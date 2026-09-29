package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func mockEnv(envMap map[string]string) EnvGetter {
	return func(key string) string {
		return envMap[key]
	}
}

func TestResolvePaths_PlatformFallbacks_Darwin(t *testing.T) {
	home := "/Users/testuser"
	sys := SysInfo{
		OS:      "darwin",
		GetEnv:  mockEnv(map[string]string{}),
		HomeDir: home,
		UID:     501,
		TmpDir:  "/var/folders/xx/tmp",
	}

	paths := ResolvePaths(sys)

	if paths.ConfigDir.Path != "/Users/testuser/.config/qw" {
		t.Errorf("ConfigDir got %q, want %q", paths.ConfigDir.Path, "/Users/testuser/.config/qw")
	}
	if paths.DataDir.Path != "/Users/testuser/.local/share/qw" {
		t.Errorf("DataDir got %q, want %q", paths.DataDir.Path, "/Users/testuser/.local/share/qw")
	}
	if paths.StateDir.Path != "/Users/testuser/.local/state/qw" {
		t.Errorf("StateDir got %q, want %q", paths.StateDir.Path, "/Users/testuser/.local/state/qw")
	}
	if paths.CacheDir.Path != "/Users/testuser/.cache/qw" {
		t.Errorf("CacheDir got %q, want %q", paths.CacheDir.Path, "/Users/testuser/.cache/qw")
	}
	if paths.RuntimeDir.Path != "/var/folders/xx/tmp/qw" {
		t.Errorf("RuntimeDir got %q, want %q", paths.RuntimeDir.Path, "/var/folders/xx/tmp/qw")
	}
}

func TestResolvePaths_PlatformFallbacks_Linux(t *testing.T) {
	home := "/home/testuser"
	sys := SysInfo{
		OS:      "linux",
		GetEnv:  mockEnv(map[string]string{}),
		HomeDir: home,
		UID:     99999,
		TmpDir:  "/tmp",
	}

	paths := ResolvePaths(sys)

	if paths.ConfigDir.Path != "/home/testuser/.config/qw" {
		t.Errorf("ConfigDir got %q, want %q", paths.ConfigDir.Path, "/home/testuser/.config/qw")
	}
	if paths.DataDir.Path != "/home/testuser/.local/share/qw" {
		t.Errorf("DataDir got %q, want %q", paths.DataDir.Path, "/home/testuser/.local/share/qw")
	}
	if paths.StateDir.Path != "/home/testuser/.local/state/qw" {
		t.Errorf("StateDir got %q, want %q", paths.StateDir.Path, "/home/testuser/.local/state/qw")
	}
	if paths.CacheDir.Path != "/home/testuser/.cache/qw" {
		t.Errorf("CacheDir got %q, want %q", paths.CacheDir.Path, "/home/testuser/.cache/qw")
	}
	if paths.RuntimeDir.Path != "/tmp/qw" {
		t.Errorf("RuntimeDir got %q, want %q", paths.RuntimeDir.Path, "/tmp/qw")
	}
}

func TestResolvePaths_AbsoluteEnvOverrides(t *testing.T) {
	home := "/home/testuser"
	env := map[string]string{
		"XDG_CONFIG_HOME": "/custom/config",
		"XDG_DATA_HOME":   "/custom/data",
		"XDG_STATE_HOME":  "/custom/state",
		"XDG_CACHE_HOME":  "/custom/cache",
		"XDG_RUNTIME_DIR": "/custom/runtime",
	}

	sys := SysInfo{
		OS:      "linux",
		GetEnv:  mockEnv(env),
		HomeDir: home,
		UID:     1000,
		TmpDir:  "/tmp",
	}

	paths := ResolvePaths(sys)

	if paths.ConfigDir.Path != "/custom/config/qw" {
		t.Errorf("ConfigDir got %q, want /custom/config/qw", paths.ConfigDir.Path)
	}
	if paths.ConfigDir.Source != "env: $XDG_CONFIG_HOME" {
		t.Errorf("ConfigDir source got %q, want env: $XDG_CONFIG_HOME", paths.ConfigDir.Source)
	}
	if paths.DataDir.Path != "/custom/data/qw" {
		t.Errorf("DataDir got %q, want /custom/data/qw", paths.DataDir.Path)
	}
	if paths.DataDir.Source != "env: $XDG_DATA_HOME" {
		t.Errorf("DataDir source got %q, want env: $XDG_DATA_HOME", paths.DataDir.Source)
	}
	if paths.StateDir.Path != "/custom/state/qw" {
		t.Errorf("StateDir got %q, want /custom/state/qw", paths.StateDir.Path)
	}
	if paths.CacheDir.Path != "/custom/cache/qw" {
		t.Errorf("CacheDir got %q, want /custom/cache/qw", paths.CacheDir.Path)
	}
	if paths.RuntimeDir.Path != "/custom/runtime/qw" {
		t.Errorf("RuntimeDir got %q, want /custom/runtime/qw", paths.RuntimeDir.Path)
	}
	if paths.RuntimeDir.Source != "env: $XDG_RUNTIME_DIR" {
		t.Errorf("RuntimeDir source got %q, want env: $XDG_RUNTIME_DIR", paths.RuntimeDir.Source)
	}
}

