package cmd

import "github.com/spf13/cobra"

var watchCmd = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"w"},
	Short:   "Watch for changes and rebuild",
	Long:    "Watch the project for file changes and automatically rebuild documentation.",
	GroupID: "building",
	Example: `  loko watch
  loko watch --debounce 1000
  loko watch --output ./docs`,
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringP("output", "o", "dist", "output directory")
	watchCmd.Flags().Int("debounce", 500, "debounce delay in milliseconds")
}

func runWatch(cmd *cobra.Command, args []string) error {
	watchCommand := NewWatchCommand(ProjectRoot)

	if output, _ := cmd.Flags().GetString("output"); output != "dist" {
		watchCommand.WithOutputDir(output)
	}
	if debounce, _ := cmd.Flags().GetInt("debounce"); debounce != 500 {
		watchCommand.WithDebounce(debounce)
	}

	return watchCommand.Execute(cmd.Context())
}
