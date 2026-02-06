package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init <project-name>",
	Short:   "Initialize a new loko project",
	Long:    "Create a new C4 architecture documentation project with loko.toml configuration and directory structure.",
	GroupID: "scaffolding",
	Args:    cobra.ExactArgs(1),
	Example: `  loko init myproject
  loko init myproject --description "Payment system architecture"
  loko init myproject --template serverless`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("description", "d", "", "project description")
	initCmd.Flags().String("path", "", "project path (defaults to project name)")
	initCmd.Flags().StringP("template", "t", "standard-3layer", "template to use")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	initCommand := NewInitCommand(projectName)

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		initCommand.WithDescription(desc)
	}
	if path, _ := cmd.Flags().GetString("path"); path != "" {
		initCommand.WithPath(path)
	}

	if err := initCommand.Execute(cmd.Context()); err != nil {
		return err
	}

	fmt.Printf("✓ Project '%s' initialized\n", projectName)
	return nil
}
