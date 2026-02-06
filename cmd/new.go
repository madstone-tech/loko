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
}

// NewNewCommand creates a new 'new' command.
// entityType should be "system", "container", or "component".
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

// Execute runs the new command.
// Creates a new system, container, or component in the project.
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

	// Use default template if not specified
	templateName := nc.templateName
	if templateName == "" {
		templateName = "standard-3layer"
	}

	// Validate template exists
	if err := nc.validateTemplate(templateName); err != nil {
		return err
	}

	// Create template engine and add search paths
	templateEngine := ason.NewTemplateEngine()
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		templateEngine.AddSearchPath(filepath.Join(exeDir, "..", "templates", templateName))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", templateName))
	}

	// Load project
	repo := filesystem.NewProjectRepository()
	repo.SetTemplateEngine(templateEngine)
	project, err := repo.LoadProject(ctx, nc.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	switch nc.entityType {
	case "system":
		return nc.createSystem(ctx, repo, project, templateEngine)
	case "container":
		return nc.createContainer(ctx, repo, project, templateEngine)
	case "component":
		return nc.createComponent(ctx, repo, project, templateEngine)
	default:
		return fmt.Errorf("unknown entity type: %s", nc.entityType)
	}
}

// createSystem creates a new system.
func (nc *NewCommand) createSystem(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project, templateEngine *ason.TemplateEngine) error {
	// Create interactive prompts if description not provided
	prompts := cli.NewPrompts(bufio.NewReader(os.Stdin))

	// Get description if not provided
	description := nc.description
	if description == "" {
		description = prompts.PromptString("System description", "")
	}

	// Prompt for system details
	fmt.Println("\n📋 System Details")
	fmt.Println("================")

	responsibilities := prompts.PromptStringMulti("Key responsibilities (e.g., Process payments, Store data)")
	keyUsers := prompts.PromptStringMulti("Key users/actors (e.g., User, Admin, Payment Gateway)")
	dependencies := prompts.PromptStringMulti("External dependencies (e.g., Database, Cache)")
	externalSystems := prompts.PromptStringMulti("External systems integration (e.g., Payment API, Email Service)")

	fmt.Println("\n🔧 Technology Stack")
	fmt.Println("===================")

	primaryLanguage := prompts.PromptString("Primary language", "Go")
	framework := prompts.PromptString("Framework/Library", "")
	database := prompts.PromptString("Database", "")

	// Use CreateSystem use case with full details
	uc := usecases.NewCreateSystem(repo)
	system, err := uc.Execute(ctx, &usecases.CreateSystemRequest{
		Name:             nc.entityName,
		Description:      description,
		Responsibilities: responsibilities,
		KeyUsers:         keyUsers,
		Dependencies:     dependencies,
		ExternalSystems:  externalSystems,
		PrimaryLanguage:  primaryLanguage,
		Framework:        framework,
		Database:         database,
	})
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	// Set path
	system.Path = filepath.Join(project.Path, project.Config.SourceDir, system.ID)

	// Add to project
	if err := project.AddSystem(system); err != nil {
		return fmt.Errorf("failed to add system to project: %w", err)
	}

	// Save system
	if err := repo.SaveSystem(ctx, project.Path, system); err != nil {
		return fmt.Errorf("failed to save system: %w", err)
	}

	// Create D2 diagram using template engine if available, fallback to generator
	if err := nc.createSystemD2(ctx, system, templateEngine); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create system D2 template: %v\n", err)
	}

	fmt.Printf("\n✅ System '%s' created successfully!\n", system.Name)
	return nil
}

