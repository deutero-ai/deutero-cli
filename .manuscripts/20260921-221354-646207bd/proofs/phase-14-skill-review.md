# Phase 4.8 — Agentic SKILL review

4 errors, 7 warnings, 1 CLI bug. **All fixed**; no findings deferred.

| # | Severity | Finding | Fix |
|---|---|---|---|
| 1 | error | Tagline claimed answers "only exist once your research corpus is sitting in a local database" — false; all 10 novel commands are `pp:data-source live` | Headline rewritten at source (`research.json`) and regenerated, so all five description surfaces changed together |
| 2 | error | "Every Deutero endpoint" vs a Command Reference that omits the `studies <group>` subtrees | Claim is true of the CLI (78 endpoints exist under `studies …`); value_prop now states what `sync` does and does not cover so the doc does not imply an offline corpus |
| 3 | error | `answers matrix` advertised characteristic widening it did not implement | **Implemented** `--with-characteristics` rather than deleting the approved claim; CSV and JSON both carry the columns |
| 4 | error | `study bundle` described replay, but `study apply` was undocumented | Description now names `study apply --file … --confirm` as the write half |
| 5 | warning | `study apply` silently omits per-item collections | Stated in the description and in the command's own help |
| 6 | warning | flow commands empty for linear studies, unstated | "Flow-mode studies only" added to those descriptions |
| 7 | warning | 500-interview default cap unstated | `--limit` cap now stated in value_prop |
| 8 | warning | saturation oversold as "the evidence for calling a study done" | Reworded to "a mechanical word-level proxy … evidence for that call rather than the call itself" |
| 9 | warning | Auth narrative cited a setup URL that does not exist | Rewritten around `DEUTERO_API_KEY` / `auth set-token` / `auth status` |
| 10 | warning | Recipes used literal `mock-value` ids that 404 | Replaced with a real interview id |
| 11 | warning | "Offline-friendly" implied an offline message index | Troubleshooting states sync covers org-level resources only; search is server-side |
| — | bug | `crosstab --help` and `study apply --help` cited non-existent top-level `characteristics` / `questions` / `screening` commands | Corrected to `studies characteristics get`, `studies questions`, `studies screening` |

Passing: verified-set alignment (10 capabilities exactly match `novel_features_built`), no marketing-copy smell, recipe `--select` paths verified against the `branchStat` struct.

Re-verified after fixes: shipcheck PASS 7/7, tests pass, sync 6/6.
