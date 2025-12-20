package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/html"
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

func (t *CreateSystemTool) Name() string {
	return "create_system"
}

func (t *CreateSystemTool) Description() string {
	return "Create a new system in the project with name, description, and optional tags"
}

func (t *CreateSystemTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "System name (e.g., 'Payment Service')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "What does this system do?",
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional tags for categorization",
			},
		},
		"required": []string{"project_root", "name"},
	}
}

func (t *CreateSystemTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	tagsIface, _ := args["tags"].([]interface{})

	if projectRoot == "" {
		projectRoot = "."
	}

	// Create system
	uc := usecases.NewCreateSystem(t.repo)
	req := &usecases.CreateSystemRequest{
		Name:        name,
		Description: description,
	}

	// Convert tags
	for _, tag := range tagsIface {
		if tagStr, ok := tag.(string); ok {
			req.Tags = append(req.Tags, tagStr)
		}
	}

	system, err := uc.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}

	// Save the system
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system: %w", err)
	}

	// Attempt to create a basic D2 diagram template (optional)
	diagramMsg := "Use 'update_diagram' tool to add D2 diagram"
	if err := createSystemD2Template(ctx, projectRoot, system); err == nil {
		diagramMsg = "D2 template created at " + system.ID + "/" + system.ID + ".d2"
	}

	return map[string]interface{}{
		"system": map[string]interface{}{
			"id":          system.ID,
			"name":        system.Name,
			"description": system.Description,
			"tags":        system.Tags,
			"path":        system.Path,
			"diagram":     diagramMsg,
		},
	}, nil
}

// CreateContainerTool creates a new container in a system.
type CreateContainerTool struct {
	repo usecases.ProjectRepository
}

// NewCreateContainerTool creates a new create_container tool.
func NewCreateContainerTool(repo usecases.ProjectRepository) *CreateContainerTool {
	return &CreateContainerTool{repo: repo}
}

func (t *CreateContainerTool) Name() string {
	return "create_container"
}

func (t *CreateContainerTool) Description() string {
	return "Create a new container in a system"
}

func (t *CreateContainerTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]interface{}{
				"type":        "string",
				"description": "Parent system name",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Container name",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "What does this container do?",
			},
			"technology": map[string]interface{}{
				"type":        "string",
				"description": "Technology stack (e.g., 'Go + Fiber', 'Node.js + Express')",
			},
		},
		"required": []string{"project_root", "system_name", "name"},
	}
}

func (t *CreateContainerTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	systemName, _ := args["system_name"].(string)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	technology, _ := args["technology"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Normalize system name to ID
	systemID := entities.NormalizeName(systemName)

	// Load system
	system, err := t.repo.LoadSystem(ctx, projectRoot, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to load system: %w", err)
	}

	// Create container
	container, err := entities.NewContainer(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	container.Description = description
	container.Technology = technology

	// Add to system
	if err := system.AddContainer(container); err != nil {
		return nil, fmt.Errorf("failed to add container to system: %w", err)
	}

	// Save container
	if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	// Attempt to create a basic D2 diagram template (optional)
	diagramMsg := "Use 'update_diagram' tool to add D2 diagram"
	if err := createContainerD2Template(ctx, projectRoot, systemID, container); err == nil {
		diagramMsg = "D2 template created at " + systemID + "/" + container.ID + "/" + container.ID + ".d2"
	}

	return map[string]interface{}{
		"container": map[string]interface{}{
			"id":          container.ID,
			"name":        container.Name,
			"description": container.Description,
			"technology":  container.Technology,
			"diagram":     diagramMsg,
		},
	}, nil
}

// CreateComponentTool creates a new component in a container.
type CreateComponentTool struct {
	repo usecases.ProjectRepository
}

// NewCreateComponentTool creates a new create_component tool.
func NewCreateComponentTool(repo usecases.ProjectRepository) *CreateComponentTool {
	return &CreateComponentTool{repo: repo}
}

func (t *CreateComponentTool) Name() string {
	return "create_component"
}

func (t *CreateComponentTool) Description() string {
	return "Create a new component in a container"
}

func (t *CreateComponentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]interface{}{
				"type":        "string",
				"description": "Parent system name",
			},
			"container_name": map[string]interface{}{
				"type":        "string",
				"description": "Parent container name",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Component name",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "What does this component do?",
			},
		},
		"required": []string{"project_root", "system_name", "container_name", "name"},
	}
}

