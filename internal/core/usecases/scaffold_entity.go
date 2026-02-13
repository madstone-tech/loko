package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// ScaffoldEntityRequest defines the input for the ScaffoldEntity use case.
type ScaffoldEntityRequest struct {
	ProjectRoot string   // filesystem path to project
	EntityType  string   // "system" | "container" | "component"
	ParentPath  []string // hierarchy path: [] for system, [system] for container, [system, container] for component
	Name        string   // entity display name
	Description string   // optional description
	Technology  string   // optional technology string
	Tags        []string // optional tags
	Template    string   // template name (empty = use project default)
}

// ScaffoldEntityResult defines the output of the ScaffoldEntity use case.
type ScaffoldEntityResult struct {
	EntityID     string   // normalized ID of created entity
	FilesCreated []string // all files created/modified
	DiagramPath  string   // path to generated D2 diagram (empty if no diagram)
}

// ScaffoldEntity orchestrates the full entity creation workflow.
type ScaffoldEntity struct {
	projectRepo      ProjectRepository
	templateEngine   TemplateEngine
	diagramGenerator DiagramGenerator
	logger           Logger
}

// ScaffoldEntityOption is a functional option for configuring ScaffoldEntity.
type ScaffoldEntityOption func(*ScaffoldEntity)

// WithTemplateEngine sets the optional template engine.
func WithTemplateEngine(te TemplateEngine) ScaffoldEntityOption {
	return func(s *ScaffoldEntity) {
		s.templateEngine = te
	}
}

// WithDiagramGenerator sets the optional diagram generator.
func WithDiagramGenerator(dg DiagramGenerator) ScaffoldEntityOption {
	return func(s *ScaffoldEntity) {
		s.diagramGenerator = dg
	}
}

// WithLogger sets the optional logger.
func WithLogger(l Logger) ScaffoldEntityOption {
	return func(s *ScaffoldEntity) {
		s.logger = l
	}
}

