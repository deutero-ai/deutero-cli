// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// The interview × question rectangle. No endpoint returns it: bulk transcripts
// return message streams, and the analysis endpoints return per-question
// aggregates. Anything that wants to count or correlate answers has to rebuild
// the rectangle by hand, which is what this command does once and properly.
//
// Both interview modes are covered deliberately. The spec is explicit that
// captured variables are "Empty for a study that captures nothing, including
// every linear study", so a variables-only implementation would return an empty
// table for every linear study on the platform. Linear answers come from the
// message stream instead, keyed on question_number.

type answersCell = string

type answersRow struct {
	InterviewID           string            `json:"interview_id"`
	ExternalParticipantID string            `json:"external_participant_id,omitempty"`
	Completed             bool              `json:"completed"`
	Simulated             bool              `json:"simulated"`
	Characteristics       map[string]string `json:"characteristics,omitempty"`
	Cells                 map[string]string `json:"cells"`
}

type answersMatrix struct {
	StudyID               string       `json:"study_id"`
	Columns               []string     `json:"columns"`
	CharacteristicCols    []string     `json:"characteristic_columns,omitempty"`
	Rows                  []answersRow `json:"rows"`
	Source                string       `json:"source"`
	CharacteristicsFailed int          `json:"characteristics_failed,omitempty"`
	Note                  string       `json:"note"`
}

func newNovelAnswersMatrixCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int
	var flagIncludeSimulated bool
	var flagCompletedOnly bool
	var flagWithCharacteristics bool

	cmd := &cobra.Command{
		Use:   "matrix",
		Short: "One row per interview, one column per question or captured variable",
		Long: `Builds the interview by question rectangle: one row per interview, one column
per question or captured variable, plus the external participant id so the table
joins back to your own records.

Works in both interview modes. Flow studies contribute captured variables; linear
studies contribute their message-stream answers keyed on question number, because
captured variables are empty for every linear study.

--with-characteristics adds participant attribute columns, at the cost of one
extra request per interview.

Use this command to count or correlate answers. Do NOT use this command to break
one question down by a participant characteristic; use 'crosstab' instead.`,
		Example: "  deutero-pp-cli answers matrix --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "answers matrix")
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			if err := deuteroCheckLimit(flagLimit); err != nil {
				return err
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			progress := deuteroProgressWriter(cmd, flags)
			trs, err := deuteroBulkFetchTranscripts(cmd.Context(), c, flagStudy, flagIncludeSimulated, flagCompletedOnly, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			matrix := buildAnswersMatrix(flagStudy, trs)

			// Characteristics live only on the per-interview detail record, so
			// widening the rectangle with them costs one request per interview.
			// Off by default: most callers want the answers, and paying an N+1
			// silently for columns nobody asked for is the wrong default.
			if flagWithCharacteristics && len(matrix.Rows) > 0 {
				ivs := make([]deuteroInterviewSummary, 0, len(matrix.Rows))
				for _, r := range matrix.Rows {
					ivs = append(ivs, deuteroInterviewSummary{ID: r.InterviewID})
				}
				details, detailsFailed, derr := deuteroFetchInterviewDetails(cmd.Context(), c, ivs, progress)
				if derr != nil {
					return classifyAPIError(cmd.OutOrStdout(), derr, flags)
				}
				widenWithCharacteristics(&matrix, details)
				matrix.CharacteristicsFailed = detailsFailed
			}

			if flags.csv {
				return writeAnswersCSV(cmd.OutOrStdout(), matrix)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderAnswersMatrix(cmd.OutOrStdout(), matrix)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), matrix, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to build the answer table for (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to include (1-10000)")
	cmd.Flags().BoolVar(&flagIncludeSimulated, "include-simulated", false, "Include simulated (persona) interviews")
	cmd.Flags().BoolVar(&flagCompletedOnly, "completed", false, "Only include completed interviews")
	cmd.Flags().BoolVar(&flagWithCharacteristics, "with-characteristics", false, "Add participant characteristic columns (one extra request per interview)")
	return cmd
}

// buildAnswersMatrix is pure so it can be tested without a network.
func buildAnswersMatrix(studyID string, trs []deuteroBulkTranscript) answersMatrix {
	cols := map[string]bool{}
	rows := make([]answersRow, 0, len(trs))
	sawVariables, sawMessages := false, false

	for _, tr := range trs {
		cells := map[string]string{}
		for _, v := range tr.Variables {
			sawVariables = true
			// Namespaced: a flow study may legitimately declare a variable
			// called "q3", which would otherwise collide with linear question 3
			// and silently suppress one of them.
			col := "var:" + v.Name
			cells[col] = deuteroValueString(v.Value)
			cols[col] = true
		}
		// Linear studies carry answers only in the message stream. Take the
		// participant's turn that follows each numbered question.
		for _, m := range tr.Messages {
			if m.QuestionNumber == nil || m.Content == nil {
				continue
			}
			if m.Type == nil || !isParticipantTurn(*m.Type) {
				continue
			}
			col := "q" + strconv.Itoa(*m.QuestionNumber)
			sawMessages = true
			// Keep every participant turn for a question, joined. A probe
			// follow-up is often where the substance is, and first-wins would
			// throw it away.
			if prev, taken := cells[col]; taken && prev != "" {
				cells[col] = prev + " ⏎ " + strings.TrimSpace(*m.Content)
			} else {
				cells[col] = strings.TrimSpace(*m.Content)
			}
			cols[col] = true
		}
		rows = append(rows, answersRow{
			InterviewID:           tr.InterviewID,
			ExternalParticipantID: deuteroDeref(tr.ExternalParticipantID),
			Completed:             tr.Completed,
			Simulated:             tr.Simulated,
			Cells:                 cells,
		})
	}

	source := "no answers observed"
	switch {
	case sawVariables && sawMessages:
		source = "captured variables + message stream"
	case sawVariables:
		source = "captured variables (flow study)"
	case sawMessages:
		source = "message stream (linear study)"
	}

	return answersMatrix{
		StudyID: studyID,
		Columns: sortAnswerColumns(cols),
		Rows:    rows,
		Source:  source,
		Note:    deuteroObservedNote,
	}
}

// isParticipantTurn identifies the answering side of a transcript. The upstream
// vocabulary has varied ("user"/"participant"/"human"), so accept the family
// rather than pinning one spelling and silently producing an empty table.
func isParticipantTurn(t string) bool {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "user", "participant", "human", "answer", "response":
		return true
	}
	return false
}

// sortAnswerColumns puts numbered question columns in numeric order (q2 before
// q10) and named variables after them, alphabetically.
func sortAnswerColumns(set map[string]bool) []string {
	var qs, names []string
	for k := range set {
		if n, ok := questionColumnNumber(k); ok {
			_ = n
			qs = append(qs, k)
			continue
		}
		names = append(names, k)
	}
	sort.Slice(qs, func(i, j int) bool {
		a, _ := questionColumnNumber(qs[i])
		b, _ := questionColumnNumber(qs[j])
		return a < b
	})
	sort.Strings(names)
	return append(qs, names...)
}

func questionColumnNumber(col string) (int, bool) {
	if !strings.HasPrefix(col, "q") {
		return 0, false
	}
	n, err := strconv.Atoi(col[1:])
	if err != nil {
		return 0, false
	}
	return n, true
}

func writeAnswersCSV(w io.Writer, m answersMatrix) error {
	cw := csv.NewWriter(w)
	header := []string{"interview_id", "external_participant_id", "completed", "simulated"}
	header = append(header, m.CharacteristicCols...)
	header = append(header, m.Columns...)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, r := range m.Rows {
		rec := []string{r.InterviewID, r.ExternalParticipantID, fmt.Sprintf("%t", r.Completed), fmt.Sprintf("%t", r.Simulated)}
		for _, c := range m.CharacteristicCols {
			rec = append(rec, deuteroCSVSafe(r.Characteristics[c]))
		}
		for _, c := range m.Columns {
			rec = append(rec, deuteroCSVSafe(r.Cells[c]))
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func renderAnswersMatrix(w io.Writer, m answersMatrix) {
	fmt.Fprintf(w, "Answer matrix for study %s\n", m.StudyID)
	fmt.Fprintf(w, "  %d interviews × %d columns   source: %s\n", len(m.Rows), len(m.Columns), m.Source)
	if len(m.Rows) == 0 || len(m.Columns) == 0 {
		fmt.Fprintln(w, "\nNothing to tabulate yet.")
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}
	shown := m.Columns
	truncated := false
	if len(shown) > 6 {
		shown = shown[:6]
		truncated = true
	}
	fmt.Fprintf(w, "\n  %-38s", "interview")
	for _, c := range shown {
		fmt.Fprintf(w, " %-18s", truncateCell(c, 18))
	}
	fmt.Fprintln(w)
	for _, r := range m.Rows {
		label := r.InterviewID
		if r.ExternalParticipantID != "" {
			label = r.ExternalParticipantID + " (" + shortID(r.InterviewID) + ")"
		}
		fmt.Fprintf(w, "  %-38s", truncateCell(label, 38))
		for _, c := range shown {
			fmt.Fprintf(w, " %-18s", truncateCell(r.Cells[c], 18))
		}
		fmt.Fprintln(w)
	}
	if truncated {
		fmt.Fprintf(w, "\n  (%d more columns — use --csv or --json for the full table)\n", len(m.Columns)-len(shown))
	}
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}

func truncateCell(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// widenWithCharacteristics adds participant attribute columns to the rectangle.
// Kept separate from buildAnswersMatrix so the pure builder stays testable
// without a network, and so the extra fan-out is visibly opt-in.
func widenWithCharacteristics(m *answersMatrix, details []deuteroInterviewDetail) {
	byID := make(map[string][]deuteroCharacteristic, len(details))
	for _, d := range details {
		byID[d.ID] = d.Characteristics
	}
	cols := map[string]bool{}
	for i := range m.Rows {
		chs := byID[m.Rows[i].InterviewID]
		if len(chs) == 0 {
			continue
		}
		vals := make(map[string]string, len(chs))
		for _, ch := range chs {
			vals[ch.Variable] = deuteroValueString(ch.Value)
			cols[ch.Variable] = true
		}
		m.Rows[i].Characteristics = vals
	}
	m.CharacteristicCols = deuteroSortedKeys(cols)
}
