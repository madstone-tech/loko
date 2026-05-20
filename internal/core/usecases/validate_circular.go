// Package usecases — circular dependency detection for ValidateArchitecture.
// See validate_architecture.go for the use case type, Execute, and ArchitectureReport.
package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

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
