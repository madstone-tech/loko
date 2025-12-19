package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/adapters/html"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// WatchCommand watches for file changes and rebuilds documentation.
type WatchCommand struct {
	projectRoot string
	outputDir   string
	debounceMs  int
}

// NewWatchCommand creates a new watch command.
func NewWatchCommand(projectRoot string) *WatchCommand {
	return &WatchCommand{
		projectRoot: projectRoot,
		outputDir:   "dist",
		debounceMs:  500,
	}
}

// WithOutputDir sets the output directory.
func (c *WatchCommand) WithOutputDir(dir string) *WatchCommand {
	c.outputDir = dir
	return c
}

// WithDebounce sets the debounce delay in milliseconds.
func (c *WatchCommand) WithDebounce(ms int) *WatchCommand {
	c.debounceMs = ms
	return c
}

// Execute runs the watch command.
func (c *WatchCommand) Execute(ctx context.Context) error {
	// Load the project
	projectRepo := filesystem.NewProjectRepository()
	project, err := projectRepo.LoadProject(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Create file watcher
	watcher, err := filesystem.NewFileWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Stop()

	// Start watching
	events, err := watcher.Watch(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	fmt.Println("👁  Watching for changes...")
	fmt.Printf("   Project: %s\n", c.projectRoot)
	fmt.Printf("   Output: %s\n", c.outputDir)
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()

	// Create adapters
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return fmt.Errorf("failed to create site builder: %w", err)
	}

	progressReporter := &simpleProgressReporter{}
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	// Track debounce timer
	debounceTimer := time.NewTimer(time.Duration(c.debounceMs) * time.Millisecond)
	debounceTimer.Stop()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initial build
	systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
	if err == nil && len(systems) > 0 {
		fmt.Println("🔨 Initial build...")
		if err := buildDocs.Execute(ctx, project, systems, c.outputDir); err != nil {
			fmt.Printf("✗ Build failed: %v\n", err)
		} else {
			fmt.Println("✓ Initial build complete")
		}
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\n✓ Watch stopped")
			return nil

		case event := <-events:
			if event.Path == "" {
				// Channel closed
				return nil
			}

			// Reset debounce timer
			debounceTimer.Reset(time.Duration(c.debounceMs) * time.Millisecond)
			fmt.Printf("📝 Change detected: %s\n", event.Path)

		case <-debounceTimer.C:
			// Debounce time elapsed, rebuild
			fmt.Println("🔨 Rebuilding...")

			systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
			if err != nil {
				fmt.Printf("✗ Error loading systems: %v\n", err)
				continue
			}

			if len(systems) == 0 {
				fmt.Println("⚠  No systems found")
				continue
			}

			startTime := time.Now()
			if err := buildDocs.Execute(ctx, project, systems, c.outputDir); err != nil {
				fmt.Printf("✗ Build failed: %v\n", err)
			} else {
				elapsed := time.Since(startTime)
				fmt.Printf("✓ Rebuild complete (%v)\n", elapsed.Round(10*time.Millisecond))
			}
			fmt.Println()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
