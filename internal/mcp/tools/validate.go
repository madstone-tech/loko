package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
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

// Name returns the tool name.
func (t *ValidateTool) Name() string { return "validate" }

// Description returns the tool description.
func (t *ValidateTool) Description() string {
	return "Validate the project architecture for errors and warnings"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *ValidateTool) InputSchema() map[string]any { return Schemas["validate"].(map[string]any) }

// Call executes the validate tool by delegating to the ValidateArchitectureUseCase.
func (t *ValidateTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	graph, systems, err := t.buildGraph(ctx, projectRoot)
	if err != nil {
		return nil, err
	}
	report := usecases.NewValidateArchitecture().Execute(graph, systems)
	var warnings []string
	for _, sys := range systems {
		if sys.ContainerCount() == 0 {
			warnings = append(warnings, fmt.Sprintf("System %q has no containers", sys.Name))
		}
	}
	return map[string]any{
		"valid": len(warnings) == 0 && report.IsValid, "warnings": warnings, "report": report,
	}, nil
}

func (t *ValidateTool) buildGraph(ctx context.Context, projectRoot string) (*entities.ArchitectureGraph, []*entities.System, error) {
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project: %w", err)
	}
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load systems: %w", err)
	}
	graphUC := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := graphUC.Execute(ctx, project, systems)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}
	return graph, systems, nil
}
