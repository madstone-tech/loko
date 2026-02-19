package entities

import (
	"strings"
)

// TemplateCategory represents a category of templates for different technology types.
type TemplateCategory string

const (
	TemplateCategoryCompute   TemplateCategory = "compute"
	TemplateCategoryDatastore TemplateCategory = "datastore"
	TemplateCategoryMessaging TemplateCategory = "messaging"
	TemplateCategoryAPI       TemplateCategory = "api"
	TemplateCategoryEvent     TemplateCategory = "event"
	TemplateCategoryStorage   TemplateCategory = "storage"
	TemplateCategoryGeneric   TemplateCategory = "generic"
)

// TechnologyPattern maps technology keywords to template categories.
type TechnologyPattern struct {
	// Keywords are the technology terms to match (case insensitive)
	Keywords []string

	// Category is the template category for this technology
	Category TemplateCategory

	// Priority determines matching precedence (higher = matched first)
	Priority int
}

// TemplateSelector selects appropriate templates based on technology fields.
type TemplateSelector struct {
	// patterns maps technology keywords to template categories
	patterns []TechnologyPattern
}

// NewTemplateSelector creates a new template selector with default patterns.
func NewTemplateSelector() *TemplateSelector {
	return &TemplateSelector{
		patterns: defaultTechnologyPatterns(),
	}
}

// SelectTemplateCategory determines the appropriate template category based on technology description.
// Returns the determined category and a boolean indicating if a match was found.
func (ts *TemplateSelector) SelectTemplateCategory(technology string) (TemplateCategory, bool) {
	if technology == "" {
		return TemplateCategoryGeneric, false
	}

	// Normalize the technology string for comparison
	normalizedTech := strings.ToLower(strings.TrimSpace(technology))

	// Try to match against patterns, prioritizing higher priority patterns
	for _, pattern := range ts.patterns {
		for _, keyword := range pattern.Keywords {
			if strings.Contains(normalizedTech, strings.ToLower(keyword)) {
				return pattern.Category, true
			}
		}
	}

	// No match found, return generic
	return TemplateCategoryGeneric, false
}

// AddPattern adds a new technology pattern to the selector.
func (ts *TemplateSelector) AddPattern(pattern TechnologyPattern) {
	ts.patterns = append(ts.patterns, pattern)
	// Sort by priority (highest first)
	for i := len(ts.patterns) - 1; i > 0; i-- {
		if ts.patterns[i].Priority > ts.patterns[i-1].Priority {
			ts.patterns[i], ts.patterns[i-1] = ts.patterns[i-1], ts.patterns[i]
		}
	}
}

// defaultTechnologyPatterns returns the default technology pattern mappings.
func defaultTechnologyPatterns() []TechnologyPattern {
	return []TechnologyPattern{
		{
			Keywords: []string{"lambda", "function", "serverless", "aws lambda"},
			Category: TemplateCategoryCompute,
			Priority: 10,
		},
		{
			Keywords: []string{"dynamodb", "database", "db", "sql", "nosql", "table"},
			Category: TemplateCategoryDatastore,
			Priority: 9,
		},
		{
			Keywords: []string{"sqs", "sns", "queue", "pubsub", "kafka", "rabbitmq"},
			Category: TemplateCategoryMessaging,
			Priority: 8,
		},
		{
			Keywords: []string{"api gateway", "rest", "graphql", "endpoint"},
			Category: TemplateCategoryAPI,
			Priority: 7,
		},
		{
			Keywords: []string{"eventbridge", "event", "trigger", "schedule"},
			Category: TemplateCategoryEvent,
			Priority: 6,
		},
		{
			Keywords: []string{"s3", "bucket", "storage", "file"},
			Category: TemplateCategoryStorage,
			Priority: 5,
		},
		{
			Keywords: []string{"ec2", "container", "docker", "ecs", "eks", "vm"},
			Category: TemplateCategoryCompute,
			Priority: 4,
		},
	}
}
