"""HTTP client wrapper for the Deutero API."""

from __future__ import annotations

from typing import Any, Dict, Optional

import httpx

from deutero_cli.config import get_api_key, get_base_url


class DeuteroAPIError(Exception):
    """Raised when the API returns a non-2xx response."""

    def __init__(self, status_code: int, detail: str):
        self.status_code = status_code
        self.detail = detail
        super().__init__(f"HTTP {status_code}: {detail}")


class DeuteroClient:
    """Thin synchronous wrapper around the Deutero Interview Admin API."""

    def __init__(
        self,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None,
        timeout: float = 120.0,
    ):
        self.base_url = (base_url or get_base_url()).rstrip("/")
        self.api_key = api_key or get_api_key()
        self.timeout = timeout

    def _headers(self) -> Dict[str, str]:
        headers: Dict[str, str] = {"Content-Type": "application/json"}
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        return headers

    def _url(self, path: str) -> str:
        return f"{self.base_url}/api/v1{path}"

    def _handle_response(self, resp: httpx.Response) -> Any:
        if resp.status_code >= 400:
            try:
                detail = resp.json().get("detail", resp.text)
            except Exception:
                detail = resp.text
            raise DeuteroAPIError(resp.status_code, detail)
        return resp.json()

    def post(self, path: str, json_body: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.post(self._url(path), json=json_body or {}, headers=self._headers())
        return self._handle_response(resp)

    def get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.get(self._url(path), params=params, headers=self._headers())
        return self._handle_response(resp)

    # ── Survey endpoints ─────────────────────────────────────────────

    def generate_survey(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/surveys/generate", payload)

    def get_participation(self, survey_id: str) -> Dict[str, Any]:
        return self.get(f"/surveys/{survey_id}/participation")

    def get_agent_requirements(self, survey_id: str) -> Dict[str, Any]:
        return self.post(f"/surveys/{survey_id}/agent-requirements")

    # ── Question endpoints ───────────────────────────────────────────

    def generate_questions(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/questions/generate", payload)

    # ── Persona endpoints ────────────────────────────────────────────

    def generate_personas(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/personas/generate", payload)

    # ── Interview endpoints ──────────────────────────────────────────

    def simulate_interview(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/interviews/simulate", payload)

    # ── Analysis endpoints ───────────────────────────────────────────

    def run_analysis(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/analysis/run", payload)

    def get_analysis_status(self, params: Dict[str, Any]) -> Dict[str, Any]:
        return self.get("/analysis/status", params)

    def get_interview_analysis_results(self, interview_id: str, phase: str) -> Dict[str, Any]:
        return self.get("/analysis/results/interview", {"interview_id": interview_id, "phase": phase})

    def get_survey_analysis_results(self, survey_id: str) -> Dict[str, Any]:
        return self.get("/analysis/results/survey", {"survey_id": survey_id})
