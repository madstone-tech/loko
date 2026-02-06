package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:     "mcp",
	Short:   "Start MCP server",
	Long:    "Start the Model Context Protocol server for LLM integration via stdio.",
	GroupID: "serving",
	RunE:    runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().String("env", "", "environment variable (KEY=VALUE)")
}

func runMCP(cmd *cobra.Command, args []string) error {
	if envVar, _ := cmd.Flags().GetString("env"); envVar != "" {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			_ = os.Setenv(parts[0], parts[1])
		}
	}

	return NewMCPCommand(ProjectRoot).Execute(cmd.Context())
}
