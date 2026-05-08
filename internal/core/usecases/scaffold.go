package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// SelectComponentTemplateName maps a technology string to the template filename
// (without extension) used for component .md scaffolding (T055/T057).
// Returns "component" when technology is empty or unrecognised.
// This function lives in usecases so that cmd/ need not import entities directly.
func SelectComponentTemplateName(technology string) string {
	if technology == "" {
		return "component"
	}
	selector := entities.NewTemplateSelector()
	category, matched := selector.SelectTemplateCategory(technology)
	if !matched {
		return "component"
	}
	switch category {
	case entities.TemplateCategoryCompute:
		return "compute"
	case entities.TemplateCategoryDatastore:
		return "datastore"
	case entities.TemplateCategoryMessaging:
		return "messaging"
	case entities.TemplateCategoryAPI:
		return "api"
	case entities.TemplateCategoryEvent:
		return "event"
	case entities.TemplateCategoryStorage:
		return "storage"
	default:
		return "generic"
	}
}

// ScaffoldEntityRequest defines the input for the ScaffoldEntity use case.
type ScaffoldEntityRequest struct {
	ProjectRoot     string   // filesystem path to project
	EntityType      string   // "system" | "container" | "component"
	ParentPath      []string // hierarchy path: [] for system, [system] for container, [system, container] for component
	Name            string   // entity display name
	Description     string   // optional description
	Technology      string   // optional technology string
	Tags            []string // optional tags
	Template        string   // template name (empty = use project default)
	ContentTemplate string   // T055: technology-specific component content template (e.g. "compute", "datastore")
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
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

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

// renderTemplates renders the named template and writes the output alongside the entity.
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
