package cmd

import (
	"fmt"
	"os"

	"github.com/madstone-tech/loko/internal/core/usecases"

	"github.com/spf13/cobra"
)

var (
	validateStrict   bool
	validateFormat   string
	validateSummary  bool
	validateSeverity string
	validateIgnore   []string
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"val"},
	Short:   "Compile the architecture and report diagnostics",
	Long: `Compile every *.loko.hcl file beneath the project root and report every
problem found, each with a file, line, and column.

All diagnostics from one run are reported together: five independent mistakes
produce five diagnostics, not one.

Exit codes:
  0  no errors (warnings present without --strict still exit 0)
  1  one or more errors
  2  warnings present and --strict was given`,
	GroupID: "building",
	Example: `  loko validate
  loko validate --strict                 # warnings fail the run
  loko validate --format json | jq .     # machine-readable, for CI
  loko validate --summary                # one line per diagnostic code
  loko validate --strict --ignore missing_docs`,
	// Diagnostics are the output; cobra should not print its own error or
	// usage text on top of them.
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"Exit 2 when warnings are present")
	validateCmd.Flags().StringVar(&validateFormat, "format", formatText,
		"Output format: text or json")
	validateCmd.Flags().BoolVar(&validateSummary, "summary", false,
		"Group diagnostics by code: a count and the affected addresses")
	validateCmd.Flags().StringVar(&validateSeverity, "severity", "",
		"Show only this severity: error (the exit code is unchanged)")
	validateCmd.Flags().StringSliceVar(&validateIgnore, "ignore", nil,
		"Warning codes to leave out, on top of the project's ignore_warnings")
}

func runValidate(cmd *cobra.Command, _ []string) error {
	if validateFormat != formatText && validateFormat != formatJSON {
		return fmt.Errorf("unsupported --format %q: use text or json", validateFormat)
	}

	code, err := runValidateWith(cmd.Context(), ValidateOptions{
		Root:   ProjectRoot,
		Strict: validateStrict,
		Format: validateFormat,
		Ignore: validateIgnore, Severity: validateSeverity, Summary: validateSummary,
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.ErrOrStderr(),
	})
	if err != nil {
		return err
	}
	if code != usecases.ExitSuccess {
		os.Exit(code)
	}
	return nil
}
