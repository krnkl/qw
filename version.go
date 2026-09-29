package qw

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"
)

//go:embed VERSION
var rawVersion string

var (
	// Injected at build time via -ldflags if provided
	commit = "dev"
	date   = "unknown"
)

type versionInfo struct {
	version   string
	commit    string
	date      string
	goVersion string
	os        string
	arch      string
}

func getVersionInfo() versionInfo {
	v := strings.TrimSpace(rawVersion)
	if v != "" && !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if v == "" {
		v = "dev"
	}

	return versionInfo{
		version:   v,
		commit:    commit,
		date:      date,
		goVersion: runtime.Version(),
		os:        runtime.GOOS,
		arch:      runtime.GOARCH,
	}
}

// VersionString returns formatted human-readable version output for the CLI.
func VersionString() string {
	info := getVersionInfo()
	return fmt.Sprintf("qw version %s (commit: %s, built: %s, go: %s, os/arch: %s/%s)",
		info.version, info.commit, info.date, info.goVersion, info.os, info.arch)
}
