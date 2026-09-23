// Package main is the entry point for the loko CLI.
// loko is a C4 model architecture documentation tool.
package main

import (
	"fmt"
	"os"

	"github.com/madstone-tech/loko/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date, builtBy)
	if err := cmd.Execute(); err != nil {
		// Commands set SilenceErrors so cobra does not print on top of the
		// diagnostics they render themselves. Operational failures still have
		// to reach the user, so they are printed here, at the process
		// boundary — a command that exits non-zero with no message is worse
		// than no error handling at all.
		fmt.Fprintln(os.Stderr, "loko: "+err.Error())
		os.Exit(1)
	}
}