// createContainer creates a new container in a system.
func (nc *NewCommand) createContainer(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project, templateEngine *ason.TemplateEngine) error {
	if nc.parentName == "" {
		return fmt.Errorf("parent system name is required for container")
	}

	// Get parent system
	system, err := project.GetSystem(entities.NormalizeName(nc.parentName))
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	// Create container
	container, err := entities.NewContainer(nc.entityName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	container.Description = nc.description
	container.Technology = nc.technology
	container.Path = filepath.Join(system.Path, container.ID)

	// Add to system
	if err := system.AddContainer(container); err != nil {
		return fmt.Errorf("failed to add container to system: %w", err)
	}

	// Save container
	if err := repo.SaveContainer(ctx, project.Path, system.ID, container); err != nil {
		return fmt.Errorf("failed to save container: %w", err)
	}

	// Create D2 diagram using template engine if available, fallback to hardcoded
	if err := nc.createContainerD2(ctx, container, templateEngine); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create D2 template: %v\n", err)
	}

	// Update parent system's D2 diagram to include the new container
	d2Gen := NewD2Generator()
	if err := d2Gen.UpdateSystemD2File(system); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not update parent system D2 diagram: %v\n", err)
	} else {
		fmt.Printf("✓ Updated %s D2 diagram with new container\n", system.Name)
	}

	return nil
}

// createComponent creates a new component in a container.
func (nc *NewCommand) createComponent(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project, templateEngine *ason.TemplateEngine) error {
	if nc.parentName == "" {
		return fmt.Errorf("parent container name is required for component")
	}

	// We need to find which system and container this component belongs to
	// Search through all systems and containers to find the parent
	var targetSystem *entities.System
	var targetContainer *entities.Container

	for _, system := range project.Systems {
		for _, container := range system.Containers {
			if container.ID == entities.NormalizeName(nc.parentName) {
				targetSystem = system
				targetContainer = container
				break
			}
		}
		if targetContainer != nil {
			break
		}
	}

	if targetContainer == nil {
		return fmt.Errorf("failed to find parent container: %s", nc.parentName)
	}

	// Create component
	component, err := entities.NewComponent(nc.entityName)
	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	component.Description = nc.description
	component.Technology = nc.technology
	component.Path = filepath.Join(targetContainer.Path, component.ID)

	// Add to container
	if err := targetContainer.AddComponent(component); err != nil {
		return fmt.Errorf("failed to add component to container: %w", err)
	}

	// Save component
	if err := repo.SaveComponent(ctx, project.Path, targetSystem.ID, targetContainer.ID, component); err != nil {
		return fmt.Errorf("failed to save component: %w", err)
	}

	// Create D2 diagram using template engine if available, fallback to hardcoded
	if err := nc.createComponentD2(ctx, component, templateEngine); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create D2 template: %v\n", err)
	}

	// Update parent container's D2 diagram to include the new component
	d2Gen := NewD2Generator()
	if err := d2Gen.UpdateContainerD2File(targetContainer); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not update parent container D2 diagram: %v\n", err)
	} else {
		fmt.Printf("✓ Updated %s D2 diagram with new component\n", targetContainer.Name)
	}

	return nil
}

// createContainerD2Template creates a container diagram template.
func (nc *NewCommand) createContainerD2Template(container *entities.Container) error {
	var desc string
	if container.Description != "" {
		desc = fmt.Sprintf("\n  description: %q", container.Description)
	}
	var tech string
	if container.Technology != "" {
		tech = fmt.Sprintf("\n  technology: %q", container.Technology)
	}

	d2Content := fmt.Sprintf(`# %s Container Diagram
# C4 Level 2 - Container
# Architecture: %s
# Description: %s

direction: right

%s: "%s" {%s%s
}

# Internal components (uncomment):
# Handler: "HTTP Handler" { style { fill: "#E3F2FD" } }
# Service: "Service" { style { fill: "#F3E5F5" } }
# Repository: "Repository" { style { fill: "#E8F5E9" } }

# Handler -> Service: "uses"
# Service -> Repository: "queries"

# External dependencies:
# Database: "Database" { 
#   icon: "https://icons.terrastruct.com/gcp/databases/Cloud%%20SQL.svg"
# }
# Service -> Database: "queries"

%s: {
  style {
    fill: "#E1F5FF"
    stroke: "#01579B"
  }
}
`, container.Name, container.Technology, container.Description, container.ID, container.Name, desc, tech, container.ID)

	d2Path := filepath.Join(container.Path, container.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}

// validateTemplate checks if the specified template exists.
// If not found, returns an error listing available templates.
func (nc *NewCommand) validateTemplate(templateName string) error {
	// Check in relative path first (for development)
	relPath := filepath.Join(".", "templates", templateName)
	if info, err := os.Stat(relPath); err == nil && info.IsDir() {
		return nil
	}

	// Check relative to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		binPath := filepath.Join(exeDir, "..", "templates", templateName)
		if info, err := os.Stat(binPath); err == nil && info.IsDir() {
			return nil
		}
	}

	// Template not found - list available templates
	available := nc.listAvailableTemplates()
	return fmt.Errorf("template %q not found. Available templates: %s", templateName, strings.Join(available, ", "))
}

