"""CLI integration tests for interview commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_INTERVIEW_ID, SAMPLE_PERSONA_ID, SAMPLE_SURVEY_ID


class TestInterviewsSimulate:
    def test_simulate_interview(self, invoke, interview_simulation_response):
        with patch("deutero_cli.client.DeuteroClient.simulate_interview", return_value=interview_simulation_response):
            result = invoke([
                "interviews", "simulate",
                SAMPLE_SURVEY_ID,
                SAMPLE_PERSONA_ID,
            ])
            assert result.exit_code == 0
            assert "simulation started" in result.output.lower() or "transcript" in result.output.lower()

    def test_simulate_with_output(self, invoke, interview_simulation_response, tmp_path):
        out = tmp_path / "sim.json"
        with patch("deutero_cli.client.DeuteroClient.simulate_interview", return_value=interview_simulation_response):
            result = invoke([
                "interviews", "simulate",
                SAMPLE_SURVEY_ID,
                SAMPLE_PERSONA_ID,
                "--output", str(out),
            ])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["survey_id"] == SAMPLE_SURVEY_ID

    def test_simulate_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.simulate_interview", side_effect=DeuteroAPIError(404, "Survey not found")):
            result = invoke([
                "interviews", "simulate",
                SAMPLE_SURVEY_ID,
                SAMPLE_PERSONA_ID,
            ])
            assert result.exit_code == 1

    def test_simulate_payload_correctness(self, invoke, interview_simulation_response):
        with patch("deutero_cli.client.DeuteroClient.simulate_interview", return_value=interview_simulation_response) as mock:
            invoke([
                "interviews", "simulate",
                SAMPLE_SURVEY_ID,
                SAMPLE_PERSONA_ID,
            ])
            payload = mock.call_args[0][0]
            assert payload["survey_id"] == SAMPLE_SURVEY_ID
            assert payload["persona_id"] == SAMPLE_PERSONA_ID
