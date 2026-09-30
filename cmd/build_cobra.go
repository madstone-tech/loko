package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var (
	buildFormats []string
	buildOut     string
	buildStrict  bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Render diagrams, markdown and a site from the architecture",
	Long: `Compile the project, derive its views, and render every requested format into
the output directory.

Views need no configuration: a landscape, one per system with containers, one
per container with components, and one per environment. Declared view blocks
are rendered alongside them.

Output is byte-identical across runs and machines, so it can be committed.
Every file opens with a notice saying it is generated and from which source.
Only files a previous build wrote are ever removed; nothing is written when
compilation reports errors, and no external program is run.`,
	GroupID:       "building",
	Example:       "  loko build\n  loko build --format svg,html --out public\n  loko build --strict",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringSliceVar(&buildFormats, "format", nil,
		"Formats to produce: d2, svg, md, html (default: all)")
	buildCmd.Flags().StringVar(&buildOut, "out", "",
		"Output directory (default: <project>/dist)")
	buildCmd.Flags().BoolVar(&buildStrict, "strict", false,
		"Exit 2 when warnings are present")
}

func runBuild(cmd *cobra.Command, _ []string) error {
	code, err := runBuildWith(cmd.Context(), BuildOptions{
		Root:    ProjectRoot,
		Formats: buildFormats,
		Out:     buildOut,
		Strict:  buildStrict,
		Stdout:  cmd.OutOrStdout(),
		Stderr:  cmd.ErrOrStderr(),
	})
	if err != nil {
		return err
	}
	if code != usecases.ExitSuccess {
		os.Exit(code)
	}
	return nil
}
