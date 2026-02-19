package unit

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestDriftIssue_SeverityAssignment verifies that NewDriftIssue correctly assigns
// severity levels based on drift type according to T019 requirements:
// - DriftDescriptionMismatch → DriftWarning (cosmetic)
// - DriftMissingComponent → DriftError (broken reference)
// - DriftOrphanedRelationship → DriftError (data integrity)
func TestDriftIssue_SeverityAssignment(t *testing.T) {
	tests := []struct {
		name             string
		componentID      string
		driftType        entities.DriftType
		message          string
		context          string
		expectedSeverity entities.DriftSeverity
	}{
		{
			name:             "description mismatch assigns WARNING severity",
			componentID:      "api-gateway",
			driftType:        entities.DriftDescriptionMismatch,
			message:          "D2 tooltip differs from frontmatter",
			context:          "D2: 'API Gateway', Frontmatter: 'API Gateway Service'",
			expectedSeverity: entities.DriftWarning,
		},
		{
			name:             "missing component assigns ERROR severity",
			componentID:      "user-service",
			driftType:        entities.DriftMissingComponent,
			message:          "D2 diagram references non-existent component",
			context:          "Referenced: 'deleted-service', Available: ['user-service', 'auth-service']",
			expectedSeverity: entities.DriftError,
		},
		{
			name:             "orphaned relationship assigns ERROR severity",
			componentID:      "payment-service",
			driftType:        entities.DriftOrphanedRelationship,
			message:          "Frontmatter references deleted component",
			context:          "Relationship: 'payment-service → legacy-db', Target 'legacy-db' not found",
			expectedSeverity: entities.DriftError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := entities.NewDriftIssue(tt.componentID, tt.driftType, tt.message, tt.context)

			if issue.Severity != tt.expectedSeverity {
				t.Errorf("NewDriftIssue(%q, %v) severity = %v, want %v",
					tt.componentID, tt.driftType, issue.Severity, tt.expectedSeverity)
			}

			// Verify all fields are set correctly
			if issue.ComponentID != tt.componentID {
				t.Errorf("ComponentID = %q, want %q", issue.ComponentID, tt.componentID)
			}
			if issue.Type != tt.driftType {
				t.Errorf("Type = %v, want %v", issue.Type, tt.driftType)
			}
			if issue.Message != tt.message {
				t.Errorf("Message = %q, want %q", issue.Message, tt.message)
			}
			if issue.Context != tt.context {
				t.Errorf("Context = %q, want %q", issue.Context, tt.context)
			}
		})
	}
}

// TestDriftIssue_FieldValidation ensures DriftIssue fields are properly populated
func TestDriftIssue_FieldValidation(t *testing.T) {
	issue := entities.NewDriftIssue(
		"test-component",
		entities.DriftDescriptionMismatch,
		"Test message",
		"Test context",
	)

	if issue == nil {
		t.Fatal("NewDriftIssue() returned nil")
	}

	if issue.ComponentID == "" {
		t.Error("ComponentID should not be empty")
	}

	if issue.Message == "" {
		t.Error("Message should not be empty")
	}
}

// TestDriftIssue_AllDriftTypes verifies all DriftType constants work correctly
func TestDriftIssue_AllDriftTypes(t *testing.T) {
	driftTypes := []struct {
		driftType        entities.DriftType
		expectedSeverity entities.DriftSeverity
	}{
		{entities.DriftDescriptionMismatch, entities.DriftWarning},
		{entities.DriftMissingComponent, entities.DriftError},
		{entities.DriftOrphanedRelationship, entities.DriftError},
	}

	for _, tt := range driftTypes {
		issue := entities.NewDriftIssue("test", tt.driftType, "msg", "ctx")
		if issue.Severity != tt.expectedSeverity {
			t.Errorf("DriftType %v: got severity %v, want %v",
				tt.driftType, issue.Severity, tt.expectedSeverity)
		}
	}
}
