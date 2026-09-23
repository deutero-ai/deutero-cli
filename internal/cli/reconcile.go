// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// The gap between "this person finished" and "our system heard about it".
//
// Completions live on the interviews endpoint; delivery outcomes live on a
// webhook's delivery log; nothing joins them. The interesting set is the
// anti-join — completions with no successful delivery — which is exactly the
// list a payout run needs and exactly what neither endpoint can produce.
//
// This depends on deliveries carrying external_participant_id. Older deliveries
// predate that field and report it as null; they are counted separately rather
// than silently treated as unmatched, because "we cannot tell" and "we failed to
// notify" are different answers and only one of them is actionable.

type reconcileRow struct {
	ExternalParticipantID string `json:"external_participant_id"`
	InterviewID           string `json:"interview_id"`
	CompletedAt           string `json:"completed_at,omitempty"`
	DeliveryAttempts      int    `json:"delivery_attempts"`
	Delivered             bool   `json:"delivered"`
	LastStatus            string `json:"last_status,omitempty"`
	LastError             string `json:"last_error,omitempty"`
}

type reconcileReport struct {
	StudyID                string         `json:"study_id"`
	WebhookID              string         `json:"webhook_id"`
	CompletedInterviews    int            `json:"completed_interviews"`
	WithExternalID         int            `json:"with_external_id"`
	WithoutExternalID      int            `json:"without_external_id"`
	Delivered              int            `json:"delivered"`
	NeverDelivered         int            `json:"never_delivered"`
	UnattributedDeliveries int            `json:"unattributed_deliveries"`
	Missing                []reconcileRow `json:"missing"`
	Note                   string         `json:"note"`
}

type deuteroDelivery struct {
	ID                    string  `json:"id"`
	EventType             string  `json:"event_type"`
	MsgID                 string  `json:"msg_id"`
	StatusCode            *int    `json:"status_code"`
	Success               bool    `json:"success"`
	ErrorMessage          *string `json:"error_message"`
	ExternalParticipantID *string `json:"external_participant_id"`
	CreatedAt             *string `json:"created_at"`
}

