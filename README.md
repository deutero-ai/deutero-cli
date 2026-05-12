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

### Development install

```bash
git clone https://github.com/deutero/deutero-cli.git
cd deutero-cli
python -m venv .venv && source .venv/bin/activate
pip install -e ".[dev]"
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

Optionally set a custom base URL:

```bash
export DEUTERO_BASE_URL="https://api.deutero.app"
# or
deutero config set-url https://api.deutero.app
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

### 4. Generate personas

```bash
deutero personas generate <SURVEY_ID> --count 3
```

### 5. Simulate an interview

```bash
deutero interviews simulate <SURVEY_ID> <PERSONA_ID>
```

### 6. Run thematic analysis

```bash
# Phases 1–4 on a single interview
deutero analysis run --interview-id <INTERVIEW_ID> --model-tier premium

# Phases 1–4 on all completed interviews in a survey
deutero analysis run --survey-id <SURVEY_ID>

# Cross-case analysis (phase 5)
deutero analysis run --survey-id <SURVEY_ID> --cross-case --xml-output cross_case.xml
```

### 7. Check analysis status

```bash
deutero analysis status --interview-id <INTERVIEW_ID>
deutero analysis status --survey-id <SURVEY_ID>
```

### 8. Retrieve analysis results

```bash
# Per-interview phase results
deutero analysis results-interview <INTERVIEW_ID> --phase emergent_themes --output themes.xml

# Cross-case survey results
deutero analysis results-survey <SURVEY_ID> --output cross_case.xml
```

### 9. Check credit balance and estimates

```bash
# Current balance
deutero credits balance

# Estimate credits for simulating interviews
deutero credits estimate-simulation <SURVEY_ID> --model-tier premium --participants 10

# Estimate credits for thematic analysis
deutero credits estimate-analysis <SURVEY_ID> --model-tier open_weights --interviews 5

# Estimate credits for a full survey (interviews + optional analysis)
deutero credits estimate-survey <SURVEY_ID> --model-tier premium --participants 10 --include-analysis
```

### 10. List projects and surveys

```bash
# List all projects
deutero projects list

# List surveys in a project
deutero projects list-surveys <PROJECT_ID>
```

### 11. Inspect survey details

```bash
# Get the full question list for a survey
deutero surveys question-list <SURVEY_ID>

# List all interviews for a survey
deutero surveys interviews <SURVEY_ID>
```

### 12. Retrieve interview transcripts

```bash
deutero interviews transcript <INTERVIEW_ID>
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

Most data commands also accept:

- `--output / -o <file>` — write the full JSON response to a file
- `--xml-output <file>` — write XML content to a file (where applicable)

---

## Command Reference

```
deutero
├── auth
│   ├── login                       Authenticate via OAuth2 + PKCE
│   ├── logout                      Clear stored tokens
│   └── status                      Show auth status
├── config
│   ├── show                        View current configuration
│   ├── set-key <KEY>               Save API key
│   └── set-url <URL>               Save base URL
├── credits
│   ├── balance                     Get credit balance and reservations
│   ├── estimate-simulation <ID>    Estimate credits for simulation
│   ├── estimate-analysis <ID>      Estimate credits for analysis
│   └── estimate-survey <ID>        Estimate credits for full survey
├── projects
│   ├── list                        List all projects
│   └── list-surveys <ID>           List surveys for a project
├── surveys
│   ├── generate                    Generate a research survey
│   ├── participation <ID>          Get participation statistics
│   ├── agent-requirements <ID>     Get agent requirements markdown
│   ├── model-tier <ID>             Get or set survey model tier
│   ├── suggest-from-site <URL>     Generate study suggestion from a URL
│   ├── question-list <ID>          Get full question list for a survey
│   └── interviews <ID>             List interviews for a survey
├── questions
│   ├── generate <ID>               Generate interview questions
│   ├── create <SURVEY_ID>          Create a single question
│   ├── get <QUESTION_ID>           Get a question by ID
│   ├── update <QUESTION_ID>        Update a question
│   ├── delete <QUESTION_ID>        Delete a question
│   ├── reorder <SURVEY_ID> <IDs>   Reorder questions in a survey
│   ├── upload-image <QUESTION_ID>  Upload an image to a question
│   └── reorder-images <QID> <IDs>  Reorder images on a question
├── personas
│   └── generate <ID>               Generate interviewee personas
├── interviews
│   ├── simulate <SID> <PID>        Simulate an interview
│   └── transcript <ID>             Get interview transcript
└── analysis
    ├── run                         Run thematic analysis
    ├── status                      Get analysis status
    ├── results-interview <ID>      Get per-phase results
    └── results-survey <ID>         Get cross-case results
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
  "base_url": "https://api.deutero.app",
  "access_token": "...",
  "refresh_token": "...",
  "token_expiry": 1234567890,
  "client_id": "your-client-id",
  "project_id": "your-project-id"
}
```

---

## Development

```bash
# Install in editable mode with dev dependencies
pip install -e ".[dev]"

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
