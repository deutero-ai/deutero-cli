// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

// Behaviour tests for the hand-written Deutero analysis functions. Each builder
// is pure, so these run without a network and assert on content rather than on
// "it did not crash".

package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func sp(s string) *string { return &s }
func ip(i int) *int       { return &i }

// --------------------------------------------------------------------------
// flow coverage
// --------------------------------------------------------------------------

func TestFlowCoverageCountsArmsAndDefaults(t *testing.T) {
	trs := []deuteroTranscript{
		{
			InterviewID: "i1",
			Decisions: []deuteroDecision{
				{NodeID: "triage", Mode: sp("llm_classifier"), Classes: []string{"churned", "active", "dormant"}, MatchedClass: sp("churned")},
			},
			Messages:  []deuteroMessage{{NodeID: sp("triage")}, {NodeID: sp("followup")}},
			Variables: []deuteroVariable{{Name: "reason", Value: json.RawMessage(`"price"`), ValueType: sp("string")}},
		},
		{
			InterviewID: "i2",
			Decisions: []deuteroDecision{
				{NodeID: "triage", Mode: sp("llm_classifier"), Classes: []string{"churned", "active", "dormant"}, UsedDefault: true},
			},
			Messages: []deuteroMessage{{NodeID: sp("triage")}},
		},
	}
	got := buildFlowCoverage("s1", 2, 0, trs)

	if len(got.Branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(got.Branches))
	}
	b := got.Branches[0]
	if b.Traversals != 2 {
		t.Errorf("traversals = %d, want 2", b.Traversals)
	}
	if b.UsedDefault != 1 || b.UsedDefaultRate != 0.5 {
		t.Errorf("used_default = %d rate %.2f, want 1 and 0.50", b.UsedDefault, b.UsedDefaultRate)
	}
	// "active" and "dormant" were offered and never taken — the whole point of
	// the command is naming those.
	if len(b.OfferedNotTaken) != 2 {
		t.Errorf("offered_not_taken = %v, want the 2 arms that never fired", b.OfferedNotTaken)
	}
	// Step reach must be per-interview, not per-message. Index first so an
	// empty result fails loudly instead of skipping the loop body.
	reach := map[string]int{}
	for _, s := range got.Steps {
		reach[s.Step] = s.Interviews
	}
	if len(reach) != 2 {
		t.Fatalf("expected 2 steps, got %v", reach)
	}
	if reach["triage"] != 2 {
		t.Errorf("triage reach = %d, want 2", reach["triage"])
	}
	if reach["followup"] != 1 {
		t.Errorf("followup reach = %d, want 1", reach["followup"])
	}
	if len(got.Variables) != 1 || got.Variables[0].CaptureRate != 0.5 {
		t.Errorf("variable capture = %+v, want reason at 0.5", got.Variables)
	}
}

func TestFlowCoverageEmptyIsNotAnError(t *testing.T) {
	got := buildFlowCoverage("s1", 0, 0, nil)
	if len(got.Branches) != 0 || len(got.Steps) != 0 {
		t.Fatalf("expected empty report, got %+v", got)
	}
	if got.Note == "" {
		t.Error("empty report must still carry the not-observed note")
	}
}

// --------------------------------------------------------------------------
// answers matrix — the both-modes requirement
// --------------------------------------------------------------------------

func TestAnswersMatrixUsesMessagesForLinearStudies(t *testing.T) {
	// A linear study: no captured variables at all (the spec says variables are
	// empty for every linear study). A variables-only implementation would
	// return an empty table here.
	trs := []deuteroBulkTranscript{{
		InterviewID: "i1",
		Messages: []deuteroMessage{
			{Type: sp("assistant"), Content: sp("How did you hear about us?"), QuestionNumber: ip(1)},
			{Type: sp("user"), Content: sp("a friend"), QuestionNumber: ip(1)},
			{Type: sp("user"), Content: sp("weekly"), QuestionNumber: ip(2)},
		},
	}}
	got := buildAnswersMatrix("s1", trs)
	if len(got.Columns) != 2 {
		t.Fatalf("columns = %v, want q1 and q2", got.Columns)
	}
	if got.Rows[0].Cells["q1"] != "a friend" {
		t.Errorf("q1 = %q, want the participant turn not the question text", got.Rows[0].Cells["q1"])
	}
	if got.Source != "message stream (linear study)" {
		t.Errorf("source = %q, want the linear-study source", got.Source)
	}
}

