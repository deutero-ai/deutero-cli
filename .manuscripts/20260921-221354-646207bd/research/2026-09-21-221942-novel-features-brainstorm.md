# Novel-features brainstorm — deutero (reprint)

Run `20260921-221354-646207bd`. Subagent output, verbatim. Customer model and killed
candidates are retained here for retro/dogfood debugging; they do not enter the manifest.

## Customer model

**Priya — research ops lead running 6-10 studies at once for a product org.**

*Today (without this CLI):* She keeps the Deutero dashboard open on the studies list, a second tab on one study's Responses view, and a third on a Google Sheet she maintains by hand. When a PM asks "how's the enterprise study doing," she reads `get_study_stats` off the dashboard (starts, completions, quota fill) and eyeballs whether it is tracking. When a PM asks "will it fill by Thursday," she cannot answer — nothing in the product or the API exposes a rate, only totals. She cannot answer "which segment is under-recruited" either, because `get_scale_responses` / `get_options_responses` return ungrouped counts and the characteristics live on a different endpoint.

*Weekly ritual:* Monday morning, walk every live study: completions since last week, quota remaining, whether anything stalled, and whether she can stop fielding one of them. Friday, hand a PM a table of answers by segment.

*Frustration:* Every question she is actually asked is a join — completions over time against quota, answers against characteristics — and the API offers no joins, so she rebuilds the same spreadsheet weekly from CSV exports that arrive as message streams rather than rows.

**Marco — the flow author who moved his studies into graph mode.**

*Today (without this CLI):* He has the LiteGraph editor open, `describe_graph_nodes` output in a scratch file, and the transcript view open on one interview at a time. When a Smart Branch behaves oddly, he opens interviews one by one and reads the routing — because `get_transcript` carries `decisions` and the bulk endpoint does not. To ask "did the `otherwise` path fire more than it should have across 80 interviews," he would have to make 80 calls and tally by hand, so he does not ask it. He also cannot tell which arm of a branch has never been taken, or which declared variable is never getting captured, without reading everything.

*Weekly ritual:* Edit the flow, `check_graph`, `set_graph`, run a simulation or two, then — once real interviews land — spot-check a handful of transcripts to confirm the branching is doing what he intended.

*Frustration:* `decisions` is per-interview only. The one question that matters about a branching flow — how it behaved *in aggregate* — is N+1 calls today, and the spec itself nudges him toward an analysis (`used_default` "worth checking when a branch looks like it fired the wrong way") that nothing lets him run.

**Dana — an agent/vibe-coding builder wiring Deutero into a product loop.**

*Today (without this CLI):* Her agent has the in-house MCP server with all 78 tools. Authoring a study end to end is ~15 tool calls that must be issued in the right order, and there is no way to snapshot the result: to clone a working study into a second one, the agent re-derives welcome, screening, characteristics, questions, graph and recruitment call by call, and drifts. She also wires `interview.completed` webhooks to a payout job and then has no way to check that everything which completed actually got delivered — `list_webhook_deliveries` shows status codes, and `list_interviews` shows completions, and nothing puts them side by side.

*Weekly ritual:* Spin up or amend a study from a config file in her repo, field it behind an embed key, and reconcile last week's completions against her own payouts table.

*Frustration:* Nothing survives the session. The study definition lives only on the server, the delivery evidence lives only in a paginated log, and both die when the agent context is cleared.

**Sam — the PM who rehearses before spending participants.**

*Today (without this CLI):* Generates personas, runs `run_simulation` per persona, polls `get_simulation` until each settles, then opens each produced interview's transcript individually. Judges "did this work" by reading. Once real interviews arrive, he has no way to ask whether reality resembled the rehearsal — whether real participants took the same paths, whether questions the personas answered fluently are the ones real people quit on.

*Weekly ritual:* Before each launch, a rehearsal round; after each launch week, a sanity check that the study is producing usable answers.

*Frustration:* Simulations mirror interviewer runtime faithfully enough to be worth trusting, and then the comparison that would make them *evidence* — simulated cohort versus real cohort on the same measures — has to be done by eye across two sets of transcripts.

## Candidates (pre-cut)

1. **Flow behavior coverage** — `flow coverage --study <id>` — (source: a/b/f, serves Marco) Across every synced interview, per Branch step: traversals, `matched_class` distribution, `used_default` rate, classifier `error` count; per flow step: interviews that reached it; per declared variable: capture rate and `value_type`. All ids reported as authored ids; anything unseen labeled "not observed", never "did not happen". **Keep** — no LLM, no external service, read-only auth, local SQLite only.

2. **Branch classifier audit** — `flow branches --study <id>` — (source: b, Marco) Just the Smart Branch half of #1: raw_label vs matched_class, error rate. **Kill inline (sibling overlap)** — strict subset of candidate 1; two commands over one table is how agents pick wrong.

