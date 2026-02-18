package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// CreateComponentRequest holds the input for creating a component.
type CreateComponentRequest struct {
	// Name is the display name of the component (e.g., "Authentication Handler")
	Name string

	// Description explains what the component does
	Description string

	// Technology describes the implementation (e.g., "Go package", "React component")
	Technology string

	// Tags for categorization
	Tags []string

	// TemplateOverride is an optional explicit template name that bypasses
	// automatic selection (maps to --template CLI flag).
	TemplateOverride string
}

// CreateComponentResult holds the output of a CreateComponent execution.
type CreateComponentResult struct {
	// Component is the newly created component entity.
	Component *entities.Component

	// SelectedTemplate is the name of the template chosen for this component.
	// Empty string means the default template should be used.
	SelectedTemplate string

	// TemplateCategory is the technology category matched.
	TemplateCategory entities.TemplateCategory
}

// CreateComponent is the use case for creating a new component.
type CreateComponent struct {
	repo             ProjectRepository
	templateRegistry TemplateRegistry // Optional: if nil, no auto-selection
}

// NewCreateComponent creates a new CreateComponent use case without template selection.
func NewCreateComponent(repo ProjectRepository) *CreateComponent {
	return &CreateComponent{repo: repo}
}

// NewCreateComponentWithTemplates creates a use case with template auto-selection
// based on the Technology field.
func NewCreateComponentWithTemplates(repo ProjectRepository, registry TemplateRegistry) *CreateComponent {
	return &CreateComponent{repo: repo, templateRegistry: registry}
}

// Execute creates a new component with the given request.
// Returns the created component and selected template info, or an error.
func (uc *CreateComponent) Execute(ctx context.Context, req *CreateComponentRequest) (*CreateComponentResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create the component (validates name)
	component, err := entities.NewComponent(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	// Set basic fields
	component.Description = req.Description
	component.Technology = req.Technology
	component.Tags = req.Tags

	// Validate the complete component
	if err := component.Validate(); err != nil {
		return nil, fmt.Errorf("component validation failed: %w", err)
	}

	result := &CreateComponentResult{Component: component}

	// T055: Select template based on technology (if registry is configured)
	if uc.templateRegistry != nil {
		selectUC := NewSelectTemplate(uc.templateRegistry)
		selResult, err := selectUC.Execute(ctx, &SelectTemplateRequest{
			Technology:       req.Technology,
			EntityType:       "component",
			OverrideTemplate: req.TemplateOverride,
		})
		if err == nil {
			result.SelectedTemplate = selResult.SelectedTemplate
			result.TemplateCategory = selResult.Category
		}
		// Graceful degradation: if template selection fails, proceed without template
	}

	return result, nil
}
