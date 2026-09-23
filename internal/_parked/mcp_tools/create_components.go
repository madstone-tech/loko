package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateComponentsTool creates multiple components in a container in a single operation.
type CreateComponentsTool struct {
	repo usecases.ProjectRepository
}

// NewCreateComponentsTool creates a new create_components tool.
func NewCreateComponentsTool(repo usecases.ProjectRepository) *CreateComponentsTool {
	return &CreateComponentsTool{repo: repo}
}

// Name returns the tool name.
func (t *CreateComponentsTool) Name() string { return "create_components" }

// Description returns the tool description.
func (t *CreateComponentsTool) Description() string {
	return "Create multiple components in a container in a single operation"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *CreateComponentsTool) InputSchema() map[string]any { return createComponentsSchema }

// Call executes the create_components tool, scaffolding each component individually.
// Individual component failures do not abort the batch.
func (t *CreateComponentsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	systemName, _ := args["system_name"].(string)
	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}
	containerName, _ := args["container_name"].(string)
	if containerName == "" {
		return nil, fmt.Errorf("container_name is required")
	}
	componentsIface, ok := args["components"].([]any)
	if !ok || len(componentsIface) == 0 {
		return nil, fmt.Errorf("components array must have at least one item")
	}
	created, failed, results := t.scaffoldAll(ctx, projectRoot, systemName, containerName, componentsIface)
	return map[string]any{"created": created, "failed": failed, "results": results}, nil
}

func (t *CreateComponentsTool) scaffoldAll(ctx context.Context, projectRoot, systemName, containerName string, items []any) (created, failed int, results []map[string]any) {
	results = make([]map[string]any, 0, len(items))
	for i, compIface := range items {
		compMap, ok := compIface.(map[string]any)
		if !ok {
			results = append(results, map[string]any{"status": "error", "error": fmt.Sprintf("component %d is not a valid object", i)})
			failed++
			continue
		}
		item, entityID := scaffoldOneComponent(ctx, t.repo, projectRoot, systemName, containerName, compMap)
		if entityID != "" {
			item["id"] = entityID
			created++
		} else {
			failed++
		}
		results = append(results, item)
	}
	return created, failed, results
}
