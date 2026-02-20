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
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository // Optional: loads relationships.toml into graph
	cache   GraphCache                      // Optional cache for graph reuse
}

// GraphCache interface for graph caching.
type GraphCache interface {
	Get(projectRoot string) (*entities.ArchitectureGraph, bool)
	Set(projectRoot string, graph *entities.ArchitectureGraph)
	// Invalidate removes the cached graph so the next access triggers a rebuild.
	Invalidate(projectRoot string)
}

// NewQueryDependenciesTool creates a new query_dependencies tool.
func NewQueryDependenciesTool(repo usecases.ProjectRepository) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo}
}

// NewQueryDependenciesToolWithCache creates a new query_dependencies tool with caching support.
func NewQueryDependenciesToolWithCache(repo usecases.ProjectRepository, cache GraphCache) *QueryDependenciesTool {
	return &QueryDependenciesTool{
		repo:  repo,
		cache: cache,
	}
}

// NewQueryDependenciesToolFull creates a new query_dependencies tool with relationship repo and cache.
func NewQueryDependenciesToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository, cache GraphCache) *QueryDependenciesTool {
	return &QueryDependenciesTool{
		repo:    repo,
		relRepo: relRepo,
		cache:   cache,
	}
}

func (t *QueryDependenciesTool) Name() string {
	return "query_dependencies"
}

func (t *QueryDependenciesTool) Description() string {
	return "Query the architecture graph to find dependencies. Omit component_id to query all dependencies of a container; provide component_id to query a specific component and optionally find a path to another component."
}

func (t *QueryDependenciesTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]any{
				"type":        "string",
				"description": "ID of the system (e.g., 'payment-service')",
			},
			"container_id": map[string]any{
				"type":        "string",
				"description": "ID of the container (e.g., 'api-server')",
			},
			"component_id": map[string]any{
				"type":        "string",
				"description": "Optional: ID of the component (e.g., 'auth'). Omit to get all dependencies of the container.",
			},
			"target_component_id": map[string]any{
				"type":        "string",
				"description": "Optional: ID of target component to find path to (only used when component_id is set)",
			},
		},
		"required": []string{"project_root", "system_id", "container_id"},
	}
}

