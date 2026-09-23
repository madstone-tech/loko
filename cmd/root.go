// Package cmd implements the loko CLI commands using Cobra.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time version information, set via SetVersionInfo from main.go.
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
	appBuiltBy = "unknown"
)

// Persistent flag values accessible to all subcommands.
var (
	ProjectRoot string
	Verbose     bool
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "loko",
	Short: "Guardian of Architectural Wisdom",
	Long: `loko compiles a C4 software architecture from HCL.

You describe the architecture once, in *.loko.hcl files: the people, systems,
containers and components, the relationships between them, and the environments
they run in. loko resolves every reference at compile time, so a broken
relationship is an error with a file, line and column rather than a diagram that
has quietly gone stale.

Everything else is a projection of the compiled result. Nothing generated is
ever edited by hand.`,
	SilenceUsage: true,
}

func init() {
	// Persistent flags available to all subcommands.
	rootCmd.PersistentFlags().StringVarP(&ProjectRoot, "project", "p", ".", "project root directory")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose output (env: LOKO_VERBOSE)")

	// Command groups for organized help output.
	// The "scaffolding" group went with the scaffolding commands in feature
	// 013; authoring is editing source, or the MCP write tools once the
	// authoring stage lands.
	rootCmd.AddGroup(
		&cobra.Group{ID: "building", Title: "Building"},
		&cobra.Group{ID: "serving", Title: "Serving"},
	)
}

// Execute runs the root command. This is the main entry point called from main.go.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets build-time version information from ldflags.
// Call this from main.go before Execute().
func SetVersionInfo(version, commit, date, builtBy string) {
	appVersion = version
	appCommit = commit
	appDate = date
	appBuiltBy = builtBy

	rootCmd.Version = version
	rootCmd.SetVersionTemplate(
		fmt.Sprintf("loko %s (commit: %s, built: %s by %s)\n", version, commit, date, builtBy),
	)
}
