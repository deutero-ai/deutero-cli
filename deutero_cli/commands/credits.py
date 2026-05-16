"""Credit balance and estimation commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_survey_id
from deutero_cli.output import print_error, print_json, print_key_value


MODEL_TIER = click.Choice(["open_weights", "premium", "frontier"])


@click.group("credits")
def credits_group() -> None:
    """Manage credit balance and estimates."""


def _estimate_payload(
    survey_id: str,
    model_tier: str,
    count: int,
    include_analysis: Optional[bool] = None,
) -> dict:
    payload: dict = {
        "survey_id": survey_id,
        "model_tier": model_tier,
        "num_participants": count,
    }
    if include_analysis is not None:
        payload["include_analysis"] = include_analysis
    return payload


def _print_estimate(result: dict, output_file: Optional[str]) -> None:
    print_key_value(
        {
            "Estimated credits": result.get("estimated_credits"),
            "Credits per interview": result.get("credits_per_interview"),
            "Credits for analysis": result.get("credits_for_analysis"),
            "Model tier": result.get("model_tier"),
            "Participants/interviews": result.get("num_participants"),
            "Questions": result.get("num_questions"),
        },
        title="Credit Estimate",
    )
    if output_file:
        print_json(result, output_file)


@credits_group.command("balance")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def credits_balance(ctx: click.Context, output_file: Optional[str]) -> None:
    """Get current credit balance and pending reservations."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_credit_balance()
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_key_value(
        {
            "Available credits": result.get("available_credits"),
            "Net available": result.get("net_available"),
            "Credits used": result.get("credits_used"),
            "Credits reserved": result.get("credits_reserved"),
            "Base limit": result.get("base_limit"),
            "Purchased credits": result.get("purchased_credits"),
            "Rollover credits": result.get("rollover_credits"),
            "Trial active": result.get("is_trial_active"),
            "Trial credits used": result.get("trial_credits_used"),
        },
        title="Credit Balance",
    )
    if output_file:
        print_json(result, output_file)


@credits_group.command("estimate-simulation")
@click.argument("survey_id", required=False, default=None)
@click.option("--model-tier", "-m", type=MODEL_TIER, default="open_weights", help="Model tier to use.")
@click.option("--participants", "-n", type=int, default=1, show_default=True, help="Number of participants to simulate.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def credits_estimate_simulation(
    ctx: click.Context,
    survey_id: Optional[str],
    model_tier: str,
    participants: int,
    output_file: Optional[str],
) -> None:
    """Estimate credits for a simulation run.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
    if participants < 1:
        print_error("Participants must be greater than zero.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.estimate_simulation_credits(_estimate_payload(survey_id, model_tier, participants))
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    _print_estimate(result, output_file)


@credits_group.command("estimate-analysis")
@click.argument("survey_id", required=False, default=None)
@click.option("--model-tier", "-m", type=MODEL_TIER, default="open_weights", help="Model tier to use.")
@click.option("--interviews", "-n", type=int, default=1, show_default=True, help="Number of interviews to analyze.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def credits_estimate_analysis(
    ctx: click.Context,
    survey_id: Optional[str],
    model_tier: str,
    interviews: int,
    output_file: Optional[str],
) -> None:
    """Estimate credits for thematic analysis.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
    if interviews < 1:
        print_error("Interviews must be greater than zero.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.estimate_analysis_credits(_estimate_payload(survey_id, model_tier, interviews))
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    _print_estimate(result, output_file)


@credits_group.command("estimate-survey")
@click.argument("survey_id", required=False, default=None)
@click.option("--model-tier", "-m", type=MODEL_TIER, default="open_weights", help="Model tier to use.")
@click.option("--participants", "-n", type=int, default=1, show_default=True, help="Number of participants.")
@click.option("--include-analysis/--no-include-analysis", default=False, help="Include analysis credits in the estimate.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def credits_estimate_survey(
    ctx: click.Context,
    survey_id: Optional[str],
    model_tier: str,
    participants: int,
    include_analysis: bool,
    output_file: Optional[str],
) -> None:
    """Estimate credits for a full survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
    if participants < 1:
        print_error("Participants must be greater than zero.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.estimate_survey_credits(
            _estimate_payload(survey_id, model_tier, participants, include_analysis)
        )
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    _print_estimate(result, output_file)
