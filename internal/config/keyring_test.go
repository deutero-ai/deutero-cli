// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

// Config.SaveTokens/Load/ClearTokens integration with the OS keyring layer
// added in internal/cliutil/keyring.go: keyring-available should keep the
// secrets out of credentials.toml and reconstruct them on the next Load;
// keyring-unavailable (mirroring "a keychain cannot be found to store the
// credential" and a headless Linux box with no Secret Service session) must
// fall back to the pre-existing file-only behavior with no loss of function.
package config

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deutero-ai/deutero-cli/internal/cliutil"
	"github.com/zalando/go-keyring"
)

func TestSaveTokensKeyringAvailable(t *testing.T) {
	isolateCredentials(t)
	keyring.MockInit()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	if err := cfg.SaveTokens("client-1", "secret-1", "access-1", "refresh-1", expiry); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	if cfg.TokenStorage != "keyring" {
		t.Fatalf("TokenStorage = %q, want keyring", cfg.TokenStorage)
	}

	credsPath, err := cliutil.CredentialsFilePath()
	if err != nil {
		t.Fatalf("CredentialsFilePath: %v", err)
	}
	raw, err := os.ReadFile(credsPath)
	if err != nil {
		t.Fatalf("reading credentials file: %v", err)
	}
	body := string(raw)
	for _, secret := range []string{"access-1", "refresh-1", "secret-1"} {
		if strings.Contains(body, secret) {
			t.Fatalf("credentials.toml unexpectedly contains secret %q when keyring storage succeeded:\n%s", secret, body)
		}
	}
	if !strings.Contains(body, "client-1") {
		t.Fatalf("credentials.toml should still carry the non-secret client_id:\n%s", body)
	}

	// A fresh Load (new process, in effect) must reconstruct the secrets from
	// the keyring even though the file doesn't carry them.
	reloaded, err := Load("")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if reloaded.AccessToken != "access-1" {
		t.Fatalf("reloaded AccessToken = %q, want access-1", reloaded.AccessToken)
	}
	if reloaded.RefreshToken != "refresh-1" {
		t.Fatalf("reloaded RefreshToken = %q, want refresh-1", reloaded.RefreshToken)
	}
	if reloaded.ClientSecret != "secret-1" {
		t.Fatalf("reloaded ClientSecret = %q, want secret-1", reloaded.ClientSecret)
	}
	if reloaded.TokenStorage != "keyring" {
		t.Fatalf("reloaded TokenStorage = %q, want keyring", reloaded.TokenStorage)
	}
	if got := reloaded.AuthHeader(); got != "Bearer access-1" {
		t.Fatalf("AuthHeader() = %q, want Bearer access-1", got)
	}

	// Logout must remove the keyring entry, not just the file.
	if err := reloaded.ClearTokens(); err != nil {
		t.Fatalf("ClearTokens: %v", err)
	}
	if _, ok := cliutil.LoadKeyringTokens(credsPath); ok {
		t.Fatal("expected keyring entry to be gone after ClearTokens")
	}
	loggedOut, err := Load("")
	if err != nil {
		t.Fatalf("reloading after logout: %v", err)
	}
	if loggedOut.AuthHeader() != "" {
		t.Fatalf("AuthHeader() after logout = %q, want empty", loggedOut.AuthHeader())
	}
}

func TestSaveTokensKeyringUnavailableFallsBackToFile(t *testing.T) {
	isolateCredentials(t)
	keyring.MockInitWithError(errors.New("a keychain cannot be found to store the credential"))

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	if err := cfg.SaveTokens("client-2", "secret-2", "access-2", "refresh-2", expiry); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
	if cfg.TokenStorage != "file" {
		t.Fatalf("TokenStorage = %q, want file (keyring unavailable)", cfg.TokenStorage)
	}

	credsPath, err := cliutil.CredentialsFilePath()
	if err != nil {
		t.Fatalf("CredentialsFilePath: %v", err)
	}
	raw, err := os.ReadFile(credsPath)
	if err != nil {
		t.Fatalf("reading credentials file: %v", err)
	}
	body := string(raw)
	for _, secret := range []string{"access-2", "refresh-2", "secret-2"} {
		if !strings.Contains(body, secret) {
			t.Fatalf("credentials.toml should carry secret %q when the keyring is unavailable:\n%s", secret, body)
		}
	}

	reloaded, err := Load("")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if got := reloaded.AuthHeader(); got != "Bearer access-2" {
		t.Fatalf("AuthHeader() = %q, want Bearer access-2", got)
	}
	if reloaded.TokenStorage != "" {
		t.Fatalf("reloaded TokenStorage = %q, want empty (no keyring entry to find)", reloaded.TokenStorage)
	}

	if err := reloaded.ClearTokens(); err != nil {
		t.Fatalf("ClearTokens: %v", err)
	}
	loggedOut, err := Load("")
	if err != nil {
		t.Fatalf("reloading after logout: %v", err)
	}
	if loggedOut.AuthHeader() != "" {
		t.Fatalf("AuthHeader() after logout = %q, want empty", loggedOut.AuthHeader())
	}
}
