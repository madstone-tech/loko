package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/mcp"
	"github.com/madstone-tech/loko/internal/mcp/tools"
)

// MCPCommand starts the MCP server.
type MCPCommand struct {
	projectRoot string
}

// NewMCPCommand creates a new MCP command.
func NewMCPCommand(projectRoot string) *MCPCommand {
	return &MCPCommand{
		projectRoot: projectRoot,
	}
}

// Execute runs the MCP server.
func (c *MCPCommand) Execute(ctx context.Context) error {
	// Create repository
	repo := filesystem.NewProjectRepository()

	// Create MCP server
	server := mcp.NewServer(c.projectRoot, os.Stdin, os.Stdout)

	// Register all tools
	if err := registerTools(server, repo); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Signal to stderr that we're ready (empty line - MCP clients may check for this)
	// This allows Claude Code to detect that the server has initialized
	fmt.Fprintln(os.Stderr)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- server.Run(ctx)
	}()

	// Wait for either server error or signal
	select {
	case <-sigChan:
		return nil
	case err := <-serverErrChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// registerTools registers all MCP tools with the server.
func registerTools(server *mcp.Server, repo *filesystem.ProjectRepository) error {
	// Create diagram renderer
	renderer := d2.NewRenderer()

	toolList := []mcp.Tool{
		tools.NewQueryProjectTool(repo),
		tools.NewQueryArchitectureTool(repo),
		tools.NewCreateSystemTool(repo),
		tools.NewCreateContainerTool(repo),
		tools.NewCreateComponentTool(repo),
		tools.NewUpdateDiagramTool(repo),
		tools.NewUpdateSystemTool(repo),
		tools.NewUpdateContainerTool(repo),
		tools.NewUpdateComponentTool(repo),
		tools.NewBuildDocsTool(repo),
		tools.NewValidateTool(repo),
		tools.NewValidateDiagramTool(renderer),
		tools.NewQueryDependenciesTool(repo),
		tools.NewQueryRelatedComponentsTool(repo),
		tools.NewAnalyzeCouplingTool(repo),
		tools.NewSearchElementsTool(repo),    // New: search by pattern/filters
		tools.NewFindRelationshipsTool(repo), // New: find relationships
	}

	for _, tool := range toolList {
		if err := server.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %q: %w", tool.Name(), err)
		}
	}

	return nil
}
