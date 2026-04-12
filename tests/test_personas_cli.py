"""CLI integration tests for persona commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_SURVEY_ID


class TestPersonasGenerate:
    def test_generate_personas(self, invoke, persona_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_personas", return_value=persona_generation_response):
            result = invoke([
                "personas", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
            ])
            assert result.exit_code == 0
            assert "2 persona(s)" in result.output or "Generated" in result.output

    def test_generate_with_instructions(self, invoke, persona_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_personas", return_value=persona_generation_response) as mock:
            result = invoke([
                "personas", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
                "--instructions", "Include diverse age groups",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["additional_instructions"] == "Include diverse age groups"

    def test_generate_with_output(self, invoke, persona_generation_response, tmp_path):
        out = tmp_path / "personas.json"
        with patch("deutero_cli.client.DeuteroClient.generate_personas", return_value=persona_generation_response):
            result = invoke([
                "personas", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
                "--output", str(out),
            ])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert len(data["personas"]) == 2

    def test_generate_invalid_count_zero(self, invoke):
        result = invoke(["personas", "generate", SAMPLE_SURVEY_ID, "--count", "0"])
        assert result.exit_code == 1

    def test_generate_invalid_count_too_high(self, invoke):
        result = invoke(["personas", "generate", SAMPLE_SURVEY_ID, "--count", "26"])
        assert result.exit_code == 1

    def test_generate_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.generate_personas", side_effect=DeuteroAPIError(404, "Survey not found")):
            result = invoke(["personas", "generate", SAMPLE_SURVEY_ID, "--count", "3"])
            assert result.exit_code == 1
