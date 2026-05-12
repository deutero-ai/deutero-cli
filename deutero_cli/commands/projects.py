"""Project listing commands."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_json, print_key_value, print_success


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
    print_json(result, output_file)


@projects_group.command("list-surveys")
@click.argument("project_id")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def projects_list_surveys(ctx: click.Context, project_id: str, output_file: Optional[str]) -> None:
    """List surveys for a project."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_project_surveys(project_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    surveys = result.get("surveys", [])
    print_success(f"Found {len(surveys)} survey(s) for project {project_id}")
    print_json(result, output_file)
