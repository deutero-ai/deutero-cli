// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

// Shared plumbing for the hand-written Deutero commands.
//
// Three facts about this API shape everything in this file:
//
//  1. Branch routing (`decisions`) comes back from the per-interview transcript
//     endpoint and NEVER from the bulk one. Anything that asks how a flow
//     behaved across a study is an N+1 fan-out. We pay it once, bounded and
//     visible, rather than hiding hundreds of requests behind a progress-free
//     command.
//
//  2. Captured answers (`variables`) exist ONLY for flows that capture. The spec
//     is explicit: variables are "Empty for a study that captures nothing,
//     including every linear study", and decisions are "Empty for a linear study
//     or a flow with no branching". A linear study's answers live in the message
//     stream (`question_number` + participant content) and nowhere else, so any
//     answer-shaped command must read BOTH sources or it silently returns
//     nothing for half the studies on the platform.
//
//  3. `variables`, `decisions`, fetches and signals are best-effort projections:
//     a write failure is logged upstream and the interview carries on. Absence
//     of a row is therefore NOT proof a step did not run. Everything in this
//     package says "not observed", never "did not happen".
package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/deutero-ai/deutero-cli/internal/client"
)

// deuteroMaxFanout bounds the concurrent per-interview requests. The upstream
// documents no rate limit, so this is politeness plus predictable latency, not
// a limit we were told about.
const deuteroMaxFanout = 6

// deuteroDefaultInterviewCap bounds how many interviews a study-wide command
// will pull transcripts for before asking the caller to raise the ceiling
// explicitly. An unbounded default is how a "quick look" turns into a thousand
// requests.
const deuteroDefaultInterviewCap = 500

// deuteroMaxInterviewCap is the ceiling --limit may be raised to. These commands
// hold the whole sample in memory and fan out per interview, so an accidental
// extra zero should fail fast with a usage error rather than run for an hour.
const deuteroMaxInterviewCap = 10000

// deuteroCheckLimit rejects a --limit outside 1..deuteroMaxInterviewCap. Zero
// and negatives are refused too: they used to fall back to the default
// silently, which reads as "no limit" to anyone who passed --limit 0.
func deuteroCheckLimit(limit int) error {
	if limit < 1 || limit > deuteroMaxInterviewCap {
		return usageErr(fmt.Errorf("--limit must be between 1 and %d (got %d)", deuteroMaxInterviewCap, limit))
	}
	return nil
}

// ---------------------------------------------------------------------------
// API payload shapes. These mirror the wire shape, including fields these
// commands do not yet read — keeping them documents the contract rather than
// implying the unread ones are an oversight.
// ---------------------------------------------------------------------------

type deuteroInterviewSummary struct {
	ID                    string  `json:"id"`
	ParticipantID         *string `json:"participant_id"`
	ParticipantName       *string `json:"participant_name"`
	StartTime             *string `json:"start_time"`
	EndTime               *string `json:"end_time"`
	Completed             bool    `json:"completed"`
	Simulated             bool    `json:"simulated"`
	TerminationReason     *string `json:"termination_reason"`
	WebSource             *string `json:"web_source"`
	ExternalParticipantID *string `json:"external_participant_id"`
	Referrer              *string `json:"referrer"`
}

type deuteroInterviewList struct {
	StudyID    string                    `json:"study_id"`
	Total      int                       `json:"total"`
	Limit      int                       `json:"limit"`
	Offset     int                       `json:"offset"`
	Interviews []deuteroInterviewSummary `json:"interviews"`
}

// deuteroVariable mirrors CapturedVariableOut. `Value` stays raw because the
// upstream type is genuinely polymorphic (string, number, boolean, list,
// object) and `value_type` names which.
type deuteroVariable struct {
	Name       string          `json:"name"`
	Value      json.RawMessage `json:"value"`
	ValueType  *string         `json:"value_type"`
	Source     *string         `json:"source"`
	NodeID     *string         `json:"node_id"`
	NodeVisit  *int            `json:"node_visit"`
	QuestionID *int            `json:"question_id"`
	Producer   *string         `json:"producer"`
	UpdatedAt  *string         `json:"updated_at"`
}

