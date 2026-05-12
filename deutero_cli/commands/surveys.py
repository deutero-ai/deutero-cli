"""Survey commands — generate surveys, view participation, get agent requirements."""

from __future__ import annotations

from pathlib import Path
from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_key_value, print_success, print_xml


@click.group("surveys")
def surveys_group() -> None:
    """Manage research surveys."""


@surveys_group.command("generate")
@click.option("--survey-type", "-t", type=click.Choice(["user_experience", "sociology", "customer_development", "polling"]), default="user_experience", help="Type of survey to generate.")
@click.option("--language", "-l", default="English", help="Language for the survey.")
@click.option("--business-context", help="Business/product context (user_experience).")
@click.option("--research-need", help="Why the research is needed (user_experience).")
@click.option("--target-users", help="Primary audience (user_experience).")
@click.option("--constraints", help="Key constraints (user_experience).")
@click.option("--research-question", help="Research question (sociology/polling).")
@click.option("--population-of-interest", help="Population of interest (sociology).")
@click.option("--context-or-setting", help="Context or setting (sociology).")
@click.option("--key-concepts", help="Key concepts (sociology).")
@click.option("--scope-and-boundaries", help="Scope and boundaries (sociology).")
@click.option("--problem-hypothesis", help="Problem hypothesis (customer_development).")
@click.option("--customer-segment", help="Customer segment (customer_development).")
@click.option("--solution-concept", help="Solution concept (customer_development).")
@click.option("--key-assumptions", help="Key assumptions (customer_development).")
@click.option("--success-criteria", help="Success criteria (customer_development).")
@click.option("--population-segment", help="Population segment (polling).")
@click.option("--geographic-scope", help="Geographic scope (polling).")
@click.option("--survey-context", help="Survey context (polling).")
@click.option("--data-quality-requirements", help="Data quality requirements (polling).")
@click.option("--model-tier", type=click.Choice(["open_weights", "premium", "frontier"]), default=None, help="Model tier to set on the created survey.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.option("--xml-output", default=None, help="Write the XML specification to a file.")
@click.pass_context
def surveys_generate(
    ctx: click.Context,
    survey_type: str,
    language: str,
    business_context: Optional[str],
    research_need: Optional[str],
    target_users: Optional[str],
    constraints: Optional[str],
    research_question: Optional[str],
    population_of_interest: Optional[str],
    context_or_setting: Optional[str],
    key_concepts: Optional[str],
    scope_and_boundaries: Optional[str],
    problem_hypothesis: Optional[str],
    customer_segment: Optional[str],
    solution_concept: Optional[str],
    key_assumptions: Optional[str],
    success_criteria: Optional[str],
    population_segment: Optional[str],
    geographic_scope: Optional[str],
    survey_context: Optional[str],
    data_quality_requirements: Optional[str],
    model_tier: Optional[str],
    output_file: Optional[str],
    xml_output: Optional[str],
) -> None:
    """Generate a new research survey specification."""
    payload: dict = {"survey_type": survey_type, "language": language}

    field_map = {
        "business_context": business_context,
        "research_need": research_need,
        "target_users": target_users,
        "constraints": constraints,
        "research_question": research_question,
        "population_of_interest": population_of_interest,
        "context_or_setting": context_or_setting,
        "key_concepts": key_concepts,
        "scope_and_boundaries": scope_and_boundaries,
        "problem_hypothesis": problem_hypothesis,
        "customer_segment": customer_segment,
        "solution_concept": solution_concept,
        "key_assumptions": key_assumptions,
        "success_criteria": success_criteria,
        "population_segment": population_segment,
        "geographic_scope": geographic_scope,
        "survey_context": survey_context,
        "data_quality_requirements": data_quality_requirements,
        "model_tier": model_tier,
    }
    for key, value in field_map.items():
        if value is not None:
            payload[key] = value

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.generate_survey(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    if xml_output and result.get("xml_file"):
        print_xml(result["xml_file"], xml_output)

    print_success(f"Survey created — ID: {result.get('study_id')}")
    print_json(result, output_file)


@surveys_group.command("participation")
@click.argument("survey_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_participation(ctx: click.Context, survey_id: str, output_file: Optional[str]) -> None:
    """Get participation statistics for a survey."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_participation(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_key_value(
        {
            "Survey ID": result.get("survey_id"),
            "Total interviews": result.get("total_interviews"),
            "Completed": result.get("completed_interviews"),
            "Incomplete": result.get("incomplete_interviews"),
            "Completion rate": f"{result.get('completion_rate', 0)}%",
            "Quota fill rate": f"{result.get('quota_fill_rate', '—')}%"
            if result.get("quota_fill_rate") is not None
            else "—",
            "Quota remaining": result.get("quota_remaining", "—"),
        },
        title="Survey Participation",
    )
    if output_file:
        print_json(result, output_file)


@surveys_group.command("agent-requirements")
@click.argument("survey_id")
@click.option("--output", "-o", "output_file", default=None, help="Write the markdown to a file.")
@click.pass_context
def surveys_agent_requirements(ctx: click.Context, survey_id: str, output_file: Optional[str]) -> None:
    """Retrieve or generate agent requirements markdown for a survey."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_agent_requirements(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    md = result.get("markdown", "")
    filename = result.get("filename", "requirements.md")
    target = output_file or filename

    Path(target).write_text(md, encoding="utf-8")
    print_success(f"Agent requirements written to {target}")


@surveys_group.command("model-tier")
@click.argument("survey_id")
@click.option("--set", "model_tier", type=click.Choice(["open_weights", "premium", "frontier"]), default=None, help="Set the survey model tier.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_model_tier(ctx: click.Context, survey_id: str, model_tier: Optional[str], output_file: Optional[str]) -> None:
    """Get or set the model tier for a survey."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.set_survey_model_tier(survey_id, model_tier) if model_tier else client.get_survey_model_tier(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_key_value(
        {
            "Survey ID": result.get("survey_id"),
            "Model tier": result.get("model_tier"),
            "Model ID": result.get("model_id"),
            "Provider": result.get("model_provider"),
        },
        title="Survey Model Tier",
    )
    if output_file:
        print_json(result, output_file)


@surveys_group.command("suggest-from-site")
@click.argument("url")
@click.option("--language", "-l", default="English", help="Language for the generated study suggestion.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_suggest_from_site(ctx: click.Context, url: str, language: str, output_file: Optional[str]) -> None:
    """Generate a study suggestion from a website URL."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.suggest_from_site({"url": url, "language": language})
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    status = "valid" if result.get("is_valid") else "not valid"
    print_success(f"Site suggestion completed — {status}")
    print_json(result, output_file)


@surveys_group.command("question-list")
@click.argument("survey_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_question_list(ctx: click.Context, survey_id: str, output_file: Optional[str]) -> None:
    """Get the full question list for a survey."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_question_list(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    questions = result.get("questions", [])
    print_success(f"Found {len(questions)} question(s) for survey {survey_id}")
    print_json(result, output_file)


@surveys_group.command("interviews")
@click.argument("survey_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_interviews(ctx: click.Context, survey_id: str, output_file: Optional[str]) -> None:
    """List interviews for a survey."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_survey_interviews(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    interviews = result.get("interviews", [])
    print_success(f"Found {len(interviews)} interview(s) for survey {survey_id}")
    print_json(result, output_file)
