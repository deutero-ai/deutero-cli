# Deutero Study Management API — CLI Brief

*Reprint of the 2026-07-20 print (run `20260719-230500`, printing-press v4.29.0). Research
redone from scratch against the current surface; the prior brief is **not** reused — it
predates three whole capability domains.*

## API Identity

- **Domain:** AI-moderated qualitative research. A researcher defines a study, an AI
  interviewer conducts asynchronous interviews with participants, and the platform
  aggregates, clusters and searches the responses.
- **Users:** Research ops, PMs, and agent/"vibe-coding" builders who want validated user
  feedback wired into an IDE or agent loop rather than a dashboard. Increasingly the
  *caller is an agent*, not a person — the field descriptions in this spec are written
  for an agent reading them cold, and the MCP tool surface is generated from it.
- **Data profile:** Deeply hierarchical and durable. Projects → Studies → (welcome,
  screening, characteristics, questions **or** an Interview Flow graph) → Interviews →
  Transcripts (message turns, captured variables, branch decisions) → Analysis (scale and
  option tallies, k-means clustering over response embeddings). Plus org-level Webhooks
  and Embed keys.
- **Surface:** OpenAPI 3.1, **59 paths / 78 operations / 16 tag groups**, base path
  `/study-api`, routes under `/api/v1`. Hand-authored; descriptions are real and
  agent-facing.
- **Auth:** `X-API-Key` header, or Bearer (API key **or** Stytch M2M access token). Four
  declared schemes (`APIKeyHeader`, `HTTPBearer`, `api_key`, `bearer_auth`) — two
  spellings of the same two mechanisms. No top-level `security` block; it is per-operation.

### Growth since the prior print

| | Prior print (2026-07) | Now |
|---|---|---|
| Paths / operations | 39 / 50 | **59 / 78** |
| Tag groups | 10 | **16** |

28 genuinely new routes. Three are entirely new domains:

- **Interview Flow (graph mode)** — 9 routes. A study now runs *either* a linear question
  list *or* a branching graph, selected by `interview_mode`. `get/set/patch_graph`,
  `check_graph` (validate without storing), `activate/deactivate_graph`,
  `import_questions_to_graph`, `describe_graph_nodes`, `get_flow_signals`.
- **Personas & Simulation** — 9 routes. Rehearse a study against AI personas before
  spending real participants: `create/update/delete/list_personas`, `generate_personas`,
  `run_simulation`, `get/list/delete_simulation`.
- **Webhooks** — 7 routes. Org-level endpoints over one event catalogue, plus delivery
  history and secret rotation: `create/update/delete/list_webhooks`,
  `list_webhook_event_types`, `list_webhook_deliveries`, `rotate_webhook_secret`.

Plus attribution/observability: `find_interviews_by_external_id`,
`get_interview_fetches_and_signals`, `validate_questions`.

## Reachability Risk

**None — and this is a change from the prior print.** The study-api is now deployed to
production.

Measured (an initial `curl -m 12` returned nothing, which was a Cloud Run **cold start**,
not an outage — `probe-reachability` then saw 4083 ms on the first request and 296 ms on
the second; a warm curl returns in 0.27 s):

- `GET https://dashboard.deutero.ai/study-api/health` → **200** `application/json`
  `{"status":"ok","service":"study-api","version":"1.0.0"}`
- `GET /study-api/api/v1/projects` unauthenticated → **401** with an actionable body:
  `"Missing authentication credentials. Provide an API key in the X-API-Key header, or an
  API key or Stytch M2M access token as a Bearer token."`
- `cli-printing-press probe-reachability` → `mode: standard_http`, confidence 0.95,
  `needs_browser_capture: false`, `needs_clearance_cookie: false`. Plain stdlib HTTP is the
  correct runtime; no Surf, no clearance cookie, no browser transport.
- `deutero.ai` → NXDOMAIN (the apex is not used; `dashboard.deutero.ai` is the host).

