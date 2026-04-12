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

## Getting Your API Key

1. Log in to your [Deutero dashboard](https://app.deutero.ai)
2. Navigate to **Settings**
3. Scroll to the **API Keys** section
4. Click **Generate New API Key**
5. Copy the key immediately — it is shown only once
6. Store it securely (use an environment variable or a secrets manager; never commit it to version control)

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

### 1. Configure your API key

```bash
# Option A – environment variable (recommended)
export DEUTERO_API_KEY="your-api-key"

# Option B – save to ~/.deutero/config.json
deutero config set-key your-api-key
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
  --target-users "Engineering managers at mid-size companies"

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
```

### 3. Generate interview questions

```bash
deutero questions generate <SURVEY_ID> --count 5
deutero questions generate <SURVEY_ID> --count 10 --instructions "Focus on emotional responses"
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

---

## Global Options

Every command supports:

| Flag | Env var | Description |
|------|---------|-------------|
| `--api-key` | `DEUTERO_API_KEY` | API key for authentication |
| `--base-url` | `DEUTERO_BASE_URL` | API base URL |
| `--version` | — | Show version and exit |
| `--help` | — | Show help and exit |

Most data commands also accept:

- `--output / -o <file>` — write the full JSON response to a file
- `--xml-output <file>` — write XML content to a file (where applicable)

---

## Command Reference

```
deutero
├── config
│   ├── show                    View current configuration
│   ├── set-key <KEY>           Save API key
│   └── set-url <URL>           Save base URL
├── surveys
│   ├── generate                Generate a research survey
│   ├── participation <ID>      Get participation statistics
│   └── agent-requirements <ID> Get agent requirements markdown
├── questions
│   └── generate <ID>           Generate interview questions
├── personas
│   └── generate <ID>           Generate interviewee personas
├── interviews
│   └── simulate <SID> <PID>    Simulate an interview
└── analysis
    ├── run                     Run thematic analysis
    ├── status                  Get analysis status
    ├── results-interview <ID>  Get per-phase results
    └── results-survey <ID>     Get cross-case results
```

---

## Configuration

The CLI reads configuration from two sources (environment variables take precedence):

1. **Environment variables**: `DEUTERO_API_KEY`, `DEUTERO_BASE_URL`
2. **Config file**: `~/.deutero/config.json`

```json
{
  "api_key": "your-api-key",
  "base_url": "https://api.deutero.app"
}
```

---

## MCP Server Integration

Deutero provides a remote MCP server that connects your AI coding assistant (Cursor, Windsurf, Claude Code, and others) directly to the Deutero API. This lets your coding agent create studies, generate questions, run simulations, and retrieve research insights without leaving your IDE.

### Prerequisites

- Active Deutero account with a valid subscription and credits
- Your API key (see [Getting Your API Key](#getting-your-api-key) above)
- An MCP-compatible IDE (Cursor, Windsurf, Claude Desktop/Code, etc.)

### Available MCP Tools

| Tool | Category | Purpose |
|------|----------|---------|
| `create_study` | Study management | Create a new study (UX, sociology, customer dev, polling) |
| `create_study_questions` | Study management | Generate 1–25 interview questions for a study |
| `create_simulation_persona` | Simulation | Generate 1–25 AI personas for a study |
| `simulate_interviews` | Simulation | Run one simulated interview (survey + persona) |
| `run_thematic_analysis` | Analysis | Start thematic analysis (single interview or full survey; optional cross-case) |
| `get_analysis_status` | Analysis | Check analysis progress for an interview or survey |
| `get_agent_requirements` | Analysis | Get or generate LLM-oriented requirements Markdown from cross-case analysis |
| `get_survey_participation` | Participation | Get completion and quota statistics for a survey |

All tools require your API key via the `X-API-Key` header in the MCP configuration.

### Configuration

#### Cursor

**Global** — create or edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "deutero": {
      "url": "http://agents.deutero.ai/mcp",
      "headers": {
        "X-API-Key": "your-api-key-here"
      }
    }
  }
}
```

**Project-level** — create or edit `.cursor/mcp.json` in your project root (use an env var to avoid committing the key):

```json
{
  "mcpServers": {
    "deutero": {
      "url": "http://agents.deutero.ai/mcp",
      "headers": {
        "X-API-Key": "${env:DEUTERO_API_KEY}"
      }
    }
  }
}
```

#### Windsurf

**Global** — create or edit `~/.windsurf/mcp.json`:

```json
{
  "mcpServers": {
    "deutero": {
      "url": "http://agents.deutero.ai/mcp",
      "headers": {
        "X-API-Key": "${env:DEUTERO_API_KEY}"
      }
    }
  }
}
```

**Project-level** — create or edit `.windsurf/mcp.json` in your project root (same format as above).

#### Claude Desktop / Claude Code

**macOS** — edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "deutero": {
      "url": "http://agents.deutero.ai/mcp",
      "headers": {
        "X-API-Key": "your-api-key-here"
      }
    }
  }
}
```

**Windows** — edit `%APPDATA%\Claude\claude_desktop_config.json` with the same structure.

#### Other MCP-Compatible Tools

```json
{
  "mcpServers": {
    "deutero": {
      "url": "http://agents.deutero.ai/mcp",
      "headers": {
        "X-API-Key": "your-api-key-here"
      }
    }
  }
}
```

Consult your tool's documentation for the exact configuration file location.

### MCP Tool Reference

#### `create_study`

Create a new research study. Study type determines which fields are required; missing required fields are elicited interactively.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_type` | string | Yes | `user_experience`, `sociology`, `customer_development`, or `polling` |
| `language` | string | No | Language for the study (default: `English`) |
| `business_context` | string | UX only | Business or product context |
| `research_need` | string | UX only | Why the research is being done |
| `target_users` | string | UX only | Target audience or user segment |
| `constraints` | string | UX only | Constraints, timing, or other considerations |
| `research_question` | string | Sociology / Polling | The research question |
| `population_of_interest` | string | Sociology | Population of interest |
| `context_or_setting` | string | Sociology | Context or setting |
| `key_concepts` | string | Sociology | Key concepts |
| `scope_and_boundaries` | string | Sociology | Scope and boundaries |
| `problem_hypothesis` | string | Customer dev | Problem hypothesis |
| `customer_segment` | string | Customer dev | Customer segment |
| `solution_concept` | string | Customer dev | Solution concept |
| `key_assumptions` | string | Customer dev | Key assumptions |
| `success_criteria` | string | Customer dev | Success criteria |
| `population_segment` | string | Polling | Population segment |
| `geographic_scope` | string | Polling | Geographic scope |
| `survey_context` | string | Polling | Survey context |
| `data_quality_requirements` | string | Polling | Data quality requirements |

