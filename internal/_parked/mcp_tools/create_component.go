package tools

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// CreateComponentTool creates a new component in a container.
type CreateComponentTool struct {
	repo usecases.ProjectRepository
}

// NewCreateComponentTool creates a new create_component tool.
func NewCreateComponentTool(repo usecases.ProjectRepository) *CreateComponentTool {
	return &CreateComponentTool{repo: repo}
}

// Name returns the tool name.
func (t *CreateComponentTool) Name() string { return "create_component" }

// Description returns the tool description.
func (t *CreateComponentTool) Description() string {
	return "Create a new component in a container"
}

// InputSchema returns the JSON schema for this tool's inputs.
func (t *CreateComponentTool) InputSchema() map[string]any { return createComponentSchema }

// Call executes the create component tool by delegating to the ScaffoldEntityUseCase.
func (t *CreateComponentTool) Call(ctx context.Context, args map[string]any) (any, error) {
	projectRoot, _ := args["project_root"].(string)
	if projectRoot == "" {
		projectRoot = "."
	}
	cp, err := parseComponentArgs(args)
	if err != nil {
		return nil, err
	}
	entityID, err := t.scaffoldComponent(ctx, projectRoot, cp)
	if err != nil {
		return nil, err
	}
	return t.buildResponse(ctx, entityID, cp), nil
}

func (t *CreateComponentTool) scaffoldComponent(ctx context.Context, projectRoot string, cp componentArgs) (string, error) {
	scaffoldUC := usecases.NewScaffoldEntity(t.repo)
	result, err := scaffoldUC.Execute(ctx, &usecases.ScaffoldEntityRequest{
		ProjectRoot: projectRoot, EntityType: "component",
		ParentPath: []string{cp.systemName, cp.containerName},
		Name:       cp.name, Description: cp.description, Technology: cp.technology, Tags: cp.tags,
	})
	if err != nil {
		return "", fmt.Errorf("failed to scaffold component: %w", err)
	}
	return result.EntityID, nil
}

func (t *CreateComponentTool) buildResponse(ctx context.Context, entityID string, cp componentArgs) map[string]any {
	response := map[string]any{
		"component": map[string]any{
			"id": entityID, "name": cp.name,
			"description": cp.description, "technology": cp.technology, "tags": cp.tags,
		},
	}
	if cp.preview {
		if svgContent, pErr := t.generatePreview(ctx, cp.name, cp.technology, cp.containerName); pErr != nil {
			response["preview_error"] = pErr.Error()
		} else {
			response["diagram_preview"] = svgContent
		}
	}
	return response
}

// componentArgs holds parsed and validated arguments for create_component.
type componentArgs struct {
	systemName, containerName, name, description, technology string
	tags                                                     []string
	preview                                                  bool
}

// parseComponentArgs extracts and validates required fields from raw args.
func parseComponentArgs(args map[string]any) (componentArgs, error) {
	var cp componentArgs
	cp.systemName, _ = args["system_name"].(string)
	if cp.systemName == "" {
		return cp, fmt.Errorf("system_name is required")
	}
	cp.containerName, _ = args["container_name"].(string)
	if cp.containerName == "" {
		return cp, fmt.Errorf("container_name is required")
	}
	cp.name, _ = args["name"].(string)
	if cp.name == "" {
		return cp, fmt.Errorf("name is required")
	}
	cp.description, _ = args["description"].(string)
	cp.technology, _ = args["technology"].(string)
	cp.preview, _ = args["preview"].(bool)
	tagsIface, _ := args["tags"].([]any)
	cp.tags = convertInterfaceSlice(tagsIface)
	return cp, nil
}

// generatePreview creates a diagram preview for a component.
func (t *CreateComponentTool) generatePreview(ctx context.Context, componentName, technology, containerName string) (string, error) {
	renderer := d2.NewRenderer()
	if !renderer.IsAvailable() {
		return "", fmt.Errorf("d2 binary not found in PATH - install from https://d2lang.com/")
	}
	previewRenderer := d2.NewPreviewRenderer(renderer)
	svgContent, err := previewRenderer.RenderComponentPreview(ctx, componentName, technology, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to render preview: %w", err)
	}
	return svgContent, nil
}
