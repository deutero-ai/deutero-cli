# Absorb Manifest — deutero-pp-cli (reprint)

Run `20260921-221354-646207bd`. Spec: 59 paths / 78 operations / 16 tag groups, dumped from
`study_api.main:app.openapi()`.

**The incumbent is Deutero's own MCP server**, which generates one tool per route function
from this exact spec (`asgi.py::_build_mcp_names`). Endpoint mirroring is therefore table
stakes, not value — an absorbed row only earns its place by adding offline persistence,
agent-shaped output, `--dry-run`, or typed exit codes on top.

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-------------|--------------------|-------------|
| 1 | List/create/get/update projects | Deutero MCP (generated from spec) | `(generated endpoint) projects list/create/get/update` | Offline cache, `--json`/`--select`, typed exit codes |
| 2 | Create/get/update study, list studies in project | Deutero MCP | `(generated endpoint) studies create/get/update/list` | `--dry-run` on mutations, stdin batch |
| 3 | Get/set welcome & consent | Deutero MCP | `(generated endpoint) welcome get/set` | Local diff before write |
| 4 | Welcome translations list/upsert/delete | Deutero MCP | `(generated endpoint) welcome translations list/upsert/delete` | Bulk locale management |
| 5 | Screening: get, questions CRUD, reorder | Deutero MCP | `(generated endpoint) screening get/questions create/update/delete/reorder` | Batch author from file |
| 6 | Screening settings (incl. screen-out redirect) | Deutero MCP | `(generated endpoint) screening settings` | `redirect_url_warning` surfaced, not swallowed |
| 7 | Characteristics: get, questions CRUD, reorder | Deutero MCP | `(generated endpoint) characteristics get/questions create/update/delete/reorder` | Batch author from file |
| 8 | Characteristics settings | Deutero MCP | `(generated endpoint) characteristics settings` | Scriptable |
| 9 | Credit balance | Deutero MCP | `(generated endpoint) credits balance` | Scriptable quota/cost checks |
| 10 | Interview questions CRUD + reorder | Deutero MCP | `(generated endpoint) questions list/create/update/delete/reorder` | Offline copy of the instrument |
| 11 | Validate questions before write | Deutero MCP `validate_questions` | `(generated endpoint) questions validate` | Fails closed in scripts via exit code |
| 12 | Interview Flow: get/set/patch graph | Deutero MCP | `(generated endpoint) graph get/set/patch` | Atomic writes; flow document round-trips |
| 13 | Interview Flow: check (validate without storing) | Deutero MCP `check_graph` | `(generated endpoint) graph check` | Gate a flow in CI before activating |
| 14 | Interview Flow: activate/deactivate | Deutero MCP | `(generated endpoint) graph activate/deactivate` | Mode switch scriptable |
| 15 | Interview Flow: import linear questions to graph | Deutero MCP | `(generated endpoint) graph import-questions` | One-shot migration |
| 16 | Interview Flow: node type catalogue | Deutero MCP `describe_graph_nodes` | `(generated endpoint) graph node-types` | Cached locally for authoring |
| 17 | Interview Flow: per-step signal webhooks + secrets | Deutero MCP `get_flow_signals` | `(generated endpoint) graph signals` | Secrets read back safely |
| 18 | Recruitment get/update (quota, link, redirect) | Deutero MCP | `(generated endpoint) recruitment get/update` | `{{external_participant_id}}` templating preserved |
| 19 | Embed keys list/create/update | Deutero MCP | `(generated endpoint) embed keys list/create/update` | Origin scoping scriptable |
| 20 | Embed snippet for a study | Deutero MCP | `(generated endpoint) embed snippet` | Snippet to stdout for piping into a page |
| 21 | Personas CRUD | Deutero MCP | `(generated endpoint) personas list/create/update/delete` | Persona library reusable offline |
| 22 | Generate personas | Deutero MCP | `(generated endpoint) personas generate` | Seeded from study context |
| 23 | Run / get / list / delete simulation | Deutero MCP | `(generated endpoint) simulations run/get/list/delete` | Poll to completion, typed exit codes |
| 24 | Get interview; list interviews with filters | Deutero MCP | `(generated endpoint) interviews get/list` | Local corpus, offline re-read |
| 25 | Get one transcript (messages + variables + decisions) | Deutero MCP | `(generated endpoint) interviews transcript` | The only source of `decisions` |
| 26 | Study stats (starts/completions/quota) | Deutero MCP | `(generated endpoint) studies stats` | Cached for trend math |
| 27 | Find interviews by external participant id | Deutero MCP | `(generated endpoint) interviews by-external-id` | Join key to the researcher's own users |
| 28 | Interview fetches & signals | Deutero MCP | `(generated endpoint) interviews fetches-and-signals` | Integration debugging per interview |
| 29 | Bulk transcripts | Deutero MCP | `(generated endpoint) transcripts bulk` | Paginated sync into local store |
| 30 | Transcript search (string / semantic / hybrid) | Deutero MCP; dovetail-mcp `search_insights` | `(generated endpoint) search` | Server semantic search preserved |
| 31 | Full-text search over the corpus | dovetail-mcp (category expectation) | `(behavior in deutero-pp-cli search)` | **Corrected during Phase 3:** transcript messages are not syncable (study-scoped), so this is the server's string/semantic/hybrid search with agent-native output and typed exit codes — not an offline index |
| 32 | Analysis: scale + options tallies, analyzable questions | Deutero MCP | `(generated endpoint) analysis responses scale/options, analysis questions` | Cached tallies |
| 33 | Clustering: run, optimal-k, latest | Deutero MCP | `(generated endpoint) analysis cluster/optimal-clusters/clustering-latest` | Stored run history |
| 34 | LLM-assisted thematic grouping | QualiGPT | `(behavior in deutero-pp-cli studies analysis run-clustering)` server-side embeddings | No extra LLM key needed |
| 35 | Webhooks CRUD | Deutero MCP | `(generated endpoint) webhooks list/create/update/delete` | `--dry-run` before wiring |
| 36 | Webhook event catalogue | Deutero MCP | `(generated endpoint) webhooks event-types` | Discover payload fields offline |
| 37 | Webhook delivery history | Deutero MCP | `(generated endpoint) webhooks deliveries` | Paginated failure triage |
| 38 | Rotate webhook signing secret | Deutero MCP | `(generated endpoint) webhooks rotate-secret` | Secret shown once, never persisted |
| 39 | Health check | Deutero MCP | `(generated endpoint) health` | Wired into `doctor` |
| 40 | Export anything to CSV | dovetail-mcp export; Dedoose | `(behavior in deutero-pp-cli answers matrix)` plus the global `--csv` flag on every command | Every command exports, not one |
| 41 | Incremental sync of org-level resources | No competitor has this | `deutero-pp-cli sync` | **Corrected during Phase 3:** covers projects, webhooks, webhook deliveries, embed keys, graph node types — the flat and parent-keyed resources sync can reach. Interviews, transcripts and simulations are study-scoped and are read live by the analysis commands instead |