3. **Variable capture coverage** — `flow variables --study <id>` — (source: b, Marco) Per declared flow variable: how many interviews hold it, `source` mix, `node_visit > 1` (loop-merged) share. **Kill inline (sibling overlap)** — folded into candidate 1 as its variables section; same synced table, same grouping key.

4. **Answer rectangle** — `answers matrix --study <id> --csv` — (source: d `prior-reframe`, Priya) One row per interview, one column per question/variable, widened with characteristics, `external_participant_id` as the join key column. Built from synced `variables` (`question_id` for linear studies, `node_id`→authored id for graph studies) with message text as fallback for un-captured text questions. **Keep, reframed** — the prior version parsed message streams; `variables` is the correct source and is the only one that works in both interview modes.

5. **Cross-tab** — `crosstab --study <id> --by <characteristic> --question <id>` — (source: d `prior-keep`, Priya) Per-cell counts for any scale/options question segmented by any characteristic, plus per-segment completion rate; column marginals reconcile against `get_scale_responses` / `get_options_responses`. **Keep** — the API collects characteristics *for* segmentation and returns tallies ungrouped.

6. **Participation funnel** — `funnel --study <id>` — (source: d `prior-drop`, Priya) Screened-out → started → per-question abandonment → completed. **Kill inline** — its premise is a monotonic `question_number`; in graph mode the trajectory branches, so "the question that bled participants" is not well defined. Its live half (last step reached, by authored id) belongs in candidate 1's reach column.

7. **Saturation curve** — `saturation --study <id>` — (source: d `prior-keep`, Priya/Sam) Chronologically ordered corpus, counting first-appearance terms contributed by each successive interview, flagging where the curve flattens. **Keep** — purely mechanical term counting over local messages, mode-independent, no endpoint has any concept of saturation.

8. **Fielding burn rate** — `fielding --study <id>` — (source: d `prior-keep`, Priya) Completions per day from local interview `start_time`/`end_time` against `max_responses` from `get_recruitment`, projecting the fill date. **Keep** — the API exposes totals and quota, never a rate.

9. **Rehearsal vs reality diff** — `rehearse diff --study <id>` — (source: a/b, Sam) Compares the simulated cohort (`simulated=true`) against the real one on the same synced measures: step reach, branch arm distribution, variable capture rate, message count and duration per interview. **Keep** — one SQLite table, two filters, honest deltas; no LLM judgment involved.

10. **Flow effects audit** — `flow effects --study <id>` — (source: b/f, Marco/Dana) Across synced `fetches` and `signals` (from `get_interview_fetches_and_signals`, an N+1 paid once at sync): failed fetches grouped by authored step id with status codes and error text, signal deliveries by success/`attempt`, and `dry_run` deliveries called out so a simulation's un-sent signals are never counted as sent. **Keep** — this is integration debugging over data that exists only one interview at a time.

11. **Payout reconciliation** — `reconcile --study <id> --webhook <id>` — (source: e, Dana) Joins completed interviews against `interview.completed` webhook deliveries by `external_participant_id`. **Kill inline (verifiability / spec contradiction)** — `WebhookDeliveryOut` carries only `id`, `event_type`, `msg_id`, `status_code`, `success`, `error_message`, `created_at`. There is **no participant identifier on a delivery record**, so the per-participant join the brief hoped for is not constructible from this API. Its honest residue (counts and failures in a window) is absorbed into candidate 10's signal section rather than sold as reconciliation.

12. **Study bundle export/apply** — `study bundle --study <id>` / `study apply --file <f>` — (source: e, Dana) One file holding study details, welcome + translations, screening (+settings), characteristics (+settings), question list or flow document, and recruitment; `apply` replays it into a new or existing study with `--dry-run` showing the diff first. **Keep** — collapses ~15 ordered MCP calls into one artifact that lives in a repo and survives the session.

13. **Participant 360** — `participant <external-id>` — (source: b, Dana) Everything about one person: interview, transcript, variables, decisions, effects. **Kill inline (thin)** — `find_interviews_by_external_id` already returns the interview with `qualifications`, `characteristics` and `variables`; the increment over one existing endpoint does not justify a command.

14. **Cluster exemplar quotes** — `clusters quotes --study <id>` — (source: c, Priya) Joins `get_latest_clustering` data points back to local message text for verbatims per cluster. **Kill inline (verifiability)** — the point-to-message identity in `ClusterPoints`/`Centroid` cannot be confirmed from the spec alone, and picking "representative" quotes slides toward judgment rather than mechanism.

15. **Rehearsal cost report** — `rehearse cost --study <id>` — (source: b, Sam) Aggregates `SimulationOut.credits_used` by persona and `model_tier` against `get_credit_balance`. **Kill inline (frequency)** — read before a launch, not weekly; also a straight sum over one list endpoint.

16. **Launch preflight** — `preflight --study <id>` — (source: a, Dana/Sam) Chains `validate_questions` or `check_graph`, welcome/consent presence, recruitment quota, credit balance and enabled webhooks into one pass/fail. **Kill inline (wrapper)** — it is five live calls in sequence with no join and no local data; an agent holding the MCP can already chain them, and each already returns its own verdict.

