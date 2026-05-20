// Package usecases — relationship-edge construction helpers for BuildArchitectureGraph.
// See build_architecture_graph.go for the use case type and Execute method.
package usecases

import (
	"context"
	"os"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// resolveToComponentIDs resolves a TOML element path to one or more component-level
// qualified node IDs. The resolution strategy is:
//
//  1. If the path directly matches a component node (3 segments), return it.
//  2. If the path matches a container node (2 segments), return all child component IDs.
//     This is the "container fan-out" that makes TOML relationships at the container
//     level visible to component-level coupling analysis.
//  3. If direct lookup fails, attempt short-ID resolution via the graph's ShortIDMap,
//     then re-apply the container fan-out if the resolved node is a container.
//  4. Returns nil if the path cannot be resolved (skipped gracefully).
func resolveToComponentIDs(path string, graph *entities.ArchitectureGraph) []string {
	// Direct lookup first — handles both component and container paths.
	if node := graph.GetNode(path); node != nil {
		return expandNodeToComponents(node.ID, graph)
	}

	// Short-ID fallback (e.g. caller stored just "api-lambda" without system prefix).
	if qid, ok := graph.ResolveID(path); ok {
		return expandNodeToComponents(qid, graph)
	}

	return nil // not found — caller skips gracefully
}

// expandNodeToComponents returns the component-level node IDs reachable from nodeID.
// If nodeID is a component node, returns [nodeID].
// If nodeID is a container node, returns all immediate child component node IDs.
// For any other level (system), returns nil — system-level fan-out is too broad.
func expandNodeToComponents(nodeID string, graph *entities.ArchitectureGraph) []string {
	node := graph.GetNode(nodeID)
	if node == nil {
		return nil
	}

	switch node.Level {
	case 3: // component — use directly
		return []string{nodeID}
	case 2: // container — fan out to children
		children := graph.ChildrenMap[nodeID]
		if len(children) == 0 {
			// Container exists but has no components yet; return the container ID
			// so the edge is at least recorded at container level.
			return []string{nodeID}
		}
		return children
	default:
		return nil
	}
}

// parseComponentD2 reads the D2 diagram file for a component (if present) and
// returns the relationships defined there. Returns nil, nil when no D2 file exists.
func (uc *BuildArchitectureGraph) parseComponentD2(ctx context.Context, componentPath string) ([]entities.D2Relationship, error) {
	// Look for any .d2 file inside the component directory
	entries, err := os.ReadDir(componentPath)
	if err != nil {
		// Directory not accessible — treat as no D2 file (graceful degradation)
		return nil, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".d2") {
			d2Path := componentPath + "/" + name
			data, err := os.ReadFile(d2Path)
			if err != nil {
				return nil, err
			}
			return uc.d2Parser.ParseRelationships(ctx, string(data))
		}
	}

	return nil, nil // no D2 file found — valid state
}
