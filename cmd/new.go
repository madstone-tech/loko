package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/cli"
	d2adapter "github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// NewCommand creates new C4 entities (system, container, component).
type NewCommand struct {
	entityType   string // "system", "container", "component"
	entityName   string
	parentName   string // For container/component: parent system/container
	description  string
	technology   string
	projectRoot  string
	templateName string // Template to use (default: "standard-3layer")
	autoTemplate bool   // Whether to auto-select template based on technology
	preview      bool   // Whether to show diagram preview after creation
}

// NewNewCommand creates a new 'new' command.
func NewNewCommand(entityType, entityName string) *NewCommand {
	return &NewCommand{
		entityType: entityType,
		entityName: entityName,
	}
}

// WithParent sets the parent entity name (for containers/components).
func (nc *NewCommand) WithParent(parent string) *NewCommand {
	nc.parentName = parent
	return nc
}

// WithDescription sets the entity description.
func (nc *NewCommand) WithDescription(desc string) *NewCommand {
	nc.description = desc
	return nc
}

// WithTechnology sets the technology stack.
func (nc *NewCommand) WithTechnology(tech string) *NewCommand {
	nc.technology = tech
	return nc
}

// WithProjectRoot sets the project root directory.
func (nc *NewCommand) WithProjectRoot(root string) *NewCommand {
	nc.projectRoot = root
	return nc
}

// WithTemplate sets the template to use for scaffolding.
func (nc *NewCommand) WithTemplate(name string) *NewCommand {
	nc.templateName = name
	return nc
}

// WithAutoTemplate enables automatic template selection based on technology.
func (nc *NewCommand) WithAutoTemplate(auto bool) *NewCommand {
	nc.autoTemplate = auto
	return nc
}

// WithPreview enables diagram preview after creation.
func (nc *NewCommand) WithPreview(preview bool) *NewCommand {
	nc.preview = preview
	return nc
}

// Execute runs the new command.
func (nc *NewCommand) Execute(ctx context.Context) error {
	if nc.entityName == "" {
		return fmt.Errorf("entity name is required")
	}
	if nc.projectRoot == "" {
		var err error
		nc.projectRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	templateName := nc.templateName
	if templateName == "" {
		if nc.autoTemplate && nc.technology != "" {
			// Auto-select template based on technology
			templateSelector := entities.NewTemplateSelector()
			_, _ = templateSelector.SelectTemplateCategory(nc.technology)
			// In a real implementation, we would map categories to actual template names
			// For now, we'll use a placeholder approach
			templateName = "standard-3layer" // Default fallback
		} else {
			templateName = "standard-3layer"
		}
	}
	if err := nc.validateTemplate(templateName); err != nil {
		return err
	}

	if nc.entityType == "system" && nc.description == "" {
		if err := nc.gatherSystemDetails(); err != nil {
			return err
		}
	}

	req, err := nc.buildScaffoldRequest(ctx, templateName)
	if err != nil {
		return err
	}

	result, err := nc.executeScaffold(ctx, req, templateName)
	if err != nil {
		return fmt.Errorf("failed to scaffold %s: %w", nc.entityType, err)
	}

	// Convert first letter to uppercase for display
	entityTypeDisplay := nc.entityType
	if len(entityTypeDisplay) > 0 {
		entityTypeDisplay = strings.ToUpper(string(entityTypeDisplay[0])) + entityTypeDisplay[1:]
	}
	fmt.Printf("\n✅ %s '%s' created successfully!\n", entityTypeDisplay, nc.entityName)
	if result.DiagramPath != "" {
		fmt.Printf("✓ D2 diagram: %s\n", result.DiagramPath)
	}

	// Handle preview if requested
	if nc.preview && nc.entityType == "component" {
		if err := nc.showPreview(ctx, req); err != nil {
			fmt.Printf("⚠️  Preview failed: %v\n", err)
		}
	}

	return nil
}

// buildScaffoldRequest creates the scaffold request with parent path resolution.
func (nc *NewCommand) buildScaffoldRequest(ctx context.Context, templateName string) (*usecases.ScaffoldEntityRequest, error) {
	req := &usecases.ScaffoldEntityRequest{
		ProjectRoot: nc.projectRoot,
		EntityType:  nc.entityType,
		Name:        nc.entityName,
		Description: nc.description,
		Technology:  nc.technology,
		Template:    templateName,
	}

	switch nc.entityType {
	case "container":
		if nc.parentName == "" {
			return nil, fmt.Errorf("parent system name is required for container")
		}
		req.ParentPath = []string{nc.parentName}
	case "component":
		if nc.parentName == "" {
			return nil, fmt.Errorf("parent container name is required for component")
		}
		req.ParentPath = nc.resolveComponentParent(ctx)
		// T055: set technology-specific content template for component .md file
		req.ContentTemplate = nc.componentTemplateName()
	}

	return req, nil
}

// executeScaffold creates and runs the scaffold use case with adapters.
func (nc *NewCommand) executeScaffold(ctx context.Context, req *usecases.ScaffoldEntityRequest, templateName string) (*usecases.ScaffoldEntityResult, error) {
	repo := filesystem.NewProjectRepository()
	templateEngine := nc.createTemplateEngine(templateName)
	repo.SetTemplateEngine(templateEngine)

	scaffold := usecases.NewScaffoldEntity(repo,
		usecases.WithTemplateEngine(templateEngine),
		usecases.WithDiagramGenerator(d2adapter.NewGenerator()),
	)
	return scaffold.Execute(ctx, req)
}

// gatherSystemDetails prompts for interactive system details.
func (nc *NewCommand) gatherSystemDetails() error {
	prompts := cli.NewPrompts(bufio.NewReader(os.Stdin))
	desc, err := prompts.PromptString("System description", "")
	if err != nil {
		return fmt.Errorf("failed to read description: %w", err)
	}
	nc.description = desc
	return nil
}

// resolveComponentParent finds the parent system for a component's container.
func (nc *NewCommand) resolveComponentParent(ctx context.Context) []string {
	repo := filesystem.NewProjectRepository()
	project, err := repo.LoadProject(ctx, nc.projectRoot)
	if err != nil {
		return []string{"", nc.parentName}
	}

	// Search through systems to find which one contains the parent container
	for _, system := range project.Systems {
		for _, container := range system.Containers {
			if container.ID == nc.parentName || container.Name == nc.parentName {
				return []string{system.ID, container.ID}
			}
		}
	}

	return []string{"", nc.parentName}
}

// createTemplateEngine creates a template engine with standard search paths.
// The component category templates directory is always included so that
// technology-specific component files (compute.md, datastore.md, etc.) are
// available alongside the scaffolding template files.
func (nc *NewCommand) createTemplateEngine(templateName string) *ason.TemplateEngine {
	templateEngine := ason.NewTemplateEngine()
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		templateEngine.AddSearchPath(filepath.Join(exeDir, "..", "templates", templateName))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", templateName))
		// T057: Add component category templates (compute.md, datastore.md, etc.)
		templateEngine.AddSearchPath(filepath.Join(exeDir, "..", "templates", "component"))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", "component"))
	} else {
		templateEngine.AddSearchPath(filepath.Join(".", "templates", templateName))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", "component"))
	}
	return templateEngine
}

