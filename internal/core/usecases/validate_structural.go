// Package usecases — structural quality checks for ValidateArchitecture.
// See validate_architecture.go for the use case type, Execute, and ArchitectureReport.
package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// checkIsolatedComponents finds components with no relationships.
// When the graph has zero edges (no relationships defined), no findings are emitted —
// this suppresses false alarms on freshly initialized projects (FR-012).
func (uc *ValidateArchitecture) checkIsolatedComponents(
	graph *entities.ArchitectureGraph,
	report *ArchitectureReport,
) {
	if graph.EdgeCount() == 0 {
		return
	}

	isolated := make([]string, 0)

	// Only check components (level 3) - systems and containers don't have relationship edges
	for nodeID, node := range graph.Nodes {
		if node.Type != "component" {
			continue // Skip systems and containers
		}

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

	// Only check components (level 3) - systems and containers don't have relationship edges
	for nodeID, node := range graph.Nodes {
		if node.Type != "component" {
			continue // Skip systems and containers
		}

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
