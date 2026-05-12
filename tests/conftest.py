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
SAMPLE_QUESTION_ID = "01020304-0506-0708-090a-0b0c0d0e0f10"
SAMPLE_IMAGE_ID = "11111111-2222-3333-4444-555555555555"
SAMPLE_PROJECT_ID = "99999999-8888-7777-6666-555555555555"


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


@pytest.fixture
def model_tier_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "model_tier": "premium",
        "model_id": "claude-sonnet",
        "model_provider": "anthropic",
    }


@pytest.fixture
def suggest_from_site_response():
    return {
        "is_valid": True,
        "rationale": "The site contains enough product context.",
        "study_data": {"study_name": "Landing Page UX Study"},
        "run_id": "run-123",
        "run_index": 1,
        "total_runs": 1,
    }


@pytest.fixture
def question_response():
    return {
        "id": SAMPLE_QUESTION_ID,
        "question": "What is your biggest onboarding challenge?",
        "explanation": "Probe for specifics.",
        "scale": None,
        "options": None,
        "slots": None,
        "follow_up": True,
        "min_turns": 1,
        "max_turns": 3,
        "expected_image": None,
    }


@pytest.fixture
def question_create_response(question_response):
    return {
        **question_response,
        "type": "text",
        "question_list_id": "question-list-123",
    }


@pytest.fixture
def question_image_response():
    return {
        "id": SAMPLE_IMAGE_ID,
        "question_id": SAMPLE_QUESTION_ID,
        "survey_id": SAMPLE_SURVEY_ID,
        "gcs_url": "gs://bucket/image.png",
        "label": "Concept A",
        "position": 1,
    }


@pytest.fixture
def credit_estimation_response():
    return {
        "estimated_credits": 12,
        "credits_per_interview": 2,
        "credits_for_analysis": 4,
        "model_tier": "premium",
        "num_participants": 4,
        "num_questions": 5,
    }


@pytest.fixture
def credit_balance_response():
    return {
        "available_credits": 100,
        "credits_used": 20,
        "credits_reserved": 5,
        "base_limit": 100,
        "purchased_credits": 10,
        "rollover_credits": 0,
        "is_trial_active": False,
        "trial_credits_used": 0,
        "net_available": 75,
    }


@pytest.fixture
def project_list_response():
    return {
        "projects": [
            {"id": SAMPLE_PROJECT_ID, "name": "UX Research Q1 2025", "description": "Quarterly UX research initiative."},
            {"id": "88888888-7777-6666-5555-444444444444", "name": "Customer Discovery", "description": None},
        ],
    }


@pytest.fixture
def project_surveys_response():
    return {
        "project_id": SAMPLE_PROJECT_ID,
        "surveys": [
            {"id": SAMPLE_SURVEY_ID, "alias": "Onboarding Study", "description": "Study about onboarding.", "survey_type": "user_experience", "language": "English", "date_created": "2025-01-15T10:00:00", "max_responses": 20},
            {"id": "77777777-6666-5555-4444-333333333333", "alias": "Feature Prioritization", "description": None, "survey_type": "user_experience", "language": "English", "date_created": "2025-02-01T10:00:00", "max_responses": None},
        ],
    }


@pytest.fixture
def question_list_response():
    return {
        "id": "question-list-123",
        "survey_id": SAMPLE_SURVEY_ID,
        "questions": [
            {
                "id": SAMPLE_QUESTION_ID,
                "question_number": 1,
                "question": "Tell me about your experience.",
                "type": "text",
                "explanation": None,
                "scale": None,
                "options": None,
                "slots": None,
                "follow_up": True,
                "min_turns": None,
                "max_turns": None,
                "expected_image": None,
                "images": None,
            },
            {
                "id": "02030405-0607-0809-0a0b-0c0d0e0f1011",
                "question_number": 2,
                "question": "Rate your satisfaction.",
                "type": "scale",
                "explanation": "Use the 1-5 scale.",
                "scale": {"minScale": 1, "maxScale": 5, "minLabel": "Low", "maxLabel": "High"},
                "options": None,
                "slots": None,
                "follow_up": False,
                "min_turns": None,
                "max_turns": None,
                "expected_image": None,
                "images": None,
            },
        ],
    }


@pytest.fixture
def survey_interviews_response():
    return {
        "survey_id": SAMPLE_SURVEY_ID,
        "total": 2,
        "interviews": [
            {"id": SAMPLE_INTERVIEW_ID, "participant_name": "Sarah", "start_time": "2025-01-20T09:00:00", "end_time": "2025-01-20T09:25:00", "completed": True, "simulated": True, "termination_reason": None},
            {"id": "eeeeeeee-dddd-cccc-bbbb-aaaaaaaaaaaa", "participant_name": "James", "start_time": "2025-01-21T10:00:00", "end_time": None, "completed": False, "simulated": True, "termination_reason": "timeout"},
        ],
    }


@pytest.fixture
def interview_transcript_response():
    return {
        "interview_id": SAMPLE_INTERVIEW_ID,
        "total": 3,
        "messages": [
            {"id": "msg-1", "role": "assistant", "content": "Welcome! Tell me about your experience.", "timestamp": "2025-01-20T09:00:00", "question_id": SAMPLE_QUESTION_ID, "question_number": 1},
            {"id": "msg-2", "role": "user", "content": "It was great, very intuitive.", "timestamp": "2025-01-20T09:00:30", "question_id": None, "question_number": None},
            {"id": "msg-3", "role": "assistant", "content": "Can you elaborate on what made it intuitive?", "timestamp": "2025-01-20T09:01:00", "question_id": SAMPLE_QUESTION_ID, "question_number": 1},
        ],
    }
