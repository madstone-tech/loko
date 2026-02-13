package cmd

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateCommand validates the project architecture for errors and warnings.
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
	project, err := projectRepo.LoadProject(ctx, c.projectRoot)
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

	// Build architecture graph
	graphBuilder := usecases.NewBuildArchitectureGraph()
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Validate architecture
	validator := usecases.NewValidateArchitecture()
	report := validator.Execute(graph, systems)

	// Print validation results
	c.printReport(report)

	// Return error if validation failed
	if !report.IsValid {
		return fmt.Errorf("validation failed with %d error(s)", report.Errors)
	}

	return nil
}

// printReport prints the validation report to stdout.
func (c *ValidateCommand) printReport(report *usecases.ArchitectureReport) {
	report.Print()
}
