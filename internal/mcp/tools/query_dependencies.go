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
	return &QueryDependenciesTool{repo: repo, cache: cache}
}

// NewQueryDependenciesToolFull creates a new query_dependencies tool with relationship repo and cache.
func NewQueryDependenciesToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository, cache GraphCache) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo, relRepo: relRepo, cache: cache}
}

// Name returns the tool name.
func (t *QueryDependenciesTool) Name() string { return "query_dependencies" }

// Description returns the tool description.
func (t *QueryDependenciesTool) Description() string {
	return "Query the architecture graph to find dependencies. Omit component_id to query all dependencies of a container; provide component_id to query a specific component and optionally find a path to another component."
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *QueryDependenciesTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root":        map[string]any{"type": "string", "description": "Root directory of the project"},
			"system_id":           map[string]any{"type": "string", "description": "ID of the system (e.g., 'payment-service')"},
			"container_id":        map[string]any{"type": "string", "description": "ID of the container (e.g., 'api-server')"},
			"component_id":        map[string]any{"type": "string", "description": "Optional: ID of the component (e.g., 'auth'). Omit to get all dependencies of the container."},
			"target_component_id": map[string]any{"type": "string", "description": "Optional: ID of target component to find path to (only used when component_id is set)"},
		},
		"required": []string{"project_root", "system_id", "container_id"},
	}
}

// Call executes the query_dependencies tool.
func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]any) (any, error) {
	var typedArgs QueryDependenciesArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}
	project, systems, err := t.loadProjectAndSystems(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, err
	}
	_, targetContainer, err := t.resolveSystemContainer(ctx, typedArgs, systems)
	if err != nil {
		return nil, err
	}
	graphBuilder := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphBuilder.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}
	if typedArgs.ComponentID == "" {
		return t.queryContainerDependencies(targetContainer, graph), nil
	}
	return t.queryComponentDependencies(typedArgs, targetContainer, graph)
}

func (t *QueryDependenciesTool) loadProjectAndSystems(ctx context.Context, projectRoot string) (*entities.Project, []*entities.System, error) {
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project: %w", err)
	}
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load systems: %w", err)
	}
	return project, systems, nil
}

func (t *QueryDependenciesTool) resolveSystemContainer(ctx context.Context, args QueryDependenciesArgs, systems []*entities.System) (*entities.System, *entities.Container, error) {
	var sys *entities.System
	for _, s := range systems {
		if s.ID == args.SystemID {
			sys = s
			break
		}
	}
	if sys == nil {
		graph, _ := getGraphFromProject(ctx, t.repo, args.ProjectRoot)
		return nil, nil, notFoundError("system", args.SystemID, suggestSlugID(args.SystemID, graph))
	}
	var container *entities.Container
	for _, c := range sys.Containers {
		if c.ID == args.ContainerID {
			container = c
			break
		}
	}
	if container == nil {
		graph, _ := getGraphFromProject(ctx, t.repo, args.ProjectRoot)
		return nil, nil, notFoundError("container", args.ContainerID, suggestSlugID(args.ContainerID, graph))
	}
	return sys, container, nil
}

func (t *QueryDependenciesTool) queryComponentDependencies(args QueryDependenciesArgs, container *entities.Container, graph *entities.ArchitectureGraph) (any, error) {
	comp, exists := container.Components[args.ComponentID]
	if !exists {
		return nil, notFoundError("component", args.ComponentID, suggestSlugID(args.ComponentID, graph))
	}
	compQID, ok := graph.ResolveID(args.ComponentID)
	if !ok {
		if graph.GetNode(args.ComponentID) != nil {
			compQID = args.ComponentID
		} else {
			return nil, notFoundError("component", args.ComponentID, "")
		}
	}
	deps := graph.GetDependencies(compQID)
	result := map[string]any{
		"component":          map[string]any{"id": comp.ID, "name": comp.Name, "type": "component", "level": 3},
		"dependencies":       graphNodesToMapList(deps),
		"relationship_count": len(deps),
	}
	if args.TargetComponentID != "" {
		appendPathToTarget(result, graph, compQID, args.ComponentID, args.TargetComponentID)
	}
	return result, nil
}

// appendPathToTarget adds path_to_target to the result map when a target is specified.
func appendPathToTarget(result map[string]any, graph *entities.ArchitectureGraph, compQID, compID, targetID string) {
	targetQID := targetID
	if qid, ok := graph.ResolveID(targetID); ok {
		targetQID = qid
	}
	if path := graph.GetPath(compQID, targetQID); path != nil {
		result["path_to_target"] = graphNodesToMapList(path)
	} else {
		result["path_to_target"] = nil
		result["note"] = fmt.Sprintf("No path found from %s to %s", compID, targetID)
	}
}

// graphNodesToMapList converts a slice of graph nodes to JSON-friendly maps.
func graphNodesToMapList(nodes []*entities.GraphNode) []map[string]any {
	list := make([]map[string]any, len(nodes))
	for i, n := range nodes {
		list[i] = map[string]any{"id": n.ID, "name": n.Name, "type": n.Type, "level": n.Level}
	}
	return list
}

// queryContainerDependencies returns the union of all dependencies from all
// components in the container — the "container-level dependency view".
func (t *QueryDependenciesTool) queryContainerDependencies(container *entities.Container, graph *entities.ArchitectureGraph) map[string]any {
	seen := make(map[string]bool)
	var allDeps []map[string]any
	for shortCompID := range container.Components {
		qualifiedID, ok := graph.ResolveID(shortCompID)
		if !ok {
			continue
		}
		for _, dep := range graph.GetDependencies(qualifiedID) {
			if !seen[dep.ID] {
				seen[dep.ID] = true
				allDeps = append(allDeps, map[string]any{"id": dep.ID, "name": dep.Name, "type": dep.Type, "level": dep.Level})
			}
		}
	}
	if allDeps == nil {
		allDeps = []map[string]any{}
	}
	return map[string]any{
		"container":        map[string]any{"id": container.ID, "name": container.Name, "type": "container", "level": 2},
		"dependencies":     allDeps,
		"dependency_count": len(allDeps),
		"component_count":  len(container.Components),
	}
}
