# Generation and Importing Reference

Complete reference for AI-generating surveys, questions, and personas, and for importing existing research guides.

---

## AI Survey Generation

`deutero surveys generate` creates a fully-specified survey using AI. The survey type determines which input fields are relevant. All fields are optional but richer inputs produce better outputs.

### Common Options (all types)

| Option | Default | Description |
|--------|---------|-------------|
| `--survey-type` / `-t` | `user_experience` | Survey type: `user_experience`, `sociology`, `customer_development`, `polling` |
| `--language` / `-l` | `English` | Language for the generated study |
| `--model-tier` | `open_weights` | Model tier to set on the created survey |
| `-o, --output <file>` | — | Write full JSON response to a file |
| `--xml-output <file>` | — | Write the XML specification to a file |

### Type: `user_experience`

For UX and product research. Focuses on how people use a product and where they face friction.

| Option | Description |
|--------|-------------|
| `--business-context` | Business or product context — what the product does and its market |
| `--research-need` | Why the research is needed — what decision it will inform |
| `--target-users` | Primary audience — who will be interviewed |
| `--constraints` | Key constraints — timeline, budget, known issues |

```bash
deutero surveys generate \
  --survey-type user_experience \
  --business-context "B2B SaaS project management tool for engineering teams" \
  --research-need "Understand why users abandon the onboarding flow before creating their first project" \
  --target-users "Engineering managers at companies with 50–500 employees" \
  --constraints "Must work for remote-first teams" \
  --model-tier premium
```

### Type: `sociology`

For academic or social research exploring behaviors, attitudes, and group dynamics.

| Option | Description |
|--------|-------------|
| `--research-question` | The central research question |
| `--population-of-interest` | Who or what group is being studied |
| `--context-or-setting` | The social or environmental context |
| `--key-concepts` | Core theoretical concepts or constructs |
| `--scope-and-boundaries` | What is in and out of scope |

```bash
deutero surveys generate \
  --survey-type sociology \
  --research-question "How do online communities form and maintain trust among strangers?" \
  --population-of-interest "Reddit moderators with 2+ years of experience" \
  --context-or-setting "Volunteer-run moderation of large (100k+) subreddits" \
  --key-concepts "Social capital, norm enforcement, reputation systems" \
  --language Spanish
```

### Type: `customer_development`

For startup and product validation — testing problem hypotheses and understanding customer needs.

| Option | Description |
|--------|-------------|
| `--problem-hypothesis` | The problem you believe customers have |
| `--customer-segment` | The customer segment to validate with |
| `--solution-concept` | The proposed solution (optional, keeps focus on problem space) |
| `--key-assumptions` | Assumptions that must be true for the business to work |
| `--success-criteria` | What would make this research successful |

```bash
deutero surveys generate \
  --survey-type customer_development \
  --problem-hypothesis "Remote engineering teams waste 3+ hours per week on manual status update meetings" \
  --customer-segment "Engineering teams of 5–20 people working fully remotely" \
  --solution-concept "Async standup tool that auto-summarizes updates" \
  --key-assumptions "Teams currently hold synchronous standups, managers care about visibility" \
  --success-criteria "Confirm the problem is top-3 pain point for at least 60% of respondents"
```

### Type: `polling`

For structured opinion gathering and quantitative insights at scale.

| Option | Description |
|--------|-------------|
| `--research-question` | The central question or hypothesis to poll on |
| `--population-segment` | The target population segment |
| `--geographic-scope` | Geographic scope (e.g. "US adults", "EU tech workers") |
| `--survey-context` | Context and purpose of the poll |
| `--data-quality-requirements` | Requirements for data quality or representativeness |

```bash
deutero surveys generate \
  --survey-type polling \
  --research-question "What is public sentiment on mandatory AI content labeling?" \
  --population-segment "US adults aged 18–65 who use social media daily" \
  --geographic-scope "United States" \
  --survey-context "Research for a policy brief on AI regulation" \
  --data-quality-requirements "Balanced representation across age groups and political affiliations"
```

