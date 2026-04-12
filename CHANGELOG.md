# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-04-11

### Added

- Initial release of `deutero-cli`.
- `deutero config show|set-key|set-url` — manage API key and base URL.
- `deutero surveys generate` — generate research surveys (UX, sociology, customer development, polling).
- `deutero surveys participation` — retrieve survey participation statistics.
- `deutero surveys agent-requirements` — generate agent requirements markdown.
- `deutero questions generate` — generate interview questions for a survey.
- `deutero personas generate` — generate interviewee personas for a survey.
- `deutero interviews simulate` — simulate an interview for a survey persona.
- `deutero analysis run` — run thematic analysis (phases 1-4) or cross-case analysis (phase 5).
- `deutero analysis status` — check analysis progress.
- `deutero analysis results-interview` — retrieve per-phase XML output.
- `deutero analysis results-survey` — retrieve cross-case analysis XML output.
- Rich-formatted terminal output with JSON/XML syntax highlighting.
- Configuration via environment variables (`DEUTERO_API_KEY`, `DEUTERO_BASE_URL`) or `~/.deutero/config.json`.
