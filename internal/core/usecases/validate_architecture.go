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

// checkCircularDependencies detects cycles in the dependency graph.
// Uses DFS with color marking to detect cycles.
func (uc *ValidateArchitecture) checkCircularDependencies(
	graph *entities.ArchitectureGraph,
	report *ArchitectureReport,
) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)         // Recursion stack for cycle detection
	cyclePathMap := make(map[string][]string) // Store the cycle path for each node

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			uc.dfs(nodeID, graph, visited, recStack, []string{}, report, cyclePathMap)
		}
	}
}

// dfs performs depth-first search for cycle detection.
func (uc *ValidateArchitecture) dfs(
	nodeID string,
	graph *entities.ArchitectureGraph,
	visited map[string]bool,
	recStack map[string]bool,
	path []string,
	report *ArchitectureReport,
	cyclePathMap map[string][]string,
) {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	// Check all outgoing edges
	edges := graph.GetOutgoingEdges(nodeID)
	for _, edge := range edges {
		if !visited[edge.Target] {
			uc.dfs(edge.Target, graph, visited, recStack, append([]string{}, path...), report, cyclePathMap)
		} else if recStack[edge.Target] {
			// Found a cycle
			cycleStart := -1
			for i, n := range path {
				if n == edge.Target {
					cycleStart = i
					break
				}
			}

			if cycleStart != -1 {
				cyclePath := path[cycleStart:]
				cyclePath = append(cyclePath, edge.Target) // Complete the cycle

				issue := ArchitectureIssue{
					Severity:    "error",
					Code:        "circular_dependency",
					Title:       fmt.Sprintf("Circular dependency detected: %s", formatCyclePath(cyclePath)),
					Description: fmt.Sprintf("Components form a circular dependency: %s", formatCyclePath(cyclePath)),
					Affected:    cyclePath,
					Suggestion:  "Refactor components to break the cycle. Consider extracting shared logic into a common component or reversing one of the dependencies.",
				}
				report.Issues = append(report.Issues, issue)
				report.Errors++
			}
		}
	}

	recStack[nodeID] = false
}

// formatCyclePath formats a cycle path for display.
func formatCyclePath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	result := ""
	for i, id := range path {
		result += id
		if i < len(path)-1 {
			result += " → "
		}
	}
	return result
}

// checkIsolatedComponents finds components with no relationships.
func (uc *ValidateArchitecture) checkIsolatedComponents(
	graph *entities.ArchitectureGraph,
	report *ArchitectureReport,
) {
	isolated := make([]string, 0)

	for nodeID := range graph.Nodes {
		deps := graph.GetDependencies(nodeID)
		dependents := graph.GetDependents(nodeID)

		if len(deps) == 0 && len(dependents) == 0 {
			isolated = append(isolated, nodeID)
		}
	}

	if len(isolated) > 0 {
		issue := ArchitectureIssue{
			Severity:    "info",
			Code:        "isolated_component",
			Title:       fmt.Sprintf("%d isolated component(s) found", len(isolated)),
			Description: "These components have no relationships with other components, which might indicate incomplete architecture modeling or truly independent components.",
			Affected:    isolated,
			Suggestion:  "Review if these components should have relationships with other components, or if they are truly independent.",
		}
		report.Issues = append(report.Issues, issue)
		report.Infos++
	}
}

// checkHighCoupling finds components with many relationships.
const highCouplingThreshold = 5

func (uc *ValidateArchitecture) checkHighCoupling(
	graph *entities.ArchitectureGraph,
	report *ArchitectureReport,
) {
	highlyCoupled := make([]string, 0)

	for nodeID := range graph.Nodes {
		deps := graph.GetDependencies(nodeID)
		if len(deps) >= highCouplingThreshold {
			highlyCoupled = append(highlyCoupled, nodeID)
		}
	}

	if len(highlyCoupled) > 0 {
		issue := ArchitectureIssue{
			Severity:    "warning",
			Code:        "high_coupling",
			Title:       fmt.Sprintf("%d highly coupled component(s) found", len(highlyCoupled)),
			Description: fmt.Sprintf("These components depend on %d or more other components, indicating high coupling:", highCouplingThreshold),
			Affected:    highlyCoupled,
			Suggestion:  "Consider breaking down these components or extracting common functionality to reduce coupling and improve maintainability.",
		}
		report.Issues = append(report.Issues, issue)
		report.Warnings++
	}
}

// checkDanglingReferences finds relationships that point to non-existent components.
func (uc *ValidateArchitecture) checkDanglingReferences(
	graph *entities.ArchitectureGraph,
	systems []*entities.System,
	report *ArchitectureReport,
) {
	danglingRefs := make(map[string][]string) // component -> dangling targets

	// Check all components in all systems
	for _, sys := range systems {
		for _, container := range sys.Containers {
			for _, comp := range container.Components {
				for targetID := range comp.Relationships {
					// Check if target exists in graph
					if _, exists := graph.Nodes[targetID]; !exists {
						if _, ok := danglingRefs[comp.ID]; !ok {
							danglingRefs[comp.ID] = make([]string, 0)
						}
						danglingRefs[comp.ID] = append(danglingRefs[comp.ID], targetID)
					}
				}
			}
		}
	}

	if len(danglingRefs) > 0 {
		affected := make([]string, 0)
		var description string
		for comp, targets := range danglingRefs {
			affected = append(affected, comp)
			description += fmt.Sprintf("  %s references: %v\n", comp, targets)
		}

		issue := ArchitectureIssue{
			Severity:    "error",
			Code:        "dangling_reference",
			Title:       fmt.Sprintf("%d component(s) with dangling references found", len(danglingRefs)),
			Description: "These components reference components that don't exist:\n" + description,
			Affected:    affected,
			Suggestion:  "Either create the referenced components or remove the references from the source components.",
		}
		report.Issues = append(report.Issues, issue)
		report.Errors++
	}
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
