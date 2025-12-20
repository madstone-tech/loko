// Package filesystem provides file system implementations of the core ports.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ProjectRepository implements the ProjectRepository port using the file system.
// Projects are stored in a directory structure with loko.toml configuration
// and markdown files with YAML frontmatter.
type ProjectRepository struct {
	templateEngine usecases.TemplateEngine
}

// NewProjectRepository creates a new file system project repository.
func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		templateEngine: nil, // Can be set with SetTemplateEngine if needed
	}
}

// SetTemplateEngine allows setting a custom template engine for D2 rendering.
func (pr *ProjectRepository) SetTemplateEngine(te usecases.TemplateEngine) {
	pr.templateEngine = te
}

// LoadProject retrieves a project by its root directory path.
// Returns ErrProjectNotFound if the project doesn't exist.
func (pr *ProjectRepository) LoadProject(ctx context.Context, projectRoot string) (*entities.Project, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	// Check if project directory exists
	if _, err := os.Stat(projectRoot); err != nil {
		return nil, fmt.Errorf("project directory not found: %w", err)
	}

	// Load loko.toml configuration
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, projectName, err := loadConfigWithName(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// If project name not found in config, use directory name
	if projectName == "" {
		projectName = filepath.Base(projectRoot)
		// Normalize the project name (replace underscores with hyphens)
		projectName = entities.NormalizeName(projectName)
	}

	project, err := entities.NewProject(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	project.Path = projectRoot
	project.Config = config

	// Load systems from src directory
	srcDir := filepath.Join(projectRoot, config.SourceDir)
	if _, err := os.Stat(srcDir); err == nil {
		systems, err := pr.loadSystems(ctx, srcDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load systems: %w", err)
		}

		for _, sys := range systems {
			if err := project.AddSystem(sys); err != nil {
				return nil, fmt.Errorf("failed to add system: %w", err)
			}
		}
	}

	return project, nil
}

// SaveProject persists a project to disk.
// Creates directories and files as needed; returns error if write fails.
func (pr *ProjectRepository) SaveProject(ctx context.Context, project *entities.Project) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	if project.Path == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Create project root directory
	if err := os.MkdirAll(project.Path, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create src directory
	srcDir := filepath.Join(project.Path, project.Config.SourceDir)
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	// Save loko.toml with project metadata
	configPath := filepath.Join(project.Path, "loko.toml")
	if err := saveConfigWithProject(configPath, project); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// SaveSystem persists a system to disk.
func (pr *ProjectRepository) SaveSystem(ctx context.Context, projectRoot string, system *entities.System) error {
	if system == nil {
		return fmt.Errorf("system cannot be nil")
	}

	if projectRoot == "" {
		return fmt.Errorf("project root cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create system directory
	systemDir := filepath.Join(projectRoot, config.SourceDir, system.ID)
	if err := os.MkdirAll(systemDir, 0755); err != nil {
		return fmt.Errorf("failed to create system directory: %w", err)
	}

	system.Path = systemDir

	// Create system.md with YAML frontmatter
	systemMdPath := filepath.Join(systemDir, "system.md")
	content := pr.generateSystemMarkdown(system)
	if err := os.WriteFile(systemMdPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write system.md: %w", err)
	}

	return nil
}

// SaveContainer persists a container to disk.
func (pr *ProjectRepository) SaveContainer(ctx context.Context, projectRoot, systemName string, container *entities.Container) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	if projectRoot == "" {
		return fmt.Errorf("project root cannot be empty")
	}

	if systemName == "" {
		return fmt.Errorf("system name cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create container directory
	containerDir := filepath.Join(projectRoot, config.SourceDir, systemName, container.ID)
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return fmt.Errorf("failed to create container directory: %w", err)
	}

	container.Path = containerDir

	// Create container.md with YAML frontmatter
	containerMdPath := filepath.Join(containerDir, "container.md")
	content := pr.generateContainerMarkdown(container)
	if err := os.WriteFile(containerMdPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write container.md: %w", err)
	}

	return nil
}

// ListSystems returns all systems in a project.
func (pr *ProjectRepository) ListSystems(ctx context.Context, projectRoot string) ([]*entities.System, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	srcDir := filepath.Join(projectRoot, config.SourceDir)
	return pr.loadSystems(ctx, srcDir)
}

// LoadSystem retrieves a system by name within a project.
func (pr *ProjectRepository) LoadSystem(ctx context.Context, projectRoot, systemName string) (*entities.System, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	if systemName == "" {
		return nil, fmt.Errorf("system name cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	systemDir := filepath.Join(projectRoot, config.SourceDir, systemName)
	return pr.loadSystemFromDir(ctx, systemDir)
}

// LoadContainer retrieves a container by name within a system.
func (pr *ProjectRepository) LoadContainer(ctx context.Context, projectRoot, systemName, containerName string) (*entities.Container, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	if systemName == "" {
		return nil, fmt.Errorf("system name cannot be empty")
	}

	if containerName == "" {
		return nil, fmt.Errorf("container name cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	containerDir := filepath.Join(projectRoot, config.SourceDir, systemName, containerName)
	return pr.loadContainerFromDir(ctx, containerDir)
}

// SaveComponent persists a component to disk.
func (pr *ProjectRepository) SaveComponent(ctx context.Context, projectRoot, systemName, containerName string, component *entities.Component) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	if projectRoot == "" {
		return fmt.Errorf("project root cannot be empty")
	}

	if systemName == "" {
		return fmt.Errorf("system name cannot be empty")
	}

	if containerName == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create component directory
	componentDir := filepath.Join(projectRoot, config.SourceDir, systemName, containerName, component.ID)
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	component.Path = componentDir

	// Create component.md with YAML frontmatter
	componentMdPath := filepath.Join(componentDir, "component.md")
	content := pr.generateComponentMarkdown(component)
	if err := os.WriteFile(componentMdPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write component.md: %w", err)
	}

	// Create basic D2 diagram template (optional - if it doesn't exist)
	d2Path := filepath.Join(componentDir, component.ID+".d2")
	if _, err := os.Stat(d2Path); os.IsNotExist(err) {
		var d2Content string

		// Try to render from template if templateEngine is available
		if pr.templateEngine != nil {
			variables := map[string]string{
				"ComponentName": component.Name,
				"ComponentID":   component.ID,
				"Description":   component.Description,
				"Technology":    component.Technology,
			}
			rendered, err := pr.templateEngine.RenderTemplate(context.Background(), "component.d2", variables)
			if err == nil {
				d2Content = rendered
			} else {
				// Fall back to hardcoded template if ason template not found
				d2Content = pr.generateComponentD2Template(component)
			}
		} else {
			// Use hardcoded template if no template engine
			d2Content = pr.generateComponentD2Template(component)
		}

		_ = os.WriteFile(d2Path, []byte(d2Content), 0644) // Ignore errors - diagram is optional
	}

	return nil
}

// LoadComponent retrieves a component by name within a container.
func (pr *ProjectRepository) LoadComponent(ctx context.Context, projectRoot, systemName, containerName, componentName string) (*entities.Component, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	if systemName == "" {
		return nil, fmt.Errorf("system name cannot be empty")
	}

	if containerName == "" {
		return nil, fmt.Errorf("container name cannot be empty")
	}

	if componentName == "" {
		return nil, fmt.Errorf("component name cannot be empty")
	}

	// Load config to get source directory
	configPath := filepath.Join(projectRoot, "loko.toml")
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	componentDir := filepath.Join(projectRoot, config.SourceDir, systemName, containerName, componentName)
	return pr.loadComponentFromDir(ctx, componentDir)
}

// Helper functions

// loadSystems loads all systems from a source directory.
func (pr *ProjectRepository) loadSystems(ctx context.Context, srcDir string) ([]*entities.System, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}

	var systems []*entities.System
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			sys, err := pr.loadSystemFromDir(ctx, filepath.Join(srcDir, entry.Name()))
			if err != nil {
				// Log but continue loading other systems
				continue
			}
			systems = append(systems, sys)
		}
	}

	return systems, nil
}

// loadSystemFromDir loads a system from a directory.
func (pr *ProjectRepository) loadSystemFromDir(ctx context.Context, systemDir string) (*entities.System, error) {
	// Check if system.md exists
	systemMdPath := filepath.Join(systemDir, "system.md")
	if _, err := os.Stat(systemMdPath); err != nil {
		return nil, fmt.Errorf("system.md not found: %w", err)
	}

	// Read system.md
	content, err := os.ReadFile(systemMdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read system.md: %w", err)
	}

	// Parse frontmatter and create system
	name, description := pr.parseFrontmatter(string(content))
	if name == "" {
		name = filepath.Base(systemDir)
	}

	system, err := entities.NewSystem(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}

	system.Description = description
	system.Path = systemDir

	// Load system diagram if it exists
	system.Diagram = pr.loadDiagramFromDir(systemDir)

	// Load containers
	entries, err := os.ReadDir(systemDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				container, err := pr.loadContainerFromDir(ctx, filepath.Join(systemDir, entry.Name()))
				if err == nil {
					_ = system.AddContainer(container)
				}
			}
		}
	}

	return system, nil
}

// loadContainerFromDir loads a container from a directory.
func (pr *ProjectRepository) loadContainerFromDir(ctx context.Context, containerDir string) (*entities.Container, error) {
	// Check if container.md exists
	containerMdPath := filepath.Join(containerDir, "container.md")
	if _, err := os.Stat(containerMdPath); err != nil {
		return nil, fmt.Errorf("container.md not found: %w", err)
	}

	// Read container.md
	content, err := os.ReadFile(containerMdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read container.md: %w", err)
	}

	// Parse frontmatter and create container
	name, description := pr.parseFrontmatter(string(content))
	if name == "" {
		name = filepath.Base(containerDir)
	}

	container, err := entities.NewContainer(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	container.Description = description
	container.Path = containerDir

	// Load container diagram if it exists
	container.Diagram = pr.loadDiagramFromDir(containerDir)

	// Load components
	entries, err := os.ReadDir(containerDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				component, err := pr.loadComponentFromDir(ctx, filepath.Join(containerDir, entry.Name()))
				if err == nil {
					_ = container.AddComponent(component)
				}
			}
		}
	}

	return container, nil
}

// parseFrontmatter extracts name and description from YAML frontmatter.
// Format: ---\nname: ".."\ndescription: ".."\n---\n
func (pr *ProjectRepository) parseFrontmatter(content string) (name, description string) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", ""
	}

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "---" {
			break
		}

		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			name = strings.Trim(name, "\"'")
		}

		if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			description = strings.Trim(description, "\"'")
		}
	}

	return name, description
}

