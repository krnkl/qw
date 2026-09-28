package main

import (
	"os"

	"github.com/krnkl/qw/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
