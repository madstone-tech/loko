package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateComponentTool creates a new component in a container.
type CreateComponentTool struct {
	repo usecases.ProjectRepository
}

// NewCreateComponentTool creates a new create_component tool.
func NewCreateComponentTool(repo usecases.ProjectRepository) *CreateComponentTool {
	return &CreateComponentTool{repo: repo}
}

func (t *CreateComponentTool) Name() string {
	return "create_component"
}

func (t *CreateComponentTool) Description() string {
	return "Create a new component in a container"
}

func (t *CreateComponentTool) InputSchema() map[string]any {
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
				"description": "Parent container name",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Component name (e.g., 'Auth Handler', 'Product Service', 'Cache Manager')",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this component do? (e.g., 'Handles JWT authentication')",
			},
			"technology": map[string]any{
				"type":        "string",
				"description": "Technology/implementation details (e.g., 'Go', 'React Component', 'Python module')",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Tags for categorization (e.g., 'auth', 'handler', 'service')",
			},
		},
		"required": []string{"project_root", "system_name", "container_name", "name"},
	}
}

// Call executes the create component tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateComponentTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
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
		EntityType:  "component",
		ParentPath:  []string{systemName, containerName},
		Name:        name,
		Description: description,
		Technology:  technology,
		Tags:        tags,
	}

	scaffoldUC := usecases.NewScaffoldEntity(t.repo)
	result, err := scaffoldUC.Execute(ctx, scaffoldReq)
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold component: %w", err)
	}

	// 3. Format response
	return map[string]any{
		"component": map[string]any{
			"id":          result.EntityID,
			"name":        name,
			"description": description,
			"technology":  technology,
			"tags":        tags,
		},
	}, nil
}
