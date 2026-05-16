"""Main CLI entry point — assembles all command groups under `deutero`."""

from __future__ import annotations

from typing import Optional

import click

try:
    from trogon import tui
    import trogon.widgets.parameter_controls as _trogon_pc
    from textual.widgets import Select as _Select
    from click._utils import UNSET as _CLICK_UNSET

    @staticmethod  # type: ignore[misc]
    def _patched_apply_default_value(control_widget: object, default_value: object) -> None:
        if isinstance(control_widget, _Select):
            if default_value is _CLICK_UNSET or default_value is None:
                control_widget.clear()
            else:
                control_widget.value = str(default_value)
            control_widget.prompt = f"{default_value} (default)"
        elif hasattr(control_widget, "value"):
            control_widget.value = str(default_value)  # type: ignore[union-attr]
            if hasattr(control_widget, "placeholder"):
                control_widget.placeholder = f"{default_value} (default)"  # type: ignore[union-attr]

    _trogon_pc.ParameterControls._apply_default_value = _patched_apply_default_value
except ImportError:
    def tui(*args, **kwargs):  # type: ignore[misc]
        def decorator(f):
            return f
        return decorator


from deutero_cli import __version__
from deutero_cli.client import DeuteroClient
from deutero_cli.commands.analysis import analysis_group
from deutero_cli.commands.auth_cmd import auth_group
from deutero_cli.commands.config_cmd import config_group
from deutero_cli.commands.credits import credits_group
from deutero_cli.commands.interviews import interviews_group
from deutero_cli.commands.personas import personas_group
from deutero_cli.commands.projects import projects_group
from deutero_cli.commands.questions import questions_group
from deutero_cli.commands.simulate import simulate_group
from deutero_cli.commands.surveys import surveys_group


@tui()
@click.group()
@click.version_option(version=__version__, prog_name="deutero")
@click.option("--api-key", envvar="DEUTERO_API_KEY", default=None, help="API key (or set DEUTERO_API_KEY).")
@click.option("--base-url", envvar="DEUTERO_BASE_URL", default=None, help="API base URL (or set DEUTERO_BASE_URL).")
@click.pass_context
def cli(ctx: click.Context, api_key: Optional[str], base_url: Optional[str]) -> None:
    """Deutero CLI — manage research surveys, interviews, and analysis."""
    ctx.ensure_object(dict)
    ctx.obj["client"] = DeuteroClient(base_url=base_url, api_key=api_key)


cli.add_command(auth_group)
cli.add_command(config_group)
cli.add_command(surveys_group)
cli.add_command(questions_group)
cli.add_command(personas_group)
cli.add_command(projects_group)
cli.add_command(interviews_group)
cli.add_command(simulate_group)
cli.add_command(analysis_group)
cli.add_command(credits_group)


if __name__ == "__main__":
    cli()
