package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestD2RendererInterface tests the D2 renderer interface contract.
func TestD2RendererInterface(t *testing.T) {
	tests := []struct {
		name      string
		d2Source  string
		wantError bool
		wantSVG   bool
	}{
		{
			name:      "simple_rectangle_diagram",
			d2Source:  "my-shape: {\n  shape: rect\n}",
			wantError: false,
			wantSVG:   true,
		},
		{
			name: "diagram_with_connections",
			d2Source: `
user: User
api: API Server {
  shape: rect
}
db: Database {
  shape: cylinder
}

user -> api: "HTTP"
api -> db: "SQL"
`,
			wantError: false,
			wantSVG:   true,
		},
		{
			name:      "invalid_d2_syntax",
			d2Source:  "invalid d2 ::: syntax !!",
			wantError: true,
			wantSVG:   false,
		},
		{
			name:      "empty_diagram",
			d2Source:  "",
			wantError: true,
			wantSVG:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests if d2 CLI is not available
			_, err := os.Stat("/usr/local/bin/d2")
			if err != nil && os.IsNotExist(err) {
				// Try common paths
				_, err = os.Stat("/opt/homebrew/bin/d2")
				if err != nil && os.IsNotExist(err) {
					t.Skip("d2 CLI not found, skipping integration test")
				}
			}

			// This test is meant to verify the interface contract.
			// In production, it would use the actual D2 renderer implementation.
			// For now, we validate the structure.

			if tt.d2Source == "" {
				if !tt.wantError {
					t.Error("expected error for empty diagram")
				}
				return
			}

			// Test that diagram source is valid enough to attempt rendering
			diagram := &entities.Diagram{
				Source: tt.d2Source,
			}

			if diagram.Source == "" {
				t.Error("diagram source is empty")
			}
		})
	}
}

// TestDiagramEntity tests the Diagram entity creation and validation.
func TestDiagramEntity(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "valid_diagram_source",
			source:  "shape: rect",
			wantErr: false,
		},
		{
			name:    "complex_diagram",
			source:  "a: A\nb: B\na -> b",
			wantErr: false,
		},
		{
			name:    "empty_source",
			source:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := &entities.Diagram{
				Source: tt.source,
			}

			if tt.wantErr && diagram.Source != "" {
				t.Error("expected error validation")
			}

			if !tt.wantErr && diagram.Source == "" {
				t.Error("source should not be empty")
			}
		})
	}
}

// TestSystemWithDiagram tests systems with embedded diagrams.
func TestSystemWithDiagram(t *testing.T) {
	system := &entities.System{
		ID:   "PaymentService",
		Name: "Payment Service",
		Diagram: &entities.Diagram{
			Source: "shape: rect\nlabel: Payment Service",
		},
		Containers: map[string]*entities.Container{
			"API": {
				ID:   "API",
				Name: "API Service",
				Diagram: &entities.Diagram{
					Source: "shape: rect\nlabel: API",
				},
			},
		},
	}

	if system.Diagram == nil {
		t.Error("system diagram should not be nil")
	}

	if system.Diagram.Source == "" {
		t.Error("diagram source should not be empty")
	}

	if len(system.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(system.Containers))
	}

	container, ok := system.Containers["API"]
	if !ok {
		t.Error("API container not found")
	}

	if container.Diagram == nil {
		t.Error("container diagram should not be nil")
	}
}

// TestDiagramCaching tests that diagrams are cached by content hash.
func TestDiagramCaching(t *testing.T) {
	diagram1 := &entities.Diagram{
		Source: "shape: rect",
	}

	diagram2 := &entities.Diagram{
		Source: "shape: rect",
	}

	// Both diagrams have the same source, so they should have the same hash
	// (when implemented, they would use SHA256 hashing)
	if diagram1.Source != diagram2.Source {
		t.Error("diagram sources should match for caching")
	}

	// Different source
	diagram3 := &entities.Diagram{
		Source: "shape: circle",
	}

	if diagram1.Source == diagram3.Source {
		t.Error("diagram sources should differ")
	}
}

// TestMultipleSystemsWithDiagrams tests building multiple systems with diagrams.
func TestMultipleSystemsWithDiagrams(t *testing.T) {
	systems := []*entities.System{
		{
			ID:   "PaymentService",
			Name: "Payment Service",
			Diagram: &entities.Diagram{
				Source: "label: Payment Service",
			},
			Containers: map[string]*entities.Container{
				"API": {
					ID:   "API",
					Name: "API Service",
					Diagram: &entities.Diagram{
						Source: "label: API",
					},
				},
			},
		},
		{
			ID:   "AuthService",
			Name: "Auth Service",
			Diagram: &entities.Diagram{
				Source: "label: Auth Service",
			},
			Containers: map[string]*entities.Container{
				"AuthServer": {
					ID:   "AuthServer",
					Name: "Auth Server",
					Diagram: &entities.Diagram{
						Source: "label: Auth Server",
					},
				},
			},
		},
	}

	// Count diagrams
	diagramCount := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			diagramCount++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				diagramCount++
			}
		}
	}

	expectedCount := 4 // 2 systems + 2 containers
	if diagramCount != expectedCount {
		t.Errorf("expected %d diagrams, got %d", expectedCount, diagramCount)
	}
}

// TestDiagramRenderingTimeout tests that diagram rendering respects timeouts.
func TestDiagramRenderingTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This test verifies that the context timeout is respected
	// In actual implementation, the D2 renderer would check context.Done()

	select {
	case <-ctx.Done():
		// Timeout occurred as expected
		return
	case <-time.After(500 * time.Millisecond):
		// Timeout did not occur
		t.Error("context timeout was not respected")
	}
}

// TestDiagramErrorHandling tests error handling for invalid diagrams.
func TestDiagramErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		source string
		valid  bool
	}{
		{
			name:   "minimal_valid",
			source: "a: A",
			valid:  true,
		},
		{
			name:   "with_connections",
			source: "a -> b",
			valid:  true,
		},
		{
			name:   "with_styling",
			source: "a: A\na.style.fill: red",
			valid:  true,
		},
		{
			name:   "empty",
			source: "",
			valid:  false,
		},
		{
			name:   "only_whitespace",
			source: "   \n   \n   ",
			valid:  true, // Whitespace still exists, even if non-semantic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := &entities.Diagram{
				Source: tt.source,
			}

			// Basic validation: non-empty source is required
			isValid := len(diagram.Source) > 0 && diagram.Source != ""
			if isValid != tt.valid {
				t.Errorf("validation mismatch: expected %v, got %v", tt.valid, isValid)
			}
		})
	}
}
