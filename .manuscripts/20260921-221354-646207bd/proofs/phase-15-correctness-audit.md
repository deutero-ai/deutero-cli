# Phase 4.9 — README / SKILL / AGENTS correctness audit

Auditor walked the full recursive `--help` tree (161 commands), the Go sources, the sync
registry, exit codes and the `--select` filter. 3 errors, 2 warnings.

**Verified correct:** all 10 Unique Features/Capabilities match `novel_features_built` and
resolve; every flag in every example exists; `--select branches.step,...` matches the real
`branchStat` JSON tags; exit codes 0/2/3/4/5/6/7/10 all exist; the documented sync coverage
matches the registry; positional-arg shapes are right; the write surface is disclosed;
4 anti-triggers present; "Deutero" used as the prose brand.

| # | Severity | Finding | Fix |
|---|---|---|---|
| 1 | error | `mock-value` placeholder in 5 executable README examples | Replaced with a real interview UUID |
| 2 | error | Same placeholder in 2 SKILL examples | Replaced |
| 3 | error | Docs said the key "is sent as X-API-Key"; the CLI actually sends `Authorization: Bearer` (`config.go:301`, `client.go:1121`) — there is no X-API-Key path in the binary | Auth narrative rewritten at source to describe the request the CLI really makes, noting the API also accepts X-API-Key but this CLI does not use it |
| 4 | warning | `<command>` / `<capability>` placeholders inside AGENTS.md bash blocks | Left as-is: generated boilerplate, reads as deliberate generic instruction rather than unsubstituted output |
| 5 | error→fixed | README's two sync-coverage lists disagreed (one omitted node types and health) | Both now render from one string: projects, webhooks, webhook deliveries, embed keys, graph node types and health |

Finding 3 is the one that mattered: the docs described a request the binary never makes.

**Regeneration note.** Fixing the description at source required a regenerate, which reverted
the hand-applied `webhooks-event-types` ID override in `internal/store/store.go` back to
`envelope` (the generator re-derives it and warned `TEMPLATED-VALUE-DRIFT`). Re-applied with a
comment explaining why. This is a generator-side mis-derivation worth a retro: `envelope` is a
sibling field of the list response, not a key on each item.
