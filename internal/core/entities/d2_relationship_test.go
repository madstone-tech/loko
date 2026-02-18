package entities

import (
	"testing"
)

// TestNewD2Relationship tests creating a new D2Relationship.
func TestNewD2Relationship(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		target   string
		label    string
		wantErr  bool
		validate func(t *testing.T, rel *D2Relationship)
	}{
		{
			name:    "valid relationship with label",
			source:  "auth-service",
			target:  "database",
			label:   "queries",
			wantErr: false,
			validate: func(t *testing.T, rel *D2Relationship) {
				if rel.Source != "auth-service" {
					t.Errorf("expected source 'auth-service', got %q", rel.Source)
				}
				if rel.Target != "database" {
					t.Errorf("expected target 'database', got %q", rel.Target)
				}
				if rel.Label != "queries" {
					t.Errorf("expected label 'queries', got %q", rel.Label)
				}
			},
		},
		{
			name:    "valid relationship without label",
			source:  "api-service",
			target:  "auth-service",
			label:   "",
			wantErr: false,
			validate: func(t *testing.T, rel *D2Relationship) {
				if rel.Source != "api-service" {
					t.Errorf("expected source 'api-service', got %q", rel.Source)
				}
				if rel.Target != "auth-service" {
					t.Errorf("expected target 'auth-service', got %q", rel.Target)
				}
				if rel.Label != "" {
					t.Errorf("expected empty label, got %q", rel.Label)
				}
			},
		},
		{
			name:    "empty source",
			source:  "",
			target:  "database",
			label:   "queries",
			wantErr: true,
		},
		{
			name:    "empty target",
			source:  "auth-service",
			target:  "",
			label:   "queries",
			wantErr: true,
		},
		{
			name:    "both empty",
			source:  "",
			target:  "",
			label:   "queries",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := NewD2Relationship(tt.source, tt.target, tt.label)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewD2Relationship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if rel == nil {
				t.Fatal("NewD2Relationship() returned nil")
			}

			if tt.validate != nil {
				tt.validate(t, rel)
			}
		})
	}
}

// TestD2Relationship_Key tests the Key method of D2Relationship.
func TestD2Relationship_Key(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		target   string
		label    string
		expected string
	}{
		{
			name:     "with label",
			source:   "auth",
			target:   "db",
			label:    "queries",
			expected: "auth->db:queries",
		},
		{
			name:     "without label",
			source:   "api",
			target:   "auth",
			label:    "",
			expected: "api->auth:",
		},
		{
			name:     "complex names",
			source:   "payment-service",
			target:   "transaction-db",
			label:    "writes transactions",
			expected: "payment-service->transaction-db:writes transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := NewD2Relationship(tt.source, tt.target, tt.label)
			if err != nil {
				t.Fatalf("NewD2Relationship() error = %v", err)
			}

			key := rel.Key()
			if key != tt.expected {
				t.Errorf("Key() = %q, want %q", key, tt.expected)
			}
		})
	}
}
