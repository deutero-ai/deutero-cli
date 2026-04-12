"""Persona generation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_success


@click.group("personas")
def personas_group() -> None:
    """Manage interviewee personas."""


@personas_group.command("generate")
@click.argument("survey_id")
@click.option("--count", "-n", "number_of_personas", type=int, required=True, help="Number of personas to generate (1–25).")
@click.option("--instructions", "-i", "additional_instructions", default=None, help="Additional instructions for persona generation.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def personas_generate(
    ctx: click.Context,
    survey_id: str,
    number_of_personas: int,
    additional_instructions: Optional[str],
    output_file: Optional[str],
) -> None:
    """Generate interviewee personas for a survey."""
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
    print_json(result, output_file)
