package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// convertInterfaceSlice converts a slice of interface{} to a slice of strings.
func convertInterfaceSlice(slice []any) []string {
	if len(slice) == 0 {
		return nil
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

// countDiagrams counts total diagrams in all systems and containers.
func countDiagrams(systems []*entities.System) int {
	count := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			count++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				count++
			}
		}
	}
	return count
}

// createContainerD2Template creates a basic D2 diagram template for a container.
func createContainerD2Template(_ context.Context, projectRoot, systemID string, container *entities.Container) error {
	containerDir := filepath.Join(projectRoot, "src", systemID, container.ID)
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return err
	}

	d2Template := fmt.Sprintf(`# %s Container Diagram
# C4 Level 2 - Container

direction: right

%s: "%s" {
  description: "%s"
  technology: "%s"
}
`, container.Name, container.ID, container.Name, container.Description, container.Technology)

	diagramPath := filepath.Join(containerDir, container.ID+".d2")
	return os.WriteFile(diagramPath, []byte(d2Template), 0644)
}

// validateDiagramStructure checks for structural and C4 compliance issues.
func validateDiagramStructure(d2Source, level string) ([]string, []string) {
	var warnings []string
	var suggestions []string

	// Check for comments
	if !containsSubstring(d2Source, "#") {
		suggestions = append(suggestions, "Add comments to explain diagram structure")
	}

	// Level-specific checks
	switch level {
	case "system":
		if !containsSubstring(d2Source, "User") && !containsSubstring(d2Source, "user") {
			suggestions = append(suggestions, "C4 Level 1 typically includes 'User' - consider adding user/actor")
		}
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "System context diagram should have at least 2 nodes (User and System)")
		}

	case "container":
		if countDiagramNodes(d2Source) < 2 {
			warnings = append(warnings, "Container diagram should have at least 2 components")
		}
		if !containsSubstring(d2Source, "{\n") {
			suggestions = append(suggestions, "Consider grouping related components with container blocks { }")
		}

	case "component":
		if countDiagramNodes(d2Source) < 1 {
			warnings = append(warnings, "Component diagram should have at least 1 component")
		}
	}

	// General best practices
	if !containsSubstring(d2Source, "direction:") && !containsSubstring(d2Source, "direction ") {
		suggestions = append(suggestions, "Consider specifying diagram direction (e.g., 'direction: right') for clarity")
	}

	if !containsSubstring(d2Source, "tooltip") && !containsSubstring(d2Source, "description") {
		suggestions = append(suggestions, "Add tooltips or descriptions to nodes for better documentation")
	}

	return warnings, suggestions
}

// containsSubstring checks if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// countDiagramNodes counts the number of nodes in a D2 diagram source.
func countDiagramNodes(d2Source string) int {
	count := 0
	// Count lines with nodes (heuristic: lines with colons that aren't comments or directives)
	for line := range strings.SplitSeq(d2Source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "direction") {
			count++
		}
	}
	return count
}

