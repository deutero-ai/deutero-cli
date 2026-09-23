# Phase 5.5 — Polish

`ship_recommendation: ship`, `further_polish_recommended: no`.

| Signal | Before | After |
|---|---|---|
| Scorecard | 96/100 | **97/100** |
| Verify | 100% | 100% |
| Dogfood | WARN | **PASS** |
| gosec (hand-authored) | 2 | **0** |
| tools-audit pending | 2 | **0** |

## Applied

- **gosec G304 ×2** on the two hand-authored bundle reads (`study_diff.go --against`,
  `study_apply.go --file`): `filepath.Clean` with a narrow `#nosec` rationale.
- **Dead code removed**: `handleBinaryResponseDelivery` plus the four helpers it alone kept
  alive. Dead Functions 1 → 0, Dead Code 4/5 → 5/5, dogfood WARN → PASS.
- **`--csv` made honest on all nine commands that accepted and ignored it.** Real CSV
  writers for `reconcile`, `fielding` and `saturation` through a formula-injection-safe
  writer; the six genuinely nested reports now return a usage error pointing at
  `--json`/`--select` instead of writing JSON into a file called `.csv`.
- **`--limit` clamped to 1..10000** on the eight fan-out commands, range documented per flag;
  `--limit 0` and negatives no longer silently mean "default".
- New `deutero_output_test.go` covering the CSV writers, the `--csv` refusal and limit bounds.

## Deliberately not changed

Cache Freshness 5/10, MCP Quality 8/10 (78 auth-required tools), `reconcile`'s org-wide join,
and the three separate fan-out helpers — all previously reasoned about and recorded.

## Handed to the retro / upstream

- **Terse flag descriptions on update commands** (`--max-turns` "Max turns", ~40 similar).
  Root cause is upstream in the dashboard repo: the `*Update` Pydantic schemas carry no field
  descriptions while `*Create` does — `study_api/schemas/questions.py` `QuestionUpdate` and
  `study_api/schemas/studies.py` `StudyUpdate`. Fixing those and re-extracting the spec fixes
  every generated CLI, so it was left alone here rather than hand-patched.
- gosec G202 `internal/platform/migration.go` and G302 `internal/cliutil/testenv/testenv.go`,
  plus `dogfood`'s `which.go` sync emitting non-gofmt output on every run — generator-owned;
  added to `retro-candidates.md`.

## Live evidence caveat, and what was done about it

Polish ran in a forked session where Go processes could not reach the network (the Little
Snitch condition), so its own `scorecard --live-check` reported every feature as a 10s
timeout and it recorded `live_matrix: not_exercised`. Its scorecard Insight figure moves
with that, not with quality.

Because polish changed real behaviour (CSV paths, limit validation, deleted helpers), the
388/388 acceptance evidence collected before polish no longer described the shipping binary.
The live matrix was therefore re-run on the post-polish binary in the parent session; that
result is the one recorded in `phase5-acceptance.json`.

Verification reported by polish: `go build`, `go vet`, `go test -count=1 ./...` and
`go test -race` clean; shipcheck PASS 7/7; verify 117/117; verify-skill 0 findings;
workflow-verify `workflow-pass`. No regenerate was run — the `event_type` override, the five
template dry-run patches and the README/SKILL example ids were all confirmed intact
afterwards.

## Post-polish live re-verification (parent session)

`dogfood --live --level full` re-run against the post-polish binary: **388/388 PASS**,
`phase5-acceptance.json` rewritten by the runner with `status: pass`.

An intermediate attempt reported 93 failures, all `exit -1` (timeout). That run overlapped
polish's own background work, which rebuilds the binary — a binary swapped mid-matrix, plus
the per-binary firewall re-challenging each new build. Re-run once no Printing Press process
was live and three consecutive cold, cache-less calls returned in ~0.7 s: clean 388/388.
Recorded here rather than dropped, because "93 failures" appears in this run's history and
should not be mistaken for a regression.
