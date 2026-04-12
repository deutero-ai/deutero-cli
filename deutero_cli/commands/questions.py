"""Question generation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_success, print_xml


@click.group("questions")
def questions_group() -> None:
    """Manage interview questions."""


@questions_group.command("generate")
@click.argument("survey_id")
@click.option("--count", "-n", "number_of_questions", type=int, required=True, help="Number of questions to generate (1–25).")
@click.option("--instructions", "-i", "additional_instructions", default=None, help="Additional instructions for question generation.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.option("--xml-output", default=None, help="Write the XML specification to a file.")
@click.pass_context
def questions_generate(
    ctx: click.Context,
    survey_id: str,
    number_of_questions: int,
    additional_instructions: Optional[str],
    output_file: Optional[str],
    xml_output: Optional[str],
) -> None:
    """Generate interview questions for a survey."""
    if number_of_questions < 1 or number_of_questions > 25:
        print_error("Number of questions must be between 1 and 25.")
        raise SystemExit(1)

    payload: dict = {
        "survey_id": survey_id,
        "number_of_questions": number_of_questions,
    }
    if additional_instructions:
        payload["additional_instructions"] = additional_instructions

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.generate_questions(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    if xml_output and result.get("xml_file"):
        print_xml(result["xml_file"], xml_output)

    question_list = result.get("question_list", [])
    print_success(f"Generated {len(question_list)} question(s) for survey {survey_id}")
    print_json(result, output_file)
