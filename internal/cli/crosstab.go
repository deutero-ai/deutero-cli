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

// The platform collects participant characteristics specifically so answers can
// be segmented by them — and then the scale and options tally endpoints return
// flat value→count maps with no grouping parameter. "Does this differ by
// segment?" is therefore unanswerable upstream and trivial once both sides are
// local.
//
// Column marginals here are the same population the server's own tally
// endpoints count, so the numbers are checkable rather than a parallel
// universe.

type crosstabCell struct {
	Segment string `json:"segment"`
	Answer  string `json:"answer"`
	Count   int    `json:"count"`
}

type crosstabSegment struct {
	Segment        string         `json:"segment"`
	Interviews     int            `json:"interviews"`
	Completed      int            `json:"completed"`
	CompletionRate float64        `json:"completion_rate"`
	Answers        map[string]int `json:"answers"`
}

type crosstabReport struct {
	StudyID       string            `json:"study_id"`
	By            string            `json:"by"`
	Question      string            `json:"question"`
	Segments      []crosstabSegment `json:"segments"`
	AnswerTotals  map[string]int    `json:"answer_totals"`
	Cells         []crosstabCell    `json:"cells"`
	Unsegmented   int               `json:"unsegmented"`
	DetailsFailed int               `json:"details_failed"`
	Note          string            `json:"note"`
}

