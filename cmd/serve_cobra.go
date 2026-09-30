package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Preview the site locally, rebuilding as the architecture changes",
	Long: `Build the site in memory, serve it on http://127.0.0.1, and rebuild whenever an
architecture source file, a prose file it references, or a theme override
changes. The browser reloads by itself.

A change that does not compile replaces every page with the diagnostics, with
file, line and column; fixing it recovers without restarting. Nothing is
written to disk.`,
	GroupID:       "serving",
	Example:       "  loko serve\n  loko serve --port 3000",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to listen on (loopback only)")
}

func runServe(cmd *cobra.Command, _ []string) error {
	code, err := runServeWith(cmd.Context(), ServeOptions{
		Root:   ProjectRoot,
		Port:   servePort,
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