// loadDiagramFromDir loads a D2 diagram from a directory if it exists.
// Returns nil if no diagram file is found (diagram is optional).
func (pr *ProjectRepository) loadDiagramFromDir(dirPath string) *entities.Diagram {
	// Check for system.d2 or container.d2
	diagramPath := filepath.Join(dirPath, filepath.Base(dirPath)+".d2")

	// Try alternate naming: just ".d2" in the directory
	if _, err := os.Stat(diagramPath); err != nil {
		diagramPath = filepath.Join(dirPath, "system.d2")
		if _, err := os.Stat(diagramPath); err != nil {
			// No diagram file found
			return nil
		}
	}

	// Read the D2 file
	content, err := os.ReadFile(diagramPath)
	if err != nil {
		// If reading fails, just return nil (diagram is optional)
		return nil
	}

	// Create diagram entity
	diagram, err := entities.NewDiagram(diagramPath)
	if err != nil {
		return nil
	}

	// Set the source content
	diagram.SetSource(string(content))

	return diagram
}

// generateSystemMarkdown generates markdown content for a system using the template engine.
func (pr *ProjectRepository) generateSystemMarkdown(system *entities.System) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %q\n", system.Name))
	if system.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %q\n", system.Description))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", system.Name))
	if system.Description != "" {
		sb.WriteString(system.Description)
		sb.WriteString("\n\n")
	}

	// Add C4 context section
	sb.WriteString("## Context\n\n")
	sb.WriteString("This is a **C4 Level 1 - System Context Diagram** showing this system in the broader architecture.\n\n")

	sb.WriteString("## Containers\n\n")
	sb.WriteString("The system is composed of the following containers:\n\n")
	sb.WriteString("| Container | Description | Technology |\n")
	sb.WriteString("|-----------|-------------|------------|\n")
	sb.WriteString("| (Add your containers here) | | |\n\n")

	sb.WriteString("## System Responsibilities\n\n")
	sb.WriteString("This system is responsible for:\n\n")
	sb.WriteString("- (Add key responsibility 1)\n")
	sb.WriteString("- (Add key responsibility 2)\n")
	sb.WriteString("- (Add key responsibility 3)\n\n")

	sb.WriteString("## Dependencies\n\n")
	sb.WriteString("This system may depend on:\n\n")
	sb.WriteString("- (List external systems)\n\n")

	sb.WriteString("## Technology Stack\n\n")
	sb.WriteString("- Primary Language: (e.g., Go, Java, Python)\n")
	sb.WriteString("- Framework: (e.g., Spring Boot, FastAPI)\n")
	sb.WriteString("- Database: (e.g., PostgreSQL, MongoDB)\n\n")

	return sb.String()
}

