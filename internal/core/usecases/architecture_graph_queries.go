// Package usecases — read-side query methods for BuildArchitectureGraph.
// See build_architecture_graph.go for the use case type and Execute method.
package usecases

import (
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// GetSystemGraph returns a subgraph containing only a specific system and its descendants.
func (uc *BuildArchitectureGraph) GetSystemGraph(
	graph *entities.ArchitectureGraph,
	systemID string,
) (*entities.ArchitectureGraph, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph cannot be nil")
	}

	systemNode := graph.GetNode(systemID)
	if systemNode == nil || systemNode.Type != "system" {
		return nil, fmt.Errorf("system %q not found", systemID)
	}

	subgraph := entities.NewArchitectureGraph()

	// Add system node
	if err := subgraph.AddNode(systemNode); err != nil {
		return nil, fmt.Errorf("failed to add system to subgraph: %w", err)
	}

	// Add all descendants
	descendants := graph.GetDescendants(systemID)
	for _, descendant := range descendants {
		if err := subgraph.AddNode(descendant); err != nil {
			return nil, fmt.Errorf("failed to add descendant to subgraph: %w", err)
		}
	}

	// Add relevant edges
	for _, sourceNode := range subgraph.Nodes {
		outgoing := graph.GetOutgoingEdges(sourceNode.ID)
		for _, edge := range outgoing {
			// Only add edge if target is in subgraph
			if subgraph.Nodes[edge.Target] != nil {
				if err := subgraph.AddEdge(edge); err != nil {
					return nil, fmt.Errorf("failed to add edge to subgraph: %w", err)
				}
			}
		}
	}

	return subgraph, nil
}

// AnalyzeDependencies analyzes dependency patterns in the graph.
// Returns a strongly-typed report of isolated components, coupling metrics, etc.
func (uc *BuildArchitectureGraph) AnalyzeDependencies(
	graph *entities.ArchitectureGraph,
) *entities.DependencyReport {
	report := entities.NewDependencyReport()

	// Count nodes by level
	systems := graph.GetNodesByLevel(1)
	containers := graph.GetNodesByLevel(2)
	components := graph.GetNodesByLevel(3)

	report.SystemsCount = len(systems)
	report.ContainersCount = len(containers)
	report.ComponentsCount = len(components)
	report.TotalNodes = graph.Size()
	report.TotalEdges = graph.EdgeCount()

	// Find isolated components (no dependencies or dependents)
	for _, node := range components {
		incoming := graph.GetIncomingEdges(node.ID)
		outgoing := graph.GetOutgoingEdges(node.ID)

		if len(incoming) == 0 && len(outgoing) == 0 {
			report.IsolatedComponents = append(report.IsolatedComponents, node.ID)
		}
	}

	// Find highly coupled components (many dependencies)
	for _, node := range components {
		deps := graph.GetDependencies(node.ID)
		if len(deps) > 2 {
			report.HighlyCoupledComponents[node.ID] = len(deps)
		}
	}

	// Find central components (many dependents)
	for _, node := range components {
		dependents := graph.GetDependents(node.ID)
		if len(dependents) > 2 {
			report.CentralComponents[node.ID] = len(dependents)
		}
	}

	return report
}
