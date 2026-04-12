"""Tests for config management and the config CLI commands."""

from __future__ import annotations

import json
from pathlib import Path

from deutero_cli.config import (
    get_api_key,
    get_base_url,
    save_api_key,
    save_base_url,
    DEFAULT_BASE_URL,
)


class TestConfigFunctions:
    def test_get_api_key_from_env(self, monkeypatch):
        monkeypatch.setenv("DEUTERO_API_KEY", "env-key-abc")
        assert get_api_key() == "env-key-abc"

    def test_get_api_key_from_file(self, monkeypatch, tmp_path):
        monkeypatch.delenv("DEUTERO_API_KEY", raising=False)
        cfg = tmp_path / "config.json"
        cfg.write_text(json.dumps({"api_key": "file-key-xyz"}))
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        assert get_api_key() == "file-key-xyz"

    def test_get_api_key_returns_none(self, monkeypatch, tmp_path):
        monkeypatch.delenv("DEUTERO_API_KEY", raising=False)
        cfg = tmp_path / "config.json"
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        assert get_api_key() is None

    def test_get_base_url_default(self, monkeypatch, tmp_path):
        monkeypatch.delenv("DEUTERO_BASE_URL", raising=False)
        cfg = tmp_path / "config.json"
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        assert get_base_url() == DEFAULT_BASE_URL

    def test_get_base_url_from_env(self, monkeypatch):
        monkeypatch.setenv("DEUTERO_BASE_URL", "http://custom:9000")
        assert get_base_url() == "http://custom:9000"

    def test_save_and_load_api_key(self, monkeypatch, tmp_path):
        monkeypatch.delenv("DEUTERO_API_KEY", raising=False)
        cfg = tmp_path / "config.json"
        cfg_dir = tmp_path
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        monkeypatch.setattr("deutero_cli.config.CONFIG_DIR", cfg_dir)
        save_api_key("saved-key-123")
        assert get_api_key() == "saved-key-123"

    def test_save_and_load_base_url(self, monkeypatch, tmp_path):
        monkeypatch.delenv("DEUTERO_BASE_URL", raising=False)
        cfg = tmp_path / "config.json"
        cfg_dir = tmp_path
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        monkeypatch.setattr("deutero_cli.config.CONFIG_DIR", cfg_dir)
        save_base_url("http://new-url:8080")
        assert get_base_url() == "http://new-url:8080"


class TestConfigCLI:
    def test_config_show(self, invoke, monkeypatch):
        monkeypatch.setenv("DEUTERO_API_KEY", "test-api-key-0123456789abcdef")
        result = invoke(["config", "show"])
        assert result.exit_code == 0

    def test_config_set_key(self, invoke, monkeypatch, tmp_path):
        cfg = tmp_path / "config.json"
        cfg_dir = tmp_path
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        monkeypatch.setattr("deutero_cli.config.CONFIG_DIR", cfg_dir)
        result = invoke(["config", "set-key", "new-key-value"])
        assert result.exit_code == 0
        assert "saved" in result.output.lower() or "saved" in result.stderr.lower()
        saved = json.loads(cfg.read_text())
        assert saved["api_key"] == "new-key-value"

    def test_config_set_url(self, invoke, monkeypatch, tmp_path):
        cfg = tmp_path / "config.json"
        cfg_dir = tmp_path
        monkeypatch.setattr("deutero_cli.config.CONFIG_FILE", cfg)
        monkeypatch.setattr("deutero_cli.config.CONFIG_DIR", cfg_dir)
        result = invoke(["config", "set-url", "http://new:1234"])
        assert result.exit_code == 0
        saved = json.loads(cfg.read_text())
        assert saved["base_url"] == "http://new:1234"