func (t *CreateComponentTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	systemName, _ := args["system_name"].(string)
	containerName, _ := args["container_name"].(string)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Normalize system and container names to IDs
	systemID := entities.NormalizeName(systemName)
	containerID := entities.NormalizeName(containerName)

	// Load container
	container, err := t.repo.LoadContainer(ctx, projectRoot, systemID, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Create component
	component, err := entities.NewComponent(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	component.Description = description

	// Add to container
	if err := container.AddComponent(component); err != nil {
		return nil, fmt.Errorf("failed to add component to container: %w", err)
	}

	// Save container
	if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	return map[string]interface{}{
		"component": map[string]interface{}{
			"id":          component.ID,
			"name":        component.Name,
			"description": component.Description,
		},
	}, nil
}

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

func (t *UpdateDiagramTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]interface{}{
				"type":        "string",
				"description": "System name",
			},
			"container_name": map[string]interface{}{
				"type":        "string",
				"description": "Container name (optional, for container diagrams)",
			},
			"d2_source": map[string]interface{}{
				"type":        "string",
				"description": "New D2 diagram source code",
			},
		},
		"required": []string{"project_root", "system_name", "d2_source"},
	}
}

func (t *UpdateDiagramTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	systemName, _ := args["system_name"].(string)
	containerName, _ := args["container_name"].(string)
	d2Source, _ := args["d2_source"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	if systemName == "" {
		return nil, fmt.Errorf("system_name is required")
	}

	if d2Source == "" {
		return nil, fmt.Errorf("d2_source is required")
	}

	// Load the target (system or container)
	if containerName != "" {
		// Normalize names to IDs
		systemID := entities.NormalizeName(systemName)
		containerID := entities.NormalizeName(containerName)

		// Update container diagram
		container, err := t.repo.LoadContainer(ctx, projectRoot, systemID, containerID)
		if err != nil {
			return nil, fmt.Errorf("failed to load container: %w", err)
		}

		// Update diagram
		diagram, err := entities.NewDiagram("")
		if err != nil {
			return nil, fmt.Errorf("failed to create diagram: %w", err)
		}
		diagram.SetSource(d2Source)
		container.Diagram = diagram

		// Save container
		if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
			return nil, fmt.Errorf("failed to save container: %w", err)
		}

		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Diagram updated for container %q", containerName),
			"type":    "container",
		}, nil
	}

	// Update system diagram
	// Normalize system name to ID
	systemID := entities.NormalizeName(systemName)

	system, err := t.repo.LoadSystem(ctx, projectRoot, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to load system: %w", err)
	}

	// Update diagram
	diagram, err := entities.NewDiagram("")
	if err != nil {
		return nil, fmt.Errorf("failed to create diagram: %w", err)
	}
	diagram.SetSource(d2Source)
	system.Diagram = diagram

	// Save system
	if err := t.repo.SaveSystem(ctx, projectRoot, system); err != nil {
		return nil, fmt.Errorf("failed to save system: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Diagram updated for system %q", systemName),
		"type":    "system",
	}, nil
}

// BuildDocsTool triggers documentation build.
type BuildDocsTool struct {
	repo usecases.ProjectRepository
}

// NewBuildDocsTool creates a new build_docs tool.
func NewBuildDocsTool(repo usecases.ProjectRepository) *BuildDocsTool {
	return &BuildDocsTool{repo: repo}
}

func (t *BuildDocsTool) Name() string {
	return "build_docs"
}

func (t *BuildDocsTool) Description() string {
	return "Build HTML documentation for the project"
}

func (t *BuildDocsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"output_dir": map[string]interface{}{
				"type":        "string",
				"description": "Output directory for HTML files",
			},
		},
		"required": []string{"project_root", "output_dir"},
	}
}

