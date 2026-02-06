package entities

import "time"

// Project represents the root of a loko architecture documentation project.
// It corresponds to a loko.toml file and its directory structure.
type Project struct {
	// Name is the project name
	Name string

	// Description explains what this project documents
	Description string

	// Version is the documentation version
	Version string

	// Systems within this project
	Systems map[string]*System

	// Config holds the parsed loko.toml configuration
	Config *ProjectConfig

	// Path is the root filesystem path
	Path string

	// Metadata holds additional fields
	Metadata map[string]any

	// CreatedAt is when the project was created
	CreatedAt time.Time

	// UpdatedAt is when the project was last modified
	UpdatedAt time.Time
}

// ProjectConfig holds the loko.toml configuration values.
type ProjectConfig struct {
	// Paths configuration
	SourceDir string // Default: "./src"
	OutputDir string // Default: "./dist"

	// Template configuration
	Template string // Default: "standard-3layer"

	// D2 configuration
	D2Theme  string // Default: "neutral-default"
	D2Layout string // Default: "elk"
	D2Cache  bool   // Default: true

	// Output configuration
	HTMLEnabled     bool // Default: true
	MarkdownEnabled bool // Default: false
	PDFEnabled      bool // Default: false

	// Build configuration
	Parallel   bool // Default: true
	MaxWorkers int  // Default: 4

	// Server configuration
	ServePort int  // Default: 8080
	APIPort   int  // Default: 8081
	HotReload bool // Default: true
}

// DefaultProjectConfig returns the default configuration.
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		SourceDir:       "./src",
		OutputDir:       "./dist",
		Template:        "standard-3layer",
		D2Theme:         "neutral-default",
		D2Layout:        "elk",
		D2Cache:         true,
		HTMLEnabled:     true,
		MarkdownEnabled: false,
		PDFEnabled:      false,
		Parallel:        true,
		MaxWorkers:      4,
		ServePort:       8080,
		APIPort:         8081,
		HotReload:       true,
	}
}

// NewProject creates a new project with the given name.
func NewProject(name string) (*Project, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("Project", "Name", name, "invalid name", err)
	}

	now := time.Now()
	return &Project{
		Name:      name,
		Systems:   make(map[string]*System),
		Config:    DefaultProjectConfig(),
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate checks if the project is valid.
func (p *Project) Validate() error {
	var errs ValidationErrors

	if err := ValidateName(p.Name); err != nil {
		errs.Add("Project", "Name", p.Name, "invalid name", err)
	}

	// Validate all systems
	for _, sys := range p.Systems {
		if err := sys.Validate(); err != nil {
			errs.Add("Project", "System", sys.ID, "invalid system", err)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// AddSystem adds a system to this project.
func (p *Project) AddSystem(sys *System) error {
	if sys == nil {
		return NewValidationError("Project", "System", "", "system cannot be nil", nil)
	}

	if _, exists := p.Systems[sys.ID]; exists {
		return &DuplicateError{Entity: "System", ID: sys.ID, Parent: p.Name}
	}

	p.Systems[sys.ID] = sys
	p.UpdatedAt = time.Now()
	return nil
}

// GetSystem retrieves a system by ID.
func (p *Project) GetSystem(id string) (*System, error) {
	sys, exists := p.Systems[id]
	if !exists {
		return nil, &NotFoundError{Entity: "System", ID: id, Parent: p.Name}
	}
	return sys, nil
}

// RemoveSystem removes a system by ID.
func (p *Project) RemoveSystem(id string) error {
	if _, exists := p.Systems[id]; !exists {
		return &NotFoundError{Entity: "System", ID: id, Parent: p.Name}
	}
	delete(p.Systems, id)
	p.UpdatedAt = time.Now()
	return nil
}

// ListSystems returns all systems.
func (p *Project) ListSystems() []*System {
	result := make([]*System, 0, len(p.Systems))
	for _, sys := range p.Systems {
		result = append(result, sys)
	}
	return result
}

// SystemCount returns the number of systems.
func (p *Project) SystemCount() int {
	return len(p.Systems)
}

// ContainerCount returns the total number of containers across all systems.
func (p *Project) ContainerCount() int {
	count := 0
	for _, sys := range p.Systems {
		count += sys.ContainerCount()
	}
	return count
}

// ComponentCount returns the total number of components across all systems.
func (p *Project) ComponentCount() int {
	count := 0
	for _, sys := range p.Systems {
		count += sys.ComponentCount()
	}
	return count
}

// GetContainer retrieves a container by system ID and container ID.
func (p *Project) GetContainer(systemID, containerID string) (*Container, error) {
	sys, err := p.GetSystem(systemID)
	if err != nil {
		return nil, err
	}
	return sys.GetContainer(containerID)
}

// GetComponent retrieves a component by system, container, and component ID.
func (p *Project) GetComponent(systemID, containerID, componentID string) (*Component, error) {
	sys, err := p.GetSystem(systemID)
	if err != nil {
		return nil, err
	}
	return sys.GetComponent(containerID, componentID)
}

// SetDescription sets the project description.
func (p *Project) SetDescription(desc string) {
	p.Description = desc
	p.UpdatedAt = time.Now()
}

// SetVersion sets the project version.
func (p *Project) SetVersion(version string) {
	p.Version = version
	p.UpdatedAt = time.Now()
}

// Stats returns project statistics.
func (p *Project) Stats() ProjectStats {
	return ProjectStats{
		Systems:    p.SystemCount(),
		Containers: p.ContainerCount(),
		Components: p.ComponentCount(),
	}
}

// ProjectStats holds project statistics for reporting.
type ProjectStats struct {
	Systems    int
	Containers int
	Components int
}
