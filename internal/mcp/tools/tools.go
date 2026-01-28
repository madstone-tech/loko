package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			"responsibilities": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Key responsibilities (e.g., 'Process payments', 'Store user data')",
			},
			"key_users": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Primary users/actors (e.g., 'User', 'Admin', 'Payment Gateway')",
			},
			"dependencies": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "External dependencies (e.g., 'Database', 'Cache', 'Message Queue')",
			},
			"external_systems": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "External system integrations (e.g., 'Payment API', 'Email Service')",
			},
			"primary_language": map[string]interface{}{
				"type":        "string",
				"description": "Primary programming language (e.g., 'Go', 'Python', 'JavaScript')",
			},
			"framework": map[string]interface{}{
				"type":        "string",
				"description": "Framework/library (e.g., 'Fiber', 'Django', 'React')",
			},
			"database": map[string]interface{}{
				"type":        "string",
				"description": "Database technology (e.g., 'PostgreSQL', 'MongoDB', 'Redis')",
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
	primaryLanguage, _ := args["primary_language"].(string)
	framework, _ := args["framework"].(string)
	database, _ := args["database"].(string)

	if projectRoot == "" {
		projectRoot = "."
	}

	// Convert array interfaces to string slices
	responsibilitiesIface, _ := args["responsibilities"].([]interface{})
	responsibilities := convertInterfaceSlice(responsibilitiesIface)

	keyUsersIface, _ := args["key_users"].([]interface{})
	keyUsers := convertInterfaceSlice(keyUsersIface)

	dependenciesIface, _ := args["dependencies"].([]interface{})
	dependencies := convertInterfaceSlice(dependenciesIface)

	externalSystemsIface, _ := args["external_systems"].([]interface{})
	externalSystems := convertInterfaceSlice(externalSystemsIface)

	tagsIface, _ := args["tags"].([]interface{})
	tags := convertInterfaceSlice(tagsIface)

	// Create system
	uc := usecases.NewCreateSystem(t.repo)
	req := &usecases.CreateSystemRequest{
		Name:             name,
		Description:      description,
		Responsibilities: responsibilities,
		KeyUsers:         keyUsers,
		Dependencies:     dependencies,
		ExternalSystems:  externalSystems,
		PrimaryLanguage:  primaryLanguage,
		Framework:        framework,
		Database:         database,
		Tags:             tags,
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
			"path":             system.Path,
			"diagram":          diagramMsg,
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
				"description": "Container name (e.g., 'API Server', 'Web Frontend', 'Database')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "What does this container do? (e.g., 'Handles all REST API requests')",
			},
			"technology": map[string]interface{}{
				"type":        "string",
				"description": "Technology stack (e.g., 'Go + Fiber', 'Node.js + Express', 'PostgreSQL 15')",
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Tags for categorization (e.g., 'backend', 'database', 'frontend')",
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
	tagsIface, _ := args["tags"].([]interface{})

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
	container.Tags = convertInterfaceSlice(tagsIface)

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

	// Attempt to update parent system's D2 diagram to include the new container (optional)
	updateMsg := ""
	system, sysErr := t.repo.LoadSystem(ctx, projectRoot, systemID)
	if sysErr == nil {
		if err := updateSystemD2Diagram(ctx, projectRoot, system); err == nil {
			updateMsg = " | System D2 auto-synced"
		}
	}

	return map[string]interface{}{
		"container": map[string]interface{}{
			"id":          container.ID,
			"name":        container.Name,
			"description": container.Description,
			"technology":  container.Technology,
			"tags":        container.Tags,
			"diagram":     diagramMsg + updateMsg,
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
				"description": "Component name (e.g., 'Auth Handler', 'Product Service', 'Cache Manager')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "What does this component do? (e.g., 'Handles JWT authentication')",
			},
			"technology": map[string]interface{}{
				"type":        "string",
				"description": "Technology/implementation details (e.g., 'Go', 'React Component', 'Python module')",
			},
			"tags": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Tags for categorization (e.g., 'auth', 'handler', 'service')",
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
	technology, _ := args["technology"].(string)
	tagsIface, _ := args["tags"].([]interface{})

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
	component.Technology = technology
	component.Tags = convertInterfaceSlice(tagsIface)

	// Add to container
	if err := container.AddComponent(component); err != nil {
		return nil, fmt.Errorf("failed to add component to container: %w", err)
	}

	// Save container
	if err := t.repo.SaveContainer(ctx, projectRoot, systemID, container); err != nil {
		return nil, fmt.Errorf("failed to save container: %w", err)
	}

	// Attempt to update parent container's D2 diagram to include the new component (optional)
	syncMsg := ""
	if err := updateContainerD2Diagram(ctx, projectRoot, container); err == nil {
		syncMsg = " | Container D2 auto-synced"
	}

	return map[string]interface{}{
		"component": map[string]interface{}{
			"id":          component.ID,
			"name":        component.Name,
			"description": component.Description,
			"technology":  component.Technology,
			"tags":        component.Tags,
			"sync":        syncMsg,
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

// ValidateDiagramTool validates D2 diagram source code.
type ValidateDiagramTool struct {
	renderer usecases.DiagramRenderer
}

// NewValidateDiagramTool creates a new validate_diagram tool.
func NewValidateDiagramTool(renderer usecases.DiagramRenderer) *ValidateDiagramTool {
	return &ValidateDiagramTool{renderer: renderer}
}

func (t *ValidateDiagramTool) Name() string {
	return "validate_diagram"
}

func (t *ValidateDiagramTool) Description() string {
	return `Validate D2 diagram source code and report syntax errors.
This tool checks if D2 source code is syntactically valid and provides helpful error messages if there are issues.
It also provides recommendations for improving diagram structure and C4 Model compliance.`
}

func (t *ValidateDiagramTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"d2_source": map[string]interface{}{
				"type":        "string",
				"description": "The D2 diagram source code to validate",
			},
			"level": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"system", "container", "component"},
				"description": "C4 Model level for context-aware validation",
			},
		},
		"required": []string{"d2_source"},
	}
}

func (t *ValidateDiagramTool) Call(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	d2Source, _ := args["d2_source"].(string)
	level, _ := args["level"].(string)

	if d2Source == "" {
		return map[string]interface{}{
			"valid":  false,
			"errors": []string{"d2_source cannot be empty"},
		}, nil
	}

	// Validate D2 syntax by attempting to render
	result := map[string]interface{}{
		"valid":        true,
		"errors":       []string{},
		"warnings":     []string{},
		"suggestions":  []string{},
		"syntax_valid": false,
		"d2_available": t.renderer.IsAvailable(),
	}

	// Try to render the diagram
	if t.renderer.IsAvailable() {
		_, err := t.renderer.RenderDiagram(ctx, d2Source)
		if err != nil {
			result["valid"] = false
			result["errors"] = []string{fmt.Sprintf("D2 syntax error: %v", err)}
		} else {
			result["syntax_valid"] = true
		}
	} else {
		result["warnings"] = []string{"D2 CLI not available - syntax validation skipped. Install D2 from https://d2lang.com"}
	}

	// Perform structural validation
	warnings, suggestions := validateDiagramStructure(d2Source, level)
	result["warnings"] = warnings
	result["suggestions"] = suggestions

	// Overall validity check
	errors := result["errors"].([]string)
	result["valid"] = len(errors) == 0 && result["syntax_valid"].(bool)

	return result, nil
}

// validateDiagramStructure checks for structural and C4 compliance issues.
func validateDiagramStructure(d2Source, level string) ([]string, []string) {
	var warnings []string
	var suggestions []string

	// Check for comments
	if !containsSubstring(d2Source, "#") {
		suggestions = append(suggestions, "Add comments to explain diagram structure")
	}

	// Level-specific checks
	switch level {
	case "system":
		if !containsSubstring(d2Source, "User") && !containsSubstring(d2Source, "user") {
			suggestions = append(suggestions, "C4 Level 1 typically includes 'User' - consider adding user/actor")
		}
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "System context diagram should have at least 2 nodes (User and System)")
		}

	case "container":
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "Container diagram should have at least 2 components")
		}
		if !containsSubstring(d2Source, "{\n") {
			suggestions = append(suggestions, "Consider grouping related components with container blocks { }")
		}

	case "component":
		if countDiagramNodes(d2Source) < 1 {
			warnings = append(warnings, "Component diagram should have at least 1 component")
		}
	}

	// General best practices
	if !containsSubstring(d2Source, "direction:") && !containsSubstring(d2Source, "direction ") {
		suggestions = append(suggestions, "Consider specifying diagram direction (e.g., 'direction: right') for clarity")
	}

	if !containsSubstring(d2Source, "tooltip") && !containsSubstring(d2Source, "description") {
		suggestions = append(suggestions, "Add tooltips or descriptions to nodes for better documentation")
	}

	return warnings, suggestions
}

