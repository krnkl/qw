package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/krnkl/qw/internal/config"
	"github.com/krnkl/qw/internal/version"
)

// Run executes the CLI with the provided arguments and standard I/O.
func Run(args []string) int {
	return RunWithIO(args, os.Stdout, os.Stderr)
}

// RunWithIO executes the CLI with injected standard outputs for testing.
func RunWithIO(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "version", "-v", "--version":
		return runVersion(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command %q. Run 'qw help' for usage.\n", args[0])
		return 1
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "qw (kiwi 🥝): Unified Workspace Navigator")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  qw <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Available Commands:")
	fmt.Fprintln(w, "  config    Inspect resolved XDG paths and configuration parameters")
	fmt.Fprintln(w, "  version   Display version and build information (-v, --version)")
	fmt.Fprintln(w, "  help      Show help for commands (-h, --help)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --json    Output in structured JSON format (supported on config and version)")
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "Output configuration paths as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "Error loading configuration: %v\n", err)
		return 1
	}

	if *jsonOutput {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(stderr, "Error formatting JSON: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}

	fileStatus := "[missing]"
	if cfg.FileExists {
		fileStatus = "[found]"
	}

	fmt.Fprintln(stdout, "Configuration Paths:")
	fmt.Fprintf(stdout, "  Config File:    %s %s (%s)\n", cfg.Paths.ConfigFile.Path, fileStatus, cfg.Paths.ConfigFile.Source)
	fmt.Fprintf(stdout, "  Config Dir:     %s (resolved via: %s)\n", cfg.Paths.ConfigDir.Path, cfg.Paths.ConfigDir.Source)
	fmt.Fprintf(stdout, "  Data Dir:       %s (resolved via: %s)\n", cfg.Paths.DataDir.Path, cfg.Paths.DataDir.Source)
	fmt.Fprintf(stdout, "  State Dir:      %s (resolved via: %s)\n", cfg.Paths.StateDir.Path, cfg.Paths.StateDir.Source)
	fmt.Fprintf(stdout, "  Cache Dir:      %s (resolved via: %s)\n", cfg.Paths.CacheDir.Path, cfg.Paths.CacheDir.Source)
	fmt.Fprintf(stdout, "  Runtime Dir:    %s (resolved via: %s)\n", cfg.Paths.RuntimeDir.Path, cfg.Paths.RuntimeDir.Source)
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Settings:")
	fmt.Fprintf(stdout, "  Workspaces Dir: %s (resolved via: %s)\n", cfg.Workspaces.Path, cfg.Workspaces.Source)

	return 0
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "Output version information as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *jsonOutput {
		fmt.Fprintln(stdout, version.JSON())
		return 0
	}

	fmt.Fprintln(stdout, version.String())
	return 0
}
