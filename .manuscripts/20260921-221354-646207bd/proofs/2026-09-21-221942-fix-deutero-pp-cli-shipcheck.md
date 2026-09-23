# Shipcheck — deutero-pp-cli (reprint)

Run `20260921-221354-646207bd`.

## Final verdict: **ship**

```
LEG                 RESULT  EXIT      ELAPSED
verify              PASS    0         24.183s
validate-narrative  PASS    0         482ms
dogfood             PASS    0         7.157s
workflow-verify     PASS    0         35ms
apify-audit         PASS    0         131ms
verify-skill        PASS    0         11.531s
scorecard           PASS    0         17.672s

Verdict: PASS (7/7 legs passed)
```

Scorecard **95/100 — Grade A**. Verify pass rate **100% (117/117, 0 critical)**.

## Before / after

| Signal | First run | Final |
|---|---|---|
| Legs passing | 4/7 | **7/7** |
| Scorecard | 91/100 | **95/100** |
| Verify verdict | FAIL (sync crashed) | **PASS** |
| Live sample probe | 1/10 | 7–9/10 (cold-start bound, see below) |
| Data Pipeline Integrity | 5/10 | **10/10** |
| Insight | 4/10 | **10/10** |

## Blockers found and fixed

1. **Examples named a study the key cannot reach.** Every live probe returned
   `403 Access denied`. The study id came from a prior session's notes and belongs to a
   different organization than the supplied key. All examples now use studies the key
   actually owns, and a 403-specific troubleshooting entry was added, since "your key is
   valid but that study is in another org" is otherwise a confusing failure.

2. **`sync` crashed — generator mis-derived an ID field.** `webhooks-event-types` had its
   ID override set to `envelope`, which is a sibling field of the *list* response rather
   than a key on each item, so all 7 rows failed ID extraction and sync reported a critical
   failure. Corrected to `event_type` in **both** copies of the override map
   (`internal/cli/sync.go` and `internal/store/store.go` — the second is the one
   `ExtractResourceID` actually reads). Sync now completes 6/6 resources, 24 records, 0 errors.

3. **`studies list` did not exist.** The README and quickstart advertised
   `studies list --project-id …`; the real path is `projects studies list <project_id>`.
   Caught by verify-skill and validate-narrative.

4. **`sync --resources interviews,transcripts` cannot work** — those endpoints are scoped to
   a `{study_id}` that sync has no parent chain to supply, so they always warn and skip. The
   quickstart advertised this as the key step. Removed, and two absorbed manifest rows that
   overclaimed on it (row 31 "offline FTS over synced messages", row 41 "incremental sync of
   a research corpus") were corrected to describe what actually works. The novel commands are
   unaffected: they read live and never needed the local store.

5. **`study diff --against typo.json` silently became a remote fetch.** A missing or mistyped
   bundle path fell through to treating the filename as a study id, producing eight slow 404s
   and an empty diff — which reads as "no differences", the most misleading answer this
   command can give. Now a value that is plainly a path fails fast with a clear error.

6. **Bundle reads were sequential.** Eight round trips against a host that cold-starts at ~4s
   blew the 10s sample budget. Section reads are now concurrent, and a two-study diff fetches
   both sides in parallel: 8.0s → 4.9s.

## Known gap (not a blocker)

The live sample probe reports 7–9/10 across runs, and every failure is a **timeout, not wrong
output**. The probe allows 10s per command; the production host cold-starts at ~4s after idle
(measured: 4083 ms first request, 296 ms second). Run warm, the same commands complete
correctly in ~1.0–1.7s:

- `flow coverage` on a 2-interview study: 1.69s, correctly reports **no branches** for a
  linear study rather than inventing any
- `reconcile`: 1.04s, correctly reports 1 completion with **nothing to join on** (no external
  participant id) rather than miscounting it as a delivery failure
- `study diff` between two live studies: 4.88s

This is environmental latency, not a defect in the CLI. It is recorded here rather than
suppressed.

## Remaining scorecard points

- Dead Code 4/5 and Cache Freshness 5/10 — polish territory, handled in Phase 19.
- `mcp_description_quality` and `mcp_token_efficiency` are omitted from the denominator.

## Re-run after Phase 4.95 fixes (final)

`PASS (7/7 legs)`, scorecard **96/100 Grade A**, Insight 10/10, live sample probe **8/10**.

**Root cause of the intermittent "network" failures in the middle of this run:** Little
Snitch's network extension on the operator's Mac. It filters per binary, and every rebuild
produces a new unsigned binary that it treats as a new application — connections were denied
or held for a prompt a headless process cannot answer, while `curl` (already allowed) always
worked. That also accounts for the earlier `govulncheck` failure. Resolved with a path-keyed
allow rule for `dashboard.deutero.ai`; cold, cache-less runs of the staged binary then took
0.3–1.7s.

Lesson for the next run on this machine: some of the "fast" timings measured before this was
found were response-cache hits (0.03s is not a network round trip). Timings recorded in this
proof after the fix were taken with an isolated, empty HOME.

The two remaining probe timeouts are `flow coverage` (per-interview fan-out) and `study diff`
(two full eight-section reads) — the two highest round-trip commands — against a host that
cold-starts at ~4s. Both complete correctly when run directly (1.7s and 4.9s).
