package d2

import (
	"context"
	"testing"
)

// TestD2Parser_ParseRelationships_SingleArrow verifies parsing a single relationship arrow.
// This is part of T028 unit tests for D2Parser.
func TestD2Parser_ParseRelationships_SingleArrow(t *testing.T) {
	d2Source := `
api-gateway -> lambda-function: Invokes
`

	parser := NewD2Parser()
	ctx := context.Background()

	relationships, err := parser.ParseRelationships(ctx, d2Source)
	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(relationships))
	}

	rel := relationships[0]
	if rel.Source != "api-gateway" {
		t.Errorf("Source = %q, want %q", rel.Source, "api-gateway")
	}
	if rel.Target != "lambda-function" {
		t.Errorf("Target = %q, want %q", rel.Target, "lambda-function")
	}
	if rel.Label != "Invokes" {
		t.Errorf("Label = %q, want %q", rel.Label, "Invokes")
	}
}

// TestD2Parser_ParseRelationships_MultipleArrows verifies parsing multiple relationships.
func TestD2Parser_ParseRelationships_MultipleArrows(t *testing.T) {
	d2Source := `
# Component diagram
order-service -> inventory-db: Reads stock levels
order-service -> payment-gateway: Processes payment
inventory-db -> s3-backup: Backup data
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(relationships))
	}

	// Verify first relationship
	if relationships[0].Source != "order-service" || relationships[0].Target != "inventory-db" {
		t.Errorf("First relationship: got %q -> %q, want order-service -> inventory-db",
			relationships[0].Source, relationships[0].Target)
	}

	// Verify labels are preserved
	if relationships[0].Label != "Reads stock levels" {
		t.Errorf("First relationship label = %q, want %q", relationships[0].Label, "Reads stock levels")
	}
}

// TestD2Parser_ParseRelationships_UnlabeledArrow verifies arrows without labels are parsed.
func TestD2Parser_ParseRelationships_UnlabeledArrow(t *testing.T) {
	d2Source := `
frontend -> backend
backend -> database
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(relationships))
	}

	// Unlabeled arrows should have empty label
	for i, rel := range relationships {
		if rel.Label != "" {
			t.Errorf("relationship[%d] Label = %q, want empty string for unlabeled arrow", i, rel.Label)
		}
	}

	// Verify source/target are correct
	if relationships[0].Source != "frontend" || relationships[0].Target != "backend" {
		t.Errorf("First relationship: got %q -> %q, want frontend -> backend",
			relationships[0].Source, relationships[0].Target)
	}
}

// TestD2Parser_ParseRelationships_EmptyFile verifies empty D2 source returns empty slice.
func TestD2Parser_ParseRelationships_EmptyFile(t *testing.T) {
	testCases := []struct {
		name     string
		d2Source string
	}{
		{"completely empty", ""},
		{"whitespace only", "   \n\n   "},
		{"comments only", "# Just a comment\n# Another comment"},
		{"shapes without relationships", "api-gateway\nlambda-function\ndatabase"},
	}

	parser := NewD2Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			relationships, err := parser.ParseRelationships(context.Background(), tc.d2Source)

			if err != nil {
				t.Errorf("ParseRelationships() should not error on %q, got: %v", tc.name, err)
			}

			if relationships == nil {
				t.Error("ParseRelationships() should return empty slice, not nil")
			}

			if len(relationships) != 0 {
				t.Errorf("expected 0 relationships for %q, got %d", tc.name, len(relationships))
			}
		})
	}
}

// TestD2Parser_ParseRelationships_InvalidSyntax verifies parse errors are returned.
func TestD2Parser_ParseRelationships_InvalidSyntax(t *testing.T) {
	testCases := []struct {
		name     string
		d2Source string
	}{
		{
			name:     "unclosed brace",
			d2Source: "api-gateway { \n  props: { ",
		},
		{
			name:     "malformed arrow",
			d2Source: "api-gateway --> -> lambda",
		},
		{
			name:     "incomplete arrow syntax",
			d2Source: "api-gateway ->",
		},
	}

	parser := NewD2Parser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseRelationships(context.Background(), tc.d2Source)

			// We expect an error for invalid syntax
			if err == nil {
				t.Errorf("ParseRelationships() should return error for %q, got nil", tc.name)
			}
		})
	}
}

