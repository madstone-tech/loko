// Package usecases — markdown table generators for BuildDocs.
// See build_docs.go for the use case type, constructors, and Execute methods.
package usecases

import (
	"fmt"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// GenerateComponentTable generates a Markdown table of components in a container.
// Returns a table with columns: Name, Technology, Description.
// If container has no components, returns an empty string.
func GenerateComponentTable(container *entities.Container) string {
	if container == nil || container.Components == nil || len(container.Components) == 0 {
		return ""
	}

	// Get all components and sort by name
	components := container.ListComponents()

	// Sort components by name
	slices.SortFunc(components, func(a, b *entities.Component) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Build the markdown table
	var sb strings.Builder
	sb.WriteString("| Name | Technology | Description |\n")
	sb.WriteString("|------|------------|-------------|\n")

	for _, comp := range components {
		// Escape pipe characters in description to avoid breaking table
		description := strings.ReplaceAll(comp.Description, "|", "\\|")
		technology := strings.ReplaceAll(comp.Technology, "|", "\\|")
		fmt.Fprintf(&sb, "| %s | %s | %s |\n", comp.Name, technology, description)
	}

	return sb.String()
}

// GenerateContainerTable generates a Markdown table of containers in a system.
// Returns a table with columns: Name, Technology, Description.
func GenerateContainerTable(system *entities.System) string {
	if system == nil || system.Containers == nil || len(system.Containers) == 0 {
		return ""
	}

	// Get all containers and sort by name
	containers := system.ListContainers()

	// Sort containers by name
	slices.SortFunc(containers, func(a, b *entities.Container) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Build the markdown table
	var sb strings.Builder
	sb.WriteString("| Name | Technology | Description |\n")
	sb.WriteString("|------|------------|-------------|\n")

	for _, cont := range containers {
		// Escape pipe characters in description to avoid breaking table
		description := strings.ReplaceAll(cont.Description, "|", "\\|")
		technology := strings.ReplaceAll(cont.Technology, "|", "\\|")
		fmt.Fprintf(&sb, "| %s | %s | %s |\n", cont.Name, technology, description)
	}

	return sb.String()
}
