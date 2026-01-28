package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/api"
)

// APICommand starts the HTTP API server.
type APICommand struct {
	port        int
	projectRoot string
	apiKey      string
}

// NewAPICommand creates a new API command.
func NewAPICommand() *APICommand {
	return &APICommand{
		port:        8081,
		projectRoot: ".",
	}
}

// WithPort sets the port.
func (c *APICommand) WithPort(port int) *APICommand {
	c.port = port
	return c
}

// WithProjectRoot sets the project root.
func (c *APICommand) WithProjectRoot(root string) *APICommand {
	c.projectRoot = root
	return c
}

// WithAPIKey sets the API key for authentication.
func (c *APICommand) WithAPIKey(key string) *APICommand {
	c.apiKey = key
	return c
}

// Execute starts the API server.
func (c *APICommand) Execute(ctx context.Context) error {
	// Create repository
	repo := filesystem.NewProjectRepository()

	// Create server config
	config := api.DefaultConfig()
	config.Port = c.port
	config.ProjectRoot = c.projectRoot
	config.APIKey = c.apiKey

	// Create server
	server := api.NewServer(config, repo)

	// Print startup message
	fmt.Fprintf(os.Stderr, "Starting loko API server on port %d\n", c.port)
	if c.apiKey != "" {
		fmt.Fprintf(os.Stderr, "Authentication: enabled (API key required)\n")
	} else {
		fmt.Fprintf(os.Stderr, "Authentication: disabled\n")
	}
	fmt.Fprintf(os.Stderr, "Project root: %s\n", c.projectRoot)
	fmt.Fprintf(os.Stderr, "\nEndpoints:\n")
	fmt.Fprintf(os.Stderr, "  GET  /health           - Health check\n")
	fmt.Fprintf(os.Stderr, "  GET  /api/v1/project   - Get project info\n")
	fmt.Fprintf(os.Stderr, "  GET  /api/v1/systems   - List all systems\n")
	fmt.Fprintf(os.Stderr, "  GET  /api/v1/systems/{id} - Get system details\n")
	fmt.Fprintf(os.Stderr, "  POST /api/v1/build     - Trigger documentation build\n")
	fmt.Fprintf(os.Stderr, "  GET  /api/v1/build/{id} - Get build status\n")
	fmt.Fprintf(os.Stderr, "  GET  /api/v1/validate  - Validate architecture\n")
	fmt.Fprintf(os.Stderr, "\nPress Ctrl+C to stop\n\n")

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nShutting down...\n")
		cancel()
	}()

	// Start server (blocks until context is cancelled)
	return server.Start(ctx)
}