func TestAnswersMatrixUsesVariablesForFlowStudies(t *testing.T) {
	trs := []deuteroBulkTranscript{{
		InterviewID:           "i1",
		ExternalParticipantID: sp("panel-42"),
		Variables: []deuteroVariable{
			{Name: "plan", Value: json.RawMessage(`"pro"`)},
			{Name: "seats", Value: json.RawMessage(`12`)},
		},
	}}
	got := buildAnswersMatrix("s1", trs)
	// Variables are namespaced so a variable named "q3" cannot collide with
	// linear question 3.
	if got.Rows[0].Cells["var:plan"] != "pro" || got.Rows[0].Cells["var:seats"] != "12" {
		t.Errorf("cells = %+v, want var:plan=pro var:seats=12", got.Rows[0].Cells)
	}
	if got.Rows[0].ExternalParticipantID != "panel-42" {
		t.Error("external participant id must survive into the rectangle — it is the join key")
	}
}

func TestAnswerColumnsSortNumerically(t *testing.T) {
	cols := sortAnswerColumns(map[string]bool{"q10": true, "q2": true, "zeta": true, "alpha": true})
	want := []string{"q2", "q10", "alpha", "zeta"}
	if len(cols) != len(want) {
		t.Fatalf("columns = %v, want %v", cols, want)
	}
	for i := range want {
		if cols[i] != want[i] {
			t.Fatalf("columns = %v, want %v", cols, want)
		}
	}
}

// --------------------------------------------------------------------------
// crosstab
// --------------------------------------------------------------------------

func TestCrosstabSegmentsAndReconcilesMarginals(t *testing.T) {
	details := []deuteroInterviewDetail{
		{ID: "i1", Completed: true,
			Characteristics: []deuteroCharacteristic{{Variable: "role", Value: json.RawMessage(`"pm"`)}},
			Variables:       []deuteroVariable{{Name: "nps", Value: json.RawMessage(`9`)}}},
		{ID: "i2", Completed: false,
			Characteristics: []deuteroCharacteristic{{Variable: "role", Value: json.RawMessage(`"pm"`)}},
			Variables:       []deuteroVariable{{Name: "nps", Value: json.RawMessage(`9`)}}},
		{ID: "i3", Completed: true,
			Characteristics: []deuteroCharacteristic{{Variable: "role", Value: json.RawMessage(`"eng"`)}},
			Variables:       []deuteroVariable{{Name: "nps", Value: json.RawMessage(`3`)}}},
		{ID: "i4", Completed: true}, // no characteristic — must not be silently bucketed
	}
	got := buildCrosstab("s1", "role", "nps", details, nil)

	if got.Unsegmented != 1 {
		t.Errorf("unsegmented = %d, want 1", got.Unsegmented)
	}
	if len(got.Segments) != 2 {
		t.Fatalf("expected 2 segments (pm, eng), got %d", len(got.Segments))
	}
	if len(got.AnswerTotals) == 0 {
		t.Fatal("no answer totals — an empty table would pass every check below")
	}
	// Column marginals must equal the per-segment sums, or the table does not
	// reconcile against the server's own tallies.
	for answer, total := range got.AnswerTotals {
		sum := 0
		for _, s := range got.Segments {
			sum += s.Answers[answer]
		}
		if sum != total {
			t.Errorf("marginal for %q = %d but segments sum to %d", answer, total, sum)
		}
	}
	for _, s := range got.Segments {
		if s.Segment == "pm" && s.CompletionRate != 0.5 {
			t.Errorf("pm completion rate = %.2f, want 0.50", s.CompletionRate)
		}
	}
}

