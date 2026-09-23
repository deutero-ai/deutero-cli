// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

package cliutil

import (
	"encoding/json"

	"github.com/zalando/go-keyring"
)

// keyringService namespaces this CLI's entries in the OS credential store
// (macOS Keychain, Linux Secret Service via D-Bus, Windows Credential
// Manager) from every other application using the same store.
const keyringService = "deutero-pp-cli"

// KeyringTokens is the secret bundle mirrored into the OS keyring alongside
// (or instead of) the credentials.toml file, following the storage approach
// in Stytch's Connected Apps CLI guide.
type KeyringTokens struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

func (t KeyringTokens) empty() bool {
	return t.AccessToken == "" && t.RefreshToken == "" && t.ClientSecret == ""
}

// SaveKeyringTokens best-effort stores tok in the OS keyring under account
// (the resolved credentials.toml path, so relocated profiles/homes don't
// collide). It returns false — never an error — when no keyring backend is
// available: headless Linux without a Secret Service session, CI, containers,
// and agent sandboxes all fall into this case, and callers are expected to
// fall back to the existing permission-verified credentials.toml.
func SaveKeyringTokens(account string, tok KeyringTokens) bool {
	if account == "" || tok.empty() {
		return false
	}
	blob, err := json.Marshal(tok) // #nosec G117 -- tok is the payload intentionally stored in the OS keyring (keyring.Set below), not logged or transmitted
	if err != nil {
		return false
	}
	return keyring.Set(keyringService, account, string(blob)) == nil
}

// LoadKeyringTokens best-effort reads tok back. ok is false when the keyring
// is unavailable, the account has no entry, or the stored value is
// unreadable — every case is treated identically as "nothing to overlay".
func LoadKeyringTokens(account string) (tok KeyringTokens, ok bool) {
	if account == "" {
		return tok, false
	}
	raw, err := keyring.Get(keyringService, account)
	if err != nil {
		return tok, false
	}
	if err := json.Unmarshal([]byte(raw), &tok); err != nil {
		return tok, false
	}
	return tok, !tok.empty()
}

// DeleteKeyringTokens best-effort removes account's entry. Safe to call
// whether or not an entry (or a keyring backend) exists.
func DeleteKeyringTokens(account string) {
	if account == "" {
		return
	}
	_ = keyring.Delete(keyringService, account)
}
