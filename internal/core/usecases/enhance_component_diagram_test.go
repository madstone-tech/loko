package usecases

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// combineD2Sources is a test-only helper that appends enhancements after the last
// top-level "}" in a raw D2 source string.  It was extracted from EnhanceComponentDiagram
// when production code was rewritten to synthesise diagrams from entities; the tests
// below keep the helper alive to verify the splice logic independently.
func (uc *EnhanceComponentDiagram) combineD2Sources(
	baseDiagram string,
	relationships string,
	codeAnnotations string,
	externalDeps string,
) string {
	enhancements := relationships + codeAnnotations + externalDeps
	if enhancements == "" {
		return baseDiagram
	}

	lastTopLevelBrace := -1
	lines := strings.Split(baseDiagram, "\n")
	for i, line := range lines {
		if line == "}" {
			lastTopLevelBrace = i
		}
	}

	if lastTopLevelBrace != -1 {
		core := strings.Join(lines[:lastTopLevelBrace+1], "\n")
		return core + "\n" + enhancements
	}

	return strings.TrimRight(baseDiagram, "\n") + "\n" + enhancements
}

// helper: build a minimal populated system + container with named components.
func buildTestScaffold() (
	*entities.System,
	*entities.Container,
	*entities.Component, // auth (focal)
	*entities.Component, // auth-cache
	*entities.Component, // user-database
) {
	system, _ := entities.NewSystem("User Service")
	container, _ := entities.NewContainer("API Server")
	container.SetDescription("REST API server")

	auth, _ := entities.NewComponent("Authentication")
	auth.SetDescription("Handles user authentication")
	auth.SetTechnology("Go / JWT")

	authCache, _ := entities.NewComponent("Auth Cache")
	authCache.SetDescription("Caches authentication tokens")

	userDB, _ := entities.NewComponent("User Database")
	userDB.SetDescription("Stores user data")

	container.AddComponent(auth)
	container.AddComponent(authCache)
	container.AddComponent(userDB)

	return system, container, auth, authCache, userDB
}

func TestEnhanceComponentDiagramShowsAllSiblings(t *testing.T) {
	uc := NewEnhanceComponentDiagram()
	system, container, auth, _, _ := buildTestScaffold()

	enhanced, err := uc.Execute(auth, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	for _, want := range []string{"Authentication", "Auth Cache", "User Database"} {
		if !strings.Contains(enhanced, want) {
			t.Errorf("diagram missing sibling node %q", want)
		}
	}
}

func TestEnhanceComponentDiagramIncludesDescriptions(t *testing.T) {
	uc := NewEnhanceComponentDiagram()
	system, container, auth, _, _ := buildTestScaffold()

	enhanced, err := uc.Execute(auth, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "Handles user authentication") {
		t.Error("diagram missing focal component description")
	}
	if !strings.Contains(enhanced, "Caches authentication tokens") {
		t.Error("diagram missing sibling description")
	}
}

func TestEnhanceComponentDiagramHighlightsFocalComponent(t *testing.T) {
	uc := NewEnhanceComponentDiagram()
	system, container, auth, _, _ := buildTestScaffold()

	enhanced, err := uc.Execute(auth, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "stroke-width: 3") {
		t.Error("focal component not highlighted (missing stroke-width: 3)")
	}
}

func TestEnhanceComponentDiagramWithRelationships(t *testing.T) {
	uc := NewEnhanceComponentDiagram()
	system, container, auth, _, _ := buildTestScaffold()

	auth.AddRelationship("auth-cache", "stores sessions in")
	auth.AddRelationship("user-database", "queries user data from")

	enhanced, err := uc.Execute(auth, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "authentication -> auth-cache") {
		t.Error("missing edge auth -> auth-cache")
	}
	if !strings.Contains(enhanced, "authentication -> user-database") {
		t.Error("missing edge auth -> user-database")
	}
	if !strings.Contains(enhanced, "stores sessions in") {
		t.Error("missing edge label 'stores sessions in'")
	}
}

func TestEnhanceComponentDiagramFiltersExternalRelationships(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service A")
	container, _ := entities.NewContainer("API")

	comp1, _ := entities.NewComponent("Component A")
	comp2, _ := entities.NewComponent("Component B")

	comp1.AddRelationship("external-component", "calls external service")
	comp1.AddRelationship("component-b", "internal dependency")

	container.AddComponent(comp1)
	container.AddComponent(comp2)

	enhanced, err := uc.Execute(comp1, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "component-a -> component-b") {
		t.Error("missing internal relationship edge")
	}
	if strings.Contains(enhanced, "external-component") {
		t.Error("diagram incorrectly included external relationship")
	}
}

func TestEnhanceComponentDiagramWithCodeAnnotations(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("Auth Service")
	component.AddCodeAnnotation("internal/auth", "JWT token handling")
	component.AddCodeAnnotation("internal/middleware", "HTTP middleware")

	container.AddComponent(component)

	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "# Code Annotations") {
		t.Error("missing code annotations section")
	}
	if !strings.Contains(enhanced, "JWT token handling") {
		t.Error("missing annotation description")
	}
}

