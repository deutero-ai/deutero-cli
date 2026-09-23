// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Authoring a study is roughly fifteen ordered API calls and nothing captures
// the result, so a working study exists only on the server. A bundle is that
// definition as one file: reviewable in a pull request, diffable against another
// study, and replayable when you want a second study just like the first.

func newNovelStudyBundleCmd(flags *rootFlags) *cobra.Command {
	var flagStudy string
	var flagOut string
	var flagIncludeSecrets bool

	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Export a whole study definition as one file",
		Long: `Reads every part of a study definition — details, welcome and consent,
welcome translations, screening, characteristics, the question list or the
interview flow, and recruitment — and writes them as a single JSON document.

Sections that do not apply are reported under unreadable_sections rather than
failing the export: a linear study has no flow, a flow study has no question
list, and a study with no consent screen has no welcome.

Credential-shaped values — a flow step's Authorization header, tokens, API keys —
are redacted unless --include-secrets is passed, so a bundle is safe to commit.

Use this command to snapshot or version a study definition. Do NOT use this
command to read collected responses; use 'answers matrix' instead.`,
		Example: "  deutero-pp-cli study bundle --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{
			"mcp:read-only":  "true",
			"pp:data-source": "live",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "study bundle")
			}
			if err := deuteroRejectCSV(cmd, flags); err != nil {
				return err
			}
			if err := deuteroRequireStudy(cmd.CommandPath(), flagStudy); err != nil {
				return err
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			bundle, err := deuteroFetchBundle(cmd.Context(), c, flagStudy, deuteroProgressWriter(cmd, flags))
			if err != nil {
				return classifyAPIError(cmd.OutOrStdout(), err, flags)
			}
			if !flagIncludeSecrets {
				deuteroRedactBundle(bundle)
			}
			if flagOut != "" {
				blob, err := json.MarshalIndent(bundle, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(flagOut, blob, 0o600); err != nil {
					return fmt.Errorf("writing %s: %w", flagOut, err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", flagOut)
				return nil
			}
			return printJSONFiltered(cmd.OutOrStdout(), bundle, flags)
		},
	}
	cmd.Flags().StringVar(&flagStudy, "study", "", "Study ID to export (required)")
	cmd.Flags().StringVar(&flagOut, "output", "", "Write the bundle to this file instead of stdout")
	cmd.Flags().BoolVar(&flagIncludeSecrets, "include-secrets", false, "Include credential-shaped values (flow step headers, tokens) instead of redacting them")
	return cmd
}
