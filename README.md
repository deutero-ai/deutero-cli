# deutero-cli

Command-line client for the **[Deutero](https://deutero.ai) Interview Admin API** — run AI-powered user research, generate questions, simulate interviews, and turn conversations into actionable insights, all from the terminal.

[![PyPI version](https://img.shields.io/pypi/v/deutero-cli.svg)](https://pypi.org/project/deutero-cli/)
[![Python](https://img.shields.io/pypi/pyversions/deutero-cli.svg)](https://pypi.org/project/deutero-cli/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## What is Deutero?

[Deutero](https://deutero.ai) is an AI-powered user research platform built for AI-first and vibe-coding teams. Instead of spending weeks on manual research, Deutero lets you:

- **Define a study** in seconds — UX, sociology, customer development, or polling
- **AI interviews users or personas** — no scheduling, no Zoom fatigue; async text-based conversations that dig past surface-level answers
- **Get actionable insights** — thematic analysis with pattern detection, user segmentation, and supporting quotes
- **Export as code context** — deliver findings as Markdown directly to your coding agent via the MCP server, or as PDF/PowerPoint for stakeholders

> *"Vibe coding is not a shortcut to product-market fit. It's a bridge between discovery and development — a way to test reality faster, not bypass it."*

---

## Authentication

The CLI supports two authentication methods:

### OAuth2 + PKCE (recommended)

Authenticate via your browser using Stytch Connected Apps. This is the most secure and convenient method.

```bash
# Option A – environment variables
export DEUTERO_CLIENT_ID="your-connected-app-client-id"
export DEUTERO_STYTCH_PROJECT_ID="your-stytch-project-id"
deutero auth login

# Option B – pass directly
deutero auth login --client-id <CLIENT_ID> --project-id <PROJECT_ID>
```

This opens your browser, authenticates you, and securely stores the access token in `~/.deutero/config.json`.

Check your login status:

```bash
deutero auth status
```

Log out (removes stored tokens):

```bash
deutero auth logout
```

Tokens are automatically refreshed when they expire using the stored `refresh_token`.

### API Key (fallback)

If you prefer not to use OAuth, you can authenticate with an API key:

1. Log in to your [Deutero dashboard](https://app.deutero.ai)
2. Navigate to **Settings**
3. Scroll to the **API Keys** section
4. Click **Generate New API Key**
5. Copy the key immediately — it is shown only once
6. Store it securely (use an environment variable or a secrets manager; never commit it to version control)

```bash
# Environment variable
export DEUTERO_API_KEY="your-api-key"

# Or save to config
deutero config set-key your-api-key
```

An active Deutero subscription with available credits is required to use the API.

---

## Installation

### Run without installing (recommended)

```bash
uvx deutero-cli --help
```

### Install globally with uv

```bash
uv tool install deutero-cli
deutero --help
```

### Install with pip

```bash
pip install deutero-cli
```

### Install with TUI support

Install the optional [Trogon](https://github.com/Textualize/trogon) TUI extra to get a fully interactive terminal user interface:

```bash
pip install "deutero-cli[tui]"
# or with uv
uv tool install "deutero-cli[tui]"
```

Once installed, launch the TUI with:

```bash
deutero --tui
```

### Development install

```bash
git clone https://github.com/deutero/deutero-cli.git
cd deutero-cli
python -m venv .venv && source .venv/bin/activate
pip install -e ".[dev,tui]"
```

---

## Quick Start

### 1. Authenticate

```bash
# OAuth2 (recommended)
export DEUTERO_CLIENT_ID="your-client-id"
export DEUTERO_STYTCH_PROJECT_ID="your-project-id"
deutero auth login

# Or use an API key
export DEUTERO_API_KEY="your-api-key"
```

### 2. Generate a research survey

```bash
# UX research
deutero surveys generate \
  --survey-type user_experience \
  --business-context "B2B SaaS project management tool" \
  --research-need "Understand onboarding friction" \
  --target-users "Engineering managers at mid-size companies" \
  --model-tier premium

# Sociology
deutero surveys generate \
  --survey-type sociology \
  --research-question "How do online communities form trust?" \
  --population-of-interest "Reddit moderators" \
  --language Spanish

# Customer development
deutero surveys generate \
  --survey-type customer_development \
  --problem-hypothesis "Teams waste time on status updates" \
  --customer-segment "Remote engineering teams"

# Polling
deutero surveys generate \
  --survey-type polling \
  --research-question "Public sentiment on AI regulation" \
  --population-segment "US adults 18-65"

# Generate a suggestion from a website
deutero surveys suggest-from-site https://example.com --language English

# Get or set model tier for an existing survey
deutero surveys model-tier <SURVEY_ID>
deutero surveys model-tier <SURVEY_ID> --set premium

# Create a survey directly (without AI generation)
deutero surveys create \
  --project-id <PROJECT_ID> \
  --survey-type user_experience \
  --name "Onboarding Study" \
  --model-tier premium

# Update an existing survey's properties
deutero surveys update <SURVEY_ID> --name "Renamed Study" --language Spanish
deutero surveys update <SURVEY_ID> --model-tier frontier --anonymous
```

### 3. Generate interview questions

```bash
deutero questions generate <SURVEY_ID> --count 5
deutero questions generate <SURVEY_ID> --count 10 --instructions "Focus on emotional responses"

# CRUD operations
deutero questions create <SURVEY_ID> --question "How do you feel about X?" --follow-up
deutero questions get <QUESTION_ID>
deutero questions update <QUESTION_ID> --question "Updated text" --max-turns 5
deutero questions delete <QUESTION_ID> --survey-id <SURVEY_ID>

# Reorder questions within a survey
deutero questions reorder <SURVEY_ID> <Q_ID_1> <Q_ID_2> <Q_ID_3>

# Upload and reorder images on a question
deutero questions upload-image <QUESTION_ID> --image-file photo.png --label "Prototype A"
deutero questions reorder-images <QUESTION_ID> <IMG_ID_1> <IMG_ID_2>
```

### 4. Generate and list personas

```bash
# Generate new personas for a survey
deutero personas generate <SURVEY_ID> --count 3

# List existing personas (includes interview count per persona)
deutero personas list
deutero personas list <SURVEY_ID>
```

### 5. Set active project and survey (optional)

Most commands that require a survey or project ID will use the **active** context if the argument is omitted, or prompt you interactively.

```bash
# Pick your active project from an interactive list
deutero projects set-active

# Pick your active survey (uses active project, or prompts for one)
deutero surveys set-active

# Verify your active context
deutero config show
```

Once set, commands like `deutero interviews list`, `deutero questions generate`, and `deutero personas generate` all work without supplying an explicit ID.

### 6. Simulate an interview

```bash
# With explicit IDs
deutero simulate run <SURVEY_ID> <PERSONA_ID>

# With active survey set — prompts only for PERSONA_ID
deutero simulate run <PERSONA_ID>

# Both IDs prompted interactively if omitted
deutero simulate run
```

### 7. Run thematic analysis

```bash
# Phases 1–4 on a single interview
deutero analysis run --interview-id <INTERVIEW_ID> --model-tier premium

# Phases 1–4 on all completed interviews in a survey
deutero analysis run --survey-id <SURVEY_ID>

# Cross-case analysis (phase 5)
deutero analysis run --survey-id <SURVEY_ID> --cross-case --xml-output cross_case.xml

# Run analysis directly from the interviews group
deutero interviews analysis <INTERVIEW_ID>
```

### 8. Check analysis status

```bash
deutero analysis status --interview-id <INTERVIEW_ID>
deutero analysis status --survey-id <SURVEY_ID>
```

### 9. Retrieve analysis results

```bash
# Per-interview phase results
deutero analysis results-interview <INTERVIEW_ID> --phase emergent_themes --output themes.xml

# Cross-case survey results
deutero analysis results-survey <SURVEY_ID> --output cross_case.xml
```

### 10. Inspect interviews and transcripts

```bash
# List interviews for a survey (uses active survey if set)
# Columns: id, participant name, start time, completed, simulated, message count
deutero interviews list
deutero interviews list <SURVEY_ID>

# Pretty-print a transcript
# If INTERVIEW_ID is omitted, an interactive list is shown to select from
deutero interviews transcript
deutero interviews transcript <INTERVIEW_ID>
```

### 11. Check credit balance and estimates

```bash
# Current balance
deutero credits balance

# Estimate credits for simulating interviews (uses active survey if set)
deutero credits estimate-simulation --model-tier premium --participants 10
deutero credits estimate-simulation <SURVEY_ID> --model-tier premium --participants 10

# Estimate credits for thematic analysis
deutero credits estimate-analysis --model-tier open_weights --interviews 5

# Estimate credits for a full survey (interviews + optional analysis)
deutero credits estimate-survey --model-tier premium --participants 10 --include-analysis
```

### 12. List projects and surveys

```bash
# List all projects (pretty table output)
deutero projects list

# List surveys in a project (uses active project if set)
deutero projects list-surveys
deutero projects list-surveys <PROJECT_ID>
```

### 13. Inspect survey details

```bash
# Get the full question list for a survey (uses active survey if set)
deutero surveys question-list
deutero surveys question-list <SURVEY_ID>

# List interviews for a survey directly from the surveys group
deutero surveys interviews
deutero surveys interviews <SURVEY_ID>
```

### 14. Extract details and questions from a research guide

```bash
# Extract both study details and interview questions (default)
deutero surveys extract-from-guide guide.txt

# Extract only study metadata
deutero surveys extract-from-guide guide.txt --extract details

# Extract only interview questions
deutero surveys extract-from-guide guide.txt --extract questions

# Save full JSON response to a file
deutero surveys extract-from-guide guide.txt --output extracted.json
```

### 15. Manage webhook endpoints

```bash
# List all configured webhook endpoints
deutero webhooks list

# Update an endpoint's label, URL, enabled state, or subscribed events
deutero webhooks update <ENDPOINT_ID> --label "Production Hook"
deutero webhooks update <ENDPOINT_ID> --url https://new.example.com/hook
deutero webhooks update <ENDPOINT_ID> --disable
deutero webhooks update <ENDPOINT_ID> --enable --event interview.completed --event survey.completed

# Delete a webhook endpoint (prompts for confirmation)
deutero webhooks delete <ENDPOINT_ID>
deutero webhooks delete <ENDPOINT_ID> --yes   # skip confirmation
```

---

## Global Options

Every command supports:

| Flag | Env var | Description |
|------|---------|-------------|
| `--api-key` | `DEUTERO_API_KEY` | API key for authentication (fallback) |
| `--base-url` | `DEUTERO_BASE_URL` | API base URL |
| `--client-id` | `DEUTERO_CLIENT_ID` | Stytch Connected App client ID |
| `--project-id` | `DEUTERO_STYTCH_PROJECT_ID` | Stytch project ID |
| `--version` | — | Show version and exit |
| `--help` | — | Show help and exit |
| `--tui` | — | Launch the interactive TUI (requires `[tui]` extra) |

Most data commands also accept:

- `--output / -o <file>` — write the full JSON response to a file
- `--xml-output <file>` — write XML content to a file (where applicable)

### Interactive prompts

All commands that require an object ID (survey, project, interview, question) will:

1. Use the **active survey / project** stored in `~/.deutero/config.json` if one is set
2. **Prompt you interactively** if no active context is configured and no argument is supplied

This means every command works with or without arguments — no more copy-pasting UUIDs.

---

## Command Reference

```
deutero
├── auth
│   ├── login                            Authenticate via OAuth2 + PKCE
│   ├── logout                           Clear stored tokens
│   └── status                           Show auth status
├── config
│   ├── show                             View configuration (incl. active context)
│   ├── set-key [KEY]                    Save API key (prompts if omitted)
│   └── set-url [URL]                    Save base URL (prompts if omitted)
├── credits
│   ├── balance                          Get credit balance and reservations
│   ├── estimate-simulation [SURVEY_ID]  Estimate credits for simulation
│   ├── estimate-analysis [SURVEY_ID]    Estimate credits for analysis
│   └── estimate-survey [SURVEY_ID]      Estimate credits for full survey
├── projects
│   ├── list                             List all projects
│   ├── list-surveys [PROJECT_ID]        List surveys for a project
│   └── set-active                       Set active project (interactive list)
├── surveys
│   ├── list [PROJECT_ID]                List surveys for a project
│   ├── generate                         Generate a research survey
│   ├── create                           Create a survey directly
│   ├── update <SURVEY_ID>               Update survey properties
│   ├── participation [SURVEY_ID]        Get participation statistics
│   ├── agent-requirements [SURVEY_ID]   Get agent requirements markdown
│   ├── model-tier [SURVEY_ID]           Get or set survey model tier
│   ├── suggest-from-site <URL>          Generate study suggestion from a URL
│   ├── question-list [SURVEY_ID]        Get full question list for a survey
│   ├── interviews [SURVEY_ID]           List interviews for a survey
│   ├── extract-from-guide <GUIDE_FILE>  Extract details/questions from a guide
│   └── set-active                       Set active survey (interactive list)
├── questions
│   ├── generate [SURVEY_ID]             Generate interview questions
│   ├── create [SURVEY_ID]               Create a single question
│   ├── get [QUESTION_ID]                Get a question by ID
│   ├── update [QUESTION_ID]             Update a question
│   ├── delete [QUESTION_ID]             Delete a question
│   ├── reorder [SURVEY_ID] <IDs>        Reorder questions in a survey
│   ├── upload-image [QUESTION_ID]       Upload an image to a question
│   └── reorder-images [QID] <IDs>       Reorder images on a question
├── personas
│   ├── list [SURVEY_ID]                 List personas with interview counts
│   └── generate [SURVEY_ID]             Generate interviewee personas
├── interviews
│   ├── list [SURVEY_ID]                 List interviews for a survey
│   ├── transcript [INTERVIEW_ID]        Pretty-print transcript (interactive select if ID omitted)
│   └── analysis [INTERVIEW_ID]          Run thematic analysis for an interview
├── simulate
│   └── run [SURVEY_ID] [PERSONA_ID]     Simulate an interview
├── analysis
│   ├── run                              Run thematic analysis
│   ├── status                           Get analysis status
│   ├── results-interview [ID]           Get per-phase results
│   └── results-survey [ID]              Get cross-case results
└── webhooks
    ├── list                             List all webhook endpoints
    ├── update <ENDPOINT_ID>             Update a webhook endpoint
    └── delete <ENDPOINT_ID>             Delete a webhook endpoint
```

> **[ID]** — argument is optional; falls back to the active survey/project from config, then prompts interactively.

---

## Detailed Command Reference

### `deutero auth`

Manage authentication via Stytch Connected Apps (OAuth2 + PKCE).

**`deutero auth login`**
Authenticate via browser-based OAuth2 + PKCE flow.

| Option | Env var | Description |
|--------|---------|-------------|
| `--client-id` | `DEUTERO_CLIENT_ID` | Stytch Connected App client ID (required) |
| `--project-id` | `DEUTERO_STYTCH_PROJECT_ID` | Stytch project ID (required) |

```bash
deutero auth login --client-id <CLIENT_ID> --project-id <PROJECT_ID>
```

**`deutero auth logout`**
Clear all stored authentication tokens from `~/.deutero/config.json`.

```bash
deutero auth logout
```

**`deutero auth status`**
Show current authentication state: logged-in status, token expiry, and refresh token availability.

```bash
deutero auth status
```

---

### `deutero config`

View and persist CLI settings.

**`deutero config show`**
Display the current configuration, including the masked API key, base URL, active project, and active survey.

```bash
deutero config show
```

**`deutero config set-key [KEY]`**
Save an API key to `~/.deutero/config.json`. Prompts securely for input if `KEY` is omitted.

```bash
deutero config set-key my-api-key
deutero config set-key   # prompts interactively
```

**`deutero config set-url [URL]`**
Save a custom API base URL. Prompts if `URL` is omitted.

```bash
deutero config set-url https://api.deutero.app
```

---

### `deutero credits`

Inspect credit balance and estimate consumption.

**`deutero credits balance`**
Show available credits, reservations, usage, limits, and trial status.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero credits balance
deutero credits balance -o balance.json
```

**`deutero credits estimate-simulation [SURVEY_ID]`**
Estimate credits required to simulate interviews for a survey.

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier (`open_weights`, `premium`, `frontier`) |
| `-n, --participants` | `1` | Number of participants to simulate |
| `-o, --output <file>` | — | Write JSON response to a file |

```bash
deutero credits estimate-simulation --model-tier premium --participants 10
deutero credits estimate-simulation <SURVEY_ID> -m premium -n 10
```

**`deutero credits estimate-analysis [SURVEY_ID]`**
Estimate credits required for thematic analysis.

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier |
| `-n, --interviews` | `1` | Number of interviews to analyze |
| `-o, --output <file>` | — | Write JSON response to a file |

```bash
deutero credits estimate-analysis -m open_weights --interviews 5
```

**`deutero credits estimate-survey [SURVEY_ID]`**
Estimate credits for a full survey run (simulation + optional analysis).

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier |
| `-n, --participants` | `1` | Number of participants |
| `--include-analysis` | `False` | Include analysis in the estimate |
| `-o, --output <file>` | — | Write JSON response to a file |

```bash
deutero credits estimate-survey --model-tier premium --participants 10 --include-analysis
```

---

### `deutero projects`

List projects and their surveys.

**`deutero projects list`**
List all projects in your account (pretty table output).

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero projects list
```

**`deutero projects list-surveys [PROJECT_ID]`**
List surveys belonging to a project.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero projects list-surveys
deutero projects list-surveys <PROJECT_ID>
```

**`deutero projects set-active`**
Interactively select the active project from a list. The chosen project ID is saved to `~/.deutero/config.json`.

```bash
deutero projects set-active
```

---

### `deutero surveys`

Manage research surveys.

**`deutero surveys list [PROJECT_ID]`**
List surveys for a project (defaults to active project).

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys list
deutero surveys list <PROJECT_ID>
```

**`deutero surveys generate`**
Generate a new research survey specification.

| Option | Description |
|--------|-------------|
| `-t, --survey-type` | Survey type: `user_experience`, `sociology`, `customer_development`, `polling` (default: `user_experience`) |
| `-l, --language` | Language for the survey (default: `English`) |
| `--business-context` | Business/product context (UX) |
| `--research-need` | Why the research is needed (UX) |
| `--target-users` | Primary audience (UX) |
| `--constraints` | Key constraints (UX) |
| `--research-question` | Research question (sociology/polling) |
| `--population-of-interest` | Population of interest (sociology) |
| `--context-or-setting` | Context or setting (sociology) |
| `--key-concepts` | Key concepts (sociology) |
| `--scope-and-boundaries` | Scope and boundaries (sociology) |
| `--problem-hypothesis` | Problem hypothesis (customer development) |
| `--customer-segment` | Customer segment (customer development) |
| `--solution-concept` | Solution concept (customer development) |
| `--key-assumptions` | Key assumptions (customer development) |
| `--success-criteria` | Success criteria (customer development) |
| `--population-segment` | Population segment (polling) |
| `--geographic-scope` | Geographic scope (polling) |
| `--survey-context` | Survey context (polling) |
| `--data-quality-requirements` | Data quality requirements (polling) |
| `--model-tier` | Set model tier on the created survey (`open_weights`, `premium`, `frontier`) |
| `-o, --output <file>` | Write JSON response to a file |
| `--xml-output <file>` | Write the XML specification to a file |

```bash
deutero surveys generate \
  --survey-type user_experience \
  --business-context "B2B SaaS" \
  --research-need "Onboarding friction" \
  --target-users "Engineering managers" \
  --model-tier premium
```

**`deutero surveys create`**
Create a new survey directly with explicit field values. Does not use AI generation — all fields are set as provided.

| Option | Description |
|--------|-------------|
| `--project-id` | Project ID to create the survey in (required) |
| `-t, --survey-type` | Survey type: `sociology`, `user_experience`, `customer_development`, `polling` |
| `--name` | Display name for the survey |
| `--description` | Description of the survey |
| `--research-questions` | Research questions |
| `--objectives` | Research objectives |
| `--target-population` | Target population |
| `--anonymous` | Mark the survey as anonymous |
| `--no-anonymous` | Mark the survey as non-anonymous |
| `--model-tier` | Model tier: `open_weights`, `premium`, `frontier` |
| `--redirect-url` | URL to redirect participants to after completing the survey |
| `-l, --language` | Language for the survey |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys create --project-id <PROJECT_ID>
deutero surveys create --project-id <PROJECT_ID> --survey-type sociology --name "Trust Study"
deutero surveys create --project-id <PROJECT_ID> --model-tier premium --anonymous -o study.json
```

**`deutero surveys update <SURVEY_ID>`**
Update one or more properties of an existing survey. Only supplied fields are changed; omitted fields are left untouched. At least one option must be provided.

| Option | Description |
|--------|-------------|
| `-t, --survey-type` | New survey type |
| `--name` | New display name |
| `--description` | New description |
| `--research-questions` | New research questions |
| `--objectives` | New research objectives |
| `--target-population` | New target population |
| `--anonymous` | Mark the survey as anonymous |
| `--no-anonymous` | Mark the survey as non-anonymous |
| `--model-tier` | New model tier (`open_weights`, `premium`, `frontier`) |
| `--redirect-url` | New redirect URL |
| `-l, --language` | New language |
| `-o, --output <file>` | Write JSON response to a file |

When both `--anonymous` and `--no-anonymous` are passed, `--anonymous` takes effect.

```bash
deutero surveys update <SURVEY_ID> --name "Renamed Study"
deutero surveys update <SURVEY_ID> --model-tier frontier --language Spanish
deutero surveys update <SURVEY_ID> --anonymous --redirect-url https://example.com/done
deutero surveys update <SURVEY_ID> --description "Updated description" -o updated.json
```

**`deutero surveys participation [SURVEY_ID]`**
Get participation statistics: total, completed, incomplete, completion rate, and quota fill rate.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys participation
deutero surveys participation <SURVEY_ID>
```

**`deutero surveys agent-requirements [SURVEY_ID]`**
Retrieve or generate agent requirements markdown for a survey. Writes a `.md` file.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write markdown to a custom file path |

```bash
deutero surveys agent-requirements
deutero surveys agent-requirements <SURVEY_ID> -o reqs.md
```

**`deutero surveys model-tier [SURVEY_ID]`**
Get or set the model tier for a survey.

| Option | Description |
|--------|-------------|
| `--set <tier>` | Set the tier (`open_weights`, `premium`, `frontier`) |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys model-tier              # get current tier
deutero surveys model-tier --set premium
deutero surveys model-tier <SURVEY_ID> --set frontier
```

**`deutero surveys suggest-from-site <URL>`**
Generate a study suggestion by scraping a website URL.

| Option | Default | Description |
|--------|---------|-------------|
| `-l, --language` | `English` | Language for the generated suggestion |
| `-o, --output <file>` | — | Write JSON response to a file |

```bash
deutero surveys suggest-from-site https://example.com --language English
```

**`deutero surveys question-list [SURVEY_ID]`**
Get the full question list for a survey.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys question-list
deutero surveys question-list <SURVEY_ID>
```

**`deutero surveys interviews [SURVEY_ID]`**
List interviews for a survey directly from the surveys group. Columns: ID, participant name, start time, end time, completed, simulated, termination reason.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero surveys interviews
deutero surveys interviews <SURVEY_ID>
```

**`deutero surveys extract-from-guide <GUIDE_FILE>`**
Extract structured study details and/or interview questions from a research guide text file.

| Option | Default | Description |
|--------|---------|-------------|
| `--extract` | `both` | What to extract: `details`, `questions`, or `both` |
| `-o, --output <file>` | — | Write JSON response to a file |

Outputs:
- **`details`** — study name, description, research questions, objectives, target population
- **`questions`** — interview questions with type, interviewer guidance, and scale/option metadata

```bash
deutero surveys extract-from-guide guide.txt
deutero surveys extract-from-guide guide.txt --extract details
deutero surveys extract-from-guide guide.txt --extract questions
deutero surveys extract-from-guide guide.txt -o extracted.json
```

**`deutero surveys set-active`**
Interactively select the active survey from a list. Saves the survey ID to `~/.deutero/config.json`.

| Option | Description |
|--------|-------------|
| `--project-id` | Project ID to list surveys from (defaults to active project) |

```bash
deutero surveys set-active
deutero surveys set-active --project-id <PROJECT_ID>
```

---

### `deutero questions`

Manage interview questions within a survey.

**`deutero questions generate [SURVEY_ID]`**
Generate AI-powered interview questions.

| Option | Description |
|--------|-------------|
| `-n, --count` | Number of questions to generate, 1–25 (required) |
| `-i, --instructions` | Additional instructions for the generator |
| `-o, --output <file>` | Write JSON response to a file |
| `--xml-output <file>` | Write XML specification to a file |

```bash
deutero questions generate --count 5
deutero questions generate <SURVEY_ID> -n 10 -i "Focus on emotional responses"
```

**`deutero questions create [SURVEY_ID]`**
Create a single question manually.

| Option | Description |
|--------|-------------|
| `--question` | Question text (required) |
| `--explanation` | Interviewer guidance / explanation |
| `--scale-json` | Scale configuration as a JSON object |
| `--option` | Fixed-choice option (repeatable) |
| `--slot` | Slot name (repeatable) |
| `--follow-up / --no-follow-up` | Enable or disable follow-up |
| `--min-turns` | Minimum conversation turns |
| `--max-turns` | Maximum conversation turns |
| `--expected-image` | Expected image description |
| `--payload-json` | Additional request fields as JSON |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions create --question "How do you feel about X?" --follow-up
deutero questions create <SURVEY_ID> --question "What is your role?" --option "Engineer" --option "Designer"
```

**`deutero questions get [QUESTION_ID]`**
Retrieve a single question by ID.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions get <QUESTION_ID>
```

**`deutero questions update [QUESTION_ID]`**
Update an existing question. Only provided fields are changed.

| Option | Description |
|--------|-------------|
| `--question` | New question text |
| `--explanation` | New interviewer guidance |
| `--scale-json` | New scale configuration |
| `--option` | Replace fixed-choice options (repeatable) |
| `--slot` | Replace slot names (repeatable) |
| `--follow-up / --no-follow-up` | Toggle follow-up |
| `--min-turns` | New minimum turns |
| `--max-turns` | New maximum turns |
| `--expected-image` | New expected image description |
| `--null-field` | Send `null` for a field (choices: `question`, `explanation`, `scale`, `options`, `slots`, `follow_up`, `min_turns`, `max_turns`, `expected_image`) |
| `--payload-json` | Additional request fields as JSON |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions update <QUESTION_ID> --question "Updated text" --max-turns 5
deutero questions update <QID> --null-field explanation
```

**`deutero questions delete [QUESTION_ID]`**
Delete a question.

| Option | Description |
|--------|-------------|
| `--survey-id` | Survey ID the question belongs to (defaults to active survey) |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions delete <QUESTION_ID> --survey-id <SURVEY_ID>
```

**`deutero questions reorder [SURVEY_ID] <IDs...>`**
Reorder questions by passing space-separated question IDs in the desired order.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions reorder <SURVEY_ID> <Q1> <Q2> <Q3>
```

**`deutero questions upload-image [QUESTION_ID]`**
Upload an image to a question.

| Option | Description |
|--------|-------------|
| `--image-file <path>` | Local image file to upload |
| `--image-data <data-url>` | Base64 data URL to upload |
| `--label` | Optional image label |
| `-o, --output <file>` | Write JSON response to a file |

Exactly one of `--image-file` or `--image-data` must be provided.

```bash
deutero questions upload-image <QUESTION_ID> --image-file photo.png --label "Prototype A"
```

**`deutero questions reorder-images [QUESTION_ID] <IDs...>`**
Reorder images on a question by passing space-separated image IDs in the desired order.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero questions reorder-images <QUESTION_ID> <IMG1> <IMG2>
```

---

### `deutero personas`

Manage interviewee personas.

**`deutero personas list [SURVEY_ID]`**
List personas for a survey, including interview counts per persona.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero personas list
deutero personas list <SURVEY_ID>
```

**`deutero personas generate [SURVEY_ID]`**
Generate AI interviewee personas for a survey.

| Option | Description |
|--------|-------------|
| `-n, --count` | Number of personas to generate, 1–25 (required) |
| `-i, --instructions` | Additional instructions for generation |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero personas generate --count 3
deutero personas generate <SURVEY_ID> -n 5 -i "Focus on enterprise users"
```

---

### `deutero interviews`

Inspect and analyze interviews.

**`deutero interviews list [SURVEY_ID]`**
List interviews for a survey. Columns include ID, participant name, start time, completion status, simulated flag, and message count.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero interviews list
deutero interviews list <SURVEY_ID>
```

**`deutero interviews transcript [INTERVIEW_ID]`**
Pretty-print an interview transcript. If `INTERVIEW_ID` is omitted, an interactive list of interviews for the active survey is shown.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero interviews transcript
deutero interviews transcript <INTERVIEW_ID>
```

**`deutero interviews analysis [INTERVIEW_ID]`**
Run thematic analysis (phases 1–4) directly on a single interview.

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier |
| `-o, --output <file>` | — | Write JSON response to a file |

```bash
deutero interviews analysis <INTERVIEW_ID> --model-tier premium
```

---

### `deutero simulate`

Run simulated AI interviews.

**`deutero simulate run [SURVEY_ID] [PERSONA_ID]`**
Simulate an interview with a specific persona. Both IDs fall back to active context or interactive prompts.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero simulate run <SURVEY_ID> <PERSONA_ID>
deutero simulate run <PERSONA_ID>       # uses active survey
deutero simulate run                    # prompts for both
```

---

### `deutero analysis`

Run and inspect thematic analysis.

**`deutero analysis run`**
Submit a thematic analysis job. Requires `--interview-id` or `--survey-id`.

| Option | Default | Description |
|--------|---------|-------------|
| `--interview-id` | — | Run on a single interview |
| `--survey-id` | — | Run on all completed interviews in a survey |
| `-m, --model-tier` | `open_weights` | Model tier |
| `--cross-case` | `False` | Run cross-case analysis (requires `--survey-id`) |
| `-o, --output <file>` | — | Write JSON response to a file |
| `--xml-output <file>` | — | Write cross-case XML to a file |

```bash
deutero analysis run --interview-id <ID> --model-tier premium
deutero analysis run --survey-id <ID> --cross-case --xml-output cross_case.xml
```

**`deutero analysis status`**
Check the status of an analysis job.

| Option | Description |
|--------|-------------|
| `--interview-id` | Get status for a single interview |
| `--survey-id` | Get status for all interviews in a survey |
| `-o, --output <file>` | Write JSON response to a file |

```bash
deutero analysis status --interview-id <ID>
deutero analysis status --survey-id <ID>
```

**`deutero analysis results-interview [INTERVIEW_ID]`**
Retrieve per-phase analysis results for a single interview.

| Option | Description |
|--------|-------------|
| `-p, --phase` | Analysis phase: `initial_engagement`, `initial_noting`, `emergent_themes`, `connections` (required) |
| `-o, --output <file>` | Write XML result to a file |

```bash
deutero analysis results-interview <ID> --phase emergent_themes -o themes.xml
```

**`deutero analysis results-survey [SURVEY_ID]`**
Retrieve cross-case analysis results for a survey.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write XML result to a file |

```bash
deutero analysis results-survey
deutero analysis results-survey <SURVEY_ID> -o cross_case.xml
```

---

### `deutero webhooks`

Manage webhook endpoints for your organization. Signing secrets are never exposed through the CLI — use the Deutero dashboard to create endpoints and retrieve secrets.

**`deutero webhooks list`**
List all webhook endpoints configured for the authenticated organization.

| Option | Description |
|--------|-------------|
| `-o, --output <file>` | Write JSON response to a file |

Also prints the list of available event types.

```bash
deutero webhooks list
deutero webhooks list -o webhooks.json
```

**`deutero webhooks update <ENDPOINT_ID>`**
Update a webhook endpoint's label, delivery URL, enabled state, or subscribed events. Only supplied fields are changed.

| Option | Description |
|--------|-------------|
| `--label <text>` | New label for the endpoint |
| `--url <url>` | New delivery URL (must be `http` or `https`) |
| `--enable` | Enable the endpoint |
| `--disable` | Disable the endpoint |
| `--event <type>` | Subscribed event type (repeatable — replaces the existing list) |
| `-o, --output <file>` | Write JSON response to a file |

At least one option must be provided. When both `--enable` and `--disable` are passed, `--enable` takes effect.

```bash
deutero webhooks update <ENDPOINT_ID> --label "Production Hook"
deutero webhooks update <ENDPOINT_ID> --url https://hooks.example.com/deutero
deutero webhooks update <ENDPOINT_ID> --enable
deutero webhooks update <ENDPOINT_ID> --disable
deutero webhooks update <ENDPOINT_ID> \
  --event interview.completed \
  --event survey.completed
deutero webhooks update <ENDPOINT_ID> --label "Updated" --disable -o result.json
```

**`deutero webhooks delete <ENDPOINT_ID>`**
Permanently delete a webhook endpoint and all its delivery logs. Prompts for confirmation unless `--yes` is supplied.

| Option | Description |
|--------|-------------|
| `-y, --yes` | Skip the confirmation prompt |

```bash
deutero webhooks delete <ENDPOINT_ID>
deutero webhooks delete <ENDPOINT_ID> --yes
```

---

## Configuration

The CLI reads configuration from two sources (environment variables take precedence):

1. **Environment variables**:
   - `DEUTERO_CLIENT_ID` / `DEUTERO_STYTCH_PROJECT_ID` — for OAuth login
   - `DEUTERO_API_KEY` — for API key fallback
   - `DEUTERO_BASE_URL` — API base URL

2. **Config file**: `~/.deutero/config.json`

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

The `active_project_id` and `active_survey_id` fields are written by `deutero projects set-active` and `deutero surveys set-active` respectively. They are used as default arguments by all commands that accept a survey or project ID.

---

## Development

```bash
# Install in editable mode with dev dependencies and TUI support
pip install -e ".[dev,tui]"

# Run the dev version explicitly using the venv Python
# (avoids being shadowed by a system/conda `deutero` install)
.venv/bin/deutero --help
# or
python -m deutero_cli.cli --help

# Run tests
pytest

# Run tests with coverage
coverage run -m pytest && coverage report

# Lint
ruff check .

# Type check
mypy deutero_cli
```

---

## License

[MIT](LICENSE)
