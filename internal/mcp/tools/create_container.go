package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateContainerTool creates a new container in a system.
type CreateContainerTool struct {
	repo             usecases.ProjectRepository
	diagramGenerator usecases.DiagramGenerator
}

// NewCreateContainerTool creates a new create_container tool.
func NewCreateContainerTool(repo usecases.ProjectRepository, diagramGenerator usecases.DiagramGenerator) *CreateContainerTool {
	return &CreateContainerTool{repo: repo, diagramGenerator: diagramGenerator}
}

// Name returns the tool name.
func (t *CreateContainerTool) Name() string { return "create_container" }

// Description returns the tool description.
func (t *CreateContainerTool) Description() string {
	return "Create a new container in a system"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *CreateContainerTool) InputSchema() map[string]any { return createContainerSchema }

// Call executes the create container tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateContainerTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	systemName, _ := args["system_name"].(string)
	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}
	name, _ := args["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	description, _ := args["description"].(string)
	technology, _ := args["technology"].(string)
	tagsIface, _ := args["tags"].([]any)
	tags := convertInterfaceSlice(tagsIface)

	result, err := t.scaffold(ctx, projectRoot, systemName, name, description, technology, tags)
	if err != nil {
		return nil, err
	}
	return map[string]any{"container": map[string]any{
		"id": result.EntityID, "name": name, "description": description,
		"technology": technology, "tags": tags, "diagram": diagramMessageFor(result.DiagramPath),
	}}, nil
}

func (t *CreateContainerTool) scaffold(ctx context.Context, projectRoot, systemName, name, description, technology string, tags []string) (*usecases.ScaffoldEntityResult, error) {
	scaffoldUC := usecases.NewScaffoldEntity(t.repo, usecases.WithDiagramGenerator(t.diagramGenerator))
	result, err := scaffoldUC.Execute(ctx, &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot, EntityType: "container",
		ParentPath: []string{systemName}, Name: name,
		Description: description, Technology: technology, Tags: tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold container: %w", err)
	}
	return result, nil
}