// updateSystemD2Diagram updates a system's D2 diagram with its current containers.
// This mirrors the CLI behavior of auto-syncing diagrams when containers are added.
func updateSystemD2Diagram(_ context.Context, _ string, system *entities.System) error {
	if system.Path == "" {
		return fmt.Errorf("system path not set")
	}

	// Generate container diagram content
	d2Content := generateContainerDiagram(system)

	// Write to system D2 file
	d2Path := filepath.Join(system.Path, system.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}

// updateContainerD2Diagram updates a container's D2 diagram with its current components.
// This mirrors the CLI behavior of auto-syncing diagrams when components are added.
func updateContainerD2Diagram(_ context.Context, _ string, container *entities.Container) error {
	if container.Path == "" {
		return fmt.Errorf("container path not set")
	}

	// Generate component diagram content
	d2Content := generateComponentDiagram(container)

	// Write to container D2 file
	d2Path := filepath.Join(container.Path, container.ID+".d2")
	return os.WriteFile(d2Path, []byte(d2Content), 0644)
}

// generateContainerDiagram creates a C4 Level 2 container diagram.
func generateContainerDiagram(system *entities.System) string {
	var sb strings.Builder

	sb.WriteString("# Container Diagram\n")
	sb.WriteString("# C4 Level 2 - Container View\n")
	sb.WriteString(fmt.Sprintf("# System: %s\n\n", system.Name))

	sb.WriteString("direction: right\n\n")

	// Add users
	sb.WriteString("# External users\n")
	if len(system.KeyUsers) > 0 {
		for i, user := range system.KeyUsers {
			userID := fmt.Sprintf("user_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s: \"%s\" { style { fill: \"#FFF3E0\" } }\n", userID, user))
		}
	} else {
		sb.WriteString("user: \"User/Actor\" { style { fill: \"#FFF3E0\" } }\n")
	}
	sb.WriteString("\n")

	// Add system as container group
	sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", system.ID, system.Name))
	sb.WriteString(fmt.Sprintf("  description: \"%s\"\n\n", system.Description))

	// Add containers
	if len(system.Containers) > 0 {
		for _, container := range system.Containers {
			sb.WriteString(fmt.Sprintf("  %s: \"%s\" {\n", container.ID, container.Name))
			if container.Description != "" {
				sb.WriteString(fmt.Sprintf("    description: \"%s\"\n", container.Description))
			}
			if container.Technology != "" {
				sb.WriteString(fmt.Sprintf("    technology: \"%s\"\n", container.Technology))
			}
			sb.WriteString("    style { fill: \"#E3F2FD\" }\n")
			sb.WriteString("  }\n")
		}
	} else {
		sb.WriteString("  # (Add containers here)\n")
	}

	sb.WriteString("}\n\n")

	// Add relationships
	sb.WriteString("# User interactions\n")
	if len(system.KeyUsers) > 0 {
		for i := range system.KeyUsers {
			userID := fmt.Sprintf("user_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s -> %s: \"Uses\"\n", userID, system.ID))
		}
	} else {
		sb.WriteString(fmt.Sprintf("user -> %s: \"Uses\"\n", system.ID))
	}

	sb.WriteString("\n")

	// System styling
	sb.WriteString(fmt.Sprintf("%s: {\n", system.ID))
	sb.WriteString("  style {\n")
	sb.WriteString("    fill: \"#E1F5FF\"\n")
	sb.WriteString("    stroke: \"#01579B\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateComponentDiagram creates a C4 Level 3 component diagram.
func generateComponentDiagram(container *entities.Container) string {
	var sb strings.Builder

	sb.WriteString("# Component Diagram\n")
	sb.WriteString("# C4 Level 3 - Component View\n")
	sb.WriteString(fmt.Sprintf("# Container: %s\n\n", container.Name))

	sb.WriteString("direction: right\n\n")

	// Add components
	if len(container.Components) > 0 {
		sb.WriteString("# Components\n")
		for _, component := range container.Components {
			sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", component.ID, component.Name))
			if component.Description != "" {
				sb.WriteString(fmt.Sprintf("  description: \"%s\"\n", component.Description))
			}
			if component.Technology != "" {
				sb.WriteString(fmt.Sprintf("  technology: \"%s\"\n", component.Technology))
			}
			sb.WriteString("  style { fill: \"#E3F2FD\" }\n")
			sb.WriteString("}\n")
		}
	} else {
		sb.WriteString("# (Add components here)\n")
	}

	sb.WriteString("\n")

	// Component relationships (optional)
	if len(container.Components) > 1 {
		sb.WriteString("# Component interactions (add as needed)\n")
		components := make([]*entities.Component, 0, len(container.Components))
		for _, c := range container.Components {
			components = append(components, c)
		}
		if len(components) >= 2 {
			sb.WriteString(fmt.Sprintf("# %s -> %s: \"Communicates via\"\n",
				components[0].ID, components[1].ID))
		}
	}

	sb.WriteString("\n")

	// Styling
	sb.WriteString(fmt.Sprintf("%s: {\n", container.ID))
	sb.WriteString("  style {\n")
	sb.WriteString("    fill: \"#E3F2FD\"\n")
	sb.WriteString("    stroke: \"#01579B\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}