// NewScaffoldEntity creates a new ScaffoldEntity use case.
func NewScaffoldEntity(repo ProjectRepository, opts ...ScaffoldEntityOption) *ScaffoldEntity {
	uc := &ScaffoldEntity{
		projectRepo: repo,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// Execute orchestrates the entity creation workflow.
func (uc *ScaffoldEntity) Execute(ctx context.Context, req *ScaffoldEntityRequest) (*ScaffoldEntityResult, error) {
	if uc.logger != nil {
		uc.logger.Info("scaffolding entity", "type", req.EntityType, "name", req.Name)
	}

	// Load the project
	project, err := uc.projectRepo.LoadProject(ctx, req.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	result := &ScaffoldEntityResult{
		FilesCreated: []string{},
	}

	switch req.EntityType {
	case "system":
		if err := uc.scaffoldSystem(ctx, req, project, result); err != nil {
			return nil, err
		}
	case "container":
		if err := uc.scaffoldContainer(ctx, req, project, result); err != nil {
			return nil, err
		}
	case "component":
		if err := uc.scaffoldComponent(ctx, req, project, result); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown entity type: %s", req.EntityType)
	}

	// Optionally render templates
	if uc.templateEngine != nil && req.Template != "" {
		if err := uc.renderTemplates(ctx, req, result); err != nil {
			return nil, fmt.Errorf("failed to render templates: %w", err)
		}
	}

	if uc.logger != nil {
		uc.logger.Info("scaffolded entity", "type", req.EntityType, "id", result.EntityID)
	}

	return result, nil
}

func (uc *ScaffoldEntity) scaffoldSystem(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Create system entity
	system, err := entities.NewSystem(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	// Set optional fields
	system.Description = req.Description
	if len(req.Tags) > 0 {
		system.Tags = req.Tags
	}

	// Set path
	system.Path = filepath.Join(req.ProjectRoot, project.Config.SourceDir, system.ID)

	// Add to project
	if err := project.AddSystem(system); err != nil {
		return fmt.Errorf("failed to add system to project: %w", err)
	}

	// Save system
	if err := uc.projectRepo.SaveSystem(ctx, req.ProjectRoot, system); err != nil {
		return fmt.Errorf("failed to save system: %w", err)
	}

	result.EntityID = system.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(system.Path, "system.toml"))

	// Generate system context diagram
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateSystemContextDiagram(system)
		if err != nil {
			return fmt.Errorf("failed to generate system context diagram: %w", err)
		}

		d2Path := filepath.Join(system.Path, "system.d2")
		if err := os.MkdirAll(system.Path, 0755); err != nil {
			return fmt.Errorf("failed to create system directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}

func (uc *ScaffoldEntity) scaffoldContainer(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Validate parent path
	if len(req.ParentPath) == 0 {
		return fmt.Errorf("parent path must contain system name for container")
	}

	// Load parent system
	systemID := entities.NormalizeName(req.ParentPath[0])
	system, err := uc.projectRepo.LoadSystem(ctx, req.ProjectRoot, systemID)
	if err != nil {
		return fmt.Errorf("failed to load parent system: %w", err)
	}

	// Create container entity
	container, err := entities.NewContainer(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Set optional fields
	container.Description = req.Description
	if req.Technology != "" {
		container.Technology = req.Technology
	}
	if len(req.Tags) > 0 {
		container.Tags = req.Tags
	}

	// Set path
	container.Path = filepath.Join(system.Path, container.ID)

	// Add to system
	if err := system.AddContainer(container); err != nil {
		return fmt.Errorf("failed to add container to system: %w", err)
	}

	// Save container
	if err := uc.projectRepo.SaveContainer(ctx, req.ProjectRoot, systemID, container); err != nil {
		return fmt.Errorf("failed to save container: %w", err)
	}

	result.EntityID = container.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(container.Path, "container.toml"))

	// Generate container diagram (shows all containers in system)
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateContainerDiagram(system)
		if err != nil {
			return fmt.Errorf("failed to generate container diagram: %w", err)
		}

		d2Path := filepath.Join(container.Path, "container.d2")
		if err := os.MkdirAll(container.Path, 0755); err != nil {
			return fmt.Errorf("failed to create container directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}

func (uc *ScaffoldEntity) scaffoldComponent(ctx context.Context, req *ScaffoldEntityRequest, project *entities.Project, result *ScaffoldEntityResult) error {
	// Validate parent path
	if len(req.ParentPath) < 2 {
		return fmt.Errorf("parent path must contain system and container names for component")
	}

	// Load parent system
	systemID := entities.NormalizeName(req.ParentPath[0])
	system, err := uc.projectRepo.LoadSystem(ctx, req.ProjectRoot, systemID)
	if err != nil {
		return fmt.Errorf("failed to load parent system: %w", err)
	}

	// Find container in system
	containerID := entities.NormalizeName(req.ParentPath[1])
	container, ok := system.Containers[containerID]
	if !ok {
		return fmt.Errorf("container %s not found in system %s", containerID, systemID)
	}

	// Create component entity
	component, err := entities.NewComponent(req.Name)
	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	// Set optional fields
	component.Description = req.Description
	if req.Technology != "" {
		component.Technology = req.Technology
	}
	if len(req.Tags) > 0 {
		component.Tags = req.Tags
	}

	// Set path
	component.Path = filepath.Join(container.Path, component.ID)

	// Add to container
	if err := container.AddComponent(component); err != nil {
		return fmt.Errorf("failed to add component to container: %w", err)
	}

	// Save component
	if err := uc.projectRepo.SaveComponent(ctx, req.ProjectRoot, systemID, containerID, component); err != nil {
		return fmt.Errorf("failed to save component: %w", err)
	}

	result.EntityID = component.ID
	result.FilesCreated = append(result.FilesCreated, filepath.Join(component.Path, "component.toml"))

	// Generate component diagram
	if uc.diagramGenerator != nil {
		d2Source, err := uc.diagramGenerator.GenerateComponentDiagram(container)
		if err != nil {
			return fmt.Errorf("failed to generate component diagram: %w", err)
		}

		d2Path := filepath.Join(component.Path, "component.d2")
		if err := os.MkdirAll(component.Path, 0755); err != nil {
			return fmt.Errorf("failed to create component directory: %w", err)
		}
		if err := os.WriteFile(d2Path, []byte(d2Source), 0644); err != nil {
			return fmt.Errorf("failed to write D2 diagram: %w", err)
		}

		result.DiagramPath = d2Path
		result.FilesCreated = append(result.FilesCreated, d2Path)
	}

	return nil
}

func (uc *ScaffoldEntity) renderTemplates(ctx context.Context, req *ScaffoldEntityRequest, result *ScaffoldEntityResult) error {
	if uc.logger != nil {
		uc.logger.Info("rendering template", "template", req.Template)
	}

	// Build template variables
	variables := map[string]string{
		"name":        req.Name,
		"description": req.Description,
		"technology":  req.Technology,
		"entity_id":   result.EntityID,
		"entity_type": req.EntityType,
	}

	// Render template
	content, err := uc.templateEngine.RenderTemplate(ctx, req.Template, variables)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", req.Template, err)
	}

	// Write rendered content to appropriate location
	var outputPath string
	switch req.EntityType {
	case "system":
		outputPath = filepath.Join(req.ProjectRoot, result.EntityID, "README.md")
	case "container":
		systemID := entities.NormalizeName(req.ParentPath[0])
		outputPath = filepath.Join(req.ProjectRoot, systemID, result.EntityID, "README.md")
	case "component":
		systemID := entities.NormalizeName(req.ParentPath[0])
		containerID := entities.NormalizeName(req.ParentPath[1])
		outputPath = filepath.Join(req.ProjectRoot, systemID, containerID, result.EntityID, "README.md")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for template output: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template output: %w", err)
	}

	result.FilesCreated = append(result.FilesCreated, outputPath)

	return nil
}
