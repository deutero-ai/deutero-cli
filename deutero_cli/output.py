"""Output formatting helpers using Rich."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any, Dict, Optional

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
