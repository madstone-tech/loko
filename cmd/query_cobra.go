package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/madstone-tech/loko/internal/core/usecases"

	"github.com/spf13/cobra"
)

var (
	queryFormat     string
	queryTransitive bool
	queryLimit      int
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Ask the architecture a question",
	Long: `Answer graph questions about the compiled architecture: who depends on an
element, what it depends on, the path between two elements, orphans, and
coupling. An element includes everything inside it, so the dependents of a
container include components elsewhere that call into it.

These are the same answers the MCP query tool gives: --format json prints
exactly what the tool returns.

Exit codes:
  0  answered
  1  the project does not compile, or an address is unknown`,
	GroupID: "building",
	Example: `  loko query dependents container.orders_db
  loko query dependencies container.web --transitive
  loko query path person.customer external.bank
  loko query orphans
  loko query coupling --limit 10 --format json`,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVar(&queryFormat, "format", formatText, "Output format: text, json or toon")
	for _, kind := range []string{"dependents", "dependencies"} {
		c := querySubcommand(kind, "<address>", "List the elements that "+map[string]string{
			"dependents": "depend on an element", "dependencies": "an element depends on"}[kind], cobra.ExactArgs(1))
		c.Flags().BoolVar(&queryTransitive, "transitive", false, "Follow relationships transitively")
		queryCmd.AddCommand(c)
	}
	queryCmd.AddCommand(querySubcommand("path", "<from> <to>", "Show the shortest chain of relationships between two elements", cobra.ExactArgs(2)))
	queryCmd.AddCommand(querySubcommand("orphans", "", "List elements with no relationships", cobra.NoArgs))
	coupling := querySubcommand("coupling", "", "Rank elements by fan-in plus fan-out", cobra.NoArgs)
	coupling.Flags().IntVar(&queryLimit, "limit", 20, "Rows to show")
	queryCmd.AddCommand(coupling)
}

func querySubcommand(kind, args, short string, nargs cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:           kind + " " + args,
		Short:         short,
		Args:          nargs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, positional []string) error {
			return runQuery(cmd, kind, positional)
		},
	}
}

func runQuery(cmd *cobra.Command, kind string, args []string) error {
	if !slices.Contains([]string{formatText, formatJSON, formatTOON}, queryFormat) {
		return fmt.Errorf("unsupported --format %q: use text, json or toon", queryFormat)
	}
	req := usecases.QueryRequest{Kind: kind, Transitive: queryTransitive, Limit: queryLimit}
	if len(args) > 0 {
		req.Address = args[0]
	}
	if len(args) > 1 {
		req.To = args[1]
	}
	code, err := runQueryWith(cmd.Context(), QueryOptions{
		Root: ProjectRoot, Format: queryFormat, Request: req,
		Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr(),
	})
	if err != nil {
		return err
	}
	if code != usecases.ExitSuccess {
		os.Exit(code)
	}
	return nil
}