// deuteroDecision mirrors DecisionOutcomeOut.
type deuteroDecision struct {
	NodeID       string   `json:"node_id"`
	NodeVisit    *int     `json:"node_visit"`
	Mode         *string  `json:"mode"`
	Classes      []string `json:"classes"`
	RawLabel     *string  `json:"raw_label"`
	MatchedClass *string  `json:"matched_class"`
	ChosenNodeID *string  `json:"chosen_node_id"`
	UsedDefault  bool     `json:"used_default"`
	Error        *string  `json:"error"`
	DecidedAt    *string  `json:"decided_at"`
}

type deuteroMessage struct {
	ID             *string `json:"id"`
	Type           *string `json:"type"`
	Content        *string `json:"content"`
	QuestionNumber *int    `json:"question_number"`
	NodeID         *string `json:"node_id"`
	Timestamp      *string `json:"timestamp"`
}

// deuteroTranscript mirrors TranscriptOut — the per-interview shape, the only
// one carrying `decisions`.
type deuteroTranscript struct {
	InterviewID           string            `json:"interview_id"`
	StudyID               *string           `json:"study_id"`
	ParticipantID         *string           `json:"participant_id"`
	Completed             bool              `json:"completed"`
	ExternalParticipantID *string           `json:"external_participant_id"`
	WebSource             *string           `json:"web_source"`
	Messages              []deuteroMessage  `json:"messages"`
	Variables             []deuteroVariable `json:"variables"`
	Decisions             []deuteroDecision `json:"decisions"`
}

// deuteroBulkTranscript mirrors InterviewTranscript — the bulk shape. Note the
// absent Decisions field; that absence is the whole reason `flow coverage`
// needs a fan-out.
type deuteroBulkTranscript struct {
	InterviewID           string            `json:"interview_id"`
	ParticipantID         *string           `json:"participant_id"`
	ParticipantName       *string           `json:"participant_name"`
	ExternalParticipantID *string           `json:"external_participant_id"`
	WebSource             *string           `json:"web_source"`
	StartTime             *string           `json:"start_time"`
	Completed             bool              `json:"completed"`
	Simulated             bool              `json:"simulated"`
	Messages              []deuteroMessage  `json:"messages"`
	Variables             []deuteroVariable `json:"variables"`
}

type deuteroBulkTranscripts struct {
	StudyID         string                  `json:"study_id"`
	TotalInterviews int                     `json:"total_interviews"`
	Limit           int                     `json:"limit"`
	Offset          int                     `json:"offset"`
	Transcripts     []deuteroBulkTranscript `json:"transcripts"`
}

type deuteroStudyStats struct {
	StudyID              string   `json:"study_id"`
	TotalInterviews      int      `json:"total_interviews"`
	CompletedInterviews  int      `json:"completed_interviews"`
	IncompleteInterviews int      `json:"incomplete_interviews"`
	CompletionRate       *float64 `json:"completion_rate"`
	MaxResponses         *int     `json:"max_responses"`
	QuotaFillRate        *float64 `json:"quota_fill_rate"`
	QuotaRemaining       *int     `json:"quota_remaining"`
	SimulatedInterviews  int      `json:"simulated_interviews"`
}

type deuteroNodeFetch struct {
	NodeID     string  `json:"node_id"`
	NodeVisit  *int    `json:"node_visit"`
	URL        *string `json:"url"`
	Method     *string `json:"method"`
	StatusCode *int    `json:"status_code"`
	Error      *string `json:"error"`
	OK         bool    `json:"ok"`
	CreatedAt  *string `json:"created_at"`
}

type deuteroSignalDelivery struct {
	NodeID     string  `json:"node_id"`
	NodeVisit  *int    `json:"node_visit"`
	EventType  *string `json:"event_type"`
	URL        *string `json:"url"`
	MsgID      *string `json:"msg_id"`
	StatusCode *int    `json:"status_code"`
	Success    bool    `json:"success"`
	Attempt    *int    `json:"attempt"`
	DryRun     bool    `json:"dry_run"`
	Error      *string `json:"error"`
	CreatedAt  *string `json:"created_at"`
}

