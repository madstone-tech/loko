package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// ValidateArchitecture checks the architecture for violations and issues.
// It detects:
// 1. Circular dependencies (A -> B -> A)
// 2. Isolated components (no relationships)
// 3. Overly coupled components (too many relationships)
// 4. Missing relationships (dangling references)
type ValidateArchitecture struct{}

// NewValidateArchitecture creates a new ValidateArchitecture use case.
func NewValidateArchitecture() *ValidateArchitecture {
	return &ValidateArchitecture{}
}

// ArchitectureIssue represents a single architecture violation or concern.
type ArchitectureIssue struct {
	Severity    string   // "error", "warning", "info"
	Code        string   // "circular_dependency", "isolated_component", "high_coupling", "dangling_reference", "missing_component"
	Title       string   // Human-readable title
	Description string   // Detailed description
	Affected    []string // IDs of affected components
	Suggestion  string   // How to fix it
}

// ArchitectureReport contains all validation issues found in the architecture.
type ArchitectureReport struct {
	Issues   []ArchitectureIssue
	IsValid  bool
	Summary  string
	Total    int
	Errors   int
	Warnings int
	Infos    int
}

// Execute validates the architecture graph and returns a report of all issues found.
//
// It checks:
// - Circular dependencies using cycle detection
// - Isolated components with no relationships
// - High coupling (components with many relationships)
// - Dangling references (relationships to non-existent components)
// - Missing components in the hierarchy
func (uc *ValidateArchitecture) Execute(
	graph *entities.ArchitectureGraph,
	systems []*entities.System,
) *ArchitectureReport {
	if graph == nil {
		return &ArchitectureReport{
			IsValid: false,
			Summary: "Graph is nil",
			Issues: []ArchitectureIssue{
				{
					Severity:    "error",
					Code:        "invalid_graph",
					Title:       "Graph is nil",
					Description: "Cannot validate a nil architecture graph",
				},
			},
		}
	}

	report := &ArchitectureReport{
		Issues: make([]ArchitectureIssue, 0),
	}

	// Check for circular dependencies
	uc.checkCircularDependencies(graph, report)

	// Check for isolated components
	uc.checkIsolatedComponents(graph, report)

	// Check for high coupling
	uc.checkHighCoupling(graph, report)

	// Check for dangling references
	uc.checkDanglingReferences(graph, systems, report)

	// Determine overall validity
	report.IsValid = report.Errors == 0
	report.Total = len(report.Issues)

	// Generate summary
	if report.IsValid {
		report.Summary = "Architecture is valid - no critical issues found"
	} else {
		report.Summary = fmt.Sprintf("Architecture has %d issue(s): %d error(s), %d warning(s), %d info(s)",
			report.Total, report.Errors, report.Warnings, report.Infos)
	}

	return report
}

// Print outputs the validation report to stdout.
func (report *ArchitectureReport) Print() {
	fmt.Println()

	if len(report.Issues) > 0 {
		for _, severity := range []string{"error", "warning", "info"} {
			issues := report.GetIssuesBySeverity(severity)
			if len(issues) == 0 {
				continue
			}
			switch severity {
			case "error":
				fmt.Println("Errors:")
			case "warning":
				fmt.Println("Warnings:")
			case "info":
				fmt.Println("Information:")
			}
			for _, issue := range issues {
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
	}

	fmt.Println("Summary:")
	fmt.Printf("  Total Issues: %d\n", report.Total)
	fmt.Printf("  Errors: %d\n", report.Errors)
	fmt.Printf("  Warnings: %d\n", report.Warnings)
	fmt.Printf("  Info: %d\n", report.Infos)

	if report.IsValid {
		fmt.Println("\nArchitecture is valid!")
	}
}

// GetIssuesBySeverity filters issues by severity level.
func (report *ArchitectureReport) GetIssuesBySeverity(severity string) []ArchitectureIssue {
	filtered := make([]ArchitectureIssue, 0)
	for _, issue := range report.Issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// GetIssuesByCode filters issues by code.
func (report *ArchitectureReport) GetIssuesByCode(code string) []ArchitectureIssue {
	filtered := make([]ArchitectureIssue, 0)
	for _, issue := range report.Issues {
		if issue.Code == code {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
