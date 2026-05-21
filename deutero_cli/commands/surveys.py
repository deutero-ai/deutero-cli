"""Survey commands — generate surveys, view participation, get agent requirements."""

from __future__ import annotations

from pathlib import Path
from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_project_id, get_active_survey_id, save_active_survey_id
from deutero_cli.output import print_error, print_items_table, print_json, print_key_value, print_success, print_xml, prompt_select_from_list


@click.group("surveys")
def surveys_group() -> None:
    """Manage research surveys."""


@surveys_group.command("list")
@click.argument("project_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_list(ctx: click.Context, project_id: Optional[str], output_file: Optional[str]) -> None:
    """List surveys for a project.

    PROJECT_ID defaults to the active project (set via `deutero projects set-active`).
    """
    if project_id is None:
        project_id = get_active_project_id()
    if project_id is None:
        project_id = click.prompt("Project ID")

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_project_surveys(project_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    surveys = [{**s, "name": s.get("alias")} for s in result.get("surveys", [])]
    print_success(f"Found {len(surveys)} survey(s) for project {project_id}")
    print_items_table(surveys, title="Surveys", columns=["id", "name", "description"])
    if output_file:
        print_json(result, output_file)


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
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_participation(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
    """Get participation statistics for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
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
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write the markdown to a file.")
@click.pass_context
def surveys_agent_requirements(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
    """Retrieve or generate agent requirements markdown for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
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
@click.argument("survey_id", required=False, default=None)
@click.option("--set", "model_tier", type=click.Choice(["open_weights", "premium", "frontier"]), default=None, help="Set the survey model tier.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_model_tier(ctx: click.Context, survey_id: Optional[str], model_tier: Optional[str], output_file: Optional[str]) -> None:
    """Get or set the model tier for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
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
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_question_list(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
    """Get the full question list for a survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    if survey_id is None:
        survey_id = get_active_survey_id()
    if survey_id is None:
        survey_id = click.prompt("Survey ID")
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.get_question_list(survey_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    questions = result.get("questions", [])
    print_success(f"Found {len(questions)} question(s) for survey {survey_id}")
    print_items_table(questions, title="Questions", columns=["id", "question", "explanation"])
    if output_file:
        print_json(result, output_file)


@surveys_group.command("extract-from-guide")
@click.argument("guide_file", type=click.Path(exists=True, readable=True))
@click.option(
    "--extract",
    type=click.Choice(["details", "questions", "both"]),
    default="both",
    show_default=True,
    help="What to extract: 'details', 'questions', or 'both'.",
)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_extract_from_guide(ctx: click.Context, guide_file: str, extract: str, output_file: Optional[str]) -> None:
    """Extract study details and/or questions from a research guide file.

    GUIDE_FILE is the path to a text file containing the research guide content.
    """
    text = Path(guide_file).read_text(encoding="utf-8")
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.extract_from_guide({"text": text, "extract": extract})
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    if extract in ("details", "both"):
        details = result.get("details") or {}
        if details:
            print_key_value(
                {
                    "Name": details.get("name", ""),
                    "Description": details.get("description", ""),
                    "Research questions": details.get("research_questions", ""),
                    "Objectives": details.get("objectives", ""),
                    "Target population": details.get("target_population", ""),
                },
                title="Extracted Study Details",
            )

    if extract in ("questions", "both"):
        questions = result.get("questions") or []
        print_success(f"Extracted {len(questions)} question(s)")
        print_items_table(questions, title="Extracted Questions", columns=["question", "type", "interviewer_guidance"])

    if output_file:
        print_json(result, output_file)


@surveys_group.command("interviews")
@click.argument("survey_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_interviews(ctx: click.Context, survey_id: Optional[str], output_file: Optional[str]) -> None:
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
    print_success(f"Found {len(interviews)} interview(s) for survey {survey_id}")
    print_items_table(
        interviews,
        title="Interviews",
        columns=["id", "participant_name", "start_time", "end_time", "completed", "simulated", "termination_reason"],
    )
    if output_file:
        print_json(result, output_file)


@surveys_group.command("create")
@click.option("--project-id", required=True, help="Project ID to create the survey in.")
@click.option("--survey-type", "-t", type=click.Choice(["sociology", "user_experience", "customer_development", "polling"]), default=None, help="Type of survey.")
@click.option("--name", default=None, help="Display name for the survey.")
@click.option("--description", default=None, help="Description of the survey.")
@click.option("--research-questions", default=None, help="Research questions for the survey.")
@click.option("--objectives", default=None, help="Research objectives.")
@click.option("--target-population", default=None, help="Target population for the survey.")
@click.option("--anonymous", "set_anonymous", is_flag=True, default=False, help="Mark the survey as anonymous.")
@click.option("--no-anonymous", "set_not_anonymous", is_flag=True, default=False, help="Mark the survey as non-anonymous.")
@click.option("--model-tier", type=click.Choice(["open_weights", "premium", "frontier"]), default=None, help="Model tier.")
@click.option("--redirect-url", default=None, help="URL to redirect participants after completing the survey.")
@click.option("--language", "-l", default=None, help="Language for the survey.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_create(
    ctx: click.Context,
    project_id: str,
    survey_type: Optional[str],
    name: Optional[str],
    description: Optional[str],
    research_questions: Optional[str],
    objectives: Optional[str],
    target_population: Optional[str],
    set_anonymous: bool,
    set_not_anonymous: bool,
    model_tier: Optional[str],
    redirect_url: Optional[str],
    language: Optional[str],
    output_file: Optional[str],
) -> None:
    """Create a new survey directly."""
    payload: dict = {"project_id": project_id}
    if survey_type is not None:
        payload["survey_type"] = survey_type
    if name is not None:
        payload["name"] = name
    if description is not None:
        payload["description"] = description
    if research_questions is not None:
        payload["research_questions"] = research_questions
    if objectives is not None:
        payload["objectives"] = objectives
    if target_population is not None:
        payload["target_population"] = target_population
    if set_anonymous or set_not_anonymous:
        payload["anonymous"] = True if set_anonymous else False
    if model_tier is not None:
        payload["model_tier"] = model_tier
    if redirect_url is not None:
        payload["redirect_url"] = redirect_url
    if language is not None:
        payload["language"] = language

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.create_survey(payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Survey created — ID: {result.get('study_id')}")
    print_key_value(
        {
            "Study ID": result.get("study_id"),
            "Project ID": result.get("project_id"),
            "Name": result.get("name"),
            "Survey type": result.get("survey_type"),
            "Model tier": result.get("model_tier"),
            "Language": result.get("language"),
            "Date created": result.get("date_created"),
            "URL": result.get("url"),
        },
        title="Survey",
    )
    if output_file:
        print_json(result, output_file)


@surveys_group.command("update")
@click.argument("survey_id")
@click.option("--survey-type", "-t", type=click.Choice(["sociology", "user_experience", "customer_development", "polling"]), default=None, help="Type of survey.")
@click.option("--name", default=None, help="Display name for the survey.")
@click.option("--description", default=None, help="Description of the survey.")
@click.option("--research-questions", default=None, help="Research questions for the survey.")
@click.option("--objectives", default=None, help="Research objectives.")
@click.option("--target-population", default=None, help="Target population for the survey.")
@click.option("--anonymous", "set_anonymous", is_flag=True, default=False, help="Mark the survey as anonymous.")
@click.option("--no-anonymous", "set_not_anonymous", is_flag=True, default=False, help="Mark the survey as non-anonymous.")
@click.option("--model-tier", type=click.Choice(["open_weights", "premium", "frontier"]), default=None, help="Model tier.")
@click.option("--redirect-url", default=None, help="URL to redirect participants after completing the survey.")
@click.option("--language", "-l", default=None, help="Language for the survey.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def surveys_update(
    ctx: click.Context,
    survey_id: str,
    survey_type: Optional[str],
    name: Optional[str],
    description: Optional[str],
    research_questions: Optional[str],
    objectives: Optional[str],
    target_population: Optional[str],
    set_anonymous: bool,
    set_not_anonymous: bool,
    model_tier: Optional[str],
    redirect_url: Optional[str],
    language: Optional[str],
    output_file: Optional[str],
) -> None:
    """Update properties of an existing survey.

    SURVEY_ID defaults to the active survey (set via `deutero surveys set-active`).
    """
    payload: dict = {}
    if survey_type is not None:
        payload["survey_type"] = survey_type
    if name is not None:
        payload["name"] = name
    if description is not None:
        payload["description"] = description
    if research_questions is not None:
        payload["research_questions"] = research_questions
    if objectives is not None:
        payload["objectives"] = objectives
    if target_population is not None:
        payload["target_population"] = target_population
    if set_anonymous or set_not_anonymous:
        payload["anonymous"] = True if set_anonymous else False
    if model_tier is not None:
        payload["model_tier"] = model_tier
    if redirect_url is not None:
        payload["redirect_url"] = redirect_url
    if language is not None:
        payload["language"] = language

    if not payload:
        print_error("No fields to update. Provide at least one option.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.update_survey(survey_id, payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Survey updated — ID: {result.get('study_id')}")
    print_key_value(
        {
            "Study ID": result.get("study_id"),
            "Project ID": result.get("project_id"),
            "Name": result.get("name"),
            "Survey type": result.get("survey_type"),
            "Model tier": result.get("model_tier"),
            "Language": result.get("language"),
            "Date created": result.get("date_created"),
            "URL": result.get("url"),
        },
        title="Survey",
    )
    if output_file:
        print_json(result, output_file)


@surveys_group.command("set-active")
@click.option("--project-id", default=None, help="Project ID to list surveys from (defaults to active project).")
@click.pass_context
def surveys_set_active(ctx: click.Context, project_id: str) -> None:
    """Set the active survey (interactive selection)."""
    if project_id is None:
        project_id = get_active_project_id()
    if project_id is None:
        project_id = click.prompt("Project ID")

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_project_surveys(project_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    surveys = result.get("surveys", [])
    if not surveys:
        print_error(f"No surveys found for project {project_id}.")
        raise SystemExit(1)

    chosen = prompt_select_from_list(
        surveys,
        lambda s: f"[bold]{s.get('alias', '\u2014')}[/bold]  [dim]{s.get('id')}[/dim]",
        "Select survey",
    )
    save_active_survey_id(chosen["id"])
    print_success(f"Active survey set to: {chosen.get('alias', chosen['id'])} ({chosen['id']})")

