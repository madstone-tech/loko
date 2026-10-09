package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/madstone-tech/loko/internal/mcp"
)

// MCPCommand runs the MCP server over stdio.
type MCPCommand struct {
	projectRoot string
}

// NewMCPCommand creates a new MCP command.
func NewMCPCommand(projectRoot string) *MCPCommand {
	return &MCPCommand{projectRoot: projectRoot}
}

// Execute runs the MCP server with the authoring tools registered.
func (c *MCPCommand) Execute(ctx context.Context) error {
	server := mcp.NewServer(c.projectRoot, os.Stdin, os.Stdout)
	for _, tool := range newMCPTools(c.projectRoot) {
		if err := server.RegisterTool(tool); err != nil {
			return fmt.Errorf("registering MCP tool: %w", err)
		}
	}

	// Signal readiness on stderr; MCP clients may watch for it.
	fmt.Fprintln(os.Stderr)

	return server.Run(ctx)
}
