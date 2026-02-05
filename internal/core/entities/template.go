package entities

// TemplateType represents the type of template.
type TemplateType string

const (
	TemplateTypeSystem    TemplateType = "system"
	TemplateTypeContainer TemplateType = "container"
	TemplateTypeComponent TemplateType = "component"
	TemplateTypeProject   TemplateType = "project"
)

// TemplateVariable defines a variable that can be set when using a template.
type TemplateVariable struct {
	// Name is the variable name (used in templates)
	Name string

	// Type is the variable type (string, bool, int)
	Type string

	// Description explains what this variable is for
	Description string

	// Required indicates if the variable must be provided
	Required bool

	// Default is the default value if not provided
	Default string

	// Prompt is the interactive prompt text
	Prompt string
}

// Template represents a scaffolding template for generating C4 documentation.
type Template struct {
	// ID is the unique identifier
	ID string

	// Name is the display name
	Name string

	// Description explains what this template creates
	Description string

	// Type is the template type (system, container, component, project)
	Type TemplateType

	// Version is the template version
	Version string

	// Variables are the configurable variables
	Variables []TemplateVariable

	// Files are the template files to render
	Files []TemplateFile

	// Path is the filesystem path to this template
	Path string

	// Global indicates if this is a global template (~/.loko/templates/)
	Global bool
}

// TemplateFile represents a single file in a template.
type TemplateFile struct {
	// Source is the template source path (relative to template root)
	Source string

	// Target is the output path pattern (can include variables)
	Target string

	// Condition is an optional condition for including this file
	Condition string
}

// NewTemplate creates a new template with the given name and type.
func NewTemplate(name string, templateType TemplateType) (*Template, error) {
	if err := ValidateName(name); err != nil {
		return nil, NewValidationError("Template", "Name", name, "invalid name", err)
	}

	return &Template{
		ID:        NormalizeName(name),
		Name:      name,
		Type:      templateType,
		Variables: []TemplateVariable{},
		Files:     []TemplateFile{},
	}, nil
}

// Validate checks if the template is valid.
func (t *Template) Validate() error {
	var errs ValidationErrors

	if err := ValidateName(t.Name); err != nil {
		errs.Add("Template", "Name", t.Name, "invalid name", err)
	}

	if err := ValidateID(t.ID); err != nil {
		errs.Add("Template", "ID", t.ID, "invalid id", err)
	}

	if t.Type == "" {
		errs.Add("Template", "Type", "", "type is required", nil)
	}

	if len(t.Files) == 0 {
		errs.Add("Template", "Files", "", "at least one file is required", nil)
	}

	// Validate variables
	for _, v := range t.Variables {
		if v.Name == "" {
			errs.Add("Template", "Variables", "", "variable name is required", nil)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// AddVariable adds a variable to the template.
func (t *Template) AddVariable(v TemplateVariable) {
	t.Variables = append(t.Variables, v)
}

// AddFile adds a file to the template.
func (t *Template) AddFile(f TemplateFile) {
	t.Files = append(t.Files, f)
}

// GetVariable retrieves a variable by name.
func (t *Template) GetVariable(name string) (*TemplateVariable, bool) {
	for i := range t.Variables {
		if t.Variables[i].Name == name {
			return &t.Variables[i], true
		}
	}
	return nil, false
}

// RequiredVariables returns all required variables.
func (t *Template) RequiredVariables() []TemplateVariable {
	var result []TemplateVariable
	for _, v := range t.Variables {
		if v.Required {
			result = append(result, v)
		}
	}
	return result
}

// DefaultValues returns a map of default values for all variables.
func (t *Template) DefaultValues() map[string]string {
	result := make(map[string]string)
	for _, v := range t.Variables {
		if v.Default != "" {
			result[v.Name] = v.Default
		}
	}
	return result
}
