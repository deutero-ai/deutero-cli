"""CLI integration tests for credit commands."""

from __future__ import annotations

from unittest.mock import patch

from tests.conftest import SAMPLE_SURVEY_ID


class TestCreditsBalance:
    def test_balance(self, invoke, credit_balance_response):
        with patch("deutero_cli.client.DeuteroClient.get_credit_balance", return_value=credit_balance_response):
            result = invoke(["credits", "balance"])
            assert result.exit_code == 0
            assert "75" in result.output


class TestCreditsEstimate:
    def test_estimate_simulation(self, invoke, credit_estimation_response):
        with patch("deutero_cli.client.DeuteroClient.estimate_simulation_credits", return_value=credit_estimation_response) as mock:
            result = invoke([
                "credits", "estimate-simulation", SAMPLE_SURVEY_ID,
                "--model-tier", "premium",
                "--participants", "4",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["survey_id"] == SAMPLE_SURVEY_ID
            assert payload["model_tier"] == "premium"
            assert payload["num_participants"] == 4

    def test_estimate_analysis(self, invoke, credit_estimation_response):
        with patch("deutero_cli.client.DeuteroClient.estimate_analysis_credits", return_value=credit_estimation_response) as mock:
            result = invoke([
                "credits", "estimate-analysis", SAMPLE_SURVEY_ID,
                "--model-tier", "premium",
                "--interviews", "4",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["num_participants"] == 4

    def test_estimate_survey(self, invoke, credit_estimation_response):
        with patch("deutero_cli.client.DeuteroClient.estimate_survey_credits", return_value=credit_estimation_response) as mock:
            result = invoke([
                "credits", "estimate-survey", SAMPLE_SURVEY_ID,
                "--model-tier", "premium",
                "--participants", "4",
                "--include-analysis",
            ])
            assert result.exit_code == 0
            payload = mock.call_args[0][0]
            assert payload["include_analysis"] is True

    def test_estimate_rejects_zero_count(self, invoke):
        result = invoke(["credits", "estimate-simulation", SAMPLE_SURVEY_ID, "--participants", "0"])
        assert result.exit_code == 1