func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// Convert map to typed struct for compile-time type safety
	var typedArgs QueryDependenciesArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Apply defaults
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Find the specified system
	var targetSystem *entities.System
	for _, sys := range systems {
		if sys.ID == typedArgs.SystemID {
			targetSystem = sys
			break
		}
	}
	if targetSystem == nil {
		// Try to get a suggestion for the error message
		graph, graphErr := getGraphFromProject(ctx, t.repo, typedArgs.ProjectRoot)
		if graphErr != nil {
			// If we can't build a graph, return the original error
			return nil, fmt.Errorf("system %q not found", typedArgs.SystemID)
		}

		// Try to find a suggestion using the graph
		suggestion := suggestSlugID(typedArgs.SystemID, graph)
		return nil, notFoundError("system", typedArgs.SystemID, suggestion)
	}

	// Find the specified container
	var targetContainer *entities.Container
	for _, container := range targetSystem.Containers {
		if container.ID == typedArgs.ContainerID {
			targetContainer = container
			break
		}
	}
	if targetContainer == nil {
		// Try to get a suggestion for the error message
		graph, graphErr := getGraphFromProject(ctx, t.repo, typedArgs.ProjectRoot)
		if graphErr != nil {
			// If we can't build a graph, return the original error
			return nil, fmt.Errorf("container %q not found in system %q", typedArgs.ContainerID, typedArgs.SystemID)
		}

		// Try to find a suggestion using the graph
		suggestion := suggestSlugID(typedArgs.ContainerID, graph)
		return nil, notFoundError("container", typedArgs.ContainerID, suggestion)
	}

	// Build architecture graph (includes relationships.toml when relRepo is wired).
	graphBuilder := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// If component_id is omitted, return all dependencies of the container.
	if typedArgs.ComponentID == "" {
		return t.queryContainerDependencies(targetContainer, graph), nil
	}

	// Find the specified component
	comp, exists := targetContainer.Components[typedArgs.ComponentID]
	if !exists {
		suggestion := suggestSlugID(typedArgs.ComponentID, graph)
		return nil, notFoundError("component", typedArgs.ComponentID, suggestion)
	}

	// Get direct dependencies
	deps := graph.GetDependencies(typedArgs.ComponentID)
	depList := make([]map[string]any, len(deps))
	for i, dep := range deps {
		depList[i] = map[string]any{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	result := map[string]any{
		"component": map[string]any{
			"id":    comp.ID,
			"name":  comp.Name,
			"type":  "component",
			"level": 3,
		},
		"dependencies":       depList,
		"relationship_count": len(comp.Relationships),
	}

	// If target component specified, find path
	if typedArgs.TargetComponentID != "" {
		path := graph.GetPath(typedArgs.ComponentID, typedArgs.TargetComponentID)
		if path != nil {
			pathList := make([]map[string]any, len(path))
			for i, node := range path {
				pathList[i] = map[string]any{
					"id":   node.ID,
					"name": node.Name,
					"type": node.Type,
				}
			}
			result["path_to_target"] = pathList
		} else {
			result["path_to_target"] = nil
			result["note"] = fmt.Sprintf("No path found from %s to %s", typedArgs.ComponentID, typedArgs.TargetComponentID)
		}
	}

	return result, nil
}

// queryContainerDependencies returns the union of all dependencies from all
// components in the container — the "container-level dependency view".
func (t *QueryDependenciesTool) queryContainerDependencies(
	container *entities.Container,
	graph *entities.ArchitectureGraph,
) map[string]any {
	seen := make(map[string]bool)
	var allDeps []map[string]any

	for compID := range container.Components {
		for _, dep := range graph.GetDependencies(compID) {
			if !seen[dep.ID] {
				seen[dep.ID] = true
				allDeps = append(allDeps, map[string]any{
					"id":    dep.ID,
					"name":  dep.Name,
					"type":  dep.Type,
					"level": dep.Level,
				})
			}
		}
	}

	if allDeps == nil {
		allDeps = []map[string]any{}
	}

	return map[string]any{
		"container": map[string]any{
			"id":    container.ID,
			"name":  container.Name,
			"type":  "container",
			"level": 2,
		},
		"dependencies":     allDeps,
		"dependency_count": len(allDeps),
		"component_count":  len(container.Components),
	}
}

// QueryRelatedComponentsTool finds related components based on relationships.
type QueryRelatedComponentsTool struct {
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository // Optional: loads relationships.toml into graph
}

// NewQueryRelatedComponentsTool creates a new query_related_components tool.
func NewQueryRelatedComponentsTool(repo usecases.ProjectRepository) *QueryRelatedComponentsTool {
	return &QueryRelatedComponentsTool{repo: repo}
}

// NewQueryRelatedComponentsToolFull creates a new query_related_components tool with relationship repo.
func NewQueryRelatedComponentsToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository) *QueryRelatedComponentsTool {
	return &QueryRelatedComponentsTool{repo: repo, relRepo: relRepo}
}

func (t *QueryRelatedComponentsTool) Name() string {
	return "query_related_components"
}

func (t *QueryRelatedComponentsTool) Description() string {
	return "Query the architecture graph to find components that depend on or are depended upon by a given component"
}

func (t *QueryRelatedComponentsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]any{
				"type":        "string",
				"description": "ID of the system",
			},
			"container_id": map[string]any{
				"type":        "string",
				"description": "ID of the container",
			},
			"component_id": map[string]any{
				"type":        "string",
				"description": "ID of the component to find related components for",
			},
		},
		"required": []string{"project_root", "system_id", "container_id", "component_id"},
	}
}

