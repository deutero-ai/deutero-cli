"""Configuration commands — set API key and base URL."""

from __future__ import annotations

import click

from deutero_cli.config import get_api_key, get_base_url, save_api_key, save_base_url
from deutero_cli.output import print_key_value, print_success


@click.group("config")
def config_group() -> None:
    """View and manage CLI configuration."""


@config_group.command("show")
def config_show() -> None:
    """Display current configuration."""
    api_key = get_api_key()
    masked = f"{api_key[:8]}…{api_key[-4:]}" if api_key and len(api_key) > 12 else api_key
    print_key_value(
        {
            "API Key": masked or "(not set)",
            "Base URL": get_base_url(),
        },
        title="Deutero CLI Configuration",
    )


@config_group.command("set-key")
@click.argument("api_key")
def config_set_key(api_key: str) -> None:
    """Save an API key to ~/.deutero/config.json."""
    save_api_key(api_key)
    print_success("API key saved.")


@config_group.command("set-url")
@click.argument("base_url")
def config_set_url(base_url: str) -> None:
    """Save a custom base URL to ~/.deutero/config.json."""
    save_base_url(base_url)
    print_success(f"Base URL set to {base_url}")