## Survivors and kills

### Survivors

The full transcendence table is in the absorb manifest
(`2026-09-21-221942-feat-deutero-pp-cli-absorb-manifest.md`) — 8 survivors, scores 6–10/10,
all tagged `hand-code`.

Pass 3 force-answers, per survivor: (1) weekly for Marco, not a wrapper — no endpoint returns study-level decisions at all; transcendence from local SQLite + authoring-id translation; sibling killed: `flow branches`; `hand-code`. (2) weekly for Priya; not a wrapper — no rectangle endpoint; transcendence from local join; sibling killed: `participant`; `hand-code`. (3) weekly for Marco/Dana; not a wrapper — `fetches-and-signals` is per-interview; transcendence from local aggregation over an N+1; sibling killed: `reconcile`; `hand-code`. (4) weekly for Dana; not a wrapper — it is 15 endpoints in one artifact with a diff; transcendence from agent-shaped file output; sibling killed: `preflight`; `hand-code`. (5) weekly for Sam during fielding; not a wrapper — no cross-cohort endpoint; transcendence from local store; sibling killed: `rehearse cost`; `hand-code`. (6) weekly for Priya; not a wrapper — rate is computed, never returned; transcendence from local time series; sibling killed: `funnel`; `hand-code`. (7) weekly for Priya; not a wrapper — tallies come back ungrouped; transcendence from local join; sibling killed: `clusters quotes`; `hand-code`. (8) weekly-to-biweekly for Priya (softest of the eight, hence 6/10); not a wrapper — no saturation concept exists; transcendence from the ordered local corpus; sibling killed: `funnel`; `hand-code`.

All eight are `// pp:data-source local` except `study bundle`/`study apply`, which is `// pp:data-source live` (it reads and writes the authoritative definition and must reject `--data-source local`). Every local command calls `hintIfUnsynced(cmd, db, "<resource>")` / `hintIfStale(...)` before returning results — `"interviews"` for `fielding` and `rehearse diff`, `"transcripts"` for `flow coverage`, `answers matrix`, `crosstab` and `saturation`, `"interviews"` for `flow effects`. All local queries drain the parent `*sql.Rows` into structs before issuing authoring-id resolution or child lookups, and any custom-table write transaction commits before calling `store.Upsert`/`UpsertBatch`.

### Killed candidates

| Feature | Kill reason | Closest surviving sibling |
|---|---|---|
| Branch classifier audit (`flow branches`) | Strict subset of `flow coverage` over the same synced `decisions` table; two commands on one grouping key invite wrong tool selection | `flow coverage` |
| Variable capture coverage (`flow variables`) | Same table, same grouping key as `flow coverage`; absorbed as that command's variables section | `flow coverage` |
| Participation funnel (`funnel`) | Premise is a monotonic `question_number`, which graph mode breaks; "last step reached" survives inside `flow coverage` | `flow coverage` |
| ~~Payout reconciliation (`reconcile`)~~ **— kill overturned at the gate** | The subagent's reasoning was correct about the spec as it stood: `WebhookDeliveryOut` carried no participant identifier. The user identified that absence as a **bug in the API rather than a constraint on it**, and it was fixed in the repo during this run. `reconcile` is transcendence row 9 in the manifest. | — (now shipping) |
| Participant 360 (`participant <external-id>`) | Thin over `find_interviews_by_external_id`, which already returns qualifications, characteristics and variables | `answers matrix` |
| Cluster exemplar quotes (`clusters quotes`) | Point-to-message identity is unverifiable from the spec and "representative quote" selection is judgment, not mechanism | `crosstab` |
| Rehearsal cost report (`rehearse cost`) | Pre-launch, not weekly, and a plain sum over one list endpoint | `rehearse diff` |
| Launch preflight (`preflight`) | Five live calls chained with no join and no local data — the in-house MCP already chains them and each returns its own verdict | `study bundle` |

## Reprint verdicts

| Prior feature | Command | Verdict | Justification |
|---|---|---|---|
| Answer matrix export | `answers matrix` | **Reframe** (command name retained; data source changed) | Right output, wrong source — rebuild from synced `variables`, which carries `question_id` for linear studies and `node_id` for graph studies, instead of parsing message streams that only work in linear mode. |
| Cross-tab | `crosstab` | **Keep** | Characteristics and ungrouped tallies are unchanged in the new spec; segmentation is still unserved by all 78 operations. |
| Participation funnel | `funnel` | **Drop** | Depends on a monotonic `question_number` that graph mode abolishes; its useful half (which step interviews last reached, by authored id) is absorbed into `flow coverage`. |
| Saturation curve | `saturation` | **Keep** | Chronological first-appearance term counting over message content is mode-independent and untouched by the graph/simulation/webhook additions. |
| Fielding burn rate | `fielding` | **Keep** | `StudyStatsOut` and `RecruitmentOut` still expose totals and quota without any rate; local interview start times still supply the time series. |