type deuteroInterviewEffects struct {
	InterviewID           string                  `json:"interview_id"`
	StudyID               *string                 `json:"study_id"`
	ExternalParticipantID *string                 `json:"external_participant_id"`
	WebSource             *string                 `json:"web_source"`
	Simulated             bool                    `json:"simulated"`
	Fetches               []deuteroNodeFetch      `json:"fetches"`
	Signals               []deuteroSignalDelivery `json:"signals"`
}

// ---------------------------------------------------------------------------
// Fetch helpers
// ---------------------------------------------------------------------------

// deuteroCohort selects which interviews a study-wide command looks at. The
// simulated/real split is exact upstream, which is what makes `rehearse diff`
// honest rather than a guess.
type deuteroCohort string

const (
	cohortAll       deuteroCohort = "all"
	cohortReal      deuteroCohort = "real"
	cohortSimulated deuteroCohort = "simulated"
)

// listInterviews pages `GET /studies/{id}/interviews` until the study is
// exhausted or `cap` is reached. Paging is explicit because the endpoint offers
// only limit/offset — there is no cursor to resume from.
func deuteroListInterviews(
	ctx context.Context,
	c *client.Client,
	studyID string,
	cohort deuteroCohort,
	completedOnly bool,
	limit int,
	progress io.Writer,
) ([]deuteroInterviewSummary, error) {
	if limit <= 0 {
		limit = deuteroDefaultInterviewCap
	}
	path := replacePathParam("/api/v1/studies/{study_id}/interviews", "study_id", studyID)

	var out []deuteroInterviewSummary
	const page = 100
	for offset := 0; len(out) < limit; offset += page {
		params := map[string]string{
			"limit":  fmt.Sprintf("%d", page),
			"offset": fmt.Sprintf("%d", offset),
		}
		switch cohort {
		case cohortReal:
			params["simulated"] = "false"
		case cohortSimulated:
			params["simulated"] = "true"
		}
		if completedOnly {
			params["completed"] = "true"
		}
		raw, err := c.Get(ctx, path, params)
		if err != nil {
			return nil, err
		}
		var batch deuteroInterviewList
		if err := json.Unmarshal(raw, &batch); err != nil {
			return nil, fmt.Errorf("decoding interview list: %w", err)
		}
		out = append(out, batch.Interviews...)
		// Compare against the limit the server echoed, not the one we asked
		// for: an endpoint that caps the page size would otherwise look like a
		// short final page and silently truncate the cohort.
		effective := batch.Limit
		if effective <= 0 {
			effective = page
		}
		if len(batch.Interviews) < effective || len(out) >= batch.Total {
			break
		}
		if progress != nil {
			fmt.Fprintf(progress, "\rfetched %d/%d interviews", len(out), batch.Total)
		}
	}
	if progress != nil && len(out) > 0 {
		fmt.Fprintf(progress, "\rfetched %d interviews%s\n", len(out), strings.Repeat(" ", 20))
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// deuteroFetchTranscripts fans out over per-interview transcripts. This is the
// N+1 the bulk endpoint cannot avoid when `decisions` are needed; it is bounded,
// concurrent and reports progress so the cost is visible rather than implied.
//
// A per-interview failure does not abort the run: one unreadable interview
// should not deny an answer about the other 200. Failures are counted and
// returned so callers can say how complete the picture is.
func deuteroFetchTranscripts(
	ctx context.Context,
	c *client.Client,
	interviews []deuteroInterviewSummary,
	progress io.Writer,
) ([]deuteroTranscript, int, error) {
	if len(interviews) == 0 {
		return nil, 0, nil
	}
	var (
		mu        sync.Mutex
		results   = make([]*deuteroTranscript, len(interviews))
		failed    int
		done      int
		firstErr  error
		cancelled bool
	)

	sem := make(chan struct{}, deuteroMaxFanout)
	var wg sync.WaitGroup
	for i, iv := range interviews {
		select {
		case <-ctx.Done():
			// Stop launching, but do not return here: an early return would read
			// the shared counters unsynchronised and leave in-flight goroutines
			// writing into results and the progress writer after the caller has
			// moved on. Break, wait, then report what completed.
			cancelled = true
		default:
		}
		if cancelled {
			break
		}
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			path := replacePathParam("/api/v1/interviews/{interview_id}/transcript", "interview_id", id)
			raw, err := c.Get(ctx, path, map[string]string{})
			mu.Lock()
			defer mu.Unlock()
			done++
			if progress != nil {
				fmt.Fprintf(progress, "\rreading transcripts %d/%d", done, len(interviews))
			}
			if err != nil {
				failed++
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			var tr deuteroTranscript
			if err := json.Unmarshal(raw, &tr); err != nil {
				failed++
				return
			}
			results[idx] = &tr
		}(i, iv.ID)
	}
	wg.Wait()
	if progress != nil {
		fmt.Fprintf(progress, "\rread %d transcripts%s\n", done-failed, strings.Repeat(" ", 20))
	}

	out := make([]deuteroTranscript, 0, len(interviews))
	for _, r := range results {
		if r != nil {
			out = append(out, *r)
		}
	}
	// One unreadable interview must not deny an answer about the other 200, so
	// partial failure is reported, not fatal. Total failure is different: an
	// empty corpus that looks like "nothing to report" would be a lie.
	if len(out) == 0 && failed > 0 && firstErr != nil {
		return nil, failed, fmt.Errorf("could not read any of %d transcripts: %w", failed, firstErr)
	}
	return out, failed, nil
}

// deuteroFetchEffects fans out over per-interview fetches-and-signals, which is
// likewise per-interview only.
func deuteroFetchEffects(
	ctx context.Context,
	c *client.Client,
	interviews []deuteroInterviewSummary,
	progress io.Writer,
) ([]deuteroInterviewEffects, int, error) {
	if len(interviews) == 0 {
		return nil, 0, nil
	}
	var (
		mu        sync.Mutex
		results   = make([]*deuteroInterviewEffects, len(interviews))
		failed    int
		done      int
		firstErr  error
		cancelled bool
	)
	sem := make(chan struct{}, deuteroMaxFanout)
	var wg sync.WaitGroup
	for i, iv := range interviews {
		select {
		case <-ctx.Done():
			// Stop launching, but do not return here: an early return would read
			// the shared counters unsynchronised and leave in-flight goroutines
			// writing into results and the progress writer after the caller has
			// moved on. Break, wait, then report what completed.
			cancelled = true
		default:
		}
		if cancelled {
			break
		}
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := replacePathParam("/api/v1/interviews/{interview_id}/fetches-and-signals", "interview_id", id)
			raw, err := c.Get(ctx, path, map[string]string{})
			mu.Lock()
			defer mu.Unlock()
			done++
			if progress != nil {
				fmt.Fprintf(progress, "\rreading flow effects %d/%d", done, len(interviews))
			}
			if err != nil {
				failed++
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			var ef deuteroInterviewEffects
			if err := json.Unmarshal(raw, &ef); err != nil {
				failed++
				return
			}
			results[idx] = &ef
		}(i, iv.ID)
	}
	wg.Wait()
	if progress != nil {
		fmt.Fprintf(progress, "\rread %d effect records%s\n", done-failed, strings.Repeat(" ", 20))
	}
	out := make([]deuteroInterviewEffects, 0, len(interviews))
	for _, r := range results {
		if r != nil {
			out = append(out, *r)
		}
	}
	if len(out) == 0 && failed > 0 && firstErr != nil {
		return nil, failed, fmt.Errorf("could not read any of %d effect records: %w", failed, firstErr)
	}
	return out, failed, nil
}

// deuteroBulkFetchTranscripts pages the bulk endpoint. Cheaper than the
// per-interview fan-out and sufficient whenever `decisions` are not needed
// (answers matrix, crosstab, saturation all live here).
func deuteroBulkFetchTranscripts(
	ctx context.Context,
	c *client.Client,
	studyID string,
	includeSimulated bool,
	completedOnly bool,
	limit int,
	progress io.Writer,
) ([]deuteroBulkTranscript, error) {
	if limit <= 0 {
		limit = deuteroDefaultInterviewCap
	}
	path := replacePathParam("/api/v1/studies/{study_id}/transcripts", "study_id", studyID)
	var out []deuteroBulkTranscript
	const page = 50
	for offset := 0; len(out) < limit; offset += page {
		params := map[string]string{
			"limit":             fmt.Sprintf("%d", page),
			"offset":            fmt.Sprintf("%d", offset),
			"include_simulated": fmt.Sprintf("%t", includeSimulated),
		}
		if completedOnly {
			params["completed"] = "true"
		}
		raw, err := c.Get(ctx, path, params)
		if err != nil {
			return nil, err
		}
		var batch deuteroBulkTranscripts
		if err := json.Unmarshal(raw, &batch); err != nil {
			return nil, fmt.Errorf("decoding bulk transcripts: %w", err)
		}
		out = append(out, batch.Transcripts...)
		if progress != nil {
			fmt.Fprintf(progress, "\rfetched %d transcripts", len(out))
		}
		if len(batch.Transcripts) < page || len(out) >= batch.TotalInterviews {
			break
		}
	}
	if progress != nil && len(out) > 0 {
		fmt.Fprintf(progress, "\rfetched %d transcripts%s\n", len(out), strings.Repeat(" ", 20))
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func deuteroGetStudyStats(ctx context.Context, c *client.Client, studyID string) (*deuteroStudyStats, error) {
	path := replacePathParam("/api/v1/studies/{study_id}/stats", "study_id", studyID)
	raw, err := c.Get(ctx, path, map[string]string{})
	if err != nil {
		return nil, err
	}
	var s deuteroStudyStats
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("decoding study stats: %w", err)
	}
	return &s, nil
}

// ---------------------------------------------------------------------------
// Presentation helpers
// ---------------------------------------------------------------------------

// deuteroStepLabel renders a flow step id for a researcher.
//
// The graph compiler names IR nodes after the editor canvas (`n3`) while the
// author's own step id lives in `properties._authoring_id`; interview rows carry
// the compiled id and the API is responsible for translating back. If a raw
// compiled-looking id reaches us anyway, say so rather than printing `n3` as
// though it were something the researcher could find in their flow.
func deuteroStepLabel(nodeID string) string {
	if nodeID == "" {
		return "(unnamed step)"
	}
	if looksLikeCompiledNodeID(nodeID) {
		return nodeID + " (compiled id — not an authored step name)"
	}
	return nodeID
}

// looksLikeCompiledNodeID reports the canvas-generated shape: "n" followed by
// digits only. Authored ids are author-chosen names, so this is a narrow test
// that will not misfire on a step someone happened to call "notes".
func looksLikeCompiledNodeID(id string) bool {
	if len(id) < 2 || id[0] != 'n' {
		return false
	}
	for _, r := range id[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// deuteroObservedNote is the sentence every coverage-shaped command prints.
// The upstream projections are best-effort, so a zero count means we did not
// see it, not that it did not happen.
const deuteroObservedNote = "Counts are of observed records only. Captured answers, decisions, fetches and signals are best-effort projections upstream: a write failure is logged and the interview continues, so a zero here reads as \"not observed\", never as \"did not happen\"."

// deuteroParseTime parses the timestamp shapes this API returns, tolerating
// both offset-bearing and naive forms.
func deuteroParseTime(s *string) (time.Time, bool) {
	if s == nil || *s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, *s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func deuteroDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// deuteroValueString renders a captured variable's polymorphic value for a
// table or CSV cell without lying about its type.
func deuteroValueString(v json.RawMessage) string {
	if len(v) == 0 {
		return ""
	}
	var asString string
	if err := json.Unmarshal(v, &asString); err == nil {
		return asString
	}
	var asAny any
	if err := json.Unmarshal(v, &asAny); err == nil {
		switch t := asAny.(type) {
		case float64:
			if t == float64(int64(t)) {
				return fmt.Sprintf("%d", int64(t))
			}
			return fmt.Sprintf("%g", t)
		case bool:
			return fmt.Sprintf("%t", t)
		case []any:
			parts := make([]string, 0, len(t))
			for _, e := range t {
				parts = append(parts, fmt.Sprintf("%v", e))
			}
			return strings.Join(parts, "; ")
		}
	}
	return strings.TrimSpace(string(v))
}

// deuteroSortedKeys keeps table and CSV column order deterministic, which
// matters for diffing output across runs.
func deuteroSortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// deuteroRequireStudy validates the --study flag the same way across every
// command, so a missing id is a usage error (exit 2) rather than a 404.
func deuteroRequireStudy(cmdPath, studyID string) error {
	if strings.TrimSpace(studyID) == "" {
		return usageErr(fmt.Errorf("--study is required\nUsage: %s --study <study-id>", cmdPath))
	}
	return nil
}

// deuteroProgressWriter returns the stream to narrate a fan-out on, or nil when
// the caller wants machine output. Progress goes to stderr so stdout stays a
// clean JSON/CSV stream, and is suppressed entirely for --quiet and for
// non-terminal stderr so logs do not fill with carriage returns.
func deuteroProgressWriter(cmd *cobra.Command, flags *rootFlags) io.Writer {
	if flags.quiet || flags.agent {
		return nil
	}
	w := cmd.ErrOrStderr()
	if !isTerminal(w) {
		return nil
	}
	return w
}

// deuteroCharacteristic mirrors CharacteristicValueOut: the participant
// attributes a study collects specifically so answers can be segmented by them.
type deuteroCharacteristic struct {
	Variable string          `json:"variable"`
	Value    json.RawMessage `json:"value"`
}

// deuteroInterviewDetail mirrors InterviewDetailOut. It is the only shape that
// carries characteristics, which is why segmentation needs a per-interview
// fan-out rather than the bulk endpoint.
type deuteroInterviewDetail struct {
	ID                    string                  `json:"id"`
	StudyID               *string                 `json:"study_id"`
	Completed             bool                    `json:"completed"`
	Simulated             bool                    `json:"simulated"`
	ExternalParticipantID *string                 `json:"external_participant_id"`
	MessageCount          *int                    `json:"message_count"`
	Characteristics       []deuteroCharacteristic `json:"characteristics"`
	Variables             []deuteroVariable       `json:"variables"`
}

// deuteroFetchInterviewDetails fans out over per-interview detail records. Same
// bounded, visible pattern as the transcript fan-out; a single unreadable
// interview does not deny an answer about the rest.
func deuteroFetchInterviewDetails(
	ctx context.Context,
	c *client.Client,
	interviews []deuteroInterviewSummary,
	progress io.Writer,
) ([]deuteroInterviewDetail, int, error) {
	if len(interviews) == 0 {
		return nil, 0, nil
	}
	var (
		mu        sync.Mutex
		results   = make([]*deuteroInterviewDetail, len(interviews))
		failed    int
		done      int
		firstErr  error
		cancelled bool
	)
	sem := make(chan struct{}, deuteroMaxFanout)
	var wg sync.WaitGroup
	for i, iv := range interviews {
		select {
		case <-ctx.Done():
			// Stop launching, but do not return here: an early return would read
			// the shared counters unsynchronised and leave in-flight goroutines
			// writing into results and the progress writer after the caller has
			// moved on. Break, wait, then report what completed.
			cancelled = true
		default:
		}
		if cancelled {
			break
		}
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := replacePathParam("/api/v1/interviews/{interview_id}", "interview_id", id)
			raw, err := c.Get(ctx, path, map[string]string{})
			mu.Lock()
			defer mu.Unlock()
			done++
			if progress != nil {
				fmt.Fprintf(progress, "\rreading interviews %d/%d", done, len(interviews))
			}
			if err != nil {
				failed++
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			var d deuteroInterviewDetail
			if err := json.Unmarshal(raw, &d); err != nil {
				failed++
				return
			}
			results[idx] = &d
		}(i, iv.ID)
	}
	wg.Wait()
	if progress != nil {
		fmt.Fprintf(progress, "\rread %d interviews%s\n", done-failed, strings.Repeat(" ", 20))
	}
	out := make([]deuteroInterviewDetail, 0, len(interviews))
	for _, r := range results {
		if r != nil {
			out = append(out, *r)
		}
	}
	if len(out) == 0 && failed > 0 && firstErr != nil {
		return nil, failed, fmt.Errorf("could not read any of %d interviews: %w", failed, firstErr)
	}
	return out, failed, nil
}

// deuteroStudyBundle is a whole study definition in one artifact: the ~8 reads
// it takes to describe a study, captured so the definition can live in a repo,
// be diffed, and be replayed. `Graph` and `Questions` are mutually exclusive in
// practice — a study runs one interview mode or the other — but both are
// captured when present rather than guessed at.
type deuteroStudyBundle struct {
	StudyID         string          `json:"study_id"`
	CapturedAt      string          `json:"captured_at"`
	Study           json.RawMessage `json:"study,omitempty"`
	Welcome         json.RawMessage `json:"welcome,omitempty"`
	Translations    json.RawMessage `json:"welcome_translations,omitempty"`
	Screening       json.RawMessage `json:"screening,omitempty"`
	Characteristics json.RawMessage `json:"characteristics,omitempty"`
	Questions       json.RawMessage `json:"questions,omitempty"`
	Graph           json.RawMessage `json:"graph,omitempty"`
	Recruitment     json.RawMessage `json:"recruitment,omitempty"`
	Partial         []string        `json:"unreadable_sections,omitempty"`
}

// deuteroBundleSections is the read plan, in the order a human would describe a
// study. Keeping it a table means bundle and diff can never drift apart on
// which sections make up "a study".
var deuteroBundleSections = []struct {
	Name string
	Path string
	Set  func(*deuteroStudyBundle, json.RawMessage)
}{
	{"study", "/api/v1/studies/{study_id}", func(b *deuteroStudyBundle, v json.RawMessage) { b.Study = v }},
	{"welcome", "/api/v1/studies/{study_id}/welcome", func(b *deuteroStudyBundle, v json.RawMessage) { b.Welcome = v }},
	{"welcome_translations", "/api/v1/studies/{study_id}/welcome/translations", func(b *deuteroStudyBundle, v json.RawMessage) { b.Translations = v }},
	{"screening", "/api/v1/studies/{study_id}/screening", func(b *deuteroStudyBundle, v json.RawMessage) { b.Screening = v }},
	{"characteristics", "/api/v1/studies/{study_id}/characteristics", func(b *deuteroStudyBundle, v json.RawMessage) { b.Characteristics = v }},
	{"questions", "/api/v1/studies/{study_id}/questions", func(b *deuteroStudyBundle, v json.RawMessage) { b.Questions = v }},
	{"graph", "/api/v1/studies/{study_id}/graph", func(b *deuteroStudyBundle, v json.RawMessage) { b.Graph = v }},
	{"recruitment", "/api/v1/studies/{study_id}/recruitment", func(b *deuteroStudyBundle, v json.RawMessage) { b.Recruitment = v }},
}

// deuteroFetchBundle reads every section of a study definition.
//
// A section that 404s is normal, not an error: a linear study has no graph, a
// graph study has no question list, and a study with no consent screen has no
// welcome. Those are recorded as unreadable rather than failing the whole
// export, so a bundle is always producible and always says what it could not
// see.
func deuteroFetchBundle(
	ctx context.Context,
	c *client.Client,
	studyID string,
	progress io.Writer,
) (*deuteroStudyBundle, error) {
	bundle := &deuteroStudyBundle{
		StudyID:    studyID,
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
	}
	// Read the sections concurrently. Eight sequential round trips against a
	// host that cold-starts at ~4s is enough to blow a 10s budget on its own,
	// and the sections are independent reads.
	type sectionResult struct {
		idx int
		raw json.RawMessage
		err error
	}
	results := make([]sectionResult, len(deuteroBundleSections))
	sem := make(chan struct{}, deuteroMaxFanout)
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	for i, section := range deuteroBundleSections {
		wg.Add(1)
		go func(idx int, name, tmpl string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := replacePathParam(tmpl, "study_id", studyID)
			raw, err := c.Get(ctx, path, map[string]string{})
			mu.Lock()
			done++
			if progress != nil {
				fmt.Fprintf(progress, "\rreading study definition %d/%d", done, len(deuteroBundleSections))
			}
			mu.Unlock()
			results[idx] = sectionResult{idx: idx, raw: raw, err: err}
		}(i, section.Name, section.Path)
	}
	wg.Wait()

	for i, section := range deuteroBundleSections {
		r := results[i]
		// A 404 here is normal, not an error: a linear study has no flow, a
		// flow study has no question list, and a study with no consent screen
		// has no welcome. Record what could not be read and carry on, so a
		// bundle is always producible and always says what it missed.
		if r.err != nil {
			bundle.Partial = append(bundle.Partial, section.Name)
			continue
		}
		section.Set(bundle, r.raw)
	}
	if progress != nil {
		fmt.Fprintf(progress, "\rread %d study sections%s\n", len(deuteroBundleSections)-len(bundle.Partial), strings.Repeat(" ", 20))
	}
	return bundle, nil
}

// deuteroSecretKeyRe matches the field names a study definition can legitimately
// carry a credential in. A flow's "Bring in data" (context_fetch) and "Send a
// signal" (webhook) steps hold a free-form headers dict, which is exactly where
// a researcher puts the bearer token for their own backend — so a study
// definition is not automatically safe to print or commit.
var deuteroSecretKeyRe = regexp.MustCompile(`(?i)(authorization|secret|token|api[-_]?key|password|cookie)`)

// deuteroRedactSecrets walks decoded JSON and replaces the value of any
// secret-shaped key with a marker. Applied to whole study definitions before
// they are printed or written, so a bundle can be committed to a repo and a
// diff can be pasted into a ticket without leaking somebody's backend
// credential. Pass --include-secrets to opt out deliberately.
func deuteroRedactSecrets(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if deuteroSecretKeyRe.MatchString(k) {
				if sv, ok := val.(string); ok && sv != "" {
					t[k] = "[redacted — pass --include-secrets to show]"
					continue
				}
			}
			t[k] = deuteroRedactSecrets(val)
		}
		return t
	case []any:
		for i := range t {
			t[i] = deuteroRedactSecrets(t[i])
		}
		return t
	}
	return v
}

// deuteroRedactRaw applies deuteroRedactSecrets to a raw JSON section, leaving
// it untouched when it does not decode.
func deuteroRedactRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw
	}
	out, err := json.Marshal(deuteroRedactSecrets(v))
	if err != nil {
		return raw
	}
	return out
}

// deuteroRedactBundle redacts every section of a study definition in place.
func deuteroRedactBundle(b *deuteroStudyBundle) {
	b.Study = deuteroRedactRaw(b.Study)
	b.Welcome = deuteroRedactRaw(b.Welcome)
	b.Translations = deuteroRedactRaw(b.Translations)
	b.Screening = deuteroRedactRaw(b.Screening)
	b.Characteristics = deuteroRedactRaw(b.Characteristics)
	b.Questions = deuteroRedactRaw(b.Questions)
	b.Graph = deuteroRedactRaw(b.Graph)
	b.Recruitment = deuteroRedactRaw(b.Recruitment)
}

// deuteroCSVSafe neutralises spreadsheet formula injection. Participant answers
// are free text written by someone outside the researcher's trust boundary, and
// the documented happy path for this data is "--csv > answers.csv", opened in
// Excel. A leading =, +, -, @, tab or CR is enough to execute.
func deuteroCSVSafe(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + v
	}
	return v
}

// deuteroRejectCSV refuses --csv on commands whose report is nested rather than
// one rectangle. Without it the generic printer silently falls back to JSON, so
// "--csv > out.csv" produces a file a spreadsheet cannot open and nothing says
// why.
func deuteroRejectCSV(cmd *cobra.Command, flags *rootFlags) error {
	if flags == nil || !flags.csv {
		return nil
	}
	return usageErr(fmt.Errorf("%s does not support --csv: its report is nested, not a single table; use --json (with --select to pick fields) instead", cmd.CommandPath()))
}

// deuteroWriteCSV writes one header and its rows through encoding/csv, passing
// every cell through deuteroCSVSafe. Callers only pass non-negative numbers, so
// the formula-injection guard never alters a numeric cell.
func deuteroWriteCSV(w io.Writer, header []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, r := range rows {
		safe := make([]string, len(r))
		for i, v := range r {
			safe[i] = deuteroCSVSafe(v)
		}
		if err := cw.Write(safe); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
