"""Interview management commands — list, transcript, and analysis."""

from __future__ import annotations

from datetime import datetime
from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.commands.simulate import simulate_run
from deutero_cli.config import get_active_survey_id
from deutero_cli.output import (
    print_error,
    print_items_table,
    print_json,
    print_key_value,
    print_success,
    print_transcript,
    prompt_select_from_list,
)


@click.group("interviews")
def interviews_group() -> None:
    """Manage and inspect interviews."""


interviews_group.add_command(simulate_run, "simulate")


@interviews_group.command("list")
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def interviews_list(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
    """List interviews for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).

    Use -o/--output to save the full JSON response to a file.
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_survey_interviews(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    interviews = result.get("interviews", [])
    print_json(result, output_file)
    print_success(f"Found {len(interviews)} interview(s) for survey {survey_id}")
    print_items_table(
        interviews,
        title="Interviews",
        columns=["id", "participant_name", "start_time", "completed", "simulated", "message_count"],
    )
    if output_file:
        print_json(result, output_file)


@interviews_group.command("transcript")
@click.argument("interview_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def interviews_transcript(ctx: click.Context, interview_id: Optional[str], output_file: Optional[str]) -> None:
    """Get the transcript for an interview.

    If INTERVIEW_ID is omitted, a list of interviews for the active survey is shown
    and you will be prompted to select one.

    Use -o/--output to save the full JSON response to a file.
    """
    client: DeuteroClient = ctx.obj["client"]

    if interview_id is None:
        survey_id = get_active_survey_id()
        if survey_id is None:
            survey_id = click.prompt("Survey ID")
        try:
            list_result = client.list_survey_interviews(survey_id)
        except Exception as exc:
            print_error(str(exc))
            raise SystemExit(1)
        interviews = list_result.get("interviews", [])
        if not interviews:
            print_error("No interviews found for this survey.")
            raise SystemExit(1)

        def _display(iv: dict) -> str:
            name = iv.get("participant_name") or "Anon"
            raw_start = iv.get("start_time")
            try:
                start = datetime.fromisoformat(raw_start).strftime("%d/%m %H:%M") if raw_start else "—"
            except ValueError:
                start = raw_start or "—"
            completed = iv.get("completed")
            completed_str = "Completed" if completed is True else ("Error" if completed is False else "Pending")
            msg_count = iv.get("message_count")
            count_str = f"  {msg_count} msgs" if msg_count is not None else ""
            return f"[bold]{name}[/bold]  [dim]{start}[/dim]  {completed_str}{count_str}  [dim]{iv.get('id')}[/dim]"

        chosen = prompt_select_from_list(interviews, _display, "Select interview")
        interview_id = chosen["id"]
    try:
        result = client.get_interview_transcript(interview_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    messages = result.get("messages", [])
    print_success(f"Transcript for interview {interview_id} — {len(messages)} message(s)")
    print_transcript(messages, interview_id)
    if output_file:
        print_json(result, output_file)


@interviews_group.command("analysis")
@click.argument("interview_id", required=False, default=None)
@click.option(
    "--model-tier",
    "-m",
    type=click.Choice(["open_weights", "premium", "frontier"]),
    default="open_weights",
    help="Model tier to use.",
)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def interviews_analysis(
    ctx: click.Context,
    interview_id: Optional[str],
    model_tier: str,
    output_file: Optional[str],
) -> None:
    """Run thematic analysis for a specific interview.

    Use -m/--model-tier to select the analysis model (default: open_weights).
    Use -o/--output to save the full JSON response to a file.
    """
    if interview_id is None:
        interview_id = click.prompt("Interview ID")

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.run_analysis(
            {
                "interview_id": interview_id,
                "model_tier": model_tier,
                "cross_case_analysis": False,
            }
        )
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(result.get("message", "Analysis submitted."))
    print_key_value(
        {
            "Interview ID": interview_id,
            "Model tier": model_tier,
            "Status": result.get("status"),
            "Message": result.get("message"),
        },
        title="Analysis",
    )
    if output_file:
        print_json(result, output_file)
