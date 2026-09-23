Manifest transcendence rows: 10 planned, 10 built. Phase 3 completion gate passed.

# Build log — deutero-pp-cli (reprint)

Run `20260921-221354-646207bd`. Generated 78 endpoints from the source-dumped spec;
10 hand-code transcendence rows approved at the absorb gate.

## Slice plan

| Slice | Scope |
|---|---|
| A | Shared novel-command helpers: interview/transcript fetch with bounded fan-out, authoring-id handling, "not observed" semantics, output plumbing |
| B | `flow coverage`, `flow effects`, `rehearse diff` |
| C | `answers matrix`, `crosstab`, `saturation`, `fielding` |
| D | `study bundle`, `study diff`, `reconcile` |

Every slice ends build-green (`go build ./...` + `go vet ./...`).

## What was built

All 10 approved transcendence rows, plus `study apply` (the write half of manifest row 4,
which the generator never scaffolded — building it was required rather than optional, since
dropping it would have been a silent mid-build downgrade of approved scope).

| Command | File | Data source | Notes |
|---|---|---|---|
| `flow coverage` | `flow_coverage.go` | live | Branch arms, `used_default` rate, classifier errors, step reach, variable capture |
| `flow effects` | `flow_effects.go` | live | Fetch failures and signal deliveries, dry-run kept separate |
| `rehearse diff` | `rehearse_diff.go` | live | Simulated vs real cohort on shared measures |
| `answers matrix` | `answers_matrix.go` | live | Rectangle from variables **and** message stream |
| `crosstab` | `crosstab.go` | live | Segment by characteristic; marginals reconcile |
| `saturation` | `saturation.go` | live | First-appearance term curve, plateau detection |
| `fielding` | `fielding.go` | live | Trailing-window rate, projected fill date |
| `reconcile` | `reconcile.go` | live | Completion↔delivery anti-join on `external_participant_id` |
| `study bundle` | `study_bundle.go` | live | 8-section definition export |
| `study diff` | `study_diff.go` | live | Recursive field diff, bookkeeping fields ignored |
| `study apply` | `study_apply.go` | live | Plan by default; writes only with `--confirm` |

Shared plumbing in `deutero_shared.go`: bounded concurrent fan-out (6 workers) with stderr
progress, payload types, authoring-id guard, "not observed" note, value rendering.

## Decisions worth recording

**A spec reading was corrected mid-build.** The absorb manifest and the novel-features
brainstorm both assumed captured `variables` key on `question_id` for linear studies, so one
implementation would serve both interview modes. The spec says the opposite: variables are
"Empty for a study that captures nothing, **including every linear study**", and decisions
are "Empty for a linear study or a flow with no branching". A variables-only `answers matrix`
would therefore have returned an empty table for every linear study on the platform. The
command now reads variables *and* the message stream and reports which source it used. The
manifest, `research.json` and the shared-code header were all corrected to match; two tests
pin the behaviour per mode.

**Three manifest rows were malformed and were normalized at the completion gate**, not
worked around: row 34 named `analysis cluster` (the real path is
`studies analysis run-clustering`), row 40's implementation cell was `— global --csv flag`
which is not a command path at all, and row 31's `search` takes a positional so its usage
line is `search <query> [flags]`.

**Fan-out is bounded and visible** rather than hidden, which is the direct answer to the
"pagination / N+1 cost" frustration raised at the gate: every study-wide command takes
`--limit` (default 500), pages explicitly, runs at most 6 concurrent requests, and narrates
progress on stderr so stdout stays a clean JSON/CSV stream.

**`study apply` prints by default and writes only on `--confirm`**, and short-circuits under
`dryRunOK` before any IO. Replaying a definition overwrites a live study's welcome text,
screening rules and flow; that is not something to do as a side effect of a typo.

**Per-interview failures degrade rather than abort.** One unreadable transcript should not
deny an answer about the other 200, so failures are counted and surfaced in the report.

## Gates

- `go build ./...`, `go vet ./...` — pass after every slice
- `go test -count=1 ./...` — all 17 packages pass
- Phase 3 per-row Cobra resolution — 15/15 approved paths resolve
- `dogfood --json` `novel_features_check` — planned 10, found 10, none missing, not skipped

## Deferred

Nothing from the approved manifest. `study apply` deliberately does not replay per-item
collections (question lists, individual screening/characteristic questions); those have their
own create/reorder semantics and are reported as needing the dedicated commands rather than
half-applied. This is stated in the command's own help, not hidden.
