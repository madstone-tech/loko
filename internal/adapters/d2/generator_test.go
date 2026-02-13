package d2_test

import (
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/d2"
	"github.com/madstone-tech/loko/internal/core/entities"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestGeneratorImplementsInterface verifies that Generator implements DiagramGenerator.
func TestGeneratorImplementsInterface(t *testing.T) {
	var _ usecases.DiagramGenerator = (*d2.Generator)(nil)
}

// TestGenerateSystemContextDiagram tests system context diagram generation.
func TestGenerateSystemContextDiagram(t *testing.T) {
	gen := d2.NewGenerator()

	system, err := entities.NewSystem("payment-system")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	system.Description = "Payment processing system"
	system.KeyUsers = []string{"Customer", "Admin"}
	system.ExternalSystems = []string{"Payment Gateway", "Email Service"}

	result, err := gen.GenerateSystemContextDiagram(system)
	if err != nil {
		t.Fatalf("GenerateSystemContextDiagram() error = %v", err)
	}

	if result == "" {
		t.Error("GenerateSystemContextDiagram() returned empty string")
	}

	// Verify key elements are present
	expectedElements := []string{
		"System Context Diagram",
		"payment-system",
		"Payment processing system",
		"Customer",
		"Admin",
		"Payment Gateway",
		"Email Service",
	}

	for _, elem := range expectedElements {
		if !contains(result, elem) {
			t.Errorf("GenerateSystemContextDiagram() missing element: %q", elem)
		}
	}
}

// TestGenerateContainerDiagram tests container diagram generation.
func TestGenerateContainerDiagram(t *testing.T) {
	gen := d2.NewGenerator()

	system, err := entities.NewSystem("payment-system")
	if err != nil {
		t.Fatalf("failed to create system: %v", err)
	}
	system.Description = "Payment processing system"
	system.KeyUsers = []string{"Customer"}

	// Add a container
	container, err := entities.NewContainer("api-gateway")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.Description = "REST API Gateway"
	container.Technology = "Go + Gin"

	if err := system.AddContainer(container); err != nil {
		t.Fatalf("failed to add container: %v", err)
	}

	result, err := gen.GenerateContainerDiagram(system)
	if err != nil {
		t.Fatalf("GenerateContainerDiagram() error = %v", err)
	}

	if result == "" {
		t.Error("GenerateContainerDiagram() returned empty string")
	}

	// Verify key elements are present
	expectedElements := []string{
		"Container Diagram",
		"payment-system",
		"api-gateway",
		"REST API Gateway",
		"Go + Gin",
	}

	for _, elem := range expectedElements {
		if !contains(result, elem) {
			t.Errorf("GenerateContainerDiagram() missing element: %q", elem)
		}
	}
}

// TestGenerateComponentDiagram tests component diagram generation.
func TestGenerateComponentDiagram(t *testing.T) {
	gen := d2.NewGenerator()

	container, err := entities.NewContainer("api-gateway")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	container.Description = "REST API Gateway"
	container.Technology = "Go + Gin"

	// Add a component
	component, err := entities.NewComponent("auth-handler")
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}
	component.Description = "Authentication handler"
	component.Technology = "JWT"

	if err := container.AddComponent(component); err != nil {
		t.Fatalf("failed to add component: %v", err)
	}

	result, err := gen.GenerateComponentDiagram(container)
	if err != nil {
		t.Fatalf("GenerateComponentDiagram() error = %v", err)
	}

	if result == "" {
		t.Error("GenerateComponentDiagram() returned empty string")
	}

	// Verify key elements are present
	expectedElements := []string{
		"Component Diagram",
		"api-gateway",
		"auth-handler",
		"Authentication handler",
		"JWT",
	}

	for _, elem := range expectedElements {
		if !contains(result, elem) {
			t.Errorf("GenerateComponentDiagram() missing element: %q", elem)
		}
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
