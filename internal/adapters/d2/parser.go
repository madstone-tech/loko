package d2

import (
	"context"
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

// D2Parser implements the D2Parser port interface using the official D2 library.
// It parses D2 diagram source code and extracts relationship arrows.
type D2Parser struct {
	// Future: Could add configuration options like compile options
}

// NewD2Parser creates a new D2Parser instance.
func NewD2Parser() *D2Parser {
	return &D2Parser{}
}

// ParseRelationships extracts relationship arrows from D2 source code.
//
// Implementation uses oss.terrastruct.com/d2 library to:
// 1. Compile D2 source into an AST
// 2. Walk the graph to find edges (connections)
// 3. Extract source, target, and label for each edge
//
// Error handling:
// - Invalid syntax: Returns error with parse details
// - Empty file: Returns empty slice (valid state)
// - No relationships: Returns empty slice (valid state)
func (p *D2Parser) ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error) {
	// Handle empty input gracefully
	if strings.TrimSpace(d2Source) == "" {
		return []entities.D2Relationship{}, nil
	}

	// Create minimal compile options with a text ruler for dimension calculation
	ruler, _ := textmeasure.NewRuler()
	compileOpts := &d2lib.CompileOptions{
		Ruler: ruler,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			return d2dagrelayout.DefaultLayout, nil
		},
	}

	// Compile D2 source using official library
	// We only need the graph, not the rendered diagram (pass nil for renderOpts)
	_, graph, err := d2lib.Compile(ctx, d2Source, compileOpts, nil)
	if err != nil {
		return nil, fmt.Errorf("D2 parse error: %w", err)
	}

	if graph == nil {
		// Valid parse but no graph produced
		return []entities.D2Relationship{}, nil
	}

	// Extract relationships from the compiled graph
	relationships := extractRelationshipsFromGraph(graph)

	return relationships, nil
}

// extractRelationshipsFromGraph walks the D2 graph and extracts relationships (edges).
func extractRelationshipsFromGraph(graph *d2graph.Graph) []entities.D2Relationship {
	if graph == nil {
		return []entities.D2Relationship{}
	}

	// Initialize with empty slice instead of nil
	relationships := []entities.D2Relationship{}

	// Iterate through all edges in the graph
	for _, edge := range graph.Edges {
		if edge == nil || edge.Src == nil || edge.Dst == nil {
			continue
		}

		// Extract source and target IDs
		sourceID := getNodeID(edge.Src)
		targetID := getNodeID(edge.Dst)

		// Skip if we couldn't extract valid IDs
		if sourceID == "" || targetID == "" {
			continue
		}

		// Extract edge label if present
		label := ""
		if edge.Label.Value != "" {
			label = edge.Label.Value
		}

		// Create D2Relationship entity
		rel, err := entities.NewD2Relationship(sourceID, targetID, label)
		if err == nil && rel != nil {
			relationships = append(relationships, *rel)
		}
	}

	return relationships
}

// getNodeID extracts the identifier from a D2 graph node.
// Handles nested shapes by constructing the full path (e.g., "backend.api-server").
func getNodeID(node *d2graph.Object) string {
	if node == nil {
		return ""
	}

	// Build the full path from root to this node
	var parts []string
	current := node

	for current != nil {
		if current.ID != "" {
			parts = append([]string{current.ID}, parts...) // Prepend to maintain order
		}
		current = current.Parent
	}

	// Join with dots for nested shapes (D2 convention)
	fullPath := strings.Join(parts, ".")

	// Clean up: Remove root container prefix if present (diagram root is unnamed)
	fullPath = strings.TrimPrefix(fullPath, ".")

	return fullPath
}