func TestEnhanceComponentDiagramWithExternalDependencies(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("API Gateway")
	component.AddDependency("github.com/golang-jwt/jwt")
	component.AddDependency("github.com/gorilla/mux")

	container.AddComponent(component)

	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "# External Dependencies") {
		t.Error("missing external dependencies section")
	}
	if !strings.Contains(enhanced, "github.com/golang-jwt/jwt") {
		t.Error("missing jwt dependency")
	}
	if !strings.Contains(enhanced, "github.com/gorilla/mux") {
		t.Error("missing gorilla/mux dependency")
	}
}

func TestEnhanceComponentDiagramWithoutBaseDiagram(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	component, _ := entities.NewComponent("Simple Component")
	otherComp, _ := entities.NewComponent("Component B")

	component.AddRelationship("component-b", "depends on")
	container.AddComponent(component)
	container.AddComponent(otherComp)

	enhanced, err := uc.Execute(component, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(enhanced, "Simple Component") {
		t.Error("missing focal component name")
	}
	if !strings.Contains(enhanced, "Component B") {
		t.Error("missing sibling component name")
	}
	if !strings.Contains(enhanced, "simple-component -> component-b") {
		t.Error("missing relationship edge")
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

	closingBrace := strings.Index(result, "}")
	relIdx := strings.Index(result, "# Relationships")
	if relIdx <= closingBrace {
		t.Errorf("Relationships injected inside node block (brace@%d, rel@%d); must be top-level", closingBrace, relIdx)
	}

	if !strings.Contains(result, "# Relationships") {
		t.Error("Result missing relationships section")
	}
	if !strings.Contains(result, "# Code") {
		t.Error("Result missing code annotations section")
	}
}

func TestCombineD2SourcesStripsTrailingComments(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	baseD2 := `direction: right

host-header-router: "Host Header Router" {
  tooltip: "routes traffic"
}

# Dependencies (if any)
# Example:
# cache: "Cache Layer"
# 
# host-header-router -> cache: uses

# Relationships (if any)
# Add component relationships here using the format:
# host-header-router -> other_component: "relationship_type"

`
	relationships := "\n# Relationships\nhost-header-router -> internal-alb: \"routes\""

	result := uc.combineD2Sources(baseD2, relationships, "", "")

	if !strings.Contains(result, "host-header-router -> internal-alb") {
		t.Error("Result missing relationship edge")
	}

	closingBrace := strings.Index(result, "}")
	edgeIdx := strings.Index(result, "host-header-router -> internal-alb")
	if edgeIdx < closingBrace {
		t.Errorf("Relationship edge is inside node block (brace@%d, edge@%d); must be top-level", closingBrace, edgeIdx)
	}

	commentIdx := strings.Index(result, "# Dependencies (if any)")
	if commentIdx != -1 && commentIdx < edgeIdx {
		t.Error("Template comment block appears before relationship edge")
	}
}

func TestEnhanceComponentDiagramWithQuotesInDescription(t *testing.T) {
	uc := NewEnhanceComponentDiagram()

	system, _ := entities.NewSystem("Service")
	container, _ := entities.NewContainer("API")

	comp1, _ := entities.NewComponent("Component 1")
	comp2, _ := entities.NewComponent("Component 2")

	comp1.AddRelationship("comp-2", `calls "other" service`)
	container.AddComponent(comp1)
	container.AddComponent(comp2)

	enhanced, err := uc.Execute(comp1, container, system)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if strings.Contains(enhanced, `calls "other" service`) && !strings.Contains(enhanced, `calls \"other\" service`) {
		t.Error("Enhanced diagram failed to escape quotes in description")
	}
}
