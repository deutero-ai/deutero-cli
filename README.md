# Deutero CLI

**Every Deutero endpoint, plus the flow-behaviour and fielding answers that take many API calls to assemble and that no single endpoint returns.**

Deutero's own MCP server already exposes every endpoint, so mirroring the API is table stakes. This CLI earns its place on the questions the API cannot answer in one call: how a branching Interview Flow actually routed across a study (`flow coverage`), the interview-by-question rectangle nobody returns (`answers matrix`), what a flow's fetches and signals really did (`flow effects`), which completions your webhook never reported (`reconcile`), and whether a study will fill by Thursday (`fielding`). Branch decisions come back one interview at a time and never in bulk, so these commands fan out for you — bounded, concurrent, and with the cost shown. They read live; `--limit` caps the fan-out at 500 interviews by default (raise it to at most 10,000). `sync` is separate and covers org-level resources only (projects, webhooks, webhook deliveries, embed keys, graph node types and health) — interviews and transcripts are study-scoped and cannot be synced.

## Install

Prebuilt binaries cover macOS (Apple Silicon), Linux (amd64 and arm64), and Windows (amd64). Every push of a `vX.Y.Z` tag builds and publishes them via [GoReleaser](https://goreleaser.com).

### Homebrew (macOS + Linux)

```bash
brew install deutero-ai/deutero-cli/deutero-pp-cli
```

This installs both `deutero-pp-cli` and `deutero-pp-mcp`. Upgrade with `brew upgrade deutero-pp-cli`.

### Pre-built binary (all platforms, including Windows)

Download the archive for your platform from the [Releases page](https://github.com/deutero-ai/deutero-cli/releases):

| Platform | Archive |
|---|---|
| macOS (Apple Silicon) | `deutero-pp-cli_<version>_darwin_arm64.tar.gz` |
| Linux (amd64) | `deutero-pp-cli_<version>_linux_amd64.tar.gz` |
| Linux (arm64) | `deutero-pp-cli_<version>_linux_arm64.tar.gz` |
| Windows (amd64) | `deutero-pp-cli_<version>_windows_amd64.zip` |

Each archive contains both `deutero-pp-cli` and `deutero-pp-mcp`, plus a `checksums.txt` for verification. On macOS, clear the Gatekeeper quarantine after extracting: `xattr -d com.apple.quarantine <binary>`. On Unix, mark the binary executable: `chmod +x <binary>`. Then move it onto your `PATH`, e.g. `sudo mv deutero-pp-cli /usr/local/bin/`.

### Build from source (requires Go 1.26.6 or newer)

```bash
git clone https://github.com/deutero-ai/deutero-cli.git
cd deutero-cli
go build -o deutero-pp-cli ./cmd/deutero-pp-cli
```

Move the resulting binary onto your `PATH`, e.g.:

```bash
sudo mv deutero-pp-cli /usr/local/bin/
```

To build the MCP server binary instead (for use with Claude Desktop or another MCP client):

```bash
go build -o deutero-pp-mcp ./cmd/deutero-pp-mcp
```

### `go install`

```bash
go install github.com/deutero-ai/deutero-cli/cmd/deutero-pp-cli@latest
go install github.com/deutero-ai/deutero-cli/cmd/deutero-pp-mcp@latest
```

This places the binaries in `$(go env GOPATH)/bin` (or `$HOME/go/bin` if `GOPATH` is unset) — make sure that directory is on your `PATH`.

## Use with Claude Desktop

Build or install `deutero-pp-mcp` (see above), then add it to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "deutero": {
      "command": "deutero-pp-mcp",
      "env": {
        "DEUTERO_API_KEY": "<your-key>"
      }
    }
  }
}
```

## Authentication

Two ways to authenticate:

- **API key** — set `DEUTERO_API_KEY`, or run `deutero-pp-cli auth set-token`. The CLI sends it as an `Authorization: Bearer` header; the API also accepts a Stytch M2M access token there. (The API separately accepts an `X-API-Key` header, but this CLI does not use it.) Mint a key from the Deutero dashboard.
- **Browser login** — `deutero-pp-cli auth login` runs a Stytch Connected Apps OAuth2 + PKCE flow: it opens your browser, captures the redirect on a loopback server, and exchanges the code for tokens. It uses Deutero's first-party public Connected App by default; pass `--client-id` (or `DEUTERO_OAUTH_CLIENT_ID`) to use a different Connected App — distinct from your API key or project ID. `--domain` (or `DEUTERO_OAUTH_DOMAIN`) points at a different login domain (e.g. a test project), which also requires an explicit `--client-id`; endpoints are otherwise resolved automatically via OIDC discovery. Use `--no-launch` to print the URL instead of opening a browser (e.g. over SSH).

Login tokens (access token, refresh token, and any confidential-client secret) are stored in your OS keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager) when one is available, falling back automatically to a permission-locked `credentials.toml` otherwise — headless servers, containers, CI, and agent sandboxes commonly fall into the fallback case, and both paths work identically from the CLI's perspective. `auth status` shows what's configured, including which backend (`Token storage: keyring` or `file`) currently holds the tokens; `auth logout` clears both.

## Quick Start

```bash
# Confirms DEUTERO_API_KEY is set and the host answers before anything else runs
deutero-pp-cli doctor

