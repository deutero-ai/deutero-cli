"""Question generation commands."""

from __future__ import annotations

import base64
import json
import mimetypes
from pathlib import Path
from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_success, print_xml


@click.group("questions")
def questions_group() -> None:
    """Manage interview questions."""


def _json_object(value: Optional[str], option_name: str) -> Optional[dict]:
    if value is None:
        return None
    try:
        parsed = json.loads(value)
    except json.JSONDecodeError as exc:
        raise click.BadParameter(f"{option_name} must be valid JSON: {exc}") from exc
    if not isinstance(parsed, dict):
        raise click.BadParameter(f"{option_name} must be a JSON object.")
    return parsed


def _question_payload(
    question: Optional[str],
    explanation: Optional[str],
    scale_json: Optional[str],
    options: tuple[str, ...],
    slots: tuple[str, ...],
    follow_up: Optional[bool],
    min_turns: Optional[int],
    max_turns: Optional[int],
    expected_image: Optional[str],
) -> dict:
    payload: dict = {}
    field_map = {
        "question": question,
        "explanation": explanation,
        "follow_up": follow_up,
        "min_turns": min_turns,
        "max_turns": max_turns,
        "expected_image": expected_image,
    }
    for key, value in field_map.items():
        if value is not None:
            payload[key] = value
    if scale_json is not None:
        payload["scale"] = _json_object(scale_json, "--scale-json")
    if options:
        payload["options"] = list(options)
    if slots:
        payload["slots"] = list(slots)
    return payload


def _image_data_from_file(image_file: str) -> str:
    path = Path(image_file)
    mime_type, _ = mimetypes.guess_type(path.name)
    encoded = base64.b64encode(path.read_bytes()).decode("ascii")
    return f"data:{mime_type or 'application/octet-stream'};base64,{encoded}"


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


