// Copyright 2026 francis and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"path/filepath"
	"testing"
)

func TestNormalizeHostAppendsAPIBasePath(t *testing.T) {
	cases := map[string]string{
		"dashboard.deutero.ai":                   "https://dashboard.deutero.ai/study-api",
		"https://dashboard.deutero.ai":           "https://dashboard.deutero.ai/study-api",
		"https://dashboard.deutero.ai/":          "https://dashboard.deutero.ai/study-api",
		"https://dashboard.deutero.ai/study-api": "https://dashboard.deutero.ai/study-api",
		"http://localhost:5000":                  "http://localhost:5000/study-api",
		"https://dashboard.deutero.ai/?x=1#frag": "https://dashboard.deutero.ai/study-api",
	}
	for in, want := range cases {
		got, err := normalizeHost(in)
		if err != nil {
			t.Fatalf("normalizeHost(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("normalizeHost(%q) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "   ", "ftp://example.com"} {
		if _, err := normalizeHost(bad); err == nil {
			t.Errorf("normalizeHost(%q) accepted an invalid host", bad)
		}
	}
}

// Precedence is the whole point of the command: a --host typo must not fall
// through to a stored value and silently talk to the wrong deployment.
func TestHostPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	t.Setenv("DEUTERO_SETTINGS_PATH", path)
	t.Setenv("DEUTERO_HOST", "")
	t.Setenv("DEUTERO_BASE_URL", "")

	if base, source := computeHost(""); base != DefaultHost+APIBasePath || source != "built-in default" {
		t.Fatalf("empty state: got %q from %q", base, source)
	}

	if err := saveSettings(settings{Host: "https://stored.example.com"}); err != nil {
		t.Fatalf("saveSettings: %v", err)
	}
	base, source := computeHost("")
	if base != "https://stored.example.com/study-api" {
		t.Errorf("stored host ignored: got %q (%s)", base, source)
	}

	t.Setenv("DEUTERO_HOST", "https://env.example.com")
	if base, _ := computeHost(""); base != "https://env.example.com/study-api" {
		t.Errorf("env should beat the settings file: got %q", base)
	}

	if base, source := computeHost("https://flag.example.com"); base != "https://flag.example.com/study-api" || source != "--host flag" {
		t.Errorf("flag should beat env: got %q from %q", base, source)
	}

	if err := applyHostOverride("nope://bad"); err == nil {
		t.Error("an invalid --host should fail loudly, not fall through to a lower-precedence source")
	}
}

func TestConfigUnsetClearsOnlyTheNamedSetting(t *testing.T) {
	t.Setenv("DEUTERO_SETTINGS_PATH", filepath.Join(t.TempDir(), "settings.json"))
	if err := saveSettings(settings{Host: "https://a.example.com", DefaultStudy: "s1", DefaultProject: "p1"}); err != nil {
		t.Fatalf("saveSettings: %v", err)
	}
	cmd := newConfigUnsetCmd(&rootFlags{quiet: true})
	if err := cmd.RunE(cmd, []string{"study"}); err != nil {
		t.Fatalf("unset study: %v", err)
	}
	got := loadSettings()
	if got.DefaultStudy != "" {
		t.Errorf("default study still set: %q", got.DefaultStudy)
	}
	if got.Host != "https://a.example.com" || got.DefaultProject != "p1" {
		t.Errorf("unset clobbered a sibling setting: %+v", got)
	}
	if err := cmd.RunE(cmd, []string{"nonsense"}); err == nil {
		t.Error("unset of an unknown key should error")
	}
}
