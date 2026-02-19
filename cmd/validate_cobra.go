package cmd

import "github.com/spf13/cobra"

var (
	validateStrict     bool
	validateExitCode   bool
	validateCheckDrift bool
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"val"},
	Short:   "Validate project architecture",
	Long: `Check the project for structural errors, orphaned references, and convention violations.

Flags:
  --strict      Treat warnings as errors (useful for CI/CD)
  --exit-code   Return non-zero exit code on validation failures`,
	GroupID: "building",
	Example: `  loko validate
  loko validate --project ./myproject
  loko validate --strict --exit-code    # For CI/CD pipelines`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Treat warnings as errors")
	validateCmd.Flags().BoolVar(&validateExitCode, "exit-code", false, "Exit with non-zero status on validation failures")
	validateCmd.Flags().BoolVar(&validateCheckDrift, "check-drift", false, "Check for drift between D2 diagrams and frontmatter")
}

func runValidate(cmd *cobra.Command, args []string) error {
	return NewValidateCommand(ProjectRoot, validateStrict, validateExitCode).Execute(cmd.Context())
}
