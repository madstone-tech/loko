package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var (
	exportFormat string
	exportOut    string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Write the compiled architecture as a machine-readable artefact",
	Long: `Compile the project and write the resolved intermediate representation.

Output is byte-identical across runs, machines, and file-system orderings, so
the artefact can be committed and diffed. Every artefact carries a schemaVersion
and contains no timestamp, hostname, tool version, or absolute path.

Nothing is written when compilation reports errors: an existing --out file is
left untouched rather than replaced with a partial export.

The artefact goes to stdout and diagnostics to stderr, so this works unfiltered:
  loko export --format json | conftest test -`,
	GroupID:       "building",
	Example:       "  loko export\n  loko export --format toon\n  loko export --out dist/ir.json",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportFormat, "format", string(usecases.ExportJSON),
		"Output encoding: json or toon")
	exportCmd.Flags().StringVar(&exportOut, "out", "",
		"Destination file (default: stdout)")
}

func runExport(cmd *cobra.Command, _ []string) error {
	if !usecases.ValidExportFormat(exportFormat) {
		return fmt.Errorf("unsupported --format %q: use json or toon", exportFormat)
	}

	code, err := runExportWith(cmd.Context(), ExportOptions{
		Root:   ProjectRoot,
		Format: exportFormat,
		Out:    exportOut,
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