// componentTemplateName resolves the template filename for a component based on
// its technology field, honoring any explicit template override (T055/T057).
func (nc *NewCommand) componentTemplateName() string {
	if nc.templateName != "" {
		return nc.templateName
	}
	if nc.technology == "" {
		return "component" // generic fallback
	}
	selector := entities.NewTemplateSelector()
	category, matched := selector.SelectTemplateCategory(nc.technology)
	if !matched {
		return "component"
	}
	// Map category to template filename (without .md suffix)
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

// showPreview renders and displays a diagram preview for the newly created component.
func (nc *NewCommand) showPreview(ctx context.Context, req *usecases.ScaffoldEntityRequest) error {
	// Create a diagram renderer
	renderer := d2adapter.NewRenderer()

	// Check if renderer is available
	if !renderer.IsAvailable() {
		return fmt.Errorf("d2 binary not found in PATH - install from https://d2lang.com/")
	}

	// Create preview renderer
	previewRenderer := d2adapter.NewPreviewRenderer(renderer)

	// Get container name
	containerName := ""
	if len(req.ParentPath) > 0 {
		containerName = req.ParentPath[len(req.ParentPath)-1]
	}

	// Render preview
	svgContent, err := previewRenderer.RenderComponentPreview(ctx, nc.entityName, nc.technology, containerName)
	if err != nil {
		return fmt.Errorf("failed to render preview: %w", err)
	}

	// Display preview
	fmt.Println("\n📊 Diagram Preview:")
	fmt.Println("==================")
	fmt.Println(svgContent)

	return nil
}

// validateTemplate checks if the specified template exists.
func (nc *NewCommand) validateTemplate(templateName string) error {
	relPath := filepath.Join(".", "templates", templateName)
	if info, err := os.Stat(relPath); err == nil && info.IsDir() {
		return nil
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		binPath := filepath.Join(exeDir, "..", "templates", templateName)
		if info, err := os.Stat(binPath); err == nil && info.IsDir() {
			return nil
		}
	}

	available := nc.listAvailableTemplates()
	return fmt.Errorf("template %q not found. Available templates: %s", templateName, strings.Join(available, ", "))
}

// listAvailableTemplates returns a list of available template names.
func (nc *NewCommand) listAvailableTemplates() []string {
	var templates []string
	seen := make(map[string]bool)

	if entries, err := os.ReadDir(filepath.Join(".", "templates")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !seen[entry.Name()] {
				templates = append(templates, entry.Name())
				seen[entry.Name()] = true
			}
		}
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if entries, err := os.ReadDir(filepath.Join(exeDir, "..", "templates")); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !seen[entry.Name()] {
					templates = append(templates, entry.Name())
					seen[entry.Name()] = true
				}
			}
		}
	}

	if len(templates) == 0 {
		return []string{"standard-3layer"}
	}
	return templates
}
