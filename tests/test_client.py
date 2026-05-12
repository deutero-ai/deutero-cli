"""Tests for the DeuteroClient HTTP wrapper."""

from __future__ import annotations

import httpx
import pytest

from deutero_cli.client import DeuteroAPIError, DeuteroClient

from .conftest import SAMPLE_IMAGE_ID, SAMPLE_INTERVIEW_ID, SAMPLE_PERSONA_ID, SAMPLE_QUESTION_ID, SAMPLE_SURVEY_ID, TEST_API_KEY, TEST_BASE_URL


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

    def test_survey_model_tier(self, httpx_mock, client, model_tier_response):
        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/model-tier",
            json=model_tier_response,
        )
        result = client.get_survey_model_tier(SAMPLE_SURVEY_ID)
        assert result["model_tier"] == "premium"

    def test_set_survey_model_tier(self, httpx_mock, client, model_tier_response):
        httpx_mock.add_response(
            method="PUT",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/model-tier",
            json=model_tier_response,
        )
        result = client.set_survey_model_tier(SAMPLE_SURVEY_ID, "premium")
        assert result["model_provider"] == "anthropic"

    def test_suggest_from_site(self, httpx_mock, client, suggest_from_site_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/suggest-from-site",
            json=suggest_from_site_response,
        )
        result = client.suggest_from_site({"url": "https://example.com"})
        assert result["is_valid"] is True

    def test_question_crud_and_ordering(self, httpx_mock, client, question_response, question_create_response):
        httpx_mock.add_response(method="POST", url=f"{TEST_BASE_URL}/api/v1/questions", json=question_create_response)
        assert client.create_question({"survey_id": SAMPLE_SURVEY_ID, "question": "Q"})["id"] == SAMPLE_QUESTION_ID

        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/questions/{SAMPLE_QUESTION_ID}",
            json=question_response,
        )
        assert client.get_question(SAMPLE_QUESTION_ID)["question"]

        httpx_mock.add_response(
            method="PUT",
            url=f"{TEST_BASE_URL}/api/v1/questions/{SAMPLE_QUESTION_ID}",
            json=question_response,
        )
        assert client.update_question(SAMPLE_QUESTION_ID, {"question": "Updated"})["id"] == SAMPLE_QUESTION_ID

        httpx_mock.add_response(
            method="DELETE",
            url=f"{TEST_BASE_URL}/api/v1/questions/{SAMPLE_QUESTION_ID}?survey_id={SAMPLE_SURVEY_ID}",
            json={"success": True},
        )
        assert client.delete_question(SAMPLE_QUESTION_ID, SAMPLE_SURVEY_ID)["success"] is True

        httpx_mock.add_response(
            method="PUT",
            url=f"{TEST_BASE_URL}/api/v1/surveys/{SAMPLE_SURVEY_ID}/questions/reorder",
            json={"success": True},
        )
        assert client.reorder_questions(SAMPLE_SURVEY_ID, [SAMPLE_QUESTION_ID])["success"] is True

    def test_question_images(self, httpx_mock, client, question_image_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/questions/{SAMPLE_QUESTION_ID}/images",
            json=question_image_response,
        )
        result = client.upload_question_image(SAMPLE_QUESTION_ID, {"image_data": "data:image/png;base64,abc"})
        assert result["id"] == SAMPLE_IMAGE_ID

        httpx_mock.add_response(
            method="PUT",
            url=f"{TEST_BASE_URL}/api/v1/questions/{SAMPLE_QUESTION_ID}/images/reorder",
            json={"success": True},
        )
        assert client.reorder_question_images(SAMPLE_QUESTION_ID, [SAMPLE_IMAGE_ID])["success"] is True

    def test_credits(self, httpx_mock, client, credit_estimation_response, credit_balance_response):
        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/credits/estimate/simulation",
            json=credit_estimation_response,
        )
        assert client.estimate_simulation_credits({"survey_id": SAMPLE_SURVEY_ID})["estimated_credits"] == 12

        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/credits/estimate/analysis",
            json=credit_estimation_response,
        )
        assert client.estimate_analysis_credits({"survey_id": SAMPLE_SURVEY_ID})["num_questions"] == 5

        httpx_mock.add_response(
            method="POST",
            url=f"{TEST_BASE_URL}/api/v1/credits/estimate/survey",
            json=credit_estimation_response,
        )
        assert client.estimate_survey_credits({"survey_id": SAMPLE_SURVEY_ID})["model_tier"] == "premium"

        httpx_mock.add_response(
            method="GET",
            url=f"{TEST_BASE_URL}/api/v1/credits/balance",
            json=credit_balance_response,
        )
        assert client.get_credit_balance()["net_available"] == 75
