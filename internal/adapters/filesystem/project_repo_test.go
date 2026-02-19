package filesystem

import (
	"testing"
)

// TestParseComponentFrontmatter_Relationships verifies that parseComponentFrontmatter
// correctly extracts the relationships map from component frontmatter YAML.
// This covers T020 requirement for frontmatter relationship parsing.
func TestParseComponentFrontmatter_Relationships(t *testing.T) {
	tests := []struct {
		name                     string
		frontmatter              string
		expectedRelationshipsLen int
		expectedTargets          map[string]string
	}{
		{
			name: "no relationships",
			frontmatter: `---
name: "API Gateway"
description: "REST API endpoint"
technology: "AWS API Gateway"
---
`,
			expectedRelationshipsLen: 0,
			expectedTargets:          map[string]string{},
		},
		{
			name: "single relationship",
			frontmatter: `---
name: "User Service"
description: "Handles user operations"
technology: "Go microservice"
relationships:
  user-db: "Reads/writes user data"
---
`,
			expectedRelationshipsLen: 1,
			expectedTargets: map[string]string{
				"user-db": "Reads/writes user data",
			},
		},
		{
			name: "multiple relationships",
			frontmatter: `---
name: "Payment Service"
description: "Payment processing"
technology: "Node.js"
relationships:
  payment-db: "Stores payment transactions"
  notification-queue: "Sends payment notifications"
  fraud-checker: "Validates transactions"
---
`,
			expectedRelationshipsLen: 3,
			expectedTargets: map[string]string{
				"payment-db":         "Stores payment transactions",
				"notification-queue": "Sends payment notifications",
				"fraud-checker":      "Validates transactions",
			},
		},
		{
			name: "relationship with quoted description",
			frontmatter: `---
name: "Auth Service"
relationships:
  user-db: "Queries user credentials and permissions"
  session-cache: 'Stores active sessions'
---
`,
			expectedRelationshipsLen: 2,
			expectedTargets: map[string]string{
				"user-db":       "Queries user credentials and permissions",
				"session-cache": "Stores active sessions",
			},
		},
		{
			name: "relationship IDs with hyphens and underscores",
			frontmatter: `---
name: "Data Pipeline"
relationships:
  s3-bucket: "Reads raw data files"
  dynamo_table: "Writes processed records"
  lambda-processor: "Triggers data transformation"
---
`,
			expectedRelationshipsLen: 3,
			expectedTargets: map[string]string{
				"s3-bucket":        "Reads raw data files",
				"dynamo_table":     "Writes processed records",
				"lambda-processor": "Triggers data transformation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := NewProjectRepository()

			_, _, _, _, relationships, _, _ := pr.parseComponentFrontmatter(tt.frontmatter)

			// Verify relationship count
			if len(relationships) != tt.expectedRelationshipsLen {
				t.Errorf("parseComponentFrontmatter() relationships count = %d, want %d",
					len(relationships), tt.expectedRelationshipsLen)
			}

			// Verify each expected target exists with correct description
			for targetID, expectedDesc := range tt.expectedTargets {
				actualDesc, exists := relationships[targetID]
				if !exists {
					t.Errorf("expected relationship to %q not found in parsed relationships", targetID)
					continue
				}
				if actualDesc != expectedDesc {
					t.Errorf("relationship %q description = %q, want %q",
						targetID, actualDesc, expectedDesc)
				}
			}

			// Ensure no unexpected relationships were parsed
			for targetID := range relationships {
				if _, expected := tt.expectedTargets[targetID]; !expected {
					t.Errorf("unexpected relationship to %q found in parsed relationships", targetID)
				}
			}
		})
	}
}

// TestParseComponentFrontmatter_EmptyRelationships ensures empty relationships
// section results in empty map (not nil).
func TestParseComponentFrontmatter_EmptyRelationships(t *testing.T) {
	frontmatter := `---
name: "Isolated Service"
relationships:
---
`
	pr := NewProjectRepository()
	_, _, _, _, relationships, _, _ := pr.parseComponentFrontmatter(frontmatter)

	if relationships == nil {
		t.Error("parseComponentFrontmatter() should return empty map, not nil")
	}

	if len(relationships) != 0 {
		t.Errorf("empty relationships section should result in zero entries, got %d", len(relationships))
	}
}

// TestParseComponentFrontmatter_RelationshipsWithOtherFields verifies relationships
// are correctly parsed alongside other frontmatter fields.
func TestParseComponentFrontmatter_RelationshipsWithOtherFields(t *testing.T) {
	frontmatter := `---
name: "Order Service"
description: "Manages customer orders"
technology: "Java Spring Boot"
tags:
  - microservice
  - core
relationships:
  inventory-service: "Checks product availability"
  payment-gateway: "Processes payments"
code_annotations:
  "src/orders": "Order domain logic"
dependencies:
  - "spring-boot-starter-web"
---
`
	pr := NewProjectRepository()
	name, description, technology, tags, relationships, annotations, dependencies := pr.parseComponentFrontmatter(frontmatter)

	// Verify all fields are parsed correctly
	if name != "Order Service" {
		t.Errorf("name = %q, want %q", name, "Order Service")
	}

	if description != "Manages customer orders" {
		t.Errorf("description = %q, want %q", description, "Manages customer orders")
	}

	if technology != "Java Spring Boot" {
		t.Errorf("technology = %q, want %q", technology, "Java Spring Boot")
	}

	if len(tags) != 2 {
		t.Errorf("tags count = %d, want 2", len(tags))
	}

	if len(relationships) != 2 {
		t.Errorf("relationships count = %d, want 2", len(relationships))
	}

	if relationships["inventory-service"] != "Checks product availability" {
		t.Errorf("relationship[inventory-service] = %q, want %q",
			relationships["inventory-service"], "Checks product availability")
	}

	if relationships["payment-gateway"] != "Processes payments" {
		t.Errorf("relationship[payment-gateway] = %q, want %q",
			relationships["payment-gateway"], "Processes payments")
	}

	if len(annotations) != 1 {
		t.Errorf("annotations count = %d, want 1", len(annotations))
	}

	if len(dependencies) != 1 {
		t.Errorf("dependencies count = %d, want 1", len(dependencies))
	}
}
