package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// BuildArchitectureGraph constructs an ArchitectureGraph from a Project.
// This use case converts the hierarchical C4 model into a graph representation
// for easier querying, traversal, and relationship analysis.
type BuildArchitectureGraph struct {
	// Could add dependencies here if needed
}

// NewBuildArchitectureGraph creates a new BuildArchitectureGraph use case.
func NewBuildArchitectureGraph() *BuildArchitectureGraph {
	return &BuildArchitectureGraph{}
}

// Execute builds an ArchitectureGraph from the given project and systems.
//
// The graph includes:
// - Nodes for all systems, containers, and components
// - Hierarchy edges (parent-child relationships)
// - Relationship edges (component dependencies)
//
// C4 Level mapping:
// - Level 1: Systems
// - Level 2: Containers
// - Level 3: Components
func (uc *BuildArchitectureGraph) Execute(
	ctx context.Context,
	project *entities.Project,
	systems []*entities.System,
) (*entities.ArchitectureGraph, error) {
	if project == nil {
		return nil, fmt.Errorf("project cannot be nil")
	}

	graph := entities.NewArchitectureGraph()

	// First pass: Add all nodes (systems, containers, components)
	// Track component entities and their qualified node IDs for relationship resolution
	componentToQualifiedID := make(map[*entities.Component]string)

	for _, system := range systems {
		if system == nil {
			continue
		}

		systemNode := &entities.GraphNode{
			ID:          entities.QualifiedNodeID("system", system.ID, "", ""),
			Type:        "system",
			Name:        system.Name,
			Description: system.Description,
			Level:       1,
			Data:        system,
			Metadata:    make(map[string]string),
		}

		if err := graph.AddNode(systemNode); err != nil {
			return nil, fmt.Errorf("failed to add system node: %w", err)
		}

		// Add container nodes
		for _, container := range system.Containers {
			if container == nil {
				continue
			}

			containerNode := &entities.GraphNode{
				ID:          entities.QualifiedNodeID("container", system.ID, container.ID, ""),
				Type:        "container",
				Name:        container.Name,
				Description: container.Description,
				Level:       2,
				ParentID:    entities.QualifiedNodeID("system", system.ID, "", ""),
				Data:        container,
				Metadata: map[string]string{
					"technology": container.Technology,
				},
			}

			if err := graph.AddNode(containerNode); err != nil {
				return nil, fmt.Errorf("failed to add container node: %w", err)
			}

			// Add component nodes
			for _, component := range container.Components {
				if component == nil {
					continue
				}

				componentNode := &entities.GraphNode{
					ID:          entities.QualifiedNodeID("component", system.ID, container.ID, component.ID),
					Type:        "component",
					Name:        component.Name,
					Description: component.Description,
					Level:       3,
					ParentID:    entities.QualifiedNodeID("container", system.ID, container.ID, ""),
					Data:        component,
					Metadata: map[string]string{
						"technology": component.Technology,
					},
				}

				if err := graph.AddNode(componentNode); err != nil {
					return nil, fmt.Errorf("failed to add component node: %w", err)
				}

				// Track component and its qualified ID for relationship processing in second pass
				componentToQualifiedID[component] = componentNode.ID
			}
		}
	}

	// Second pass: Add relationship edges after all nodes are in the graph
	for component, sourceQualifiedID := range componentToQualifiedID {
		for relatedID, relDescription := range component.Relationships {
			// Resolve short ID to qualified ID
			targetQualifiedID, ok := graph.ResolveID(relatedID)
			if !ok {
				// Resolution failed - could be ambiguous or not found
				// Check if it's in ShortIDMap (indicating ambiguity)
				qualifiedIDs, exists := graph.ShortIDMap[relatedID]
				if exists && len(qualifiedIDs) > 1 {
					// Ambiguous - filter out self-references
					var candidates []string
					for _, qid := range qualifiedIDs {
						if qid != sourceQualifiedID {
							candidates = append(candidates, qid)
						}
					}

					if len(candidates) == 1 {
						// After filtering self-reference, only one candidate remains
						targetQualifiedID = candidates[0]
					} else if len(candidates) == 0 {
						// Only self-reference - skip (no external dependency)
						continue
					} else {
						// Still ambiguous after filtering - skip with warning
						// TODO: Log warning about ambiguous relationship
						continue
					}
				} else if graph.Nodes[relatedID] != nil {
					// ID might already be qualified - try using as-is
					targetQualifiedID = relatedID
				} else {
					// Not found - skip relationship
					continue
				}
			}

			edge := &entities.GraphEdge{
				Source:      sourceQualifiedID,
				Target:      targetQualifiedID,
				Type:        "depends-on",
				Description: relDescription,
				Weight:      0.8, // Default coupling weight
				Metadata: map[string]string{
					"explicit": "true",
				},
			}

			// Ignore errors on individual edges - continue building graph
			_ = graph.AddEdge(edge)
		}
	}

	// Validate graph integrity
	if err := graph.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	return graph, nil
}

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
