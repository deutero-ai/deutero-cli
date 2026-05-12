"""HTTP client wrapper for the Deutero API."""

from __future__ import annotations

from typing import Any, Dict, Optional

import httpx

from deutero_cli.auth import get_access_token
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
        token = get_access_token()
        if token:
            headers["Authorization"] = f"Bearer {token}"
        elif self.api_key:
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
        if not resp.content:
            return {}
        return resp.json()

    def post(self, path: str, json_body: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.post(self._url(path), json=json_body or {}, headers=self._headers())
        return self._handle_response(resp)

    def get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.get(self._url(path), params=params, headers=self._headers())
        return self._handle_response(resp)

    def put(self, path: str, json_body: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.put(self._url(path), json=json_body or {}, headers=self._headers())
        return self._handle_response(resp)

    def delete(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        with httpx.Client(timeout=self.timeout) as http:
            resp = http.delete(self._url(path), params=params, headers=self._headers())
        return self._handle_response(resp)

    # ── Survey endpoints ─────────────────────────────────────────────

    def generate_survey(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/surveys/generate", payload)

    def get_participation(self, survey_id: str) -> Dict[str, Any]:
        return self.get(f"/surveys/{survey_id}/participation")

    def get_agent_requirements(self, survey_id: str) -> Dict[str, Any]:
        return self.post(f"/surveys/{survey_id}/agent-requirements")

    def get_survey_model_tier(self, survey_id: str) -> Dict[str, Any]:
        return self.get(f"/surveys/{survey_id}/model-tier")

    def set_survey_model_tier(self, survey_id: str, model_tier: str) -> Dict[str, Any]:
        return self.put(f"/surveys/{survey_id}/model-tier", {"model_tier": model_tier})

    def suggest_from_site(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/suggest-from-site", payload)

    # ── Question endpoints ───────────────────────────────────────────

    def generate_questions(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/questions/generate", payload)

    def create_question(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/questions", payload)

    def get_question(self, question_id: str) -> Dict[str, Any]:
        return self.get(f"/questions/{question_id}")

    def update_question(self, question_id: str, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.put(f"/questions/{question_id}", payload)

    def delete_question(self, question_id: str, survey_id: str) -> Dict[str, Any]:
        return self.delete(f"/questions/{question_id}", {"survey_id": survey_id})

    def reorder_questions(self, survey_id: str, question_ids: list[str]) -> Dict[str, Any]:
        return self.put(f"/surveys/{survey_id}/questions/reorder", {"question_ids": question_ids})

    def upload_question_image(self, question_id: str, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post(f"/questions/{question_id}/images", payload)

    def reorder_question_images(self, question_id: str, image_ids: list[str]) -> Dict[str, Any]:
        return self.put(f"/questions/{question_id}/images/reorder", {"image_ids": image_ids})

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

    def estimate_simulation_credits(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/credits/estimate/simulation", payload)

    def estimate_analysis_credits(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/credits/estimate/analysis", payload)

    def estimate_survey_credits(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return self.post("/credits/estimate/survey", payload)

    def get_credit_balance(self) -> Dict[str, Any]:
        return self.get("/credits/balance")

    # ── Project endpoints ────────────────────────────────────────────

    def list_projects(self) -> Dict[str, Any]:
        return self.get("/projects")

    def list_project_surveys(self, project_id: str) -> Dict[str, Any]:
        return self.get(f"/projects/{project_id}/surveys")

    # ── Survey detail endpoints ──────────────────────────────────────

    def get_question_list(self, survey_id: str) -> Dict[str, Any]:
        return self.get(f"/surveys/{survey_id}/question-list")

    def list_survey_interviews(self, survey_id: str) -> Dict[str, Any]:
        return self.get(f"/surveys/{survey_id}/interviews")

    # ── Interview detail endpoints ───────────────────────────────────

    def get_interview_transcript(self, interview_id: str) -> Dict[str, Any]:
        return self.get(f"/interviews/{interview_id}/transcript")
