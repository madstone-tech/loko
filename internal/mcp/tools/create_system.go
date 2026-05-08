package tools

import (
	"context"
	"fmt"

	d2gen "github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/entities"
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

// Name returns the tool name.
func (t *CreateSystemTool) Name() string { return "create_system" }

// Description returns the tool description.
func (t *CreateSystemTool) Description() string {
	return "Create a new system in the project with name, description, and optional tags"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *CreateSystemTool) InputSchema() map[string]any { return createSystemSchema }

// Call executes the create system tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateSystemTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	name, _ := args["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	sp := parseSystemProps(args)
	scaffoldUC := usecases.NewScaffoldEntity(t.repo, usecases.WithDiagramGenerator(d2gen.NewGenerator()))
	result, err := scaffoldUC.Execute(ctx, &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot, EntityType: "system",
		Name: name, Description: sp.description, Tags: sp.tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scaffold system: %w", err)
	}
	system, err := t.repo.LoadSystem(ctx, projectRoot, result.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created system: %w", err)
	}
	sp.applyTo(system)
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system with additional properties: %w", err)
	}
	return map[string]any{"system": systemToMap(system, diagramMessageFor(result.DiagramPath))}, nil
}

// systemProps holds parsed system metadata fields from tool args.
type systemProps struct {
	description, primaryLanguage, framework, database               string
	responsibilities, keyUsers, dependencies, externalSystems, tags []string
}

// parseSystemProps extracts system metadata from raw tool args.
func parseSystemProps(args map[string]any) systemProps {
	str := func(k string) string { s, _ := args[k].(string); return s }
	sl := func(k string) []string { v, _ := args[k].([]any); return convertInterfaceSlice(v) }
	return systemProps{
		description: str("description"), primaryLanguage: str("primary_language"),
		framework: str("framework"), database: str("database"),
		responsibilities: sl("responsibilities"), keyUsers: sl("key_users"),
		dependencies: sl("dependencies"), externalSystems: sl("external_systems"),
		tags: sl("tags"),
	}
}

// applyTo sets all additional system properties on the given system entity.
func (sp systemProps) applyTo(system *entities.System) {
	system.Responsibilities = sp.responsibilities
	system.KeyUsers = sp.keyUsers
	system.Dependencies = sp.dependencies
	system.ExternalSystems = sp.externalSystems
	system.PrimaryLanguage = sp.primaryLanguage
	system.Framework = sp.framework
	system.Database = sp.database
}

// systemToMap converts a System entity to a JSON-friendly map for MCP responses.
func systemToMap(system *entities.System, diagramMsg string) map[string]any {
	return map[string]any{
		"id": system.ID, "name": system.Name, "description": system.Description,
		"responsibilities": system.Responsibilities, "key_users": system.KeyUsers,
		"dependencies": system.Dependencies, "external_systems": system.ExternalSystems,
		"primary_language": system.PrimaryLanguage, "framework": system.Framework,
		"database": system.Database, "tags": system.Tags,
		"path": system.Path, "diagram": diagramMsg,
	}
}
