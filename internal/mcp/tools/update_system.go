package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// UpdateSystemTool updates an existing system's metadata.
type UpdateSystemTool struct {
	repo usecases.ProjectRepository
}

// NewUpdateSystemTool creates a new update_system tool.
func NewUpdateSystemTool(repo usecases.ProjectRepository) *UpdateSystemTool {
	return &UpdateSystemTool{repo: repo}
}

func (t *UpdateSystemTool) Name() string {
	return "update_system"
}

func (t *UpdateSystemTool) Description() string {
	return "Update an existing system's metadata (description, tags, responsibilities, etc.)"
}

func (t *UpdateSystemTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "System name or ID to update",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New description (leave empty to keep current)",
			},
			"responsibilities": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace responsibilities list",
			},
			"key_users": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace key users list",
			},
			"dependencies": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace dependencies list",
			},
			"external_systems": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace external systems list",
			},
			"primary_language": map[string]any{
				"type":        "string",
				"description": "Primary programming language",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Framework/library",
			},
			"database": map[string]any{
				"type":        "string",
				"description": "Database technology",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Replace tags list",
			},
		},
		"required": []string{"project_root", "system_name"},
	}
}

// Call executes the update system tool.
func (t *UpdateSystemTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	systemName, _ := args["system_name"].(string)
	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}

	systemID := entities.NormalizeName(systemName)

	// Load existing system
	system, err := t.repo.LoadSystem(ctx, projectRoot, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to load system %q: %w", systemID, err)
	}

	// Update only non-empty fields
	if desc, ok := args["description"].(string); ok && desc != "" {
		system.Description = desc
	}
	if lang, ok := args["primary_language"].(string); ok && lang != "" {
		system.PrimaryLanguage = lang
	}
	if fw, ok := args["framework"].(string); ok && fw != "" {
		system.Framework = fw
	}
	if db, ok := args["database"].(string); ok && db != "" {
		system.Database = db
	}

	// Update array fields if provided
	if v, ok := args["responsibilities"].([]any); ok {
		system.Responsibilities = convertInterfaceSlice(v)
	}
	if v, ok := args["key_users"].([]any); ok {
		system.KeyUsers = convertInterfaceSlice(v)
	}
	if v, ok := args["dependencies"].([]any); ok {
		system.Dependencies = convertInterfaceSlice(v)
	}
	if v, ok := args["external_systems"].([]any); ok {
		system.ExternalSystems = convertInterfaceSlice(v)
	}
	if v, ok := args["tags"].([]any); ok {
		system.Tags = convertInterfaceSlice(v)
	}

	// Save
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system: %w", err)
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
		},
		"message": fmt.Sprintf("System %q updated", system.Name),
	}, nil
}
