"""CLI integration tests for survey commands."""

from __future__ import annotations

import json
from unittest.mock import patch, MagicMock

from tests.conftest import SAMPLE_SURVEY_ID


class TestSurveysGenerate:
    def test_generate_ux_survey(self, invoke, survey_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "user_experience",
                "--business-context", "SaaS product for team collaboration",
                "--research-need", "Understand onboarding pain points",
                "--target-users", "Engineering managers at mid-size companies",
            ])
            assert result.exit_code == 0
            assert SAMPLE_SURVEY_ID in result.output

    def test_generate_sociology_survey(self, invoke, survey_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "sociology",
                "--research-question", "How do communities form online?",
                "--population-of-interest", "Reddit users",
            ])
            assert result.exit_code == 0

    def test_generate_customer_development_survey(self, invoke, survey_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "customer_development",
                "--problem-hypothesis", "Teams struggle with async communication",
                "--customer-segment", "Remote-first startups",
            ])
            assert result.exit_code == 0

    def test_generate_polling_survey(self, invoke, survey_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "polling",
                "--research-question", "Public opinion on remote work policies",
                "--population-segment", "US workers 25-55",
            ])
            assert result.exit_code == 0

    def test_generate_with_output_file(self, invoke, survey_generation_response, tmp_path):
        out = tmp_path / "out.json"
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "user_experience",
                "--business-context", "test",
                "--research-need", "test",
                "--output", str(out),
            ])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["study_id"] == SAMPLE_SURVEY_ID

    def test_generate_with_xml_output(self, invoke, survey_generation_response, tmp_path):
        xml_out = tmp_path / "spec.xml"
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response):
            result = invoke([
                "surveys", "generate",
                "--survey-type", "user_experience",
                "--business-context", "test",
                "--research-need", "test",
                "--xml-output", str(xml_out),
            ])
            assert result.exit_code == 0
            assert xml_out.exists()
            assert "surveyExport" in xml_out.read_text()

    def test_generate_with_language(self, invoke, survey_generation_response):
        with patch("deutero_cli.client.DeuteroClient.generate_survey", return_value=survey_generation_response) as mock:
            result = invoke([
                "surveys", "generate",
                "--survey-type", "sociology",
                "--language", "Spanish",
                "--research-question", "¿Cómo se forman las comunidades?",
                "--population-of-interest", "Usuarios de redes sociales",
            ])
            assert result.exit_code == 0
            call_payload = mock.call_args[0][0]
            assert call_payload["language"] == "Spanish"

    def test_generate_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.generate_survey", side_effect=DeuteroAPIError(400, "Missing required fields")):
            result = invoke(["surveys", "generate", "--survey-type", "user_experience"])
            assert result.exit_code == 1


class TestSurveysParticipation:
    def test_participation_success(self, invoke, participation_response):
        with patch("deutero_cli.client.DeuteroClient.get_participation", return_value=participation_response):
            result = invoke(["surveys", "participation", SAMPLE_SURVEY_ID])
            assert result.exit_code == 0
            assert "70" in result.output

    def test_participation_with_output(self, invoke, participation_response, tmp_path):
        out = tmp_path / "stats.json"
        with patch("deutero_cli.client.DeuteroClient.get_participation", return_value=participation_response):
            result = invoke(["surveys", "participation", SAMPLE_SURVEY_ID, "--output", str(out)])
            assert result.exit_code == 0
            assert out.exists()

    def test_participation_not_found(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.get_participation", side_effect=DeuteroAPIError(404, "Survey not found")):
            result = invoke(["surveys", "participation", SAMPLE_SURVEY_ID])
            assert result.exit_code == 1


class TestSurveysAgentRequirements:
    def test_agent_requirements_success(self, invoke, agent_requirements_response, tmp_path):
        out = tmp_path / "reqs.md"
        with patch("deutero_cli.client.DeuteroClient.get_agent_requirements", return_value=agent_requirements_response):
            result = invoke(["surveys", "agent-requirements", SAMPLE_SURVEY_ID, "--output", str(out)])
            assert result.exit_code == 0
            assert out.exists()
            assert "Requirement 1" in out.read_text()

    def test_agent_requirements_default_filename(self, invoke, agent_requirements_response, monkeypatch, tmp_path):
        monkeypatch.chdir(tmp_path)
        with patch("deutero_cli.client.DeuteroClient.get_agent_requirements", return_value=agent_requirements_response):
            result = invoke(["surveys", "agent-requirements", SAMPLE_SURVEY_ID])
            assert result.exit_code == 0
