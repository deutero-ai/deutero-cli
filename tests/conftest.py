"""Shared fixtures for deutero-cli tests."""

from __future__ import annotations

import json
import os
from unittest.mock import patch

import pytest
from click.testing import CliRunner

from deutero_cli.cli import cli
from deutero_cli.client import DeuteroClient


TEST_API_KEY = "test-api-key-0123456789abcdef"
TEST_BASE_URL = "http://testserver"

SAMPLE_SURVEY_ID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
SAMPLE_INTERVIEW_ID = "f0e1d2c3-b4a5-6789-0abc-def123456789"
SAMPLE_PERSONA_ID = "11223344-5566-7788-99aa-bbccddeeff00"


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def invoke(runner):
    """Return a helper that invokes the CLI with test credentials pre-set."""

    def _invoke(args, **kwargs):
        return runner.invoke(
            cli,
            ["--api-key", TEST_API_KEY, "--base-url", TEST_BASE_URL] + list(args),
            catch_exceptions=False,
            **kwargs,
        )

    return _invoke


@pytest.fixture
def client():
    return DeuteroClient(base_url=TEST_BASE_URL, api_key=TEST_API_KEY)


# ── Canonical API response fixtures ─────────────────────────────────

@pytest.fixture
def survey_generation_response():
    return {
        "study_id": SAMPLE_SURVEY_ID,
        "study_name": "UX Study: Onboarding Flow",
        "study_description": "A study about onboarding.",
        "research_questions": [{"id": "1", "question": "How do users onboard?"}],
        "research_objectives": [{"id": "1", "objective": "Understand onboarding pain points"}],
        "xml_file": '<?xml version="1.0"?><surveyExport><survey><alias>Test</alias></survey></surveyExport>',
        "url": "http://127.0.0.1:5000/survey-details?survey_id=" + SAMPLE_SURVEY_ID,
        "agent_instructions": "Create a .xml file.",
    }


@pytest.fixture
def participation_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "total_interviews": 10,
        "completed_interviews": 7,
        "incomplete_interviews": 3,
        "max_responses": 20,
        "completion_rate": 70.0,
        "quota_fill_rate": 35.0,
        "quota_remaining": 13,
    }


@pytest.fixture
def agent_requirements_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "markdown": "# User Requirements\n\n- Requirement 1\n- Requirement 2\n",
        "filename": "test-study_user_requirements.md",
    }


@pytest.fixture
def question_generation_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "question_list": [
            {
                "question_number": 1,
                "question_type": "text",
                "question_content": "Tell me about your experience.",
                "question_scale": None,
                "question_options": None,
                "question_slots": None,
                "question_images": None,
                "question_follow_up": True,
                "question_expected_image": None,
            },
            {
                "question_number": 2,
                "question_type": "scale",
                "question_content": "Rate your satisfaction.",
                "question_scale": '{"minScale": 1, "maxScale": 5, "minLabel": "Low", "maxLabel": "High"}',
                "question_options": None,
                "question_slots": None,
                "question_images": None,
                "question_follow_up": False,
                "question_expected_image": None,
            },
        ],
        "edit_questions_url": f"http://127.0.0.1:5000/edit-survey?survey_id={SAMPLE_SURVEY_ID}",
        "interview_url": f"http://127.0.0.1:5000/?survey_id={SAMPLE_SURVEY_ID}",
        "xml_file": "<surveyExport/>",
        "agent_instructions": "Create or update a .xml file.",
    }


@pytest.fixture
def persona_generation_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "personas": [
            {"persona": "Sarah, 32, product manager at a mid-size SaaS company.", "persona_id": SAMPLE_PERSONA_ID},
            {"persona": "James, 45, enterprise IT director.", "persona_id": "22334455-6677-8899-aabb-ccddeeff0011"},
        ],
    }


@pytest.fixture
def interview_simulation_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "persona_id": SAMPLE_PERSONA_ID,
        "transcript_url": f"/interviews/?survey_id={SAMPLE_SURVEY_ID}&interview_id={SAMPLE_INTERVIEW_ID}",
    }


@pytest.fixture
def analysis_run_response():
    return {
        "success": True,
        "interviews_queued": 2,
        "interviews": [
            {"interview_id": SAMPLE_INTERVIEW_ID, "survey_id": SAMPLE_SURVEY_ID},
        ],
        "message": "Successfully queued phases 1-4 analysis for 2 interview(s)",
        "cross_case_xml": None,
    }


@pytest.fixture
def analysis_run_cross_case_response():
    return {
        "success": True,
        "interviews_queued": 0,
        "interviews": [],
        "message": "Cross-case analysis completed for survey with 5 interviews",
        "cross_case_xml": "<crossCaseAnalysis><themes/></crossCaseAnalysis>",
    }


@pytest.fixture
def analysis_status_response():
    return {
        "interview_id": SAMPLE_INTERVIEW_ID,
        "survey_id": None,
        "total_interviews": 1,
        "interviews": [
            {
                "interview_id": SAMPLE_INTERVIEW_ID,
                "survey_id": SAMPLE_SURVEY_ID,
                "analysis_id": "aaa-bbb-ccc",
                "status": "in_progress",
                "phases_completed": 2,
                "total_phases": 4,
                "phases": [
                    {"phase": 1, "completed": True, "completed_at": "2025-01-01T00:00:00", "has_output": True},
                    {"phase": 2, "completed": True, "completed_at": "2025-01-01T00:01:00", "has_output": True},
                    {"phase": 3, "completed": False, "completed_at": None, "has_output": False},
                    {"phase": 4, "completed": False, "completed_at": None, "has_output": False},
                ],
                "runs": [],
                "error_message": None,
                "created_at": "2025-01-01T00:00:00",
                "updated_at": "2025-01-01T00:01:00",
            }
        ],
    }


@pytest.fixture
def interview_results_response():
    return {
        "interview_id": SAMPLE_INTERVIEW_ID,
        "phase": "initial_engagement",
        "xml_output": "<phase1><summary>Initial engagement output</summary></phase1>",
        "completed_at": "2025-01-01T00:00:00",
        "error": None,
    }


@pytest.fixture
def survey_results_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "xml_output": "<crossCaseAnalysis><themes/></crossCaseAnalysis>",
        "completed_at": "2025-01-01T12:00:00",
        "error": None,
    }