func (t *QueryRelatedComponentsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// Convert map to typed struct for compile-time type safety
	var typedArgs QueryRelatedComponentsArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Apply defaults
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Build architecture graph (includes relationships.toml when relRepo is wired).
	graphBuilder := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Check if component exists in the graph
	componentNode := graph.GetNode(typedArgs.ComponentID)
	if componentNode == nil {
		// Try to resolve using ShortIDMap
		if qualifiedID, ok := graph.ResolveID(typedArgs.ComponentID); ok {
			// Use the resolved ID
			typedArgs.ComponentID = qualifiedID
		} else {
			// Component not found, try to get a suggestion
			suggestion := suggestSlugID(typedArgs.ComponentID, graph)
			return nil, notFoundError("component", typedArgs.ComponentID, suggestion)
		}
	}

	// Get dependencies (outgoing edges)
	deps := graph.GetDependencies(typedArgs.ComponentID)
	depList := make([]map[string]any, len(deps))
	for i, dep := range deps {
		depList[i] = map[string]any{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	// Get dependents (incoming edges)
	dependents := graph.GetDependents(typedArgs.ComponentID)
	dependentList := make([]map[string]any, len(dependents))
	for i, dep := range dependents {
		dependentList[i] = map[string]any{
			"id":    dep.ID,
			"name":  dep.Name,
			"type":  dep.Type,
			"level": dep.Level,
		}
	}

	return map[string]any{
		"component_id":     typedArgs.ComponentID,
		"dependencies":     depList,
		"dependents":       dependentList,
		"dependency_count": len(depList),
		"dependent_count":  len(dependentList),
	}, nil
}

// AnalyzeCouplingTool analyzes coupling metrics in the architecture.
type AnalyzeCouplingTool struct {
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository // Optional: loads relationships.toml into graph
}

// NewAnalyzeCouplingTool creates a new analyze_coupling tool.
func NewAnalyzeCouplingTool(repo usecases.ProjectRepository) *AnalyzeCouplingTool {
	return &AnalyzeCouplingTool{repo: repo}
}

// NewAnalyzeCouplingToolFull creates a new analyze_coupling tool with relationship repo.
func NewAnalyzeCouplingToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository) *AnalyzeCouplingTool {
	return &AnalyzeCouplingTool{repo: repo, relRepo: relRepo}
}

func (t *AnalyzeCouplingTool) Name() string {
	return "analyze_coupling"
}

func (t *AnalyzeCouplingTool) Description() string {
	return "Analyze coupling metrics for a system, identifying highly coupled and central components"
}

func (t *AnalyzeCouplingTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_id": map[string]any{
				"type":        "string",
				"description": "ID of the system to analyze (optional - analyzes whole project if not specified)",
			},
		},
	}
}

func (t *AnalyzeCouplingTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// Convert map to typed struct for compile-time type safety
	var typedArgs AnalyzeCouplingArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Apply defaults
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}

	// Load project and systems
	project, err := t.repo.LoadProject(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// Build architecture graph (includes relationships.toml when relRepo is wired).
	graphBuilder := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Get subgraph if system specified
	var targetGraph *entities.ArchitectureGraph
	if typedArgs.SystemID != "" {
		subgraph, err := graphBuilder.GetSystemGraph(graph, typedArgs.SystemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get system graph: %w", err)
		}
		targetGraph = subgraph
	} else {
		targetGraph = graph
	}

	// Analyze dependencies
	report := graphBuilder.AnalyzeDependencies(targetGraph)

	return map[string]any{
		"total_systems":             report.SystemsCount,
		"total_components":          report.ComponentsCount,
		"isolated_components":       report.IsolatedComponents,
		"highly_coupled_components": report.HighlyCoupledComponents,
		"central_components":        report.CentralComponents,
		"note":                      "Isolated components have no relationships; Central components have high in-degree (many dependents)",
	}, nil
}
