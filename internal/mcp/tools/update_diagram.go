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

func (t *UpdateDiagramTool) Name() string {
	return "update_diagram"
}

func (t *UpdateDiagramTool) Description() string {
	return "Update a system or container D2 diagram source code"
}

func (t *UpdateDiagramTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "System name",
			},
			"container_name": map[string]any{
				"type":        "string",
				"description": "Container name (optional, for container diagrams)",
			},
			"d2_source": map[string]any{
				"type":        "string",
				"description": "New D2 diagram source code",
			},
		},
		"required": []string{"project_root", "system_name", "d2_source"},
	}
}

// Call executes the update diagram tool by delegating to the UpdateDiagramUseCase.
func (t *UpdateDiagramTool) Call(ctx context.Context, args map[string]any) (any, error) {
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
	d2Source, _ := args["d2_source"].(string)
	if d2Source == "" {
		return nil, fmt.Errorf("d2_source is required")
	}

	// 2. Load project to get source directory
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	sourceDir := project.Config.SourceDir
	if sourceDir == "" {
		sourceDir = "./src"
	}

	// Build diagram path using project source directory
	var diagramPath string
	systemID := entities.NormalizeName(systemName)
	if containerName != "" {
		// Update container diagram: src/{systemID}/{containerID}/{containerID}.d2
		containerID := entities.NormalizeName(containerName)
		diagramPath = filepath.Join(sourceDir, systemID, containerID, containerID+".d2")
	} else {
		// Update system diagram: src/{systemID}/system.d2
		diagramPath = filepath.Join(sourceDir, systemID, "system.d2")
	}

	updateReq := &usecases.UpdateDiagramRequest{
		ProjectRoot: projectRoot,
		DiagramPath: diagramPath,
		D2Source:    d2Source,
	}

	updateUC := usecases.NewUpdateDiagram()
	_, err = updateUC.Execute(ctx, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update diagram: %w", err)
	}

	// 3. Format response
	message := fmt.Sprintf("Diagram updated for %s %q",
		map[bool]string{true: "container", false: "system"}[containerName != ""],
		map[bool]string{true: containerName, false: systemName}[containerName != ""])

	return map[string]any{
		"success": true,
		"message": message,
		"type":    map[bool]string{true: "container", false: "system"}[containerName != ""],
	}, nil
}
