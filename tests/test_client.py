"""Tests for the DeuteroClient HTTP wrapper."""

from __future__ import annotations

import httpx
import pytest

from deutero_cli.client import DeuteroAPIError, DeuteroClient

from .conftest import SAMPLE_INTERVIEW_ID, SAMPLE_PERSONA_ID, SAMPLE_SURVEY_ID, TEST_API_KEY, TEST_BASE_URL


class TestClientHeaders:
    def test_headers_include_api_key(self, client):
        headers = client._headers()
        assert headers["X-API-Key"] == TEST_API_KEY
        assert headers["Content-Type"] == "application/json"

    def test_url_construction(self, client):
        assert client._url("/surveys/generate") == f"{TEST_BASE_URL}/api/v1/surveys/generate"


class TestClientPost:
    def test_post_success(self, httpx_mock, client, survey_generation_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/surveys/generate",
            json=survey_generation_response,
        )
        result = client.generate_survey({"survey_type": "user_experience", "business_context": "test", "research_need": "test"})
        assert result["study_id"] == SAMPLE_SURVEY_ID

    def test_post_api_error(self, httpx_mock, client):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/surveys/generate",
            status_code=400,
            json={"detail": "Missing required fields"},
        )
        with pytest.raises(DeuteroAPIError) as exc_info:
            client.generate_survey({})
        assert exc_info.value.status_code == 400
        assert "Missing required fields" in str(exc_info.value)


class TestClientGet:
    def test_get_participation(self, httpx_mock, client, participation_response):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/participation",
            json=participation_response,
        )
        result = client.get_participation(SAMPLE_SURVEY_ID)
        assert result["total_interviews"] == 10
        assert result["completed_interviews"] == 7

    def test_get_analysis_status(self, httpx_mock, client, analysis_status_response):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/analysis/status?interview_id={SAMPLE_INTERVIEW_ID}",
            json=analysis_status_response,
        )
        result = client.get_analysis_status({"interview_id": SAMPLE_INTERVIEW_ID})
        assert result["total_interviews"] == 1

    def test_get_404_error(self, httpx_mock, client):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/participation",
            status_code=404,
            json={"detail": "Survey not found"},
        )
        with pytest.raises(DeuteroAPIError) as exc_info:
            client.get_participation(SAMPLE_SURVEY_ID)
        assert exc_info.value.status_code == 404


class TestClientEndpoints:
    def test_generate_questions(self, httpx_mock, client, question_generation_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/questions/generate",
            json=question_generation_response,
        )
        result = client.generate_questions({"survey_id": SAMPLE_SURVEY_ID, "number_of_questions": 2})
        assert len(result["question_list"]) == 2

    def test_generate_personas(self, httpx_mock, client, persona_generation_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/personas/generate",
            json=persona_generation_response,
        )
        result = client.generate_personas({"survey_id": SAMPLE_SURVEY_ID, "number_of_personas": 2})
        assert len(result["personas"]) == 2

    def test_simulate_interview(self, httpx_mock, client, interview_simulation_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/interviews/simulate",
            json=interview_simulation_response,
        )
        result = client.simulate_interview({"survey_id": SAMPLE_SURVEY_ID, "persona_id": SAMPLE_PERSONA_ID})
        assert result["transcript_url"]

    def test_run_analysis(self, httpx_mock, client, analysis_run_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/analysis/run",
            json=analysis_run_response,
        )
        result = client.run_analysis({"survey_id": SAMPLE_SURVEY_ID, "model_tier": "open_weights"})
        assert result["success"] is True

    def test_get_agent_requirements(self, httpx_mock, client, agent_requirements_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/agent-requirements",
            json=agent_requirements_response,
        )
        result = client.get_agent_requirements(SAMPLE_SURVEY_ID)
        assert "markdown" in result

    def test_get_interview_analysis_results(self, httpx_mock, client, interview_results_response):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/analysis/results/interview?interview_id={SAMPLE_INTERVIEW_ID}&phase=initial_engagement",
            json=interview_results_response,
        )
        result = client.get_interview_analysis_results(SAMPLE_INTERVIEW_ID, "initial_engagement")
        assert result["xml_output"]

    def test_get_survey_analysis_results(self, httpx_mock, client, survey_results_response):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/analysis/results/survey?survey_id={SAMPLE_SURVEY_ID}",
            json=survey_results_response,
        )
        result = client.get_survey_analysis_results(SAMPLE_SURVEY_ID)
        assert result["xml_output"]