---

## Question Generation

Generate AI-powered interview questions for an existing survey.

```bash
deutero questions generate --count 5
deutero questions generate <SURVEY_ID> --count 10
deutero questions generate --count 8 --instructions "Focus on emotional responses and avoid yes/no questions"
deutero questions generate <SURVEY_ID> -n 5 -i "Include at least one image-based question" --xml-output qs.xml
```

| Option | Description |
|--------|-------------|
| `SURVEY_ID` | Optional — defaults to active survey |
| `-n, --count` | Number of questions to generate, 1–25 (required) |
| `-i, --instructions` | Additional instructions for the generator |
| `-o, --output <file>` | Write JSON response to a file |
| `--xml-output <file>` | Write XML specification to a file |

The generated questions are immediately created on the survey and returned in the response.

---

## Extract Study Details and Questions from a Research Guide

Parse an existing research guide document to extract structured study metadata and/or interview questions.

```bash
deutero surveys extract-from-guide guide.txt
deutero surveys extract-from-guide guide.txt --extract details    # metadata only
deutero surveys extract-from-guide guide.txt --extract questions  # questions only
deutero surveys extract-from-guide guide.txt --extract both       # default
deutero surveys extract-from-guide guide.txt -o extracted.json
```

`guide.txt` must be a local text file (UTF-8). The content should be the raw text of the research guide.

### Extracted `details` fields

| Field | Description |
|-------|-------------|
| `name` | Study name |
| `description` | Study description |
| `research_questions` | Research questions |
| `objectives` | Research objectives |
| `target_population` | Target population |

### Extracted `questions` fields

Each extracted question includes:

| Field | Description |
|-------|-------------|
| `question` | Question text |
| `type` | Question type (open, scale, multiple_choice, etc.) |
| `interviewer_guidance` | Notes for the AI interviewer |
| `scale` | Scale config, if applicable |
| `options` | Fixed-choice options, if applicable |

### Post-extraction workflow

After extracting, use the results to:
1. Create a survey: `deutero surveys create --project-id <ID> --name "..." --description "..."`
2. Create questions individually: `deutero questions create --question "..." --explanation "..."`

Or save the JSON and process it programmatically with `-o extracted.json`.

---

## Generate a Study Suggestion from a Website

Scrape a website URL and generate a study suggestion based on its content.

```bash
deutero surveys suggest-from-site https://myproduct.com
deutero surveys suggest-from-site https://myproduct.com --language Spanish
deutero surveys suggest-from-site https://myproduct.com -o suggestion.json
```

| Option | Default | Description |
|--------|---------|-------------|
| `URL` | required | The website URL to scrape |
| `--language` / `-l` | `English` | Language for the generated suggestion |
| `-o, --output <file>` | — | Write JSON response to a file |

The response includes an `is_valid` field indicating whether a valid suggestion was generated, along with the suggested study details.

---

## Persona Generation

Generate AI interviewee personas for a survey. Personas are used for simulated interviews.

```bash
deutero personas generate --count 3
deutero personas generate <SURVEY_ID> --count 5
deutero personas generate --count 4 --instructions "Enterprise users, risk-averse, non-technical decision-makers"
deutero personas generate -n 3 -i "Focus on early adopters willing to pay for premium tools" -o personas.json
```

| Option | Description |
|--------|-------------|
| `SURVEY_ID` | Optional — defaults to active survey |
| `-n, --count` | Number of personas to generate, 1–25 (required) |
| `-i, --instructions` | Additional instructions for persona generation |
| `-o, --output <file>` | Write JSON response to a file |

Personas are immediately created on the survey and stored. Use `deutero personas list` to view existing personas and their interview counts.

---

## Post-Generation Workflow

After generating a survey, set it as active to avoid repeating the survey ID:

```bash
# Find the new survey ID in the generation output, then:
deutero surveys set-active

# Verify
deutero config show

# Generate questions without specifying survey ID
deutero questions generate --count 5

# Generate personas without specifying survey ID
deutero personas generate --count 3
```
