package entities

import (
	"strings"
	"testing"
)

// TestGenerateRelationshipID verifies deterministic ID generation.
func TestGenerateRelationshipID(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target string
		label  string
	}{
		{
			name:   "basic relationship",
			source: "agwe/api-lambda",
			target: "agwe/sqs-queue",
			label:  "Enqueue email job",
		},
		{
			name:   "container level",
			source: "backend/api",
			target: "backend/database",
			label:  "Reads data",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id1 := GenerateRelationshipID(tc.source, tc.target, tc.label)
			id2 := GenerateRelationshipID(tc.source, tc.target, tc.label)

			// Must be deterministic
			if id1 != id2 {
				t.Errorf("GenerateRelationshipID is not deterministic: %q != %q", id1, id2)
			}

			// Must be exactly 8 hex chars
			if len(id1) != 8 {
				t.Errorf("expected 8-char ID, got %d chars: %q", len(id1), id1)
			}

			// Must be lowercase hex only
			for _, c := range id1 {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("ID contains non-hex character %q in %q", string(c), id1)
				}
			}
		})
	}

	// Different inputs must produce different IDs
	t.Run("different inputs produce different IDs", func(t *testing.T) {
		id1 := GenerateRelationshipID("a/b", "c/d", "label")
		id2 := GenerateRelationshipID("a/b", "c/d", "different-label")
		id3 := GenerateRelationshipID("x/y", "c/d", "label")
		if id1 == id2 {
			t.Error("different labels should produce different IDs")
		}
		if id1 == id3 {
			t.Error("different sources should produce different IDs")
		}
	})
}

// TestNewRelationship verifies the constructor, defaults, and validation rules.
func TestNewRelationship(t *testing.T) {
	t.Run("happy path with defaults", func(t *testing.T) {
		rel, err := NewRelationship("agwe/api-lambda", "agwe/sqs-queue", "Enqueue email job")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rel.Source != "agwe/api-lambda" {
			t.Errorf("wrong Source: %q", rel.Source)
		}
		if rel.Target != "agwe/sqs-queue" {
			t.Errorf("wrong Target: %q", rel.Target)
		}
		if rel.Label != "Enqueue email job" {
			t.Errorf("wrong Label: %q", rel.Label)
		}
		if rel.Type != "sync" {
			t.Errorf("expected default Type 'sync', got %q", rel.Type)
		}
		if rel.Direction != "forward" {
			t.Errorf("expected default Direction 'forward', got %q", rel.Direction)
		}
		if rel.ID == "" {
			t.Error("ID must be generated")
		}
		// ID must match independent GenerateRelationshipID call
		expectedID := GenerateRelationshipID("agwe/api-lambda", "agwe/sqs-queue", "Enqueue email job")
		if rel.ID != expectedID {
			t.Errorf("ID mismatch: got %q, want %q", rel.ID, expectedID)
		}
	})

	t.Run("with all options", func(t *testing.T) {
		rel, err := NewRelationship(
			"backend/api", "backend/worker", "Dispatch job",
			WithRelType("async"),
			WithRelTechnology("AWS SQS"),
			WithRelDirection("bidirectional"),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rel.Type != "async" {
			t.Errorf("expected Type 'async', got %q", rel.Type)
		}
		if rel.Technology != "AWS SQS" {
			t.Errorf("expected Technology 'AWS SQS', got %q", rel.Technology)
		}
		if rel.Direction != "bidirectional" {
			t.Errorf("expected Direction 'bidirectional', got %q", rel.Direction)
		}
	})

	t.Run("event type is valid", func(t *testing.T) {
		_, err := NewRelationship("a/b", "c/d", "label", WithRelType("event"))
		if err != nil {
			t.Errorf("event type should be valid, got: %v", err)
		}
	})

	validationTests := []struct {
		name   string
		source string
		target string
		label  string
		opts   []RelationshipOption
		errMsg string
	}{
		{
			name:   "empty source",
			source: "",
			target: "a/b",
			label:  "label",
			errMsg: "source",
		},
		{
			name:   "blank source",
			source: "   ",
			target: "a/b",
			label:  "label",
			errMsg: "source",
		},
		{
			name:   "empty target",
			source: "a/b",
			target: "",
			label:  "label",
			errMsg: "target",
		},
		{
			name:   "source equals target (self-reference)",
			source: "a/b",
			target: "a/b",
			label:  "label",
			errMsg: "self-reference",
		},
		{
			name:   "empty label",
			source: "a/b",
			target: "c/d",
			label:  "",
			errMsg: "label",
		},
		{
			name:   "blank label",
			source: "a/b",
			target: "c/d",
			label:  "  ",
			errMsg: "label",
		},
		{
			name:   "invalid type",
			source: "a/b",
			target: "c/d",
			label:  "label",
			opts:   []RelationshipOption{WithRelType("grpc")},
			errMsg: "type",
		},
		{
			name:   "invalid direction",
			source: "a/b",
			target: "c/d",
			label:  "label",
			opts:   []RelationshipOption{WithRelDirection("reverse")},
			errMsg: "direction",
		},
	}

	for _, tc := range validationTests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewRelationship(tc.source, tc.target, tc.label, tc.opts...)
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("expected error to contain %q, got: %v", tc.errMsg, err)
			}
		})
	}
}

