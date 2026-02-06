package cmd

import "github.com/spf13/cobra"

var apiCmd = &cobra.Command{
	Use:     "api",
	Short:   "Start HTTP API server",
	Long:    "Start the loko HTTP REST API server.",
	GroupID: "serving",
	RunE:    runAPI,
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func runAPI(cmd *cobra.Command, args []string) error {
	apiCommand := NewAPICommand()
	apiCommand.WithProjectRoot(ProjectRoot)
	return apiCommand.Execute(cmd.Context())
}
