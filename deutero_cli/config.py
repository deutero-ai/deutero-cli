"""Configuration management for the Deutero CLI.

Reads settings from environment variables and a local config file (~/.deutero/config.json).
Environment variables take precedence over the config file.
"""

from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Optional


CONFIG_DIR = Path.home() / ".deutero"
CONFIG_FILE = CONFIG_DIR / "config.json"

ENV_API_KEY = "DEUTERO_API_KEY"
ENV_BASE_URL = "DEUTERO_BASE_URL"

DEFAULT_BASE_URL = "http://dashboard.deutero.ai:8000"


def _load_config_file() -> dict:
    if CONFIG_FILE.exists():
        try:
            return json.loads(CONFIG_FILE.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, OSError):
            return {}
    return {}


def _save_config_file(data: dict) -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    CONFIG_FILE.write_text(json.dumps(data, indent=2), encoding="utf-8")


def get_api_key() -> Optional[str]:
    return os.environ.get(ENV_API_KEY) or _load_config_file().get("api_key")


def get_base_url() -> str:
    return os.environ.get(ENV_BASE_URL) or _load_config_file().get("base_url") or DEFAULT_BASE_URL


def save_api_key(api_key: str) -> None:
    data = _load_config_file()
    data["api_key"] = api_key
    _save_config_file(data)


def save_base_url(base_url: str) -> None:
    data = _load_config_file()
    data["base_url"] = base_url
    _save_config_file(data)


def get_active_project_id() -> Optional[str]:
    return _load_config_file().get("active_project_id")


def save_active_project_id(project_id: str) -> None:
    data = _load_config_file()
    data["active_project_id"] = project_id
    _save_config_file(data)


def get_active_survey_id() -> Optional[str]:
    return _load_config_file().get("active_survey_id")


def save_active_survey_id(survey_id: str) -> None:
    data = _load_config_file()
    data["active_survey_id"] = survey_id
    _save_config_file(data)
