package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// AppName is the identifier used for subdirectory names in XDG paths.
const AppName = "qw"

// PathInfo holds a resolved filesystem path along with the exact source of resolution.
type PathInfo struct {
	Path   string `json:"path" yaml:"path"`
	Source string `json:"source" yaml:"source"`
}

// Paths represents all XDG base directories and the config file for qw.
type Paths struct {
	ConfigFile PathInfo `json:"config_file" yaml:"config_file"`
	ConfigDir  PathInfo `json:"config_dir" yaml:"config_dir"`
	DataDir    PathInfo `json:"data_dir" yaml:"data_dir"`
	StateDir   PathInfo `json:"state_dir" yaml:"state_dir"`
	CacheDir   PathInfo `json:"cache_dir" yaml:"cache_dir"`
	RuntimeDir PathInfo `json:"runtime_dir" yaml:"runtime_dir"`
}

// WorkspacesConfig holds the effective workspace directory and its source.
type WorkspacesConfig struct {
	Path   string `json:"path" yaml:"path"`
	Source string `json:"source" yaml:"source"`
}

// Config represents the evaluated configuration state.
type Config struct {
	Paths      Paths            `json:"paths" yaml:"paths"`
	Workspaces WorkspacesConfig `json:"workspaces" yaml:"workspaces"`
	FileExists bool             `json:"file_exists" yaml:"file_exists"`
}

const configTemplateText = `Configuration Paths:
  Config File:    {{.Paths.ConfigFile.Path}} {{if .FileExists}}[found]{{else}}[missing]{{end}} ({{.Paths.ConfigFile.Source}})
  Config Dir:     {{.Paths.ConfigDir.Path}} (resolved via: {{.Paths.ConfigDir.Source}})
  Data Dir:       {{.Paths.DataDir.Path}} (resolved via: {{.Paths.DataDir.Source}})
  State Dir:      {{.Paths.StateDir.Path}} (resolved via: {{.Paths.StateDir.Source}})
  Cache Dir:      {{.Paths.CacheDir.Path}} (resolved via: {{.Paths.CacheDir.Source}})
  Runtime Dir:    {{.Paths.RuntimeDir.Path}} (resolved via: {{.Paths.RuntimeDir.Source}})

Settings:
  Workspaces Dir: {{.Workspaces.Path}} (resolved via: {{.Workspaces.Source}})
`

var configTmpl = template.Must(template.New("config").Parse(configTemplateText))

// Format renders the human-readable configuration view using Go text/template.
func (c *Config) Format(w io.Writer) error {
	return configTmpl.Execute(w, c)
}

// String returns formatted human-readable configuration state.
func (c *Config) String() string {
	var buf bytes.Buffer
	_ = c.Format(&buf)
	return buf.String()
}

// EnvGetter is an abstraction for querying environment variables (for testing).
type EnvGetter func(string) string

// SysInfo abstracts platform detection for path resolution tests.
type SysInfo struct {
	OS      string
	GetEnv  EnvGetter
	HomeDir string
	UID     int
	TmpDir  string
}

// DefaultSysInfo returns real system information and environment lookups.
func DefaultSysInfo() SysInfo {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return SysInfo{
		OS:      runtime.GOOS,
		GetEnv:  os.Getenv,
		HomeDir: home,
		UID:     os.Getuid(),
		TmpDir:  os.TempDir(),
	}
}

// ResolvePaths resolves all XDG paths and the config file based on SysInfo.
func ResolvePaths(sys SysInfo) Paths {
	home := sys.HomeDir
	osName := sys.OS

	configDir := resolveDir(
		sys.GetEnv("XDG_CONFIG_HOME"),
		filepath.Join(home, ".config"),
		"XDG_CONFIG_HOME",
		fmt.Sprintf("platform fallback: %s (~/.config)", osName),
	)

	dataDir := resolveDir(
		sys.GetEnv("XDG_DATA_HOME"),
		filepath.Join(home, ".local", "share"),
		"XDG_DATA_HOME",
		fmt.Sprintf("platform fallback: %s (~/.local/share)", osName),
	)

	stateDir := resolveDir(
		sys.GetEnv("XDG_STATE_HOME"),
		filepath.Join(home, ".local", "state"),
		"XDG_STATE_HOME",
		fmt.Sprintf("platform fallback: %s (~/.local/state)", osName),
	)

	cacheDir := resolveDir(
		sys.GetEnv("XDG_CACHE_HOME"),
		filepath.Join(home, ".cache"),
		"XDG_CACHE_HOME",
		fmt.Sprintf("platform fallback: %s (~/.cache)", osName),
	)

	runtimeDir := resolveRuntimeDir(sys)

	configFile := PathInfo{
		Path:   filepath.Join(configDir.Path, "config.yaml"),
		Source: fmt.Sprintf("config dir: %s", configDir.Source),
	}

	return Paths{
		ConfigFile: configFile,
		ConfigDir:  configDir,
		DataDir:    dataDir,
		StateDir:   stateDir,
		CacheDir:   cacheDir,
		RuntimeDir: runtimeDir,
	}
}

