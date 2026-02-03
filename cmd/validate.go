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
	fmt.Println()

	// Print issues grouped by severity
	if len(report.Issues) > 0 {
		// Errors
		var errors []usecases.ArchitectureIssue
		for _, issue := range report.Issues {
			if issue.Severity == "error" {
				errors = append(errors, issue)
			}
		}
		if len(errors) > 0 {
			fmt.Println("❌ Errors:")
			for _, issue := range errors {
				fmt.Printf("  [%s] %s\n", issue.Code, issue.Title)
				fmt.Printf("    %s\n", issue.Description)
				if len(issue.Affected) > 0 {
					fmt.Printf("    Affected: %v\n", issue.Affected)
				}
				if issue.Suggestion != "" {
					fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
				}
			}
			fmt.Println()
		}

		// Warnings
		var warnings []usecases.ArchitectureIssue
		for _, issue := range report.Issues {
			if issue.Severity == "warning" {
				warnings = append(warnings, issue)
			}
		}
		if len(warnings) > 0 {
			fmt.Println("⚠️  Warnings:")
			for _, issue := range warnings {
				fmt.Printf("  [%s] %s\n", issue.Code, issue.Title)
				fmt.Printf("    %s\n", issue.Description)
				if len(issue.Affected) > 0 {
					fmt.Printf("    Affected: %v\n", issue.Affected)
				}
				if issue.Suggestion != "" {
					fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
				}
			}
			fmt.Println()
		}

		// Infos
		var infos []usecases.ArchitectureIssue
		for _, issue := range report.Issues {
			if issue.Severity == "info" {
				infos = append(infos, issue)
			}
		}
		if len(infos) > 0 {
			fmt.Println("ℹ️  Information:")
			for _, issue := range infos {
				fmt.Printf("  [%s] %s\n", issue.Code, issue.Title)
				fmt.Printf("    %s\n", issue.Description)
			}
			fmt.Println()
		}
	}

	// Print summary
	fmt.Println("Summary:")
	fmt.Printf("  Total Issues: %d\n", report.Total)
	fmt.Printf("  Errors: %d\n", report.Errors)
	fmt.Printf("  Warnings: %d\n", report.Warnings)
	fmt.Printf("  Info: %d\n", report.Infos)

	if report.IsValid {
		fmt.Println("\n✓ Architecture is valid!")
	}
}
