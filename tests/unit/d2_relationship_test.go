package unit

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestD2Relationship_NewD2Relationship tests the D2Relationship constructor validation
func TestD2Relationship_NewD2Relationship(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		target    string
		label     string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid relationship with label",
			source:  "email-queue",
			target:  "email-sender",
			label:   "triggers",
			wantErr: false,
		},
		{
			name:    "valid relationship with empty label",
			source:  "email-queue",
			target:  "email-sender",
			label:   "",
			wantErr: false,
		},
		{
			name:      "empty source",
			source:    "",
			target:    "email-sender",
			label:     "triggers",
			wantErr:   true,
			errString: "source cannot be empty",
		},
		{
			name:      "empty target",
			source:    "email-queue",
			target:    "",
			label:     "triggers",
			wantErr:   true,
			errString: "target cannot be empty",
		},
		{
			name:      "both empty",
			source:    "",
			target:    "",
			label:     "triggers",
			wantErr:   true,
			errString: "source cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := entities.NewD2Relationship(tt.source, tt.target, tt.label)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewD2Relationship() expected error but got nil")
					return
				}
				if err.Error() != tt.errString {
					t.Errorf("NewD2Relationship() error = %v, want %v", err.Error(), tt.errString)
				}
				return
			}

			if err != nil {
				t.Errorf("NewD2Relationship() unexpected error = %v", err)
				return
			}

			if rel == nil {
				t.Errorf("NewD2Relationship() returned nil relationship")
				return
			}

			if rel.Source != tt.source {
				t.Errorf("NewD2Relationship() Source = %v, want %v", rel.Source, tt.source)
			}
			if rel.Target != tt.target {
				t.Errorf("NewD2Relationship() Target = %v, want %v", rel.Target, tt.target)
			}
			if rel.Label != tt.label {
				t.Errorf("NewD2Relationship() Label = %v, want %v", rel.Label, tt.label)
			}
		})
	}
}

// TestD2Relationship_Key tests the Key() method for deduplication
func TestD2Relationship_Key(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		target  string
		label   string
		wantKey string
	}{
		{
			name:    "relationship with label",
			source:  "email-queue",
			target:  "email-sender",
			label:   "triggers",
			wantKey: "email-queue->email-sender:triggers",
		},
		{
			name:    "relationship without label",
			source:  "email-queue",
			target:  "email-sender",
			label:   "",
			wantKey: "email-queue->email-sender:",
		},
		{
			name:    "different source same target",
			source:  "api",
			target:  "email-sender",
			label:   "triggers",
			wantKey: "api->email-sender:triggers",
		},
		{
			name:    "different label same source/target",
			source:  "email-queue",
			target:  "email-sender",
			label:   "uses",
			wantKey: "email-queue->email-sender:uses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := entities.NewD2Relationship(tt.source, tt.target, tt.label)
			if err != nil {
				t.Fatalf("NewD2Relationship() unexpected error = %v", err)
			}

			gotKey := rel.Key()
			if gotKey != tt.wantKey {
				t.Errorf("Key() = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}

// TestD2Relationship_KeyUniqueness tests that different relationships produce different keys
func TestD2Relationship_KeyUniqueness(t *testing.T) {
	rel1, _ := entities.NewD2Relationship("a", "b", "label1")
	rel2, _ := entities.NewD2Relationship("a", "b", "label2")
	rel3, _ := entities.NewD2Relationship("a", "c", "label1")
	rel4, _ := entities.NewD2Relationship("b", "a", "label1")

	keys := []string{rel1.Key(), rel2.Key(), rel3.Key(), rel4.Key()}

	// Check all keys are unique
	seen := make(map[string]bool)
	for i, key := range keys {
		if seen[key] {
			t.Errorf("Key %d (%s) is duplicate", i, key)
		}
		seen[key] = true
	}

	if len(seen) != 4 {
		t.Errorf("Expected 4 unique keys, got %d", len(seen))
	}
}