# Find the project you want; every study hangs off one
deutero-pp-cli projects list

# Get the study id the rest of the commands take
deutero-pp-cli projects studies list be8fc6bf-8819-4f73-81f9-961cf751b13a

# Totals and quota fill — the starting point the API does answer
deutero-pp-cli studies stats get-study f0ed9214-06fc-4420-87b3-31240b670fc1

# The question no single API call answers: how the flow actually routed
deutero-pp-cli flow coverage --study f0ed9214-06fc-4420-87b3-31240b670fc1

# The rectangle for counting and correlating, straight to a spreadsheet
deutero-pp-cli answers matrix --study f0ed9214-06fc-4420-87b3-31240b670fc1 --csv

```

## Unique Features

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

## Usage

Run `deutero-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `DEUTERO_CONFIG_DIR`, `DEUTERO_DATA_DIR`, `DEUTERO_STATE_DIR`, or `DEUTERO_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `DEUTERO_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export DEUTERO_HOME=/srv/deutero
deutero-pp-cli doctor
```

Under `DEUTERO_HOME=/srv/deutero`, the four dirs resolve to `/srv/deutero/config`, `/srv/deutero/data`, `/srv/deutero/state`, and `/srv/deutero/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

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

Precedence matters in fleets: an ambient per-kind variable such as `DEUTERO_DATA_DIR` overrides an explicit `--home` for that kind. Use `DEUTERO_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `DEUTERO_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `deutero-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### credits

Organization credit balances, usage, and reservations.

- **`deutero-pp-cli credits`** - Get credit balance

### embed

Embed-widget keys and install snippets.

- **`deutero-pp-cli embed create`** - Create a publishable embed key for the caller's organization, optionally
pinned to one study.

**The response includes ``publishable_key`` and ``signing_secret`` — this is
the only time they are returned. Store them immediately.**
- **`deutero-pp-cli embed list-keys`** - Every embed key owned by the caller's organization (secrets omitted).
- **`deutero-pp-cli embed update`** - Change allowed origins, revoke/reactivate (``status``), or adjust the metadata size limit.

### graph

Manage graph

- **`deutero-pp-cli graph`** - The configuration schema for every step type an interview flow can contain,
generated from the same registry the validator enforces — so it cannot drift
from what `check_graph` will accept.

Read this before authoring a flow. It returns, per step type: what the step
does, whether it sits on the interview path or attaches to a question (and with
`after` or `before`), and every config field with its type, whether it is
required, and what it means. Plus the valid question types, condition
operators, decision modes, merge modes, the built-in variables available to
piped text, and the structural rules the validator applies.

### health

Manage health

- **`deutero-pp-cli health`** - Liveness check

### interviews

Manage interviews

- **`deutero-pp-cli interviews <interview_id>`** - One interview with its screening answers, participant characteristics, embed-widget metadata (with per-key provenance), message count, and the answers it captured into Interview Flow variables. ``web_source``, ``external_participant_id`` and ``referrer`` say where the participant arrived from: the first two are set by building the entry link with ``?source=`` and ``?participant_id=`` (or, for an embedded study, a ``participant_id`` key in the widget's metadata bag). All three are participant-supplied text — labels to join on, never proof of who answered.

### projects

Containers that group studies.

- **`deutero-pp-cli projects create`** - Create a new project owned by the caller's organization.
- **`deutero-pp-cli projects get`** - Get a project
- **`deutero-pp-cli projects list`** - List every project in the caller's organization.
- **`deutero-pp-cli projects update`** - Update a project's name and/or description. Omitted fields are untouched.

### simulations

Manage simulations

- **`deutero-pp-cli simulations delete`** - Removes the run's bookkeeping row. The interview it produced is kept — delete that separately if you want the transcript gone. A run still in flight cannot be deleted.
- **`deutero-pp-cli simulations get`** - One run's current state. Poll this after starting a run: ``running`` until the interview finishes, then ``completed`` (with ``credits_used``) or ``failed`` (reservation released, nothing charged — ``error`` says why).

### studies

Study creation and core configuration.

- **`deutero-pp-cli studies create`** - Create a new study inside a project. The study starts with an empty main
question list; configure the rest of the lifecycle with the Welcome,
Screening, Characteristics, Questions, Recruitment, and Embed endpoints.

Interview mode is **linear** (ordered question list).
- **`deutero-pp-cli studies get-study`** - Full study configuration, including per-lifecycle-stage status (welcome configured, screening/characteristics enabled and question counts, main question count) and participation links.
- **`deutero-pp-cli studies update`** - Partial update of study properties. Only fields present in the request body are changed. A ``redirect_url`` is normalized on the way in: a single-braced ``{external_participant_id}`` is corrected to the ``{{external_participant_id}}`` the interviewer substitutes, and anything else in double braces is kept but named in ``redirect_url_warning``.

### webhooks

Organization-level event subscriptions: signed HTTP callbacks for interview, analysis, simulation, study and credit events. Distinct from the per-step signals an Interview Flow sends mid-interview.

- **`deutero-pp-cli webhooks create`** - Subscribe a URL of yours to organization events. The response carries the
**signing secret, the only time it is ever shown** — store it before doing
anything else, and verify every delivery against it.

The endpoint receives events for every study the organization owns. Name the
events you want in ``events``; an empty list means all of them.
- **`deutero-pp-cli webhooks delete`** - Permanently removes the endpoint and its delivery log. To stop deliveries temporarily, set ``enabled`` false with update_webhook instead — deleting throws the signing secret away with the endpoint.
- **`deutero-pp-cli webhooks list`** - Every webhook endpoint in your organization, newest first. Signing secrets are not included — they are shown only when an endpoint is created or its secret is rotated.
- **`deutero-pp-cli webhooks list-event-types`** - Every event an endpoint can subscribe to, what makes it fire, and the fields its payload carries — read this before creating an endpoint, so the subscription names events that exist and the receiver knows what it will get. ``interview.completed`` and ``interview.started`` carry ``external_participant_id``, your own id for the participant, which is what lets a receiver act on a finished interview without calling back.
- **`deutero-pp-cli webhooks update`** - Partial update: only the fields you send change. Set ``enabled`` false to pause deliveries without losing the endpoint or its secret. Changing ``events`` replaces the subscription list rather than adding to it.


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`deutero-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`deutero-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`deutero-pp-cli learnings list`** - Inspect taught rows
- **`deutero-pp-cli learnings forget <query>`** - Undo a teach
- **`deutero-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`deutero-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`deutero-pp-cli teach-pattern`** - Install a query/resource template up front
- **`deutero-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `DEUTERO_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `deutero-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4