// generateContainerMarkdown generates markdown content for a container.
func (pr *ProjectRepository) generateContainerMarkdown(container *entities.Container) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %q\n", container.Name))
	if container.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %q\n", container.Description))
	}
	if container.Technology != "" {
		sb.WriteString(fmt.Sprintf("technology: %q\n", container.Technology))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", container.Name))
	if container.Description != "" {
		sb.WriteString(container.Description)
		sb.WriteString("\n\n")
	}

	// Add C4 context
	sb.WriteString("## Context\n\n")
	sb.WriteString("This is a **C4 Level 2 - Container** representing a deployable unit within the system.\n\n")

	sb.WriteString("## Purpose\n\n")
	if container.Description != "" {
		sb.WriteString(fmt.Sprintf("This container is responsible for %s.\n\n", container.Description))
	} else {
		sb.WriteString("This container is responsible for (add purpose here).\n\n")
	}

	sb.WriteString("## Technology Stack\n\n")
	if container.Technology != "" {
		sb.WriteString(fmt.Sprintf("- **Primary**: %s\n", container.Technology))
	}
	sb.WriteString("- **Runtime**: (e.g., Docker, JVM, Node.js)\n")
	sb.WriteString("- **Database**: (e.g., PostgreSQL, Redis)\n\n")

	sb.WriteString("## Components\n\n")
	sb.WriteString("This container is composed of the following components:\n\n")
	sb.WriteString("| Component | Description | Technology |\n")
	sb.WriteString("|-----------|-------------|------------|\n")
	sb.WriteString("| (Add your components here) | | |\n\n")

	sb.WriteString("## Interfaces\n\n")
	sb.WriteString("### Inbound\n\n")
	sb.WriteString("- REST API endpoints\n")
	sb.WriteString("- gRPC services\n")
	sb.WriteString("- Message queue consumers\n\n")

	sb.WriteString("### Outbound\n\n")
	sb.WriteString("- Database connections\n")
	sb.WriteString("- External service calls\n")
	sb.WriteString("- Cache operations\n\n")

	sb.WriteString("## Deployment\n\n")
	sb.WriteString("- **Container Type**: (e.g., Docker, Pod)\n")
	sb.WriteString("- **Port**: (e.g., 8080)\n")
	sb.WriteString("- **Environment**: (e.g., dev, staging, prod)\n\n")

	sb.WriteString("## Monitoring\n\n")
	sb.WriteString("- Health checks: `/health`\n")
	sb.WriteString("- Metrics: Prometheus format\n")
	sb.WriteString("- Logs: Structured JSON\n\n")

	return sb.String()
}