@questions_group.command("create")
@click.argument("survey_id")
@click.option("--question", required=True, help="Question text.")
@click.option("--explanation", default=None, help="Interviewer guidance/explanation.")
@click.option("--scale-json", default=None, help="Scale config JSON object.")
@click.option("--option", "options", multiple=True, help="Fixed choice option. Repeat for multiple options.")
@click.option("--slot", "slots", multiple=True, help="Slot name. Repeat for multiple slots.")
@click.option("--follow-up/--no-follow-up", default=None, help="Whether follow-up is enabled.")
@click.option("--min-turns", type=int, default=None, help="Minimum conversation turns.")
@click.option("--max-turns", type=int, default=None, help="Maximum conversation turns.")
@click.option("--expected-image", default=None, help="Expected image description.")
@click.option("--payload-json", default=None, help="Additional request fields as a JSON object.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_create(
    ctx: click.Context,
    survey_id: str,
    question: str,
    explanation: Optional[str],
    scale_json: Optional[str],
    options: tuple[str, ...],
    slots: tuple[str, ...],
    follow_up: Optional[bool],
    min_turns: Optional[int],
    max_turns: Optional[int],
    expected_image: Optional[str],
    payload_json: Optional[str],
    output_file: Optional[str],
) -> None:
    payload = _json_object(payload_json, "--payload-json") or {}
    payload.update(
        _question_payload(
            question,
            explanation,
            scale_json,
            options,
            slots,
            follow_up,
            min_turns,
            max_turns,
            expected_image,
        )
    )
    payload["survey_id"] = survey_id

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.create_question(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Question created — ID: {result.get('id', '—')}")
    print_json(result, output_file)


@questions_group.command("get")
@click.argument("question_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_get(ctx: click.Context, question_id: str, output_file: Optional[str]) -> None:
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_question(question_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_json(result, output_file)


@questions_group.command("update")
@click.argument("question_id")
@click.option("--question", default=None, help="Question text.")
@click.option("--explanation", default=None, help="Interviewer guidance/explanation.")
@click.option("--scale-json", default=None, help="Scale config JSON object.")
@click.option("--option", "options", multiple=True, help="Fixed choice option. Repeat for multiple options.")
@click.option("--slot", "slots", multiple=True, help="Slot name. Repeat for multiple slots.")
@click.option("--follow-up/--no-follow-up", default=None, help="Whether follow-up is enabled.")
@click.option("--min-turns", type=int, default=None, help="Minimum conversation turns.")
@click.option("--max-turns", type=int, default=None, help="Maximum conversation turns.")
@click.option("--expected-image", default=None, help="Expected image description.")
@click.option("--null-field", "null_fields", multiple=True, type=click.Choice(["question", "explanation", "scale", "options", "slots", "follow_up", "min_turns", "max_turns", "expected_image"]), help="Send null for a field.")
@click.option("--payload-json", default=None, help="Request fields as a JSON object.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_update(
    ctx: click.Context,
    question_id: str,
    question: Optional[str],
    explanation: Optional[str],
    scale_json: Optional[str],
    options: tuple[str, ...],
    slots: tuple[str, ...],
    follow_up: Optional[bool],
    min_turns: Optional[int],
    max_turns: Optional[int],
    expected_image: Optional[str],
    null_fields: tuple[str, ...],
    payload_json: Optional[str],
    output_file: Optional[str],
) -> None:
    payload = _json_object(payload_json, "--payload-json") or {}
    payload.update(
        _question_payload(
            question,
            explanation,
            scale_json,
            options,
            slots,
            follow_up,
            min_turns,
            max_turns,
            expected_image,
        )
    )
    for field in null_fields:
        payload[field] = None
    if not payload:
        print_error("Provide at least one field to update.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.update_question(question_id, payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Question updated — ID: {result.get('id', question_id)}")
    print_json(result, output_file)


@questions_group.command("delete")
@click.argument("question_id")
@click.option("--survey-id", required=True, help="Survey ID the question belongs to.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_delete(ctx: click.Context, question_id: str, survey_id: str, output_file: Optional[str]) -> None:
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.delete_question(question_id, survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Question deleted — ID: {question_id}")
    print_json(result, output_file)


@questions_group.command("reorder")
@click.argument("survey_id")
@click.argument("question_ids", nargs=-1)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_reorder(ctx: click.Context, survey_id: str, question_ids: tuple[str, ...], output_file: Optional[str]) -> None:
    if not question_ids:
        print_error("Provide at least one question ID.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.reorder_questions(survey_id, list(question_ids))
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Reordered {len(question_ids)} question(s) for survey {survey_id}")
    print_json(result, output_file)


@questions_group.command("upload-image")
@click.argument("question_id")
@click.option("--image-file", type=click.Path(exists=True, dir_okay=False), default=None, help="Image file to upload.")
@click.option("--image-data", default=None, help="Base64 data URL to upload.")
@click.option("--label", default=None, help="Optional image label.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_upload_image(
    ctx: click.Context,
    question_id: str,
    image_file: Optional[str],
    image_data: Optional[str],
    label: Optional[str],
    output_file: Optional[str],
) -> None:
    if bool(image_file) == bool(image_data):
        print_error("Provide exactly one of --image-file or --image-data.")
        raise SystemExit(1)

    payload = {"image_data": image_data or _image_data_from_file(image_file or "")}
    if label is not None:
        payload["label"] = label

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.upload_question_image(question_id, payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Image uploaded — ID: {result.get('id', '—')}")
    print_json(result, output_file)


@questions_group.command("reorder-images")
@click.argument("question_id")
@click.argument("image_ids", nargs=-1)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def questions_reorder_images(
    ctx: click.Context,
    question_id: str,
    image_ids: tuple[str, ...],
    output_file: Optional[str],
) -> None:
    if not image_ids:
        print_error("Provide at least one image ID.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.reorder_question_images(question_id, list(image_ids))
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Reordered {len(image_ids)} image(s) for question {question_id}")
    print_json(result, output_file)
