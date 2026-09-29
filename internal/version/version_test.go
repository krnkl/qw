package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestString_Defaults(t *testing.T) {
	out := String()

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
			t.Errorf("String() missing %s %q; got: %s", tc.name, tc.expected, out)
		}
	}
}

func TestString_Injected(t *testing.T) {
	origVer, origCommit, origDate := version, commit, date
	defer func() {
		version, commit, date = origVer, origCommit, origDate
	}()

	version = "v0.9.9"
	commit = "abc1234"
	date = "2026-09-29T12:00:00Z"

	out := String()

	expected := "qw version v0.9.9 (commit: abc1234, built: 2026-09-29T12:00:00Z"
	if !strings.HasPrefix(out, expected) {
		t.Errorf("String() got %q, want prefix %q", out, expected)
	}
}