**Returns**: `study_id`, `study_name`, `study_description`, `research_questions`, `research_objectives`, `url`, `xml_file`.

---

#### `create_study_questions`

Generate interview questions for an existing study.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_id` | string (UUID) | Yes | ID of the survey |
| `number_of_questions` | integer | No | Number of questions to generate (1–25; default: 10) |
| `additional_instructions` | string | No | Extra instructions for question generation |

**Returns**: `survey_id`, `question_list`, `edit_questions_url`, `interview_url`, `xml_file`.

---

#### `create_simulation_persona`

Generate AI personas for simulated interviews.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_id` | string (UUID) | Yes | ID of the survey |
| `number_of_personas` | integer | No | Number of personas to generate (1–25; default: 1) |
| `additional_instructions` | string | No | Extra instructions for persona generation |

**Returns**: `survey_id`, `personas` (list of `{ persona, persona_id }`).

---

#### `simulate_interviews`

Run a simulated interview for a survey and persona. Can take up to 5 minutes.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_id` | string (UUID) | Yes | ID of the survey |
| `persona_id` | string | Yes | ID of the persona (from `create_simulation_persona`) |

**Returns**: `survey_id`, `persona_id`, `transcript_url`.

---

#### `run_thematic_analysis`

Start thematic analysis in the background (returns immediately; use `get_analysis_status` to poll).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `interview_id` | string (UUID) | One of | Run analysis on this single interview |
| `survey_id` | string (UUID) | One of | Run analysis on all completed interviews for this survey |
| `model_tier` | string | No | `open_weights`, `premium`, or `frontier` (default: `open_weights`) |
| `cross_case_analysis` | boolean | No | Run cross-case analysis across interviews; requires `survey_id` (default: false) |

**Returns**: `success`, `interviews_queued`, `interviews`, `message`.

---

#### `get_analysis_status`

Check analysis progress for an interview or survey.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `interview_id` | string (UUID) | One of | Status for a single interview |
| `survey_id` | string (UUID) | One of | Status for all interviews in a survey |

**Returns**: `interview_id`, `survey_id`, `total_interviews`, `interviews` (with phase completion and run metadata).

---

#### `get_agent_requirements`

Retrieve the LLM-oriented requirements Markdown document derived from cross-case thematic analysis. Cross-case analysis must have been run first.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_id` | string (UUID) | Yes | ID of the survey |

**Returns**: `survey_id`, `markdown`, `filename`.

---

#### `get_survey_participation`

Get participation and completion statistics for a survey.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `survey_id` | string (UUID) | Yes | ID of the survey |

**Returns**: `survey_id`, `total_interviews`, `completed_interviews`, `incomplete_interviews`, `max_responses`, `completion_rate`, and quota metrics if applicable.

---

### MCP Usage Example

```
You: I need to research why users churn from our SaaS product.

AI: [Uses create_study — survey_type: user_experience]
    Study "SaaS User Churn Analysis" created. Saved as saas-churn-study.xml.

    [Uses create_study_questions — number_of_questions: 10]
    10 questions generated. Review them in the XML file or at [edit_questions_url].

You: Create 3 personas and run simulations.

AI: [Uses create_simulation_persona — number_of_personas: 3]
    3 personas created.

    [Uses simulate_interviews × 3]
    Simulations running. Transcripts available at:
    1. "Power User" — [url]
    2. "Occasional User" — [url]
    3. "Churned User" — [url]

You: Run thematic analysis with cross-case, then get the agent requirements.

AI: [Uses run_thematic_analysis — survey_id, cross_case_analysis: true]
    Analysis started.

    [Uses get_analysis_status to poll until complete]
    [Uses get_agent_requirements]
    Here is the requirements Markdown for your coding agent...
```

### MCP Security Best Practices

- **Never commit API keys** to version control
- **Use environment variables** (`${env:DEUTERO_API_KEY}`) in project-level config files
- **Rotate keys regularly** from the Settings page in your Deutero dashboard
- **Use separate keys** for development and production environments

### Troubleshooting MCP

| Problem | Solutions |
|---------|-----------|
| Server not connecting | Verify API key; check connectivity to `agents.deutero.ai`; confirm config is valid JSON; restart your IDE |
| "Unauthorized" errors | Regenerate key from Settings; check for extra spaces or quotes in the key value |
| Tools not visible in IDE | Confirm MCP is enabled in IDE settings; check the config file path is correct; review IDE MCP logs |
| "Insufficient credits" | Check balance in Deutero dashboard; enable auto top-up; review your spend cap |
| Simulations timing out | Simulations normally complete in 2–5 minutes; check transcript URL after 5 minutes; verify the study has valid questions |

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
