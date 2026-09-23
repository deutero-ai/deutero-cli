// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// "Will this study fill by Thursday?" is the question a research ops lead is
// asked every week, and no endpoint answers it. Study stats and recruitment
// expose totals and a quota; neither exposes a rate. The rate comes from the
// interview time series, which means holding the interviews locally and doing
// the arithmetic here.

type fieldingDay struct {
	Date       string `json:"date"`
	Started    int    `json:"started"`
	Completed  int    `json:"completed"`
	Cumulative int    `json:"cumulative_completed"`
}

type fieldingReport struct {
	StudyID           string        `json:"study_id"`
	Completed         int           `json:"completed"`
	Quota             *int          `json:"quota,omitempty"`
	QuotaRemaining    *int          `json:"quota_remaining,omitempty"`
	ActiveDays        int           `json:"active_days"`
	CompletionsPerDay float64       `json:"completions_per_day"`
	ProjectedFill     string        `json:"projected_fill_date,omitempty"`
	Verdict           string        `json:"verdict"`
	Days              []fieldingDay `json:"days"`
	Note              string        `json:"note"`
}

func newNovelFieldingCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagLimit int
	var flagWindow int

	cmd := &cobra.Command{
		Use:   "fielding",
		Short: "Completions per day against quota, projecting the fill date",
		Long: `Completions per day from the interview time series, measured against the
study's response quota, with a projected date for when the quota fills.

The rate is taken over the trailing --window days of activity rather than the
whole life of the study, so a burst at launch does not flatter a study that has
since gone quiet.

--csv writes the daily series (date, started, completed, cumulative); the rate,
quota and projected fill date are in --json.`,
		Example: "  deutero-pp-cli fielding --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "fielding")
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

			interviews, err := deuteroListInterviews(ctx, c, flagStudy, cohortReal, false, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			stats, err := deuteroGetStudyStats(ctx, c, flagStudy)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			report := buildFielding(flagStudy, interviews, stats, flagWindow, time.Now().UTC())
			if flags.csv {
				return writeFieldingCSV(cmd.OutOrStdout(), report)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderFielding(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to measure fielding for (required)")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum interviews to read (1-10000)")
	cmd.Flags().IntVar(&flagWindow, "window", 14, "Trailing days of activity the rate is measured over")
	return cmd
}

// buildFielding is pure so it can be tested without a network or a clock.
func buildFielding(
	studyID string,
	interviews []deuteroInterviewSummary,
	stats *deuteroStudyStats,
	window int,
	now time.Time,
) fieldingReport {
	if window <= 0 {
		window = 14
	}
	started := map[string]int{}
	completed := map[string]int{}
	for _, iv := range interviews {
		if t, ok := deuteroParseTime(iv.StartTime); ok {
			started[t.UTC().Format("2006-01-02")]++
		}
		if !iv.Completed {
			continue
		}
		when := iv.EndTime
		if when == nil {
			when = iv.StartTime
		}
		if t, ok := deuteroParseTime(when); ok {
			completed[t.UTC().Format("2006-01-02")]++
		}
	}

	dates := map[string]bool{}
	for d := range started {
		dates[d] = true
	}
	for d := range completed {
		dates[d] = true
	}
	ordered := deuteroSortedKeys(dates)
	sort.Strings(ordered)

	report := fieldingReport{
		StudyID: studyID,
		Days:    []fieldingDay{},
		Note:    deuteroObservedNote,
	}
	cum := 0
	for _, d := range ordered {
		cum += completed[d]
		report.Days = append(report.Days, fieldingDay{
			Date:       d,
			Started:    started[d],
			Completed:  completed[d],
			Cumulative: cum,
		})
	}
	report.Completed = cum
	if stats != nil {
		report.Quota = stats.MaxResponses
		report.QuotaRemaining = stats.QuotaRemaining
		if stats.CompletedInterviews > report.Completed {
			// Trust the server's own total over our windowed sample: --limit
			// may have truncated the interview list.
			report.Completed = stats.CompletedInterviews
		}
	}

	// Rate over the trailing window, per CALENDAR day — idle days count against
	// the rate, because a projection that ignores them promises a fill date the
	// study will miss. The denominator is the elapsed span inside the window,
	// not the flag value: a study fielded for three days under --window 14 is
	// running at its three-day rate, not at a rate diluted by eleven days that
	// had not happened yet.
	cutoff := now.AddDate(0, 0, -window).AddDate(0, 0, 1).Format("2006-01-02")
	windowCompleted, activeDays := 0, 0
	firstInWindow := ""
	for _, d := range report.Days {
		if d.Date < cutoff {
			continue
		}
		windowCompleted += d.Completed
		if d.Completed > 0 || d.Started > 0 {
			activeDays++
		}
		if firstInWindow == "" || d.Date < firstInWindow {
			firstInWindow = d.Date
		}
	}
	report.ActiveDays = activeDays
	elapsed := window
	if firstInWindow != "" {
		if t, err := time.Parse("2006-01-02", firstInWindow); err == nil {
			if spanned := int(now.Sub(t).Hours()/24) + 1; spanned > 0 && spanned < window {
				elapsed = spanned
			}
		}
	}
	if elapsed > 0 {
		report.CompletionsPerDay = float64(windowCompleted) / float64(elapsed)
	}

	switch {
	case report.Quota == nil:
		report.Verdict = "No response quota set on this study, so there is nothing to fill."
	case report.QuotaRemaining != nil && *report.QuotaRemaining <= 0:
		report.Verdict = "Quota already filled."
	case report.CompletionsPerDay <= 0:
		report.Verdict = fmt.Sprintf("No completions in the last %d days — at this rate the quota never fills.", window)
	default:
		remaining := 0
		if report.QuotaRemaining != nil {
			remaining = *report.QuotaRemaining
		} else if report.Quota != nil {
			remaining = *report.Quota - report.Completed
		}
		if remaining <= 0 {
			report.Verdict = "Quota already filled."
			break
		}
		days := int(math.Ceil(float64(remaining) / report.CompletionsPerDay))
		fill := now.AddDate(0, 0, days)
		report.ProjectedFill = fill.Format("2006-01-02")
		report.Verdict = fmt.Sprintf("%d to go at %.1f/day → fills about %s (%d days).",
			remaining, report.CompletionsPerDay, report.ProjectedFill, days)
	}
	return report
}

func renderFielding(w io.Writer, r fieldingReport) {
	fmt.Fprintf(w, "Fielding for study %s\n", r.StudyID)
	fmt.Fprintf(w, "  completed %d", r.Completed)
	if r.Quota != nil {
		fmt.Fprintf(w, " of %d", *r.Quota)
	}
	fmt.Fprintf(w, "   rate %.1f/day\n", r.CompletionsPerDay)

	if len(r.Days) > 0 {
		fmt.Fprintln(w, "\n  date         started  completed  cumulative")
		show := r.Days
		if len(show) > 21 {
			show = show[len(show)-21:]
		}
		for _, d := range show {
			fmt.Fprintf(w, "  %-10s %7d  %9d  %10d\n", d.Date, d.Started, d.Completed, d.Cumulative)
		}
	}
	fmt.Fprintf(w, "\n  %s\n", r.Verdict)
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}

// writeFieldingCSV emits the daily time series; the rate, quota and projected
// fill date are in --json.
func writeFieldingCSV(w io.Writer, r fieldingReport) error {
	rows := make([][]string, 0, len(r.Days))
	for _, d := range r.Days {
		rows = append(rows, []string{d.Date, strconv.Itoa(d.Started), strconv.Itoa(d.Completed), strconv.Itoa(d.Cumulative)})
	}
	return deuteroWriteCSV(w, []string{"date", "started", "completed", "cumulative_completed"}, rows)
}
