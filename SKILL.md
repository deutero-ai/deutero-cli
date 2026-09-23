---
name: pp-deutero
description: "Every Deutero endpoint, plus the flow-behaviour and fielding answers that take many API calls to assemble and that no single endpoint returns. Trigger phrases: `how did my interview flow route`, `which branch is misrouting`, `export study answers as a table`, `will my study fill by`, `have I reached saturation`, `compare my simulation to real interviews`, `use deutero`, `run deutero`."
author: "francis"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - deutero-pp-cli
    install:
      - kind: go
        bins: [deutero-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/ai/deutero/cmd/deutero-pp-cli
---

# Deutero — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `deutero-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install deutero --cli-only
   ```
2. Verify: `deutero-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.6 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/ai/deutero/cmd/deutero-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Deutero's own MCP server already exposes every endpoint, so mirroring the API is table stakes. This CLI earns its place on the questions the API cannot answer in one call: how a branching Interview Flow actually routed across a study (`flow coverage`), the interview-by-question rectangle nobody returns (`answers matrix`), what a flow's fetches and signals really did (`flow effects`), which completions your webhook never reported (`reconcile`), and whether a study will fill by Thursday (`fielding`). Branch decisions come back one interview at a time and never in bulk, so these commands fan out for you — bounded, concurrent, and with the cost shown. They read live; `--limit` caps the fan-out at 500 interviews by default (raise it to at most 10,000). `sync` is separate and covers org-level resources only (projects, webhooks, webhook deliveries, embed keys, graph node types and health) — interviews and transcripts are study-scoped and cannot be synced.

## When to Use This CLI

Reach for this CLI when the question spans more than one Deutero API call: how a branching Interview Flow behaved across a study, the interview-by-question table no endpoint returns, whether a simulated rehearsal matched real participants, which completions a webhook never reported, and when a study will hit its quota. Also when you want a study definition as a file you can review, diff and replay. For single-resource reads and writes, an agent already holding Deutero's MCP server should use that instead.

## Anti-triggers

Do not use this CLI for:
- Conducting or moderating an interview — the AI interviewer does that, not this CLI
- Recruiting participants or buying a panel
- Analysing transcripts from a source other than Deutero
- One-off single-endpoint reads when the Deutero MCP server is already connected

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Flow behavior you cannot see upstream
- **`flow coverage`** — Shows how an Interview Flow actually routed across a whole study: every branch arm's share, how often the otherwise path fired, classifier errors, which steps interviews reached, and which declared variables are being captured. Flow-mode studies only: a linear study has no branch or flow records and correctly returns nothing.

  _Reach for this when a branching interview might be misrouting: it names the arms that never fired and the share that fell through to the default, instead of making you read transcripts one at a time._

  ```bash
  deutero-pp-cli flow coverage --study f0ed9214-06fc-4420-87b3-31240b670fc1 --agent
  ```
- **`flow effects`** — Aggregates what a flow's side steps actually did across a study: failed data fetches with status codes, undelivered signals, and dry-run signals from simulations kept separate from real ones. Flow-mode studies only: a linear study has no branch or flow records and correctly returns nothing.

  _Reach for this when an interview-time webhook or data fetch looks unreliable and you need the failures grouped by step rather than buried per interview._

  ```bash
  deutero-pp-cli flow effects --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json
  ```
- **`rehearse diff`** — Compares the simulated persona cohort against real participants on the same measures: step reach, branch arm mix, variable capture rate, message count and duration. Needs both cohorts; branch and variable measures are flow-mode only.

  _Reach for this after fielding starts to check whether real participants behave like the personas you rehearsed on, before trusting the pilot._

  ```bash
  deutero-pp-cli rehearse diff --study 092103ce-0e01-4e81-8ea5-e85fa560504f --agent
  ```

### Agent-native plumbing
- **`reconcile`** — Finds participants who completed a study but were never successfully notified, by joining completed interviews against the webhook deliveries that were supposed to tell your system about them.

  _Reach for this after a payout or reward run to find the people your system was never told about, instead of trusting that every webhook landed._

  ```bash
  deutero-pp-cli reconcile --study f0ed9214-06fc-4420-87b3-31240b670fc1 --webhook 9ca6a913-54ff-495d-92f0-2baaa24ee9c6 --json
  ```
- **`study diff`** — Compares two complete study definitions field by field — details, welcome and translations, screening, characteristics, questions or flow, recruitment — or compares a live study against a saved bundle file.

  _Use this before replaying a bundle over an existing study, or to see exactly how two studies' instruments differ._

  ```bash
  deutero-pp-cli study diff --study f0ed9214-06fc-4420-87b3-31240b670fc1 --against 092103ce-0e01-4e81-8ea5-e85fa560504f
  ```
- **`study bundle`** — Snapshots an entire study definition — details, welcome and translations, screening, characteristics, questions or flow, recruitment — into one file. The write half is a separate command, 'study apply --file <f>', which plans by default and writes only with --confirm, and which replays whole-resource sections only (question lists replay via 'studies questions').

  _Use this to clone or version a study definition instead of re-deriving fifteen calls; pair it with 'study diff' to see what a replay would change before running 'study apply --confirm'._

  ```bash
  deutero-pp-cli study bundle --study f0ed9214-06fc-4420-87b3-31240b670fc1 --output study.json
  ```

### Answers the API cannot assemble
- **`answers matrix`** — Turns a study into a real table: one row per interview, one column per question or captured variable, with the external participant id as the join key and optional participant characteristic columns via --with-characteristics. Works for both linear and flow studies.

  _Use this when you need to count or correlate answers rather than read them — it produces the rectangle you would otherwise rebuild by hand from transcripts._

  ```bash
  deutero-pp-cli answers matrix --study f0ed9214-06fc-4420-87b3-31240b670fc1 --csv
  ```
- **`crosstab`** — Segments any scale or options question by any participant characteristic, with per-cell counts and per-segment completion rate.

  _Use this to answer whether an answer differs by segment — column marginals reconcile against the server's own tallies, so the numbers are checkable._

  ```bash
  deutero-pp-cli crosstab --study f0ed9214-06fc-4420-87b3-31240b670fc1 --by role --question 7 --agent
  ```

### Fielding decisions
- **`fielding`** — Completions per day against the response quota, projecting the date the quota fills.

  _Use this to answer whether a study will fill by a date, which the totals-only stats endpoint cannot tell you._

  ```bash
  deutero-pp-cli fielding --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json
  ```
- **`saturation`** — Counts how many genuinely new terms each successive interview contributes, and marks the point where the curve flattens. A mechanical word-level proxy for new content, not a judgement about meaning.

  _Reach for this when deciding whether to stop fielding — it shows when additional interviews stopped adding new language, as evidence for that call rather than the call itself._

  ```bash
  deutero-pp-cli saturation --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json
  ```

## Command Reference

**credits** — Organization credit balances, usage, and reservations.

- `deutero-pp-cli credits` — Get credit balance

**embed** — Embed-widget keys and install snippets.

- `deutero-pp-cli embed create` — Create a publishable embed key for the caller's organization, optionally pinned to one study.
- `deutero-pp-cli embed list-keys` — Every embed key owned by the caller's organization (secrets omitted).
- `deutero-pp-cli embed update` — Change allowed origins, revoke/reactivate (``status``), or adjust the metadata size limit.

**graph** — Manage graph

- `deutero-pp-cli graph` — The configuration schema for every step type an interview flow can contain

**health** — Manage health

- `deutero-pp-cli health` — Liveness check

**interviews** — Manage interviews

- `deutero-pp-cli interviews <interview_id>` — One interview with its screening answers, participant characteristics, embed-widget metadata (with per-key provenance)

**projects** — Containers that group studies.

- `deutero-pp-cli projects create` — Create a new project owned by the caller's organization.
- `deutero-pp-cli projects get` — Get a project
- `deutero-pp-cli projects list` — List every project in the caller's organization.
- `deutero-pp-cli projects update` — Update a project's name and/or description. Omitted fields are untouched.

**simulations** — Manage simulations

- `deutero-pp-cli simulations delete` — Removes the run's bookkeeping row.
- `deutero-pp-cli simulations get` — One run's current state.

**studies** — Study creation and core configuration.

- `deutero-pp-cli studies create` — Create a new study inside a project.
- `deutero-pp-cli studies get-study` — Full study configuration, including per-lifecycle-stage status (welcome configured
- `deutero-pp-cli studies update` — Partial update of study properties. Only fields present in the request body are changed.

**webhooks** — Organization-level event subscriptions: signed HTTP callbacks for interview, analysis, simulation, study and credit events. Distinct from the per-step signals an Interview Flow sends mid-interview.

- `deutero-pp-cli webhooks create` — Subscribe a URL of yours to organization events.
- `deutero-pp-cli webhooks delete` — Permanently removes the endpoint and its delivery log.
- `deutero-pp-cli webhooks list` — Every webhook endpoint in your organization, newest first.
- `deutero-pp-cli webhooks list-event-types` — Every event an endpoint can subscribe to, what makes it fire
- `deutero-pp-cli webhooks update` — Partial update: only the fields you send change.


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
deutero-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query. `--json` (and other machine formats) keep that exit-2 contract and write `{"matches":[]}` on stdout so agents can inspect the envelope without treating a miss as success.

## Recipes

### Find the branch that is misrouting

```bash
deutero-pp-cli flow coverage --study f0ed9214-06fc-4420-87b3-31240b670fc1 --agent --select branches.step,branches.used_default_rate,branches.error_count
```

Narrows a large coverage report to just the signal that matters — which branch fell through to its default most often, and whether the classifier was erroring — instead of pulling the whole nested payload into context.

### Answers as a spreadsheet, segmented

```bash
deutero-pp-cli crosstab --study f0ed9214-06fc-4420-87b3-31240b670fc1 --by role --question 7 --csv
```

Cross-tabulates one scale question by a participant characteristic; the column totals reconcile against the server's own tally endpoint, so the numbers are checkable.

### Decide whether to stop fielding

```bash
deutero-pp-cli saturation --study f0ed9214-06fc-4420-87b3-31240b670fc1 --json
```

Shows where new interviews stopped contributing new language, which is the evidence for calling a study done rather than running it to an arbitrary quota.

### Version a study definition

```bash
deutero-pp-cli study bundle --study f0ed9214-06fc-4420-87b3-31240b670fc1 --output study.json
```

Captures details, welcome, screening, characteristics, questions or flow and recruitment as one file you can commit, diff and replay into a second study.

### Check the rehearsal predicted reality

```bash
deutero-pp-cli rehearse diff --study 092103ce-0e01-4e81-8ea5-e85fa560504f --agent
```

Compares simulated persona interviews against real ones on step reach, branch mix and variable capture, so a pilot becomes evidence rather than a vibe.

## Auth Setup

Set DEUTERO_API_KEY, or run 'deutero-pp-cli auth set-token'. The CLI sends it as an Authorization: Bearer header; the API accepts either an API key or a Stytch M2M access token there. (The API also accepts an X-API-Key header, but this CLI does not use it.) There is no browser or cookie login and no hosted setup URL — mint a key from the Deutero dashboard. 'auth status' shows what is configured.

Run `deutero-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color`.

Global format flags share one contract on promoted, novel, sync, and `--deliver` paths:

- `--json` — one JSON document on stdout (sync progress events go to stderr)
- `--compact` — keep identity/status/timestamp fields; does not change the document vs stream shape
- `--csv` / `--plain` — tabular rows (collection envelopes unwrap to the row array)
- `--quiet` — one identity value per row, no envelope

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4 --agent --select characteristics,completed,embed_metadata
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Explicit confirmation** — `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success, and use `--ignore-missing` only when a missing delete target should count as success

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Paths and state

Agents should treat the CLI's path resolver as part of the runtime contract:

- Use `--home <dir>` for one invocation, or set `DEUTERO_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `DEUTERO_CONFIG_DIR`, `DEUTERO_DATA_DIR`, `DEUTERO_STATE_DIR`, `DEUTERO_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `DEUTERO_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `deutero-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

  ```json
  {
    "mcpServers": {
      "deutero": {
        "command": "deutero-pp-mcp",
        "env": {
          "DEUTERO_HOME": "/srv/deutero"
        }
      }
    }
  }
  ```

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `DEUTERO_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `DEUTERO_HOME`, or `doctor` will not find credentials left under the former root.

## Automatic learning

This CLI ships a self-capturing learning loop. The CLI does its own bookkeeping: every invocation is journaled locally, a failed flag followed by a corrected retry auto-derives a `flag_alias` candidate, and a `teach` on a query family without a playbook auto-synthesizes a `playbook_candidate` from the session's journal. Your job is judgment only: `recall` first, act on surfaced candidates, `teach` the final answer, `playbook amend` when you observe a correction. You never record failures by hand.

### Step 1: `recall` before any discovery

Before list/search/drill commands on a new user question, pass the question as an argv or MCP tool argument to `recall --agent`. Do not interpolate user-controlled text into a shell command line.

Quoted `recall "<question>"` breaks on an apostrophe, which is ordinary English. A quoted heredoc breaks when a body line equals the delimiter, and that delimiter is published in these docs. Write the question with a non-shell file-writing tool, then read it back as data:

```bash
# Write the question verbatim with your file-writing tool (no shell involved).
# Command substitution on a file only ever yields data — the shell never
# parses the file's bytes as syntax.
QUERY=$(cat /path/to/question.txt)
deutero-pp-cli recall "$QUERY" --agent
```

Prefer MCP: pass the question as the tool's query argument. `"$QUERY"` after a file read is argv-safe; putting the question itself in the command text is not.

The response envelope:

```json
{
  "query": "...",
  "normalized": "<normalized form>",
  "query_entities": ["..."],
  "found": true | false,
  "match_score": 0.0,
  "results": [
    { "resource_id": "...", "resource_type": "...", "venue": "...",
      "confidence": 2, "entity_match": "exact|partial|unknown",
      "source": "taught|preseed|pattern", "warnings": ["..."] }
  ],
  "mismatches": [ /* only when --debug-mismatches */ ],
  "warnings": [ /* top-level */ ],
  "candidates": [
    { "id": 12, "class": "flag_alias | playbook_candidate",
      "summary": "...", "sightings": 3, "last_seen": "...",
      "rationale": "...",
      "next_action": ["<trial command>", "deutero-pp-cli learnings confirm 12"] }
  ],
  "playbook": {
    "query_family": "...",
    "playbook": {
      "steps": [ { "cmd": "<command with {slot} substitution>", "purpose": "..." } ],
      "entity_slots": ["$ENTITY"],
      "expected_tool_calls": 3
    },
    "slots_resolved": { "$ENTITY": { "token": "<live token>", "canonical": "<canonical>" } },
    "notes": "<workarounds + gotchas for this query family>"
  },
  "notes": "<duplicate surface for non-playbook callers>"
}
```

Empty-store short-circuit: if the store has no learnings, playbooks, or candidates yet (recall finds nothing and `learnings list` and `learnings candidates` are both empty), skip recall for the rest of this session instead of taxing every query; resume recall-first once something has been taught.

### Step 2: decision tree

Read `candidates`, `playbook`, `notes`, `results[0]`, and warnings in that order:

```
if Candidates present (warnings include "candidates_present"):
    -> candidates are try-then-confirm, never facts. Follow each candidate's
       two-step next_action verbatim: run the trial command first, then run
       `learnings confirm <id>` only after the trial verified the behavior.
       Reject a wrong candidate with `learnings reject <id>`.
    -> NEVER re-teach something recall surfaced as a candidate; confirm or
       reject that candidate instead of teaching a duplicate.
    -> candidates ride alongside playbooks and resource hits, not instead of
       them; continue with the branches below after acting on them.

if Playbook present:
    -> READ Playbook.notes verbatim FIRST (workarounds + gotchas the CLI surface doesn't expose)
    -> replay Playbook.steps in order, substituting Playbook.slots_resolved entries
       for the entity slot tokens. If a step's slot is unresolved, fall back to
       discovery for that step only.
    -> the Playbook's expected_tool_calls is a budget; if you find yourself running
       materially more, record the divergence via `deutero-pp-cli playbook amend`
       at end-of-session.

elif Notes present (no Playbook):
    -> read Notes verbatim before any discovery step; they carry known gotchas
       for this query family even when no structured choreography exists yet.

elif Found AND Results[0].EntityMatch == "exact" AND Results[0].Confidence >= 2:
    -> skip discovery; fetch live data for Results[*].ResourceID in parallel

elif Found AND Results[0].EntityMatch == "partial":
    -> candidate hint, NOT a hit; read the resource title to validate before trusting

elif (any row in Mismatches[] when --debug-mismatches was passed):
    -> treat as cold start; the stored learning is for a different entity
       (different canonical resolved from query_entities)

else:  // Found == false, no playbook, no notes
    -> cold start; run discovery normally; teach the answer afterward (Step 4).
       If the family has no playbook yet, that teach auto-synthesizes a
       playbook candidate from this session's journal - you do not need to
       record one by hand.
```

Playbook and Notes are orthogonal to the per-resource path. A recall response can carry both a Playbook AND a `Results[]` hit - use both: the Playbook tells you which choreography to run; the resource hits short-circuit specific steps. Default to skipping `mismatches`; pass `--debug-mismatches` only when investigating cold-start surprises.

Candidate judgment details: `learnings confirm <id>` prints the candidate's full payload before materializing it - check that the printed payload matches the behavior you verified. `learnings reject <id>` tombstones the derivation signature so the same candidate does not resurface. The envelope carries only the few candidates worth acting on now; `deutero-pp-cli learnings candidates` lists the full open set.

Graceful degradation: if `learnings confirm` is an unknown command, you are driving an older binary - ignore the candidates guidance and follow the rest of the protocol.

### Step 3: always read `warnings`

- `low_confidence`: row exists at `confidence<2`. Treat as a hint, not a skip-discovery hit.
- `resource_not_in_store`: the local store doesn't have the resource the learning points at. The match validator couldn't classify entities — direct-fetch and re-evaluate.
- `cross_alias_match` (per-result): the row was taught under a different alias and matched the live query's canonical via `entity_lookups` (e.g., a "USA" teach satisfying a "United States" recall). Trust the resource_id.
- `similar_shape_different_entity:<canonical>` (top-level): a structurally matching row exists but its canonical entity differs from the live query's. Treated as cold start; the warning carries the conflicting canonical as a hint, but the row is NOT promoted into Results.
- `ambiguous_alias` (top-level): a single query entity resolved to multiple canonicals (e.g., "Cards" → Arizona Cardinals + St. Louis Cardinals). Surface the ambiguity from context before committing to a resource.
- `candidates_present` (top-level): the envelope carries a `candidates` section. Handle it via the candidates branch in Step 2 before anything else.
- `lookup_refresh_available` (top-level): an entity in the query has no lookup row yet, but synced data could provide one. Run `deutero-pp-cli sync` to refresh entity lookups.
- Top-level `no_learnings_for_query_family`: the table had no rows above the Jaccard floor. Pure cold start.

### Step 4: `teach &` after finalizing your response - always

Teaching is unconditional. After resolving a query the store could not answer, background-teach the final resource mapping - no call-count threshold, no judging whether it was "worth" learning. The teach is the anchor of the loop: it triggers playbook synthesis for a family without a playbook, and same-referent phrasings fold into one family so near-duplicate teaches do not fragment the store. Fire it after assembling your user-facing response but BEFORE emitting it, with a shell `&` so the call returns immediately. Pass the query the same way as recall — argv/MCP, or file-then-`$QUERY`. Do not splice the question into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
deutero-pp-cli teach --query "$QUERY" --resource-type <type> --resource <id1> --resource <id2>
# (append shell `&` to background it)
```

Silent on success. Errors only land in `teach.log` under the resolved state dir. Teach the **most specific** resource - if the user asked a broad question and you walked through parent records to find the specific answer, teach the leaf id, not the parent. The CLI uses seeded `entity_lookups` for cross-alias resolution at recall time, so a teach under one alias (e.g., "Niners") satisfies future queries under another alias (e.g., "49ers", "San Francisco") automatically.

PII rule: teach the structural question with identifiers stripped - never include names, emails, phone numbers, account ids, or other personal identifiers in taught queries or notes. The CLI scans teach queries for obvious email/phone shapes and warns, but does not block; strip before teaching rather than relying on the warning.

### Step 5: playbooks - optional flags, automatic synthesis

You do not need to decide whether a session "deserves" a playbook: a teach on a family without one auto-synthesizes a `playbook_candidate` from the session's journal, and the next session judges it via confirm/reject. Attach explicit playbook flags only when you already hold choreography worth recording verbatim - workarounds the CLI didn't surface (silently-dropped flags, undocumented params, pagination tricks, payload gotchas). Prefer the **integrated one-call form** - record the resource learning and the playbook in the same `teach` invocation:

```bash
# Common case: record both the resource learning AND the playbook in one call.
QUERY=$(cat /path/to/question.txt)
deutero-pp-cli teach \
  --query "$QUERY" \
  --resource <id> \
  --playbook-file ~/playbooks/<shape>.json \
  --playbook-notes-file ~/playbooks/<shape>-notes.md
# (append shell `&` to background it)

# Alternate: playbook-only (no resource to record alongside).
QUERY=$(cat /path/to/question.txt)
deutero-pp-cli teach-playbook \
  --query "$QUERY" \
  --playbook-file ~/playbooks/<shape>.json \
  --notes-file ~/playbooks/<shape>-notes.md
```

Playbook files are JSON with `steps`, `entity_slots`, `expected_tool_calls`. Notes files are markdown carrying the gotchas verbatim. File-free callers (MCP-only agents) pass the same content inline: `--playbook-json` and `--playbook-notes` on the integrated `teach` form, `--playbook-json` and `--notes` on `teach-playbook`. On the integrated `teach` form, the playbook flags are optional - omit them entirely for a resource-only teach. On the standalone `teach-playbook` form, at least one of the playbook and notes flags must be set; both empty is rejected. Playbooks are keyed on the structural query family (entities stripped) so a recipe taught from one entity-shaped query applies to every other query of the same shape, with `slots_resolved` binding the live query's canonical at recall time.

When you DO find a playbook on a future recall, treat it as ground truth: replay the steps with `slots_resolved` substitutions, skip the discovery that the choreography already documents, and read `notes` before any step.

### Step 6: `playbook amend &` when your debug response identifies a correction

If your debug-protocol response identifies a concrete correction the notes or playbook should know — a workaround, an undocumented endpoint shape, a stale field name, observed schema drift, an empty-payload fallback — fire `playbook amend` BEFORE emitting your user-facing response. Same fire-and-forget posture as `teach`. Pass the query and note as argv/MCP arguments, or write each with a non-shell file tool and read them back (`QUERY=$(cat ...)`, `NOTE=$(cat ...)`). Do not interpolate either string into the command text:

```bash
QUERY=$(cat /path/to/question.txt)
NOTE=$(cat /path/to/note.txt)
deutero-pp-cli playbook amend \
  --query "$QUERY" \
  --add-note "$NOTE"
# (append shell `&` to background it)
```

What counts as worth amending: a behavior you OBSERVED this session that future-you would benefit from knowing. Examples worth amending:

- A workaround for a CLI surface that silently drops or misorders a flag.
- An undocumented endpoint shape (response wrapped in `{meta, results}`, payload nested two levels deeper than the docs claim).
- Observed schema drift (a field renamed, an index that shifted between seasons, a category label that the API now returns lower-cased).

What does NOT belong in notes:

- The year-specific or entity-specific answer to the user's question. That's the response, not a learning.
- Per-team / per-athlete / per-row data the playbook already retrieves at runtime.
- Statements that paraphrase what the existing notes already say.

The amend command appends to the family's existing notes with a timestamped marker (`[amend YYYY-MM-DDTHH:MMZ]: <text>`). Multiple amends accumulate; the audit trail is visible. If no playbook exists yet for the family, amend creates a notes-only one (so cold-start corrections still land).

#### PII discipline for amend notes

`playbook amend` notes are designed to potentially flow upstream as shared knowledge in future versions of the Printing Press. Keep them clean of user-identifying content so the upstream-contribution path stays open without retroactive scrubbing:

- **Do NOT embed** paths to user filesystems, personal API keys or tokens, user email addresses, user GitHub handles, or specific query histories tied to a single user.
- **Acceptable**: endpoint shapes, undocumented field names, API gotchas, observed schema drift, workarounds for CLI surfaces, generalizable pagination or retry tactics.

If a correction is only meaningful with user-specific context, it belongs in a personal note, not in the playbook amend.

### Measuring the loop

`deutero-pp-cli learnings stats` reports recall hit rate, teach-to-reuse, playbook resolution rate, and candidate confirm/reject counts from the local `learn_events` table. Rates are null until they have a denominator; everything stays on this machine. Use it to check whether the loop is earning its keep for this CLI.

### Disabling learning

- `--no-learn` on a single command short-circuits both `recall` and the `teach` write path. Use for deterministic agent flows or tests that must not be affected by accumulated learnings.
- `DEUTERO_NO_LEARN=true` in the environment globally disables the pipeline.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
deutero-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
deutero-pp-cli feedback --stdin < notes.txt
deutero-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `DEUTERO_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `DEUTERO_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename). Binary-response commands write decoded payload bytes (not the base64 JSON envelope) and print a small JSON receipt on stdout; `--json`/`--csv` do not refuse when this sink is set. |
| `webhook:<url>` | POST the output body to the URL (`application/json`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled or recurring agent reuses the same saved flags while providing different input each run.

```
deutero-pp-cli profile save briefing --json
deutero-pp-cli --profile briefing interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4
deutero-pp-cli profile list --json
deutero-pp-cli profile show briefing
deutero-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 6 | Partial failure |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `deutero-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/ai/deutero/cmd/deutero-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add deutero-pp-mcp -- deutero-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which deutero-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   deutero-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `deutero-pp-cli <command> --help`.
