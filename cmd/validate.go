package cmd

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateCommand validates the project architecture for errors and warnings.
type ValidateCommand struct {
	projectRoot string
	strict      bool
	exitCode    bool
	checkDrift  bool
}

// NewValidateCommand creates a new validate command.
func NewValidateCommand(projectRoot string, strict, exitCode bool) *ValidateCommand {
	return &ValidateCommand{
		projectRoot: projectRoot,
		strict:      strict,
		exitCode:    exitCode,
		checkDrift:  validateCheckDrift, // Access the global flag
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

	// Check for drift if requested
	if c.checkDrift {
		return c.executeDriftCheck(ctx, projectRepo, systems)
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

// executeDriftCheck runs drift detection and formats output according to the contract.
func (c *ValidateCommand) executeDriftCheck(ctx context.Context, projectRepo usecases.ProjectRepository, systems []*entities.System) error {
	// Create drift detection use case
	driftUC := usecases.NewDetectDrift(projectRepo)

	// Execute drift detection
	req := &usecases.DetectDriftRequest{
		ProjectRoot: c.projectRoot,
		Systems:     systems,
	}

	result, err := driftUC.Execute(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to check for drift: %w", err)
	}

	// Format output according to the contract
	if result.HasErrors {
		fmt.Println("❌ Validation failed - Critical drift detected")
		fmt.Println("Issues found:")
		for _, issue := range result.Issues {
			if issue.Severity == entities.DriftError {
				fmt.Printf("  %s (ERROR): %s\n", issue.ComponentID, issue.Message)
			}
		}
		return fmt.Errorf("drift detection failed with %d error(s)", len(result.Issues))
	} else if result.HasWarnings {
		fmt.Println("⚠️  Validation passed with warnings")
		fmt.Println("Issues found:")
		for _, issue := range result.Issues {
			if issue.Severity == entities.DriftWarning {
				fmt.Printf("  %s (WARNING): %s\n", issue.ComponentID, issue.Message)
				if issue.Context != "" {
					fmt.Printf("    %s\n", issue.Context)
				}
			}
		}
		return nil
	} else {
		fmt.Printf("✅ Validation passed - No drift detected\n")
		fmt.Printf("  Components checked: %d\n", result.ComponentsChecked)
		fmt.Printf("  Drift issues found: %d\n", len(result.Issues))
		return nil
	}
}

// printReport prints the validation report to stdout.
func (c *ValidateCommand) printReport(report *usecases.ArchitectureReport) {
	report.Print()
}
