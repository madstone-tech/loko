package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

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

// Name returns the tool name.
func (t *QueryRelatedComponentsTool) Name() string { return "query_related_components" }

// Description returns the tool description.
func (t *QueryRelatedComponentsTool) Description() string {
	return "Query the architecture graph to find components that depend on or are depended upon by a given component"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *QueryRelatedComponentsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{"type": "string", "description": "Root directory of the project"},
			"system_id":    map[string]any{"type": "string", "description": "ID of the system"},
			"container_id": map[string]any{"type": "string", "description": "ID of the container"},
			"component_id": map[string]any{"type": "string", "description": "ID of the component to find related components for"},
		},
		"required": []string{"project_root", "system_id", "container_id", "component_id"},
	}
}

// Call executes the query_related_components tool.
func (t *QueryRelatedComponentsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	var typedArgs QueryRelatedComponentsArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}
	graph, err := getGraphFromProjectWithRel(ctx, t.repo, t.relRepo, typedArgs.ProjectRoot)
	if err != nil {
		return nil, err
	}
	componentID, err := resolveComponentInGraph(typedArgs.ComponentID, graph)
	if err != nil {
		return nil, err
	}
	deps := graph.GetDependencies(componentID)
	dependents := graph.GetDependents(componentID)
	return map[string]any{
		"component_id":     componentID,
		"dependencies":     graphNodesToMapList(deps),
		"dependents":       graphNodesToMapList(dependents),
		"dependency_count": len(deps),
		"dependent_count":  len(dependents),
	}, nil
}

// resolveComponentInGraph resolves a short or qualified component ID,
// returning a not-found error with suggestion if the component does not exist.
func resolveComponentInGraph(componentID string, graph *entities.ArchitectureGraph) (string, error) {
	if graph.GetNode(componentID) != nil {
		return componentID, nil
	}
	if qualifiedID, ok := graph.ResolveID(componentID); ok {
		return qualifiedID, nil
	}
	return "", notFoundError("component", componentID, suggestSlugID(componentID, graph))
}
