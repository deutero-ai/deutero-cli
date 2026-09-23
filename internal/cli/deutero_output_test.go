// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func readCSV(t *testing.T, s string) [][]string {
	t.Helper()
	recs, err := csv.NewReader(strings.NewReader(s)).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v\n%s", err, s)
	}
	return recs
}

func TestWriteReconcileCSVListsMissingRowsAndGuardsFormulas(t *testing.T) {
	var buf bytes.Buffer
	r := reconcileReport{Missing: []reconcileRow{
		{ExternalParticipantID: "=HYPERLINK(1)", InterviewID: "iv-1", CompletedAt: "2026-09-01T00:00:00Z", DeliveryAttempts: 3, LastStatus: "500", LastError: "boom, again"},
	}}
	if err := writeReconcileCSV(&buf, r); err != nil {
		t.Fatal(err)
	}
	recs := readCSV(t, buf.String())
	if len(recs) != 2 {
		t.Fatalf("want header + 1 row, got %d", len(recs))
	}
	if recs[0][0] != "external_participant_id" || recs[1][0] != "'=HYPERLINK(1)" {
		t.Fatalf("unexpected cells: %v", recs)
	}
	if recs[1][3] != "3" || recs[1][5] != "boom, again" {
		t.Fatalf("unexpected row: %v", recs[1])
	}
}

func TestWriteFieldingAndSaturationCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := writeFieldingCSV(&buf, fieldingReport{Days: []fieldingDay{{Date: "2026-09-01", Started: 4, Completed: 2, Cumulative: 2}}}); err != nil {
		t.Fatal(err)
	}
	if got := readCSV(t, buf.String()); len(got) != 2 || got[1][0] != "2026-09-01" || got[1][3] != "2" {
		t.Fatalf("fielding csv: %v", got)
	}

	buf.Reset()
	if err := writeSaturationCSV(&buf, saturationReport{Curve: []saturationPoint{{Index: 1, InterviewID: "iv-1", NewTerms: 10, TotalTerms: 10, NewShare: 1}}}); err != nil {
		t.Fatal(err)
	}
	if got := readCSV(t, buf.String()); len(got) != 2 || got[1][4] != "1.0000" {
		t.Fatalf("saturation csv: %v", got)
	}
}

func TestDeuteroRejectCSVIsUsageError(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	if err := deuteroRejectCSV(cmd, &rootFlags{}); err != nil {
		t.Fatalf("no --csv should pass, got %v", err)
	}
	err := deuteroRejectCSV(cmd, &rootFlags{csv: true})
	var ce *cliError
	if !errors.As(err, &ce) || ce.code != 2 {
		t.Fatalf("want usage error (exit 2), got %v", err)
	}
}

func TestDeuteroCheckLimitBounds(t *testing.T) {
	for _, ok := range []int{1, deuteroDefaultInterviewCap, deuteroMaxInterviewCap} {
		if err := deuteroCheckLimit(ok); err != nil {
			t.Errorf("limit %d rejected: %v", ok, err)
		}
	}
	for _, bad := range []int{0, -1, deuteroMaxInterviewCap + 1, 5000000} {
		var ce *cliError
		if err := deuteroCheckLimit(bad); !errors.As(err, &ce) || ce.code != 2 {
			t.Errorf("limit %d: want usage error, got %v", bad, err)
		}
	}
}
