"""Webhook commands — list, update, and delete webhook endpoints."""

from __future__ import annotations

from typing import Any, Dict, Optional, Tuple

import click

from deutero_cli.client import DeuteroClient
from deutero_cli.output import print_error, print_items_table, print_json, print_key_value, print_success


@click.group("webhooks")
def webhooks_group() -> None:
    """Manage webhook endpoints."""


@webhooks_group.command("list")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def webhooks_list(ctx: click.Context, output_file: Optional[str]) -> None:
    """List all webhook endpoints for the authenticated organization."""
    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.list_webhook_endpoints()
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    endpoints = result.get("endpoints", [])
    total = result.get("total", len(endpoints))
    event_types = result.get("event_types", [])
    print_success(f"Found {total} webhook endpoint(s)")
    print_items_table(endpoints, title="Webhook Endpoints", columns=["id", "label", "url", "enabled", "events"])
    if event_types:
        click.echo(f"\nAvailable event types: {', '.join(event_types)}")
    if output_file:
        print_json(result, output_file)


@webhooks_group.command("update")
@click.argument("endpoint_id")
@click.option("--label", default=None, help="New label for the webhook endpoint.")
@click.option("--url", default=None, help="New delivery URL for the webhook endpoint.")
@click.option("--enable", "set_enabled", is_flag=True, default=False, help="Enable the webhook endpoint.")
@click.option("--disable", "set_disabled", is_flag=True, default=False, help="Disable the webhook endpoint.")
@click.option("--event", "events", multiple=True, help="Subscribed event type (repeatable). Replaces the existing event list.")
@click.option("--output", "-o", "output_file", default=None, help="Write JSON response to a file.")
@click.pass_context
def webhooks_update(
    ctx: click.Context,
    endpoint_id: str,
    label: Optional[str],
    url: Optional[str],
    set_enabled: bool,
    set_disabled: bool,
    events: Tuple[str, ...],
    output_file: Optional[str],
) -> None:
    """Update a webhook endpoint's label, URL, enabled state, or subscribed events.

    ENDPOINT_ID is the UUID of the webhook endpoint to update.
    """
    enabled: Optional[bool] = True if set_enabled else (False if set_disabled else None)

    payload: Dict[str, Any] = {}
    if label is not None:
        payload["label"] = label
    if url is not None:
        payload["url"] = url
    if enabled is not None:
        payload["enabled"] = enabled
    if events:
        payload["events"] = list(events)

    if not payload:
        print_error("No update fields provided. Use --label, --url, --enable/--disable, or --event.")
        raise SystemExit(1)

    client: DeuteroClient = ctx.obj["client"]
    try:
        result = client.update_webhook_endpoint(endpoint_id, payload)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Webhook endpoint {endpoint_id} updated.")
    print_key_value(
        {
            "ID": result.get("id"),
            "Label": result.get("label"),
            "URL": result.get("url"),
            "Enabled": result.get("enabled"),
            "Events": ", ".join(result.get("events") or []) or "—",
            "Updated at": result.get("updated_at"),
        },
        title="Webhook Endpoint",
    )
    if output_file:
        print_json(result, output_file)


@webhooks_group.command("delete")
@click.argument("endpoint_id")
@click.option("--yes", "-y", is_flag=True, default=False, help="Skip confirmation prompt.")
@click.pass_context
def webhooks_delete(ctx: click.Context, endpoint_id: str, yes: bool) -> None:
    """Delete a webhook endpoint permanently.

    ENDPOINT_ID is the UUID of the webhook endpoint to delete.
    """
    if not yes:
        click.confirm(f"Delete webhook endpoint {endpoint_id}? This cannot be undone.", abort=True)

    client: DeuteroClient = ctx.obj["client"]
    try:
        client.delete_webhook_endpoint(endpoint_id)
    except Exception as exc:
        print_error(str(exc))
        raise SystemExit(1)

    print_success(f"Webhook endpoint {endpoint_id} deleted.")
