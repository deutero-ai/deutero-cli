"""CLI integration tests for question commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_IMAGE_ID, SAMPLE_QUESTION_ID, SAMPLE_SURVEY_ID


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


class TestQuestionsManage:
    def test_create_question(self, invoke, question_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_question", return_value=question_create_response) as mock:
            result = invoke([
                "questions", "create", SAMPLE_SURVEY_ID,
                "--question", "What changed?",
                "--option", "A",
                "--option", "B",
                "--follow-up",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["survey_id"] == SAMPLE_SURVEY_ID
            assert payload["options"] == ["A", "B"]
            assert payload["follow_up"] is True

    def test_get_question(self, invoke, question_response):
        with patch("deutero_cli.client.DeuteroClient.get_question", return_value=question_response):
            result = invoke(["questions", "get", SAMPLE_QUESTION_ID])
            assert result.exit_code == 0
            assert SAMPLE_QUESTION_ID in result.output

    def test_update_question(self, invoke, question_response):
        with patch("deutero_cli.client.DeuteroClient.update_question", return_value=question_response) as mock:
            result = invoke([
                "questions", "update", SAMPLE_QUESTION_ID,
                "--question", "Updated?",
                "--null-field", "scale",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][1]
            assert payload["question"] == "Updated?"
            assert payload["scale"] is None

    def test_update_question_requires_payload(self, invoke):
        result = invoke(["questions", "update", SAMPLE_QUESTION_ID])
        assert result.exit_code == 1

    def test_delete_question(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.delete_question", return_value={"success": True}) as mock:
            result = invoke(["questions", "delete", SAMPLE_QUESTION_ID, "--survey-id", SAMPLE_SURVEY_ID])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_QUESTION_ID, SAMPLE_SURVEY_ID)

    def test_reorder_questions(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.reorder_questions", return_value={"success": True}) as mock:
            result = invoke(["questions", "reorder", SAMPLE_SURVEY_ID, SAMPLE_QUESTION_ID])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_SURVEY_ID, [SAMPLE_QUESTION_ID])

    def test_upload_image_data(self, invoke, question_image_response):
        with patch("deutero_cli.client.DeuteroClient.upload_question_image", return_value=question_image_response) as mock:
            result = invoke([
                "questions", "upload-image", SAMPLE_QUESTION_ID,
                "--image-data", "data:image/png;base64,abc",
                "--label", "Concept",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][1]
            assert payload["label"] == "Concept"

    def test_reorder_images(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.reorder_question_images", return_value={"success": True}) as mock:
            result = invoke(["questions", "reorder-images", SAMPLE_QUESTION_ID, SAMPLE_IMAGE_ID])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_QUESTION_ID, [SAMPLE_IMAGE_ID])
