package version

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	// Injected at build time via -ldflags
	version = ""
	commit  = "dev"
	date    = "unknown"
)

type versionInfo struct {
	version   string
	commit    string
	date      string
	goVersion string
	os        string
	arch      string
}

func getVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	// Fallback for local development checkouts without ldflags: read root VERSION from disk if present
	for _, p := range []string{"VERSION", "../VERSION", "../../VERSION"} {
		if data, err := os.ReadFile(p); err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				if !strings.HasPrefix(v, "v") {
					return "v" + v
				}
				return v
			}
		}
	}
	return "dev"
}

func getVersionInfo() versionInfo {
	return versionInfo{
		version:   getVersion(),
		commit:    commit,
		date:      date,
		goVersion: runtime.Version(),
		os:        runtime.GOOS,
		arch:      runtime.GOARCH,
	}
}

// String returns formatted human-readable version output for the CLI.
func String() string {
	info := getVersionInfo()
	return fmt.Sprintf("qw version %s (commit: %s, built: %s, go: %s, os/arch: %s/%s)",
		info.version, info.commit, info.date, info.goVersion, info.os, info.arch)
}
