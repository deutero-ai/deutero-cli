"""Project listing commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.config import get_active_project_id, save_active_project_id
from deutero_cli.output import print_error, print_items_table, print_json, print_success, prompt_select_from_list


@click.group("projects")
def projects_group() -> None:
    """List projects and their surveys."""


@projects_group.command("list")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def projects_list(ctx: click.Context, output_file: Optional[str]) -> None:
    """List all projects."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_projects()
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    projects = result.get("projects", [])
    print_success(f"Found {len(projects)} project(s)")
    print_items_table(projects, title="Projects", columns=["id", "name", "description"])
    if output_file:
        print_json(result, output_file)


@projects_group.command("list-surveys")
@click.argument("project_id", required=False, default=None)
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def projects_list_surveys(ctx: click.Context, project_id: Optional[str], output_file: Optional[str]) -> None:
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


@projects_group.command("set-active")
@click.pass_context
def projects_set_active(ctx: click.Context) -> None:
    """Set the active project (interactive selection)."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_projects()
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    projects = result.get("projects", [])
    if not projects:
        print_error("No projects found.")
        raise SystemExit(1)

    chosen = prompt_select_from_list(
        projects,
        lambda p: f"[bold]{p.get('name', '\u2014')}[/bold]  [dim]{p.get('id')}[/dim]",
        "Select project",
    )
    save_active_project_id(chosen["id"])
    print_success(f"Active project set to: {chosen.get('name', chosen['id'])} ({chosen['id']})")