// Helper functions for validation
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

func countDiagramNodes(d2Source string) int {
	count := 0
	// Count lines with nodes (heuristic: lines with colons that aren't comments or directives)
	for _, line := range strings.Split(d2Source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "direction") {
			count++
		}
	}
	return count
}

// convertInterfaceSlice converts a slice of interface{} to a slice of strings.
func convertInterfaceSlice(ifaces []interface{}) []string {
	result := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		if str, ok := iface.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// updateSystemD2Diagram updates a system's D2 diagram with its current containers.
// This mirrors the CLI behavior of auto-syncing diagrams when containers are added.
func updateSystemD2Diagram(_ context.Context, projectRoot string, system *entities.System) error {
	if system.Path == "" {
		return fmt.Errorf("system path not set")
	}

	// Generate container diagram content
	d2Content := generateContainerDiagram(system)

	// Write to system D2 file
	d2Path := filepath.Join(system.Path, system.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}

// updateContainerD2Diagram updates a container's D2 diagram with its current components.
// This mirrors the CLI behavior of auto-syncing diagrams when components are added.
func updateContainerD2Diagram(_ context.Context, projectRoot string, container *entities.Container) error {
	if container.Path == "" {
		return fmt.Errorf("container path not set")
	}

	// Generate component diagram content
	d2Content := generateComponentDiagram(container)

	// Write to container D2 file
	d2Path := filepath.Join(container.Path, container.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}

// generateContainerDiagram creates a C4 Level 2 container diagram.
func generateContainerDiagram(system *entities.System) string {
	var sb strings.Builder

	sb.WriteString("# Container Diagram\n")
	sb.WriteString("# C4 Level 2 - Container View\n")
	sb.WriteString(fmt.Sprintf("# System: %s\n\n", system.Name))

	sb.WriteString("direction: right\n\n")

	// Add users
	sb.WriteString("# External users\n")
	if len(system.KeyUsers) > 0 {
		for i, user := range system.KeyUsers {
			userID := fmt.Sprintf("user_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s: \"%s\" { style { fill: \"#FFF3E0\" } }\n", userID, user))
		}
	} else {
		sb.WriteString("user: \"User/Actor\" { style { fill: \"#FFF3E0\" } }\n")
	}
	sb.WriteString("\n")

	// Add system as container group
	sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", system.ID, system.Name))
	sb.WriteString(fmt.Sprintf("  description: \"%s\"\n\n", system.Description))

	// Add containers
	if len(system.Containers) > 0 {
		for _, container := range system.Containers {
			sb.WriteString(fmt.Sprintf("  %s: \"%s\" {\n", container.ID, container.Name))
			if container.Description != "" {
				sb.WriteString(fmt.Sprintf("    description: \"%s\"\n", container.Description))
			}
			if container.Technology != "" {
				sb.WriteString(fmt.Sprintf("    technology: \"%s\"\n", container.Technology))
			}
			sb.WriteString("    style { fill: \"#E3F2FD\" }\n")
			sb.WriteString("  }\n")
		}
	} else {
		sb.WriteString("  # (Add containers here)\n")
	}

	sb.WriteString("}\n\n")

	// Add relationships
	sb.WriteString("# User interactions\n")
	if len(system.KeyUsers) > 0 {
		for i := range system.KeyUsers {
			userID := fmt.Sprintf("user_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s -> %s: \"Uses\"\n", userID, system.ID))
		}
	} else {
		sb.WriteString(fmt.Sprintf("user -> %s: \"Uses\"\n", system.ID))
	}

	sb.WriteString("\n")

	// System styling
	sb.WriteString(fmt.Sprintf("%s: {\n", system.ID))
	sb.WriteString("  style {\n")
	sb.WriteString("    fill: \"#E1F5FF\"\n")
	sb.WriteString("    stroke: \"#01579B\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateComponentDiagram creates a C4 Level 3 component diagram.
func generateComponentDiagram(container *entities.Container) string {
	var sb strings.Builder

	sb.WriteString("# Component Diagram\n")
	sb.WriteString("# C4 Level 3 - Component View\n")
	sb.WriteString(fmt.Sprintf("# Container: %s\n\n", container.Name))

	sb.WriteString("direction: right\n\n")

	// Add components
	if len(container.Components) > 0 {
		sb.WriteString("# Components\n")
		for _, component := range container.Components {
			sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", component.ID, component.Name))
			if component.Description != "" {
				sb.WriteString(fmt.Sprintf("  description: \"%s\"\n", component.Description))
			}
			if component.Technology != "" {
				sb.WriteString(fmt.Sprintf("  technology: \"%s\"\n", component.Technology))
			}
			sb.WriteString("  style { fill: \"#E3F2FD\" }\n")
			sb.WriteString("}\n")
		}
	} else {
		sb.WriteString("# (Add components here)\n")
	}

	sb.WriteString("\n")

	// Component relationships (optional)
	if len(container.Components) > 1 {
		sb.WriteString("# Component interactions (add as needed)\n")
		components := make([]*entities.Component, 0, len(container.Components))
		for _, c := range container.Components {
			components = append(components, c)
		}
		if len(components) >= 2 {
			sb.WriteString(fmt.Sprintf("# %s -> %s: \"Communicates via\"\n",
				components[0].ID, components[1].ID))
		}
	}

	sb.WriteString("\n")

	// Styling
	sb.WriteString(fmt.Sprintf("%s: {\n", container.ID))
	sb.WriteString("  style {\n")
	sb.WriteString("    fill: \"#E3F2FD\"\n")
	sb.WriteString("    stroke: \"#01579B\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}
