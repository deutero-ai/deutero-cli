# Analysis and Results Reference

Complete reference for running thematic analysis, polling status, and retrieving results and transcripts.

---

## Analysis Overview

Deutero uses a multi-phase thematic analysis framework based on qualitative research methodology. Analysis is asynchronous — submit a job, then poll for status, then retrieve results.

### The Five Analysis Phases

| Phase | Scope | Description |
|-------|-------|-------------|
| `initial_engagement` | Per interview | Initial reading impressions — how the participant engaged, tone, notable moments |
| `initial_noting` | Per interview | First-pass noting — interesting, recurring, or significant items without interpretation |
| `emergent_themes` | Per interview | Themes that emerge from the data — patterns, tensions, unexpected insights |
| `connections` | Per interview | Connections between themes — how they relate, reinforce, or contradict each other |
| Cross-case (phase 5) | Whole survey | Patterns across all interviews — what holds across participants, what varies |

Phases 1–4 run per-interview. Phase 5 (cross-case) runs across all completed interviews in the survey.

---

## Running Analysis

### Per-Interview Analysis (Phases 1–4)

```bash
deutero analysis run --interview-id <INTERVIEW_ID>
deutero analysis run --interview-id <INTERVIEW_ID> --model-tier premium
deutero analysis run --interview-id <INTERVIEW_ID> -m frontier -o analysis.json
```

### Survey-Wide Analysis (All Completed Interviews)

```bash
deutero analysis run --survey-id <SURVEY_ID>
deutero analysis run --survey-id <SURVEY_ID> --model-tier premium
```

Submits phases 1–4 for all completed interviews in the survey.

### Cross-Case Analysis (Phase 5)

```bash
deutero analysis run --survey-id <SURVEY_ID> --cross-case
deutero analysis run --survey-id <SURVEY_ID> --cross-case --xml-output cross_case.xml
deutero analysis run --survey-id <SURVEY_ID> --cross-case -m premium
```

Cross-case analysis requires `--survey-id`. It synthesizes findings across all interviews into a unified view of patterns, themes, and variations.

| Option | Default | Description |
|--------|---------|-------------|
| `--interview-id` | — | Run on a single interview |
| `--survey-id` | — | Run on all completed interviews |
| `-m, --model-tier` | `open_weights` | Model tier |
| `--cross-case` / `--no-cross-case` | `false` | Include cross-case (phase 5) |
| `-o, --output <file>` | — | Write JSON response to a file |
| `--xml-output <file>` | — | Write cross-case XML to a file |

### Shortcut via interviews group

```bash
deutero interviews analysis <INTERVIEW_ID>
deutero interviews analysis <INTERVIEW_ID> --model-tier premium
deutero interviews analysis                # prompts for interview ID
```

This is equivalent to `deutero analysis run --interview-id <ID>` with `--cross-case false`.

---

## Checking Analysis Status

```bash
deutero analysis status --interview-id <INTERVIEW_ID>
deutero analysis status --survey-id <SURVEY_ID>
deutero analysis status --interview-id <ID> -o status.json
```

At least one of `--interview-id` or `--survey-id` must be provided.

The response shows the analysis state for each interview (e.g. `pending`, `in_progress`, `completed`, `failed`) across all phases.

---

## Retrieving Analysis Results

Results are returned as XML documents that follow Deutero's thematic analysis schema.

### Per-Interview Phase Results

```bash
deutero analysis results-interview <INTERVIEW_ID> --phase <PHASE>
deutero analysis results-interview <INTERVIEW_ID> --phase emergent_themes -o themes.xml
deutero analysis results-interview <INTERVIEW_ID> --phase connections -o connections.xml
```

| Option | Description |
|--------|-------------|
| `INTERVIEW_ID` | Interview UUID (prompts if omitted) |
| `-p, --phase` | **Required.** One of: `initial_engagement`, `initial_noting`, `emergent_themes`, `connections` |
| `-o, --output <file>` | Write XML result to a file (prints to stdout if omitted) |

### Cross-Case Survey Results

```bash
deutero analysis results-survey
deutero analysis results-survey <SURVEY_ID>
deutero analysis results-survey <SURVEY_ID> -o cross_case.xml
```

`SURVEY_ID` defaults to the active survey. Returns XML covering the cross-case analysis (phase 5).

### Working with XML Results

Results are written as XML files. Pass `-o results.xml` to save them. If no output file is specified, the XML is printed to stdout.

The XML structure follows Deutero's analysis schema. Typical usage:
- Load into an XML parser or XSLT transform
- Pass directly to the Deutero MCP server for agent consumption
- Include in reports or documentation

---

## Interview Transcripts

### Retrieve a Transcript

```bash
# Interactive picker (lists interviews for active survey)
deutero interviews transcript

# Direct retrieval by ID
deutero interviews transcript <INTERVIEW_ID>

# Save raw JSON to file
deutero interviews transcript <INTERVIEW_ID> -o transcript.json
```

The transcript is pretty-printed using Rich formatting, showing the full conversation between the AI interviewer and the participant.

When called without `INTERVIEW_ID`, an interactive list is shown:
- Participant name
- Start time (formatted)
- Completion status (`Completed` / `Error` / `Pending`)
- Message count
- Interview UUID

### Transcript JSON Structure

The JSON response (available via `-o`) contains a `messages` array. Each message has:
- `role` — `interviewer` or `participant`
- `content` — message text
- `timestamp` — ISO 8601 datetime
- `question_id` — the survey question this message belongs to (if applicable)

---

## Full Analysis Workflow Example

```bash
# 1. Ensure active survey is set
deutero config show

# 2. Check which interviews are completed
deutero interviews list

# 3. Run survey-wide analysis (phases 1–4)
deutero analysis run --survey-id <SURVEY_ID> --model-tier premium

# 4. Poll status until complete
deutero analysis status --survey-id <SURVEY_ID>

# 5. Run cross-case analysis
deutero analysis run --survey-id <SURVEY_ID> --cross-case --model-tier premium

# 6. Retrieve per-interview emergent themes
deutero analysis results-interview <INTERVIEW_ID> --phase emergent_themes -o themes_1.xml

# 7. Retrieve cross-case results
deutero analysis results-survey -o cross_case.xml

# 8. Review specific transcripts
deutero interviews transcript <INTERVIEW_ID>
```

---

## Model Tier Selection for Analysis

| Tier | Recommendation |
|------|---------------|
| `open_weights` | Exploratory analysis, early-stage research, cost-sensitive projects |
| `premium` | Standard research deliverables, client-facing outputs |
| `frontier` | High-stakes research, complex or nuanced subject matter |

The model tier can be set on the survey itself (`deutero surveys update <ID> --model-tier premium`) or passed per-command with `-m`/`--model-tier`.