func (t *BuildDocsTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)
	outputDir, _ := args["output_dir"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}
	if outputDir == "" {
		outputDir = "dist"
	}

	// Load the project
	project, err := t.repo.LoadProject(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Load systems
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}

	if len(systems) == 0 {
		return map[string]interface{}{
			"success": true,
			"message": "No systems to build documentation for",
			"output":  outputDir,
		}, nil
	}

	// Create adapters
	diagramRenderer := d2.NewRenderer()
	siteBuilder, err := html.NewBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create site builder: %w", err)
	}

	// Create progress reporter (simple in-memory reporter)
	progressReporter := &mcpProgressReporter{}

	// Create and execute build use case
	buildDocs := usecases.NewBuildDocs(diagramRenderer, siteBuilder, progressReporter)

	err = buildDocs.Execute(ctx, project, systems, outputDir)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Return success with generated files info
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Documentation built successfully in %s", outputDir),
		"output":  outputDir,
		"systems": len(systems),
		"files": map[string]interface{}{
			"index":    "index.html",
			"systems":  len(systems),
			"diagrams": countDiagrams(systems),
		},
	}, nil
}

// ValidateTool validates the architecture.
type ValidateTool struct {
	repo usecases.ProjectRepository
}

// NewValidateTool creates a new validate tool.
func NewValidateTool(repo usecases.ProjectRepository) *ValidateTool {
	return &ValidateTool{repo: repo}
}

func (t *ValidateTool) Name() string {
	return "validate"
}

func (t *ValidateTool) Description() string {
	return "Validate the project architecture for errors and warnings"
}

func (t *ValidateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "Root directory of the project",
			},
		},
		"required": []string{"project_root"},
	}
}

func (t *ValidateTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	projectRoot, _ := args["project_root"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Load systems and validate
	systems, err := t.repo.ListSystems(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load systems: %w", err)
	}

	var warnings []string
	for _, sys := range systems {
		if sys.ContainerCount() == 0 {
			warnings = append(warnings, fmt.Sprintf("System %q has no containers", sys.Name))
		}
	}

	return map[string]interface{}{
		"valid":    len(warnings) == 0,
		"warnings": warnings,
	}, nil
}

// mcpProgressReporter implements ProgressReporter for MCP tool context.
type mcpProgressReporter struct {
}

// ReportProgress reports progress.
func (r *mcpProgressReporter) ReportProgress(step string, current int, total int, message string) {
	// Silent in MCP context; progress is implicit in tool execution
}

// ReportError reports an error.
func (r *mcpProgressReporter) ReportError(err error) {
	// Silent in MCP context; errors are returned directly
}

// ReportSuccess reports success.
func (r *mcpProgressReporter) ReportSuccess(message string) {
	// Silent in MCP context; success is implicit in return value
}

// ReportInfo reports info.
func (r *mcpProgressReporter) ReportInfo(message string) {
	// Silent in MCP context; info is implicit in return value
}

// countDiagrams counts total diagrams in all systems and containers.
func countDiagrams(systems []*entities.System) int {
	count := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			count++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				count++
			}
		}
	}
	return count
}

// createSystemD2Template creates a basic D2 diagram template for a system.
func createSystemD2Template(ctx context.Context, projectRoot string, system *entities.System) error {
	systemDir := filepath.Join(projectRoot, "src", system.ID)
	if err := os.MkdirAll(systemDir, 0755); err != nil {
		return err
	}

	d2Template := fmt.Sprintf(`# %s System Context Diagram
# C4 Level 1 - System Context

direction: right

User: "User" {
  icon: "https://icons.terrastruct.com/essentials/087-user.svg"
}

%s: "%s" {
  icon: "https://icons.terrastruct.com/gcp/compute/Cloud%%20Run.svg"
}

User -> %s: "Uses"
`, system.Name, system.ID, system.Name, system.ID)

	diagramPath := filepath.Join(systemDir, system.ID+".d2")
	return os.WriteFile(diagramPath, []byte(d2Template), 0644)
}

// createContainerD2Template creates a basic D2 diagram template for a container.
func createContainerD2Template(ctx context.Context, projectRoot, systemID string, container *entities.Container) error {
	containerDir := filepath.Join(projectRoot, "src", systemID, container.ID)
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return err
	}

	d2Template := fmt.Sprintf(`# %s Container Diagram
# C4 Level 2 - Container

direction: right

%s: "%s" {
  description: "%s"
  technology: "%s"
}
`, container.Name, container.ID, container.Name, container.Description, container.Technology)

	diagramPath := filepath.Join(containerDir, container.ID+".d2")
	return os.WriteFile(diagramPath, []byte(d2Template), 0644)
}
