// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command scaffold. Implement the RunE body before shipping.
// generate --force preserves implemented bodies; untouched TODO scaffolds may refresh.
// pp:data-source auto
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"github.com/spf13/cobra"
)

func newNovelRehearseCmd(flags *rootFlags) *cobra.Command {

	cmd := &cobra.Command{
		Use:         "rehearse",
		Short:       "Flow behavior you cannot see upstream",
		Example:     "  deutero-pp-cli rehearse diff --study 092103ce-0e01-4e81-8ea5-e85fa560504f --agent",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto"},
		RunE:        parentNoSubcommandRunE(flags),
	}
	addNovelCommandIfAbsent(cmd, newNovelRehearseDiffCmd(flags))
	return cmd
}
