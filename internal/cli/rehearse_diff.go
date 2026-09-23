// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Simulations mirror the interviewer runtime closely enough to be worth
// trusting — they execute the same steps, including fetches and signals. What
// turns a rehearsal into evidence is comparing it against reality on the same
// measures, and nothing server-side does that.
//
// The cohort split is exact rather than inferred: `simulated` is a real filter
// on the interviews endpoint, so "the personas" and "the people" are two clean
// populations over one set of measures.

type cohortMeasures struct {
	Interviews      int                `json:"interviews"`
	Completed       int                `json:"completed"`
	CompletionRate  float64            `json:"completion_rate"`
	AvgMessages     float64            `json:"avg_messages"`
	AvgDurationMin  float64            `json:"avg_duration_minutes"`
	StepReach       map[string]float64 `json:"step_reach"`
	BranchArms      map[string]float64 `json:"branch_arms"`
	VariableCapture map[string]float64 `json:"variable_capture"`
}

type rehearseDiffRow struct {
	Measure   string  `json:"measure"`
	Simulated float64 `json:"simulated"`
	Real      float64 `json:"real"`
	Delta     float64 `json:"delta"`
}

type rehearseDiffReport struct {
	StudyID   string            `json:"study_id"`
	Simulated cohortMeasures    `json:"simulated"`
	Real      cohortMeasures    `json:"real"`
	Headline  []rehearseDiffRow `json:"headline"`
	StepReach []rehearseDiffRow `json:"step_reach"`
	Branches  []rehearseDiffRow `json:"branch_arms"`
	Variables []rehearseDiffRow `json:"variable_capture"`
	Note      string            `json:"note"`
}

func newNovelRehearseDiffCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare the simulated persona cohort against real participants",
		Long: `Compares simulated (persona) interviews against real ones on the same measures:
completion rate, message count, duration, which steps each cohort reached, how
branches routed for each, and which variables got captured.

Rates are shares of each cohort, so the two sides are comparable even when the
cohorts are different sizes. A measure present in one cohort and absent in the
other shows as 0 for the absent side — read that as "not observed", not as
"cannot happen".

Use this command to compare rehearsal against reality. Do NOT use this command to
inspect one simulation run's status or credits; the generated 'simulations get'
command does that.`,
		Example: "  deutero-pp-cli rehearse diff --study 092103ce-0e01-4e81-8ea5-e85fa560504f --agent",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "rehearse diff")
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

			simIVs, err := deuteroListInterviews(ctx, c, flagStudy, cohortSimulated, false, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			realIVs, err := deuteroListInterviews(ctx, c, flagStudy, cohortReal, false, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			simTR, _, err := deuteroFetchTranscripts(ctx, c, simIVs, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			realTR, _, err := deuteroFetchTranscripts(ctx, c, realIVs, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}

			report := buildRehearseDiff(flagStudy, simIVs, realIVs, simTR, realTR)

			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderRehearseDiff(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to compare cohorts within (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to read per cohort (1-10000)")
	return cmd
}

func measureCohort(ivs []deuteroInterviewSummary, trs []deuteroTranscript) cohortMeasures {
	m := cohortMeasures{
		Interviews:      len(ivs),
		StepReach:       map[string]float64{},
		BranchArms:      map[string]float64{},
		VariableCapture: map[string]float64{},
	}
	var durSum float64
	var durN int
	for _, iv := range ivs {
		if iv.Completed {
			m.Completed++
		}
		if st, ok := deuteroParseTime(iv.StartTime); ok {
			if et, ok2 := deuteroParseTime(iv.EndTime); ok2 && et.After(st) {
				durSum += et.Sub(st).Minutes()
				durN++
			}
		}
	}
	if len(ivs) > 0 {
		m.CompletionRate = float64(m.Completed) / float64(len(ivs))
	}
	if durN > 0 {
		m.AvgDurationMin = durSum / float64(durN)
	}

	if len(trs) == 0 {
		return m
	}
	stepSeen := map[string]int{}
	armCount := map[string]int{}
	// Each arm's share must be of ITS OWN branch's traversals. Dividing by the
	// total across every branch makes a cohort that visited more branches look
	// uniformly lower on all of them, and the resulting deltas are artifacts of
	// the branch mix rather than of how any branch actually routed.
	armTotalByNode := map[string]int{}
	varSeen := map[string]int{}
	totalMsgs := 0
	for _, tr := range trs {
		totalMsgs += len(tr.Messages)
		seenStep := map[string]bool{}
		for _, msg := range tr.Messages {
			if msg.NodeID != nil && *msg.NodeID != "" {
				seenStep[*msg.NodeID] = true
			}
		}
		for s := range seenStep {
			stepSeen[s]++
		}
		for _, d := range tr.Decisions {
			label := "(unmatched)"
			switch {
			case d.MatchedClass != nil && *d.MatchedClass != "":
				label = *d.MatchedClass
			case d.UsedDefault:
				label = "(otherwise)"
			}
			node := deuteroStepLabel(d.NodeID)
			armCount[node+" → "+label]++
			armTotalByNode[node]++
		}
		seenVar := map[string]bool{}
		for _, v := range tr.Variables {
			seenVar[v.Name] = true
		}
		for n := range seenVar {
			varSeen[n]++
		}
	}
	n := float64(len(trs))
	m.AvgMessages = float64(totalMsgs) / n
	for k, v := range stepSeen {
		m.StepReach[deuteroStepLabel(k)] = float64(v) / n
	}
	for k, v := range armCount {
		node := k
		if idx := strings.Index(k, " → "); idx >= 0 {
			node = k[:idx]
		}
		if total := armTotalByNode[node]; total > 0 {
			m.BranchArms[k] = float64(v) / float64(total)
		}
	}
	for k, v := range varSeen {
		m.VariableCapture[k] = float64(v) / n
	}
	return m
}

// buildRehearseDiff is pure so it can be tested without a network.
func buildRehearseDiff(
	studyID string,
	simIVs, realIVs []deuteroInterviewSummary,
	simTR, realTR []deuteroTranscript,
) rehearseDiffReport {
	sim := measureCohort(simIVs, simTR)
	real := measureCohort(realIVs, realTR)

	diffMap := func(a, b map[string]float64) []rehearseDiffRow {
		keys := map[string]bool{}
		for k := range a {
			keys[k] = true
		}
		for k := range b {
			keys[k] = true
		}
		rows := make([]rehearseDiffRow, 0, len(keys))
		for _, k := range deuteroSortedKeys(keys) {
			rows = append(rows, rehearseDiffRow{Measure: k, Simulated: a[k], Real: b[k], Delta: b[k] - a[k]})
		}
		sort.SliceStable(rows, func(i, j int) bool {
			return absFloat(rows[i].Delta) > absFloat(rows[j].Delta)
		})
		return rows
	}

	return rehearseDiffReport{
		StudyID:   studyID,
		Simulated: sim,
		Real:      real,
		Headline: []rehearseDiffRow{
			{Measure: "interviews", Simulated: float64(sim.Interviews), Real: float64(real.Interviews), Delta: float64(real.Interviews - sim.Interviews)},
			{Measure: "completion_rate", Simulated: sim.CompletionRate, Real: real.CompletionRate, Delta: real.CompletionRate - sim.CompletionRate},
			{Measure: "avg_messages", Simulated: sim.AvgMessages, Real: real.AvgMessages, Delta: real.AvgMessages - sim.AvgMessages},
			{Measure: "avg_duration_minutes", Simulated: sim.AvgDurationMin, Real: real.AvgDurationMin, Delta: real.AvgDurationMin - sim.AvgDurationMin},
		},
		StepReach: diffMap(sim.StepReach, real.StepReach),
		Branches:  diffMap(sim.BranchArms, real.BranchArms),
		Variables: diffMap(sim.VariableCapture, real.VariableCapture),
		Note:      deuteroObservedNote,
	}
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func renderRehearseDiff(w io.Writer, r rehearseDiffReport) {
	fmt.Fprintf(w, "Rehearsal vs reality for study %s\n", r.StudyID)
	fmt.Fprintf(w, "  simulated: %d interviews    real: %d interviews\n", r.Simulated.Interviews, r.Real.Interviews)

	if r.Simulated.Interviews == 0 || r.Real.Interviews == 0 {
		missing := "simulated"
		if r.Real.Interviews == 0 {
			missing = "real"
		}
		fmt.Fprintf(w, "\nNo %s interviews to compare against — the diff needs both cohorts.\n", missing)
	}

	section := func(title string, rows []rehearseDiffRow, pct bool) {
		if len(rows) == 0 {
			return
		}
		fmt.Fprintf(w, "\n%s\n", title)
		fmt.Fprintf(w, "  %-44s %10s %10s %10s\n", "measure", "simulated", "real", "delta")
		for _, row := range rows {
			if pct {
				fmt.Fprintf(w, "  %-44s %9.0f%% %9.0f%% %+9.0f%%\n", row.Measure, row.Simulated*100, row.Real*100, row.Delta*100)
			} else {
				fmt.Fprintf(w, "  %-44s %10.2f %10.2f %+10.2f\n", row.Measure, row.Simulated, row.Real, row.Delta)
			}
		}
	}
	section("Headline", r.Headline, false)
	section("Step reach (share of cohort)", r.StepReach, true)
	section("Branch arms (share of traversals)", r.Branches, true)
	section("Variable capture (share of cohort)", r.Variables, true)
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}
