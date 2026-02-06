package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("loko %s (commit: %s, built: %s by %s)\n", appVersion, appCommit, appDate, appBuiltBy)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
