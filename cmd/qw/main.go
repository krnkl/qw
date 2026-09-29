package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/krnkl/qw/internal/cli"
)

var (
	// Injected at build time via -ldflags
	version = "dev"
	commit  = "dev"
	date    = "unknown"
)

func versionString() string {
	return fmt.Sprintf("qw version %s (commit: %s, built: %s, go: %s, os/arch: %s/%s)",
		version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func main() {
	os.Exit(cli.RunWithVersion(os.Args[1:], versionString()))
}
