package unit

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestTemplateCategory_String tests the String() method for TemplateCategory
func TestTemplateCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category entities.TemplateCategory
		want     string
	}{
		{
			name:     "compute category",
			category: entities.TemplateCategoryCompute,
			want:     "compute",
		},
		{
			name:     "datastore category",
			category: entities.TemplateCategoryDatastore,
			want:     "datastore",
		},
		{
			name:     "messaging category",
			category: entities.TemplateCategoryMessaging,
			want:     "messaging",
		},
		{
			name:     "api category",
			category: entities.TemplateCategoryAPI,
			want:     "api",
		},
		{
			name:     "event category",
			category: entities.TemplateCategoryEvent,
			want:     "event",
		},
		{
			name:     "storage category",
			category: entities.TemplateCategoryStorage,
			want:     "storage",
		},
		{
			name:     "generic category",
			category: entities.TemplateCategoryGeneric,
			want:     "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.category)
			if got != tt.want {
				t.Errorf("TemplateCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTemplateSelector_SelectTemplateCategory tests pattern matching
func TestTemplateSelector_SelectTemplateCategory(t *testing.T) {
	tests := []struct {
		name         string
		technology   string
		wantCategory entities.TemplateCategory
		wantMatched  bool
	}{
		// Compute patterns
		{
			name:         "AWS Lambda",
			technology:   "AWS Lambda",
			wantCategory: entities.TemplateCategoryCompute,
			wantMatched:  true,
		},
		{
			name:         "lambda function",
			technology:   "lambda function",
			wantCategory: entities.TemplateCategoryCompute,
			wantMatched:  true,
		},
		{
			name:         "serverless function",
			technology:   "serverless function",
			wantCategory: entities.TemplateCategoryCompute,
			wantMatched:  true,
		},
		{
			name:         "ECS container",
			technology:   "ECS container",
			wantCategory: entities.TemplateCategoryCompute,
			wantMatched:  true,
		},
		// Datastore patterns
		{
			name:         "DynamoDB",
			technology:   "DynamoDB",
			wantCategory: entities.TemplateCategoryDatastore,
			wantMatched:  true,
		},
		{
			name:         "Amazon DynamoDB Table",
			technology:   "Amazon DynamoDB Table",
			wantCategory: entities.TemplateCategoryDatastore,
			wantMatched:  true,
		},
		{
			name:         "MySQL database",
			technology:   "MySQL database",
			wantCategory: entities.TemplateCategoryDatastore,
			wantMatched:  true,
		},
		{
			name:         "PostgreSQL",
			technology:   "PostgreSQL table",
			wantCategory: entities.TemplateCategoryDatastore,
			wantMatched:  true,
		},
		// Messaging patterns
		{
			name:         "SQS Queue",
			technology:   "SQS Queue",
			wantCategory: entities.TemplateCategoryMessaging,
			wantMatched:  true,
		},
		{
			name:         "Amazon SNS Topic",
			technology:   "Amazon SNS Topic",
			wantCategory: entities.TemplateCategoryMessaging,
			wantMatched:  true,
		},
		{
			name:         "Kafka queue",
			technology:   "Kafka queue",
			wantCategory: entities.TemplateCategoryMessaging,
			wantMatched:  true,
		},
		// API patterns
		{
			name:         "API Gateway",
			technology:   "API Gateway",
			wantCategory: entities.TemplateCategoryAPI,
			wantMatched:  true,
		},
		{
			name:         "REST endpoint",
			technology:   "REST endpoint",
			wantCategory: entities.TemplateCategoryAPI,
			wantMatched:  true,
		},
		{
			name:         "GraphQL API",
			technology:   "GraphQL API",
			wantCategory: entities.TemplateCategoryAPI,
			wantMatched:  true,
		},
		// Event patterns
		{
			name:         "EventBridge",
			technology:   "EventBridge",
			wantCategory: entities.TemplateCategoryEvent,
			wantMatched:  true,
		},
		{
			name:         "Event trigger",
			technology:   "Event trigger",
			wantCategory: entities.TemplateCategoryEvent,
			wantMatched:  true,
		},
		// Storage patterns
		{
			name:         "S3 Bucket",
			technology:   "S3 Bucket",
			wantCategory: entities.TemplateCategoryStorage,
			wantMatched:  true,
		},
		{
			name:         "file storage",
			technology:   "file storage",
			wantCategory: entities.TemplateCategoryStorage,
			wantMatched:  true,
		},
		// Unknown/Generic
		{
			name:         "unknown technology",
			technology:   "some unknown tech",
			wantCategory: entities.TemplateCategoryGeneric,
			wantMatched:  false,
		},
		{
			name:         "empty technology",
			technology:   "",
			wantCategory: entities.TemplateCategoryGeneric,
			wantMatched:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := entities.NewTemplateSelector()
			gotCategory, gotMatched := selector.SelectTemplateCategory(tt.technology)

			if gotCategory != tt.wantCategory {
				t.Errorf("SelectTemplateCategory() category = %v, want %v", gotCategory, tt.wantCategory)
			}
			if gotMatched != tt.wantMatched {
				t.Errorf("SelectTemplateCategory() matched = %v, want %v", gotMatched, tt.wantMatched)
			}
		})
	}
}

// TestTemplateSelector_CaseInsensitive tests that pattern matching is case-insensitive
func TestTemplateSelector_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		technology string
		want       entities.TemplateCategory
	}{
		{
			name:       "uppercase LAMBDA",
			technology: "AWS LAMBDA",
			want:       entities.TemplateCategoryCompute,
		},
		{
			name:       "lowercase lambda",
			technology: "aws lambda",
			want:       entities.TemplateCategoryCompute,
		},
		{
			name:       "mixed case Lambda",
			technology: "AWS Lambda Function",
			want:       entities.TemplateCategoryCompute,
		},
		{
			name:       "uppercase DYNAMODB",
			technology: "DYNAMODB TABLE",
			want:       entities.TemplateCategoryDatastore,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := entities.NewTemplateSelector()
			got, _ := selector.SelectTemplateCategory(tt.technology)
			if got != tt.want {
				t.Errorf("SelectTemplateCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTemplateSelector_AddPattern tests custom pattern addition
func TestTemplateSelector_AddPattern(t *testing.T) {
	selector := entities.NewTemplateSelector()

	// Add a custom pattern for a new technology
	customPattern := entities.TechnologyPattern{
		Keywords: []string{"custom-tech", "my-service"},
		Category: entities.TemplateCategoryCompute,
		Priority: 15, // Higher priority than defaults
	}
	selector.AddPattern(customPattern)

	// Test that the custom pattern is matched
	category, matched := selector.SelectTemplateCategory("Custom-Tech Service")
	if !matched {
		t.Errorf("Custom pattern not matched")
	}
	if category != entities.TemplateCategoryCompute {
		t.Errorf("SelectTemplateCategory() = %v, want %v", category, entities.TemplateCategoryCompute)
	}

	// Test that original patterns still work
	category2, matched2 := selector.SelectTemplateCategory("AWS Lambda")
	if !matched2 {
		t.Errorf("Original pattern no longer matched after adding custom")
	}
	if category2 != entities.TemplateCategoryCompute {
		t.Errorf("SelectTemplateCategory() = %v, want %v", category2, entities.TemplateCategoryCompute)
	}
}
