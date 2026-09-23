# Phase 4.95 — Local code review

Three reviewers (correctness, security, maintainability) dispatched via the Agent tool
against the 12 hand-written files. `/review` was deliberately not used: it is PR-shaped and
there is no PR at this stage.

**28 findings. 24 fixed, 4 accepted with reasons. No finding was deferred silently.**

## Fixed — correctness

| Finding | Fix |
|---|---|
| `deuteroFetchBundle` never returned an error, so a bad study id or expired token made `study diff` print "No differences" and exit 0 | Errors when every section fails; `buildStudyDiff` now reports `unread_sections` and refuses `same: true` when either side is partial |
| Human diff output silently dropped every entry in array-rooted sections — header said `questions (14)`, body showed nothing | Render filter now matches `section[` as well as `section.` |
| `fielding` divided by the `--window` flag while counting only active days; a 3-day-old study under `--window 14` reported a rate ~4.6× too low and a fill date weeks late | Divides by the elapsed span inside the window; field renamed and comment corrected; test added where the two denominators differ |
| Cancellation raced on `failed` and leaked goroutines writing into results and the progress writer after return (3 fan-outs) | Sets a `cancelled` flag, breaks, waits, then reports |
| Pagination stopped on the requested page size, so a server that caps `limit` would silently truncate the cohort | Compares against the server's echoed `limit` |
| Per-interview fetch failures discarded at two call sites, silently shrinking crosstab denominators | `details_failed` threaded into the report and printed |
| `rehearse diff` branch-arm shares used a global denominator across all branches, making deltas artifacts of branch mix | Per-branch denominator (`armTotalByNode`) |
| `study apply --confirm` that wrote nothing exited 0 | Prints the plan, then returns a non-nil error when any section failed |
| `answers matrix` was first-wins per question and a variable named `q3` suppressed question 3 | Variables namespaced `var:<name>`; multiple participant turns joined |
| Fan-out failures reported a count with no reason | First error retained and surfaced; total failure is now an error, partial stays non-fatal |

## Fixed — security

| Finding | Fix |
|---|---|
| `study bundle` is annotated `mcp:read-only` yet exposed `--out`, which is **not** on the MCP destination-flag blocklist — a prompt-injected model could truncate `~/.zshrc` via an auto-approved tool | Flag renamed to `--output`, which is on the blocklist |
| A flow step's `headers` dict (where a researcher puts their own backend's `Authorization`) rode `GET /graph` into bundle and diff output | Secret-shaped keys redacted by default in both; `--include-secrets` to opt out |
| CSV formula injection: participant free text goes straight into `answers.csv` for Excel | Leading `= + - @ tab CR` prefixed with `'` in both CSV writers |
| `reconcile` delivery paging was unbounded and could not terminate against a server ignoring `offset` | Bounded by `--limit × 10`, breaks on an empty page, reports progress |
| Raw API error bodies bypassed the repo's own redaction | Routed through `cliutil.SanitizeErrorBody` |

## Fixed — maintainability

Dead `result` type and its `_ = result{}` prop; dead `answersCell` alias; `_ = n` no-op;
redundant re-sort in `fielding`; `cap` parameter shadowing the builtin (renamed `limit`);
`IncompleteInterview` field not matching its JSON tag; header comment that claimed the structs
hold only read fields; `study apply` addressing steps by positional index (now by section name).

## Test coverage added

10 new tests, each pinning a fixed bug: the dry-run/delivered split in `buildFlowEffects`
(previously untested, and the file's own comment calls it easy to get wrong), crosstab's
linear question-number path, the `var:q3` vs `q3` collision, multi-turn joining, the fielding
denominator, partial-read diffs, `looksLikeBundlePath`, CSV neutralisation, secret redaction,
and per-branch arm shares. Three vacuous assertions that would pass on an empty result were
given length checks first.

`go vet` clean. `go test -race ./internal/cli/` clean (161s).

## Accepted, not fixed

1. **`reconcile`'s join is org-wide, not study-scoped.** A webhook endpoint is org-level and
   `external_participant_id` is a researcher-chosen key, so a delivery from another study on
   the same endpoint can mark this study's completion as notified. **Not fixable from this
   API** — delivery records carry no study or interview id. Documented prominently in the
   command's own help so the operator can choose unique ids or a per-study endpoint.
2. **The three fan-out helpers remain separate** rather than one generic. The reviewer's case
   is sound and the drift it predicted was real (the dead `result` type existed in exactly one
   copy). Deferred deliberately: collapsing them is a structural change to the most
   concurrency-sensitive code in the CLI, at a point where the environment cannot run the live
   matrix to prove the result. Filed as a retro/polish candidate.
3. **`--csv` is accepted but ignored by nine commands.** Their reports are nested, not
   rectangular. Better handled in polish, either with real CSV writers for `reconcile` and
   `fielding` or an explicit `usageErr`.
4. **`--limit` has no upper clamp.** A pathological `--limit 5000000` would hold the corpus in
   memory. Low risk against a 500 default; noted for polish.
