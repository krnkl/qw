package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithIO([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "qw (kiwi 🥝): Unified Workspace Navigator") {
		t.Errorf("expected help output, got %q", out)
	}
}

func TestCLI_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithIO([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "qw version ") {
		t.Errorf("expected version output, got %q", out)
	}

	stdout.Reset()
	code = RunWithIO([]string{"-v"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "qw version ") {
		t.Errorf("expected version output on -v, got %q", stdout.String())
	}
}

func TestCLI_Config(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithIO([]string{"config"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d; stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Configuration Paths:") {
		t.Errorf("expected config paths header, got %q", out)
	}
	if !strings.Contains(out, "Workspaces Dir:") {
		t.Errorf("expected workspaces dir line, got %q", out)
	}

	// Test JSON flag
	stdout.Reset()
	code = RunWithIO([]string{"config", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0 for --json, got %d", code)
	}
	jsonOut := stdout.String()
	if !strings.Contains(jsonOut, `"paths":`) || !strings.Contains(jsonOut, `"workspaces":`) {
		t.Errorf("expected valid JSON structure, got %q", jsonOut)
	}
}
