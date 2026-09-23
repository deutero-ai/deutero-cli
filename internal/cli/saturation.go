// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Saturation is the qualitative-research stopping rule: keep interviewing until
// new interviews stop contributing new language. No endpoint has any concept of
// it, and it needs the whole corpus in chronological order, so it only exists
// once the transcripts are local.
//
// This counts first-appearance terms, which is a mechanical proxy for new
// content — deliberately not an LLM judgment, so the number is reproducible and
// checkable.

type saturationPoint struct {
	Index       int     `json:"index"`
	InterviewID string  `json:"interview_id"`
	NewTerms    int     `json:"new_terms"`
	TotalTerms  int     `json:"total_terms"`
	NewShare    float64 `json:"new_share"`
}

type saturationReport struct {
	StudyID     string            `json:"study_id"`
	Interviews  int               `json:"interviews"`
	Curve       []saturationPoint `json:"curve"`
	PlateauAt   int               `json:"plateau_at"`
	PlateauNote string            `json:"plateau_note"`
	Note        string            `json:"note"`
}

func newNovelSaturationCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int
	var flagIncludeSimulated bool
	var flagThreshold float64

	cmd := &cobra.Command{
		Use:   "saturation",
		Short: "Where new interviews stopped contributing new language",
		Long: `Orders the study's transcripts chronologically and counts how many genuinely
new terms each successive interview contributes, marking where the curve
flattens.

The plateau is the first interview after which every subsequent interview
contributes a new-term share below --threshold. It is a mechanical word-level
proxy for new content, not a judgement about meaning — treat it as evidence for a
stopping decision, not as the decision.

--csv writes the curve (one row per interview); the plateau is in --json.`,
		Example: "  deutero-pp-cli saturation --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "saturation")
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
			trs, err := deuteroBulkFetchTranscripts(cmd.Context(), c, flagStudy, flagIncludeSimulated, false, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			report := buildSaturation(flagStudy, trs, flagThreshold)
			if flags.csv {
				return writeSaturationCSV(cmd.OutOrStdout(), report)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderSaturation(cmd.OutOrStdout(), report, flagThreshold)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to measure saturation for (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to include (1-10000)")
	cmd.Flags().BoolVar(&flagIncludeSimulated, "include-simulated", false, "Include simulated (persona) interviews")
	cmd.Flags().Float64Var(&flagThreshold, "threshold", 0.05, "New-term share below which the curve counts as flat")
	return cmd
}

// buildSaturation is pure so it can be tested without a network.
func buildSaturation(studyID string, trs []deuteroBulkTranscript, threshold float64) saturationReport {
	ordered := append([]deuteroBulkTranscript(nil), trs...)
	sort.SliceStable(ordered, func(i, j int) bool {
		ti, oki := deuteroParseTime(ordered[i].StartTime)
		tj, okj := deuteroParseTime(ordered[j].StartTime)
		if oki && okj {
			return ti.Before(tj)
		}
		if oki != okj {
			return oki
		}
		return ordered[i].InterviewID < ordered[j].InterviewID
	})

	seen := map[string]bool{}
	report := saturationReport{
		StudyID:    studyID,
		Interviews: len(ordered),
		Curve:      []saturationPoint{},
		PlateauAt:  0,
		Note:       deuteroObservedNote,
	}
	for i, tr := range ordered {
		terms := map[string]bool{}
		for _, m := range tr.Messages {
			if m.Content == nil || m.Type == nil || !isParticipantTurn(*m.Type) {
				continue
			}
			for _, w := range tokenizeForSaturation(*m.Content) {
				terms[w] = true
			}
		}
		newCount := 0
		for t := range terms {
			if !seen[t] {
				newCount++
				seen[t] = true
			}
		}
		share := 0.0
		if len(terms) > 0 {
			share = float64(newCount) / float64(len(terms))
		}
		report.Curve = append(report.Curve, saturationPoint{
			Index:       i + 1,
			InterviewID: tr.InterviewID,
			NewTerms:    newCount,
			TotalTerms:  len(terms),
			NewShare:    share,
		})
	}

	// The plateau is the earliest index after which nothing exceeds the
	// threshold. Scanning backwards means a single late outlier correctly
	// pushes the plateau later rather than being averaged away.
	plateau := 0
	for i := len(report.Curve) - 1; i >= 0; i-- {
		if report.Curve[i].NewShare >= threshold {
			plateau = report.Curve[i].Index
			break
		}
	}
	report.PlateauAt = plateau
	switch {
	case len(report.Curve) == 0:
		report.PlateauNote = "No interviews to measure."
	case plateau == 0:
		report.PlateauNote = "Flat from the first interview — either the corpus is tiny or answers are highly repetitive."
	case plateau >= len(report.Curve):
		report.PlateauNote = "Still contributing new language at the most recent interview; saturation not reached."
	default:
		report.PlateauNote = fmt.Sprintf("Curve flattens after interview %d of %d.", plateau, len(report.Curve))
	}
	return report
}

// tokenizeForSaturation lowercases, strips punctuation and drops very short and
// stop-ish words so the curve tracks content rather than grammar.
func tokenizeForSaturation(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		w := b.String()
		b.Reset()
		if len(w) < 4 || saturationStopWords[w] {
			return
		}
		out = append(out, w)
	}
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '\'' {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

var saturationStopWords = map[string]bool{
	"that": true, "this": true, "with": true, "have": true, "just": true,
	"like": true, "really": true, "think": true, "know": true, "would": true,
	"they": true, "there": true, "then": true, "them": true, "what": true,
	"when": true, "your": true, "about": true, "because": true, "been": true,
	"were": true, "from": true, "into": true, "more": true, "some": true,
	"very": true, "much": true, "kind": true, "sort": true, "thing": true,
	"yeah": true, "okay": true, "mean": true, "also": true, "make": true,
}

func renderSaturation(w io.Writer, r saturationReport, threshold float64) {
	fmt.Fprintf(w, "Saturation for study %s\n", r.StudyID)
	fmt.Fprintf(w, "  %d interviews, threshold %.0f%% new terms\n", r.Interviews, threshold*100)
	if len(r.Curve) == 0 {
		fmt.Fprintln(w, "\nNo interviews to measure.")
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}
	fmt.Fprintln(w)
	maxNew := 1
	for _, p := range r.Curve {
		if p.NewTerms > maxNew {
			maxNew = p.NewTerms
		}
	}
	for _, p := range r.Curve {
		bar := int(float64(p.NewTerms) / float64(maxNew) * 32)
		marker := " "
		if r.PlateauAt > 0 && p.Index == r.PlateauAt {
			marker = "◀ plateau"
		}
		fmt.Fprintf(w, "  %3d  %-32s %4d new (%3.0f%%) %s\n",
			p.Index, strings.Repeat("█", bar), p.NewTerms, p.NewShare*100, marker)
	}
	fmt.Fprintf(w, "\n  %s\n", r.PlateauNote)
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}

// writeSaturationCSV emits the curve, one row per interview in chronological
// order; the plateau verdict is in --json.
func writeSaturationCSV(w io.Writer, r saturationReport) error {
	rows := make([][]string, 0, len(r.Curve))
	for _, p := range r.Curve {
		rows = append(rows, []string{
			strconv.Itoa(p.Index),
			p.InterviewID,
			strconv.Itoa(p.NewTerms),
			strconv.Itoa(p.TotalTerms),
			strconv.FormatFloat(p.NewShare, 'f', 4, 64),
		})
	}
	return deuteroWriteCSV(w, []string{"index", "interview_id", "new_terms", "total_terms", "new_share"}, rows)
}
