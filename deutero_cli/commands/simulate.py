"""Interview simulation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_survey_id
from deutero_cli.output import print_error, print_json, print_key_value, print_success


@click.group("simulate")
def simulate_group() -> None:
    """Simulate interviews."""


@simulate_group.command("run")
@click.argument("survey_id", required=False, default=None)
@click.argument("persona_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def simulate_run(
    ctx: click.Context,
    survey_id: Optional[str],
    persona_id: Optional[str],
    output_file: Optional[str],
) -> None:
    """Simulate an interview for a survey persona.

    SURVEY_ID is the UUID of the survey.
    PERSONA_ID is the UUID of the persona.
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")

    if persona_id is None:
        persona_id = click.prompt("Persona ID")

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
    print_key_value(
        {
            "Interview ID": result.get("interview_id") or result.get("id"),
            "Survey ID": result.get("survey_id"),
            "Persona ID": result.get("persona_id"),
            "Status": result.get("status"),
            "Transcript URL": result.get("transcript_url"),
        },
        title="Simulation Result",
    )
    if output_file:
        print_json(result, output_file)