# JSON for scripting and agents
deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4 --json
# Filter to specific fields
deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4 --json --select characteristics,completed,embed_metadata

# Dry run — show the request without sending
deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4 --dry-run

# Agent mode — JSON + compact + no prompts in one flag
deutero-pp-cli interviews caeaec18-cab8-44fb-9be4-fd45dac9efe4 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries and add `--ignore-missing` to delete retries when a no-op success is acceptable
- **Explicit confirmation** - `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `6` partial failure, `7` rate limited, `10` config error.

## Health Check

```bash
deutero-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `deutero-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/deutero-study-management-pp-cli/config.toml`; `--home`, `DEUTERO_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `DEUTERO_API_KEY` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `deutero-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `deutero-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $DEUTERO_API_KEY`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **401 with 'Missing authentication credentials'** — Set DEUTERO_API_KEY. The CLI sends it as an Authorization: Bearer header; a Stytch M2M access token works there too.
- **403 'Access denied' on a study id you were given** — The key belongs to one organization; a study in another org reads as denied. Run 'projects list' to see what this key can actually reach.
- **First command after an idle period takes several seconds** — The host cold-starts (~4s observed). Raise --timeout if you are scripting tight loops; subsequent calls return in well under a second.
- **A study-wide command is slow on a large study** — Branch decisions and flow effects exist only per interview, so these commands issue one request each. Lower --limit to bound the fan-out.
- **flow coverage reports a branch arm as 'not observed'** — That means no interview recorded taking it — not that it cannot fire. Captured answers and decisions are best-effort projections upstream.
- **flow coverage or answers matrix returns nothing for a study** — Linear studies have no flow records at all, and captured variables are empty for every linear study. answers matrix falls back to the message stream; flow coverage will correctly report nothing.
- **sync skips interviews or transcripts** — Those endpoints are scoped to a study id that sync cannot supply, so it syncs org-level resources only (projects, webhooks, webhook deliveries, embed keys, graph node types and health). The analysis commands read live and do not need sync.
- **A step id in the output does not match anything in your flow editor** — Report it — the API resolves compiled node ids back to your authored step ids, so a raw id like n3 is flagged in output rather than shown as a step name.
- **flow coverage or rehearse diff returns empty branches for a study that clearly has questions** — Those measures exist only in flow (graph) mode. A linear question-list study has no branch or variable records at all — answers matrix is the command for its answers.

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**dovetail-mcp**](https://github.com/dovetail/dovetail-mcp) — TypeScript
- [**QualiGPT**](https://github.com/KindOPSTAR/QualiGPT) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
