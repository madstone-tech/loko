// Package main is the entry point for the loko CLI.
// loko is a C4 model architecture documentation tool.
package main

import (
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
		os.Exit(1)
	}
}