func TestParseQuestionRef(t *testing.T) {
	for _, tc := range []struct {
		in    string
		n     int
		isNum bool
	}{{"7", 7, true}, {"q7", 7, true}, {"Q7", 7, true}, {"nps", 0, false}, {"", 0, false}} {
		n, ok := parseQuestionRef(tc.in)
		if n != tc.n || ok != tc.isNum {
			t.Errorf("parseQuestionRef(%q) = %d,%t want %d,%t", tc.in, n, ok, tc.n, tc.isNum)
		}
	}
}

// --------------------------------------------------------------------------
// fielding
// --------------------------------------------------------------------------

func TestFieldingProjectsFillDate(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	var ivs []deuteroInterviewSummary
	// 10 completions over the last 10 days = 1/day against a window of 10.
	for i := 0; i < 10; i++ {
		d := now.AddDate(0, 0, -i).Format("2006-01-02T15:04:05Z")
		ivs = append(ivs, deuteroInterviewSummary{ID: "i", StartTime: sp(d), EndTime: sp(d), Completed: true})
	}
	stats := &deuteroStudyStats{CompletedInterviews: 10, MaxResponses: ip(20), QuotaRemaining: ip(10)}
	got := buildFielding("s1", ivs, stats, 10, now)

	if got.CompletionsPerDay != 1 {
		t.Fatalf("rate = %.2f, want 1.00", got.CompletionsPerDay)
	}
	if got.ProjectedFill != "2026-10-01" {
		t.Errorf("projected fill = %q, want 2026-10-01 (10 remaining at 1/day)", got.ProjectedFill)
	}
}

func TestFieldingSaysSoWhenNothingIsHappening(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	stats := &deuteroStudyStats{CompletedInterviews: 3, MaxResponses: ip(50), QuotaRemaining: ip(47)}
	got := buildFielding("s1", nil, stats, 14, now)
	if got.ProjectedFill != "" {
		t.Errorf("projected fill = %q, want none when the rate is zero", got.ProjectedFill)
	}
	if got.Verdict == "" {
		t.Error("a stalled study must still get an explicit verdict")
	}
}

// --------------------------------------------------------------------------
// saturation
// --------------------------------------------------------------------------

func TestSaturationCountsOnlyFirstAppearances(t *testing.T) {
	mk := func(id, when, text string) deuteroBulkTranscript {
		return deuteroBulkTranscript{
			InterviewID: id,
			StartTime:   sp(when),
			Messages:    []deuteroMessage{{Type: sp("user"), Content: sp(text)}},
		}
	}
	trs := []deuteroBulkTranscript{
		mk("i2", "2026-09-02T00:00:00Z", "delivery windows delivery windows"),
		mk("i1", "2026-09-01T00:00:00Z", "packaging waste recycling"),
	}
	got := buildSaturation("s1", trs, 0.05)

	if len(got.Curve) != 2 {
		t.Fatalf("curve length = %d, want 2", len(got.Curve))
	}
	// Chronological order, not input order.
	if got.Curve[0].InterviewID != "i1" {
		t.Errorf("first point = %s, want the earliest interview i1", got.Curve[0].InterviewID)
	}
	if got.Curve[0].NewTerms != 3 {
		t.Errorf("first interview new terms = %d, want 3", got.Curve[0].NewTerms)
	}
	if got.Curve[1].NewTerms != 2 {
		t.Errorf("second interview new terms = %d, want 2 (repeats within one interview count once)", got.Curve[1].NewTerms)
	}
}

func TestSaturationRepeatContributesNothingNew(t *testing.T) {
	same := "delivery windows packaging"
	trs := []deuteroBulkTranscript{
		{InterviewID: "i1", StartTime: sp("2026-09-01T00:00:00Z"), Messages: []deuteroMessage{{Type: sp("user"), Content: sp(same)}}},
		{InterviewID: "i2", StartTime: sp("2026-09-02T00:00:00Z"), Messages: []deuteroMessage{{Type: sp("user"), Content: sp(same)}}},
	}
	got := buildSaturation("s1", trs, 0.05)
	if got.Curve[1].NewTerms != 0 {
		t.Errorf("identical second interview contributed %d new terms, want 0", got.Curve[1].NewTerms)
	}
}

