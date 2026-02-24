package usecases

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// EnhanceComponentDiagram adds relationship edges and metadata to a component's D2 diagram.
//
// This use case enhances the basic component diagram with:
// 1. Relationship edges to other components in the same container
// 2. External dependency connections
// 3. Code annotation boxes
// 4. Metadata display
type EnhanceComponentDiagram struct{}

// NewEnhanceComponentDiagram creates a new EnhanceComponentDiagram use case.
func NewEnhanceComponentDiagram() *EnhanceComponentDiagram {
	return &EnhanceComponentDiagram{}
}

// Execute enhances the component's D2 diagram with relationships and metadata.
//
// It takes:
// - component: the component to enhance
// - container: the parent container (for resolving related components)
// - system: the parent system (for context)
//
// It returns the enhanced D2 source or an error.
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

	// Start with existing diagram source if present
	var d2Source string
	if component.Diagram != nil && component.Diagram.Source != "" {
		d2Source = component.Diagram.Source
	} else {
		// Create minimal diagram structure
		d2Source = fmt.Sprintf("%s {\n}\n", component.ID)
	}

	// Build relationship edges
	relationshipEdges := uc.buildRelationshipEdges(component, container)

	// Build code annotation boxes
	codeAnnotationBoxes := uc.buildCodeAnnotationBoxes(component)

	// Build external dependencies section
	externalDeps := uc.buildExternalDependencies(component)

	// Combine all enhancements
	enhanced := uc.combineD2Sources(d2Source, relationshipEdges, codeAnnotationBoxes, externalDeps)

	return enhanced, nil
}

// buildRelationshipEdges creates D2 edges for component relationships.
// Only includes relationships to other components in the same container.
func (uc *EnhanceComponentDiagram) buildRelationshipEdges(
	component *entities.Component,
	container *entities.Container,
) string {
	if len(component.Relationships) == 0 {
		return ""
	}

	// Create a map of component IDs in this container for quick lookup
	containerComponentIDs := make(map[string]bool)
	for id := range container.Components {
		containerComponentIDs[id] = true
	}

	var edges []string
	for targetID, description := range component.Relationships {
		// Only include relationships to components in the same container
		if !containerComponentIDs[targetID] {
			continue
		}

		// Create edge from this component to target component
		if description == "" {
			edges = append(edges, fmt.Sprintf("%s -> %s", component.ID, targetID))
		} else {
			// Escape quotes in description for D2
			escapedDesc := strings.ReplaceAll(description, "\"", "\\\"")
			edges = append(edges, fmt.Sprintf("%s -> %s: \"%s\"", component.ID, targetID, escapedDesc))
		}
	}

	if len(edges) == 0 {
		return ""
	}

	return "\n# Relationships\n" + strings.Join(edges, "\n")
}

// buildCodeAnnotationBoxes creates D2 boxes for code annotations.
func (uc *EnhanceComponentDiagram) buildCodeAnnotationBoxes(component *entities.Component) string {
	if len(component.CodeAnnotations) == 0 {
		return ""
	}

	var boxes []string
	boxes = append(boxes, "\n# Code Annotations")

	for codePath, description := range component.CodeAnnotations {
		// Sanitize path for use as D2 identifier
		safeID := uc.sanitizeID(codePath)

		// Escape quotes in description
		escapedDesc := strings.ReplaceAll(description, "\"", "\\\"")

		box := fmt.Sprintf("  %s: \"%s\" {\n    label: \"%s\"\n    style.text.font-size: 12\n  }", safeID, codePath, escapedDesc)
		boxes = append(boxes, box)
	}

	return strings.Join(boxes, "\n")
}

// buildExternalDependencies creates a section for external dependencies.
func (uc *EnhanceComponentDiagram) buildExternalDependencies(component *entities.Component) string {
	if len(component.Dependencies) == 0 {
		return ""
	}

	var depsSection []string
	depsSection = append(depsSection, "\n# External Dependencies")

	for i, dep := range component.Dependencies {
		// Create a safe ID for the dependency
		depID := fmt.Sprintf("dep_%d", i)

		// Format dependency name
		depsSection = append(depsSection, fmt.Sprintf("  %s: \"%s\" {\n    style.font-size: 11\n    style.stroke: \"#666\"\n  }", depID, dep))
	}

	return strings.Join(depsSection, "\n")
}

// combineD2Sources appends relationship edges and other enhancements after the last
// top-level closing brace in the base diagram.
//
// Component .d2 files follow the pattern:
//
//	comp-id: "Label" {
//	  tooltip: "..."
//	}
//
//	# trailing comment template lines (not real D2)
//
// Enhancements must be inserted at the top level (after the closing "}"), not inside
// the node block. Trailing comment/blank lines that follow the last top-level "}" are
// stripped so they don't appear between the node block and the new edges.
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

	// Find the last top-level closing brace (a "}" that starts at column 0).
	// This is the correct splice point for top-level D2 statements.
	lastTopLevelBrace := -1
	lines := strings.Split(baseDiagram, "\n")
	for i, line := range lines {
		if line == "}" {
			lastTopLevelBrace = i
		}
	}

	if lastTopLevelBrace != -1 {
		// Keep everything up to and including the closing brace, then append enhancements.
		core := strings.Join(lines[:lastTopLevelBrace+1], "\n")
		return core + "\n" + enhancements
	}

	// Fallback: no top-level brace found — append enhancements to whatever is there.
	return strings.TrimRight(baseDiagram, "\n") + "\n" + enhancements
}

// sanitizeID converts a path or name into a safe D2 identifier.
// D2 identifiers must be alphanumeric with underscores only (no hyphens).
func (uc *EnhanceComponentDiagram) sanitizeID(input string) string {
	// Replace special characters with underscores
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		// Replace everything else (including hyphens) with underscores
		return '_'
	}, input)

	// Remove leading underscores and digits (D2 identifiers should start with letter)
	safe = strings.TrimLeft(safe, "0123456789_")
	if safe == "" {
		safe = "item"
	}

	return safe
}