The prior brief recorded prod 404ing in July ("study-api is not yet deployed to prod") and
made a persistent host override **load-bearing** as a result. That pressure is gone, but
the feature is still worth keeping as ordinary good practice — the spec has **no `servers`
block**, so the CLI must carry its own default. Precedence: `--host` flag > `DEUTERO_HOST`
env > config file > default `https://dashboard.deutero.ai/study-api`.

**Cold-start implication for the build:** the first request after idle can take ~4 s. The
generated client's default timeout must comfortably exceed that, and the Phase 18 live
matrix should not read a cold-start latency as a failure.

Live testing for this run needs only an API key; the host is now a known-good default.

Spec resolution therefore did **not** come from the network: the spec was dumped directly
from the source of truth, `study_api.main:app.openapi()`, in the local repo. That is
strictly better than a fetched copy — it cannot be stale relative to the code.

## Top Workflows

1. **Author a study end to end.** Project → study → welcome/consent (+ translations) →
   screening questions with qualifying options → characteristic questions → the interview
   question list (text/scale/choices/multi_select/ranking/slots/image) → recruitment quota
   and link. ~15 dashboard screens; it should be one file and one command.
2. **Author, validate and activate an Interview Flow.** The branching alternative to the
   question list. `check_graph` validates without storing; `set_graph` is atomic (invalid
   flows are never stored); `activate_graph` switches the study into graph mode.
   `import_questions_to_graph` lifts an existing linear list into a graph as a starting
   point. This is the single biggest change since the last print.
3. **Rehearse before fielding.** Generate or write personas, `run_simulation`, poll
   `get_simulation`, read the resulting transcripts — catching broken questions, leading
   wording, and branches that never fire *before* recruiting anyone. Simulations
   deliberately mirror interviewer runtime semantics (including executing `context_fetch`
   and `webhook` steps), so a rehearsal is worth trusting.
4. **Field and monitor.** Participation link, `get_study_stats` (starts/completions/quota
   fill), `list_interviews` filtered by completed/simulated/date/external id, spot when the
   quota fills.
5. **Read and analyse the data.** `bulk_transcripts`, per-interview `get_transcript`,
   `search_transcripts` (string / semantic / hybrid), scale and option tallies, optimal-k
   estimation, k-means clustering over response embeddings.
6. **Join results back to the researcher's own systems.** `external_participant_id` is the
   researcher's own id for a person — it rides every interview-shaped response, filters
   `list_interviews` / `bulk_transcripts` / `search_transcripts`, has its own lookup, and
   travels on the `interview.started` / `interview.completed` webhooks. Webhook + redirect
   URL templating is how a completed interview triggers a payout, credit, or CRM update.

## Table Stakes

The honest competitive picture is unusual and must shape the thesis.

- **The incumbent is Deutero's own MCP server.** `asgi.py` generates an MCP tool per route
  function from this exact OpenAPI spec (`_build_mcp_names`), so an agent already has all
  78 operations with typed inputs and no registration step. A CLI that merely mirrors
  endpoints is **strictly worse than what ships today** — this is the central design
  constraint of the reprint.
- `dovetail-mcp` (TypeScript, ~12 tools) — qualitative-research MCP for the adjacent
  incumbent product. Establishes the category expectation: search, tag, export.
- `QualiGPT` (Python, ~3 commands) — LLM-assisted qualitative coding. Establishes the
  expectation of themes/coding over transcripts.
- Generic table stakes any research tool is expected to have: export to CSV, full-text
  search, filter by participant attribute, per-question tallies.

No competing Deutero CLI exists, published or otherwise; the prior print never reached the
public library (absent from a 519-entry registry). Novelty score against the open
ecosystem remains effectively 0 — the bar to clear is the in-house MCP, not a rival CLI.

## Data Layer

The API is deliberately normalized and offers **no joins**; almost every interesting
research question is a join. That gap is the whole opportunity for a local store.

