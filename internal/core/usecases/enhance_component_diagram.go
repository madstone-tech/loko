package usecases

import (
	"fmt"
	"sort"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// EnhanceComponentDiagram generates a full C4 Level 3 component diagram for a focal
// component within its parent container.
//
// Instead of rendering the sparse single-node stub stored in the component's own .d2
// file, it synthesises a complete diagram from the live entity data so that:
//   - All sibling components in the container appear as labelled nodes with descriptions
//     and technology annotations.
//   - The focal component is visually highlighted (accent fill + thicker border).
//   - All intra-container relationships (from every component's Relationships map) are
//     rendered as directed edges with labels.
//   - Code annotations and external dependencies for the focal component are appended.
type EnhanceComponentDiagram struct{}

// NewEnhanceComponentDiagram creates a new EnhanceComponentDiagram use case.
func NewEnhanceComponentDiagram() *EnhanceComponentDiagram {
	return &EnhanceComponentDiagram{}
}

// Execute generates the enhanced D2 source for the component's diagram.
//
// It returns the complete D2 source string or an error.
func (uc *EnhanceComponentDiagram) Execute(
	component *entities.Component,
	container *entities.Container,
	system *entities.System,
) (string, error) {
	if component == nil {
		return "", fmt.Errorf("component cannot be nil")
	}
	if container == nil {
		return "", fmt.Errorf("container cannot be nil")
	}
	if system == nil {
		return "", fmt.Errorf("system cannot be nil")
	}

	var sb strings.Builder

	sb.WriteString("# Component Diagram\n")
	sb.WriteString("# C4 Level 3 - Component\n")
	sb.WriteString(fmt.Sprintf("# Container: %s / %s\n", system.Name, container.Name))
	sb.WriteString(fmt.Sprintf("# Focal component: %s\n\n", component.Name))
	sb.WriteString("direction: right\n\n")

	// Emit all sibling components as labelled nodes.
	// Sort for deterministic output.
	components := container.ListComponents()
	sort.Slice(components, func(i, j int) bool {
		return components[i].ID < components[j].ID
	})

	for _, comp := range components {
		sb.WriteString(fmt.Sprintf("%s: \"%s\" {\n", comp.ID, uc.escapeD2String(comp.Name)))
		if comp.Description != "" {
			sb.WriteString(fmt.Sprintf("  description: \"%s\"\n", uc.escapeD2String(comp.Description)))
		}
		if comp.Technology != "" {
			sb.WriteString(fmt.Sprintf("  technology: \"%s\"\n", uc.escapeD2String(comp.Technology)))
		}
		if comp.ID == component.ID {
			// Focal component: highlighted accent style
			sb.WriteString("  style {\n")
			sb.WriteString("    fill: \"#E1F5FF\"\n")
			sb.WriteString("    stroke: \"#01579B\"\n")
			sb.WriteString("    stroke-width: 3\n")
			sb.WriteString("  }\n")
		} else {
			sb.WriteString("  style { fill: \"#E3F2FD\" }\n")
		}
		sb.WriteString("}\n")
	}

	// Emit all intra-container relationship edges.
	containerIDs := make(map[string]bool, len(components))
	for _, comp := range components {
		containerIDs[comp.ID] = true
	}

	// Collect edges from all components for deterministic ordering
	type edge struct{ from, to, label string }
	var edges []edge
	for _, comp := range components {
		// Sort target IDs for determinism
		targets := make([]string, 0, len(comp.Relationships))
		for t := range comp.Relationships {
			targets = append(targets, t)
		}
		sort.Strings(targets)
		for _, targetID := range targets {
			if !containerIDs[targetID] {
				continue
			}
			edges = append(edges, edge{comp.ID, targetID, comp.Relationships[targetID]})
		}
	}

	if len(edges) > 0 {
		sb.WriteString("\n# Relationships\n")
		for _, e := range edges {
			if e.label == "" {
				sb.WriteString(fmt.Sprintf("%s -> %s\n", e.from, e.to))
			} else {
				sb.WriteString(fmt.Sprintf("%s -> %s: \"%s\"\n", e.from, e.to, uc.escapeD2String(e.label)))
			}
		}
	}

	// Append code annotations for the focal component.
	if len(component.CodeAnnotations) > 0 {
		sb.WriteString("\n# Code Annotations\n")
		paths := make([]string, 0, len(component.CodeAnnotations))
		for p := range component.CodeAnnotations {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, codePath := range paths {
			desc := component.CodeAnnotations[codePath]
			safeID := uc.sanitizeID(codePath)
			sb.WriteString(fmt.Sprintf("  %s: \"%s\" {\n", safeID, uc.escapeD2String(codePath)))
			sb.WriteString(fmt.Sprintf("    label: \"%s\"\n", uc.escapeD2String(desc)))
			sb.WriteString("    style.text.font-size: 12\n")
			sb.WriteString("  }\n")
		}
	}

	// Append external dependencies for the focal component.
	if len(component.Dependencies) > 0 {
		sb.WriteString("\n# External Dependencies\n")
		deps := component.ListDependencies()
		sort.Strings(deps)
		for i, dep := range deps {
			depID := fmt.Sprintf("dep_%d", i)
			sb.WriteString(fmt.Sprintf("  %s: \"%s\" {\n", depID, uc.escapeD2String(dep)))
			sb.WriteString("    style.font-size: 11\n")
			sb.WriteString("    style.stroke: \"#666\"\n")
			sb.WriteString("  }\n")
		}
	}

	return sb.String(), nil
}

// escapeD2String escapes double-quote characters inside a D2 string literal.
func (uc *EnhanceComponentDiagram) escapeD2String(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// sanitizeID converts a path or name into a safe D2 identifier.
// D2 identifiers must be alphanumeric with underscores only (no hyphens).
func (uc *EnhanceComponentDiagram) sanitizeID(input string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, input)
	safe = strings.TrimLeft(safe, "0123456789_")
	if safe == "" {
		safe = "item"
	}
	return safe
}
