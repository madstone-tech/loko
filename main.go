// Package main is the entry point for the loko CLI.
// loko is a C4 model architecture documentation tool.
package main

import (
	"fmt"
	"os"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("loko %s (commit: %s, built: %s by %s)\n", version, commit, date, builtBy)
		os.Exit(0)
	}

	fmt.Println("🪇 loko - Guardian of Architectural Wisdom")
	fmt.Println()
	fmt.Println("C4 model architecture documentation with LLM integration")
	fmt.Println()
	fmt.Println("Coming soon! See ROADMAP.md for development progress.")
	fmt.Println("https://github.com/madstone-tech/loko")
}
