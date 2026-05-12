"""Authentication commands — login, logout, and status via Stytch Connected Apps."""

from __future__ import annotations

import click

from deutero_cli.auth import clear_tokens, get_auth_status, login
from deutero_cli.config import _load_config_file
from deutero_cli.output import console, print_error, print_key_value, print_success


@click.group("auth")
def auth_group() -> None:
    """Authenticate with Deutero via Stytch Connected Apps."""


@auth_group.command("login")
@click.option("--client-id", envvar="DEUTERO_CLIENT_ID", required=True, help="Stytch Connected App client ID (or set DEUTERO_CLIENT_ID).")
@click.option("--project-id", envvar="DEUTERO_STYTCH_PROJECT_ID", required=True, help="Stytch project ID (or set DEUTERO_STYTCH_PROJECT_ID).")
def auth_login(client_id: str, project_id: str) -> None:
    """Authenticate via browser-based OAuth2 + PKCE flow."""
    console.print("[bold]Opening browser for authentication...[/bold]")
    try:
        token_data = login(client_id=client_id, project_id=project_id)
        print_success("Authentication successful!")
        console.print(f"  Token expires in: [cyan]{token_data.get('expires_in', '?')}[/cyan] seconds")
        if token_data.get("refresh_token"):
            console.print("  Refresh token: [green]saved[/green]")
    except RuntimeError as exc:
        print_error(str(exc))
        raise SystemExit(1)


@auth_group.command("logout")
def auth_logout() -> None:
    """Clear stored authentication tokens."""
    clear_tokens()
    print_success("Logged out. Stored tokens removed.")


@auth_group.command("status")
def auth_status() -> None:
    """Show current authentication status."""
    status = get_auth_status()
    if not status["logged_in"]:
        console.print("[yellow]Not logged in.[/yellow] Run [bold]deutero auth login[/bold] to authenticate.")
        return

    info = {
        "Logged in": "Yes",
        "Token expired": "Yes" if status["expired"] else "No",
        "Refresh token": "Available" if status["has_refresh_token"] else "Not available",
        "Client ID": status["client_id"] or "(not set)",
        "Project ID": status["project_id"] or "(not set)",
    }
    print_key_value(info, title="Auth Status")
