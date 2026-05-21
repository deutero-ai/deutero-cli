---
name: deutero-research
description: This skill should be used when the user asks to "create a study", "set up a survey", "generate interview questions", "run a simulation", "analyze interviews", "get analysis results", "fetch transcripts", "manage webhooks", "run thematic analysis", "import a research guide", "set up a Deutero study", "generate personas", "check participation", "get interview URLs", "extract questions from a guide", "suggest a study from a website", or any other task involving the deutero-cli tool for qualitative research management.
version: 0.1.0
---

# Deutero Research CLI Skill

A skill for using `deutero-cli` to plan, run, and analyze AI-powered qualitative interview studies on the [Deutero](https://deutero.ai) platform.

## Purpose

Deutero runs AI-powered async text interviews with real users or AI personas, then performs multi-phase thematic analysis to surface actionable research insights. All operations are available via `deutero-cli`, which wraps the Deutero Interview Admin API.

## Authentication

**OAuth2 (recommended):**
```bash
export DEUTERO_CLIENT_ID="your-client-id"
export DEUTERO_STYTCH_PROJECT_ID="your-project-id"
deutero auth login          # opens browser flow, stores token in ~/.deutero/config.json
deutero auth status         # verify login state
deutero auth logout         # clear stored tokens
```

**API key (fallback):**
```bash
export DEUTERO_API_KEY="your-api-key"
# or persist it:
deutero config set-key your-api-key
```

The explicit API key always takes precedence over an OAuth token if both are present.

## Active Context System

Most commands accept an optional ID argument. When omitted, the CLI:
1. Uses the active survey/project stored in `~/.deutero/config.json`
2. Falls back to an interactive picker if no active context is set

Set active context once to avoid repeating IDs across every command:
```bash
deutero projects set-active    # pick from list, persists project ID
deutero surveys set-active     # pick from list under active project, persists survey ID
deutero config show            # verify current active project + survey
```

## Model Tiers

Three tiers are available for both simulation and analysis:

| Tier | Notes |
|------|-------|
| `open_weights` | Default — lowest cost |
| `premium` | Better quality, moderate cost |
| `frontier` | Highest quality, highest cost |

Set the tier on the survey itself (`--model-tier` on create/generate/update), or pass `-m`/`--model-tier` per-command for simulation and analysis.

## End-to-End Study Workflow

### 1. Create or Generate a Study

**AI-generate a study (recommended):**
```bash
deutero surveys generate \
  --survey-type user_experience \
  --business-context "B2B SaaS project management tool" \
  --research-need "Understand onboarding friction" \
  --target-users "Engineering managers at mid-size companies" \
  --model-tier premium
```

Four survey types: `user_experience`, `sociology`, `customer_development`, `polling`. Each type takes different input options — see `references/generation-and-importing.md`.

**Create directly without AI:**
```bash
deutero surveys create \
  --project-id <PROJECT_ID> \
  --name "Onboarding Study" \
  --survey-type user_experience \
  --model-tier premium
```

**Update an existing study:**
```bash
deutero surveys update <SURVEY_ID> --name "Renamed" --language Spanish
deutero surveys update <SURVEY_ID> --anonymous --redirect-url https://example.com/done
```

**Import from a research guide file:**
```bash
deutero surveys extract-from-guide guide.txt               # extracts details + questions
deutero surveys extract-from-guide guide.txt --extract details   # metadata only
deutero surveys extract-from-guide guide.txt --extract questions # questions only
```

**Generate a study suggestion from a product website:**
```bash
deutero surveys suggest-from-site https://myproduct.com --language English
```

### 2. Manage Questions

```bash
# AI-generate questions (count required, 1–25)
deutero questions generate --count 5
deutero questions generate --count 8 --instructions "Focus on emotional responses and pain points"

# Create a question manually
deutero questions create --question "What is your biggest challenge with X?" --follow-up

# Get, update, or delete a question
deutero questions get <QUESTION_ID>
deutero questions update <QUESTION_ID> --max-turns 5 --question "Revised text"
deutero questions delete <QUESTION_ID>

# Reorder questions in a survey
deutero questions reorder <SURVEY_ID> <Q1_ID> <Q2_ID> <Q3_ID>

# Upload and reorder images on a question
deutero questions upload-image <QUESTION_ID> --image-file prototype.png --label "Variant A"
deutero questions reorder-images <QUESTION_ID> <IMG1_ID> <IMG2_ID>
```

See `references/study-and-question-management.md` for full question type details: scale, multiple choice, image-based, and slot-based questions.

### 3. Generate Personas

```bash
deutero personas generate --count 3
deutero personas generate --count 5 --instructions "Enterprise users, risk-averse, non-technical"
deutero personas list        # view all personas with their interview counts
```

### 4. Simulate Interviews

```bash
deutero simulate run                           # prompts for both IDs
deutero simulate run <PERSONA_ID>              # uses active survey
deutero simulate run <SURVEY_ID> <PERSONA_ID>  # fully explicit
```

For real participant interviews, retrieve the interview URL and agent requirements:
```bash
deutero surveys agent-requirements -o requirements.md
```

### 5. Track Participation

```bash
deutero surveys participation                  # totals, completion rate, quota fill rate
deutero interviews list                        # per-interview: id, name, status, simulated, message count
deutero surveys interviews                     # same data from surveys group
```

### 6. Run Thematic Analysis

```bash
# Per-interview (phases 1–4)
deutero analysis run --interview-id <ID> --model-tier premium

# All completed interviews in a survey
deutero analysis run --survey-id <ID>

# Cross-case analysis (phase 5, survey-wide patterns)
deutero analysis run --survey-id <ID> --cross-case --xml-output cross_case.xml

# Shortcut from interviews group
deutero interviews analysis <INTERVIEW_ID> --model-tier premium

# Check status
deutero analysis status --survey-id <ID>
deutero analysis status --interview-id <ID>
```

See `references/analysis-and-results.md` for the five analysis phases and result retrieval.

### 7. Retrieve Results and Transcripts

```bash
# Interview transcript (interactive picker if ID omitted)
deutero interviews transcript
deutero interviews transcript <INTERVIEW_ID> -o transcript.json

# Per-phase analysis results (XML output)
deutero analysis results-interview <ID> --phase emergent_themes -o themes.xml
deutero analysis results-interview <ID> --phase connections -o connections.xml

# Cross-case survey results (XML)
deutero analysis results-survey -o cross_case.xml
deutero analysis results-survey <SURVEY_ID> -o cross_case.xml
```

## Global Options and Output

Every command supports:

| Flag | Env var | Description |
|------|---------|-------------|
| `--api-key` | `DEUTERO_API_KEY` | API key (fallback auth) |
| `--base-url` | `DEUTERO_BASE_URL` | Override API base URL |
| `--version` | — | Show CLI version |
| `--help` | — | Show help |
| `--tui` | — | Launch interactive TUI (requires `[tui]` extra) |

Most data commands also accept:
- `-o, --output <file>` — write full JSON response to a file
- `--xml-output <file>` — write XML content to a file (generation, cross-case analysis)

## Credits

```bash
deutero credits balance                                           # current balance + usage
deutero credits estimate-simulation --model-tier premium --participants 10
deutero credits estimate-analysis -m premium --interviews 5
deutero credits estimate-survey --model-tier premium --participants 10 --include-analysis
```

## Configuration File

Stored at `~/.deutero/config.json`:

```json
{
  "api_key": "your-api-key",
  "active_project_id": "e927074f-...",
  "active_survey_id": "c3a1f8b2-...",
  "access_token": "...",
  "refresh_token": "...",
  "token_expiry": 1234567890,
  "client_id": "your-client-id",
  "project_id": "your-project-id"
}
```

## Additional Resources

### Reference Files

- **`references/study-and-question-management.md`** — Survey CRUD with all properties, all question types (open-ended, scale, multiple choice, image-based, slot-based), question properties, image upload and reorder
- **`references/generation-and-importing.md`** — AI survey generation (all 4 types with full option tables), question generation, `extract-from-guide`, `suggest-from-site`, persona generation
- **`references/running-studies.md`** — Simulation, real participant interview URLs, participation tracking, personas, credits estimation
- **`references/analysis-and-results.md`** — Five analysis phases explained, running analysis, polling status, retrieving per-phase and cross-case XML results, transcripts
- **`references/webhooks.md`** — Webhook endpoint management: list, update (label/url/enable/disable/events), delete, available event types
