// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// Nothing upstream diffs anything. Each side of this comparison is ~8 calls, so
// the question "what actually differs between these two studies?" only becomes
// answerable once both definitions are local.

type studyDiffEntry struct {
	Path  string `json:"path"`
	Left  string `json:"left,omitempty"`
	Right string `json:"right,omitempty"`
	Kind  string `json:"kind"` // added | removed | changed
}

type studyDiffReport struct {
	Left       string           `json:"left"`
	Right      string           `json:"right"`
	Same       bool             `json:"same"`
	Sections   map[string]int   `json:"changes_by_section"`
	Entries    []studyDiffEntry `json:"entries"`
	Unread     []string         `json:"unread_sections,omitempty"`
	Comparable bool             `json:"comparable"`
}

// looksLikeBundlePath reports whether --against names a file rather than a
// study id: a path separator, or a data-file extension. Study ids are bare
// UUIDs, so this cannot misfire on a real id.
func looksLikeBundlePath(v string) bool {
	if strings.ContainsAny(v, "/\\") {
		return true
	}
	lower := strings.ToLower(v)
	for _, ext := range []string{".json", ".yaml", ".yml"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func newNovelStudyDiffCmd(flags *rootFlags) *cobra.Command {
	var flagStudy, flagAgainst string
	var flagIncludeSecrets bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two study definitions field by field",
		Long: `Compares a study against another study or against a saved bundle file, field
by field across details, welcome and translations, screening, characteristics,
questions or flow, and recruitment.

--against accepts either a study ID or a path to a file written by
'study bundle'. Volatile identity fields (ids, timestamps, counts) are ignored so
the diff shows instrument differences rather than bookkeeping.

Use this command to see what differs before replaying one definition over
another. Do NOT use this command to write changes.`,
		Example: "  deutero-pp-cli study diff --study f0ed9214-06fc-4420-87b3-31240b670fc1 --against 092103ce-0e01-4e81-8ea5-e85fa560504f",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "study diff")
			}
			if err := deuteroRejectCSV(cmd, flags); err != nil {
				return err
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			if strings.TrimSpace(flagAgainst) == "" {
				return usageErr(fmt.Errorf("--against is required (a study ID or a bundle file)\nUsage: %s --study <id> --against <id-or-file>", cmd.CommandPath()))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			progress := deuteroProgressWriter(cmd, flags)

			var right *deuteroStudyBundle
			blob, readErr := os.ReadFile(filepath.Clean(flagAgainst)) // #nosec G304 -- operator-named bundle file from --against; reading it is the command's purpose.

			// When both sides are remote, read them concurrently: two sequential
			// eight-section reads against a cold-starting host is twice the
			// latency for no reason.
			var left *deuteroStudyBundle
			var leftErr, rightErr error
			if readErr != nil && !looksLikeBundlePath(flagAgainst) {
				var wg sync.WaitGroup
				wg.Add(2)
				go func() { defer wg.Done(); left, leftErr = deuteroFetchBundle(ctx, c, flagStudy, progress) }()
				go func() { defer wg.Done(); right, rightErr = deuteroFetchBundle(ctx, c, flagAgainst, nil) }()
				wg.Wait()
				if leftErr != nil {
					return classifyAPIError(cmd.OutOrStdout(), leftErr, flags)
				}
				if rightErr != nil {
					return classifyAPIError(cmd.OutOrStdout(), rightErr, flags)
				}
				if !flagIncludeSecrets {
					deuteroRedactBundle(left)
					deuteroRedactBundle(right)
				}
				report := buildStudyDiff(left, right, flagStudy, flagAgainst)
				if wantsHumanTable(cmd.OutOrStdout(), flags) {
					renderStudyDiff(cmd.OutOrStdout(), report)
					return nil
				}
				return printJSONFiltered(cmd.OutOrStdout(), report, flags)
			}

			left, err = deuteroFetchBundle(ctx, c, flagStudy, progress)
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			switch {
			case readErr == nil:
				right = &deuteroStudyBundle{}
				if err := json.Unmarshal(blob, right); err != nil {
					return fmt.Errorf("parsing bundle file %s: %w", flagAgainst, err)
				}
			case looksLikeBundlePath(flagAgainst):
				// A value that is plainly a filename must not fall through to a
				// remote fetch. Treating a missing or mistyped path as a study id
				// turns a typo into eight slow 404s and an empty diff, which
				// reads as "no differences" — the most misleading answer this
				// command can give.
				return usageErr(fmt.Errorf("bundle file %s not found: %w", flagAgainst, readErr))
			}

			if !flagIncludeSecrets {
				deuteroRedactBundle(left)
				deuteroRedactBundle(right)
			}
			report := buildStudyDiff(left, right, flagStudy, flagAgainst)
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderStudyDiff(cmd.OutOrStdout(), report)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), report, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID on the left of the comparison (required)")
	cmd.Flags().StringVar(&flagAgainst, "against", "", "Study ID or bundle file on the right (required)")
	cmd.Flags().BoolVar(&flagIncludeSecrets, "include-secrets", false, "Compare credential-shaped values instead of redacting them")
	return cmd
}

// deuteroVolatileKeys are bookkeeping fields that always differ between two
// studies and never mean the instrument differs. Diffing them would bury the
// real changes in noise.
var deuteroVolatileKeys = map[string]bool{
	"id": true, "study_id": true, "survey_id": true, "project_id": true,
	"created_at": true, "updated_at": true, "captured_at": true,
	"question_id": true, "translation_id": true, "key_id": true,
	"graph_version": true, "signing_secret": true,
	// Derived per-study state: counts and the links a study is reachable at.
	// Two studies always differ here, and none of it is part of the instrument,
	// so reporting it buries the real differences.
	"completed_interviews": true, "quota_remaining": true, "total_interviews": true,
	"participation_url": true, "short_participation_url": true, "short_url_slug": true,
	"video_participation_url": true, "voice_participation_url": true,
	"short_video_participation_url": true, "short_voice_participation_url": true,
	"redirect_url_warning": true,
}

// buildStudyDiff is pure so it can be tested without a network.
func buildStudyDiff(left, right *deuteroStudyBundle, leftName, rightName string) studyDiffReport {
	report := studyDiffReport{
		Left:     leftName,
		Right:    rightName,
		Sections: map[string]int{},
		Entries:  []studyDiffEntry{},
	}
	sections := []struct {
		name string
		l, r json.RawMessage
	}{
		{"study", left.Study, right.Study},
		{"welcome", left.Welcome, right.Welcome},
		{"welcome_translations", left.Translations, right.Translations},
		{"screening", left.Screening, right.Screening},
		{"characteristics", left.Characteristics, right.Characteristics},
		{"questions", left.Questions, right.Questions},
		{"graph", left.Graph, right.Graph},
		{"recruitment", left.Recruitment, right.Recruitment},
	}
	for _, s := range sections {
		var lv, rv any
		_ = json.Unmarshal(s.l, &lv)
		_ = json.Unmarshal(s.r, &rv)
		before := len(report.Entries)
		report.Entries = diffValues(report.Entries, s.name, lv, rv)
		if n := len(report.Entries) - before; n > 0 {
			report.Sections[s.name] = n
		}
	}
	// Sections that could not be read on either side are not evidence of
	// sameness. Say so rather than letting an unreadable section masquerade as
	// an identical one.
	seen := map[string]bool{}
	for _, name := range append(append([]string{}, left.Partial...), right.Partial...) {
		if !seen[name] {
			seen[name] = true
			report.Unread = append(report.Unread, name)
		}
	}
	report.Comparable = len(report.Unread) == 0
	report.Same = len(report.Entries) == 0 && report.Comparable
	return report
}

func diffValues(acc []studyDiffEntry, path string, l, r any) []studyDiffEntry {
	switch lv := l.(type) {
	case map[string]any:
		rv, ok := r.(map[string]any)
		if !ok {
			return appendDiff(acc, path, l, r)
		}
		keys := map[string]bool{}
		for k := range lv {
			keys[k] = true
		}
		for k := range rv {
			keys[k] = true
		}
		names := make([]string, 0, len(keys))
		for k := range keys {
			if deuteroVolatileKeys[k] {
				continue
			}
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			acc = diffValues(acc, path+"."+k, lv[k], rv[k])
		}
		return acc
	case []any:
		rv, ok := r.([]any)
		if !ok {
			return appendDiff(acc, path, l, r)
		}
		max := len(lv)
		if len(rv) > max {
			max = len(rv)
		}
		for i := 0; i < max; i++ {
			var li, ri any
			if i < len(lv) {
				li = lv[i]
			}
			if i < len(rv) {
				ri = rv[i]
			}
			acc = diffValues(acc, fmt.Sprintf("%s[%d]", path, i), li, ri)
		}
		return acc
	default:
		if reflect.DeepEqual(l, r) {
			return acc
		}
		return appendDiff(acc, path, l, r)
	}
}

func appendDiff(acc []studyDiffEntry, path string, l, r any) []studyDiffEntry {
	kind := "changed"
	switch {
	case l == nil:
		kind = "added"
	case r == nil:
		kind = "removed"
	}
	return append(acc, studyDiffEntry{
		Path:  path,
		Left:  truncateCell(fmt.Sprintf("%v", derefAny(l)), 80),
		Right: truncateCell(fmt.Sprintf("%v", derefAny(r)), 80),
		Kind:  kind,
	})
}

func derefAny(v any) any {
	if v == nil {
		return ""
	}
	return v
}

func renderStudyDiff(w io.Writer, r studyDiffReport) {
	fmt.Fprintf(w, "Study diff: %s → %s\n", r.Left, r.Right)
	if r.Same {
		fmt.Fprintln(w, "\n  No differences (ignoring ids, timestamps and other bookkeeping fields).")
		return
	}
	fmt.Fprintf(w, "  %d differences across %d sections\n", len(r.Entries), len(r.Sections))
	for _, s := range deuteroSortedKeys(r.Sections) {
		fmt.Fprintf(w, "\n  %s (%d)\n", s, r.Sections[s])
		for _, e := range r.Entries {
			if e.Path != s && !strings.HasPrefix(e.Path, s+".") && !strings.HasPrefix(e.Path, s+"[") {
				continue
			}
			switch e.Kind {
			case "added":
				fmt.Fprintf(w, "    + %s = %s\n", e.Path, e.Right)
			case "removed":
				fmt.Fprintf(w, "    - %s = %s\n", e.Path, e.Left)
			default:
				fmt.Fprintf(w, "    ~ %s: %s → %s\n", e.Path, e.Left, e.Right)
			}
		}
	}
}
