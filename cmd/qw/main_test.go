package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionString_Defaults(t *testing.T) {
	out := versionString()

	checks := []struct {
		name     string
		expected string
	}{
		{"prefix", "qw version "},
		{"commit label", "commit: "},
		{"built label", "built: "},
		{"go version label", "go: "},
		{"go version value", runtime.Version()},
		{"os/arch label", "os/arch: "},
		{"os value", runtime.GOOS},
		{"arch value", runtime.GOARCH},
	}

	for _, tc := range checks {
		if !strings.Contains(out, tc.expected) {
			t.Errorf("versionString() missing %s %q; got: %s", tc.name, tc.expected, out)
		}
	}
}

func TestVersionString_Injected(t *testing.T) {
	origVer, origCommit, origDate := version, commit, date
	defer func() {
		version, commit, date = origVer, origCommit, origDate
	}()

	version = "v0.1.5"
	commit = "abc1234"
	date = "2026-09-29T12:00:00Z"

	out := versionString()
	expected := "qw version v0.1.5 (commit: abc1234, built: 2026-09-29T12:00:00Z"
	if !strings.HasPrefix(out, expected) {
		t.Errorf("versionString() got %q, want prefix %q", out, expected)
	}
}