// --------------------------------------------------------------------------
// reconcile — the anti-join the bug fix unblocked
// --------------------------------------------------------------------------

func TestReconcileFindsUndeliveredCompletions(t *testing.T) {
	ivs := []deuteroInterviewSummary{
		{ID: "i1", Completed: true, ExternalParticipantID: sp("p1")},
		{ID: "i2", Completed: true, ExternalParticipantID: sp("p2")},
		{ID: "i3", Completed: true}, // no external id — nothing to join on
		{ID: "i4", Completed: false, ExternalParticipantID: sp("p4")},
	}
	dels := []deuteroDelivery{
		{EventType: "interview.completed", Success: true, ExternalParticipantID: sp("p1")},
		{EventType: "interview.completed", Success: false, StatusCode: ip(500), ErrorMessage: sp("HTTP 500"), ExternalParticipantID: sp("p2")},
		{EventType: "interview.completed", Success: true}, // predates the field
	}
	got := buildReconcile("s1", "w1", ivs, dels)

	if got.CompletedInterviews != 3 {
		t.Errorf("completed = %d, want 3 (the incomplete one is excluded)", got.CompletedInterviews)
	}
	if got.Delivered != 1 || got.NeverDelivered != 1 {
		t.Errorf("delivered=%d never=%d, want 1 and 1", got.Delivered, got.NeverDelivered)
	}
	if got.WithoutExternalID != 1 {
		t.Errorf("without external id = %d, want 1", got.WithoutExternalID)
	}
	if got.UnattributedDeliveries != 1 {
		t.Errorf("unattributed deliveries = %d, want 1 — these must not be confused with failures", got.UnattributedDeliveries)
	}
	if len(got.Missing) != 1 || got.Missing[0].ExternalParticipantID != "p2" {
		t.Fatalf("missing = %+v, want exactly p2", got.Missing)
	}
	if got.Missing[0].LastError == "" {
		t.Error("a failed delivery must carry its error so the operator can act")
	}
}

// --------------------------------------------------------------------------
// study diff
// --------------------------------------------------------------------------

func TestStudyDiffIgnoresBookkeepingButCatchesContent(t *testing.T) {
	left := &deuteroStudyBundle{
		Study:   json.RawMessage(`{"id":"aaa","name":"Churn","created_at":"2026-01-01"}`),
		Welcome: json.RawMessage(`{"title":"Welcome","body":"Hello"}`),
	}
	right := &deuteroStudyBundle{
		Study:   json.RawMessage(`{"id":"bbb","name":"Churn","created_at":"2026-06-01"}`),
		Welcome: json.RawMessage(`{"title":"Welcome","body":"Hi there"}`),
	}
	got := buildStudyDiff(left, right, "a", "b")

	if got.Same {
		t.Fatal("bodies differ, diff must not report same")
	}
	for _, e := range got.Entries {
		if e.Path == "study.id" || e.Path == "study.created_at" {
			t.Errorf("bookkeeping field %q must be ignored", e.Path)
		}
	}
	found := false
	for _, e := range got.Entries {
		if e.Path == "welcome.body" && e.Kind == "changed" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected welcome.body change, entries = %+v", got.Entries)
	}
}

func TestStudyDiffIdenticalIsSame(t *testing.T) {
	b := func() *deuteroStudyBundle {
		return &deuteroStudyBundle{Welcome: json.RawMessage(`{"title":"W"}`)}
	}
	got := buildStudyDiff(b(), b(), "a", "b")
	if !got.Same || len(got.Entries) != 0 {
		t.Errorf("identical bundles diffed: %+v", got.Entries)
	}
}