// listAvailableTemplates returns a list of available template names.
func (nc *NewCommand) listAvailableTemplates() []string {
	var templates []string
	seen := make(map[string]bool)

	// Check relative path
	if entries, err := os.ReadDir(filepath.Join(".", "templates")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !seen[entry.Name()] {
				templates = append(templates, entry.Name())
				seen[entry.Name()] = true
			}
		}
	}

	// Check relative to executable
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

// createSystemD2 creates system D2 diagram using template if available.
func (nc *NewCommand) createSystemD2(ctx context.Context, system *entities.System, templateEngine *ason.TemplateEngine) error {
	d2Path := filepath.Join(system.Path, "system.d2")

	// Try template engine first
	if templateEngine != nil {
		variables := map[string]string{
			"SystemName":  system.Name,
			"SystemID":    system.ID,
			"Description": system.Description,
		}
		rendered, err := templateEngine.RenderTemplate(ctx, "system.d2", variables)
		if err == nil {
			return os.WriteFile(d2Path, []byte(rendered), 0644)
		}
		// Fall through to D2Generator on error
	}

	// Fallback to D2Generator
	d2Gen := NewD2Generator()
	return d2Gen.SaveSystemContextD2File(system)
}

// createContainerD2 creates container D2 diagram using template if available.
func (nc *NewCommand) createContainerD2(ctx context.Context, container *entities.Container, templateEngine *ason.TemplateEngine) error {
	d2Path := filepath.Join(container.Path, container.ID+".d2")

	// Try template engine first
	if templateEngine != nil {
		variables := map[string]string{
			"ContainerName": container.Name,
			"ContainerID":   container.ID,
			"Description":   container.Description,
			"Technology":    container.Technology,
		}
		rendered, err := templateEngine.RenderTemplate(ctx, "container.d2", variables)
		if err == nil {
			return os.WriteFile(d2Path, []byte(rendered), 0644)
		}
		// Fall through to hardcoded template on error
	}

	// Fallback to hardcoded template
	return nc.createContainerD2Template(container)
}

// createComponentD2 creates component D2 diagram using template if available.
func (nc *NewCommand) createComponentD2(ctx context.Context, component *entities.Component, templateEngine *ason.TemplateEngine) error {
	d2Path := filepath.Join(component.Path, component.ID+".d2")

	// Try template engine first
	if templateEngine != nil {
		variables := map[string]string{
			"ComponentName": component.Name,
			"ComponentID":   component.ID,
			"Description":   component.Description,
			"Technology":    component.Technology,
		}
		rendered, err := templateEngine.RenderTemplate(ctx, "component.d2", variables)
		if err == nil {
			return os.WriteFile(d2Path, []byte(rendered), 0644)
		}
		// Fall through to hardcoded template on error
	}

	// Fallback to hardcoded template
	return nc.createComponentD2Template(component)
}

// createComponentD2Template creates a component diagram template.
func (nc *NewCommand) createComponentD2Template(component *entities.Component) error {
	d2Content := fmt.Sprintf(`# %s Component Diagram
# C4 Level 3 - Component
# Architecture: %s
# Description: %s

direction: right

%s: "%s" {
  tooltip: "%s"
  style { fill: "#E3F2FD" }
}

# Dependencies (uncomment):
# cache: "Cache" { style { fill: "#FFE8D6" } }
# %s -> cache: "uses"

# Relationships (add):
# %s -> other_component: "relationship_description"
`, component.Name, component.Technology, component.Description, component.ID, component.Name, component.Description, component.ID, component.ID)

	d2Path := filepath.Join(component.Path, component.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}
