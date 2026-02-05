package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ServeCommand serves the documentation locally.
type ServeCommand struct {
	outputDir string
	address   string
	port      string
}

// NewServeCommand creates a new serve command.
func NewServeCommand(outputDir string) *ServeCommand {
	return &ServeCommand{
		outputDir: outputDir,
		address:   "localhost",
		port:      "8080",
	}
}

// WithAddress sets the server address.
func (c *ServeCommand) WithAddress(address string) *ServeCommand {
	c.address = address
	return c
}

// WithPort sets the server port.
func (c *ServeCommand) WithPort(port string) *ServeCommand {
	c.port = port
	return c
}

// Execute runs the serve command.
func (c *ServeCommand) Execute(ctx context.Context) error {
	// Verify output directory exists
	if info, err := os.Stat(c.outputDir); err != nil || !info.IsDir() {
		return fmt.Errorf("output directory not found: %s", c.outputDir)
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.Dir(c.outputDir))
	mux.Handle("/", fileServer)

	// Create server
	addr := net.JoinHostPort(c.address, c.port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel for errors
	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		fmt.Printf("🚀 Server starting on http://%s\n", addr)
		fmt.Printf("   Serving documentation from: %s\n", c.outputDir)
		fmt.Println("   Press Ctrl+C to stop")
		errChan <- server.ListenAndServe()
	}()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigChan:
		fmt.Printf("\n✓ Received signal: %v\n", sig)
		fmt.Println("✓ Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		fmt.Println("✓ Server stopped")
	}

	return nil
}
