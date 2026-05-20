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

// Name returns the tool name.
func (t *UpdateComponentTool) Name() string { return "update_component" }

// Description returns the tool description.
func (t *UpdateComponentTool) Description() string {
	return "Update an existing component's metadata (description, technology, tags)"
}

// InputSchema returns the JSON schema for this tool's inputs.
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
	component, err := t.loadComponent(ctx, projectRoot, systemName, containerName, componentName)
	if err != nil {
		return nil, err
	}
	applyComponentUpdates(component, args)
	return t.saveComponent(ctx, projectRoot, systemName, containerName, component)
}

func (t *UpdateComponentTool) saveComponent(ctx context.Context, projectRoot, systemName, containerName string, component *entities.Component) (any, error) {
	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)
	if err := t.repo.SaveComponent(ctx, projectRoot, systemID, containerID, component); err != nil {
		return nil, fmt.Errorf("failed to save component: %w", err)
	}
	return map[string]any{
		"component": map[string]any{
			"id": component.ID, "name": component.Name,
			"description": component.Description, "technology": component.Technology, "tags": component.Tags,
		},
		"message": fmt.Sprintf("Component %q updated", component.Name),
	}, nil
}

func (t *UpdateComponentTool) loadComponent(ctx context.Context, projectRoot, systemName, containerName, componentName string) (*entities.Component, error) {
	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)
	componentID := entities.NormalizeName(componentName)
	component, err := t.repo.LoadComponent(ctx, projectRoot, systemID, containerID, componentID)
	if err != nil {
		graph, graphErr := getGraphFromProject(ctx, t.repo, projectRoot)
		if graphErr != nil {
			return nil, fmt.Errorf("failed to load component %q: %w", componentID, err)
		}
		return nil, notFoundError("component", componentName, suggestSlugID(componentName, graph))
	}
	return component, nil
}

// applyComponentUpdates patches non-empty fields on a Component entity from raw args.
func applyComponentUpdates(component *entities.Component, args map[string]any) {
	if desc, ok := args["description"].(string); ok && desc != "" {
		component.Description = desc
	}
	if tech, ok := args["technology"].(string); ok && tech != "" {
		component.Technology = tech
	}
	if v, ok := args["tags"].([]any); ok {
		component.Tags = convertInterfaceSlice(v)
	}
}