## Transcendence (only possible with our approach)

Eight survivors from the novel-features subagent (16 candidates → 8 kills). Audit trail,
customer model and kill reasons: `2026-09-21-221942-novel-features-brainstorm.md`.

| # | Feature | Command | Score | Buildability | Why Only We Can Do This | Long Description |
|---|---------|---------|-------|--------------|------------------------|------------------|
| 1 | Flow behavior coverage | `flow coverage --study <id>` | 10/10 | hand-code | `decisions` is returned by per-interview `get_transcript` and **never** by bulk transcripts, so study-level branch behaviour is impossible in one call. **Corrected during Phase 3:** `sync` cannot reach interviews or transcripts (they are scoped to a study id it cannot supply), so the command pays the N+1 itself — bounded, concurrent and reported — rather than reading a local store. Ids translated to authored step ids; unseen arms labeled "not observed" | Use this command for how an Interview Flow routed and what it captured across a whole study — branch arms, `otherwise` rate, classifier errors, variable capture. Do NOT use this command for what a flow's steps did outside the conversation (data fetches, signal deliveries); use `flow effects` instead. Do NOT use it to get per-participant answers as a table; use `answers matrix` instead. |
| 2 | Answer rectangle | `answers matrix --study <id> --csv` | 8/10 | hand-code | No endpoint returns the interview × question rectangle. **Corrected during Phase 3 against the spec:** `variables` is "Empty for a study that captures nothing, including every linear study", so the rectangle is built from `variables` for flow studies **and from the message stream (`question_number` + content) for linear studies** — reading only one source would return nothing for half the platform's studies | Use this command to get the interview × question rectangle for counting, correlating or handing to a spreadsheet. Do NOT use this command to break one question down by a participant characteristic; use `crosstab` instead. |
| 3 | Flow effects audit | `flow effects --study <id>` | 8/10 | hand-code | `get_interview_fetches_and_signals` is per-interview with no study-level equivalent; aggregating it needs the local store. `SignalDeliveryOut.dry_run` must be separated or a simulation's un-sent signals get counted as delivered | Use this command for what an Interview Flow's `context_fetch` and `webhook` steps actually did — failed fetches, undelivered signals, dry-run signals from simulations. Do NOT use this command for how the flow routed or what it captured; use `flow coverage` instead. Do NOT use it for org-level webhook endpoint delivery history; that is the generated `webhooks deliveries` command. |
| 4 | Study bundle export/apply | `study bundle --study <id>` / `study apply --file <f>` | 8/10 | hand-code | Collapses ~15 ordered endpoint calls (study, welcome + translations, screening + settings, characteristics + settings, questions or flow, recruitment) into one file that lives in a repo and survives the agent session; `apply` replays with `--dry-run` diffing first | Use this command to snapshot or replay an entire study definition as one file. Do NOT use this command to read collected responses; use `answers matrix` instead. |
| 5 | Rehearsal vs reality diff | `rehearse diff --study <id>` | 7/10 | hand-code | Nothing server-side compares the simulated cohort against the real one; `simulated` filters plus one local table make step reach, branch mix, variable capture and duration directly comparable | Use this command to compare simulated (persona) interviews against real ones on the same measures. Do NOT use this command to inspect one simulation run's status or credits; the generated `simulations get` command does that. |
| 6 | Fielding burn rate | `fielding --study <id>` | 7/10 | hand-code | `StudyStatsOut` and `RecruitmentOut` expose totals and quota but no rate anywhere in 78 operations; the local time series of interview start times supplies it | none |
| 7 | Cross-tab | `crosstab --study <id> --by <characteristic> --question <id>` | 7/10 | hand-code | Scale/options tallies come back as flat value→count maps with no grouping parameter, while the platform collects characteristics specifically for segmentation | Use this command to answer "does this differ by segment". Do NOT use this command when you want every answer for every participant; use `answers matrix` instead. |
| 8 | Saturation curve | `saturation --study <id>` | 6/10 | hand-code | No endpoint has any concept of saturation; first-appearance term counting over the chronologically ordered local corpus is mode-independent | none |
| 9 | Payout reconciliation | `reconcile --study <id> --webhook <id>` | 9/10 | hand-code | **Unblocked by a repo fix made during this run.** `WebhookDeliveryOut` now carries `external_participant_id`, so completed interviews can be joined to the deliveries that were supposed to notify somebody. Still impossible upstream: it spans two endpoint families with no server-side join, and the "completed but never successfully delivered" set is a local anti-join | Use this command to find participants who completed a study but were never successfully notified — the gap between what finished and what your payout job heard about. Do NOT use this command to inspect one endpoint's raw delivery log; the generated `webhooks deliveries` command does that. Do NOT use it for interview-time flow signals; use `flow effects` instead. |
| 10 | Study definition diff | `study diff --study <id> --against <id-or-file>` | 7/10 | hand-code | Compares two complete study definitions — details, welcome + translations, screening + settings, characteristics + settings, questions or flow, recruitment — field by field. Nothing upstream diffs anything; each half is ~7 calls, and the comparison only exists once both are local | Use this command to see what differs between two studies, or between a study and a saved bundle file, before replaying one over the other. Do NOT use this command to write changes; `study apply` does that, and shows its own dry-run diff first. |

