package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// QueryDependenciesTool queries the architecture graph to find component dependencies.
// It returns the dependency chain from a source component to a target component.
type QueryDependenciesTool struct {
	repo usecases.ProjectRepository
}

// NewQueryDependenciesTool creates a new query_dependencies tool.
func NewQueryDependenciesTool(repo usecases.ProjectRepository) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo}
}

func (t *QueryDependenciesTool) Name() string {
	return "query_dependencies"
}

func (t *QueryDependenciesTool) Description() string {
	return "Query the architecture graph to find dependencies of a component and the dependency path to another component"
}

func (t *QueryDependenciesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the system (e.g., 'payment-service')",
			},
			"container_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the container (e.g., 'api-server')",
			},
			"component_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the component (e.g., 'auth')",
			},
			"target_component_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: ID of target component to find path to",
			},
		},
		"required": []string{"project_root", "system_id", "container_id", "component_id"},
	}
}

func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	systemID, _ := args["system_id"].(string)
	containerID, _ := args["container_id"].(string)
	componentID, _ := args["component_id"].(string)
	targetComponentID, _ := args["target_component_id"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Find the specified system
	var targetSystem *entities.System
	for _, sys := range systems {
		if sys.ID == systemID {
			targetSystem = sys
			break
		}
	}
	if targetSystem == nil {
		return nil, fmt.Errorf("system %q not found", systemID)
	}

	// Find the specified container
	var targetContainer *entities.Container
	for _, container := range targetSystem.Containers {
		if container.ID == containerID {
			targetContainer = container
			break
		}
	}
	if targetContainer == nil {
		return nil, fmt.Errorf("container %q not found in system %q", containerID, systemID)
	}

	// Find the specified component
	comp, exists := targetContainer.Components[componentID]
	if !exists {
		return nil, fmt.Errorf("component %q not found in container %q", componentID, containerID)
	}

	// Build architecture graph
	graphBuilder := usecases.NewBuildArchitectureGraph()
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Get direct dependencies
	deps := graph.GetDependencies(componentID)
	depList := make([]map[string]interface{}, len(deps))
	for i, dep := range deps {
		depList[i] = map[string]interface{}{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	result := map[string]interface{}{
		"component": map[string]interface{}{
			"id":    comp.ID,
			"name":  comp.Name,
			"type":  "component",
			"level": 3,
		},
		"dependencies":       depList,
		"relationship_count": len(comp.Relationships),
	}

	// If target component specified, find path
	if targetComponentID != "" {
		path := graph.GetPath(componentID, targetComponentID)
		if path != nil {
			pathList := make([]map[string]interface{}, len(path))
			for i, node := range path {
				pathList[i] = map[string]interface{}{
					"id":   node.ID,
					"name": node.Name,
					"type": node.Type,
				}
			}
			result["path_to_target"] = pathList
		} else {
			result["path_to_target"] = nil
			result["note"] = fmt.Sprintf("No path found from %s to %s", componentID, targetComponentID)
		}
	}

	return result, nil
}

// QueryRelatedComponentsTool finds related components based on relationships.
type QueryRelatedComponentsTool struct {
	repo usecases.ProjectRepository
}

// NewQueryRelatedComponentsTool creates a new query_related_components tool.
func NewQueryRelatedComponentsTool(repo usecases.ProjectRepository) *QueryRelatedComponentsTool {
	return &QueryRelatedComponentsTool{repo: repo}
}

func (t *QueryRelatedComponentsTool) Name() string {
	return "query_related_components"
}

func (t *QueryRelatedComponentsTool) Description() string {
	return "Query the architecture graph to find components that depend on or are depended upon by a given component"
}

func (t *QueryRelatedComponentsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the system",
			},
			"container_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the container",
			},
			"component_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the component to find related components for",
			},
		},
		"required": []string{"project_root", "system_id", "container_id", "component_id"},
	}
}

func (t *QueryRelatedComponentsTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	_ = args["system_id"]    // Not used, but required in schema
	_ = args["container_id"] // Not used, but required in schema
	componentID, _ := args["component_id"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Build architecture graph
	graphBuilder := usecases.NewBuildArchitectureGraph()
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Get dependencies (outgoing edges)
	deps := graph.GetDependencies(componentID)
	depList := make([]map[string]interface{}, len(deps))
	for i, dep := range deps {
		depList[i] = map[string]interface{}{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	// Get dependents (incoming edges)
	dependents := graph.GetDependents(componentID)
	dependentList := make([]map[string]interface{}, len(dependents))
	for i, dep := range dependents {
		dependentList[i] = map[string]interface{}{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	return map[string]interface{}{
		"component_id":     componentID,
		"dependencies":     depList,
		"dependents":       dependentList,
		"dependency_count": len(depList),
		"dependent_count":  len(dependentList),
	}, nil
}

// AnalyzeCouplingTool analyzes coupling metrics in the architecture.
type AnalyzeCouplingTool struct {
	repo usecases.ProjectRepository
}

// NewAnalyzeCouplingTool creates a new analyze_coupling tool.
func NewAnalyzeCouplingTool(repo usecases.ProjectRepository) *AnalyzeCouplingTool {
	return &AnalyzeCouplingTool{repo: repo}
}

func (t *AnalyzeCouplingTool) Name() string {
	return "analyze_coupling"
}

func (t *AnalyzeCouplingTool) Description() string {
	return "Analyze coupling metrics for a system, identifying highly coupled and central components"
}

func (t *AnalyzeCouplingTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the system to analyze (optional - analyzes whole project if not specified)",
			},
		},
	}
}

func (t *AnalyzeCouplingTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	systemID, _ := args["system_id"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Build architecture graph
	graphBuilder := usecases.NewBuildArchitectureGraph()
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Get subgraph if system specified
	var targetGraph *entities.ArchitectureGraph
	if systemID != "" {
		subgraph, err := graphBuilder.GetSystemGraph(graph, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get system graph: %w", err)
		}
		targetGraph = subgraph
	} else {
		targetGraph = graph
	}

	// Analyze dependencies
	report := graphBuilder.AnalyzeDependencies(targetGraph)

	return map[string]interface{}{
		"total_systems":             report["systems_count"],
		"total_components":          report["components_count"],
		"isolated_components":       report["isolated_components"],
		"highly_coupled_components": report["highly_coupled_components"],
		"central_components":        report["central_components"],
		"note":                      "Isolated components have no relationships; Central components have high in-degree (many dependents)",
	}, nil
}
