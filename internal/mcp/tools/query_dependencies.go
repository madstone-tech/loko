package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// GraphCache interface for graph caching.
type GraphCache interface {
	Get(projectRoot string) (*entities.ArchitectureGraph, bool)
	Set(projectRoot string, graph *entities.ArchitectureGraph)
	// Invalidate removes the cached graph so the next access triggers a rebuild.
	Invalidate(projectRoot string)
}

// QueryDependenciesTool queries the architecture graph to find component dependencies.
type QueryDependenciesTool struct {
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository
	cache   GraphCache
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

	systems, err := t.repo.ListSystems(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	var targetContainer *entities.Container
	for _, s := range systems {
		if s.ID == typedArgs.SystemID {
			for _, c := range s.Containers {
				if c.ID == typedArgs.ContainerID {
					targetContainer = c
					break
				}
			}
			break
		}
	}
	if targetContainer == nil {
		graph, _ := getGraphFromProject(ctx, t.repo, typedArgs.ProjectRoot)
		return nil, notFoundError("container", typedArgs.ContainerID, suggestSlugID(typedArgs.ContainerID, graph))
	}

	graph, err := getGraphFromProjectWithRel(ctx, t.repo, t.relRepo, typedArgs.ProjectRoot)
	if err != nil {
		return nil, err
	}

	if typedArgs.ComponentID == "" {
		return queryContainerDependencies(targetContainer, graph), nil
	}
	return queryComponentDependencies(typedArgs, targetContainer, graph)
}