// generateComponentD2Template generates a basic D2 diagram template for a component.
func (pr *ProjectRepository) generateComponentD2Template(component *entities.Component) string {
	return fmt.Sprintf(`# %s Component Diagram
# C4 Level 3 - Component
# Architecture: %s
# Description: %s

direction: right

%s: "%s" {
  tooltip: "%s"
}

# Dependencies (if any)
# Example:
# cache: "Cache Layer"
# 
# %s -> cache: uses

# Relationships (if any)
# Add component relationships here using the format:
# %s -> other_component: "relationship_type"
`, component.Name, component.Technology, component.Description,
		component.ID, component.Name, component.Description,
		component.ID, component.ID)
}

// generateComponentMarkdown generates markdown content for a component.
func (pr *ProjectRepository) generateComponentMarkdown(component *entities.Component) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", component.ID))
	sb.WriteString(fmt.Sprintf("name: %q\n", component.Name))
	if component.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %q\n", component.Description))
	}
	if component.Technology != "" {
		sb.WriteString(fmt.Sprintf("technology: %q\n", component.Technology))
	}
	if len(component.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range component.Tags {
			sb.WriteString(fmt.Sprintf("  - %q\n", tag))
		}
	}
	if len(component.Relationships) > 0 {
		sb.WriteString("relationships:\n")
		for targetID, desc := range component.Relationships {
			sb.WriteString(fmt.Sprintf("  %s: %q\n", targetID, desc))
		}
	}
	if len(component.CodeAnnotations) > 0 {
		sb.WriteString("code_annotations:\n")
		for path, desc := range component.CodeAnnotations {
			sb.WriteString(fmt.Sprintf("  %q: %q\n", path, desc))
		}
	}
	if len(component.Dependencies) > 0 {
		sb.WriteString("dependencies:\n")
		for _, dep := range component.Dependencies {
			sb.WriteString(fmt.Sprintf("  - %q\n", dep))
		}
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", component.Name))
	if component.Description != "" {
		sb.WriteString(component.Description)
		sb.WriteString("\n\n")
	}

	// Add C4 context
	sb.WriteString("## Context\n\n")
	sb.WriteString("This is a **C4 Level 3 - Component** representing code-level abstractions within a container.\n\n")

	sb.WriteString("## Responsibility\n\n")
	if component.Description != "" {
		sb.WriteString(fmt.Sprintf("This component is responsible for %s.\n\n", component.Description))
	} else {
		sb.WriteString("This component is responsible for (add responsibility here).\n\n")
	}

	sb.WriteString("## Technology\n\n")
	if component.Technology != "" {
		sb.WriteString(fmt.Sprintf("- **Language**: %s\n", component.Technology))
	}
	sb.WriteString("- **Framework**: (specify framework)\n")
	sb.WriteString("- **Pattern**: (e.g., MVC, CQRS, Event-Sourcing)\n\n")

	sb.WriteString("## Interfaces\n\n")
	sb.WriteString("### Public Methods\n\n")
	sb.WriteString("- `Method1()` - Description of method 1\n")
	sb.WriteString("- `Method2()` - Description of method 2\n\n")

	sb.WriteString("### Dependencies\n\n")
	if len(component.Dependencies) > 0 {
		for _, dep := range component.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	} else {
		sb.WriteString("- (List external dependencies like libraries, frameworks)\n")
	}
	sb.WriteString("\n")

	if len(component.Relationships) > 0 {
		sb.WriteString("### Component Relationships\n\n")
		sb.WriteString("This component depends on:\n\n")
		for targetID, desc := range component.Relationships {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", targetID, desc))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Implementation Details\n\n")
	sb.WriteString("### Key Classes/Functions\n\n")
	sb.WriteString("- `Class1` - Description\n")
	sb.WriteString("- `Class2` - Description\n\n")

	sb.WriteString("### Data Structures\n\n")
	sb.WriteString("- (List important data structures)\n\n")

	if len(component.CodeAnnotations) > 0 {
		sb.WriteString("### Code Locations\n\n")
		for path, desc := range component.CodeAnnotations {
			sb.WriteString(fmt.Sprintf("- `%s`: %s\n", path, desc))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Testing\n\n")
	sb.WriteString("- Unit tests: (specify framework)\n")
	sb.WriteString("- Integration tests: (specify framework)\n")
	sb.WriteString("- Coverage: (target %)\n\n")

	sb.WriteString("## Performance Considerations\n\n")
	sb.WriteString("- (Note any performance-critical aspects)\n")
	sb.WriteString("- (Document caching strategies)\n")
	sb.WriteString("- (List optimization opportunities)\n\n")

	return sb.String()
}

// loadComponentFromDir loads a component from a directory.
func (pr *ProjectRepository) loadComponentFromDir(ctx context.Context, componentDir string) (*entities.Component, error) {
	// Check if component.md exists
	componentMdPath := filepath.Join(componentDir, "component.md")
	if _, err := os.Stat(componentMdPath); err != nil {
		return nil, fmt.Errorf("component.md not found: %w", err)
	}

	// Read component.md
	content, err := os.ReadFile(componentMdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component.md: %w", err)
	}

	// Parse frontmatter and create component
	name, description, technology, tags, relationships, annotations, deps := pr.parseComponentFrontmatter(string(content))
	if name == "" {
		name = filepath.Base(componentDir)
	}

	component, err := entities.NewComponent(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	component.Description = description
	component.Technology = technology
	component.Tags = tags
	component.Relationships = relationships
	component.CodeAnnotations = annotations
	component.Dependencies = deps
	component.Path = componentDir

	// Load component diagram if it exists
	component.Diagram = pr.loadDiagramFromDir(componentDir)

	return component, nil
}

// parseComponentFrontmatter extracts metadata from YAML frontmatter for components.
func (pr *ProjectRepository) parseComponentFrontmatter(content string) (name, description, technology string, tags []string, relationships map[string]string, annotations map[string]string, dependencies []string) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", "", "", []string{}, make(map[string]string), make(map[string]string), []string{}
	}

	relationships = make(map[string]string)
	annotations = make(map[string]string)

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "---" {
			break
		}

		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			name = strings.Trim(name, "\"'")
		}

		if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			description = strings.Trim(description, "\"'")
		}

		if strings.HasPrefix(line, "technology:") {
			technology = strings.TrimSpace(strings.TrimPrefix(line, "technology:"))
			technology = strings.Trim(technology, "\"'")
		}

		if strings.HasPrefix(line, "tags:") {
			// Parse tags array from next lines
			for j := i + 1; j < len(lines); j++ {
				tagLine := lines[j]
				if strings.HasPrefix(tagLine, "  - ") {
					tag := strings.TrimSpace(strings.TrimPrefix(tagLine, "  - "))
					tag = strings.Trim(tag, "\"'")
					tags = append(tags, tag)
				} else if !strings.HasPrefix(tagLine, " ") || tagLine == "---" {
					// End of tags section
					i = j - 1
					break
				}
			}
		}

		if strings.HasPrefix(line, "relationships:") {
			// Parse relationships map from next lines
			for j := i + 1; j < len(lines); j++ {
				relLine := lines[j]
				if strings.HasPrefix(relLine, "  ") && strings.Contains(relLine, ":") {
					// Parse "  componentid: \"description\"" format
					relLine = strings.TrimPrefix(relLine, "  ")
					parts := strings.SplitN(relLine, ":", 2)
					if len(parts) == 2 {
						targetID := strings.TrimSpace(parts[0])
						desc := strings.TrimSpace(parts[1])
						desc = strings.Trim(desc, "\"'")
						relationships[targetID] = desc
					}
				} else if !strings.HasPrefix(relLine, " ") || relLine == "---" {
					// End of relationships section
					i = j - 1
					break
				}
			}
		}

		if strings.HasPrefix(line, "code_annotations:") {
			// Parse code annotations map from next lines
			for j := i + 1; j < len(lines); j++ {
				annLine := lines[j]
				if strings.HasPrefix(annLine, "  ") && strings.Contains(annLine, ":") {
					// Parse "  \"path\": \"description\"" format
					annLine = strings.TrimPrefix(annLine, "  ")
					parts := strings.SplitN(annLine, ":", 2)
					if len(parts) == 2 {
						path := strings.TrimSpace(parts[0])
						path = strings.Trim(path, "\"'")
						desc := strings.TrimSpace(parts[1])
						desc = strings.Trim(desc, "\"'")
						annotations[path] = desc
					}
				} else if !strings.HasPrefix(annLine, " ") || annLine == "---" {
					// End of annotations section
					i = j - 1
					break
				}
			}
		}

		if strings.HasPrefix(line, "dependencies:") {
			// Parse dependencies array from next lines
			for j := i + 1; j < len(lines); j++ {
				depLine := lines[j]
				if strings.HasPrefix(depLine, "  - ") {
					dep := strings.TrimSpace(strings.TrimPrefix(depLine, "  - "))
					dep = strings.Trim(dep, "\"'")
					dependencies = append(dependencies, dep)
				} else if !strings.HasPrefix(depLine, " ") || depLine == "---" {
					// End of dependencies section
					i = j - 1
					break
				}
			}
		}
	}

	return name, description, technology, tags, relationships, annotations, dependencies
}
