package entities

import "slices"

// System represents a C4 system - a high-level abstraction.
// Examples: "Payment System", "Order Management System".
type System struct {
	// ID is the unique identifier (used in file paths)
	ID string

	// Name is the display name
	Name string

	// Description explains what this system does
	Description string

	// Tags for categorization and filtering
	Tags []string

	// Responsibilities lists key responsibilities of this system
	Responsibilities []string

	// Dependencies lists external systems or services this system depends on
	Dependencies []string

	// PrimaryLanguage is the main programming language used
	PrimaryLanguage string

	// Framework is the primary framework/library (e.g., Spring Boot, FastAPI, Cobra)
	Framework string

	// Database is the primary data storage technology
	Database string

	// KeyUsers lists the primary users or actors that use this system
	KeyUsers []string

	// ExternalSystems lists external systems this system integrates with
	ExternalSystems []string

	// Containers within this system
	Containers map[string]*Container

	// Diagram is the system context diagram
	Diagram *Diagram

	// DiagramPath is the relative path to the rendered diagram SVG file
	DiagramPath string

	// Metadata holds additional frontmatter fields
	Metadata map[string]any

	// Path is the filesystem path to this system's directory
	Path string

	// External indicates if this is an external system (not owned by us)
	External bool
}

// NewSystem creates a new system with the given name.
func NewSystem(name string) (*System, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("System", "Name", name, "invalid name", err)
	}

	return &System{
		ID:               NormalizeName(name),
		Name:             name,
		Tags:             []string{},
		Responsibilities: []string{},
		Dependencies:     []string{},
		KeyUsers:         []string{},
		ExternalSystems:  []string{},
		Containers:       make(map[string]*Container),
		Metadata:         make(map[string]any),
	}, nil
}

// Validate checks if the system is valid.
func (s *System) Validate() error {
	var errs ValidationErrors

	if err := ValidateName(s.Name); err != nil {
		errs.Add("System", "Name", s.Name, "invalid name", err)
	}

	if err := ValidateID(s.ID); err != nil {
		errs.Add("System", "ID", s.ID, "invalid id", err)
	}

	if s.Diagram != nil {
		if err := s.Diagram.Validate(); err != nil {
			errs.Add("System", "Diagram", s.ID, "invalid diagram", err)
		}
	}

	// Validate all containers
	for _, cont := range s.Containers {
		if err := cont.Validate(); err != nil {
			errs.Add("System", "Container", cont.ID, "invalid container", err)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// AddContainer adds a container to this system.
func (s *System) AddContainer(cont *Container) error {
	if cont == nil {
		return NewValidationError("System", "Container", "", "container cannot be nil", nil)
	}

	if _, exists := s.Containers[cont.ID]; exists {
		return &DuplicateError{Entity: "Container", ID: cont.ID, Parent: s.Name}
	}

	cont.ParentID = s.ID
	s.Containers[cont.ID] = cont
	return nil
}

// GetContainer retrieves a container by ID.
func (s *System) GetContainer(id string) (*Container, error) {
	cont, exists := s.Containers[id]
	if !exists {
		return nil, &NotFoundError{Entity: "Container", ID: id, Parent: s.Name}
	}
	return cont, nil
}

// RemoveContainer removes a container by ID.
func (s *System) RemoveContainer(id string) error {
	if _, exists := s.Containers[id]; !exists {
		return &NotFoundError{Entity: "Container", ID: id, Parent: s.Name}
	}
	delete(s.Containers, id)
	return nil
}

// ListContainers returns all containers in order.
func (s *System) ListContainers() []*Container {
	result := make([]*Container, 0, len(s.Containers))
	for _, cont := range s.Containers {
		result = append(result, cont)
	}
	return result
}

// ContainerCount returns the number of containers.
func (s *System) ContainerCount() int {
	return len(s.Containers)
}

// ComponentCount returns the total number of components across all containers.
func (s *System) ComponentCount() int {
	count := 0
	for _, cont := range s.Containers {
		count += cont.ComponentCount()
	}
	return count
}

// SetDescription sets the system description.
func (s *System) SetDescription(desc string) {
	s.Description = desc
}

// SetExternal marks the system as external.
func (s *System) SetExternal(external bool) {
	s.External = external
}

// AddTag adds a tag to the system.
func (s *System) AddTag(tag string) {
	if !slices.Contains(s.Tags, tag) {
		s.Tags = append(s.Tags, tag)
	}
}

// HasTag checks if the system has a specific tag.
func (s *System) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetComponent retrieves a component by container ID and component ID.
func (s *System) GetComponent(containerID, componentID string) (*Component, error) {
	cont, err := s.GetContainer(containerID)
	if err != nil {
		return nil, err
	}
	return cont.GetComponent(componentID)
}

// AddResponsibility adds a responsibility to the system (deduplicates).
func (s *System) AddResponsibility(resp string) {
	if resp == "" {
		return
	}
	if !slices.Contains(s.Responsibilities, resp) {
		s.Responsibilities = append(s.Responsibilities, resp)
	}
}

// AddDependency adds a dependency to the system (deduplicates).
func (s *System) AddDependency(dep string) {
	if dep == "" {
		return
	}
	if !slices.Contains(s.Dependencies, dep) {
		s.Dependencies = append(s.Dependencies, dep)
	}
}

// AddKeyUser adds a key user/actor to the system (deduplicates).
func (s *System) AddKeyUser(user string) {
	if user == "" {
		return
	}
	if !slices.Contains(s.KeyUsers, user) {
		s.KeyUsers = append(s.KeyUsers, user)
	}
}

// AddExternalSystem adds an external system this system integrates with (deduplicates).
func (s *System) AddExternalSystem(sys string) {
	if sys == "" {
		return
	}
	if !slices.Contains(s.ExternalSystems, sys) {
		s.ExternalSystems = append(s.ExternalSystems, sys)
	}
}