// TestRelationshipToD2Edge verifies D2 edge generation for all type/direction combinations.
func TestRelationshipToD2Edge(t *testing.T) {
	tests := []struct {
		name      string
		rel       Relationship
		wantParts []string // all strings must appear in output
	}{
		{
			name: "sync forward (default)",
			rel: Relationship{
				Source:    "backend/api-lambda",
				Target:    "backend/sqs-queue",
				Label:     "Enqueue job",
				Type:      "sync",
				Direction: "forward",
			},
			wantParts: []string{"api-lambda", "->", "sqs-queue", `"Enqueue job"`},
		},
		{
			name: "async forward — animated",
			rel: Relationship{
				Source:    "system/api",
				Target:    "system/worker",
				Label:     "Dispatch",
				Type:      "async",
				Direction: "forward",
			},
			wantParts: []string{"api", "->", "worker", "style.animated: true", `"Dispatch"`},
		},
		{
			name: "event forward — stroke-dash",
			rel: Relationship{
				Source:    "sys/publisher",
				Target:    "sys/subscriber",
				Label:     "Domain event",
				Type:      "event",
				Direction: "forward",
			},
			wantParts: []string{"publisher", "->", "subscriber", "style.stroke-dash: 5", `"Domain event"`},
		},
		{
			name: "sync bidirectional — double arrow",
			rel: Relationship{
				Source:    "sys/a",
				Target:    "sys/b",
				Label:     "Mutual sync",
				Type:      "sync",
				Direction: "bidirectional",
			},
			wantParts: []string{"a", "<->", "b", `"Mutual sync"`},
		},
		{
			name: "async bidirectional",
			rel: Relationship{
				Source:    "sys/a",
				Target:    "sys/b",
				Label:     "Async both ways",
				Type:      "async",
				Direction: "bidirectional",
			},
			wantParts: []string{"a", "<->", "b", "style.animated: true"},
		},
		{
			name: "short path segments extracted",
			rel: Relationship{
				Source:    "system-id/container-id/component-id",
				Target:    "system-id/container-id/another-component",
				Label:     "Uses",
				Type:      "sync",
				Direction: "forward",
			},
			wantParts: []string{"component-id", "->", "another-component"},
		},
		{
			name: "no slash in path — used as-is",
			rel: Relationship{
				Source:    "api",
				Target:    "database",
				Label:     "Reads",
				Type:      "sync",
				Direction: "forward",
			},
			wantParts: []string{"api", "->", "database"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RelationshipToD2Edge(tc.rel)
			for _, part := range tc.wantParts {
				if !strings.Contains(result, part) {
					t.Errorf("expected %q to contain %q, full output: %q", result, part, result)
				}
			}
			// Must end with newline
			if !strings.HasSuffix(result, "\n") {
				t.Errorf("D2 edge must end with newline, got: %q", result)
			}
		})
	}
}

// TestRelationshipsFile verifies the TOML wrapper struct has correct fields.
func TestRelationshipsFile(t *testing.T) {
	rf := RelationshipsFile{
		Relationships: []Relationship{
			{ID: "a1b2c3d4", Source: "sys/a", Target: "sys/b", Label: "uses"},
		},
	}
	if len(rf.Relationships) != 1 {
		t.Errorf("expected 1 relationship, got %d", len(rf.Relationships))
	}
	if rf.Relationships[0].ID != "a1b2c3d4" {
		t.Errorf("unexpected ID: %q", rf.Relationships[0].ID)
	}
}