// --------------------------------------------------------------------------
// rehearse diff
// --------------------------------------------------------------------------

func TestRehearseDiffComparesCohortsAsShares(t *testing.T) {
	simIVs := []deuteroInterviewSummary{{ID: "s1", Completed: true}, {ID: "s2", Completed: true}}
	realIVs := []deuteroInterviewSummary{{ID: "r1", Completed: true}, {ID: "r2", Completed: false}}
	simTR := []deuteroTranscript{
		{InterviewID: "s1", Messages: []deuteroMessage{{NodeID: sp("intro")}}},
		{InterviewID: "s2", Messages: []deuteroMessage{{NodeID: sp("intro")}}},
	}
	realTR := []deuteroTranscript{
		{InterviewID: "r1", Messages: []deuteroMessage{{NodeID: sp("intro")}}},
		{InterviewID: "r2", Messages: []deuteroMessage{{NodeID: sp("other")}}},
	}
	got := buildRehearseDiff("s1", simIVs, realIVs, simTR, realTR)

	if got.Simulated.CompletionRate != 1.0 || got.Real.CompletionRate != 0.5 {
		t.Errorf("completion rates sim=%.2f real=%.2f, want 1.00 and 0.50", got.Simulated.CompletionRate, got.Real.CompletionRate)
	}
	var introDelta float64
	for _, row := range got.StepReach {
		if row.Measure == "intro" {
			introDelta = row.Delta
		}
	}
	if introDelta != -0.5 {
		t.Errorf("intro reach delta = %.2f, want -0.50 (100%% of personas, 50%% of people)", introDelta)
	}
}

// --------------------------------------------------------------------------
// presentation guards
// --------------------------------------------------------------------------

func TestCompiledNodeIDsAreFlaggedNotPresentedAsAuthored(t *testing.T) {
	// Reporting a compiled id as though it were an authored step name sends a
	// researcher looking for something that appears nowhere in their flow.
	if got := deuteroStepLabel("n3"); got == "n3" {
		t.Errorf("compiled id rendered bare as %q; it must be marked", got)
	}
	if got := deuteroStepLabel("triage"); got != "triage" {
		t.Errorf("authored id was mangled: %q", got)
	}
	if got := deuteroStepLabel("notes"); got != "notes" {
		t.Errorf("authored id beginning with n must not be treated as compiled: %q", got)
	}
}

func TestValueStringRendersEachJSONShape(t *testing.T) {
	cases := map[string]string{
		`"pro"`:     "pro",
		`12`:        "12",
		`true`:      "true",
		`["a","b"]`: "a; b",
		`1.5`:       "1.5",
	}
	for in, want := range cases {
		if got := deuteroValueString(json.RawMessage(in)); got != want {
			t.Errorf("deuteroValueString(%s) = %q, want %q", in, got, want)
		}
	}
}

// --------------------------------------------------------------------------
// flow effects — the dry-run split the file's own comment calls easy to get wrong
// --------------------------------------------------------------------------

func TestFlowEffectsSeparatesDryRunFromDelivered(t *testing.T) {
	effects := []deuteroInterviewEffects{{
		InterviewID: "i1",
		Fetches: []deuteroNodeFetch{
			{NodeID: "enrich", OK: true, StatusCode: ip(200)},
			{NodeID: "enrich", OK: false, StatusCode: ip(502), Error: sp("bad gateway")},
		},
		Signals: []deuteroSignalDelivery{
			{NodeID: "notify", Success: true, StatusCode: ip(200)},
			{NodeID: "notify", Success: false, StatusCode: ip(500), Error: sp("boom")},
			{NodeID: "notify", DryRun: true, Success: true},
			{NodeID: "notify", Success: true, Attempt: ip(3)},
		},
	}}
	got := buildFlowEffects("s1", 1, 0, effects)

	if len(got.Fetches) != 1 || got.Fetches[0].Total != 2 || got.Fetches[0].Failed != 1 {
		t.Fatalf("fetches = %+v, want 1 step with 2 total / 1 failed", got.Fetches)
	}
	if got.Fetches[0].SampleError == "" {
		t.Error("a failed fetch must carry a sample error so the user can act")
	}
	if len(got.Signals) != 1 {
		t.Fatalf("signals = %+v, want one step", got.Signals)
	}
	sig := got.Signals[0]
	// A dry-run delivery is a simulation that did NOT send. Counting it as
	// delivered would report an integration healthy on rehearsals alone.
	if sig.DryRun != 1 {
		t.Errorf("dry_run = %d, want 1", sig.DryRun)
	}
	if sig.Delivered != 2 {
		t.Errorf("delivered = %d, want 2 (the dry-run one must not count)", sig.Delivered)
	}
	if sig.Failed != 1 {
		t.Errorf("failed = %d, want 1", sig.Failed)
	}
	if sig.Retried != 1 {
		t.Errorf("retried = %d, want 1 (attempt > 1)", sig.Retried)
	}
}

