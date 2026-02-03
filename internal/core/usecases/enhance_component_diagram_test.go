package usecases

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestEnhanceComponentDiagramWithRelationships(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	// Create test system and container
	system, _ := entities.NewSystem("User Service")
	container, _ := entities.NewContainer("API Server")
	container.SetDescription("REST API server")

	// Create components
	auth, _ := entities.NewComponent("Authentication")
	auth.SetDescription("Handles user authentication")

	authCache, _ := entities.NewComponent("Auth Cache")
	authCache.SetDescription("Caches authentication tokens")

	userDB, _ := entities.NewComponent("User Database")
	userDB.SetDescription("Stores user data")

	// Set up relationships
	auth.AddRelationship("auth-cache", "stores sessions in")
	auth.AddRelationship("user-database", "queries user data from")

	// Add to container
	container.AddComponent(auth)
	container.AddComponent(authCache)
	container.AddComponent(userDB)

	// Create base diagram
	auth.Diagram = &entities.Diagram{
		Source: "authentication {\n  description: \"Authentication service\"\n}\n",
	}

	// Execute enhancement
	enhanced, err := uc.Execute(auth, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify enhancements were added
	if !strings.Contains(enhanced, "authentication -> auth-cache") {
		t.Error("Enhanced diagram missing relationship edge to auth-cache")
	}

	if !strings.Contains(enhanced, "authentication -> user-database") {
		t.Error("Enhanced diagram missing relationship edge to user-database")
	}

	// Verify descriptions were included
	if !strings.Contains(enhanced, "stores sessions in") {
		t.Error("Enhanced diagram missing relationship description for auth-cache")
	}
}

func TestEnhanceComponentDiagramFiltersExternalRelationships(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	// Create test system and container
	system, _ := entities.NewSystem("Service A")
	container, _ := entities.NewContainer("API")

	// Create components
	comp1, _ := entities.NewComponent("Component A")
	comp2, _ := entities.NewComponent("Component B")

	// Add relationship to external component (not in container)
	comp1.AddRelationship("external-component", "calls external service")
	comp1.AddRelationship("component-b", "internal dependency")

	container.AddComponent(comp1)
	container.AddComponent(comp2)

	// Create diagram
	comp1.Diagram = &entities.Diagram{Source: "component-a {\n}\n"}

	// Execute enhancement
	enhanced, err := uc.Execute(comp1, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should include internal relationship
	if !strings.Contains(enhanced, "component-a -> component-b") {
		t.Error("Enhanced diagram missing internal relationship")
	}

	// Should NOT include external relationship
	if strings.Contains(enhanced, "external-component") {
		t.Error("Enhanced diagram incorrectly included external relationship")
	}
}

func TestEnhanceComponentDiagramWithCodeAnnotations(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("Auth Service")
	component.Diagram = &entities.Diagram{Source: "auth {\n}\n"}

	// Add code annotations
	component.AddCodeAnnotation("internal/auth", "JWT token handling")
	component.AddCodeAnnotation("internal/middleware", "HTTP middleware")

	container.AddComponent(component)

	// Execute enhancement
	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify code annotations section exists
	if !strings.Contains(enhanced, "# Code Annotations") {
		t.Error("Enhanced diagram missing code annotations section")
	}

	// Verify annotations included
	if !strings.Contains(enhanced, "internal/auth") {
		t.Error("Enhanced diagram missing code annotation for internal/auth")
	}

	if !strings.Contains(enhanced, "JWT token handling") {
		t.Error("Enhanced diagram missing description for JWT handling")
	}
}

func TestEnhanceComponentDiagramWithExternalDependencies(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("API Gateway")
	component.Diagram = &entities.Diagram{Source: "gateway {\n}\n"}

	// Add external dependencies
	component.AddDependency("github.com/golang-jwt/jwt")
	component.AddDependency("github.com/gorilla/mux")

	container.AddComponent(component)

	// Execute enhancement
	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify external dependencies section exists
	if !strings.Contains(enhanced, "# External Dependencies") {
		t.Error("Enhanced diagram missing external dependencies section")
	}

	// Verify dependencies included
	if !strings.Contains(enhanced, "github.com/golang-jwt/jwt") {
		t.Error("Enhanced diagram missing JWT dependency")
	}

	if !strings.Contains(enhanced, "github.com/gorilla/mux") {
		t.Error("Enhanced diagram missing gorilla/mux dependency")
	}
}

func TestEnhanceComponentDiagramWithoutBaseDiagram(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("Simple Component")
	// No diagram provided

	component.AddRelationship("component-b", "depends on")
	otherComp, _ := entities.NewComponent("Component B")
	container.AddComponent(component)
	container.AddComponent(otherComp)

	// Execute enhancement - should create minimal diagram
	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should still create relationships even without base diagram
	if !strings.Contains(enhanced, "component-b") {
		t.Error("Enhanced diagram missing relationship without base diagram")
	}
}

func TestEnhanceComponentDiagramSanitizeID(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	tests := []struct {
		input    string
		expected string
	}{
		{"internal/auth", "internal_auth"},
		{"path/to/component", "path_to_component"},
		{"Component-With-Dashes", "Component_With_Dashes"},
		{"123-starts-with-number", "starts_with_number"},
		{"valid_id", "valid_id"},
		{"!!!invalid!!!", "invalid___"},
	}

	for _, tt := range tests {
		result := uc.sanitizeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEnhanceComponentDiagramCombineSourcesWithBraces(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	baseD2 := "component {\n  label: \"Test\"\n}\n"
	relationships := "\n# Relationships\ncomponent -> other: \"depends\""
	codeAnnotations := "\n# Code\ncode_path: \"path/to/code\""
	externalDeps := ""

	result := uc.combineD2Sources(baseD2, relationships, codeAnnotations, externalDeps)

	// Should insert enhancements before closing brace
	if !strings.Contains(result, "# Relationships") {
		t.Error("Result missing relationships section")
	}

	if !strings.Contains(result, "# Code") {
		t.Error("Result missing code annotations section")
	}

	// Verify closing brace still exists
	if !strings.HasSuffix(strings.TrimSpace(result), "}") {
		t.Error("Result missing closing brace")
	}
}

func TestEnhanceComponentDiagramWithQuotesInDescription(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	comp1, _ := entities.NewComponent("Component 1")
	comp2, _ := entities.NewComponent("Component 2")

	// Add relationship with quotes in description
	comp1.AddRelationship("comp-2", `calls "other" service`)

	comp1.Diagram = &entities.Diagram{Source: "comp-1 {\n}\n"}
	container.AddComponent(comp1)
	container.AddComponent(comp2)

	// Execute enhancement
	enhanced, err := uc.Execute(comp1, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify quotes were escaped
	if strings.Contains(enhanced, `calls "other" service`) && !strings.Contains(enhanced, `calls \"other\" service`) {
		t.Error("Enhanced diagram failed to escape quotes in description")
	}
}
