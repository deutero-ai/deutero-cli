"""CLI integration tests for question commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_SURVEY_ID


class TestQuestionsGenerate:
    def test_generate_questions(self, invoke, question_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_questions", return_value=question_generation_response):
            result = invoke([
                "questions", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
            ])
            assert result.exit_code == 0
            assert "2 question(s)" in result.output or "Generated" in result.output

    def test_generate_with_instructions(self, invoke, question_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_questions", return_value=question_generation_response) as mock:
            result = invoke([
                "questions", "generate", SAMPLE_SURVEY_ID,
                "--count", "3",
                "--instructions", "Focus on emotional responses",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["additional_instructions"] == "Focus on emotional responses"

    def test_generate_with_output(self, invoke, question_generation_response, tmp_path):
        out = tmp_path / "questions.json"
        with patch("deutero_cli.client.DeuteroClient.generate_questions", return_value=question_generation_response):
            result = invoke([
                "questions", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
                "--output", str(out),
            ])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert len(data["question_list"]) == 2

    def test_generate_with_xml_output(self, invoke, question_generation_response, tmp_path):
        xml_out = tmp_path / "questions.xml"
        with patch("deutero_cli.client.DeuteroClient.generate_questions", return_value=question_generation_response):
            result = invoke([
                "questions", "generate", SAMPLE_SURVEY_ID,
                "--count", "2",
                "--xml-output", str(xml_out),
            ])
            assert result.exit_code == 0
            assert xml_out.exists()

    def test_generate_invalid_count_zero(self, invoke):
        result = invoke(["questions", "generate", SAMPLE_SURVEY_ID, "--count", "0"])
        assert result.exit_code == 1

    def test_generate_invalid_count_too_high(self, invoke):
        result = invoke(["questions", "generate", SAMPLE_SURVEY_ID, "--count", "26"])
        assert result.exit_code == 1

    def test_generate_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.generate_questions", side_effect=DeuteroAPIError(502, "Unable to generate questions.")):
            result = invoke(["questions", "generate", SAMPLE_SURVEY_ID, "--count", "5"])
            assert result.exit_code == 1
