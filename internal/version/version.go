package version

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

//go:embed VERSION
var embeddedVersion string

var (
	// Version is injected at build time via -ldflags.
	Version = ""
	// Commit is injected at build time via -ldflags.
	Commit = "unknown"
	// Date is injected at build time via -ldflags.
	Date = "unknown"
)

// Info holds version metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the current version string, falling back to embedded VERSION or "dev".
func Get() string {
	if Version != "" {
		return Version
	}
	trimmed := strings.TrimSpace(embeddedVersion)
	if trimmed != "" {
		if !strings.HasPrefix(trimmed, "v") {
			return "v" + trimmed
		}
		return trimmed
	}
	return "dev"
}

// GetInfo returns complete version information.
func GetInfo() Info {
	return Info{
		Version:   Get(),
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns formatted human-readable version output.
func String() string {
	info := GetInfo()
	return fmt.Sprintf("qw version %s (commit: %s, built: %s, go: %s, os/arch: %s/%s)",
		info.Version, info.Commit, info.Date, info.GoVersion, info.OS, info.Arch)
}

// JSON returns version information formatted as a JSON string.
func JSON() string {
	data, _ := json.MarshalIndent(GetInfo(), "", "  ")
	return string(data)
}
