// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelAnswersCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "answers",
		Short:       "Local state that compounds",
		Example:     "  deutero-pp-cli answers matrix --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelAnswersMatrixCmd(flags))
	return cmd
}
