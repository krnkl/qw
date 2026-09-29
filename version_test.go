package qw

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionString(t *testing.T) {
	output := VersionString()

	expectedVersion := strings.TrimSpace(rawVersion)
	if !strings.HasPrefix(expectedVersion, "v") {
		expectedVersion = "v" + expectedVersion
	}

	checks := []struct {
		name     string
		expected string
	}{
		{"prefix", "qw version "},
		{"version", expectedVersion},
		{"commit label", "commit: "},
		{"commit value", commit},
		{"built label", "built: "},
		{"built value", date},
		{"go version label", "go: "},
		{"go version value", runtime.Version()},
		{"os/arch label", "os/arch: "},
		{"os value", runtime.GOOS},
		{"arch value", runtime.GOARCH},
	}

	for _, tc := range checks {
		if !strings.Contains(output, tc.expected) {
			t.Errorf("VersionString() missing %s %q; got: %s", tc.name, tc.expected, output)
		}
	}
}
