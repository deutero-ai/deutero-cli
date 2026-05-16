"""Persona generation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_survey_id
from deutero_cli.output import print_error, print_items_table, print_json, print_success


@click.group("personas")
def personas_group() -> None:
    """Manage interviewee personas."""


@personas_group.command("list")
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def personas_list(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
    """List personas for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).

    Use -o/--output to save the full JSON response to a file.
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_personas(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    personas = result.get("personas", [])
    print_success(f"Found {result.get('total', len(personas))} persona(s) for survey {survey_id}")
    print_items_table(
        personas,
        title="Personas",
        columns=["persona_id", "source", "interview_count", "content"],
    )
    if output_file:
        print_json(result, output_file)


@personas_group.command("generate")
@click.argument("survey_id", required=False, default=None)
@click.option("--count", "-n", "number_of_personas", type=int, required=True, help="Number of personas to generate (1–25).")
@click.option("--instructions", "-i", "additional_instructions", default=None, help="Additional instructions for persona generation.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def personas_generate(
    ctx: click.Context,
    survey_id: Optional[str],
    number_of_personas: int,
    additional_instructions: Optional[str],
    output_file: Optional[str],
) -> None:
    """Generate interviewee personas for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")

    if number_of_personas < 1 or number_of_personas > 25:
        print_error("Number of personas must be between 1 and 25.")
        raise SystemExit(1)

    payload: dict = {
        "survey_id": survey_id,
        "number_of_personas": number_of_personas,
    }
    if additional_instructions:
        payload["additional_instructions"] = additional_instructions

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.generate_personas(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    personas = result.get("personas", [])
    print_success(f"Generated {len(personas)} persona(s) for survey {survey_id}")
    print_items_table(personas, title="Personas", columns=["id", "name", "description"])
    if output_file:
        print_json(result, output_file)
