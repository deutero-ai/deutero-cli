"""Main CLI entry point — assembles all command groups under `deutero`."""

from __future__ import annotations

from typing import Optional

import click

from deutero_cli import __version__
from deutero_cli.client import DeuteroClient
from deutero_cli.commands.analysis import analysis_group
from deutero_cli.commands.config_cmd import config_group
from deutero_cli.commands.interviews import interviews_group
from deutero_cli.commands.personas import personas_group
from deutero_cli.commands.questions import questions_group
from deutero_cli.commands.surveys import surveys_group


@click.group()
@click.version_option(version=__version__, prog_name="deutero")
@click.option("--api-key", envvar="DEUTERO_API_KEY", default=None, help="API key (or set DEUTERO_API_KEY).")
@click.option("--base-url", envvar="DEUTERO_BASE_URL", default=None, help="API base URL (or set DEUTERO_BASE_URL).")
@click.pass_context
def cli(ctx: click.Context, api_key: Optional[str], base_url: Optional[str]) -> None:
    """Deutero CLI — manage research surveys, interviews, and analysis."""
    ctx.ensure_object(dict)
    ctx.obj["client"] = DeuteroClient(base_url=base_url, api_key=api_key)


cli.add_command(config_group)
cli.add_command(surveys_group)
cli.add_command(questions_group)
cli.add_command(personas_group)
cli.add_command(interviews_group)
cli.add_command(analysis_group)


if __name__ == "__main__":
    cli()
