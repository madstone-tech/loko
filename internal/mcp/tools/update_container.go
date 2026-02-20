package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// UpdateContainerTool updates an existing container's metadata.
type UpdateContainerTool struct {
	repo usecases.ProjectRepository
}

// NewUpdateContainerTool creates a new update_container tool.
func NewUpdateContainerTool(repo usecases.ProjectRepository) *UpdateContainerTool {
	return &UpdateContainerTool{repo: repo}
}

func (t *UpdateContainerTool) Name() string {
	return "update_container"
}

func (t *UpdateContainerTool) Description() string {
	return "Update an existing container's metadata (description, technology, tags)"
}

func (t *UpdateContainerTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "Parent system name",
			},
			"container_name": map[string]any{
				"type":        "string",
				"description": "Container name or ID to update",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New description (leave empty to keep current)",
			},
			"technology": map[string]any{
				"type":        "string",
				"description": "New technology stack (leave empty to keep current)",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace tags list",
			},
		},
		"required": []string{"project_root", "system_name", "container_name"},
	}
}

// Call executes the update container tool.
func (t *UpdateContainerTool) Call(ctx context.Context, args map[string]any) (any, error) {
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

	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)

	// First try to load the container with the provided IDs
	container, err := t.repo.LoadContainer(ctx, projectRoot, systemID, containerID)
	if err != nil {
		// If that fails, try to get a suggestion for the error message
		graph, graphErr := getGraphFromProject(ctx, t.repo, projectRoot)
		if graphErr != nil {
			// If we can't build a graph, return the original error
			return nil, fmt.Errorf("failed to load container %q in system %q: %w", containerID, systemID, err)
		}

		// Try to find a suggestion using the graph
		suggestion := suggestSlugID(containerName, graph)
		return nil, notFoundError("container", containerName, suggestion)
	}

	// Update only non-empty fields
	if desc, ok := args["description"].(string); ok && desc != "" {
		container.Description = desc
	}
	if tech, ok := args["technology"].(string); ok && tech != "" {
		container.Technology = tech
	}
	if v, ok := args["tags"].([]any); ok {
		container.Tags = convertInterfaceSlice(v)
	}

	// Save
	if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return map[string]any{
		"container": map[string]any{
			"id":          container.ID,
			"name":        container.Name,
			"description": container.Description,
			"technology":  container.Technology,
			"tags":        container.Tags,
		},
		"message": fmt.Sprintf("Container %q updated", container.Name),
	}, nil
}
