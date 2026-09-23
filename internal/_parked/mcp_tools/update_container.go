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

// Name returns the tool name.
func (t *UpdateContainerTool) Name() string { return "update_container" }

// Description returns the tool description.
func (t *UpdateContainerTool) Description() string {
	return "Update an existing container's metadata (description, technology, tags)"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *UpdateContainerTool) InputSchema() map[string]any { return updateContainerSchema }

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
	container, err := t.loadContainer(ctx, projectRoot, systemName, containerName)
	if err != nil {
		return nil, err
	}
	applyContainerUpdates(container, args)
	systemID := entities.NormalizeName(systemName)
	if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}
	return map[string]any{
		"container": map[string]any{
			"id": container.ID, "name": container.Name,
			"description": container.Description, "technology": container.Technology, "tags": container.Tags,
		},
		"message": fmt.Sprintf("Container %q updated", container.Name),
	}, nil
}

func (t *UpdateContainerTool) loadContainer(ctx context.Context, projectRoot, systemName, containerName string) (*entities.Container, error) {
	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)
	container, err := t.repo.LoadContainer(ctx, projectRoot, systemID, containerID)
	if err != nil {
		graph, graphErr := getGraphFromProject(ctx, t.repo, projectRoot)
		if graphErr != nil {
			return nil, fmt.Errorf("failed to load container %q in system %q: %w", containerID, systemID, err)
		}
		return nil, notFoundError("container", containerName, suggestSlugID(containerName, graph))
	}
	return container, nil
}

// applyContainerUpdates patches non-empty fields on a Container entity from raw args.
func applyContainerUpdates(container *entities.Container, args map[string]any) {
	if desc, ok := args["description"].(string); ok && desc != "" {
		container.Description = desc
	}
	if tech, ok := args["technology"].(string); ok && tech != "" {
		container.Technology = tech
	}
	if v, ok := args["tags"].([]any); ok {
		container.Tags = convertInterfaceSlice(v)
	}
}
