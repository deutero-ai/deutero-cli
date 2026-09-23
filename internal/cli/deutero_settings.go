// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// DefaultHost is where the CLI points when nothing overrides it.
const DefaultHost = "https://dashboard.deutero.ai"

// APIBasePath is the path the study API is mounted under. Callers set a host
// ("https://example.com"); this is appended for them, because pasting a host
// out of a browser bar and getting 404s on every call is the single most
// likely first-run failure.
const APIBasePath = "/study-api"

// settings holds the values a user can persist across invocations. They live
// in their own file rather than the generated credentials config because the
// auth commands rewrite that file from a struct, which would silently drop
// any key they do not know about.
type settings struct {
	// Host is the base host, stored without the API base path.
	Host string `json:"host,omitempty"`
	// DefaultStudy and DefaultProject let study-scoped commands be run
	// without repeating an ID that rarely changes during a session.
	DefaultStudy   string `json:"default_study,omitempty"`
	DefaultProject string `json:"default_project,omitempty"`
}

func settingsPath() string {
	if v := strings.TrimSpace(os.Getenv("DEUTERO_SETTINGS_PATH")); v != "" {
		return v
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "deutero-settings.json"
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "deutero-pp-cli", "settings.json")
}

func loadSettings() settings {
	var s settings
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func saveSettings(s settings) error {
	path := settingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// normalizeHost turns whatever the user typed into a base URL the client can
// use. It accepts a bare host, a host with a scheme, a trailing slash, or a
// host that already includes the API base path.
func normalizeHost(raw string) (string, error) {
	h := strings.TrimSpace(raw)
	if h == "" {
		return "", fmt.Errorf("host is empty")
	}
	if !strings.Contains(h, "://") {
		h = "https://" + h
	}
	u, err := url.Parse(h)
	if err != nil {
		return "", fmt.Errorf("invalid host %q: %w", raw, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid host %q: no hostname", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid host %q: scheme must be http or https", raw)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	if u.Path == "" {
		u.Path = APIBasePath
	}
	u.RawQuery, u.Fragment = "", ""
	return u.String(), nil
}

// hostBase strips the API base path back off, for display.
func hostBase(baseURL string) string {
	return strings.TrimSuffix(baseURL, APIBasePath)
}

// hostResolution caches what applyHostOverride decided, and why. Without it,
// resolvedHost would re-read DEUTERO_BASE_URL after applyHostOverride had
// written the settings-file value into it, and report the config file's own
// value as though the user had exported it.
var hostResolution struct {
	baseURL string
	source  string
	done    bool
}

// resolvedHost reports the base URL in force and where it came from, using the
// precedence applyHostOverride applied: flag, env, settings file, default.
func resolvedHost(hostFlag string) (baseURL, source string) {
	if hostResolution.done {
		return hostResolution.baseURL, hostResolution.source
	}
	return computeHost(hostFlag)
}

// computeHost walks the precedence chain without mutating anything.
func computeHost(hostFlag string) (baseURL, source string) {
	if v := strings.TrimSpace(hostFlag); v != "" {
		if norm, err := normalizeHost(v); err == nil {
			return norm, "--host flag"
		}
	}
	for _, name := range []string{"DEUTERO_HOST", "DEUTERO_BASE_URL"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			if norm, err := normalizeHost(v); err == nil {
				return norm, name + " env var"
			}
		}
	}
	if v := strings.TrimSpace(loadSettings().Host); v != "" {
		if norm, err := normalizeHost(v); err == nil {
			return norm, "config file (" + settingsPath() + ")"
		}
	}
	def, _ := normalizeHost(DefaultHost)
	return def, "built-in default"
}

// applyHostOverride publishes the resolved base URL through the environment
// variable the generated config loader already reads. Doing it this way means
// the override lands for every command and every code path that builds a
// client, without forking generated code.
//
// Precedence, highest first: --host, DEUTERO_HOST, DEUTERO_BASE_URL, the
// settings file, the built-in default.
func applyHostOverride(hostFlag string) error {
	// Validate an explicit --host eagerly so a typo fails with a clear message
	// rather than falling through to a lower-precedence source.
	if v := strings.TrimSpace(hostFlag); v != "" {
		if _, err := normalizeHost(v); err != nil {
			return err
		}
	}
	baseURL, source := computeHost(hostFlag)
	hostResolution.baseURL, hostResolution.source, hostResolution.done = baseURL, source, true
	return os.Setenv("DEUTERO_BASE_URL", baseURL)
}
