package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateContainerTool creates a new container in a system.
type CreateContainerTool struct {
	repo usecases.ProjectRepository
}

// NewCreateContainerTool creates a new create_container tool.
func NewCreateContainerTool(repo usecases.ProjectRepository) *CreateContainerTool {
	return &CreateContainerTool{repo: repo}
}

func (t *CreateContainerTool) Name() string {
	return "create_container"
}

func (t *CreateContainerTool) Description() string {
	return "Create a new container in a system"
}

func (t *CreateContainerTool) InputSchema() map[string]any {
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
			"name": map[string]any{
				"type":        "string",
				"description": "Container name (e.g., 'API Server', 'Web Frontend', 'Database')",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this container do? (e.g., 'Handles all REST API requests')",
			},
			"technology": map[string]any{
				"type":        "string",
				"description": "Technology stack (e.g., 'Go + Fiber', 'Node.js + Express', 'PostgreSQL 15')",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Tags for categorization (e.g., 'backend', 'database', 'frontend')",
			},
		},
		"required": []string{"project_root", "system_name", "name"},
	}
}

// Call executes the create container tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateContainerTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
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

	// Convert array interfaces to string slices
	tagsIface, _ := args["tags"].([]any)
	tags := convertInterfaceSlice(tagsIface)

	// 2. Call ScaffoldEntityUseCase
	scaffoldReq := &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot,
		EntityType:  "container",
		ParentPath:  []string{systemName},
		Name:        name,
		Description: description,
		Technology:  technology,
		Tags:        tags,
	}

	scaffoldUC := usecases.NewScaffoldEntity(t.repo)
	result, err := scaffoldUC.Execute(ctx, scaffoldReq)
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold container: %w", err)
	}

	// 3. Format response
	diagramMsg := "Use 'update_diagram' tool to add D2 diagram"
	if result.DiagramPath != "" {
		diagramMsg = "D2 template created at " + result.DiagramPath
	}

	return map[string]any{
		"container": map[string]any{
			"id":          result.EntityID,
			"name":        name,
			"description": description,
			"technology":  technology,
			"tags":        tags,
			"diagram":     diagramMsg,
		},
	}, nil
}
