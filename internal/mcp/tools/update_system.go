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

// Name returns the tool name.
func (t *UpdateSystemTool) Name() string { return "update_system" }

// Description returns the tool description.
func (t *UpdateSystemTool) Description() string {
	return "Update an existing system's metadata (description, tags, responsibilities, etc.)"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *UpdateSystemTool) InputSchema() map[string]any { return updateSystemSchema }

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
	system, err := t.loadSystem(ctx, projectRoot, systemName)
	if err != nil {
		return nil, err
	}
	applySystemUpdates(system, args)
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system: %w", err)
	}
	return map[string]any{
		"system":  systemToMap(system, ""),
		"message": fmt.Sprintf("System %q updated", system.Name),
	}, nil
}

func (t *UpdateSystemTool) loadSystem(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
	systemID := entities.NormalizeName(systemName)
	system, err := t.repo.LoadSystem(ctx, projectRoot, systemID)
	if err != nil {
		graph, graphErr := getGraphFromProject(ctx, t.repo, projectRoot)
		if graphErr != nil {
			return nil, fmt.Errorf("failed to load system %q: %w", systemID, err)
		}
		return nil, notFoundError("system", systemName, suggestSlugID(systemName, graph))
	}
	return system, nil
}

// applySystemUpdates patches non-empty fields on a System entity from raw args.
func applySystemUpdates(system *entities.System, args map[string]any) {
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
}
