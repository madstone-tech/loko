package cmd

import "github.com/spf13/cobra"

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"val"},
	Short:   "Validate project architecture",
	Long:    "Check the project for structural errors, orphaned references, and convention violations.",
	GroupID: "building",
	Example: `  loko validate
  loko validate --project ./myproject`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	return NewValidateCommand(ProjectRoot).Execute(cmd.Context())
}
