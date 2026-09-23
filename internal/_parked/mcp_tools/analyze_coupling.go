package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// AnalyzeCouplingTool analyzes coupling metrics in the architecture.
type AnalyzeCouplingTool struct {
	repo    usecases.ProjectRepository
	relRepo usecases.RelationshipRepository // Optional: loads relationships.toml into graph
	encoder usecases.OutputEncoder
}

// NewAnalyzeCouplingTool creates a new analyze_coupling tool.
func NewAnalyzeCouplingTool(repo usecases.ProjectRepository, encoder usecases.OutputEncoder) *AnalyzeCouplingTool {
	return &AnalyzeCouplingTool{repo: repo, encoder: encoder}
}

// NewAnalyzeCouplingToolFull creates a new analyze_coupling tool with relationship repo.
func NewAnalyzeCouplingToolFull(repo usecases.ProjectRepository, relRepo usecases.RelationshipRepository, encoder usecases.OutputEncoder) *AnalyzeCouplingTool {
	return &AnalyzeCouplingTool{repo: repo, relRepo: relRepo, encoder: encoder}
}

// Name returns the tool name.
func (t *AnalyzeCouplingTool) Name() string { return "analyze_coupling" }

// Description returns the tool description.
func (t *AnalyzeCouplingTool) Description() string {
	return "Analyze coupling metrics for a system, identifying highly coupled and central components"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *AnalyzeCouplingTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{"type": "string", "description": "Root directory of the project"},
			"system_id":    map[string]any{"type": "string", "description": "Optional: ID of the system to analyze (if empty, analyzes all systems)"},
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"toon", "json"},
				"default":     "toon",
				"description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging",
			},
		},
	}
}

// Call executes the analyze_coupling tool.
func (t *AnalyzeCouplingTool) Call(ctx context.Context, args map[string]any) (any, error) {
	format, err := getFormat(args)
	if err != nil {
		return nil, err
	}
	result, err := t.analyze(ctx, args)
	if err != nil {
		return nil, err
	}
	return formatResponse(result, format, t.encoder)
}

func (t *AnalyzeCouplingTool) analyze(ctx context.Context, args map[string]any) (map[string]any, error) {
	var typedArgs AnalyzeCouplingArgs
	if err := mapToStruct(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if typedArgs.ProjectRoot == "" {
		typedArgs.ProjectRoot = "."
	}
	graphBuilder, graph, err := t.loadGraph(ctx, typedArgs.ProjectRoot)
	if err != nil {
		return nil, err
	}
	targetGraph, err := t.selectTargetGraph(graphBuilder, graph, typedArgs.SystemID)
	if err != nil {
		return nil, err
	}
	report := graphBuilder.AnalyzeDependencies(targetGraph)
	return map[string]any{
		"systems_count":             report.SystemsCount,
		"containers_count":          report.ContainersCount,
		"components_count":          report.ComponentsCount,
		"total_nodes":               report.TotalNodes,
		"total_edges":               report.TotalEdges,
		"isolated_components":       report.IsolatedComponents,
		"highly_coupled_components": report.HighlyCoupledComponents,
		"central_components":        report.CentralComponents,
		"note":                      "Isolated components have no relationships; Central components have high in-degree (many dependents)",
	}, nil
}
func (t *AnalyzeCouplingTool) loadGraph(ctx context.Context, projectRoot string) (*usecases.BuildArchitectureGraph, *entities.ArchitectureGraph, error) {
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project: %w", err)
	}
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load systems: %w", err)
	}
	builder := usecases.NewBuildArchitectureGraphWithRelRepo(t.relRepo)
	graph, err := builder.Execute(ctx, project, systems)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build architecture graph: %w", err)
	}
	return builder, graph, nil
}

func (t *AnalyzeCouplingTool) selectTargetGraph(builder *usecases.BuildArchitectureGraph, graph *entities.ArchitectureGraph, systemID string) (*entities.ArchitectureGraph, error) {
	if systemID == "" {
		return graph, nil
	}
	subgraph, err := builder.GetSystemGraph(graph, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get system graph: %w", err)
	}
	return subgraph, nil
}
