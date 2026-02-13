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
	strict      bool
	exitCode    bool
}

// NewValidateCommand creates a new validate command.
func NewValidateCommand(projectRoot string, strict, exitCode bool) *ValidateCommand {
	return &ValidateCommand{
		projectRoot: projectRoot,
		strict:      strict,
		exitCode:    exitCode,
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

	// Handle strict mode: treat warnings as errors
	hasIssues := report.Errors > 0
	if c.strict && report.Warnings > 0 {
		hasIssues = true
		fmt.Println("\n⚠  Strict mode: Treating warnings as errors")
	}

	// Return error if validation failed (or has issues in strict mode)
	if hasIssues {
		if c.exitCode {
			// exit-code flag: return error with exit code 1
			if c.strict && report.Errors == 0 {
				return fmt.Errorf("validation failed with %d warning(s) (strict mode)", report.Warnings)
			}
			return fmt.Errorf("validation failed with %d error(s)", report.Errors)
		}
		// Without exit-code flag, print message but return success
		fmt.Println("\n⚠  Note: Use --exit-code flag to exit with non-zero status")
	}

	return nil
}

// printReport prints the validation report to stdout.
func (c *ValidateCommand) printReport(report *usecases.ArchitectureReport) {
	report.Print()
}
