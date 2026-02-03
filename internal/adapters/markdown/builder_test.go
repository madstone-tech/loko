package markdown

import (
	"context"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestBuilder_BuildMarkdown(t *testing.T) {
	builder := NewBuilder()
	ctx := context.Background()

	// Create test project
	project, _ := entities.NewProject("TestProject")
	project.Description = "A test project for architecture documentation"
	project.Version = "1.0.0"

	// Create test systems
	sys1, _ := entities.NewSystem("AuthService")
	sys1.Description = "Handles authentication and authorization"
	sys1.Tags = []string{"security", "identity"}
	sys1.Responsibilities = []string{"User authentication", "Token management"}

	cont1, _ := entities.NewContainer("API")
	cont1.Description = "REST API for authentication"
	cont1.Technology = "Go"
	sys1.AddContainer(cont1)

	comp1, _ := entities.NewComponent("AuthHandler")
	comp1.Description = "Handles auth requests"
	comp1.Technology = "Go HTTP"
	cont1.AddComponent(comp1)

	sys2, _ := entities.NewSystem("UserService")
	sys2.Description = "User management service"

	systems := []*entities.System{sys1, sys2}

	// Build markdown
	content, err := builder.BuildMarkdown(ctx, project, systems)
	if err != nil {
		t.Fatalf("BuildMarkdown failed: %v", err)
	}

	// Verify content
	checks := []string{
		"# TestProject",
		"A test project for architecture documentation",
		"**Version:** 1.0.0",
		"## Table of Contents",
		"## AuthService",
		"Handles authentication and authorization",
		"**Tags:** security, identity",
		"**Responsibilities:**",
		"- User authentication",
		"### Containers",
		"#### API",
		"REST API for authentication",
		"**Technology:** Go",
		"| AuthHandler |",
		"## UserService",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Expected markdown to contain %q", check)
		}
	}
}

func TestBuilder_BuildSystemMarkdown(t *testing.T) {
	builder := NewBuilder()
	ctx := context.Background()

	sys, _ := entities.NewSystem("PaymentService")
	sys.Description = "Handles payment processing"
	sys.Tags = []string{"payments", "financial"}
	sys.Dependencies = []string{"AuthService", "DatabaseService"}

	cont, _ := entities.NewContainer("PaymentAPI")
	cont.Description = "Payment REST API"
	cont.Technology = "Node.js"
	sys.AddContainer(cont)

	containers := sys.ListContainers()

	content, err := builder.BuildSystemMarkdown(ctx, sys, containers)
	if err != nil {
		t.Fatalf("BuildSystemMarkdown failed: %v", err)
	}

	// Verify content
	checks := []string{
		"## PaymentService",
		"Handles payment processing",
		"**Tags:** payments, financial",
		"**Dependencies:**",
		"- AuthService",
		"- DatabaseService",
		"### Containers",
		"#### PaymentAPI",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Expected markdown to contain %q", check)
		}
	}
}

func TestBuilder_NilProject(t *testing.T) {
	builder := NewBuilder()
	ctx := context.Background()

	_, err := builder.BuildMarkdown(ctx, nil, nil)
	if err == nil {
		t.Error("Expected error for nil project")
	}
}

func TestBuilder_NilSystem(t *testing.T) {
	builder := NewBuilder()
	ctx := context.Background()

	_, err := builder.BuildSystemMarkdown(ctx, nil, nil)
	if err == nil {
		t.Error("Expected error for nil system")
	}
}

func TestBuilder_EmptySystems(t *testing.T) {
	builder := NewBuilder()
	ctx := context.Background()

	project, _ := entities.NewProject("EmptyProject")

	content, err := builder.BuildMarkdown(ctx, project, []*entities.System{})
	if err != nil {
		t.Fatalf("BuildMarkdown failed: %v", err)
	}

	if !strings.Contains(content, "# EmptyProject") {
		t.Error("Expected project header")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"auth_service", "auth-service"},
		{"API", "api"},
		{"User Service", "user-service"},
	}

	for _, tt := range tests {
		result := slugify(tt.input)
		if result != tt.expected {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
