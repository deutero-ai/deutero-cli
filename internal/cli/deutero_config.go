// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source computed
// Supported strategies: auto, local, live, or computed. Change this default deliberately.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// The study API is one deployment among several — production, a staging
// instance, a self-hosted one, a tunnel — and every other surface (dashboard,
// MCP client) is configured with a host. Without this the CLI can only ever
// talk to the built-in default, and the generated config file is rewritten
// wholesale by the auth commands, so a host cannot simply be parked there.

// hostFlagValue backs the persistent --host flag. It is read by
// applyHostOverride in the root PersistentPreRunE hook registered below.
var hostFlagValue string

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		root.PersistentFlags().StringVar(&hostFlagValue, "host", "",
			"Base host for this invocation (e.g. https://staging.example.com). Persist one with 'config set host'.")

		// Chain onto whatever the generated root already does rather than
		// replacing it.
		prev := root.PersistentPreRunE
		root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := applyHostOverride(hostFlagValue); err != nil {
				return err
			}
			if prev != nil {
				return prev(cmd, args)
			}
			return nil
		}

		addNovelCommandIfAbsent(root, newConfigCmd(flags))
		addCreditsBalanceAlias(root)
	})
}

func newConfigCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show and persist CLI settings (host, default study, default project)",
		Long: strings.TrimLeft(`
Show and persist settings that would otherwise have to be repeated on every
invocation.

The host is the one that matters most: this CLI points at `+DefaultHost+`
by default, and 'config set host' repoints it permanently at a self-hosted,
staging, or tunnelled instance. You give it a host; the `+APIBasePath+` base
path is appended for you.

Precedence, highest first:
  1. --host on the command line
  2. DEUTERO_HOST or DEUTERO_BASE_URL in the environment
  3. the value stored by 'config set host'
  4. the built-in default
`, "\n"),
		Example: strings.Trim(`
  deutero-pp-cli config show
  deutero-pp-cli config set host https://research.internal.example.com
  deutero-pp-cli config unset host
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newConfigShowCmd(flags), newConfigSetCmd(flags), newConfigUnsetCmd(flags))
	return cmd
}

func newConfigShowCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "show",
		Short:       "Show the resolved host and stored defaults",
		Example:     "  deutero-pp-cli config show\n  deutero-pp-cli config show --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "config show")
			}
			s := loadSettings()
			baseURL, source := resolvedHost(hostFlagValue)
			out := struct {
				Host           string `json:"host"`
				BaseURL        string `json:"base_url"`
				HostSource     string `json:"host_source"`
				DefaultStudy   string `json:"default_study,omitempty"`
				DefaultProject string `json:"default_project,omitempty"`
				SettingsPath   string `json:"settings_path"`
			}{
				Host:           hostBase(baseURL),
				BaseURL:        baseURL,
				HostSource:     source,
				DefaultStudy:   s.DefaultStudy,
				DefaultProject: s.DefaultProject,
				SettingsPath:   settingsPath(),
			}
			if flags.asJSON || flags.agent {
				return flags.printJSON(cmd, out)
			}
			rows := [][]string{
				{"host", out.Host},
				{"base url", out.BaseURL},
				{"host from", out.HostSource},
				{"default study", orNotSet(out.DefaultStudy)},
				{"default project", orNotSet(out.DefaultProject)},
				{"settings file", out.SettingsPath},
			}
			return flags.printTable(cmd, []string{"setting", "value"}, rows)
		},
	}
}

func newConfigSetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "set <host|study|project> <value>",
		Short: "Save the API host, default study, or default project to the CLI settings file",
		Example: strings.Trim(`
  deutero-pp-cli config set host https://dashboard.deutero.ai
  deutero-pp-cli config set host https://research.internal.example.com
  deutero-pp-cli config set study 3f2a9c14-8b7e-4d21-9a55-1c0e7b3d6f88
`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "config set")
			}
			key, value := strings.ToLower(strings.TrimSpace(args[0])), strings.TrimSpace(args[1])
			s := loadSettings()
			var shown string
			switch key {
			case "host":
				norm, err := normalizeHost(value)
				if err != nil {
					return err
				}
				s.Host = hostBase(norm)
				shown = norm
			case "study", "default-study":
				s.DefaultStudy = value
				shown = value
			case "project", "default-project":
				s.DefaultProject = value
				shown = value
			default:
				return fmt.Errorf("unknown setting %q: expected one of host, study, project", args[0])
			}
			if err := saveSettings(s); err != nil {
				return err
			}
			if flags.quiet {
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set %s = %s\nstored in %s\n", key, shown, settingsPath())
			return nil
		},
	}
}

func newConfigUnsetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "unset <host|study|project>",
		Short:   "Remove a stored setting and fall back to the default",
		Example: "  deutero-pp-cli config unset host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "config unset")
			}
			s := loadSettings()
			switch strings.ToLower(strings.TrimSpace(args[0])) {
			case "host":
				s.Host = ""
			case "study", "default-study":
				s.DefaultStudy = ""
			case "project", "default-project":
				s.DefaultProject = ""
			default:
				return fmt.Errorf("unknown setting %q: expected one of host, study, project", args[0])
			}
			if err := saveSettings(s); err != nil {
				return err
			}
			if flags.quiet {
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "unset %s\n", args[0])
			return nil
		},
	}
}

func orNotSet(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(not set)"
	}
	return s
}

// addCreditsBalanceAlias gives the single-endpoint `credits` command a
// `balance` subcommand. The endpoint returns a balance, and "credits balance"
// is what people type; without this the noun alone is the whole command.
func addCreditsBalanceAlias(root *cobra.Command) {
	creditsCmd, _, err := root.Find([]string{"credits"})
	if err != nil || creditsCmd == nil || creditsCmd.Name() != "credits" {
		return
	}
	run := creditsCmd.RunE
	if run == nil {
		return
	}
	creditsCmd.AddCommand(&cobra.Command{
		Use:         "balance",
		Short:       "Show the organization's credit balance",
		Long:        creditsCmd.Long,
		Example:     "  deutero-pp-cli credits balance\n  deutero-pp-cli credits balance --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE:        func(cmd *cobra.Command, args []string) error { return run(cmd, args) },
	})
}
