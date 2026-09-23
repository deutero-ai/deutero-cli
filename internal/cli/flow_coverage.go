// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"
)

// Flow coverage answers the question a branching interview makes urgent and the
// API cannot answer in one call: across a whole study, how did the flow route?
//
// `decisions` comes back from the per-interview transcript endpoint and never
// from the bulk one, so this is an N+1 fan-out by construction. That is the
// point — the cost is paid once here instead of by a researcher opening
// interviews one at a time, which is what they do today.
//
// The upstream spec's own description of `used_default` says it is "worth
// checking when a branch looks like it fired the wrong way". This is that check.

type branchArmStat struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type branchStat struct {
	Step            string          `json:"step"`
	Mode            string          `json:"mode,omitempty"`
	Traversals      int             `json:"traversals"`
	UsedDefault     int             `json:"used_default"`
	UsedDefaultRate float64         `json:"used_default_rate"`
	ErrorCount      int             `json:"error_count"`
	Arms            []branchArmStat `json:"arms"`
	OfferedNotTaken []string        `json:"offered_not_taken,omitempty"`
}

type stepReachStat struct {
	Step       string `json:"step"`
	Interviews int    `json:"interviews"`
}

type variableStat struct {
	Name        string   `json:"name"`
	Interviews  int      `json:"interviews"`
	CaptureRate float64  `json:"capture_rate"`
	ValueTypes  []string `json:"value_types,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	LoopMerged  int      `json:"loop_merged"`
}

type flowCoverageReport struct {
	StudyID            string          `json:"study_id"`
	InterviewsExamined int             `json:"interviews_examined"`
	TranscriptsRead    int             `json:"transcripts_read"`
	TranscriptsFailed  int             `json:"transcripts_failed"`
	Branches           []branchStat    `json:"branches"`
	Steps              []stepReachStat `json:"steps"`
	Variables          []variableStat  `json:"variables"`
	Note               string          `json:"note"`
}

func newNovelFlowCoverageCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int
	var flagIncludeSimulated bool

	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "How an Interview Flow actually routed across a whole study",
		Long: `How an Interview Flow actually routed across a whole study: each branch's arm
distribution, how often the 'otherwise' path fired, classifier errors, which
steps interviews reached, and which declared variables are being captured.

Branch decisions come back from the per-interview transcript endpoint and never
from the bulk one, so this reads one transcript per interview. Use --limit to
bound that.

Use this command for how a flow routed and what it captured across a study. Do
NOT use this command for what a flow's steps did outside the conversation (data
fetches, signal deliveries); use 'flow effects' instead. Do NOT use it to get
per-participant answers as a table; use 'answers matrix' instead.`,
		Example: "  deutero-pp-cli flow coverage --study f0ed9214-06fc-4420-87b3-31240b670fc1 --agent",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "flow coverage")
			}
			if err := deuteroRejectCSV(cmd, flags); err != nil {
				return err
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
			transcripts, failed, err := deuteroFetchTranscripts(ctx, c, interviews, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}

			report := buildFlowCoverage(flagStudy, len(interviews), failed, transcripts)

			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderFlowCoverage(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID whose flow behaviour to summarise (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to read transcripts for (1-10000)")
	cmd.Flags().BoolVar(&flagIncludeSimulated, "include-simulated", false, "Include simulated (persona) interviews alongside real ones")
	return cmd
}

// buildFlowCoverage is pure so it can be tested without a network.
func buildFlowCoverage(studyID string, examined, failed int, transcripts []deuteroTranscript) flowCoverageReport {
	type branchAgg struct {
		mode        string
		traversals  int
		usedDefault int
		errCount    int
		arms        map[string]int
		offered     map[string]bool
	}
	branches := map[string]*branchAgg{}
	stepReach := map[string]map[string]bool{}
	type varAgg struct {
		interviews map[string]bool
		types      map[string]bool
		sources    map[string]bool
		loopMerged int
	}
	vars := map[string]*varAgg{}

	for _, tr := range transcripts {
		for _, d := range tr.Decisions {
			b := branches[d.NodeID]
			if b == nil {
				b = &branchAgg{arms: map[string]int{}, offered: map[string]bool{}}
				branches[d.NodeID] = b
			}
			b.traversals++
			if d.Mode != nil && b.mode == "" {
				b.mode = *d.Mode
			}
			if d.UsedDefault {
				b.usedDefault++
			}
			if d.Error != nil && *d.Error != "" {
				b.errCount++
			}
			for _, cl := range d.Classes {
				b.offered[cl] = true
			}
			switch {
			case d.MatchedClass != nil && *d.MatchedClass != "":
				b.arms[*d.MatchedClass]++
			case d.UsedDefault:
				b.arms["(otherwise)"]++
			default:
				b.arms["(unmatched)"]++
			}
		}
		for _, m := range tr.Messages {
			if m.NodeID == nil || *m.NodeID == "" {
				continue
			}
			set := stepReach[*m.NodeID]
			if set == nil {
				set = map[string]bool{}
				stepReach[*m.NodeID] = set
			}
			set[tr.InterviewID] = true
		}
		for _, v := range tr.Variables {
			a := vars[v.Name]
			if a == nil {
				a = &varAgg{interviews: map[string]bool{}, types: map[string]bool{}, sources: map[string]bool{}}
				vars[v.Name] = a
			}
			a.interviews[tr.InterviewID] = true
			if v.ValueType != nil && *v.ValueType != "" {
				a.types[*v.ValueType] = true
			}
			if v.Source != nil && *v.Source != "" {
				a.sources[*v.Source] = true
			}
			if v.NodeVisit != nil && *v.NodeVisit > 1 {
				a.loopMerged++
			}
		}
	}

	read := len(transcripts)
	report := flowCoverageReport{
		StudyID:            studyID,
		InterviewsExamined: examined,
		TranscriptsRead:    read,
		TranscriptsFailed:  failed,
		Branches:           []branchStat{},
		Steps:              []stepReachStat{},
		Variables:          []variableStat{},
		Note:               deuteroObservedNote,
	}

	for _, step := range deuteroSortedKeys(branches) {
		b := branches[step]
		arms := make([]branchArmStat, 0, len(b.arms))
		taken := map[string]bool{}
		for _, label := range deuteroSortedKeys(b.arms) {
			arms = append(arms, branchArmStat{Label: label, Count: b.arms[label]})
			taken[label] = true
		}
		sort.SliceStable(arms, func(i, j int) bool { return arms[i].Count > arms[j].Count })

		notTaken := []string{}
		for _, label := range deuteroSortedKeys(b.offered) {
			if !taken[label] {
				notTaken = append(notTaken, label)
			}
		}
		rate := 0.0
		if b.traversals > 0 {
			rate = float64(b.usedDefault) / float64(b.traversals)
		}
		report.Branches = append(report.Branches, branchStat{
			Step:            deuteroStepLabel(step),
			Mode:            b.mode,
			Traversals:      b.traversals,
			UsedDefault:     b.usedDefault,
			UsedDefaultRate: rate,
			ErrorCount:      b.errCount,
			Arms:            arms,
			OfferedNotTaken: notTaken,
		})
	}

	for _, step := range deuteroSortedKeys(stepReach) {
		report.Steps = append(report.Steps, stepReachStat{
			Step:       deuteroStepLabel(step),
			Interviews: len(stepReach[step]),
		})
	}
	sort.SliceStable(report.Steps, func(i, j int) bool { return report.Steps[i].Interviews > report.Steps[j].Interviews })

	for _, name := range deuteroSortedKeys(vars) {
		a := vars[name]
		rate := 0.0
		if read > 0 {
			rate = float64(len(a.interviews)) / float64(read)
		}
		report.Variables = append(report.Variables, variableStat{
			Name:        name,
			Interviews:  len(a.interviews),
			CaptureRate: rate,
			ValueTypes:  deuteroSortedKeys(a.types),
			Sources:     deuteroSortedKeys(a.sources),
			LoopMerged:  a.loopMerged,
		})
	}
	return report
}

func renderFlowCoverage(w io.Writer, r flowCoverageReport) {
	fmt.Fprintf(w, "Flow coverage for study %s\n", r.StudyID)
	fmt.Fprintf(w, "  interviews examined: %d   transcripts read: %d", r.InterviewsExamined, r.TranscriptsRead)
	if r.TranscriptsFailed > 0 {
		fmt.Fprintf(w, "   unreadable: %d", r.TranscriptsFailed)
	}
	fmt.Fprintln(w)

	if len(r.Branches) == 0 && len(r.Steps) == 0 && len(r.Variables) == 0 {
		fmt.Fprintln(w, "\nNo flow records observed. Either this study runs a linear question list rather than a flow, or no interviews have produced routing records yet.")
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}

	if len(r.Branches) > 0 {
		fmt.Fprintln(w, "\nBranches")
		for _, b := range r.Branches {
			fmt.Fprintf(w, "  %s", b.Step)
			if b.Mode != "" {
				fmt.Fprintf(w, "  [%s]", b.Mode)
			}
			fmt.Fprintf(w, "\n    traversals %d   otherwise %d (%.0f%%)", b.Traversals, b.UsedDefault, b.UsedDefaultRate*100)
			if b.ErrorCount > 0 {
				fmt.Fprintf(w, "   classifier errors %d", b.ErrorCount)
			}
			fmt.Fprintln(w)
			for _, a := range b.Arms {
				fmt.Fprintf(w, "      %-28s %d\n", a.Label, a.Count)
			}
			if len(b.OfferedNotTaken) > 0 {
				fmt.Fprintf(w, "      not observed: %v\n", b.OfferedNotTaken)
			}
		}
	}
	if len(r.Steps) > 0 {
		fmt.Fprintln(w, "\nStep reach")
		for _, s := range r.Steps {
			fmt.Fprintf(w, "  %-36s %d interviews\n", s.Step, s.Interviews)
		}
	}
	if len(r.Variables) > 0 {
		fmt.Fprintln(w, "\nCaptured variables")
		for _, v := range r.Variables {
			fmt.Fprintf(w, "  %-28s %3d interviews (%.0f%%)", v.Name, v.Interviews, v.CaptureRate*100)
			if len(v.ValueTypes) > 0 {
				fmt.Fprintf(w, "  types %v", v.ValueTypes)
			}
			if v.LoopMerged > 0 {
				fmt.Fprintf(w, "  loop-merged %d", v.LoopMerged)
			}
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}
