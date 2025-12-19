package cmd

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
)

// ValidateCommand validates the project structure.
type ValidateCommand struct {
	projectRoot string
}

// NewValidateCommand creates a new validate command.
func NewValidateCommand(projectRoot string) *ValidateCommand {
	return &ValidateCommand{
		projectRoot: projectRoot,
	}
}

// Execute runs the validate command.
func (c *ValidateCommand) Execute(ctx context.Context) error {
	// Load the project
	projectRepo := filesystem.NewProjectRepository()
	_, err := projectRepo.LoadProject(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// List systems
	systems, err := projectRepo.ListSystems(ctx, c.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to list systems: %w", err)
	}

	if len(systems) == 0 {
		fmt.Println("⚠  No systems found in project")
		return nil
	}

	// Validate structure
	errorCount := 0
	warningCount := 0

	for _, system := range systems {
		if system == nil {
			continue
		}

		// Check system has required metadata
		if system.Name == "" {
			fmt.Printf("✗ System %s: missing name\n", system.ID)
			errorCount++
		}

		// Check containers
		if len(system.Containers) == 0 {
			fmt.Printf("⚠ System %s: has no containers\n", system.Name)
			warningCount++
			continue
		}

		for _, container := range system.Containers {
			if container == nil {
				continue
			}

			// Check container metadata
			if container.Name == "" {
				fmt.Printf("✗ System %s, Container %s: missing name\n", system.Name, container.ID)
				errorCount++
			}

			// Check components exist if referenced
			if len(container.Components) == 0 {
				fmt.Printf("⚠ System %s, Container %s: has no components\n", system.Name, container.Name)
				warningCount++
			}
		}
	}

	// Count containers
	containerCount := 0
	for _, system := range systems {
		if system != nil {
			containerCount += len(system.Containers)
		}
	}

	// Print summary
	fmt.Println()
	if errorCount == 0 && warningCount == 0 {
		fmt.Printf("✓ Validation passed: %d systems, %d containers\n", len(systems), containerCount)
		return nil
	}

	if errorCount > 0 {
		fmt.Printf("✗ %d error(s) found\n", errorCount)
	}

	if warningCount > 0 {
		fmt.Printf("⚠ %d warning(s) found\n", warningCount)
	}

	if errorCount > 0 {
		return fmt.Errorf("validation failed with %d error(s)", errorCount)
	}

	return nil
}