func newNovelCrosstabCmd(flags *rootFlags) *cobra.Command {
	var flagStudy, flagBy, flagQuestion string
	var flagLimit int
	var flagIncludeSimulated bool

	cmd := &cobra.Command{
		Use:   "crosstab",
		Short: "Segment a question's answers by a participant characteristic",
		Long: `Cross-tabulates one question's answers against one participant characteristic,
with per-cell counts and each segment's completion rate.

--question accepts a captured variable name (flow studies) or a question number
such as 7 or q7 (linear studies). --by names a characteristic variable; run
'studies characteristics get <study_id>' to see which exist.

Column totals count the same population the server's own scale and options tally
endpoints count, so the marginals reconcile.

Use this command to answer "does this differ by segment". Do NOT use this command
when you want every answer for every participant; use 'answers matrix' instead.`,
		Example: "  deutero-pp-cli crosstab --study f0ed9214-06fc-4420-87b3-31240b670fc1 --by role --question 1 --agent",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "crosstab")
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			if err := deuteroCheckLimit(flagLimit); err != nil {
				return err
			}
			if strings.TrimSpace(flagBy) == "" || strings.TrimSpace(flagQuestion) == "" {
				return usageErr(fmt.Errorf("--by and --question are both required\nUsage: %s --study <id> --by <characteristic> --question <variable-or-number>", cmd.CommandPath()))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			progress := deuteroProgressWriter(cmd, flags)

			cohort := cohortReal
			if flagIncludeSimulated {
				cohort = cohortAll
			}
			interviews, err := deuteroListInterviews(ctx, c, flagStudy, cohort, false, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			details, detailsFailed, err := deuteroFetchInterviewDetails(ctx, c, interviews, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			// Linear studies keep their answers in the message stream, so a
			// question-number crosstab needs transcripts too.
			var trs []deuteroBulkTranscript
			if _, isNum := parseQuestionRef(flagQuestion); isNum {
				trs, err = deuteroBulkFetchTranscripts(ctx, c, flagStudy, flagIncludeSimulated, false, flagLimit, progress)
				if err != nil {
					return classifyAPIError(cmd.OutOrStdout(), err, flags)
				}
			}

			report := buildCrosstab(flagStudy, flagBy, flagQuestion, details, trs)
			// A silently smaller denominator is the failure mode this command
			// exists to avoid; say how many interviews could not be read.
			report.DetailsFailed = detailsFailed
			if flags.csv {
				return writeCrosstabCSV(cmd.OutOrStdout(), report)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderCrosstab(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to cross-tabulate (required)")
	cmd.Flags().StringVar(&flagBy, "by", "", "Characteristic variable to segment by (required)")
	cmd.Flags().StringVar(&flagQuestion, "question", "", "Captured variable name, or question number like 7 / q7 (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to include (1-10000)")
	cmd.Flags().BoolVar(&flagIncludeSimulated, "include-simulated", false, "Include simulated (persona) interviews")
	return cmd
}

// parseQuestionRef reports whether --question named a linear question number
// ("7" or "q7") rather than a captured variable name.
func parseQuestionRef(q string) (int, bool) {
	s := strings.TrimSpace(strings.ToLower(q))
	s = strings.TrimPrefix(s, "q")
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// buildCrosstab is pure so it can be tested without a network.
func buildCrosstab(
	studyID, by, question string,
	details []deuteroInterviewDetail,
	trs []deuteroBulkTranscript,
) crosstabReport {
	qnum, isNum := parseQuestionRef(question)

	// Linear answers by interview id, from the message stream.
	msgAnswer := map[string]string{}
	if isNum {
		for _, tr := range trs {
			for _, m := range tr.Messages {
				if m.QuestionNumber == nil || *m.QuestionNumber != qnum {
					continue
				}
				if m.Type == nil || !isParticipantTurn(*m.Type) || m.Content == nil {
					continue
				}
				if _, seen := msgAnswer[tr.InterviewID]; !seen {
					msgAnswer[tr.InterviewID] = strings.TrimSpace(*m.Content)
				}
			}
		}
	}

	type segAgg struct {
		interviews int
		completed  int
		answers    map[string]int
	}
	segs := map[string]*segAgg{}
	totals := map[string]int{}
	unsegmented := 0

	for _, d := range details {
		segment := ""
		for _, ch := range d.Characteristics {
			if strings.EqualFold(ch.Variable, by) {
				segment = deuteroValueString(ch.Value)
				break
			}
		}
		if strings.TrimSpace(segment) == "" {
			unsegmented++
			continue
		}
		answer := ""
		if isNum {
			answer = msgAnswer[d.ID]
		} else {
			for _, v := range d.Variables {
				if strings.EqualFold(v.Name, question) {
					answer = deuteroValueString(v.Value)
					break
				}
			}
		}
		if strings.TrimSpace(answer) == "" {
			answer = "(no answer)"
		}
		a := segs[segment]
		if a == nil {
			a = &segAgg{answers: map[string]int{}}
			segs[segment] = a
		}
		a.interviews++
		if d.Completed {
			a.completed++
		}
		a.answers[answer]++
		totals[answer]++
	}

	report := crosstabReport{
		StudyID:      studyID,
		By:           by,
		Question:     question,
		Segments:     []crosstabSegment{},
		AnswerTotals: totals,
		Cells:        []crosstabCell{},
		Unsegmented:  unsegmented,
		Note:         deuteroObservedNote,
	}
	for _, name := range deuteroSortedKeys(segs) {
		a := segs[name]
		rate := 0.0
		if a.interviews > 0 {
			rate = float64(a.completed) / float64(a.interviews)
		}
		report.Segments = append(report.Segments, crosstabSegment{
			Segment:        name,
			Interviews:     a.interviews,
			Completed:      a.completed,
			CompletionRate: rate,
			Answers:        a.answers,
		})
		for _, ans := range deuteroSortedKeys(a.answers) {
			report.Cells = append(report.Cells, crosstabCell{Segment: name, Answer: ans, Count: a.answers[ans]})
		}
	}
	sort.SliceStable(report.Segments, func(i, j int) bool {
		return report.Segments[i].Interviews > report.Segments[j].Interviews
	})
	return report
}

func writeCrosstabCSV(w io.Writer, r crosstabReport) error {
	cw := csv.NewWriter(w)
	answers := deuteroSortedKeys(r.AnswerTotals)
	if err := cw.Write(append([]string{"segment", "interviews", "completion_rate"}, answers...)); err != nil {
		return err
	}
	for _, s := range r.Segments {
		rec := []string{deuteroCSVSafe(s.Segment), strconv.Itoa(s.Interviews), fmt.Sprintf("%.3f", s.CompletionRate)}
		for _, a := range answers {
			rec = append(rec, strconv.Itoa(s.Answers[a]))
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	totals := []string{"(all)", "", ""}
	for _, a := range answers {
		totals = append(totals, strconv.Itoa(r.AnswerTotals[a]))
	}
	if err := cw.Write(totals); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func renderCrosstab(w io.Writer, r crosstabReport) {
	fmt.Fprintf(w, "Crosstab for study %s — %s by %s\n", r.StudyID, r.Question, r.By)
	if len(r.Segments) == 0 {
		fmt.Fprintf(w, "\nNo interviews carry a value for characteristic %q.\n", r.By)
		if r.Unsegmented > 0 {
			fmt.Fprintf(w, "%d interviews had no value for it.\n", r.Unsegmented)
		}
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}
	answers := deuteroSortedKeys(r.AnswerTotals)
	shown := answers
	if len(shown) > 6 {
		shown = shown[:6]
	}
	fmt.Fprintf(w, "\n  %-22s %6s %7s", "segment", "n", "compl%")
	for _, a := range shown {
		fmt.Fprintf(w, " %10s", truncateCell(a, 10))
	}
	fmt.Fprintln(w)
	for _, s := range r.Segments {
		fmt.Fprintf(w, "  %-22s %6d %6.0f%%", truncateCell(s.Segment, 22), s.Interviews, s.CompletionRate*100)
		for _, a := range shown {
			fmt.Fprintf(w, " %10d", s.Answers[a])
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "  %-22s %6s %7s", "(all)", "", "")
	for _, a := range shown {
		fmt.Fprintf(w, " %10d", r.AnswerTotals[a])
	}
	fmt.Fprintln(w)
	if len(answers) > len(shown) {
		fmt.Fprintf(w, "\n  (%d more answer values — use --csv or --json for the full table)\n", len(answers)-len(shown))
	}
	if r.Unsegmented > 0 {
		fmt.Fprintf(w, "\n  %d interviews carried no value for %q and are excluded.\n", r.Unsegmented, r.By)
	}
	if r.DetailsFailed > 0 {
		fmt.Fprintf(w, "  %d interviews could not be read, so every count below is a lower bound.\n", r.DetailsFailed)
	}
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}