- **Primary entities:** `projects`, `studies`, `questions`, `interviews`, `messages`
  (turns), `variables` (captured answers), `decisions` (branch routing), `personas`,
  `simulations`, `characteristics`, `qualifications`, `webhooks`, `webhook_deliveries`,
  `embed_keys`.
- **Sync cursor:** `list_interviews` supports `started_after` / `started_before` plus
  `limit`/`offset` — a clean incremental cursor on interview start time.
  `bulk_transcripts` paginates on `limit`/`offset` only.
- **FTS/search:** transcript message content is the natural FTS corpus; the server also
  offers semantic and hybrid search, so the local store complements rather than replaces
  `search_transcripts` (offline, regex, SQL-composable).

**The asymmetry that makes a local store pay for itself:** `TranscriptOut` (per-interview
`get_transcript`) carries `messages` **+ `variables` + `decisions`**. `InterviewTranscript`
(bulk) carries `messages` + `variables` but **never `decisions`**. Branch-routing data is
therefore only reachable one interview at a time — N+1 calls. Any question about how a
*flow* behaved across a study is impossible in one API call and trivial once synced.

Relevant field shapes (all real, from the spec):

- `CapturedVariableOut`: `name`, `value`, `value_type`, `source` (`question`/`extract`/
  `fetch`), `node_id`, `node_visit`, `question_id`, `producer`, `updated_at`.
  `question_id` is populated for linear studies, `node_id` for graph studies — so one
  table serves both modes.
- `DecisionOutcomeOut`: `node_id`, `node_visit`, `mode` (`llm_classifier` /
  `deterministic`), `classes`, `raw_label`, `matched_class`, `chosen_node_id`,
  `used_default`, `error`, `decided_at`. The spec's own description of `used_default`
  says it is "worth checking when a branch looks like it fired the wrong way" — the API
  authors are naming an analysis nobody can currently run.
- `TranscriptMessage`: `question_number` (linear ordering) and `node_id` (graph
  trajectory) — the two modes' positional keys.

## Codebase Intelligence

Source: the repository itself (`CLAUDE.md` + `docs/` shared-table contracts), which is
better evidence than DeepWiki for a private product.

