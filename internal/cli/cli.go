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

const helpMessage = `qw (kiwi 🥝): Unified Workspace Navigator

Usage:
  qw <command> [flags]

Available Commands:
  config    Inspect resolved XDG paths and configuration parameters
  version   Display version and build information (-v, --version)
  help      Show help for commands (-h, --help)
`

// Run executes the CLI with the provided arguments and standard I/O.
func Run(args []string) int {
	return RunWithIO(args, os.Stdout, os.Stderr)
}

// RunWithIO executes the CLI with injected standard outputs for testing.
func RunWithIO(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, helpMessage)
		return 0
	}

	switch args[0] {
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "version", "-v", "--version":
		return runVersion(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, helpMessage)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command %q. Run 'qw help' for usage.\n", args[0])
		return 1
	}
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

	if err := cfg.Format(stdout); err != nil {
		fmt.Fprintf(stderr, "Error formatting configuration: %v\n", err)
		return 1
	}
	return 0
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	fmt.Fprintln(stdout, version.String())
	return 0
}
