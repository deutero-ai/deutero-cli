"""Interview simulation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_success


@click.group("interviews")
def interviews_group() -> None:
    """Simulate interviews."""


@interviews_group.command("simulate")
@click.argument("survey_id")
@click.argument("persona_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def interviews_simulate(
    ctx: click.Context,
    survey_id: str,
    persona_id: str,
    output_file: Optional[str],
) -> None:
    """Simulate an interview for a survey persona.

    SURVEY_ID is the UUID of the survey.
    PERSONA_ID is the UUID of the persona.
    """
    payload = {
        "survey_id": survey_id,
        "persona_id": persona_id,
    }

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.simulate_interview(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Interview simulation started — transcript: {result.get('transcript_url', '—')}")
    print_json(result, output_file)


@interviews_group.command("transcript")
@click.argument("interview_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def interviews_transcript(ctx: click.Context, interview_id: str, output_file: Optional[str]) -> None:
    """Get the transcript for an interview."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_interview_transcript(interview_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    messages = result.get("messages", [])
    print_success(f"Transcript for interview {interview_id} — {len(messages)} message(s)")
    print_json(result, output_file)
