package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// UpdateDiagramTool updates a diagram source.
type UpdateDiagramTool struct {
	repo usecases.ProjectRepository
}

// NewUpdateDiagramTool creates a new update_diagram tool.
func NewUpdateDiagramTool(repo usecases.ProjectRepository) *UpdateDiagramTool {
	return &UpdateDiagramTool{repo: repo}
}

// Name returns the tool name.
func (t *UpdateDiagramTool) Name() string { return "update_diagram" }

// Description returns the tool description.
func (t *UpdateDiagramTool) Description() string {
	return "Update a system or container D2 diagram source code"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *UpdateDiagramTool) InputSchema() map[string]any {
	return Schemas["update_diagram"].(map[string]any)
}

// Call executes the update diagram tool by delegating to the UpdateDiagramUseCase.
func (t *UpdateDiagramTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	systemName, _ := args["system_name"].(string)
	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}
	d2Source, _ := args["d2_source"].(string)
	if d2Source == "" {
		return nil, fmt.Errorf("d2_source is required")
	}
	containerName, _ := args["container_name"].(string)
	diagramPath, err := t.resolveDiagramPath(ctx, projectRoot, systemName, containerName)
	if err != nil {
		return nil, err
	}
	updateUC := usecases.NewUpdateDiagram()
	if _, err := updateUC.Execute(ctx, &usecases.UpdateDiagramRequest{
		ProjectRoot: projectRoot, DiagramPath: diagramPath, D2Source: d2Source,
	}); err != nil {
		return nil, fmt.Errorf("failed to update diagram: %w", err)
	}
	return diagramUpdateResponse(systemName, containerName), nil
}

// diagramUpdateResponse builds the success response map for an update_diagram call.
func diagramUpdateResponse(systemName, containerName string) map[string]any {
	level, targetName := "system", systemName
	if containerName != "" {
		level, targetName = "container", containerName
	}
	return map[string]any{
		"success": true, "type": level,
		"message": fmt.Sprintf("Diagram updated for %s %q", level, targetName),
	}
}

func (t *UpdateDiagramTool) resolveDiagramPath(ctx context.Context, projectRoot, systemName, containerName string) (string, error) {
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load project: %w", err)
	}
	sourceDir := project.Config.SourceDir
	if sourceDir == "" {
		sourceDir = "./src"
	}
	systemID := entities.NormalizeName(systemName)
	if containerName != "" {
		containerID := entities.NormalizeName(containerName)
		return filepath.Join(sourceDir, systemID, containerID, containerID+".d2"), nil
	}
	return filepath.Join(sourceDir, systemID, "system.d2"), nil
}
