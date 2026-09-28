package version

import (
	"strings"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	str := String()
	if !strings.HasPrefix(str, "qw version ") {
		t.Errorf("expected version string to start with 'qw version ', got %q", str)
	}

	jsonStr := JSON()
	if !strings.Contains(jsonStr, `"version":`) {
		t.Errorf("expected json output to contain version key, got %q", jsonStr)
	}
}
