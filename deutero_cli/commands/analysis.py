"""Thematic analysis commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_survey_id
from deutero_cli.output import print_error, print_json, print_key_value, print_success, print_xml


@click.group("analysis")
def analysis_group() -> None:
    """Run and inspect thematic analysis."""


@analysis_group.command("run")
@click.option("--interview-id", default=None, help="Run analysis on a single interview.")
@click.option("--survey-id", default=None, help="Run analysis on all completed interviews in a survey.")
@click.option("--model-tier", "-m", type=click.Choice(["open_weights", "premium", "frontier"]), default="open_weights", help="Model tier to use.")
@click.option("--cross-case/--no-cross-case", "cross_case_analysis", default=False, help="Run cross-case analysis (requires --survey-id).")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.option("--xml-output", default=None, help="Write cross-case XML to a file (cross-case only).")
@click.pass_context
def analysis_run(
    ctx: click.Context,
    interview_id: Optional[str],
    survey_id: Optional[str],
    model_tier: str,
    cross_case_analysis: bool,
    output_file: Optional[str],
    xml_output: Optional[str],
) -> None:
    """Run thematic analysis on interviews."""
    if not interview_id and not survey_id:
        print_error("Provide --interview-id or --survey-id.")
        raise SystemExit(1)

    payload: dict = {"model_tier": model_tier, "cross_case_analysis": cross_case_analysis}
    if interview_id:
        payload["interview_id"] = interview_id
    if survey_id:
        payload["survey_id"] = survey_id

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.run_analysis(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    if xml_output and result.get("cross_case_xml"):
        print_xml(result["cross_case_xml"], xml_output)

    print_success(result.get("message", "Analysis submitted."))
    print_json(result, output_file)


@analysis_group.command("status")
@click.option("--interview-id", default=None, help="Get status for a single interview.")
@click.option("--survey-id", default=None, help="Get status for all interviews in a survey.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def analysis_status(
    ctx: click.Context,
    interview_id: Optional[str],
    survey_id: Optional[str],
    output_file: Optional[str],
) -> None:
    """Get analysis status for interviews."""
    if not interview_id and not survey_id:
        print_error("Provide --interview-id or --survey-id.")
        raise SystemExit(1)

    params: dict = {}
    if interview_id:
        params["interview_id"] = interview_id
    if survey_id:
        params["survey_id"] = survey_id

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_analysis_status(params)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_json(result, output_file)


@analysis_group.command("results-interview")
@click.argument("interview_id", required=False, default=None)
@click.option("--phase", "-p", type=click.Choice(["initial_engagement", "initial_noting", "emergent_themes", "connections"]), required=True, help="Analysis phase to retrieve.")
@click.option("--output", "-o", "output_file", default=None, help="Write XML result to a file.")
@click.pass_context
def analysis_results_interview(
    ctx: click.Context,
    interview_id: Optional[str],
    phase: str,
    output_file: Optional[str],
) -> None:
    """Get analysis results for a specific interview phase."""
    if interview_id is None:
        interview_id = click.prompt("Interview ID")
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_interview_analysis_results(interview_id, phase)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    xml_output = result.get("xml_output")
    if xml_output:
        print_xml(xml_output, output_file)
    elif result.get("error"):
        print_error(result["error"])
    else:
        print_json(result, output_file)


@analysis_group.command("results-survey")
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write XML result to a file.")
@click.pass_context
def analysis_results_survey(
    ctx: click.Context,
    survey_id: Optional[str],
    output_file: Optional[str],
) -> None:
    """Get cross-case analysis results for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_survey_analysis_results(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    xml_output = result.get("xml_output")
    if xml_output:
        print_xml(xml_output, output_file)
    elif result.get("error"):
        print_error(result["error"])
    else:
        print_json(result, output_file)
