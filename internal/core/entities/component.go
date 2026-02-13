package entities

// Component represents a C4 component - the lowest level of the hierarchy.
// Components are code-level abstractions within a container.
type Component struct {
	// ID is the unique identifier (used in file paths)
	ID string `json:"id" toon:"id"`

	// Name is the display name
	Name string `json:"name" toon:"name"`

	// Description explains what this component does
	Description string `json:"description" toon:"description,omitempty"`

	// Technology describes the implementation (e.g., "Go package", "React component")
	Technology string `json:"technology" toon:"technology,omitempty"`

	// Tags for categorization and filtering
	Tags []string `json:"tags" toon:"tags,omitempty"`

	// Relationships to other components (maps component ID to relationship description)
	Relationships map[string]string `json:"relationships" toon:"relationships,omitempty"`

	// CodeAnnotations maps code file/package paths to descriptions (e.g., "internal/auth" -> "JWT handling logic")
	CodeAnnotations map[string]string `json:"code_annotations" toon:"code_annotations,omitempty"`

	// Dependencies lists external packages/libraries this component depends on (e.g., "github.com/golang-jwt/jwt")
	Dependencies []string `json:"dependencies" toon:"dependencies,omitempty"`

	// Diagram is the optional component diagram
	Diagram *Diagram `json:"diagram,omitempty" toon:"diagram,omitempty"`

	// DiagramPath is the path to the rendered SVG diagram in the dist/ output
	DiagramPath string `json:"diagram_path" toon:"diagram_path,omitempty"`

	// Metadata holds additional frontmatter fields
	Metadata map[string]any `json:"metadata" toon:"metadata,omitempty"`

	// Path is the filesystem path to this component's directory
	Path string `json:"path" toon:"path,omitempty"`
}

// NewComponent creates a new component with the given name.
func NewComponent(name string) (*Component, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("Component", "Name", name, "invalid name", err)
	}

	return &Component{
		ID:              NormalizeName(name),
		Name:            name,
		Tags:            []string{},
		Relationships:   make(map[string]string),
		CodeAnnotations: make(map[string]string),
		Dependencies:    []string{},
		Metadata:        make(map[string]any),
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

// AddRelationship adds a relationship to another component.
func (c *Component) AddRelationship(targetComponentID, description string) {
	if targetComponentID != "" {
		c.Relationships[targetComponentID] = description
	}
}

// GetRelationship retrieves a relationship description by component ID.
func (c *Component) GetRelationship(targetComponentID string) (string, bool) {
	desc, exists := c.Relationships[targetComponentID]
	return desc, exists
}

// RemoveRelationship removes a relationship to another component.
func (c *Component) RemoveRelationship(targetComponentID string) {
	delete(c.Relationships, targetComponentID)
}

// ListRelationships returns all component relationships.
func (c *Component) ListRelationships() map[string]string {
	result := make(map[string]string)
	for id, desc := range c.Relationships {
		result[id] = desc
	}
	return result
}

// RelationshipCount returns the number of relationships.
func (c *Component) RelationshipCount() int {
	return len(c.Relationships)
}

// AddCodeAnnotation adds a code location annotation to the component.
func (c *Component) AddCodeAnnotation(codePath, description string) {
	if codePath != "" {
		c.CodeAnnotations[codePath] = description
	}
}

// GetCodeAnnotation retrieves a code annotation by path.
func (c *Component) GetCodeAnnotation(codePath string) (string, bool) {
	desc, exists := c.CodeAnnotations[codePath]
	return desc, exists
}

// RemoveCodeAnnotation removes a code annotation.
func (c *Component) RemoveCodeAnnotation(codePath string) {
	delete(c.CodeAnnotations, codePath)
}

// ListCodeAnnotations returns all code annotations.
func (c *Component) ListCodeAnnotations() map[string]string {
	result := make(map[string]string)
	for path, desc := range c.CodeAnnotations {
		result[path] = desc
	}
	return result
}

// CodeAnnotationCount returns the number of code annotations.
func (c *Component) CodeAnnotationCount() int {
	return len(c.CodeAnnotations)
}

// AddDependency adds an external dependency to the component.
func (c *Component) AddDependency(dep string) {
	if dep == "" {
		return
	}
	// Check if already exists
	for _, d := range c.Dependencies {
		if d == dep {
			return
		}
	}
	c.Dependencies = append(c.Dependencies, dep)
}

// RemoveDependency removes a dependency from the component.
func (c *Component) RemoveDependency(dep string) {
	for i, d := range c.Dependencies {
		if d == dep {
			c.Dependencies = append(c.Dependencies[:i], c.Dependencies[i+1:]...)
			return
		}
	}
}

// HasDependency checks if the component has a specific dependency.
func (c *Component) HasDependency(dep string) bool {
	for _, d := range c.Dependencies {
		if d == dep {
			return true
		}
	}
	return false
}

// ListDependencies returns all external dependencies.
func (c *Component) ListDependencies() []string {
	result := make([]string, len(c.Dependencies))
	copy(result, c.Dependencies)
	return result
}

// DependencyCount returns the number of external dependencies.
func (c *Component) DependencyCount() int {
	return len(c.Dependencies)
}

// GetID returns the component's unique identifier (implements C4Entity).
func (c *Component) GetID() string { return c.ID }

// GetName returns the component's display name (implements C4Entity).
func (c *Component) GetName() string { return c.Name }

// GetEntityType returns "component" (implements C4Entity).
func (c *Component) GetEntityType() string { return "component" }
