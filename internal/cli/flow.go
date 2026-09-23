// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelFlowCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "flow",
		Short:       "Flow behavior you cannot see upstream",
		Example:     "  deutero-pp-cli flow coverage --study f0ed9214-06fc-4420-87b3-31240b670fc1 --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelFlowCoverageCmd(flags))
	addNovelCommandIfAbsent(cmd, newNovelFlowEffectsCmd(flags))
	return cmd
}
