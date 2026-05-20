// Package entities — traversal and relationship-walking methods for ArchitectureGraph.
// See graph.go for type definitions and core mutation/lookup primitives.
package entities

// GetParent returns the parent node of a given node.
func (ag *ArchitectureGraph) GetParent(nodeID string) *GraphNode {
	if parentID, exists := ag.ParentMap[nodeID]; exists {
		return ag.Nodes[parentID]
	}
	return nil
}

// GetChildren returns all child nodes of a given node.
// Optimized to O(1) lookup using ChildrenMap.
func (ag *ArchitectureGraph) GetChildren(nodeID string) []*GraphNode {
	var children []*GraphNode
	for _, childID := range ag.ChildrenMap[nodeID] {
		if node := ag.Nodes[childID]; node != nil {
			children = append(children, node)
		}
	}
	return children
}

// GetAncestors returns all ancestor nodes (path to root).
func (ag *ArchitectureGraph) GetAncestors(nodeID string) []*GraphNode {
	var ancestors []*GraphNode
	current := nodeID

	for {
		parent := ag.GetParent(current)
		if parent == nil {
			break
		}
		ancestors = append(ancestors, parent)
		current = parent.ID
	}

	return ancestors
}

// GetDescendants returns all descendant nodes (recursive).
func (ag *ArchitectureGraph) GetDescendants(nodeID string) []*GraphNode {
	var descendants []*GraphNode
	children := ag.GetChildren(nodeID)

	for _, child := range children {
		descendants = append(descendants, child)
		descendants = append(descendants, ag.GetDescendants(child.ID)...)
	}

	return descendants
}

// GetDependencies returns all nodes that this node depends on (outgoing edges).
//
// For components, this represents the "uses" or "depends-on" relationships declared
// in the component's Relationships map. Systems and containers typically have no
// dependencies, as relationships are modeled at the component level (C4 Level 3).
//
// The nodeID parameter should be a qualified ID. Short IDs can be resolved using
// ResolveID() first if the short ID is unambiguous.
//
// Example:
//
//	// Get dependencies of backend/api/auth component
//	authID := QualifiedNodeID("component", "backend", "api", "auth")
//	deps := graph.GetDependencies(authID)
//	for _, dep := range deps {
//	    fmt.Printf("auth depends on: %s (%s)\n", dep.Name, dep.ID)
//	}
//	// Output might be:
//	// auth depends on: Database (backend/api/database)
//	// auth depends on: Cache (backend/api/cache)
//
// Returns an empty slice if the node has no dependencies or doesn't exist.
func (ag *ArchitectureGraph) GetDependencies(nodeID string) []*GraphNode {
	var deps []*GraphNode
	edges := ag.GetOutgoingEdges(nodeID)

	for _, edge := range edges {
		if node := ag.Nodes[edge.Target]; node != nil {
			deps = append(deps, node)
		}
	}

	return deps
}

// GetDependents returns all nodes that depend on this node (incoming edges).
func (ag *ArchitectureGraph) GetDependents(nodeID string) []*GraphNode {
	var dependents []*GraphNode
	edges := ag.GetIncomingEdges(nodeID)

	for _, edge := range edges {
		if node := ag.Nodes[edge.Source]; node != nil {
			dependents = append(dependents, node)
		}
	}

	return dependents
}

// GetPath finds a path from source to target using BFS.
func (ag *ArchitectureGraph) GetPath(source, target string) []*GraphNode {
	if ag.Nodes[source] == nil || ag.Nodes[target] == nil {
		return nil
	}

	// BFS implementation
	visited := make(map[string]bool)
	queue := [][]*GraphNode{{ag.Nodes[source]}}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]

		if current.ID == target {
			return path
		}

		if visited[current.ID] {
			continue
		}

		visited[current.ID] = true

		// Explore neighbors via outgoing edges
		for _, edge := range ag.GetOutgoingEdges(current.ID) {
			neighbor := ag.Nodes[edge.Target]
			if neighbor != nil && !visited[neighbor.ID] {
				newPath := make([]*GraphNode, len(path))
				copy(newPath, path)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}

	return nil
}
