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

// Flow effects is the other half of "what did this flow do": not how it routed,
// but what it did to the outside world. `context_fetch` steps pull data in and
// `webhook` steps push signals out, and both are recorded per interview with no
// study-level endpoint — so triaging a flaky integration means reading them one
// interview at a time today.
//
// The dry_run split matters and is easy to get wrong. Simulations execute signal
// steps but do not send them; counting a simulated delivery as delivered would
// report an integration as healthy on the strength of rehearsals alone.

type effectStepStat struct {
	Step        string         `json:"step"`
	Total       int            `json:"total"`
	Failed      int            `json:"failed"`
	StatusCodes map[string]int `json:"status_codes,omitempty"`
	SampleError string         `json:"sample_error,omitempty"`
}

type signalStepStat struct {
	Step        string         `json:"step"`
	EventTypes  []string       `json:"event_types,omitempty"`
	Delivered   int            `json:"delivered"`
	Failed      int            `json:"failed"`
	DryRun      int            `json:"dry_run"`
	Retried     int            `json:"retried"`
	StatusCodes map[string]int `json:"status_codes,omitempty"`
	SampleError string         `json:"sample_error,omitempty"`
}

type flowEffectsReport struct {
	StudyID            string           `json:"study_id"`
	InterviewsExamined int              `json:"interviews_examined"`
	RecordsRead        int              `json:"records_read"`
	RecordsFailed      int              `json:"records_failed"`
	Fetches            []effectStepStat `json:"fetches"`
	Signals            []signalStepStat `json:"signals"`
	Note               string           `json:"note"`
}

func newNovelFlowEffectsCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int
	var flagIncludeSimulated bool

	cmd := &cobra.Command{
		Use:   "effects",
		Short: "What a flow's data fetches and outbound signals actually did",
		Long: `What an Interview Flow's side steps actually did across a study: failed data
fetches with their status codes, undelivered signals, and dry-run signals from
simulations kept separate from real deliveries.

Fetch and signal records are exposed one interview at a time with no study-level
endpoint, so this reads one record per interview. Use --limit to bound that.

Use this command for what a flow's 'Bring in data' and 'Send a signal' steps did.
Do NOT use this command for how the flow routed or what it captured; use
'flow coverage' instead. Do NOT use it for org-level webhook endpoint delivery
history; that is 'webhooks deliveries'.`,
		Example: "  deutero-pp-cli flow effects --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "flow effects")
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
			effects, failed, err := deuteroFetchEffects(ctx, c, interviews, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}

			report := buildFlowEffects(flagStudy, len(interviews), failed, effects)

			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderFlowEffects(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID whose flow side-effects to summarise (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to read effect records for (1-10000)")
	cmd.Flags().BoolVar(&flagIncludeSimulated, "include-simulated", false, "Include simulated interviews (their signals are dry-run and reported separately)")
	return cmd
}

// buildFlowEffects is pure so it can be tested without a network.
func buildFlowEffects(studyID string, examined, failed int, effects []deuteroInterviewEffects) flowEffectsReport {
	type fetchAgg struct {
		total  int
		failed int
		codes  map[string]int
		sample string
	}
	type signalAgg struct {
		delivered int
		failed    int
		dryRun    int
		retried   int
		codes     map[string]int
		events    map[string]bool
		sample    string
	}
	fetches := map[string]*fetchAgg{}
	signals := map[string]*signalAgg{}

	for _, ef := range effects {
		for _, f := range ef.Fetches {
			a := fetches[f.NodeID]
			if a == nil {
				a = &fetchAgg{codes: map[string]int{}}
				fetches[f.NodeID] = a
			}
			a.total++
			if !f.OK {
				a.failed++
				if a.sample == "" && f.Error != nil {
					a.sample = *f.Error
				}
			}
			if f.StatusCode != nil {
				a.codes[fmt.Sprintf("%d", *f.StatusCode)]++
			}
		}
		for _, s := range ef.Signals {
			a := signals[s.NodeID]
			if a == nil {
				a = &signalAgg{codes: map[string]int{}, events: map[string]bool{}}
				signals[s.NodeID] = a
			}
			switch {
			case s.DryRun:
				a.dryRun++
			case s.Success:
				a.delivered++
			default:
				a.failed++
				if a.sample == "" && s.Error != nil {
					a.sample = *s.Error
				}
			}
			if s.Attempt != nil && *s.Attempt > 1 {
				a.retried++
			}
			if s.StatusCode != nil {
				a.codes[fmt.Sprintf("%d", *s.StatusCode)]++
			}
			if s.EventType != nil && *s.EventType != "" {
				a.events[*s.EventType] = true
			}
		}
	}

	report := flowEffectsReport{
		StudyID:            studyID,
		InterviewsExamined: examined,
		RecordsRead:        len(effects),
		RecordsFailed:      failed,
		Fetches:            []effectStepStat{},
		Signals:            []signalStepStat{},
		Note:               deuteroObservedNote,
	}
	for _, step := range deuteroSortedKeys(fetches) {
		a := fetches[step]
		report.Fetches = append(report.Fetches, effectStepStat{
			Step:        deuteroStepLabel(step),
			Total:       a.total,
			Failed:      a.failed,
			StatusCodes: a.codes,
			SampleError: a.sample,
		})
	}
	sort.SliceStable(report.Fetches, func(i, j int) bool { return report.Fetches[i].Failed > report.Fetches[j].Failed })

	for _, step := range deuteroSortedKeys(signals) {
		a := signals[step]
		report.Signals = append(report.Signals, signalStepStat{
			Step:        deuteroStepLabel(step),
			EventTypes:  deuteroSortedKeys(a.events),
			Delivered:   a.delivered,
			Failed:      a.failed,
			DryRun:      a.dryRun,
			Retried:     a.retried,
			StatusCodes: a.codes,
			SampleError: a.sample,
		})
	}
	sort.SliceStable(report.Signals, func(i, j int) bool { return report.Signals[i].Failed > report.Signals[j].Failed })
	return report
}

func renderFlowEffects(w io.Writer, r flowEffectsReport) {
	fmt.Fprintf(w, "Flow effects for study %s\n", r.StudyID)
	fmt.Fprintf(w, "  interviews examined: %d   records read: %d", r.InterviewsExamined, r.RecordsRead)
	if r.RecordsFailed > 0 {
		fmt.Fprintf(w, "   unreadable: %d", r.RecordsFailed)
	}
	fmt.Fprintln(w)

	if len(r.Fetches) == 0 && len(r.Signals) == 0 {
		fmt.Fprintln(w, "\nNo fetch or signal records observed. Either this flow has no 'Bring in data' or 'Send a signal' steps, or no interview has run one yet.")
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}

	if len(r.Fetches) > 0 {
		fmt.Fprintln(w, "\nData fetches")
		for _, f := range r.Fetches {
			fmt.Fprintf(w, "  %-32s %d run", f.Step, f.Total)
			if f.Failed > 0 {
				fmt.Fprintf(w, "   FAILED %d", f.Failed)
			}
			if len(f.StatusCodes) > 0 {
				fmt.Fprintf(w, "   codes %v", f.StatusCodes)
			}
			fmt.Fprintln(w)
			if f.SampleError != "" {
				fmt.Fprintf(w, "      e.g. %s\n", f.SampleError)
			}
		}
	}
	if len(r.Signals) > 0 {
		fmt.Fprintln(w, "\nSignals")
		for _, s := range r.Signals {
			fmt.Fprintf(w, "  %-32s delivered %d", s.Step, s.Delivered)
			if s.Failed > 0 {
				fmt.Fprintf(w, "   FAILED %d", s.Failed)
			}
			if s.DryRun > 0 {
				fmt.Fprintf(w, "   dry-run %d (simulated, not sent)", s.DryRun)
			}
			if s.Retried > 0 {
				fmt.Fprintf(w, "   retried %d", s.Retried)
			}
			fmt.Fprintln(w)
			if s.SampleError != "" {
				fmt.Fprintf(w, "      e.g. %s\n", s.SampleError)
			}
		}
	}
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}
