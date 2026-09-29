package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
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
