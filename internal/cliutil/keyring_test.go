// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

package cliutil

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeyringTokensRoundTrip(t *testing.T) {
	keyring.MockInit()

	const account = "/home/user/.local/share/deutero-pp-cli/credentials.toml"

	if _, ok := LoadKeyringTokens(account); ok {
		t.Fatal("expected no tokens before Save")
	}

	tok := KeyringTokens{AccessToken: "at-1", RefreshToken: "rt-1", ClientSecret: "cs-1"}
	if !SaveKeyringTokens(account, tok) {
		t.Fatal("SaveKeyringTokens returned false against the mock backend")
	}

	got, ok := LoadKeyringTokens(account)
	if !ok {
		t.Fatal("expected tokens after Save")
	}
	if got != tok {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, tok)
	}

	DeleteKeyringTokens(account)
	if _, ok := LoadKeyringTokens(account); ok {
		t.Fatal("expected no tokens after Delete")
	}
}

func TestKeyringTokensEmptyAccountOrValue(t *testing.T) {
	keyring.MockInit()

	if SaveKeyringTokens("", KeyringTokens{AccessToken: "at"}) {
		t.Fatal("expected SaveKeyringTokens to refuse an empty account")
	}
	if SaveKeyringTokens("some/path", KeyringTokens{}) {
		t.Fatal("expected SaveKeyringTokens to refuse an all-empty token bundle")
	}
	if _, ok := LoadKeyringTokens(""); ok {
		t.Fatal("expected LoadKeyringTokens to refuse an empty account")
	}
	// Deleting an account that was never set, or an empty account, must not panic.
	DeleteKeyringTokens("")
	DeleteKeyringTokens("never-set")
}

func TestKeyringTokensUnavailableBackendFallsBackSilently(t *testing.T) {
	// Mirrors real-world failures like macOS's "SecKeychainAddGenericPassword:
	// A keychain cannot be found to store the credential" (no default
	// keychain in the current session — SSH, some CI runners, sandboxes) or
	// a headless Linux box with no D-Bus Secret Service session.
	keyring.MockInitWithError(errors.New("a keychain cannot be found to store the credential"))

	if SaveKeyringTokens("some/path", KeyringTokens{AccessToken: "at"}) {
		t.Fatal("expected SaveKeyringTokens to report false, not panic or succeed, when the backend errors")
	}
	if _, ok := LoadKeyringTokens("some/path"); ok {
		t.Fatal("expected LoadKeyringTokens to report false when the backend errors")
	}
}