// --------------------------------------------------------------------------
// crosstab — the linear (question-number) path, previously untested
// --------------------------------------------------------------------------

func TestCrosstabLinearQuestionUsesMessageStream(t *testing.T) {
	details := []deuteroInterviewDetail{
		{ID: "i1", Completed: true, Characteristics: []deuteroCharacteristic{{Variable: "role", Value: json.RawMessage(`"pm"`)}}},
		{ID: "i2", Completed: true, Characteristics: []deuteroCharacteristic{{Variable: "role", Value: json.RawMessage(`"eng"`)}}},
	}
	trs := []deuteroBulkTranscript{
		{InterviewID: "i1", Messages: []deuteroMessage{
			{Type: sp("assistant"), Content: sp("How often?"), QuestionNumber: ip(7)},
			{Type: sp("user"), Content: sp("weekly"), QuestionNumber: ip(7)},
		}},
		{InterviewID: "i2", Messages: []deuteroMessage{
			{Type: sp("user"), Content: sp("daily"), QuestionNumber: ip(7)},
		}},
	}
	got := buildCrosstab("s1", "role", "q7", details, trs)

	if len(got.Segments) != 2 {
		t.Fatalf("segments = %+v, want pm and eng", got.Segments)
	}
	if got.AnswerTotals["weekly"] != 1 || got.AnswerTotals["daily"] != 1 {
		t.Errorf("answer totals = %v, want weekly=1 daily=1 from the message stream", got.AnswerTotals)
	}
}

// --------------------------------------------------------------------------
// answers matrix — the mixed-source case
// --------------------------------------------------------------------------

func TestAnswersMatrixReportsBothSourcesWhenPresent(t *testing.T) {
	trs := []deuteroBulkTranscript{{
		InterviewID: "i1",
		Variables:   []deuteroVariable{{Name: "q3", Value: json.RawMessage(`"from-variable"`)}},
		Messages: []deuteroMessage{
			{Type: sp("user"), Content: sp("from-message"), QuestionNumber: ip(3)},
		},
	}}
	got := buildAnswersMatrix("s1", trs)
	// The collision case: a variable literally named q3 alongside question 3.
	// Both must survive, under distinct columns.
	if got.Rows[0].Cells["var:q3"] != "from-variable" {
		t.Errorf("var:q3 = %q, want from-variable", got.Rows[0].Cells["var:q3"])
	}
	if got.Rows[0].Cells["q3"] != "from-message" {
		t.Errorf("q3 = %q, want from-message — the variable must not suppress it", got.Rows[0].Cells["q3"])
	}
	if got.Source != "captured variables + message stream" {
		t.Errorf("source = %q, want the mixed-source label", got.Source)
	}
}

