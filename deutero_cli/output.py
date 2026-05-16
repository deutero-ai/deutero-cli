"""Output formatting helpers using Rich."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Dict, List, Optional

import click
from rich.console import Console
from rich.panel import Panel
from rich.syntax import Syntax
from rich.table import Table

console = Console()
err_console = Console(stderr=True)


def print_json(data: Any, output_file: Optional[str] = None) -> None:
    """Pretty-print JSON to stdout or write to a file."""
    formatted = json.dumps(data, indent=2, default=str)
    if output_file:
        Path(output_file).write_text(formatted, encoding="utf-8")
        err_console.print(f"[green]Output written to {output_file}[/green]")
    else:
        syntax = Syntax(formatted, "json", theme="monokai", line_numbers=False)
        console.print(syntax)


def print_error(message: str) -> None:
    err_console.print(f"[bold red]Error:[/bold red] {message}")


def print_success(message: str) -> None:
    console.print(f"[bold green]✓[/bold green] {message}")


def print_key_value(data: Dict[str, Any], title: Optional[str] = None) -> None:
    """Print a dict as a Rich table of key-value pairs."""
    table = Table(show_header=False, title=title, title_style="bold cyan", expand=True)
    table.add_column("Key", style="bold")
    table.add_column("Value")
    for key, value in data.items():
        display_val = str(value) if value is not None else "[dim]—[/dim]"
        if len(display_val) > 200:
            display_val = display_val[:200] + "…"
        table.add_row(key, display_val)
    console.print(table)


def print_xml(xml_content: str, output_file: Optional[str] = None) -> None:
    """Print or save XML content."""
    if output_file:
        Path(output_file).write_text(xml_content, encoding="utf-8")
        err_console.print(f"[green]XML written to {output_file}[/green]")
    else:
        syntax = Syntax(xml_content, "xml", theme="monokai", line_numbers=False)
        console.print(syntax)


def print_items_table(
    items: List[Dict[str, Any]],
    title: Optional[str] = None,
    columns: Optional[List[str]] = None,
) -> None:
    """Print a list of dicts as a Rich table."""
    if not items:
        console.print("[dim]No items found.[/dim]")
        return
    cols = columns if columns is not None else list(items[0].keys())
    table = Table(title=title, title_style="bold cyan", show_lines=False, box=None, pad_edge=False)
    for col in cols:
        header = col.replace("_", " ").title()
        table.add_column(header, style="cyan" if col == "id" else "", no_wrap=(col == "id"))
    for item in items:
        row: List[str] = []
        for col in cols:
            val = item.get(col)
            if val is None:
                row.append("[dim]—[/dim]")
            elif isinstance(val, bool):
                row.append("[green]✓[/green]" if val else "[dim]✗[/dim]")
            else:
                s = str(val)
                if len(s) > 80:
                    s = s[:77] + "…"
                row.append(s)
        table.add_row(*row)
    console.print(table)


def print_transcript(messages: List[Dict[str, Any]], interview_id: str) -> None:
    """Pretty-print interview transcript messages."""
    console.print(f"\n[bold cyan]Transcript — {interview_id}[/bold cyan]\n")
    for msg in messages:
        role = msg.get("role", "unknown")
        content = msg.get("content", "")
        if role == "assistant":
            label = "[bold blue]Interviewer[/bold blue]"
        elif role == "user":
            label = "[bold green]Respondent[/bold green]"
        else:
            label = f"[dim]{role}[/dim]"
        console.print(f"{label}: {content}\n")


def prompt_select_from_list(items: List[Dict[str, Any]], display_fn: Any, prompt_text: str) -> Dict[str, Any]:
    """Display a numbered list and prompt the user to pick one item."""
    for i, item in enumerate(items, 1):
        console.print(f"  [dim]{i}.[/dim] {display_fn(item)}")
    idx = click.prompt(prompt_text, type=click.IntRange(1, len(items)))
    return items[idx - 1]