func TestResolvePaths_IgnoreRelativePaths(t *testing.T) {
	home := "/Users/testuser"
	env := map[string]string{
		"XDG_CONFIG_HOME": "relative/config",
		"XDG_DATA_HOME":   "./relative/data",
		"XDG_STATE_HOME":  "relative/state",
		"XDG_CACHE_HOME":  "cache",
		"XDG_RUNTIME_DIR": "tmp/runtime",
	}

	sys := SysInfo{
		OS:      "darwin",
		GetEnv:  mockEnv(env),
		HomeDir: home,
		UID:     501,
		TmpDir:  "/var/folders/xx/tmp",
	}

	paths := ResolvePaths(sys)

	if paths.ConfigDir.Path != "/Users/testuser/.config/qw" {
		t.Errorf("Expected relative XDG_CONFIG_HOME to be ignored, got %q", paths.ConfigDir.Path)
	}
	if paths.DataDir.Path != "/Users/testuser/.local/share/qw" {
		t.Errorf("Expected relative XDG_DATA_HOME to be ignored, got %q", paths.DataDir.Path)
	}
	if paths.RuntimeDir.Path != "/var/folders/xx/tmp/qw" {
		t.Errorf("Expected relative XDG_RUNTIME_DIR to be ignored, got %q", paths.RuntimeDir.Path)
	}
}

func TestWorkspacesResolution(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(homeDir, ".config", "qw")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("Default Fallback", func(t *testing.T) {
		sys := SysInfo{
			OS:      "darwin",
			GetEnv:  mockEnv(map[string]string{}),
			HomeDir: homeDir,
			UID:     501,
			TmpDir:  os.TempDir(),
		}
		cfg, err := LoadWithSys(sys)
		if err != nil {
			t.Fatal(err)
		}
		expected := filepath.Join(homeDir, ".local", "share", "qw", "workspaces")
		if cfg.Workspaces.Path != expected {
			t.Errorf("got %q, want %q", cfg.Workspaces.Path, expected)
		}
		if cfg.FileExists {
			t.Errorf("expected FileExists to be false")
		}
	})

	t.Run("Config File workspaces override", func(t *testing.T) {
		yamlContent := "workspaces: ~/my-workspaces\n"
		configFile := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(configFile)

		sys := SysInfo{
			OS:      "darwin",
			GetEnv:  mockEnv(map[string]string{}),
			HomeDir: homeDir,
			UID:     501,
			TmpDir:  os.TempDir(),
		}
		cfg, err := LoadWithSys(sys)
		if err != nil {
			t.Fatal(err)
		}
		expected := filepath.Join(homeDir, "my-workspaces")
		if cfg.Workspaces.Path != expected {
			t.Errorf("got %q, want %q", cfg.Workspaces.Path, expected)
		}
		if !cfg.FileExists {
			t.Errorf("expected FileExists to be true")
		}
	})

	t.Run("Env Override precedence", func(t *testing.T) {
		yamlContent := "workspaces: ~/from-config\n"
		configFile := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(configFile)

		sys := SysInfo{
			OS: "darwin",
			GetEnv: mockEnv(map[string]string{
				"QW_WORKSPACES": "~/from-env",
			}),
			HomeDir: homeDir,
			UID:     501,
			TmpDir:  os.TempDir(),
		}
		cfg, err := LoadWithSys(sys)
		if err != nil {
			t.Fatal(err)
		}
		expected := filepath.Join(homeDir, "from-env")
		if cfg.Workspaces.Path != expected {
			t.Errorf("got %q, want %q", cfg.Workspaces.Path, expected)
		}
		if cfg.Workspaces.Source != "env: $QW_WORKSPACES" {
			t.Errorf("got source %q, want env: $QW_WORKSPACES", cfg.Workspaces.Source)
		}
	})
}

func TestConfigFormat(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")

	sys := SysInfo{
		OS:      "darwin",
		GetEnv:  mockEnv(map[string]string{}),
		HomeDir: homeDir,
		UID:     501,
		TmpDir:  os.TempDir(),
	}

	cfg, err := LoadWithSys(sys)
	if err != nil {
		t.Fatal(err)
	}

	out := cfg.String()
	requiredSubstrings := []string{
		"Configuration Paths:",
		"Config File:",
		"Config Dir:",
		"Data Dir:",
		"State Dir:",
		"Cache Dir:",
		"Runtime Dir:",
		"Settings:",
		"Workspaces Dir:",
	}

	for _, sub := range requiredSubstrings {
		if !strings.Contains(out, sub) {
			t.Errorf("cfg.String() missing substring %q; got:\n%s", sub, out)
		}
	}
}

func TestEnsureDirs(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	runtimeBase := filepath.Join(tempDir, "run")

	sys := SysInfo{
		OS: "darwin",
		GetEnv: mockEnv(map[string]string{
			"XDG_CONFIG_HOME": filepath.Join(homeDir, ".config"),
			"XDG_RUNTIME_DIR": runtimeBase,
		}),
		HomeDir: homeDir,
		UID:     501,
		TmpDir:  os.TempDir(),
	}

	cfg, err := LoadWithSys(sys)
	if err != nil {
		t.Fatal(err)
	}

	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	if runtime.GOOS != "windows" {
		fi, err := os.Stat(cfg.Paths.RuntimeDir.Path)
		if err != nil {
			t.Fatal(err)
		}
		perm := fi.Mode().Perm()
		if perm != 0700 {
			t.Errorf("RuntimeDir perm got %#o, want 0700", perm)
		}

		dataFi, err := os.Stat(cfg.Paths.DataDir.Path)
		if err != nil {
			t.Fatal(err)
		}
		dataPerm := dataFi.Mode().Perm()
		if dataPerm != 0755 {
			t.Errorf("DataDir perm got %#o, want 0755", dataPerm)
		}
	}
}
