"""CLI integration tests for analysis commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_INTERVIEW_ID, SAMPLE_SURVEY_ID


class TestAnalysisRun:
    def test_run_by_interview(self, invoke, analysis_run_response):
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_response):
            result = invoke([
                "analysis", "run",
                "--interview-id", SAMPLE_INTERVIEW_ID,
            ])
            assert result.exit_code == 0
            assert "queued" in result.output.lower() or "success" in result.output.lower()

    def test_run_by_survey(self, invoke, analysis_run_response):
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_response):
            result = invoke([
                "analysis", "run",
                "--survey-id", SAMPLE_SURVEY_ID,
            ])
            assert result.exit_code == 0

    def test_run_cross_case(self, invoke, analysis_run_cross_case_response):
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_cross_case_response):
            result = invoke([
                "analysis", "run",
                "--survey-id", SAMPLE_SURVEY_ID,
                "--cross-case",
            ])
            assert result.exit_code == 0

    def test_run_cross_case_xml_output(self, invoke, analysis_run_cross_case_response, tmp_path):
        xml_out = tmp_path / "cross.xml"
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_cross_case_response):
            result = invoke([
                "analysis", "run",
                "--survey-id", SAMPLE_SURVEY_ID,
                "--cross-case",
                "--xml-output", str(xml_out),
            ])
            assert result.exit_code == 0
            assert xml_out.exists()
            assert "crossCaseAnalysis" in xml_out.read_text()

    def test_run_with_model_tier(self, invoke, analysis_run_response):
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_response) as mock:
            result = invoke([
                "analysis", "run",
                "--interview-id", SAMPLE_INTERVIEW_ID,
                "--model-tier", "frontier",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["model_tier"] == "frontier"

    def test_run_with_output(self, invoke, analysis_run_response, tmp_path):
        out = tmp_path / "analysis.json"
        with patch("deutero_cli.client.DeuteroClient.run_analysis", return_value=analysis_run_response):
            result = invoke([
                "analysis", "run",
                "--interview-id", SAMPLE_INTERVIEW_ID,
                "--output", str(out),
            ])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["success"] is True

    def test_run_missing_ids(self, invoke):
        result = invoke(["analysis", "run"])
        assert result.exit_code == 1

    def test_run_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.run_analysis", side_effect=DeuteroAPIError(400, "Provide either interview_id or survey_id")):
            result = invoke(["analysis", "run", "--interview-id", SAMPLE_INTERVIEW_ID])
            assert result.exit_code == 1


class TestAnalysisStatus:
    def test_status_by_interview(self, invoke, analysis_status_response):
        with patch("deutero_cli.client.DeuteroClient.get_analysis_status", return_value=analysis_status_response):
            result = invoke([
                "analysis", "status",
                "--interview-id", SAMPLE_INTERVIEW_ID,
            ])
            assert result.exit_code == 0

    def test_status_by_survey(self, invoke, analysis_status_response):
        resp = {**analysis_status_response, "interview_id": None, "survey_id": SAMPLE_SURVEY_ID}
        with patch("deutero_cli.client.DeuteroClient.get_analysis_status", return_value=resp):
            result = invoke([
                "analysis", "status",
                "--survey-id", SAMPLE_SURVEY_ID,
            ])
            assert result.exit_code == 0

    def test_status_with_output(self, invoke, analysis_status_response, tmp_path):
        out = tmp_path / "status.json"
        with patch("deutero_cli.client.DeuteroClient.get_analysis_status", return_value=analysis_status_response):
            result = invoke([
                "analysis", "status",
                "--interview-id", SAMPLE_INTERVIEW_ID,
                "--output", str(out),
            ])
            assert result.exit_code == 0
            assert out.exists()

    def test_status_missing_ids(self, invoke):
        result = invoke(["analysis", "status"])
        assert result.exit_code == 1


class TestAnalysisResultsInterview:
    def test_results_interview(self, invoke, interview_results_response):
        with patch("deutero_cli.client.DeuteroClient.get_interview_analysis_results", return_value=interview_results_response):
            result = invoke([
                "analysis", "results-interview",
                SAMPLE_INTERVIEW_ID,
                "--phase", "initial_engagement",
            ])
            assert result.exit_code == 0

    def test_results_interview_to_file(self, invoke, interview_results_response, tmp_path):
        out = tmp_path / "phase1.xml"
        with patch("deutero_cli.client.DeuteroClient.get_interview_analysis_results", return_value=interview_results_response):
            result = invoke([
                "analysis", "results-interview",
                SAMPLE_INTERVIEW_ID,
                "--phase", "initial_engagement",
                "--output", str(out),
            ])
            assert result.exit_code == 0
            assert out.exists()

    def test_results_interview_error(self, invoke):
        resp = {
            "interview_id": SAMPLE_INTERVIEW_ID,
            "phase": "initial_engagement",
            "xml_output": None,
            "completed_at": None,
            "error": "Phase initial_engagement has not been completed yet",
        }
        with patch("deutero_cli.client.DeuteroClient.get_interview_analysis_results", return_value=resp):
            result = invoke([
                "analysis", "results-interview",
                SAMPLE_INTERVIEW_ID,
                "--phase", "initial_engagement",
            ])
            assert result.exit_code == 0


class TestAnalysisResultsSurvey:
    def test_results_survey(self, invoke, survey_results_response):
        with patch("deutero_cli.client.DeuteroClient.get_survey_analysis_results", return_value=survey_results_response):
            result = invoke([
                "analysis", "results-survey",
                SAMPLE_SURVEY_ID,
            ])
            assert result.exit_code == 0

    def test_results_survey_to_file(self, invoke, survey_results_response, tmp_path):
        out = tmp_path / "cross.xml"
        with patch("deutero_cli.client.DeuteroClient.get_survey_analysis_results", return_value=survey_results_response):
            result = invoke([
                "analysis", "results-survey",
                SAMPLE_SURVEY_ID,
                "--output", str(out),
            ])
            assert result.exit_code == 0
            assert out.exists()

    def test_results_survey_not_found(self, invoke):
        resp = {
            "survey_id": SAMPLE_SURVEY_ID,
            "xml_output": None,
            "completed_at": None,
            "error": "No cross-case analysis found for this survey",
        }
        with patch("deutero_cli.client.DeuteroClient.get_survey_analysis_results", return_value=resp):
            result = invoke([
                "analysis", "results-survey",
                SAMPLE_SURVEY_ID,
            ])
            assert result.exit_code == 0
