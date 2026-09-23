# Retro candidates — Printing Press defects found on the deutero reprint

Binary v4.32.4, skills v3.0.0. Each item below is a **generator / template / tooling**
defect, not a defect in the Deutero CLI's own code. Items marked *patched locally* were
fixed in this CLI's generated files so the Full dogfood gate could pass; a future
`generate --force` reverts them unless the generator is fixed first
(`research/patch_template_dryrun.py` re-applies them).

## Framework commands ignore --dry-run (patched locally)

1. **`doctor` ignores `--dry-run` entirely** — it runs the full diagnostic, including
   network calls, and returns a report with no `dry_run` marker. Fails live-dogfood
   `json_fidelity`.
2. **`sync --dry-run --json` omits `dry_run: true`** — the preview is correct (nothing is
   written), only the envelope marker is missing.
3. **`workflow archive` ignores `--dry-run`.**
4. **`feedback list --json --dry-run`** and 5. **`profile list --json --dry-run`** return a
   bare JSON array instead of the `{"dry_run": true, ...}` envelope, which the runner
   reports as `dry_run_json: invalid JSON`.

Suggested fix: every template-emitted command's RunE should open with the
`if dryRunOK(flags) { return writeDryRun(...) }` guard that generated endpoint commands
already use; `sync` should add the marker to its summary.

## Sync and store

6. **ID override mis-derived for a catalogue resource.** `webhooks-event-types` got
   `"envelope"` as its ID field. `envelope` is a sibling field of the *list* response, not a
   key on each item, so all 7 rows failed ID extraction and `sync` reported a critical
   failure. The natural key is `event_type`. The override exists in **two** maps
   (`internal/cli/sync.go` and `internal/store/store.go`; the store one is authoritative), and
   `generate --force` reverts both — it warns `TEMPLATED-VALUE-DRIFT` but still overwrites.
7. **Study-scoped resources are listed as syncable but can never sync.** `interviews`,
   `transcripts` and `simulations` appear in `knownSyncResourceNames` and `syncResourcePath`
   with a `{study_id}` path key, but no dependent-resource definition supplies a parent
   chain (`studies` is not syncable), so they always warn `unfilled_path_key` and skip. The
   generated README and research narrative then promised an offline interview corpus that
   cannot exist. Either wire a parent chain or stop advertising them as syncable.
8. **Single-object responses warn "no extractable ID field".** `credits balance` and
   `health` return one object; caching them as rows is meaningless, and the warning on
   every call reads as a fault.

## Examples and help

9. **No `Example` for body-only commands.** The generator synthesises examples only from
   *required* parameters; request-body schema examples (media-level or property-level) are
   ignored, so `create`/`update`/`set` commands ship with no Examples section and fail the
   help check (34 commands here). Needed `x-pp-example` on every such operation.
10. **"One of these is required" rules are invisible.** `optimal-clusters` requires
    `question_id` *or* `question_number`, enforced in code, not the schema — so the example
    and happy-args omit both and the live call 400s. Needed `x-pp-example` + `x-happy-args`.
11. **Happy paths ignore `x-pp-example`.** The live matrix drives happy paths from
    `pp:happy-args` (derived from parameter examples), not from the Cobra Example, so
    fixing an example does not fix its live test. Surprising; worth documenting or unifying.
12. **`mock-value` placeholder in the generated README/SKILL "Output Formats" block**
    ignores spec path-parameter examples entirely. Hand-replaced after every regenerate.
13. **Dry-run envelope silently leaves JSON mode under `--csv`.** A novel command whose
    Example includes `--csv` produces prose under `--dry-run` even when piped, failing
    `dry_run_json`. Hand-written commands had to drop `--csv` from their Examples.
14. **Novel-command group files keep stale Examples.** `answers.go`, `flow.go`,
    `rehearse.go`, `study.go` are preserved across regenerate and kept an Example from an
    earlier `research.json` (including a shell redirect `'>' answers.csv` passed as literal
    args) after `research.json` was corrected.

## Validation and tooling

15. **`generate --validate` cannot run `govulncheck` offline or behind a per-app
    firewall.** It always fetches `vuln.go.dev`; `GOVULNDB` is ignored (modern govulncheck
    only honours `-db`) and there is no flag to pass one. Worked around by mirroring the DB
    and running govulncheck manually against `file://` — result: no vulnerabilities.
16. **Legacy `cmd/printing-press` entrypoint vs `cmd/cli-printing-press`.** Installing the
    old entrypoint gave a binary that passes `version --json` but the v3 preflight's
    resolution order made the phase-receipt check fail confusingly. The installer and the
    v2 skill text disagree about the binary name.
17. **Skill install drift is only detected after an upgrade.** v4.29 did not report
    `skill_status`; upgrading the binary surfaced that the installed skills were v2.0.0
    (repo tag v4.8.0) — 24 minor versions stale.

## Environment note (not a Press defect, but worth a doc line)

Little Snitch on the operator's Mac blocked each freshly built, unsigned binary as a new
application; `curl` (already allowed) always worked. This made failures look like
intermittent network trouble and briefly produced timings that were response-cache hits.
Doctor could detect a per-app firewall and say so.