func TestAnswersMatrixJoinsMultipleTurnsPerQuestion(t *testing.T) {
	trs := []deuteroBulkTranscript{{
		InterviewID: "i1",
		Messages: []deuteroMessage{
			{Type: sp("user"), Content: sp("first"), QuestionNumber: ip(1)},
			{Type: sp("user"), Content: sp("and the probe answer"), QuestionNumber: ip(1)},
		},
	}}
	got := buildAnswersMatrix("s1", trs)
	cell := got.Rows[0].Cells["q1"]
	if !strings.Contains(cell, "first") || !strings.Contains(cell, "probe") {
		t.Errorf("q1 = %q, want both turns joined — the follow-up is often the substance", cell)
	}
}

// --------------------------------------------------------------------------
// fielding — the window denominator the review flagged
// --------------------------------------------------------------------------

func TestFieldingRateUsesElapsedSpanNotWindowFlag(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	var ivs []deuteroInterviewSummary
	// 30 completions across the last 3 days, asked for with a 14-day window.
	for i := 0; i < 30; i++ {
		d := now.AddDate(0, 0, -(i % 3)).Format("2006-01-02T15:04:05Z")
		ivs = append(ivs, deuteroInterviewSummary{ID: "i", StartTime: sp(d), EndTime: sp(d), Completed: true})
	}
	stats := &deuteroStudyStats{CompletedInterviews: 30, MaxResponses: ip(60), QuotaRemaining: ip(30)}
	got := buildFielding("s1", ivs, stats, 14, now)

	// Dividing by the flag (14) would report ~2.1/day and stretch the fill date
	// roughly fivefold. The study is actually running at 10/day.
	if got.CompletionsPerDay < 9 || got.CompletionsPerDay > 11 {
		t.Errorf("rate = %.2f, want about 10/day over the 3 elapsed days", got.CompletionsPerDay)
	}
}

// --------------------------------------------------------------------------
// study diff — unreadable sections must not read as "identical"
// --------------------------------------------------------------------------

func TestStudyDiffDoesNotCallPartialReadsIdentical(t *testing.T) {
	left := &deuteroStudyBundle{Welcome: json.RawMessage(`{"title":"W"}`)}
	right := &deuteroStudyBundle{
		Welcome: json.RawMessage(`{"title":"W"}`),
		Partial: []string{"graph", "questions"},
	}
	got := buildStudyDiff(left, right, "a", "b")
	if got.Same {
		t.Error("sections that could not be read must not be reported as identical")
	}
	if got.Comparable {
		t.Error("comparable must be false when a side has unread sections")
	}
	if len(got.Unread) != 2 {
		t.Errorf("unread = %v, want the two unreadable sections named", got.Unread)
	}
}

func TestLooksLikeBundlePath(t *testing.T) {
	// This guard is what stops a mistyped filename becoming a remote fetch that
	// 404s every section and prints "no differences".
	for _, in := range []string{"study.json", "./a/b.yaml", "dir/x.yml", "/tmp/s.json"} {
		if !looksLikeBundlePath(in) {
			t.Errorf("looksLikeBundlePath(%q) = false, want true", in)
		}
	}
	for _, in := range []string{"f0ed9214-06fc-4420-87b3-31240b670fc1", "abc123"} {
		if looksLikeBundlePath(in) {
			t.Errorf("looksLikeBundlePath(%q) = true, want false (it is a study id)", in)
		}
	}
}

// --------------------------------------------------------------------------
// output safety
// --------------------------------------------------------------------------

func TestCSVSafeNeutralisesFormulaInjection(t *testing.T) {
	// Participant free text is written by someone outside the trust boundary,
	// and the documented path is "--csv > answers.csv" opened in Excel.
	for _, in := range []string{"=cmd|'/c calc'!A1", "+1+1", "-1", "@SUM(A1)"} {
		got := deuteroCSVSafe(in)
		if got == in || got[0] != '\'' {
			t.Errorf("deuteroCSVSafe(%q) = %q, want a leading quote", in, got)
		}
	}
	if deuteroCSVSafe("weekly") != "weekly" {
		t.Error("ordinary text must pass through untouched")
	}
}

