package entities

// Component represents a C4 component - the lowest level of the hierarchy.
// Components are code-level abstractions within a container.
type Component struct {
	// ID is the unique identifier (used in file paths)
	ID string

	// Name is the display name
	Name string

	// Description explains what this component does
	Description string

	// Technology describes the implementation (e.g., "Go package", "React component")
	Technology string

	// Tags for categorization and filtering
	Tags []string

	// Diagram is the optional component diagram
	Diagram *Diagram

	// Metadata holds additional frontmatter fields
	Metadata map[string]any

	// Path is the filesystem path to this component's directory
	Path string
}

// NewComponent creates a new component with the given name.
func NewComponent(name string) (*Component, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("Component", "Name", name, "invalid name", err)
	}

	return &Component{
		ID:       NormalizeName(name),
		Name:     name,
		Tags:     []string{},
		Metadata: make(map[string]any),
	}, nil
}

// Validate checks if the component is valid.
func (c *Component) Validate() error {
	var errs ValidationErrors

	if err := ValidateName(c.Name); err != nil {
		errs.Add("Component", "Name", c.Name, "invalid name", err)
	}

	if err := ValidateID(c.ID); err != nil {
		errs.Add("Component", "ID", c.ID, "invalid id", err)
	}

	if c.Diagram != nil {
		if err := c.Diagram.Validate(); err != nil {
			errs.Add("Component", "Diagram", c.ID, "invalid diagram", err)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SetDescription sets the component description.
func (c *Component) SetDescription(desc string) {
	c.Description = desc
}

// SetTechnology sets the technology stack.
func (c *Component) SetTechnology(tech string) {
	c.Technology = tech
}

// AddTag adds a tag to the component.
func (c *Component) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return // Already exists
		}
	}
	c.Tags = append(c.Tags, tag)
}

// HasTag checks if the component has a specific tag.
func (c *Component) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