type deuteroDeliveryList struct {
	WebhookID  string            `json:"webhook_id"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	Deliveries []deuteroDelivery `json:"deliveries"`
}

func newNovelReconcileCmd(flags *rootFlags) *cobra.Command {
	var flagStudy, flagWebhook, flagEvent string
	var flagLimit int

	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Find completions your webhook never successfully reported",
		Long: `Joins a study's completed interviews against a webhook endpoint's delivery log
on your own participant id, and lists the completions that were never
successfully delivered.

Three populations are reported separately, because they need different actions:
completions with no successful delivery (act on these), completions carrying no
external participant id (nothing to join on — check your entry links), and
deliveries that carry no participant id (older deliveries predating that field).

Caveat worth knowing: a webhook endpoint is org-level, and external participant
ids are chosen by you rather than issued by Deutero. If the same id is reused
across studies on the same endpoint, a delivery for another study can mark this
study's completion as notified. Use ids that are unique per person across the
org, or a dedicated endpoint per study, if you need this to be exact.

--csv writes only the never-delivered rows (one per completion), ready for a
payout retry; the summary counts are in --json.

Use this command after a payout or reward run. Do NOT use this command to inspect
one endpoint's raw delivery log; 'webhooks deliveries' does that. Do NOT use it
for interview-time flow signals; use 'flow effects' instead.`,
		Example: "  deutero-pp-cli reconcile --study f0ed9214-06fc-4420-87b3-31240b670fc1 --webhook 9ca6a913-54ff-495d-92f0-2baaa24ee9c6 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "reconcile")
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			if err := deuteroCheckLimit(flagLimit); err != nil {
				return err
			}
			if strings.TrimSpace(flagWebhook) == "" {
				return usageErr(fmt.Errorf("--webhook is required\nUsage: %s --study <id> --webhook <webhook-id>", cmd.CommandPath()))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			progress := deuteroProgressWriter(cmd, flags)

			interviews, err := deuteroListInterviews(ctx, c, flagStudy, cohortReal, true, flagLimit, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}

			path := replacePathParam("/api/v1/webhooks/{webhook_id}/deliveries", "webhook_id", flagWebhook)
			var deliveries []deuteroDelivery
			const page = 200
			// Bound the delivery side explicitly. A webhook endpoint is
			// org-level and its log can be very large; without a ceiling a
			// server that ignores offset returns a full page forever while the
			// slice grows without limit.
			deliveryCap := flagLimit * 10
			if deliveryCap <= 0 {
				deliveryCap = deuteroDefaultInterviewCap * 10
			}
			for offset := 0; len(deliveries) < deliveryCap; offset += page {
				params := map[string]string{
					"limit":  fmt.Sprintf("%d", page),
					"offset": fmt.Sprintf("%d", offset),
				}
				if flagEvent != "" {
					params["event_type"] = flagEvent
				}
				raw, err := c.Get(ctx, path, params)
				if err != nil {
					return classifyAPIError(cmd.OutOrStdout(), err, flags)
				}
				var batch deuteroDeliveryList
				if err := json.Unmarshal(raw, &batch); err != nil {
					return fmt.Errorf("decoding deliveries: %w", err)
				}
				if len(batch.Deliveries) == 0 {
					break
				}
				deliveries = append(deliveries, batch.Deliveries...)
				if progress != nil {
					fmt.Fprintf(progress, "\rread %d deliveries", len(deliveries))
				}
				effective := batch.Limit
				if effective <= 0 {
					effective = page
				}
				if len(batch.Deliveries) < effective || len(deliveries) >= batch.Total {
					break
				}
			}

			report := buildReconcile(flagStudy, flagWebhook, interviews, deliveries)
			if flags.csv {
				return writeReconcileCSV(cmd.OutOrStdout(), report)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderReconcile(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID whose completions to reconcile (required)")
	cmd.Flags().StringVar(&flagWebhook, "webhook", "", "Webhook endpoint ID whose deliveries to join against (required)")
	cmd.Flags().StringVar(&flagEvent, "event-type", "interview.completed", "Delivery event type to join on")
	cmd.Flags().IntVar(&flagLimit, "limit", deuteroDefaultInterviewCap, "Maximum completed interviews to read (1-10000)")
	return cmd
}

// buildReconcile is pure so it can be tested without a network.
func buildReconcile(
	studyID, webhookID string,
	interviews []deuteroInterviewSummary,
	deliveries []deuteroDelivery,
) reconcileReport {
	type deliveryAgg struct {
		attempts   int
		delivered  bool
		lastStatus string
		lastError  string
	}
	byExternal := map[string]*deliveryAgg{}
	unattributed := 0
	for _, d := range deliveries {
		if d.ExternalParticipantID == nil || *d.ExternalParticipantID == "" {
			unattributed++
			continue
		}
		key := *d.ExternalParticipantID
		a := byExternal[key]
		if a == nil {
			a = &deliveryAgg{}
			byExternal[key] = a
		}
		a.attempts++
		if d.Success {
			a.delivered = true
		}
		if d.StatusCode != nil {
			a.lastStatus = fmt.Sprintf("%d", *d.StatusCode)
		}
		if d.ErrorMessage != nil && *d.ErrorMessage != "" {
			a.lastError = *d.ErrorMessage
		}
	}

	report := reconcileReport{
		StudyID:                studyID,
		WebhookID:              webhookID,
		UnattributedDeliveries: unattributed,
		Missing:                []reconcileRow{},
		Note:                   deuteroObservedNote,
	}
	for _, iv := range interviews {
		if !iv.Completed {
			continue
		}
		report.CompletedInterviews++
		ext := deuteroDeref(iv.ExternalParticipantID)
		if ext == "" {
			report.WithoutExternalID++
			continue
		}
		report.WithExternalID++
		a := byExternal[ext]
		if a != nil && a.delivered {
			report.Delivered++
			continue
		}
		report.NeverDelivered++
		row := reconcileRow{
			ExternalParticipantID: ext,
			InterviewID:           iv.ID,
			CompletedAt:           deuteroDeref(iv.EndTime),
		}
		if a != nil {
			row.DeliveryAttempts = a.attempts
			row.LastStatus = a.lastStatus
			row.LastError = a.lastError
		}
		report.Missing = append(report.Missing, row)
	}
	sort.SliceStable(report.Missing, func(i, j int) bool {
		return report.Missing[i].ExternalParticipantID < report.Missing[j].ExternalParticipantID
	})
	return report
}

func renderReconcile(w io.Writer, r reconcileReport) {
	fmt.Fprintf(w, "Reconciliation for study %s against webhook %s\n", r.StudyID, r.WebhookID)
	fmt.Fprintf(w, "  %d completed   %d delivered   %d never delivered\n",
		r.CompletedInterviews, r.Delivered, r.NeverDelivered)

	if r.WithoutExternalID > 0 {
		fmt.Fprintf(w, "\n  %d completions carry no participant id of yours, so there is nothing to join on.\n", r.WithoutExternalID)
		fmt.Fprintln(w, "  Add ?participant_id= to your entry links (or the embed metadata bag) to make these reconcilable.")
	}
	if r.UnattributedDeliveries > 0 {
		fmt.Fprintf(w, "\n  %d deliveries carry no participant id — these predate that field being recorded,\n", r.UnattributedDeliveries)
		fmt.Fprintln(w, "  so they cannot confirm or deny a notification either way.")
	}

	if len(r.Missing) == 0 {
		if r.WithExternalID > 0 {
			fmt.Fprintln(w, "\n  Every reconcilable completion was successfully delivered.")
		}
		fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
		return
	}
	fmt.Fprintln(w, "\n  Never successfully delivered:")
	fmt.Fprintf(w, "  %-28s %-12s %8s %s\n", "participant", "interview", "attempts", "last error")
	for _, m := range r.Missing {
		fmt.Fprintf(w, "  %-28s %-12s %8d %s\n",
			truncateCell(m.ExternalParticipantID, 28),
			shortID(m.InterviewID),
			m.DeliveryAttempts,
			truncateCell(m.LastError, 40))
	}
	fmt.Fprintf(w, "\n%s\n", deuteroObservedNote)
}

// writeReconcileCSV emits the actionable rectangle — completions never
// successfully delivered — one row each. The summary counts stay in --json.
func writeReconcileCSV(w io.Writer, r reconcileReport) error {
	rows := make([][]string, 0, len(r.Missing))
	for _, m := range r.Missing {
		rows = append(rows, []string{
			m.ExternalParticipantID,
			m.InterviewID,
			m.CompletedAt,
			strconv.Itoa(m.DeliveryAttempts),
			m.LastStatus,
			m.LastError,
		})
	}
	return deuteroWriteCSV(w, []string{"external_participant_id", "interview_id", "completed_at", "delivery_attempts", "last_status", "last_error"}, rows)
}
