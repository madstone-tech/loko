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

// Execute runs the MCP server.
//
// The server currently registers NO tools. Feature 013 deleted the v0
// file-scaffolding tools along with the model they were written against, and
// the replacements — describe, query, validate, apply_edit, move — belong to
// the authoring stage, which builds them against the compiled IR.
//
// The harness is kept running rather than removed so that an editor's MCP
// configuration keeps working across the gap: a server that starts and honestly
// advertises an empty tool list is easier to diagnose than one that has
// vanished.
func (c *MCPCommand) Execute(ctx context.Context) error {
	server := mcp.NewServer(c.projectRoot, os.Stdin, os.Stdout)

	// Signal readiness on stderr; MCP clients may watch for it.
	fmt.Fprintln(os.Stderr)

	return server.Run(ctx)
}
