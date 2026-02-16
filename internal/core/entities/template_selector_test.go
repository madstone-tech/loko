package entities

import (
	"testing"
)

func TestTemplateSelector_SelectTemplateCategory(t *testing.T) {
	selector := NewTemplateSelector()

	tests := []struct {
		name          string
		technology    string
		expectedCat   TemplateCategory
		expectedMatch bool
	}{
		{
			name:          "AWS Lambda function",
			technology:    "AWS Lambda function",
			expectedCat:   TemplateCategoryCompute,
			expectedMatch: true,
		},
		{
			name:          "DynamoDB table",
			technology:    "DynamoDB NoSQL table",
			expectedCat:   TemplateCategoryDatastore,
			expectedMatch: true,
		},
		{
			name:          "SQS queue",
			technology:    "Amazon SQS messaging queue",
			expectedCat:   TemplateCategoryMessaging,
			expectedMatch: true,
		},
		{
			name:          "API Gateway endpoint",
			technology:    "REST API Gateway endpoint",
			expectedCat:   TemplateCategoryAPI,
			expectedMatch: true,
		},
		{
			name:          "EventBridge rule",
			technology:    "EventBridge event trigger",
			expectedCat:   TemplateCategoryEvent,
			expectedMatch: true,
		},
		{
			name:          "S3 bucket",
			technology:    "S3 storage bucket",
			expectedCat:   TemplateCategoryStorage,
			expectedMatch: true,
		},
		{
			name:          "EC2 instance",
			technology:    "EC2 virtual machine",
			expectedCat:   TemplateCategoryCompute,
			expectedMatch: true,
		},
		{
			name:          "Unknown technology",
			technology:    "Some unknown technology",
			expectedCat:   TemplateCategoryGeneric,
			expectedMatch: false,
		},
		{
			name:          "Empty technology",
			technology:    "",
			expectedCat:   TemplateCategoryGeneric,
			expectedMatch: false,
		},
		{
			name:          "Case insensitive match",
			technology:    "LAMBDA FUNCTION",
			expectedCat:   TemplateCategoryCompute,
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, matched := selector.SelectTemplateCategory(tt.technology)

			if category != tt.expectedCat {
				t.Errorf("SelectTemplateCategory(%q) category = %v, want %v",
					tt.technology, category, tt.expectedCat)
			}

			if matched != tt.expectedMatch {
				t.Errorf("SelectTemplateCategory(%q) matched = %v, want %v",
					tt.technology, matched, tt.expectedMatch)
			}
		})
	}
}

func TestTemplateSelector_AddPattern(t *testing.T) {
	selector := NewTemplateSelector()

	// Add a custom pattern
	customPattern := TechnologyPattern{
		Keywords: []string{"custom-tech"},
		Category: TemplateCategoryGeneric,
		Priority: 100, // Higher priority than defaults
	}
	selector.AddPattern(customPattern)

	// Test that our custom pattern is matched first
	category, matched := selector.SelectTemplateCategory("custom-tech solution")
	if !matched {
		t.Error("Expected custom pattern to match")
	}

	if category != TemplateCategoryGeneric {
		t.Errorf("Expected category to be Generic, got %v", category)
	}
}

func TestTemplateSelector_PriorityOrdering(t *testing.T) {
	selector := NewTemplateSelector()

	// Add overlapping patterns to test priority
	lowPriority := TechnologyPattern{
		Keywords: []string{"lambda"},
		Category: TemplateCategoryGeneric,
		Priority: 1,
	}
	highPriority := TechnologyPattern{
		Keywords: []string{"lambda"},
		Category: TemplateCategoryDatastore,
		Priority: 20,
	}

	selector.AddPattern(lowPriority)
	selector.AddPattern(highPriority)

	// Should match high priority pattern
	category, matched := selector.SelectTemplateCategory("lambda function")
	if !matched {
		t.Error("Expected pattern to match")
	}

	if category != TemplateCategoryDatastore {
		t.Errorf("Expected category to be Datastore due to higher priority, got %v", category)
	}
}

func TestTemplateCategories(t *testing.T) {
	// Ensure all expected categories are defined
	categories := []TemplateCategory{
		TemplateCategoryCompute,
		TemplateCategoryDatastore,
		TemplateCategoryMessaging,
		TemplateCategoryAPI,
		TemplateCategoryEvent,
		TemplateCategoryStorage,
		TemplateCategoryGeneric,
	}

	expected := []string{"compute", "datastore", "messaging", "api", "event", "storage", "generic"}

	for i, category := range categories {
		if string(category) != expected[i] {
			t.Errorf("Category %d: expected %s, got %s", i, expected[i], string(category))
		}
	}
}
