package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var fmtCheck bool

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Rewrite source files in canonical form",
	Long: `Rewrite every *.loko.hcl file beneath the project root in canonical form.

Comments, declaration order, and meaning are preserved. A file that is already
canonical is left byte-unchanged, so running fmt twice changes nothing the
second time. A file that does not parse is reported and left untouched.

With --check nothing is written: the non-canonical paths are listed and the
command exits 1, which is the form a CI job needs.`,
	GroupID:       "building",
	Example:       "  loko fmt\n  loko fmt --check    # fail CI on unformatted source",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runFmt,
}

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false,
		"List non-canonical files and exit 1 without writing anything")
}

func runFmt(cmd *cobra.Command, _ []string) error {
	code, err := runFmtWith(cmd.Context(), FmtOptions{
		Root:   ProjectRoot,
		Check:  fmtCheck,
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