- **Auth:** `X-API-Key` header or Bearer; canonical env var `DEUTERO_API_KEY` (carried
  forward from the prior print's manifest). `study_api/dependencies.py` holds the
  `require_*` authz dependencies; `get_db` commits on success so routers only `flush()`.
- **Two node-id namespaces — the sharpest correctness trap in this API.** The compiler
  names IR nodes after the editor canvas (`n3`); the author's own step id survives only in
  `properties._authoring_id`. Interview rows (`messages.node_id`, `interview_variables`,
  `node_fetches`, `decision_outcomes`) carry the **compiled** id, and every caller-facing
  surface is supposed to translate back via `services/graph_authoring.authoring_id_map`.
  **Reporting a compiled id to a researcher is a bug** — it names something that appears
  nowhere in their flow. Any CLI feature that groups or labels by `node_id` inherits this
  obligation, and must be verified against a real graph study rather than assumed.
- **What an interview leaves behind** (`docs/interview-variables.md`,
  `docs/decision-outcomes.md`, `docs/interviewer-signals-and-fetches.md`): `interview_variables`
  is upsert-keyed on `(interview_id, name)` — the **end state**, not a change log, so a
  looped step yields a merged value with `node_visit > 1`. `decision_outcomes` is
  upsert-keyed on `(interview_id, node_id, node_visit)` — one row per pass, so a Temporal
  retry never double-counts. All are best-effort projections: a write failure logs and the
  interview carries on, so **absence of a row is not proof the step did not run**. Any
  feature reporting coverage must say "not observed", never "did not happen".
- **Rate limiting:** none documented in the spec.

## User Vision

Verbatim from the reprint request:

> "Update the deutero cli to cover the added API/MCP routes."

Read against the measured delta, this means: the prior CLI knows 50 operations and the API
now has 78, so a third of the surface — including all of graph mode, simulation and
webhooks — is simply missing. The reprint must (a) cover the new routes, and (b) revisit
novel features whose premises the new domains invalidated.

## Prior novel features — reconciliation input

Carried into the Phase 1.5 novel-features subagent as reprint reconciliation input, not
adopted verbatim:

| Prior feature | Command | Status going in |
|---|---|---|
| Answer matrix export | `answers matrix` | **Reframe.** Built from message streams; `variables` is now the better source and works in both modes (`question_id` for linear, `node_id` for graph). |
| Cross-tab | `crosstab` | **Keep, re-scope.** Characteristics still exist; segmentation still unserved by the API. |
| Participation funnel | `funnel` | **Reframe or drop.** Assumes a linear list with monotonic `question_number`. In graph mode the trajectory branches, so "which question bled participants" needs redefining over node trajectories + `decisions`. |
| Saturation curve | `saturation` | **Keep.** Chronological new-term counting over message content is mode-independent. |
| Fielding burn rate | `fielding` | **Keep.** `get_study_stats` + interview start times still support it. |

## Product Thesis

- **Name:** `deutero-pp-cli` (binary name preserved from the prior print — reprints keep
  the name).
- **Why it should exist:** Not to expose the API — the in-house MCP already does that,
  generated from this same spec, with typed inputs and zero registration. It exists to
  answer the questions the API *cannot answer in one call*, because they are joins across
  endpoints or aggregations the server never computes:
  - the interview × question rectangle (no endpoint returns it)
  - segmentation by participant characteristic (tallies come back ungrouped)
  - **how a flow actually behaved across a study** — which arms never fired, how often the
    `otherwise` path was taken, where the classifier errored. `decisions` exists per
    interview and never in bulk, so this is N+1 calls today and one local query after sync.
  - what a flow's side steps actually *did* — failed `context_fetch` calls and undelivered
    `webhook` signals, which `get_interview_fetches_and_signals` also exposes only one
    interview at a time.

  **Correction (verified against the spec after the absorb gate):** an earlier draft of this
  brief listed per-participant payout reconciliation — joining webhook deliveries to
  `external_participant_id` — as a thesis pillar. It is **not constructible from this API**.
  `WebhookDeliveryOut` carries only `id`, `event_type`, `msg_id`, `status_code`, `success`,
  `error_message`, `created_at`; there is no participant identifier on a delivery record.
  Per-step `SignalDeliveryOut` *does* carry `node_id`/`msg_id`/`dry_run`, which is why
  flow-signal auditing survives while payout reconciliation does not.
  Plus what a CLI gives that an MCP tool cannot: an offline corpus that survives the
  session, SQL composability, `--json`/`--select` output an agent can narrow, typed exit
  codes, and `--dry-run` on every mutation.

## Build Priorities

1. **Data layer + sync** for the full entity set above, with `decisions` and `variables`
   captured during transcript sync (the N+1 is paid once, at sync time). This is the
   foundation every transcendence feature stands on.
2. **Full endpoint coverage — all 78 operations**, including the three new domains. This is
   the literal ask.
3. **Graph-mode correctness plumbing:** authoring-id translation wherever `node_id` is
   surfaced or grouped, and honest "not observed" semantics for best-effort projections.
4. **Transcendence features** decided at the absorb gate, biased toward: flow/branch
   coverage, simulate-then-field diffing, webhook delivery reconciliation, and the
   reframed answer rectangle.
5. **Host configuration** — the spec has no `servers` block, so the CLI carries the default
   `https://dashboard.deutero.ai/study-api` and supports persistent override
   (`--host` > `DEUTERO_HOST` > config file > default). No longer load-bearing now that prod
   answers, but still required because the spec supplies no base URL. Set the client
   timeout above the ~4 s cold-start ceiling.
