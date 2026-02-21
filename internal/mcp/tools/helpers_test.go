package tools

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// buildTestGraph creates a small ArchitectureGraph for helper tests.
func buildTestGraph(t *testing.T) *entities.ArchitectureGraph {
	t.Helper()

	graph := entities.NewArchitectureGraph()

	// Add a node with ID "api-lambda" (slug) so lookups can find it
	graph.AddNode(&entities.GraphNode{
		ID:    "api-lambda",
		Name:  "API Lambda",
		Type:  "container",
		Level: 2,
	})

	// Add a node reachable via ShortIDMap
	graph.AddNode(&entities.GraphNode{
		ID:    "payment-service/db-proxy",
		Name:  "DB Proxy",
		Type:  "component",
		Level: 3,
	})

	return graph
}

// ─────────────────────────────────────────────────────────────────────────────
// suggestSlugID tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSuggestSlugID_NilGraphReturnsEmpty(t *testing.T) {
	result := suggestSlugID("API Lambda", nil)
	if result != "" {
		t.Errorf("expected empty string for nil graph, got %q", result)
	}
}

func TestSuggestSlugID_ExactSlugAlreadyCorrect(t *testing.T) {
	graph := buildTestGraph(t)

	result := suggestSlugID("api-lambda", graph)
	if result != "api-lambda" {
		t.Errorf("expected 'api-lambda', got %q", result)
	}
}

func TestSuggestSlugID_DisplayNameNormalizesToSlug(t *testing.T) {
	graph := buildTestGraph(t)

	// "API Lambda" normalizes to "api-lambda" which exists in graph
	result := suggestSlugID("API Lambda", graph)
	if result != "api-lambda" {
		t.Errorf("expected 'api-lambda' from display name 'API Lambda', got %q", result)
	}
}

func TestSuggestSlugID_UnknownNameReturnsEmpty(t *testing.T) {
	graph := buildTestGraph(t)

	result := suggestSlugID("Completely Unknown Service XYZ", graph)
	if result != "" {
		t.Errorf("expected empty string for unrecognized name, got %q", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// notFoundError tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNotFoundError_WithSuggestion(t *testing.T) {
	err := notFoundError("container", "API Lambda", "api-lambda")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "container") {
		t.Errorf("error message should contain entity type 'container': %q", msg)
	}
	if !strings.Contains(msg, "API Lambda") {
		t.Errorf("error message should contain input 'API Lambda': %q", msg)
	}
	if !strings.Contains(msg, "did you mean") {
		t.Errorf("error message should contain 'did you mean': %q", msg)
	}
	if !strings.Contains(msg, "api-lambda") {
		t.Errorf("error message should contain suggestion 'api-lambda': %q", msg)
	}
}

func TestNotFoundError_WithoutSuggestion_FallbackToQueryArchitecture(t *testing.T) {
	err := notFoundError("component", "XYZ Unknown", "")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "component") {
		t.Errorf("error message should contain entity type: %q", msg)
	}
	if !strings.Contains(msg, "XYZ Unknown") {
		t.Errorf("error message should contain input: %q", msg)
	}
	if !strings.Contains(msg, "query_architecture") {
		t.Errorf("fallback error should mention 'query_architecture': %q", msg)
	}
	// Must NOT contain "did you mean" when no suggestion
	if strings.Contains(msg, "did you mean") {
		t.Errorf("fallback message should not contain 'did you mean': %q", msg)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// validateElementPath tests
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateElementPath_ValidSlug(t *testing.T) {
	tests := []string{
		"agwe/api-lambda",
		"payment-service",
		"agwe/sqs-queue",
		"my-system/my-container/my-component",
	}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			_, err := validateElementPath(path)
			if err != nil {
				t.Errorf("expected valid path %q to pass validation, got: %v", path, err)
			}
		})
	}
}

func TestValidateElementPath_InvalidSlugReturnsError(t *testing.T) {
	tests := []struct {
		input    string
		wantSlug string
	}{
		{"agwe/API Lambda", "agwe/api-lambda"},
		{"Payment Service", "payment-service"},
		{"agwe/SQS Queue", "agwe/sqs-queue"},
		{"My System/My Container", "my-system/my-container"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			suggestion, err := validateElementPath(tc.input)
			if err == nil {
				t.Errorf("expected validation error for %q, got nil", tc.input)
				return
			}
			if suggestion != tc.wantSlug {
				t.Errorf("expected suggestion %q, got %q", tc.wantSlug, suggestion)
			}
			if !strings.Contains(err.Error(), tc.wantSlug) {
				t.Errorf("error message should contain corrected slug %q: %v", tc.wantSlug, err)
			}
			if !strings.Contains(err.Error(), "did you mean") {
				t.Errorf("error message should contain 'did you mean': %v", err)
			}
		})
	}
}

func TestValidateElementPath_EmptyPathIsValid(t *testing.T) {
	_, err := validateElementPath("")
	if err != nil {
		t.Errorf("expected empty path to be valid (caller handles empty check), got: %v", err)
	}
}
