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
	encoder usecases.OutputEncoder
}

// NewQueryDependenciesTool creates a new query_dependencies tool.
func NewQueryDependenciesTool(repo usecases.ProjectRepository, encoder usecases.OutputEncoder) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo, encoder: encoder}
}

// NewQueryDependenciesToolWithCache creates a new query_dependencies tool with caching support.
func NewQueryDependenciesToolWithCache(repo usecases.ProjectRepository, cache GraphCache, encoder usecases.OutputEncoder) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo, cache: cache, encoder: encoder}
}

// NewQueryDependenciesToolFull creates a new query_dependencies tool with relationship repo and cache.
func NewQueryDependenciesToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository, cache GraphCache, encoder usecases.OutputEncoder) *QueryDependenciesTool {
	return &QueryDependenciesTool{repo: repo, relRepo: relRepo, cache: cache, encoder: encoder}
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
			"system_id":           map[string]any{"type": "string", "description": "ID of the system"},
			"container_id":        map[string]any{"type": "string", "description": "ID of the container"},
			"component_id":        map[string]any{"type": "string", "description": "ID of the component (optional — omit to query all container dependencies)"},
			"target_component_id": map[string]any{"type": "string", "description": "Optional: find dependency path to this component"},
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"toon", "json"},
				"default":     "toon",
				"description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging",
			},
		},
		"required": []string{"project_root", "system_id", "container_id"},
	}
}

// Call executes the query_dependencies tool.
func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]any) (any, error) {
	format, err := getFormat(args)
	if err != nil {
		return nil, err
	}
	result, err := t.query(ctx, args)
	if err != nil {
		return nil, err
	}
	return formatResponse(result, format, t.encoder)
}

func (t *QueryDependenciesTool) query(ctx context.Context, args map[string]any) (map[string]any, error) {
	var typedArgs QueryDependenciesArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}

	targetContainer, err := t.findContainer(ctx, typedArgs)
	if err != nil {
		return nil, err
	}

	graph, err := getGraphFromProjectWithRel(ctx, t.repo, t.relRepo, typedArgs.ProjectRoot)
	if err != nil {
		return nil, err
	}

	if typedArgs.ComponentID == "" {
		return queryContainerDependencies(targetContainer, graph), nil
	}
	result, err := queryComponentDependencies(typedArgs, targetContainer, graph)
	if err != nil {
		return nil, err
	}
	return result.(map[string]any), nil
}

func (t *QueryDependenciesTool) findContainer(ctx context.Context, args QueryDependenciesArgs) (*entities.Container, error) {
	systems, err := t.repo.ListSystems(ctx, args.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}
	for _, s := range systems {
		if s.ID == args.SystemID {
			for _, c := range s.Containers {
				if c.ID == args.ContainerID {
					return c, nil
				}
			}
			break
		}
	}
	graph, _ := getGraphFromProject(ctx, t.repo, args.ProjectRoot)
	return nil, notFoundError("container", args.ContainerID, suggestSlugID(args.ContainerID, graph))
}
