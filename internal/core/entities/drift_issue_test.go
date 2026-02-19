package entities

import (
	"testing"
)

// TestNewDriftIssue tests creating a new DriftIssue.
func TestNewDriftIssue(t *testing.T) {
	tests := []struct {
		name             string
		componentID      string
		driftType        DriftType
		message          string
		context          string
		expectedSeverity DriftSeverity
	}{
		{
			name:             "description mismatch - warning severity",
			componentID:      "auth-service",
			driftType:        DriftDescriptionMismatch,
			message:          "Description mismatch between D2 and frontmatter",
			context:          "D2: 'Auth service'\nFrontmatter: 'Authentication service'",
			expectedSeverity: DriftWarning,
		},
		{
			name:             "missing component - error severity",
			componentID:      "payment-service",
			driftType:        DriftMissingComponent,
			message:          "D2 references non-existent component",
			context:          "Component 'legacy-auth' referenced in D2 but not found",
			expectedSeverity: DriftError,
		},
		{
			name:             "orphaned relationship - error severity",
			componentID:      "api-service",
			driftType:        DriftOrphanedRelationship,
			message:          "Frontmatter relationship to deleted component",
			context:          "Relationship to 'deprecated-db' but component no longer exists",
			expectedSeverity: DriftError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := NewDriftIssue(tt.componentID, tt.driftType, tt.message, tt.context)

			if issue == nil {
				t.Fatal("NewDriftIssue() returned nil")
			}

			if issue.ComponentID != tt.componentID {
				t.Errorf("expected component ID %q, got %q", tt.componentID, issue.ComponentID)
			}

			if issue.Type != tt.driftType {
				t.Errorf("expected drift type %v, got %v", tt.driftType, issue.Type)
			}

			if issue.Severity != tt.expectedSeverity {
				t.Errorf("expected severity %v, got %v", tt.expectedSeverity, issue.Severity)
			}

			if issue.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, issue.Message)
			}

			if issue.Context != tt.context {
				t.Errorf("expected context %q, got %q", tt.context, issue.Context)
			}
		})
	}
}

// TestDriftSeverityConstants tests that drift severity constants are defined correctly.
func TestDriftSeverityConstants(t *testing.T) {
	// Test that constants have expected values
	if DriftWarning != 0 {
		t.Errorf("DriftWarning should be 0, got %d", DriftWarning)
	}

	if DriftError != 1 {
		t.Errorf("DriftError should be 1, got %d", DriftError)
	}
}

// TestDriftTypeConstants tests that drift type constants are defined correctly.
func TestDriftTypeConstants(t *testing.T) {
	// Test that constants have expected values
	if DriftDescriptionMismatch != 0 {
		t.Errorf("DriftDescriptionMismatch should be 0, got %d", DriftDescriptionMismatch)
	}

	if DriftMissingComponent != 1 {
		t.Errorf("DriftMissingComponent should be 1, got %d", DriftMissingComponent)
	}

	if DriftOrphanedRelationship != 2 {
		t.Errorf("DriftOrphanedRelationship should be 2, got %d", DriftOrphanedRelationship)
	}
}
