package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// SelectTemplateRequest defines the input for the SelectTemplate use case.
type SelectTemplateRequest struct {
	// Technology is the technology description to match against
	Technology string

	// EntityType is the type of entity being created ("system", "container", "component")
	EntityType string

	// OverrideTemplate is an optional explicit template name that bypasses selection
	OverrideTemplate string
}

// SelectTemplateResult defines the output of the SelectTemplate use case.
type SelectTemplateResult struct {
	// SelectedTemplate is the name of the template that was selected
	SelectedTemplate string

	// Category is the template category that was matched
	Category entities.TemplateCategory

	// Matched indicates whether a technology match was found
	Matched bool

	// Reason explains why this template was selected
	Reason string
}

// TemplateRegistry defines the interface for template name resolution.
type TemplateRegistry interface {
	// GetTemplateName maps a category to an actual template name
	GetTemplateName(category entities.TemplateCategory, entityType string) string

	// IsValidTemplate checks if a template name is valid
	IsValidTemplate(name string) bool
}

// SelectTemplate selects an appropriate template based on technology description.
type SelectTemplate struct {
	templateSelector *entities.TemplateSelector
	templateRegistry TemplateRegistry
}

// NewSelectTemplate creates a new SelectTemplate use case.
func NewSelectTemplate(registry TemplateRegistry) *SelectTemplate {
	return &SelectTemplate{
		templateSelector: entities.NewTemplateSelector(),
		templateRegistry: registry,
	}
}

// Execute selects a template based on technology description or uses an override.
func (uc *SelectTemplate) Execute(ctx context.Context, req *SelectTemplateRequest) (*SelectTemplateResult, error) {
	// If an override template is specified, use it directly
	if req.OverrideTemplate != "" {
		if !uc.templateRegistry.IsValidTemplate(req.OverrideTemplate) {
			return nil, fmt.Errorf("template %q not found", req.OverrideTemplate)
		}

		return &SelectTemplateResult{
			SelectedTemplate: req.OverrideTemplate,
			Category:         entities.TemplateCategoryGeneric,
			Matched:          false,
			Reason:           "explicitly specified by user",
		}, nil
	}

	// If no technology is specified, use default template
	if req.Technology == "" {
		defaultTemplate := uc.templateRegistry.GetTemplateName(entities.TemplateCategoryGeneric, req.EntityType)
		return &SelectTemplateResult{
			SelectedTemplate: defaultTemplate,
			Category:         entities.TemplateCategoryGeneric,
			Matched:          false,
			Reason:           "no technology specified, using default template",
		}, nil
	}

	// Select template based on technology
	category, matched := uc.templateSelector.SelectTemplateCategory(req.Technology)
	templateName := uc.templateRegistry.GetTemplateName(category, req.EntityType)

	// Validate that the selected template exists
	if !uc.templateRegistry.IsValidTemplate(templateName) {
		// Fallback to default if category-specific template doesn't exist
		templateName = uc.templateRegistry.GetTemplateName(entities.TemplateCategoryGeneric, req.EntityType)
		if !uc.templateRegistry.IsValidTemplate(templateName) {
			return nil, fmt.Errorf("no valid template found for category %q", category)
		}
	}

	reason := "using default template"
	if matched {
		reason = fmt.Sprintf("matched technology %q to category %q", req.Technology, category)
	}

	return &SelectTemplateResult{
		SelectedTemplate: templateName,
		Category:         category,
		Matched:          matched,
		Reason:           reason,
	}, nil
}
