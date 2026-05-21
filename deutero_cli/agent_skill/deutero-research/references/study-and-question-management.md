# Study and Question Management Reference

Complete reference for creating, reading, updating, and deleting surveys and interview questions via `deutero-cli`.

---

## Projects

### List Projects

```bash
deutero projects list
deutero projects list -o projects.json
```

Returns a table of all projects with ID and name.

### List Surveys in a Project

```bash
deutero projects list-surveys
deutero projects list-surveys <PROJECT_ID>
```

Falls back to the active project if `PROJECT_ID` is omitted.

### Set Active Project

```bash
deutero projects set-active
```

Shows an interactive picker. Saves the chosen project ID to `~/.deutero/config.json`.

---

## Surveys

### Survey Properties

| Field | CLI option | Values / Notes |
|-------|-----------|----------------|
| `name` | `--name` | Display name (alias) |
| `description` | `--description` | Free text |
| `survey_type` | `--survey-type` / `-t` | `user_experience`, `sociology`, `customer_development`, `polling` |
| `research_questions` | `--research-questions` | The study's research questions |
| `objectives` | `--objectives` | Research objectives |
| `target_population` | `--target-population` | Who the study targets |
| `anonymous` | `--anonymous` / `--no-anonymous` | Whether responses are anonymous |
| `model_tier` | `--model-tier` | `open_weights`, `premium`, `frontier` |
| `redirect_url` | `--redirect-url` | URL participants are sent to after completing the survey |
| `language` | `--language` / `-l` | e.g. `English`, `Spanish`, `Portuguese` |
| `project_id` | `--project-id` | Required on `create`, not on `update` |

### Create a Survey Directly

```bash
deutero surveys create --project-id <PROJECT_ID>

deutero surveys create \
  --project-id <PROJECT_ID> \
  --survey-type sociology \
  --name "Trust in Online Communities" \
  --description "Explores how trust is formed in Reddit communities" \
  --model-tier premium \
  --language English

deutero surveys create \
  --project-id <PROJECT_ID> \
  --anonymous \
  --redirect-url https://example.com/thank-you \
  -o created_study.json
```

`--project-id` is the only required field. All other fields are optional; omitted fields are left as platform defaults.

### Update a Survey

```bash
deutero surveys update <SURVEY_ID> --name "New Name"
deutero surveys update <SURVEY_ID> --model-tier frontier --language Spanish
deutero surveys update <SURVEY_ID> --anonymous
deutero surveys update <SURVEY_ID> --no-anonymous --redirect-url https://example.com/done
deutero surveys update <SURVEY_ID> --description "Updated description" -o result.json
```

At least one option must be provided. Only supplied fields are changed; omitted fields remain untouched.

When both `--anonymous` and `--no-anonymous` are passed, `--anonymous` takes effect.

### List Surveys

```bash
deutero surveys list                   # uses active project
deutero surveys list <PROJECT_ID>
```

### Set Active Survey

```bash
deutero surveys set-active
deutero surveys set-active --project-id <PROJECT_ID>
```

Interactive picker. Saves chosen survey ID to `~/.deutero/config.json`.

### Get the Question List for a Survey

```bash
deutero surveys question-list
deutero surveys question-list <SURVEY_ID>
```

Returns a table showing question IDs, text, and explanation. Uses the active survey if `SURVEY_ID` is omitted.

### Get Participation Statistics

```bash
deutero surveys participation
deutero surveys participation <SURVEY_ID>
```

Returns:

| Field | Description |
|-------|-------------|
| `total_interviews` | Total interview attempts |
| `completed_interviews` | Successfully completed interviews |
| `incomplete_interviews` | Started but not completed |
| `completion_rate` | `completed / total * 100` |
| `quota_fill_rate` | Percentage of quota filled (if a quota is set) |
| `quota_remaining` | Remaining quota slots |

### Get Agent Requirements

```bash
deutero surveys agent-requirements
deutero surveys agent-requirements <SURVEY_ID> -o requirements.md
```

Retrieves or generates a Markdown file describing the survey's requirements for AI agents conducting interviews. Contains the interview URL for distributing to real participants.

### Get or Set Model Tier

```bash
deutero surveys model-tier                         # get current tier
deutero surveys model-tier --set premium           # set tier on active survey
deutero surveys model-tier <SURVEY_ID> --set frontier
```

### List Interviews for a Survey (via surveys group)

```bash
deutero surveys interviews
deutero surveys interviews <SURVEY_ID>
```

Returns columns: `id`, `participant_name`, `start_time`, `end_time`, `completed`, `simulated`, `termination_reason`.

---

## Questions

### Question Properties

