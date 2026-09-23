# Acceptance Report: deutero

  Level: Full Dogfood, with write lifecycle
  Tests: 388/388 passed (312 skipped by the runner — see below)
  Gate: PASS
  Marker: proofs/phase5-acceptance.json, written by `dogfood --live --write-acceptance`

## How the write lifecycle was run

The operator approved a full run including writes. The runner was **not** given
`--allow-destructive`: that flag re-enables live mutating Example probes, and at the time
several generated examples (`webhooks delete`, `webhooks rotate-secret`) pointed at a real
webhook in the operator's account. Instead:

- The runner covered every command live, with mutating commands exercised as `--dry-run`.
  Its 312 skips are those live-mutation and destructive probes, by design.
- The write lifecycle was run by hand against resources created for the purpose, never
  against an existing one. All destructive examples were then re-pointed at obviously fake
  ids (`00000000-0000-4000-8000-…`) so no future `--allow-destructive` run can touch a real
  resource.

| Step | Result |
|---|---|
| Create disposable project → appears in list | pass |
| Create disposable study → in list, `get` returns it | pass |
| Update study name → change visible | pass |
| Question create → listed → update → edit visible → delete → gone | pass |
| Webhook create → events subscribed → secret shown once | pass |
| Webhook update (relabel, disable) → visible | pass |
| Webhook rotate-secret → new secret returned | pass |
| Webhook delivery log read | pass |
| Webhook delete → gone | pass |
| `study apply --confirm` clone onto disposable study → 4 sections applied, graph skipped with reason, diff shrinks accordingly | pass (after fix below) |

**Left behind:** the API has no delete endpoint for projects or studies, so one project and
one study remain, both named "ZZ … dogfood — delete me":
project `ffc9e3f8-49c9-4f5a-a6aa-51253de09555`, study `d074b9bf-c018-4d1c-8bd5-0b6ffafaecbf`.
Delete them from the dashboard.

## Failures found and fixed during Phase 5

Starting point: 45 failures of 299. Final: 0 of 388.

**In this CLI's code**
- `study apply` sent each section's GET shape to its PUT endpoint; 4 of 5 sections failed
  validation on the first live run. Rewritten with a per-section builder: settings objects
  only, derived/read-only fields dropped, **`short_url_slug` never copied** (slugs are unique
  platform-wide, so a clone that keeps one always 409s), and a linear source's null flow
  skipped with a stated reason. 3 tests added.
- `study diff` reported derived per-study state (counts, participation URLs) as instrument
  differences. Now ignored.
- Eleven hand-written Cobra Examples still used an unreachable study id, a shell redirect
  passed as literal arguments, and a nonexistent bundle file.

**In the spec's examples** — 34 generated commands had no Example (body-only commands get
none from the generator); fixed with `x-pp-example`. Tally endpoints pointed at a free-text
question; re-pointed at a real scale/choices question via `x-happy-args`.

**In generator templates (patched locally, filed for the retro)** — `doctor`, `sync`,
`workflow archive`, `feedback list`, `profile list` did not honour the `--dry-run --json`
envelope. See `proofs/retro-candidates.md`.

## Printing Press issues: 17
Full list in `proofs/retro-candidates.md`.

## Upstream (Deutero API) issues found

1. **Transcript search 500s in its default mode.** `search_transcripts` defaults to
   `hybrid`, which runs the trigram query; SQLAlchemy's `.op("%")` renders a bare `%`, which
   the pg8000 driver rejects ("Only %s and %% are supported"). Every string and hybrid
   search fails, for every caller including every MCP client. Fixed in the dashboard repo
   (`study_api/routers/search.py`, `.op("%%")`), verified against the real database, with a
   regression test that fails on the original code and passes on the fix. **Not yet
   deployed** — until it is, the CLI's search example uses `--mode semantic`.
2. **Webhook deliveries carried no participant id** — fixed earlier in this run (migration +
   model + capture + schema + filter + tests). Not yet deployed.

## Environment note

Little Snitch on the operator's machine blocked each freshly built binary until a
path-scoped rule was added; this accounted for every "network" failure mid-run.