// TestD2Parser_ParseRelationships_NestedShapes verifies nested shapes (containers) work.
func TestD2Parser_ParseRelationships_NestedShapes(t *testing.T) {
	d2Source := `
backend {
  api-server
  worker
}

database {
  primary
  replica
}

backend.api-server -> database.primary: Queries
backend.worker -> database.replica: Reads
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) < 2 {
		t.Fatalf("expected at least 2 relationships, got %d", len(relationships))
	}

	// Verify nested paths are preserved
	foundNested := false
	for _, rel := range relationships {
		// D2 uses dot notation for nested shapes
		if (rel.Source == "backend.api-server" || rel.Source == "api-server") &&
			(rel.Target == "database.primary" || rel.Target == "primary") {
			foundNested = true
			break
		}
	}

	if !foundNested {
		t.Error("expected to find nested shape relationship (backend.api-server -> database.primary)")
	}
}

// TestD2Parser_ParseRelationships_SpecialArrowTypes verifies different arrow styles.
func TestD2Parser_ParseRelationships_SpecialArrowTypes(t *testing.T) {
	d2Source := `
# Different D2 arrow types
client -> server: HTTP request
server <-> cache: Bidirectional
auth --> database: One-way
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	// Should parse all arrow types
	if len(relationships) < 2 {
		t.Errorf("expected at least 2 relationships from different arrow types, got %d", len(relationships))
	}

	// Verify at least some relationships were extracted
	foundClient := false
	for _, rel := range relationships {
		if rel.Source == "client" && rel.Target == "server" {
			foundClient = true
		}
	}

	if !foundClient {
		t.Error("expected to find client -> server relationship")
	}
}

// TestD2Parser_ParseRelationships_MultiTarget verifies arrows with multiple targets.
func TestD2Parser_ParseRelationships_MultiTarget(t *testing.T) {
	// Note: D2 doesn't natively support multi-target arrows like "A -> B, C"
	// but we test that multiple separate arrows are correctly parsed
	d2Source := `
load-balancer -> server-1: Route traffic
load-balancer -> server-2: Route traffic
load-balancer -> server-3: Route traffic
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) != 3 {
		t.Fatalf("expected 3 relationships (one per target), got %d", len(relationships))
	}

	// Verify all have same source
	for i, rel := range relationships {
		if rel.Source != "load-balancer" {
			t.Errorf("relationship[%d] Source = %q, want %q", i, rel.Source, "load-balancer")
		}
	}

	// Verify different targets
	targets := make(map[string]bool)
	for _, rel := range relationships {
		targets[rel.Target] = true
	}

	expectedTargets := []string{"server-1", "server-2", "server-3"}
	for _, expected := range expectedTargets {
		if !targets[expected] {
			t.Errorf("expected target %q not found in relationships", expected)
		}
	}
}

// TestD2Parser_ParseRelationships_PreservesWhitespace verifies label whitespace handling.
func TestD2Parser_ParseRelationships_PreservesWhitespace(t *testing.T) {
	d2Source := `
api -> db: Queries user data and session info
`

	parser := NewD2Parser()
	relationships, err := parser.ParseRelationships(context.Background(), d2Source)

	if err != nil {
		t.Fatalf("ParseRelationships() error = %v", err)
	}

	if len(relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(relationships))
	}

	// Label should preserve internal whitespace
	if relationships[0].Label != "Queries user data and session info" {
		t.Errorf("Label = %q, whitespace not preserved correctly", relationships[0].Label)
	}
}
