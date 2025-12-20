package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/adapters/ason"
	"github.com/madstone-tech/loko/internal/adapters/filesystem"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// NewCommand creates new C4 entities (system, container, component).
type NewCommand struct {
	entityType  string // "system", "container", "component"
	entityName  string
	parentName  string // For container/component: parent system/container
	description string
	technology  string
	projectRoot string
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

	// Create template engine and add search paths
	templateEngine := ason.NewTemplateEngine()
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		templateEngine.AddSearchPath(filepath.Join(exeDir, "..", "templates", "standard-3layer"))
		templateEngine.AddSearchPath(filepath.Join(".", "templates", "standard-3layer"))
	}
	templateEngine.AddSearchPath("/Users/andhi/code/mdstn/loko/templates/standard-3layer")

	// Load project
	repo := filesystem.NewProjectRepository()
	repo.SetTemplateEngine(templateEngine)
	project, err := repo.LoadProject(ctx, nc.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	switch nc.entityType {
	case "system":
		return nc.createSystem(ctx, repo, project)
	case "container":
		return nc.createContainer(ctx, repo, project)
	case "component":
		return nc.createComponent(ctx, repo, project)
	default:
		return fmt.Errorf("unknown entity type: %s", nc.entityType)
	}
}

// createSystem creates a new system.
func (nc *NewCommand) createSystem(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project) error {
	// Use CreateSystem use case
	uc := usecases.NewCreateSystem(repo)
	system, err := uc.Execute(ctx, &usecases.CreateSystemRequest{
		Name:        nc.entityName,
		Description: nc.description,
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

	// Create default D2 diagram template
	if err := nc.createSystemD2Template(system); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create D2 template: %v\n", err)
	}

	return nil
}

// createContainer creates a new container in a system.
func (nc *NewCommand) createContainer(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project) error {
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

	// Create default D2 diagram template
	if err := nc.createContainerD2Template(container); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create D2 template: %v\n", err)
	}

	return nil
}

// createComponent creates a new component in a container.
func (nc *NewCommand) createComponent(ctx context.Context, repo *filesystem.ProjectRepository, project *entities.Project) error {
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

	// Create default D2 diagram template
	if err := nc.createComponentD2Template(component); err != nil {
		// Log warning but don't fail - D2 is optional
		fmt.Printf("⚠ Warning: Could not create D2 template: %v\n", err)
	}

	return nil
}

// createSystemD2Template creates a system context diagram template.
func (nc *NewCommand) createSystemD2Template(system *entities.System) error {
	d2Content := fmt.Sprintf(`# %s System Context Diagram
# C4 Level 1 - System Context
# Description: %s

direction: right

User: "User/Actor"

%s: "%s" {
  description: "%s"
}

User -> %s: "Uses"

# External systems (uncomment to add):
# ExternalSystem: "External System" {
#   icon: "https://icons.terrastruct.com/gcp/compute/Cloud%%20Run.svg"
# }
# %s -> ExternalSystem: "Integrates with"

%s: {
  style {
    fill: "#E1F5FF"
    stroke: "#01579B"
  }
}
`, system.Name, system.Description, system.ID, system.Name, system.Description, system.ID, system.ID, system.ID)

	d2Path := filepath.Join(system.Path, system.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
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
