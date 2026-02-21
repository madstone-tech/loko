package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// UpdateComponentTool updates an existing component's metadata.
type UpdateComponentTool struct {
	repo usecases.ProjectRepository
}

// NewUpdateComponentTool creates a new update_component tool.
func NewUpdateComponentTool(repo usecases.ProjectRepository) *UpdateComponentTool {
	return &UpdateComponentTool{repo: repo}
}

func (t *UpdateComponentTool) Name() string {
	return "update_component"
}

func (t *UpdateComponentTool) Description() string {
	return "Update an existing component's metadata (description, technology, tags)"
}

func (t *UpdateComponentTool) InputSchema() map[string]any { return updateComponentSchema }

// Call executes the update component tool.
func (t *UpdateComponentTool) Call(ctx context.Context, args map[string]any) (any, error) {
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

	componentName, _ := args["component_name"].(string)
	if componentName == "" {
		return nil, fmt.Errorf("component_name is required")
	}

	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)
	componentID := entities.NormalizeName(componentName)

	// First try to load the component with the provided IDs
	component, err := t.repo.LoadComponent(ctx, projectRoot, systemID, containerID, componentID)
	if err != nil {
		// If that fails, try to get a suggestion for the error message
		graph, graphErr := getGraphFromProject(ctx, t.repo, projectRoot)
		if graphErr != nil {
			// If we can't build a graph, return the original error
			return nil, fmt.Errorf("failed to load component %q: %w", componentID, err)
		}

		// Try to find a suggestion using the graph
		suggestion := suggestSlugID(componentName, graph)
		return nil, notFoundError("component", componentName, suggestion)
	}

	// Update only non-empty fields
	if desc, ok := args["description"].(string); ok && desc != "" {
		component.Description = desc
	}
	if tech, ok := args["technology"].(string); ok && tech != "" {
		component.Technology = tech
	}
	if v, ok := args["tags"].([]any); ok {
		component.Tags = convertInterfaceSlice(v)
	}

	// Save
	if err := t.repo.SaveComponent(ctx, projectRoot, systemID, containerID, component); err != nil {
		return nil, fmt.Errorf("failed to save component: %w", err)
	}

	return map[string]any{
		"component": map[string]any{
			"id":          component.ID,
			"name":        component.Name,
			"description": component.Description,
			"technology":  component.Technology,
			"tags":        component.Tags,
		},
		"message": fmt.Sprintf("Component %q updated", component.Name),
	}, nil
}
