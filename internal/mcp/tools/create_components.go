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

func (t *CreateComponentsTool) Name() string {
	return "create_components"
}

func (t *CreateComponentsTool) Description() string {
	return "Create multiple components in a container in a single operation"
}

func (t *CreateComponentsTool) InputSchema() map[string]any {
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
			"components": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Component name (e.g., 'Auth Handler', 'Request Validator')",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "What does this component do?",
						},
						"technology": map[string]any{
							"type":        "string",
							"description": "Technology/implementation details (e.g., 'Go', 'Python module')",
						},
						"tags": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Tags for categorization",
						},
					},
					"required": []string{"name"},
				},
				"description": "Array of component definitions to create",
			},
		},
		"required": []string{"project_root", "system_name", "container_name", "components"},
	}
}

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

	results := make([]map[string]any, 0, len(componentsIface))
	created, failed := 0, 0

	for i, compIface := range componentsIface {
		compMap, ok := compIface.(map[string]any)
		if !ok {
			results = append(results, map[string]any{
				"status": "error",
				"error":  fmt.Sprintf("component %d is not a valid object", i),
			})
			failed++
			continue
		}

		item, entityID := t.scaffoldOne(ctx, projectRoot, systemName, containerName, compMap)
		if entityID != "" {
			item["id"] = entityID
			created++
		} else {
			failed++
		}
		results = append(results, item)
	}

	return map[string]any{
		"created": created,
		"failed":  failed,
		"results": results,
	}, nil
}

// scaffoldOne creates a single component. Returns the result map and entity ID (empty on error).
func (t *CreateComponentsTool) scaffoldOne(
	ctx context.Context,
	projectRoot, systemName, containerName string,
	compMap map[string]any,
) (map[string]any, string) {
	name := getComponentString(compMap, "name")
	if name == "" {
		return map[string]any{
			"status": "error",
			"error":  "name is required",
		}, ""
	}

	var tags []string
	if tagsIface, ok := compMap["tags"].([]any); ok {
		tags = convertInterfaceSlice(tagsIface)
	}

	scaffoldUC := usecases.NewScaffoldEntity(t.repo)
	scaffoldResult, err := scaffoldUC.Execute(ctx, &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot,
		EntityType:  "component",
		ParentPath:  []string{systemName, containerName},
		Name:        name,
		Description: getComponentString(compMap, "description"),
		Technology:  getComponentString(compMap, "technology"),
		Tags:        tags,
	})
	if err != nil {
		return map[string]any{
			"name":   name,
			"status": "error",
			"error":  fmt.Sprintf("failed to scaffold component: %v", err),
		}, ""
	}

	return map[string]any{
		"name":   name,
		"status": "created",
	}, scaffoldResult.EntityID
}

// getComponentString safely extracts a string from a map.
func getComponentString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
