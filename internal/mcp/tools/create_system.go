package tools

import (
	"context"
	"fmt"

	d2gen "github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateSystemTool creates a new system in the project.
type CreateSystemTool struct {
	repo usecases.ProjectRepository
}

// NewCreateSystemTool creates a new create_system tool.
func NewCreateSystemTool(repo usecases.ProjectRepository) *CreateSystemTool {
	return &CreateSystemTool{repo: repo}
}

func (t *CreateSystemTool) Name() string {
	return "create_system"
}

func (t *CreateSystemTool) Description() string {
	return "Create a new system in the project with name, description, and optional tags"
}

func (t *CreateSystemTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "System name (e.g., 'Payment Service')",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this system do?",
			},
			"responsibilities": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Key responsibilities (e.g., 'Process payments', 'Store user data')",
			},
			"key_users": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Primary users/actors (e.g., 'User', 'Admin', 'Payment Gateway')",
			},
			"dependencies": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "External dependencies (e.g., 'Database', 'Cache', 'Message Queue')",
			},
			"external_systems": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "External system integrations (e.g., 'Payment API', 'Email Service')",
			},
			"primary_language": map[string]any{
				"type":        "string",
				"description": "Primary programming language (e.g., 'Go', 'Python', 'JavaScript')",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Framework/library (e.g., 'Fiber', 'Django', 'React')",
			},
			"database": map[string]any{
				"type":        "string",
				"description": "Database technology (e.g., 'PostgreSQL', 'MongoDB', 'Redis')",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional tags for categorization",
			},
		},
		"required": []string{"project_root", "name"},
	}
}

// Call executes the create system tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateSystemTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	name, _ := args["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	description, _ := args["description"].(string)
	primaryLanguage, _ := args["primary_language"].(string)
	framework, _ := args["framework"].(string)
	database, _ := args["database"].(string)

	// Convert array interfaces to string slices
	responsibilitiesIface, _ := args["responsibilities"].([]any)
	responsibilities := convertInterfaceSlice(responsibilitiesIface)

	keyUsersIface, _ := args["key_users"].([]any)
	keyUsers := convertInterfaceSlice(keyUsersIface)

	dependenciesIface, _ := args["dependencies"].([]any)
	dependencies := convertInterfaceSlice(dependenciesIface)

	externalSystemsIface, _ := args["external_systems"].([]any)
	externalSystems := convertInterfaceSlice(externalSystemsIface)

	tagsIface, _ := args["tags"].([]any)
	tags := convertInterfaceSlice(tagsIface)

	// 2. Call ScaffoldEntityUseCase
	scaffoldReq := &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot,
		EntityType:  "system",
		Name:        name,
		Description: description,
		Tags:        tags,
		Template:    "", // No template for now
	}

	// Create the use case with diagram generator
	scaffoldUC := usecases.NewScaffoldEntity(t.repo,
		usecases.WithDiagramGenerator(d2gen.NewGenerator()))

	result, err := scaffoldUC.Execute(ctx, scaffoldReq)
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold system: %w", err)
	}

	// Load the created system to get full details for response
	system, err := t.repo.LoadSystem(ctx, projectRoot, result.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created system: %w", err)
	}

	// Set additional system properties that weren't handled by ScaffoldEntity
	system.Responsibilities = responsibilities
	system.KeyUsers = keyUsers
	system.Dependencies = dependencies
	system.ExternalSystems = externalSystems
	system.PrimaryLanguage = primaryLanguage
	system.Framework = framework
	system.Database = database

	// Save the updated system with additional properties
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system with additional properties: %w", err)
	}

	// 3. Format response
	diagramMsg := "Use 'update_diagram' tool to add D2 diagram"
	if result.DiagramPath != "" {
		diagramMsg = "D2 template created at " + result.DiagramPath
	}

	return map[string]any{
		"system": map[string]any{
			"id":               system.ID,
			"name":             system.Name,
			"description":      system.Description,
			"responsibilities": system.Responsibilities,
			"key_users":        system.KeyUsers,
			"dependencies":     system.Dependencies,
			"external_systems": system.ExternalSystems,
			"primary_language": system.PrimaryLanguage,
			"framework":        system.Framework,
			"database":         system.Database,
			"tags":             system.Tags,
			"path":             system.Path,
			"diagram":          diagramMsg,
		},
	}, nil
}

