"""Stytch Connected Apps OAuth2 + PKCE authentication for the Deutero CLI."""

from __future__ import annotations

import base64
import hashlib
import json
import secrets
import socket
import time
import webbrowser
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Thread
from typing import Optional
from urllib.parse import urlencode

import httpx

from deutero_cli.config import CONFIG_DIR, CONFIG_FILE, _load_config_file, _save_config_file

AUTHORIZE_URL = "https://dashboard.deutero.ai/connect/authorize"
STYTCH_TOKEN_URL_TEMPLATE = "https://api.stytch.com/v1/public/{project_id}/oauth2/token"


def _generate_code_verifier() -> str:
    """Generate a random 32-byte PKCE code_verifier (base64url, no padding)."""
    return base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b"=").decode("ascii")


def _generate_code_challenge(verifier: str) -> str:
    """Compute S256 code_challenge from verifier."""
    digest = hashlib.sha256(verifier.encode("ascii")).digest()
    return base64.urlsafe_b64encode(digest).rstrip(b"=").decode("ascii")


def _get_free_port() -> int:
    """Find a free TCP port on localhost."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("dashboard.deutero.ai", 0))
        return s.getsockname()[1]


class _CallbackHandler(BaseHTTPRequestHandler):
    """HTTP handler that captures the OAuth callback code."""

    code: Optional[str] = None
    error: Optional[str] = None

    def do_GET(self) -> None:
        from urllib.parse import parse_qs, urlparse

        query = parse_qs(urlparse(self.path).query)
        code = query.get("code", [None])[0]
        error = query.get("error", [None])[0]

        if code:
            _CallbackHandler.code = code
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.end_headers()
            self.wfile.write(
                b"<html><body><h1>Authentication successful!</h1>"
                b"<p>You can close this window and return to the CLI.</p></body></html>"
            )
        else:
            _CallbackHandler.error = error or "no code received"
            self.send_response(400)
            self.send_header("Content-Type", "text/html")
            self.end_headers()
            self.wfile.write(
                b"<html><body><h1>Authentication failed.</h1>"
                b"<p>Please try again.</p></body></html>"
            )

    def log_message(self, format: str, *args: object) -> None:
        pass


def _exchange_code_for_token(
    client_id: str,
    project_id: str,
    code: str,
    code_verifier: str,
    redirect_uri: str,
) -> dict:
    """Exchange authorization code for access + refresh tokens via Stytch token endpoint."""
    token_url = STYTCH_TOKEN_URL_TEMPLATE.format(project_id=project_id)
    body = {
        "client_id": client_id,
        "code": code,
        "code_verifier": code_verifier,
        "grant_type": "authorization_code",
        "redirect_uri": redirect_uri,
    }
    with httpx.Client(timeout=30.0) as http:
        resp = http.post(token_url, json=body)
    if resp.status_code != 200:
        raise RuntimeError(f"Token exchange failed (HTTP {resp.status_code}): {resp.text}")
    return resp.json()


def _refresh_access_token(client_id: str, project_id: str, refresh_token: str) -> dict:
    """Use a refresh_token to obtain a new access_token."""
    token_url = STYTCH_TOKEN_URL_TEMPLATE.format(project_id=project_id)
    body = {
        "client_id": client_id,
        "refresh_token": refresh_token,
        "grant_type": "refresh_token",
    }
    with httpx.Client(timeout=30.0) as http:
        resp = http.post(token_url, json=body)
    if resp.status_code != 200:
        raise RuntimeError(f"Token refresh failed (HTTP {resp.status_code}): {resp.text}")
    return resp.json()


def login(client_id: str, project_id: str) -> dict:
    """Run the full OAuth2 + PKCE login flow: open browser, wait for callback, exchange code."""
    _CallbackHandler.code = None
    _CallbackHandler.error = None

    port = _get_free_port()
    redirect_uri = f"http://dashboard.deutero.ai:{port}/callback"

    code_verifier = _generate_code_verifier()
    code_challenge = _generate_code_challenge(code_verifier)

    params = urlencode({
        "client_id": client_id,
        "redirect_uri": redirect_uri,
        "response_type": "code",
        "code_challenge": code_challenge,
        "code_challenge_method": "S256",
        "scope": "offline_access",
    })
    auth_url = f"{AUTHORIZE_URL}?{params}"

    server = HTTPServer(("dashboard.deutero.ai", port), _CallbackHandler)
    server_thread = Thread(target=server.serve_forever, daemon=True)
    server_thread.start()

    try:
        opened = webbrowser.open(auth_url)
        if not opened:
            raise RuntimeError(f"Could not open browser. Please visit:\n{auth_url}")

        deadline = time.time() + 300  # 5 minute timeout
        while time.time() < deadline:
            if _CallbackHandler.code or _CallbackHandler.error:
                break
            time.sleep(0.2)

        if _CallbackHandler.error:
            raise RuntimeError(f"Authorization failed: {_CallbackHandler.error}")
        if not _CallbackHandler.code:
            raise RuntimeError("Timeout waiting for authorization (5 minutes).")

        token_data = _exchange_code_for_token(
            client_id=client_id,
            project_id=project_id,
            code=_CallbackHandler.code,
            code_verifier=code_verifier,
            redirect_uri=redirect_uri,
        )
    finally:
        server.shutdown()

    save_tokens(
        access_token=token_data.get("access_token", ""),
        refresh_token=token_data.get("refresh_token", ""),
        expires_in=token_data.get("expires_in", 3600),
    )
    save_oauth_config(client_id=client_id, project_id=project_id)

    return token_data


def save_tokens(access_token: str, refresh_token: str, expires_in: int) -> None:
    """Persist tokens to the config file."""
    data = _load_config_file()
    data["access_token"] = access_token
    data["refresh_token"] = refresh_token
    data["token_expiry"] = int(time.time()) + expires_in
    _save_config_file(data)


def save_oauth_config(client_id: str, project_id: str) -> None:
    """Persist OAuth client_id and project_id."""
    data = _load_config_file()
    data["client_id"] = client_id
    data["project_id"] = project_id
    _save_config_file(data)


def get_access_token() -> Optional[str]:
    """Return a valid access token, refreshing if expired. Returns None if not logged in."""
    data = _load_config_file()
    access_token = data.get("access_token")
    if not access_token:
        return None

    expiry = data.get("token_expiry", 0)
    if time.time() < expiry - 60:
        return access_token

    refresh_token = data.get("refresh_token")
    client_id = data.get("client_id")
    project_id = data.get("project_id")
    if not (refresh_token and client_id and project_id):
        return None

    try:
        token_data = _refresh_access_token(client_id, project_id, refresh_token)
        save_tokens(
            access_token=token_data.get("access_token", ""),
            refresh_token=token_data.get("refresh_token", refresh_token),
            expires_in=token_data.get("expires_in", 3600),
        )
        return token_data.get("access_token")
    except RuntimeError:
        return None


def clear_tokens() -> None:
    """Remove stored tokens (logout)."""
    data = _load_config_file()
    for key in ("access_token", "refresh_token", "token_expiry"):
        data.pop(key, None)
    _save_config_file(data)


def get_auth_status() -> dict:
    """Return current auth status info."""
    data = _load_config_file()
    access_token = data.get("access_token")
    expiry = data.get("token_expiry", 0)
    client_id = data.get("client_id")
    project_id = data.get("project_id")

    logged_in = bool(access_token)
    expired = time.time() >= expiry if expiry else True
    has_refresh = bool(data.get("refresh_token"))

    return {
        "logged_in": logged_in,
        "expired": expired if logged_in else None,
        "has_refresh_token": has_refresh,
        "client_id": client_id,
        "project_id": project_id,
    }
