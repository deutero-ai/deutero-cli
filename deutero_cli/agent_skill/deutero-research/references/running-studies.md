# Running Studies Reference

Complete reference for simulating interviews, distributing real participant URLs, tracking participation, and estimating credits.

---

## Simulated Interviews

Simulation runs the interview using an AI persona instead of a real participant. Simulated interviews appear in the interview list with `simulated: true`.

### Run a Simulation

```bash
# Fully explicit
deutero simulate run <SURVEY_ID> <PERSONA_ID>

# Uses active survey — prompts only for persona
deutero simulate run <PERSONA_ID>

# Both IDs resolved interactively
deutero simulate run

# Save JSON response
deutero simulate run <SURVEY_ID> <PERSONA_ID> -o simulation.json
```

The `SURVEY_ID` argument defaults to the active survey if set. `PERSONA_ID` is always required (by argument or prompt).

Both IDs fall back to:
1. The active survey/project in `~/.deutero/config.json`
2. An interactive prompt

### Simulate via interviews group (alias)

```bash
deutero interviews simulate <SURVEY_ID> <PERSONA_ID>
```

This is an alias for `deutero simulate run`.

### Simulation workflow

1. Generate personas first: `deutero personas generate --count 3`
2. List personas to get their IDs: `deutero personas list`
3. Run simulation: `deutero simulate run <PERSONA_ID>`
4. Check the resulting interview: `deutero interviews list`
5. Read the transcript: `deutero interviews transcript <INTERVIEW_ID>`

---

## Real Participant Interviews

Real participants take the interview through a web URL. The CLI is used to retrieve and distribute that URL.

### Get the Interview URL and Agent Requirements

```bash
deutero surveys agent-requirements
deutero surveys agent-requirements <SURVEY_ID>
deutero surveys agent-requirements -o requirements.md
```

This retrieves or generates a Markdown file that includes:
- The survey's interview URL for real participants
- Agent requirements describing how an AI agent should conduct the interview
- Context about the study for the interviewing agent

The output file is written to the path specified by `-o`, or to the filename returned by the API (usually something like `requirements.md`).

Distribute the interview URL from that file to your participants via email, Slack, or any other channel.

---

## Listing Interviews

### Via interviews group

```bash
deutero interviews list
deutero interviews list <SURVEY_ID>
deutero interviews list <SURVEY_ID> -o interviews.json
```

Columns in output table:

| Column | Description |
|--------|-------------|
| `id` | Interview UUID |
| `participant_name` | Name provided by participant (or "Anon" if anonymous) |
| `start_time` | When the interview started (ISO 8601) |
| `completed` | `true` (finished), `false` (error/terminated), or `null` (in progress) |
| `simulated` | Whether this was a simulated interview |
| `message_count` | Number of messages exchanged |

### Via surveys group

```bash
deutero surveys interviews
deutero surveys interviews <SURVEY_ID>
```

Same data but with additional columns: `end_time` and `termination_reason`.

---

## Participation Statistics

```bash
deutero surveys participation
deutero surveys participation <SURVEY_ID>
deutero surveys participation <SURVEY_ID> -o participation.json
```

| Field | Description |
|-------|-------------|
| `total_interviews` | All interview attempts (started) |
| `completed_interviews` | Interviews fully completed |
| `incomplete_interviews` | Started but not completed |
| `completion_rate` | `completed / total × 100` |
| `quota_fill_rate` | Percentage of quota filled (null if no quota set) |
| `quota_remaining` | Remaining quota slots (null if no quota set) |

Use this to monitor progress during fielding and decide when to close the survey.

---

## Interview Transcripts

```bash
# Interactive picker (shows list of interviews for active survey)
deutero interviews transcript

# Direct retrieval
deutero interviews transcript <INTERVIEW_ID>

# Save raw JSON to file
deutero interviews transcript <INTERVIEW_ID> -o transcript.json
```

The transcript command pretty-prints the conversation using the Rich library:
- Agent messages are styled distinctly from participant messages
- Message timestamps are shown
- Questions and follow-ups are labelled

When `INTERVIEW_ID` is omitted, an interactive picker shows all interviews for the active survey with participant name, start time, completion status, and message count.

---

## Personas

### List Personas

```bash
deutero personas list
deutero personas list <SURVEY_ID>
deutero personas list <SURVEY_ID> -o personas.json
```

Returns a table showing each persona's ID, name/description, and how many interviews have been conducted with them. Useful for identifying which personas have already been simulated.

### Generate Personas

```bash
deutero personas generate --count 3
deutero personas generate <SURVEY_ID> --count 5
deutero personas generate --count 4 \
  --instructions "Enterprise buyers, budget-conscious, skeptical of new tools"
```

See `generation-and-importing.md` for full generation options.

---

## Credit Estimation

Estimate credit consumption before running simulations or analysis.

### Check Current Balance

```bash
deutero credits balance
deutero credits balance -o balance.json
```

Returns available credits, reservations, usage, limits, and trial status.

### Estimate Simulation Credits

```bash
deutero credits estimate-simulation
deutero credits estimate-simulation <SURVEY_ID>
deutero credits estimate-simulation --model-tier premium --participants 10
deutero credits estimate-simulation <SURVEY_ID> -m frontier -n 5
```

| Option | Default | Description |
|--------|---------|-------------|
| `SURVEY_ID` | active survey | Survey to estimate for |
| `-m, --model-tier` | `open_weights` | Model tier |
| `-n, --participants` | `1` | Number of participants to simulate |

### Estimate Analysis Credits

```bash
deutero credits estimate-analysis --model-tier premium --interviews 5
deutero credits estimate-analysis -m open_weights -n 10
```

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier |
| `-n, --interviews` | `1` | Number of interviews to analyze |

### Estimate Full Survey Credits

```bash
deutero credits estimate-survey --model-tier premium --participants 10
deutero credits estimate-survey --model-tier premium --participants 10 --include-analysis
```

| Option | Default | Description |
|--------|---------|-------------|
| `-m, --model-tier` | `open_weights` | Model tier |
| `-n, --participants` | `1` | Number of participants |
| `--include-analysis` | `false` | Include analysis cost in estimate |

---

## Typical Study Run Checklist

1. **Set active context**: `deutero projects set-active` → `deutero surveys set-active`
2. **Check questions**: `deutero surveys question-list`
3. **Generate personas**: `deutero personas generate --count 3`
4. **Estimate credits**: `deutero credits estimate-simulation -m premium -n 3`
5. **Run simulations**: `deutero simulate run` (repeat for each persona)
6. **Check participation**: `deutero surveys participation`
7. **Review transcripts**: `deutero interviews transcript` (interactive picker)
8. **Run analysis**: `deutero analysis run --survey-id <ID> --model-tier premium`
9. **Fetch results**: `deutero analysis results-survey -o cross_case.xml`