| Property | CLI option | Description |
|----------|-----------|-------------|
| `question` | `--question` | The question text shown to participants |
| `explanation` | `--explanation` | Interviewer guidance — context for the AI interviewer |
| `follow_up` | `--follow-up` / `--no-follow-up` | Whether the AI can ask follow-up questions |
| `min_turns` | `--min-turns` | Minimum conversation turns before moving on |
| `max_turns` | `--max-turns` | Maximum conversation turns before moving on |
| `scale` | `--scale-json` | JSON object defining a numeric/labeled scale |
| `options` | `--option` (repeatable) | Fixed-choice options (multiple choice) |
| `slots` | `--slot` (repeatable) | Named slots for structured responses |
| `expected_image` | `--expected-image` | Description of expected image input from participant |

### Question Types

#### Open-Ended (with Follow-up)

The default type. The AI interviewer asks the question and may probe deeper.

```bash
deutero questions create \
  --question "Walk me through how you currently handle onboarding new team members." \
  --explanation "We want to understand the manual steps they take and where they feel friction." \
  --follow-up \
  --min-turns 2 \
  --max-turns 6
```

Use `--no-follow-up` to limit the AI to a single exchange on this question.

#### Scale Questions

Pass a `--scale-json` object to define a rating scale.

```bash
deutero questions create \
  --question "On a scale of 1–10, how satisfied are you with your current onboarding process?" \
  --scale-json '{"min": 1, "max": 10, "min_label": "Very unsatisfied", "max_label": "Very satisfied"}' \
  --follow-up
```

The scale JSON object should include `min`, `max`, and optionally `min_label` and `max_label`.

#### Multiple Choice

Use `--option` (repeatable) to define fixed choices:

```bash
deutero questions create \
  --question "Which of these best describes your role?" \
  --option "Engineering Manager" \
  --option "Product Manager" \
  --option "Designer" \
  --option "Individual Contributor"
```

#### Slot-Based Questions

Use `--slot` (repeatable) to capture structured named values from the participant's response:

```bash
deutero questions create \
  --question "Tell me about the last time you onboarded a new hire. What was their role, how long did it take, and what was the biggest challenge?" \
  --slot "role" \
  --slot "duration" \
  --slot "biggest_challenge"
```

#### Image-Based Questions

Use `--expected-image` to signal that this question expects the participant to submit an image. After creating the question, upload images for the participant to react to using `upload-image`.

```bash
# Create the question
deutero questions create \
  --question "Please look at these two interface designs and tell me which feels more intuitive and why." \
  --expected-image "Two UI screenshots side by side" \
  --follow-up

# Upload images to the question
deutero questions upload-image <QUESTION_ID> \
  --image-file design_a.png \
  --label "Design A"

deutero questions upload-image <QUESTION_ID> \
  --image-file design_b.png \
  --label "Design B"

# Reorder images
deutero questions reorder-images <QUESTION_ID> <IMG_B_ID> <IMG_A_ID>
```

For `upload-image`, provide exactly one of:
- `--image-file <path>` — local file, automatically base64-encoded and MIME-typed
- `--image-data <data-url>` — pre-encoded base64 data URL

### Create a Question

```bash
deutero questions create [SURVEY_ID] --question "Question text" [options...]
```

`SURVEY_ID` defaults to the active survey. `--question` is required. All other fields are optional.

### Get a Question

```bash
deutero questions get <QUESTION_ID>
deutero questions get <QUESTION_ID> -o question.json
```

### Update a Question

```bash
deutero questions update <QUESTION_ID> --question "Revised question text"
deutero questions update <QUESTION_ID> --max-turns 4 --follow-up
deutero questions update <QUESTION_ID> --option "Option A" --option "Option B"  # replaces all options
deutero questions update <QUESTION_ID> --scale-json '{"min": 1, "max": 5}'

# Clear a field by sending null
deutero questions update <QUESTION_ID> --null-field explanation
deutero questions update <QUESTION_ID> --null-field scale
```

Valid `--null-field` values: `question`, `explanation`, `scale`, `options`, `slots`, `follow_up`, `min_turns`, `max_turns`, `expected_image`.

When using `--option` or `--slot` on update, the full list **replaces** the existing options/slots.

Use `--payload-json` to pass additional fields as a raw JSON object (merged with other options):
```bash
deutero questions update <QUESTION_ID> --payload-json '{"custom_field": "value"}'
```

### Delete a Question

```bash
deutero questions delete <QUESTION_ID>
deutero questions delete <QUESTION_ID> --survey-id <SURVEY_ID>
```

`--survey-id` defaults to the active survey.

### Reorder Questions

```bash
deutero questions reorder <SURVEY_ID> <Q1_ID> <Q2_ID> <Q3_ID>
```

Pass question IDs as space-separated arguments in the desired display order. `SURVEY_ID` defaults to the active survey.

### Upload an Image to a Question

```bash
deutero questions upload-image <QUESTION_ID> --image-file photo.png --label "Prototype A"
deutero questions upload-image <QUESTION_ID> --image-data "data:image/png;base64,..." --label "Variant B"
```

Exactly one of `--image-file` or `--image-data` must be supplied.

### Reorder Images on a Question

```bash
deutero questions reorder-images <QUESTION_ID> <IMG1_ID> <IMG2_ID> <IMG3_ID>
```

Pass image IDs as space-separated arguments in the desired order.
