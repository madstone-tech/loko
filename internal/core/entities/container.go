package entities

// Container represents a C4 container - a deployable unit within a system.
// Examples: API server, database, web app, mobile app.
type Container struct {
	// ID is the unique identifier (used in file paths)
	ID string

	// Name is the display name
	Name string

	// Description explains what this container does
	Description string

	// Technology describes the implementation (e.g., "Go + Fiber", "PostgreSQL 15")
	Technology string

	// Tags for categorization and filtering
	Tags []string

	// Components within this container
	Components map[string]*Component

	// Diagram is the container diagram
	Diagram *Diagram

	// DiagramPath is the relative path to the rendered diagram SVG file
	DiagramPath string

	// Metadata holds additional frontmatter fields
	Metadata map[string]any

	// Path is the filesystem path to this container's directory
	Path string

	// ParentID is the ID of the parent system
	ParentID string
}

// NewContainer creates a new container with the given name.
func NewContainer(name string) (*Container, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("Container", "Name", name, "invalid name", err)
	}

	return &Container{
		ID:         NormalizeName(name),
		Name:       name,
		Tags:       []string{},
		Components: make(map[string]*Component),
		Metadata:   make(map[string]any),
	}, nil
}

// Validate checks if the container is valid.
func (c *Container) Validate() error {
	var errs ValidationErrors

	if err := ValidateName(c.Name); err != nil {
		errs.Add("Container", "Name", c.Name, "invalid name", err)
	}

	if err := ValidateID(c.ID); err != nil {
		errs.Add("Container", "ID", c.ID, "invalid id", err)
	}

	if c.Diagram != nil {
		if err := c.Diagram.Validate(); err != nil {
			errs.Add("Container", "Diagram", c.ID, "invalid diagram", err)
		}
	}

	// Validate all components
	for _, comp := range c.Components {
		if err := comp.Validate(); err != nil {
			errs.Add("Container", "Component", comp.ID, "invalid component", err)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// AddComponent adds a component to this container.
func (c *Container) AddComponent(comp *Component) error {
	if comp == nil {
		return NewValidationError("Container", "Component", "", "component cannot be nil", nil)
	}

	if _, exists := c.Components[comp.ID]; exists {
		return &DuplicateError{Entity: "Component", ID: comp.ID, Parent: c.Name}
	}

	c.Components[comp.ID] = comp
	return nil
}

// GetComponent retrieves a component by ID.
func (c *Container) GetComponent(id string) (*Component, error) {
	comp, exists := c.Components[id]
	if !exists {
		return nil, &NotFoundError{Entity: "Component", ID: id, Parent: c.Name}
	}
	return comp, nil
}

// RemoveComponent removes a component by ID.
func (c *Container) RemoveComponent(id string) error {
	if _, exists := c.Components[id]; !exists {
		return &NotFoundError{Entity: "Component", ID: id, Parent: c.Name}
	}
	delete(c.Components, id)
	return nil
}

// ListComponents returns all components in order.
func (c *Container) ListComponents() []*Component {
	result := make([]*Component, 0, len(c.Components))
	for _, comp := range c.Components {
		result = append(result, comp)
	}
	return result
}

// ComponentCount returns the number of components.
func (c *Container) ComponentCount() int {
	return len(c.Components)
}

// SetDescription sets the container description.
func (c *Container) SetDescription(desc string) {
	c.Description = desc
}

// SetTechnology sets the technology stack.
func (c *Container) SetTechnology(tech string) {
	c.Technology = tech
}

// AddTag adds a tag to the container.
func (c *Container) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
}

// HasTag checks if the container has a specific tag.
func (c *Container) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