### Gate amendments (added by the user at Phase Gate 1.5)

The user reviewed the showcase and added two rows, taking the transcendence set from 8 to 10.

- **Row 9, `reconcile` — un-killed.** The subagent killed it because `WebhookDeliveryOut`
  carried no participant identifier, making the join unconstructible. The user identified
  that absence as **a bug in the API, not a constraint on it**, and it was fixed in the
  repo during this run (migration + model column + delivery-time capture + schema field +
  query filter + tests). The feature is back as shipping scope. Note for Phase 18: the
  deployed production build predates the fix, so the field reads null live until the
  migration is run and the service redeployed — the command must treat a null
  `external_participant_id` as "this delivery names nobody", not as a failed join.
- **Row 10, `study diff` — added.** From the user's "authoring assistance" answer. Pairs
  with the already-surviving `study bundle`/`study apply`.

`flow lint` (unreachable steps, never-satisfiable variables) was considered for the same
answer and **not** added: `check_graph` already validates server-side and variable capture
is already a section of `flow coverage`, so it would have duplicated two existing surfaces.

### Frustrations the user named (build constraints, not features)

These came from the gate and bind Phase 3 and Phase 11:

1. **Pagination / N+1 cost.** Bulk transcripts paginate on `limit`/`offset` only, and
   `decisions` plus `fetches-and-signals` are per-interview. `sync` must therefore page
   explicitly, checkpoint so an interrupted sync resumes rather than restarting, and make
   the per-interview fan-out visible (count and progress) rather than silently issuing
   hundreds of requests. `--max-pages` must actually bound the work.