func resolveDir(envVal, defaultBase, envKey, fallbackDesc string) PathInfo {
	if envVal != "" && filepath.IsAbs(envVal) {
		return PathInfo{
			Path:   filepath.Join(envVal, AppName),
			Source: fmt.Sprintf("env: $%s", envKey),
		}
	}
	return PathInfo{
		Path:   filepath.Join(defaultBase, AppName),
		Source: fallbackDesc,
	}
}

func resolveRuntimeDir(sys SysInfo) PathInfo {
	envVal := sys.GetEnv("XDG_RUNTIME_DIR")
	if envVal != "" && filepath.IsAbs(envVal) {
		return PathInfo{
			Path:   filepath.Join(envVal, AppName),
			Source: "env: $XDG_RUNTIME_DIR",
		}
	}

	if sys.OS == "darwin" {
		tmp := sys.TmpDir
		if tmp == "" || !filepath.IsAbs(tmp) {
			tmp = "/tmp"
		}
		return PathInfo{
			Path:   filepath.Join(tmp, AppName),
			Source: "platform fallback: macOS $TMPDIR",
		}
	}

	uidDir := fmt.Sprintf("/run/user/%d", sys.UID)
	if sys.UID >= 0 {
		if fi, err := os.Stat(uidDir); err == nil && fi.IsDir() {
			return PathInfo{
				Path:   filepath.Join(uidDir, AppName),
				Source: fmt.Sprintf("platform fallback: linux (/run/user/%d)", sys.UID),
			}
		}
	}

	return PathInfo{
		Path:   filepath.Join("/tmp", AppName),
		Source: "platform fallback: linux (/tmp)",
	}
}

// ExpandHome replaces leading ~ with the user's home directory.
func ExpandHome(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

type yamlFileConfig struct {
	Workspaces string `yaml:"workspaces"`
}

// Load evaluates the full configuration state from environment and config.yaml.
func Load() (*Config, error) {
	return LoadWithSys(DefaultSysInfo())
}

// LoadWithSys allows loading configuration with custom system abstraction for testing.
func LoadWithSys(sys SysInfo) (*Config, error) {
	paths := ResolvePaths(sys)

	var fileCfg yamlFileConfig
	var fileConfigured bool
	fileExists := false

	if data, err := os.ReadFile(paths.ConfigFile.Path); err == nil {
		fileExists = true
		if err := yaml.Unmarshal(data, &fileCfg); err == nil {
			if fileCfg.Workspaces != "" {
				fileConfigured = true
			}
		}
	}

	var wsPath string
	var wsSource string

	envWorkspaces := sys.GetEnv("QW_WORKSPACES")

	if envWorkspaces != "" {
		wsPath = ExpandHome(strings.TrimSpace(envWorkspaces), sys.HomeDir)
		wsSource = "env: $QW_WORKSPACES"
	} else if fileConfigured {
		wsPath = ExpandHome(strings.TrimSpace(fileCfg.Workspaces), sys.HomeDir)
		wsSource = fmt.Sprintf("config: %s (workspaces)", paths.ConfigFile.Path)
	} else {
		wsPath = filepath.Join(paths.DataDir.Path, "workspaces")
		wsSource = fmt.Sprintf("default XDG: %s/workspaces", paths.DataDir.Path)
	}

	return &Config{
		Paths: paths,
		Workspaces: WorkspacesConfig{
			Path:   wsPath,
			Source: wsSource,
		},
		FileExists: fileExists,
	}, nil
}

// EnsureDirs creates all required XDG directories with correct POSIX permissions.
func (c *Config) EnsureDirs() error {
	dirs755 := []string{
		c.Paths.ConfigDir.Path,
		c.Paths.DataDir.Path,
		c.Paths.StateDir.Path,
		c.Paths.CacheDir.Path,
	}

	for _, dir := range dirs755 {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.MkdirAll(c.Paths.RuntimeDir.Path, 0700); err != nil {
		return fmt.Errorf("failed to create runtime directory %s: %w", c.Paths.RuntimeDir.Path, err)
	}

	return nil
}
