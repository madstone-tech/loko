package template

import (
	"github.com/madstone-tech/loko/internal/core/entities"
)

// Registry implements the TemplateRegistry interface for template name resolution.
type Registry struct {
	// templateMap maps categories and entity types to actual template names
	templateMap map[entities.TemplateCategory]map[string]string
}

// NewRegistry creates a new template registry with default mappings.
func NewRegistry() *Registry {
	return &Registry{
		templateMap: defaultTemplateMap(),
	}
}

// GetTemplateName maps a category and entity type to an actual template name.
func (r *Registry) GetTemplateName(category entities.TemplateCategory, entityType string) string {
	if categoryMap, exists := r.templateMap[category]; exists {
		if templateName, exists := categoryMap[entityType]; exists {
			return templateName
		}
	}

	// Fallback to generic templates
	if genericMap, exists := r.templateMap[entities.TemplateCategoryGeneric]; exists {
		if templateName, exists := genericMap[entityType]; exists {
			return templateName
		}
	}

	// Ultimate fallback
	return "standard-3layer"
}

// IsValidTemplate checks if a template name is valid.
// In a real implementation, this would check against available templates.
func (r *Registry) IsValidTemplate(name string) bool {
	// For now, we'll assume all non-empty template names are valid
	// In a real implementation, this would check the filesystem or template store
	return name != ""
}

// defaultTemplateMap returns the default template mappings.
func defaultTemplateMap() map[entities.TemplateCategory]map[string]string {
	return map[entities.TemplateCategory]map[string]string{
		entities.TemplateCategoryCompute: {
			"container": "compute-container",
			"component": "compute-component",
		},
		entities.TemplateCategoryDatastore: {
			"container": "datastore-container",
			"component": "datastore-component",
		},
		entities.TemplateCategoryMessaging: {
			"container": "messaging-container",
			"component": "messaging-component",
		},
		entities.TemplateCategoryAPI: {
			"container": "api-container",
			"component": "api-component",
		},
		entities.TemplateCategoryEvent: {
			"container": "event-container",
			"component": "event-component",
		},
		entities.TemplateCategoryStorage: {
			"container": "storage-container",
			"component": "storage-component",
		},
		entities.TemplateCategoryGeneric: {
			"system":    "standard-system",
			"container": "standard-container",
			"component": "standard-component",
		},
	}
}
