package usecases

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// BuildArchitectureGraph constructs an ArchitectureGraph from a Project.
// This use case converts the hierarchical C4 model into a graph representation
// for easier querying, traversal, and relationship analysis.
type BuildArchitectureGraph struct {
	d2Parser D2Parser // Optional: if nil, only frontmatter relationships are used
}

// NewBuildArchitectureGraph creates a new BuildArchitectureGraph use case.
// D2 parsing is disabled by default (no D2Parser dependency).
func NewBuildArchitectureGraph() *BuildArchitectureGraph {
	return &BuildArchitectureGraph{
		d2Parser: nil,
	}
}

// NewBuildArchitectureGraphWithD2 creates a use case with D2 diagram parsing enabled.
// Relationships will be merged from both frontmatter and D2 files.
func NewBuildArchitectureGraphWithD2(d2Parser D2Parser) *BuildArchitectureGraph {
	return &BuildArchitectureGraph{
		d2Parser: d2Parser,
	}
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

	// Second pass: Union merge relationships from frontmatter and D2, then deduplicate.
	// Key: "sourceQualifiedID->targetQualifiedID" — used to deduplicate by (source, target).
	edgeSeen := make(map[string]bool)
	var edgeMu sync.Mutex // guards edgeSeen and graph.AddEdge

	addEdgeIfNew := func(sourceQualifiedID, targetQualifiedID, description string) {
		key := sourceQualifiedID + "->" + targetQualifiedID
		edgeMu.Lock()
		defer edgeMu.Unlock()
		if edgeSeen[key] {
			return // T036: deduplicate by (source, target)
		}
		edgeSeen[key] = true

		edge := &entities.GraphEdge{
			Source:      sourceQualifiedID,
			Target:      targetQualifiedID,
			Type:        "depends-on",
			Description: description,
			Weight:      0.8,
			Metadata:    map[string]string{"explicit": "true"},
		}
		_ = graph.AddEdge(edge)
	}

	// resolveTarget resolves a short or qualified component ID to a qualified graph node ID.
	resolveTarget := func(relatedID, sourceQualifiedID string) (string, bool) {
		targetQualifiedID, ok := graph.ResolveID(relatedID)
		if ok {
			return targetQualifiedID, true
		}
		qualifiedIDs, exists := graph.ShortIDMap[relatedID]
		if exists && len(qualifiedIDs) > 1 {
			var candidates []string
			for _, qid := range qualifiedIDs {
				if qid != sourceQualifiedID {
					candidates = append(candidates, qid)
				}
			}
			if len(candidates) == 1 {
				return candidates[0], true
			}
			return "", false // ambiguous
		}
		if graph.Nodes[relatedID] != nil {
			return relatedID, true // already qualified
		}
		return "", false // not found
	}

	// T035 source 1: frontmatter relationships
	for component, sourceQualifiedID := range componentToQualifiedID {
		for relatedID, relDescription := range component.Relationships {
			targetQualifiedID, ok := resolveTarget(relatedID, sourceQualifiedID)
			if !ok {
				continue
			}
			addEdgeIfNew(sourceQualifiedID, targetQualifiedID, relDescription)
		}
	}

	// T035 source 2: D2 file relationships (if parser is configured)
	// T037: Worker pool — up to 10 goroutines parse D2 files concurrently.
	if uc.d2Parser != nil {
		const maxWorkers = 10
		sem := make(chan struct{}, maxWorkers)
		var wg sync.WaitGroup

		for component, sourceQualifiedID := range componentToQualifiedID {
			if component.Path == "" {
				continue // no filesystem path, skip D2 parsing
			}
			// Capture loop variables for goroutine
			comp := component
			srcQID := sourceQualifiedID

			wg.Add(1)
			sem <- struct{}{} // acquire slot
			go func() {
				defer wg.Done()
				defer func() { <-sem }() // release slot

				d2Rels, err := uc.parseComponentD2(ctx, comp.Path)
				if err != nil {
					// T033: graceful degradation — log warning, continue
					return
				}
				for _, d2Rel := range d2Rels {
					// Only attribute relationships whose D2 source matches this component.
					// This prevents cross-component contamination when a D2 file covers
					// multiple nodes.
					if d2Rel.Source != comp.ID {
						continue
					}
					targetQualifiedID, ok := resolveTarget(d2Rel.Target, srcQID)
					if !ok {
						continue
					}
					addEdgeIfNew(srcQID, targetQualifiedID, d2Rel.Label)
				}
			}()
		}
		wg.Wait()
	}

	// Validate graph integrity
	if err := graph.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	return graph, nil
}

// parseComponentD2 reads the D2 diagram file for a component (if present) and
// returns the relationships defined there. Returns nil, nil when no D2 file exists.
func (uc *BuildArchitectureGraph) parseComponentD2(ctx context.Context, componentPath string) ([]entities.D2Relationship, error) {
	// Look for any .d2 file inside the component directory
	entries, err := os.ReadDir(componentPath)
	if err != nil {
		// Directory not accessible — treat as no D2 file (graceful degradation)
		return nil, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".d2") {
			d2Path := componentPath + "/" + name
			data, err := os.ReadFile(d2Path)
			if err != nil {
				return nil, err
			}
			return uc.d2Parser.ParseRelationships(ctx, string(data))
		}
	}

	return nil, nil // no D2 file found — valid state
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
