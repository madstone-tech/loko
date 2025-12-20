package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// D2Generator generates D2 diagram source code from architecture entities.
type D2Generator struct{}

// NewD2Generator creates a new D2 generator.
func NewD2Generator() *D2Generator {
	return &D2Generator{}
}

// GenerateSystemContextDiagram creates a C4 Level 1 system context diagram.
// Shows the system with external users and systems.
func (dg *D2Generator) GenerateSystemContextDiagram(system *entities.System) string {
	var sb strings.Builder

	sb.WriteString("# System Context Diagram\n")
	sb.WriteString("# C4 Level 1 - System Context\n")
	sb.WriteString(fmt.Sprintf("# System: %s\n", system.Name))
	sb.WriteString(fmt.Sprintf("# Description: %s\n\n", system.Description))

	sb.WriteString("direction: right\n\n")

	// Add users
	sb.WriteString("# Primary users/actors\n")
	if len(system.KeyUsers) > 0 {
		for i, user := range system.KeyUsers {
			userID := fmt.Sprintf("user_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s: \"%s\"\n", userID, user))
		}
	} else {
		sb.WriteString("user: \"User/Actor\"\n")
	}
	sb.WriteString("\n")

	// Add main system
	sb.WriteString("# Main system\n")
	sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", system.ID, system.Name))
	sb.WriteString(fmt.Sprintf("  description: \"%s\"\n", system.Description))
	sb.WriteString("}\n\n")

	// Add relationships with users
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

	// Add external systems
	if len(system.ExternalSystems) > 0 {
		sb.WriteString("# External system integrations\n")
		for i, extSys := range system.ExternalSystems {
			extID := fmt.Sprintf("external_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", extID, extSys))
			sb.WriteString("  style { fill: \"#FFF3E0\" }\n")
			sb.WriteString("}\n")
		}
		sb.WriteString("\n")

		for i := range system.ExternalSystems {
			extID := fmt.Sprintf("external_%d", i+1)
			sb.WriteString(fmt.Sprintf("%s -> %s: \"Integrates with\"\n", system.ID, extID))
		}
		sb.WriteString("\n")
	}

	// Styling
	sb.WriteString("# Styling\n")
	sb.WriteString(fmt.Sprintf("%s: {\n", system.ID))
	sb.WriteString("  style {\n")
	sb.WriteString("    fill: \"#E1F5FF\"\n")
	sb.WriteString("    stroke: \"#01579B\"\n")
	sb.WriteString("    stroke-width: 2\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateContainerDiagram creates a C4 Level 2 container diagram.
// Shows the system's internal containers.
func (dg *D2Generator) GenerateContainerDiagram(system *entities.System) string {
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
			sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", userID, user))
			sb.WriteString("  style { fill: \"#FFF3E0\" }\n")
			sb.WriteString("}\n")
		}
	} else {
		sb.WriteString("user: \"User/Actor\" { style { fill: \"#FFF3E0\" } }\n")
	}
	sb.WriteString("\n")

	// Add system as container group
	sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", system.ID, system.Name))
	sb.WriteString(fmt.Sprintf("  description: \"%s\"\n\n", system.Description))

	// Add containers
	if system.ContainerCount() > 0 {
		for _, container := range system.ListContainers() {
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

	// Container relationships (optional)
	if system.ContainerCount() > 1 {
		sb.WriteString("\n# Container interactions (add as needed)\n")
		containers := system.ListContainers()
		if len(containers) >= 2 {
			sb.WriteString(fmt.Sprintf("# %s.%s -> %s.%s: \"Communicates via\"\n",
				system.ID, containers[0].ID, system.ID, containers[1].ID))
		}
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

// UpdateSystemD2File updates the system's D2 diagram file with current containers.
// This is called when containers are added/removed to keep the diagram in sync.
func (dg *D2Generator) UpdateSystemD2File(system *entities.System) error {
	if system.Path == "" {
		return fmt.Errorf("system path is not set")
	}

	// Generate the container diagram
	d2Content := dg.GenerateContainerDiagram(system)

	// Write to system D2 file
	d2Path := filepath.Join(system.Path, system.ID+".d2")
	if err := os.WriteFile(d2Path, []byte(d2Content), 0644); err != nil {
		return fmt.Errorf("failed to update D2 file: %w", err)
	}

	return nil
}

// SaveSystemContextD2File saves a new system context diagram.
// This is called when a system is first created.
func (dg *D2Generator) SaveSystemContextD2File(system *entities.System) error {
	if system.Path == "" {
		return fmt.Errorf("system path is not set")
	}

	d2Content := dg.GenerateSystemContextDiagram(system)

	d2Path := filepath.Join(system.Path, system.ID+".d2")
	if err := os.WriteFile(d2Path, []byte(d2Content), 0644); err != nil {
		return fmt.Errorf("failed to save system context D2 file: %w", err)
	}

	return nil
}

// GenerateComponentDiagram creates a C4 Level 3 component diagram.
// Shows the component structure within a container.
func (dg *D2Generator) GenerateComponentDiagram(container *entities.Container) string {
	var sb strings.Builder

	sb.WriteString("# Component Diagram\n")
	sb.WriteString("# C4 Level 3 - Component View\n")
	sb.WriteString(fmt.Sprintf("# Container: %s\n\n", container.Name))

	sb.WriteString("direction: right\n\n")

	// Add components
	if container.ComponentCount() > 0 {
		sb.WriteString("# Components\n")
		for _, component := range container.ListComponents() {
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
	if container.ComponentCount() > 1 {
		sb.WriteString("# Component interactions (add as needed)\n")
		components := container.ListComponents()
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

// UpdateContainerD2File updates the container's D2 diagram file with current components.
// This is called when components are added/removed to keep the diagram in sync.
func (dg *D2Generator) UpdateContainerD2File(container *entities.Container) error {
	if container.Path == "" {
		return fmt.Errorf("container path is not set")
	}

	// Generate the component diagram
	d2Content := dg.GenerateComponentDiagram(container)

	// Write to container D2 file
	d2Path := filepath.Join(container.Path, container.ID+".d2")
	if err := os.WriteFile(d2Path, []byte(d2Content), 0644); err != nil {
		return fmt.Errorf("failed to update D2 file: %w", err)
	}

	return nil
}
