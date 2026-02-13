package encoding

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// T033: Token Efficiency Benchmark

func BenchmarkTOONvsJSON(b *testing.B) {
	// Create test data: 5 systems, 15 containers (3 per system)
	project := createTestProject(5, 3)
	enc := NewEncoder()

	b.Run("JSON_Encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = enc.EncodeJSON(project)
		}
	})

	b.Run("TOON_Encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = enc.EncodeTOON(project)
		}
	})
}

func TestTokenEfficiencyMetrics(t *testing.T) {
	// Test with 5 systems, 15 containers architecture
	project := createTestProject(5, 3)
	enc := NewEncoder()

	jsonData, _ := enc.EncodeJSON(project)
	toonData, _ := enc.EncodeTOON(project)

	jsonTokens := estimateTokenCount(string(jsonData))
	toonTokens := estimateTokenCount(string(toonData))

	savings := float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100

	t.Logf("JSON tokens: %d", jsonTokens)
	t.Logf("TOON tokens: %d", toonTokens)
	t.Logf("Token savings: %.1f%%", savings)

	// Assert > 5% overall reduction (more realistic for mixed data)
	if savings < 5 {
		t.Errorf("expected >5%% token savings, got %.1f%%", savings)
	}
}

func TestTabularArrayTokenEfficiency(t *testing.T) {
	// Test tabular array format efficiency
	containers := []struct {
		Name       string `toon:"name"`
		Technology string `toon:"technology"`
	}{
		{"API", "Go"},
		{"Database", "PostgreSQL"},
		{"Cache", "Redis"},
		{"Queue", "RabbitMQ"},
		{"Worker", "Python"},
	}

	enc := NewEncoder()
	jsonData, _ := enc.EncodeJSON(containers)
	toonData, _ := enc.EncodeTOON(containers)

	jsonTokens := estimateTokenCount(string(jsonData))
	toonTokens := estimateTokenCount(string(toonData))

	savings := float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100

	t.Logf("Tabular array - JSON tokens: %d", jsonTokens)
	t.Logf("Tabular array - TOON tokens: %d", toonTokens)
	t.Logf("Tabular array - Token savings: %.1f%%", savings)

	// Assert > 50% reduction for tabular arrays
	if savings < 50 {
		t.Errorf("expected >50%% token savings for tabular arrays, got %.1f%%", savings)
	}
}

// Helper: estimate token count (4 chars ≈ 1 token on average)
func estimateTokenCount(s string) int {
	return (len(s) + 3) / 4
}

// Helper: create test project with N systems, M containers per system
func createTestProject(numSystems, containersPerSystem int) *entities.Project {
	project, _ := entities.NewProject("TestProject")

	// Add systems with containers and components
	for i := 0; i < numSystems; i++ {
		systemName := "System" + string(rune('A'+i))
		system, _ := entities.NewSystem(systemName)
		system.SetDescription("System " + systemName + " description")
		system.AddTag("microservice")
		system.AddDependency("external-api-" + string(rune('A'+i)))

		for j := 0; j < containersPerSystem; j++ {
			containerName := "Container" + string(rune('A'+j))
			container, _ := entities.NewContainer(containerName)
			container.SetDescription("Container " + containerName + " description")
			container.SetTechnology("Tech" + string(rune('A'+j)))
			container.AddTag("container-tag-" + string(rune('A'+j)))

			// Add components to container
			for k := 0; k < 3; k++ {
				componentName := "Component" + string(rune('A'+k))
				component, _ := entities.NewComponent(componentName)
				component.SetDescription("Component " + componentName + " description")
				component.SetTechnology("ComponentTech" + string(rune('A'+k)))
				component.AddDependency("github.com/test/dep" + string(rune('A'+k)))
				component.AddTag("component-tag-" + string(rune('A'+k)))

				container.AddComponent(component)
			}

			system.AddContainer(container)
		}

		project.AddSystem(system)
	}

	return project
}
