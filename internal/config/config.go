package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

	// 1. Config Dir
	configDir := resolveDir(
		sys.GetEnv("XDG_CONFIG_HOME"),
		filepath.Join(home, ".config"),
		"XDG_CONFIG_HOME",
		fmt.Sprintf("platform fallback: %s (~/.config)", osName),
	)

	// 2. Data Dir
	dataDir := resolveDir(
		sys.GetEnv("XDG_DATA_HOME"),
		filepath.Join(home, ".local", "share"),
		"XDG_DATA_HOME",
		fmt.Sprintf("platform fallback: %s (~/.local/share)", osName),
	)

	// 3. State Dir
	stateDir := resolveDir(
		sys.GetEnv("XDG_STATE_HOME"),
		filepath.Join(home, ".local", "state"),
		"XDG_STATE_HOME",
		fmt.Sprintf("platform fallback: %s (~/.local/state)", osName),
	)

	// 4. Cache Dir
	cacheDir := resolveDir(
		sys.GetEnv("XDG_CACHE_HOME"),
		filepath.Join(home, ".cache"),
		"XDG_CACHE_HOME",
		fmt.Sprintf("platform fallback: %s (~/.cache)", osName),
	)

	// 5. Runtime Dir
	runtimeDir := resolveRuntimeDir(sys)

	// Config File is config.yaml inside ConfigDir
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

	// Linux fallback: /run/user/<UID>, or /tmp if UID dir doesn't exist
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

// yamlFileConfig mirrors possible keys in config.yaml
type yamlFileConfig struct {
	Ws         string `yaml:"ws"`
	Workspaces string `yaml:"workspaces"`
}

// Load evaluates the full configuration state from environment and config.yaml.
func Load() (*Config, error) {
	return LoadWithSys(DefaultSysInfo())
}

// LoadWithSys allows loading configuration with custom system abstraction for testing.
func LoadWithSys(sys SysInfo) (*Config, error) {
	paths := ResolvePaths(sys)

	// Check if config.yaml exists
	var fileCfg yamlFileConfig
	var fileSourceKey string
	fileExists := false

	if data, err := os.ReadFile(paths.ConfigFile.Path); err == nil {
		fileExists = true
		if err := yaml.Unmarshal(data, &fileCfg); err == nil {
			if fileCfg.Ws != "" {
				fileSourceKey = "ws"
			} else if fileCfg.Workspaces != "" {
				fileSourceKey = "workspaces"
			}
		}
	}

	// Workspaces resolution priority:
	// 1. QW_WORKSPACES env
	// 2. QW_WS env
	// 3. config.yaml ws / workspaces
	// 4. Default: <DataDir>/ws
	var wsPath string
	var wsSource string

	envWorkspaces := sys.GetEnv("QW_WORKSPACES")
	envWs := sys.GetEnv("QW_WS")

	if envWorkspaces != "" {
		wsPath = ExpandHome(strings.TrimSpace(envWorkspaces), sys.HomeDir)
		wsSource = "env: $QW_WORKSPACES"
	} else if envWs != "" {
		wsPath = ExpandHome(strings.TrimSpace(envWs), sys.HomeDir)
		wsSource = "env: $QW_WS"
	} else if fileSourceKey != "" {
		val := fileCfg.Ws
		if val == "" {
			val = fileCfg.Workspaces
		}
		wsPath = ExpandHome(strings.TrimSpace(val), sys.HomeDir)
		wsSource = fmt.Sprintf("config: %s (%s)", paths.ConfigFile.Path, fileSourceKey)
	} else {
		wsPath = filepath.Join(paths.DataDir.Path, "ws")
		wsSource = fmt.Sprintf("default XDG: %s/ws", paths.DataDir.Path)
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

	// Runtime dir requires 0700 permissions
	if err := os.MkdirAll(c.Paths.RuntimeDir.Path, 0700); err != nil {
		return fmt.Errorf("failed to create runtime directory %s: %w", c.Paths.RuntimeDir.Path, err)
	}

	return nil
}