func TestRedactSecretsHidesFlowStepCredentials(t *testing.T) {
	// A flow's context_fetch/webhook steps carry a free-form headers dict, which
	// is where a researcher puts their own backend's bearer token.
	var v any
	if err := json.Unmarshal([]byte(`{"nodes":[{"headers":{"Authorization":"Bearer sk-live-123","Accept":"application/json"}}]}`), &v); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(deuteroRedactSecrets(v))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "sk-live-123") {
		t.Errorf("credential survived redaction: %s", out)
	}
	if !strings.Contains(string(out), "application/json") {
		t.Errorf("non-secret header was destroyed: %s", out)
	}
}

func TestRehearseDiffArmSharesArePerBranch(t *testing.T) {
	// Two branches, each traversed once per interview. Each arm is 100% of its
	// OWN branch; a global denominator would report 50%.
	tr := []deuteroTranscript{{
		InterviewID: "s1",
		Decisions: []deuteroDecision{
			{NodeID: "triage", MatchedClass: sp("churned")},
			{NodeID: "depth", MatchedClass: sp("deep")},
		},
	}}
	got := buildRehearseDiff("s", []deuteroInterviewSummary{{ID: "s1"}}, nil, tr, nil)
	for _, row := range got.Branches {
		if strings.HasPrefix(row.Measure, "triage") || strings.HasPrefix(row.Measure, "depth") {
			if row.Simulated != 1.0 {
				t.Errorf("%s share = %.2f, want 1.00 (share of its own branch)", row.Measure, row.Simulated)
			}
		}
	}
}

// --------------------------------------------------------------------------
// study apply — read shapes are not write shapes (found by the live run)
// --------------------------------------------------------------------------

func applySection(t *testing.T, name string) func(*deuteroStudyBundle) (json.RawMessage, string) {
	t.Helper()
	for _, s := range deuteroApplySections {
		if s.Name == name {
			return s.Build
		}
	}
	t.Fatalf("no apply section %q", name)
	return nil
}

func TestApplySendsOnlyTheSettingsObjectForScreening(t *testing.T) {
	b := &deuteroStudyBundle{Screening: json.RawMessage(`{"questions":[{"id":"q1"}],"settings":{"enabled":true,"disqualification_message":"Sorry","redirect_url":null,"redirect_url_warning":"x"}}`)}
	body, skip := applySection(t, "screening_settings")(b)
	if skip != "" {
		t.Fatalf("skipped: %s", skip)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if _, has := got["questions"]; has {
		t.Error("questions must not be PUT to the settings endpoint")
	}
	if got["enabled"] != true || got["disqualification_message"] != "Sorry" {
		t.Errorf("body = %v, want the settings fields", got)
	}
	if _, has := got["redirect_url_warning"]; has {
		t.Error("read-only warning field must be dropped")
	}
}

func TestApplyNeverCopiesTheShortURLSlug(t *testing.T) {
	// Slugs are unique across the platform; keeping the source's slug makes
	// every clone fail with 409.
	b := &deuteroStudyBundle{Recruitment: json.RawMessage(`{"short_url_slug":"hermes","max_responses":10,"participation_url":"https://x","completed_interviews":3}`)}
	body, skip := applySection(t, "recruitment")(b)
	if skip != "" {
		t.Fatalf("skipped: %s", skip)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if _, has := got["short_url_slug"]; has {
		t.Error("short_url_slug must not be copied")
	}
	if _, has := got["completed_interviews"]; has {
		t.Error("derived counts must not be sent")
	}
	if got["max_responses"] != float64(10) {
		t.Errorf("max_responses = %v, want 10", got["max_responses"])
	}
}

func TestApplySkipsTheFlowOfALinearStudy(t *testing.T) {
	b := &deuteroStudyBundle{Graph: json.RawMessage(`{"flow":null,"interview_mode":"linear"}`)}
	body, skip := applySection(t, "graph")(b)
	if body != nil || !strings.Contains(skip, "linear") {
		t.Errorf("got body=%s skip=%q, want a linear-study skip rather than a null PUT", body, skip)
	}
}
