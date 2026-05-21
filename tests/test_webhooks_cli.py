"""CLI integration tests for webhook commands."""

from __future__ import annotations

import json
from unittest.mock import patch

from tests.conftest import SAMPLE_ENDPOINT_ID


class TestWebhooksList:
    def test_list_success(self, invoke, webhook_list_response):
        with patch("deutero_cli.client.DeuteroClient.list_webhook_endpoints", return_value=webhook_list_response):
            result = invoke(["webhooks", "list"])
            assert result.exit_code == 0
            assert "2 webhook endpoint(s)" in result.output
            assert SAMPLE_ENDPOINT_ID in result.output
            assert "bbbbcccc-dddd-eeee-ffff-000011112222" in result.output

    def test_list_shows_event_types(self, invoke, webhook_list_response):
        with patch("deutero_cli.client.DeuteroClient.list_webhook_endpoints", return_value=webhook_list_response):
            result = invoke(["webhooks", "list"])
            assert result.exit_code == 0
            assert "interview.completed" in result.output

    def test_list_empty(self, invoke):
        empty = {"endpoints": [], "total": 0, "event_types": ["interview.completed"]}
        with patch("deutero_cli.client.DeuteroClient.list_webhook_endpoints", return_value=empty):
            result = invoke(["webhooks", "list"])
            assert result.exit_code == 0
            assert "0 webhook endpoint(s)" in result.output

    def test_list_with_output_file(self, invoke, webhook_list_response, tmp_path):
        out = tmp_path / "webhooks.json"
        with patch("deutero_cli.client.DeuteroClient.list_webhook_endpoints", return_value=webhook_list_response):
            result = invoke(["webhooks", "list", "--output", str(out)])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["total"] == 2

    def test_list_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.list_webhook_endpoints", side_effect=DeuteroAPIError(401, "Unauthorized")):
            result = invoke(["webhooks", "list"])
            assert result.exit_code == 1


class TestWebhooksUpdate:
    def test_update_label(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "label": "New Label"}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--label", "New Label"])
            assert result.exit_code == 0
            assert "New Label" in result.output
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"label": "New Label"})

    def test_update_url(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "url": "https://new.example.com/hook"}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--url", "https://new.example.com/hook"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"url": "https://new.example.com/hook"})

    def test_update_enable(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "enabled": True}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--enable"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"enabled": True})

    def test_update_disable(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "enabled": False}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--disable"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"enabled": False})

    def test_update_enable_wins_over_disable_when_both_passed(self, invoke, webhook_endpoint_response):
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=webhook_endpoint_response) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--enable", "--disable"])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert called_payload["enabled"] is True

    def test_update_events(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "events": ["interview.completed"]}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--event", "interview.completed"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"events": ["interview.completed"]})

    def test_update_multiple_events(self, invoke, webhook_endpoint_response):
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=webhook_endpoint_response) as mock:
            result = invoke([
                "webhooks", "update", SAMPLE_ENDPOINT_ID,
                "--event", "interview.completed",
                "--event", "survey.completed",
            ])
            assert result.exit_code == 0
            called_payload = mock.call_args[0][1]
            assert set(called_payload["events"]) == {"interview.completed", "survey.completed"}

    def test_update_multiple_fields(self, invoke, webhook_endpoint_response):
        updated = {**webhook_endpoint_response, "label": "Updated", "enabled": False}
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=updated) as mock:
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--label", "Updated", "--disable"])
            assert result.exit_code == 0
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID, {"label": "Updated", "enabled": False})

    def test_update_no_fields_error(self, invoke):
        result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID])
        assert result.exit_code == 1

    def test_update_with_output_file(self, invoke, webhook_endpoint_response, tmp_path):
        out = tmp_path / "endpoint.json"
        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", return_value=webhook_endpoint_response):
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--label", "Test", "--output", str(out)])
            assert result.exit_code == 0
            data = json.loads(out.read_text())
            assert data["id"] == SAMPLE_ENDPOINT_ID

    def test_update_api_error(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", side_effect=DeuteroAPIError(404, "Webhook endpoint not found.")):
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--label", "X"])
            assert result.exit_code == 1

    def test_update_api_422(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.update_webhook_endpoint", side_effect=DeuteroAPIError(422, "url must be a valid http or https URL.")):
            result = invoke(["webhooks", "update", SAMPLE_ENDPOINT_ID, "--url", "not-a-url"])
            assert result.exit_code == 1


class TestWebhooksDelete:
    def test_delete_with_yes_flag(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.delete_webhook_endpoint", return_value={}) as mock:
            result = invoke(["webhooks", "delete", SAMPLE_ENDPOINT_ID, "--yes"])
            assert result.exit_code == 0
            assert "deleted" in result.output
            mock.assert_called_once_with(SAMPLE_ENDPOINT_ID)

    def test_delete_with_short_yes_flag(self, invoke):
        with patch("deutero_cli.client.DeuteroClient.delete_webhook_endpoint", return_value={}):
            result = invoke(["webhooks", "delete", SAMPLE_ENDPOINT_ID, "-y"])
            assert result.exit_code == 0
            assert "deleted" in result.output

    def test_delete_confirmation_accepted(self, runner):
        from deutero_cli.cli import cli
        from tests.conftest import TEST_API_KEY, TEST_BASE_URL

        with patch("deutero_cli.client.DeuteroClient.delete_webhook_endpoint", return_value={}):
            result = runner.invoke(
                cli,
                ["--api-key", TEST_API_KEY, "--base-url", TEST_BASE_URL, "webhooks", "delete", SAMPLE_ENDPOINT_ID],
                input="y\n",
                catch_exceptions=False,
            )
            assert result.exit_code == 0
            assert "deleted" in result.output

    def test_delete_confirmation_aborted(self, runner):
        from deutero_cli.cli import cli
        from tests.conftest import TEST_API_KEY, TEST_BASE_URL

        with patch("deutero_cli.client.DeuteroClient.delete_webhook_endpoint", return_value={}):
            result = runner.invoke(
                cli,
                ["--api-key", TEST_API_KEY, "--base-url", TEST_BASE_URL, "webhooks", "delete", SAMPLE_ENDPOINT_ID],
                input="n\n",
                catch_exceptions=False,
            )
            assert result.exit_code != 0

    def test_delete_not_found(self, invoke):
        from deutero_cli.client import DeuteroAPIError

        with patch("deutero_cli.client.DeuteroClient.delete_webhook_endpoint", side_effect=DeuteroAPIError(404, "Webhook endpoint not found.")):
            result = invoke(["webhooks", "delete", SAMPLE_ENDPOINT_ID, "--yes"])
            assert result.exit_code == 1