2. **Error messages / failure modes.** Upstream 4xx bodies must be surfaced verbatim in the
   CLI's error text rather than collapsed into "request failed" — the 401 body is already
   well written ("Provide an API key in the X-API-Key header, or ...") and should reach the
   user intact. Typed exit codes for auth vs not-found vs validation.
3. **Graph authoring round-trips.** `check_graph` validator errors (`var_not_guaranteed`,
   `no_default`) must be printed in full with the offending step named by its **authored**
   id, since that is the id the researcher can act on.

### Dropped prior features (reprint surface — override at the gate if you disagree)

| Prior feature | Command | Verdict | Justification |
|---|---|---|---|
| Participation funnel | `funnel` | **Drop** | Premise is a monotonic `question_number`, which graph mode abolishes; the useful half ("which step interviews last reached", by authored id) is absorbed into `flow coverage` |
| Answer matrix export | `answers matrix` | **Reframe** | Command name retained, data source changed: synced `variables` instead of parsed message streams, so it works in graph mode too |
| Cross-tab | `crosstab` | Keep | Unchanged premise |
| Saturation curve | `saturation` | Keep | Mode-independent |
| Fielding burn rate | `fielding` | Keep | Unchanged premise |

## Stubs

**None.** Every row above is shipping scope.

## Build notes carried into Phase 3

- Data-source annotations: all transcendence commands are `// pp:data-source local` except
  `study bundle` / `study apply`, which are `// pp:data-source live` and must reject
  `--data-source local`.
- Every local command calls `hintIfUnsynced` / `hintIfStale` before returning results
  (`"interviews"` for `fielding`, `rehearse diff`, `flow effects`; `"transcripts"` for
  `flow coverage`, `answers matrix`, `crosstab`, `saturation`).
- Drain the parent `*sql.Rows` into structs before any authoring-id resolution or child
  lookup (SQLite single-connection constraint).
- Commit any custom-table write transaction before calling `store.Upsert`/`UpsertBatch`.
- **Authoring-id translation is a correctness requirement, not a nicety.** Anything that
  groups or labels by `node_id` must present the authored step id.
- **"Not observed", never "did not happen."** `variables`, `decisions`, `node_fetches` and
  `signal_deliveries` are best-effort projections; a write failure logs and the interview
  continues.
