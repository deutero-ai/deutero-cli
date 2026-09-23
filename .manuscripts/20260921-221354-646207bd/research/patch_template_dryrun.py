"""Local patches for five Printing Press template defects found by the live matrix.

These are generator bugs, not this CLI's code. They are patched here so the Full
dogfood gate passes honestly, and each is recorded for the retro so the template
gets fixed. A future `generate --force` WILL revert them unless the generator is
fixed first — re-run this script after any regeneration.

1. doctor           ignored --dry-run entirely (it even made network calls)
2. sync             previewed correctly under --dry-run but omitted dry_run:true
3. workflow archive ignored --dry-run
4/5. feedback list, profile list — see PATCH_LIST below
"""
import sys

CLI = sys.argv[1] + "/internal/cli/"
applied = []


def patch(fname, old, new, label):
    p = CLI + fname
    s = open(p).read()
    if new in s:
        applied.append(label + " (already)")
        return
    if old not in s:
        print("NOT FOUND:", label)
        return
    open(p, "w").write(s.replace(old, new, 1))
    applied.append(label)


GUARD = """
			// Template fix (retro): honour --dry-run like every other command, so
			// a --dry-run --json caller gets the standard envelope rather than a
			// live run that omits it.
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "%s")
			}"""

# 1. doctor
patch("doctor.go",
      "\t\tRunE: func(cmd *cobra.Command, args []string) error {\n\t\t\tif registeredPlatformSource != nil {",
      "\t\tRunE: func(cmd *cobra.Command, args []string) error {" + (GUARD % "doctor") +
      "\n\t\t\tif registeredPlatformSource != nil {",
      "doctor honours --dry-run")

# 2. sync: keep the real preview, just mark it
patch("sync.go",
      """				if err := printJSONFiltered(cmd.OutOrStdout(), map[string]any{
					"total_records": totalSynced,""",
      """				summary := map[string]any{
					"total_records": totalSynced,""",
      "sync summary: named map")
patch("sync.go",
      """					"duration_ms":   elapsed.Milliseconds(),
				}, flags); err != nil {
					return err
				}""",
      """					"duration_ms":   elapsed.Milliseconds(),
				}
				// Template fix (retro): a dry-run sync previews without writing;
				// say so in the envelope, as every other --dry-run --json does.
				if c.DryRun {
					summary["dry_run"] = true
					summary["action"] = "sync"
					summary["would"] = "sync the listed resources; nothing was written"
				}
				if err := printJSONFiltered(cmd.OutOrStdout(), summary, flags); err != nil {
					return err
				}""",
      "sync marks dry_run:true")

# 3. workflow archive
patch("channel_workflow.go",
      "\t\tRunE: func(cmd *cobra.Command, args []string) error {\n\t\t\tif maxPages < 0 {",
      "\t\tRunE: func(cmd *cobra.Command, args []string) error {" + (GUARD % "workflow archive") +
      "\n\t\t\tif maxPages < 0 {",
      "workflow archive honours --dry-run")

print("applied:", applied)

# 4. feedback list  5. profile list — returned a bare array under --json --dry-run
patch("feedback.go",
      "\t\tRunE: func(cmd *cobra.Command, _ []string) error {\n\t\t\tp, err := feedbackFilePath()",
      "\t\tRunE: func(cmd *cobra.Command, _ []string) error {" + (GUARD % "feedback list") +
      "\n\t\t\tp, err := feedbackFilePath()",
      "feedback list honours --dry-run")
patch("profile.go",
      "\t\tRunE: func(cmd *cobra.Command, _ []string) error {\n\t\t\ts, err := loadProfileStore()",
      "\t\tRunE: func(cmd *cobra.Command, _ []string) error {" + (GUARD % "profile list") +
      "\n\t\t\ts, err := loadProfileStore()",
      "profile list honours --dry-run")
print("applied (incl. list commands):", applied)
