package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ValidateTool validates the architecture.
type ValidateTool struct {
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository // Optional: loads relationships.toml into graph
}

// NewValidateTool creates a new validate tool.
func NewValidateTool(repo usecases.ProjectRepository) *ValidateTool {
	return &ValidateTool{repo: repo}
}

// NewValidateToolFull creates a new validate tool with relationship repo.
func NewValidateToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository) *ValidateTool {
	return &ValidateTool{repo: repo, relRepo: relRepo}
}

func (t *ValidateTool) Name() string {
	return "validate"
}

func (t *ValidateTool) Description() string {
	return "Validate the project architecture for errors and warnings"
}

func (t *ValidateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
		},
		"required": []string{"project_root"},
	}
}

// Call executes the validate tool by delegating to the ValidateArchitectureUseCase.
func (t *ValidateTool) Call(ctx context.Context, args map[string]any) (any, error) {
	// 1. Parse and validate inputs
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}

	// 2. Load project and systems
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	// 3. Call ValidateArchitectureUseCase
	validateUC := usecases.NewValidateArchitecture()

	// Build architecture graph (includes relationships.toml when relRepo is wired).
	graphUC := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphUC.Execute(ctx, project, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}

	// Validate architecture
	report := validateUC.Execute(graph, systems)

	// 4. Format response
	var warnings []string
	for _, sys := range systems {
		if sys.ContainerCount() == 0 {
			warnings = append(warnings, fmt.Sprintf("System %q has no containers", sys.Name))
		}
	}

	return map[string]any{
		"valid":    len(warnings) == 0 && report.IsValid,
		"warnings": warnings,
		"report":   report,
	}, nil
}
