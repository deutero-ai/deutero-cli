"""CLI integration tests for project commands."""

from __future__ import annotations

from unittest.mock import patch

from tests.conftest import SAMPLE_PROJECT_ID, SAMPLE_SURVEY_ID


class TestProjectsList:
    def test_list_projects(self, invoke, project_list_response):
        with patch("deutero_cli.client.DeuteroClient.list_projects", return_value=project_list_response):
            result = invoke(["projects", "list"])
            assert result.exit_code == 0
            assert "2 project(s)" in result.output

    def test_list_projects_empty(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.list_projects", return_value={"projects": []}):
            result = invoke(["projects", "list"])
            assert result.exit_code == 0
            assert "0 project(s)" in result.output

    def test_list_projects_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.list_projects", side_effect=DeuteroAPIError(401, "Unauthorized")):
            result = invoke(["projects", "list"])
            assert result.exit_code == 1


class TestProjectsListSurveys:
    def test_list_surveys_for_project(self, invoke, project_surveys_response):
        with patch("deutero_cli.client.DeuteroClient.list_project_surveys", return_value=project_surveys_response) as mock:
            result = invoke(["projects", "list-surveys", SAMPLE_PROJECT_ID])
            assert result.exit_code == 0
            assert "2 survey(s)" in result.output
            mock.assert_called_once_with(SAMPLE_PROJECT_ID)

    def test_list_surveys_empty(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.list_project_surveys", return_value={"project_id": SAMPLE_PROJECT_ID, "surveys": []}):
            result = invoke(["projects", "list-surveys", SAMPLE_PROJECT_ID])
            assert result.exit_code == 0
            assert "0 survey(s)" in result.output

    def test_list_surveys_not_found(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.list_project_surveys", side_effect=DeuteroAPIError(404, "Project not found")):
            result = invoke(["projects", "list-surveys", SAMPLE_PROJECT_ID])
            assert result.exit_code == 1
