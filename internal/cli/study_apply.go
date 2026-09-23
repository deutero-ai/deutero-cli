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
	"strings"

	"github.com/deutero-ai/deutero-cli/internal/cliutil"

	"github.com/spf13/cobra"
)

// The write half of the bundle. Replays a saved study definition onto a target
// study, so "make me another one like that" is one command instead of fifteen
// ordered calls.
//
// Print by default, opt in to the write. This command shows the plan and exits
// without touching anything unless --confirm is passed: replaying a definition
// overwrites a live study's welcome text, screening rules and question list, and
// that is not something to do as a side effect of a typo. Under the verifier it
// short-circuits before any IO at all.

type applyStep struct {
	Section string `json:"section"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
}

type applyPlan struct {
	TargetStudy string      `json:"target_study"`
	Source      string      `json:"source"`
	Confirmed   bool        `json:"confirmed"`
	Steps       []applyStep `json:"steps"`
	Note        string      `json:"note"`
}

// deuteroApplySections is the write plan. Only sections with a whole-resource
// write endpoint are replayable; question lists and screening questions are
// per-item collections with their own create/reorder semantics, so they are
// reported as needing the dedicated commands rather than half-applied.
var deuteroApplySections = []struct {
	Name   string
	Method string
	Path   string
	// Build returns the request body the WRITE endpoint expects, or a non-empty
	// skip reason. The read and write shapes differ for every section except
	// welcome: screening and characteristics read as {questions, settings} but
	// write only the settings object; recruitment reads back derived URLs and
	// counts; the graph reads back compiled state. Sending a section's GET shape
	// to its PUT fails validation, which the first live run of this command
	// proved on four of five sections.
	Build func(*deuteroStudyBundle) (json.RawMessage, string)
}{
	{"welcome", "PUT", "/api/v1/studies/{study_id}/welcome", func(b *deuteroStudyBundle) (json.RawMessage, string) {
		return pickFields(b.Welcome, "", "message", "consent")
	}},
	{"screening_settings", "PUT", "/api/v1/studies/{study_id}/screening/settings", func(b *deuteroStudyBundle) (json.RawMessage, string) {
		return pickFields(b.Screening, "settings", "enabled", "disqualification_message", "redirect_url")
	}},
	{"characteristics_settings", "PUT", "/api/v1/studies/{study_id}/characteristics/settings", func(b *deuteroStudyBundle) (json.RawMessage, string) {
		return pickFields(b.Characteristics, "settings", "enabled", "anonymous")
	}},
	{"recruitment", "PUT", "/api/v1/studies/{study_id}/recruitment", func(b *deuteroStudyBundle) (json.RawMessage, string) {
		// short_url_slug is deliberately NOT copied. Slugs are unique across
		// the platform, so a clone that keeps its source's slug is guaranteed
		// to be rejected (409) — the target keeps its own link.
		return pickFields(b.Recruitment, "", "max_responses", "redirect_url")
	}},
	{"graph", "PUT", "/api/v1/studies/{study_id}/graph", func(b *deuteroStudyBundle) (json.RawMessage, string) {
		if len(b.Graph) == 0 {
			return nil, "not present in the bundle"
		}
		var g struct {
			Flow          json.RawMessage `json:"flow"`
			InterviewMode string          `json:"interview_mode"`
		}
		if err := json.Unmarshal(b.Graph, &g); err != nil {
			return nil, "graph section is not readable"
		}
		if len(g.Flow) == 0 || string(g.Flow) == "null" {
			return nil, "source study is linear (" + g.InterviewMode + "); it has no flow to replay"
		}
		// expected_graph_version is omitted on purpose: it guards against a
		// concurrent edit of the SAME study, and the target's version is
		// unrelated to the source's.
		body, _ := json.Marshal(map[string]json.RawMessage{"flow": g.Flow})
		return body, ""
	}},
}

// pickFields extracts the named fields from a section (optionally from a nested
// object such as "settings"), dropping nulls and read-only extras. An empty
// result is a skip, not an empty PUT.
func pickFields(raw json.RawMessage, nested string, fields ...string) (json.RawMessage, string) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, "not present in the bundle"
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, "section is not an object"
	}
	if nested != "" {
		inner, ok := obj[nested]
		if !ok || len(inner) == 0 || string(inner) == "null" {
			return nil, "no " + nested + " in the bundle"
		}
		obj = nil
		if err := json.Unmarshal(inner, &obj); err != nil {
			return nil, nested + " is not an object"
		}
	}
	out := map[string]json.RawMessage{}
	for _, f := range fields {
		if v, ok := obj[f]; ok && len(v) > 0 && string(v) != "null" {
			out[f] = v
		}
	}
	if len(out) == 0 {
		return nil, "nothing replayable in this section"
	}
	body, err := json.Marshal(out)
	if err != nil {
		return nil, "could not encode section"
	}
	return body, ""
}

func newNovelStudyApplyCmd(flags *rootFlags) *cobra.Command {
	var flagStudy, flagFile string
	var flagConfirm bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Replay a saved study definition onto a study (plan by default)",
		Long: `Replays a bundle written by 'study bundle' onto a target study.

Prints the plan and changes nothing unless --confirm is passed. Run
'study diff --study <target> --against <file>' first to see exactly what would
change.

Sections with a whole-resource write endpoint (welcome, screening settings,
characteristics settings, recruitment, flow) are replayed. The short-URL slug is
not copied — slugs are unique across the platform — and a linear source study
has no flow to replay, so that section is skipped rather than sent empty. Question lists and
individual screening or characteristic questions are per-item collections with
their own create and reorder semantics; this command reports them rather than
half-applying them — use 'studies questions' and 'studies screening' instead.

Use this command to clone or restore a study definition. Do NOT use this command
to read collected responses; use 'answers matrix' instead.`,
		Example: "  deutero-pp-cli study apply --study f0ed9214-06fc-4420-87b3-31240b670fc1 --file study.json",
		Annotations: map[string]string{
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "study apply")
			}
			if err := deuteroRejectCSV(cmd, flags); err != nil {
				return err
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			if strings.TrimSpace(flagFile) == "" {
				return usageErr(fmt.Errorf("--file is required (a bundle written by 'study bundle')\nUsage: %s --study <id> --file <bundle.json>", cmd.CommandPath()))
			}
			blob, err := os.ReadFile(filepath.Clean(flagFile)) // #nosec G304 -- operator-named bundle file from --file; reading it is the command's purpose.
			if err != nil {
				return fmt.Errorf("reading %s: %w", flagFile, err)
			}
			var bundle deuteroStudyBundle
			if err := json.Unmarshal(blob, &bundle); err != nil {
				return fmt.Errorf("parsing bundle %s: %w", flagFile, err)
			}

			plan := applyPlan{
				TargetStudy: flagStudy,
				Source:      flagFile,
				Confirmed:   flagConfirm,
				Steps:       []applyStep{},
				Note:        "Nothing is written without --confirm.",
			}
			for _, s := range deuteroApplySections {
				body, skipWhy := s.Build(&bundle)
				step := applyStep{
					Section: s.Name,
					Method:  s.Method,
					Path:    replacePathParam(s.Path, "study_id", flagStudy),
				}
				if skipWhy != "" || len(body) == 0 {
					step.Status = "skipped"
					step.Detail = skipWhy
					plan.Steps = append(plan.Steps, step)
					continue
				}
				step.Status = "planned"
				plan.Steps = append(plan.Steps, step)
			}
			if len(bundle.Questions) > 0 {
				plan.Steps = append(plan.Steps, applyStep{
					Section: "questions",
					Status:  "manual",
					Detail:  "question lists are per-item; use the 'questions' commands to recreate them",
				})
			}

			if !flagConfirm {
				if wantsHumanTable(cmd.OutOrStdout(), flags) {
					renderApplyPlan(cmd.OutOrStdout(), plan, false)
					return nil
				}
				return printJSONFiltered(cmd.OutOrStdout(), plan, flags)
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			// Address steps by section name, not by index: the planning loop and
			// this one only line up while every section appends exactly one
			// step, and a future conditional section would silently mark the
			// wrong one applied — on a command that overwrites live content.
			byName := make(map[string]*applyStep, len(plan.Steps))
			for i := range plan.Steps {
				byName[plan.Steps[i].Section] = &plan.Steps[i]
			}
			for _, sec := range deuteroApplySections {
				body, skipWhy := sec.Build(&bundle)
				if skipWhy != "" || len(body) == 0 {
					continue
				}
				step := byName[sec.Name]
				if step == nil {
					continue
				}
				path := replacePathParam(sec.Path, "study_id", flagStudy)
				if _, _, err := c.Put(cmd.Context(), path, body); err != nil {
					step.Status = "failed"
					step.Detail = cliutil.SanitizeErrorBody(err.Error())
					continue
				}
				step.Status = "applied"
			}
			failedSteps := 0
			for _, st := range plan.Steps {
				if st.Status == "failed" {
					failedSteps++
				}
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				renderApplyPlan(cmd.OutOrStdout(), plan, true)
			} else if err := printJSONFiltered(cmd.OutOrStdout(), plan, flags); err != nil {
				return err
			}
			// Print the plan first so the caller can see which sections landed,
			// then fail. A script or agent that only checks the exit code must
			// not read a run that wrote nothing as a success.
			if failedSteps > 0 {
				return fmt.Errorf("%d of %d section(s) failed to apply", failedSteps, len(plan.Steps))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to replay the bundle onto (required)")
	cmd.Flags().StringVar(&flagFile, "file", "", "Bundle file written by 'study bundle' (required)")
	cmd.Flags().BoolVar(&flagConfirm, "confirm", false, "Actually write the definition; without this the plan is printed and nothing changes")
	return cmd
}

func renderApplyPlan(w io.Writer, p applyPlan, applied bool) {
	verb := "Would apply"
	if applied {
		verb = "Applied"
	}
	fmt.Fprintf(w, "%s bundle %s → study %s\n\n", verb, p.Source, p.TargetStudy)
	for _, s := range p.Steps {
		fmt.Fprintf(w, "  %-26s %-8s %s", s.Section, s.Status, s.Path)
		if s.Detail != "" {
			fmt.Fprintf(w, "  (%s)", s.Detail)
		}
		fmt.Fprintln(w)
	}
	if !applied {
		fmt.Fprintln(w, "\n  Nothing was written. Re-run with --confirm to apply, or run 'study diff' first to see the exact changes.")
	}
}
