"""CLI integration tests for surveys create and update commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_PROJECT_ID, SAMPLE_SURVEY_ID


class TestSurveysCreate:
    def test_create_required_only(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID])
            assert result.exit_code == 0
            assert "Survey created" in result.output
            assert SAMPLE_SURVEY_ID in result.output
            mock.assert_called_once_with({"project_id": SAMPLE_PROJECT_ID})

    def test_create_with_survey_type(self, invoke, survey_create_response):
        resp = {**survey_create_response, "survey_type": "sociology"}
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=resp) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--survey-type", "sociology"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["survey_type"] == "sociology"

    def test_create_with_name(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--name", "My Study"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["name"] == "My Study"

    def test_create_with_model_tier(self, invoke, survey_create_response):
        resp = {**survey_create_response, "model_tier": "frontier"}
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=resp) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--model-tier", "frontier"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["model_tier"] == "frontier"

    def test_create_with_anonymous_flag(self, invoke, survey_create_response):
        resp = {**survey_create_response, "anonymous": True}
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=resp) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--anonymous"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["anonymous"] is True

    def test_create_with_no_anonymous_flag(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--no-anonymous"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["anonymous"] is False

    def test_create_with_all_optional_fields(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response) as mock:
            result = invoke([
                "surveys", "create",
                "--project-id", SAMPLE_PROJECT_ID,
                "--survey-type", "user_experience",
                "--name", "Onboarding Study",
                "--description", "A study about onboarding.",
                "--research-questions", "How do users experience onboarding?",
                "--objectives", "Identify pain points.",
                "--target-population", "New users.",
                "--model-tier", "open_weights",
                "--redirect-url", "https://example.com/done",
                "--language", "English",
            ])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][0]
            assert called_payload["project_id"] == SAMPLE_PROJECT_ID
            assert called_payload["survey_type"] == "user_experience"
            assert called_payload["name"] == "Onboarding Study"
            assert called_payload["description"] == "A study about onboarding."
            assert called_payload["research_questions"] == "How do users experience onboarding?"
            assert called_payload["objectives"] == "Identify pain points."
            assert called_payload["target_population"] == "New users."
            assert called_payload["model_tier"] == "open_weights"
            assert called_payload["redirect_url"] == "https://example.com/done"
            assert called_payload["language"] == "English"

    def test_create_missing_project_id_error(self, invoke):
        result = invoke(["surveys", "create"])
        assert result.exit_code != 0

    def test_create_with_output_file(self, invoke, survey_create_response, tmp_path):
        out = tmp_path / "survey.json"
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response):
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID, "--output", str(out)])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["study_id"] == SAMPLE_SURVEY_ID

    def test_create_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.create_survey", side_effect=DeuteroAPIError(400, "Invalid survey_type 'bad'. Must be one of: sociology, ...")):
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID])
            assert result.exit_code == 1

    def test_create_api_403(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.create_survey", side_effect=DeuteroAPIError(403, "You do not have access to this project.")):
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID])
            assert result.exit_code == 1

    def test_create_shows_study_id_in_output(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response):
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID])
            assert SAMPLE_SURVEY_ID in result.output

    def test_create_shows_url_in_output(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.create_survey", return_value=survey_create_response):
            result = invoke(["surveys", "create", "--project-id", SAMPLE_PROJECT_ID])
            assert "dashboard.deutero.ai" in result.output


class TestSurveysUpdate:
    def test_update_name(self, invoke, survey_create_response):
        updated = {**survey_create_response, "name": "Renamed Study"}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--name", "Renamed Study"])
            assert result.exit_code == 0
            assert "Survey updated" in result.output
            mock.assert_called_once_with(SAMPLE_SURVEY_ID, {"name": "Renamed Study"})

    def test_update_survey_type(self, invoke, survey_create_response):
        updated = {**survey_create_response, "survey_type": "polling"}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--survey-type", "polling"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_SURVEY_ID, {"survey_type": "polling"})

    def test_update_model_tier(self, invoke, survey_create_response):
        updated = {**survey_create_response, "model_tier": "premium"}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--model-tier", "premium"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_SURVEY_ID, {"model_tier": "premium"})

    def test_update_anonymous(self, invoke, survey_create_response):
        updated = {**survey_create_response, "anonymous": True}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--anonymous"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["anonymous"] is True

    def test_update_no_anonymous(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--no-anonymous"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["anonymous"] is False

    def test_update_anonymous_wins_when_both_passed(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--anonymous", "--no-anonymous"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["anonymous"] is True

    def test_update_multiple_fields(self, invoke, survey_create_response):
        updated = {**survey_create_response, "name": "New Name", "language": "Spanish"}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--name", "New Name", "--language", "Spanish"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_SURVEY_ID, {"name": "New Name", "language": "Spanish"})

    def test_update_no_fields_error(self, invoke):
        result = invoke(["surveys", "update", SAMPLE_SURVEY_ID])
        assert result.exit_code == 1
        assert "No fields to update" in result.output

    def test_update_with_output_file(self, invoke, survey_create_response, tmp_path):
        out = tmp_path / "survey.json"
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=survey_create_response):
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--name", "Test", "--output", str(out)])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["study_id"] == SAMPLE_SURVEY_ID

    def test_update_api_error_404(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.update_survey", side_effect=DeuteroAPIError(404, "Survey not found.")):
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--name", "X"])
            assert result.exit_code == 1

    def test_update_api_error_400(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.update_survey", side_effect=DeuteroAPIError(400, "Invalid model_tier.")):
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--model-tier", "open_weights"])
            assert result.exit_code == 1

    def test_update_redirect_url(self, invoke, survey_create_response):
        updated = {**survey_create_response, "redirect_url": "https://example.com/thanks"}
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=updated) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--redirect-url", "https://example.com/thanks"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["redirect_url"] == "https://example.com/thanks"

    def test_update_research_questions(self, invoke, survey_create_response):
        with patch("deutero_cli.client.DeuteroClient.update_survey", return_value=survey_create_response) as mock:
            result = invoke(["surveys", "update", SAMPLE_SURVEY_ID, "--research-questions", "What motivates users?"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["research_questions"] == "What motivates users?"
