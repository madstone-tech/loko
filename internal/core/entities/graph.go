package entities

import (
	"fmt"
)

// ArchitectureGraph represents the C4 model as a directed graph.
// Nodes are entities (Systems, Containers, Components).
// Edges represent relationships and hierarchy.
type ArchitectureGraph struct {
	// Nodes maps entity ID to its node representation
	Nodes map[string]*GraphNode

	// Edges maps source node ID to list of target node IDs with relationship info
	Edges map[string][]*GraphEdge

	// Hierarchy tracks parent-child relationships
	// Maps child ID to parent ID
	ParentMap map[string]string
}

// GraphNode represents a C4 entity as a node in the graph.
type GraphNode struct {
	// ID is the unique identifier (normalized name)
	ID string

	// Type is the C4 level (System, Container, Component)
	Type string // "system", "container", "component"

	// Name is the display name
	Name string

	// Description explains the node
	Description string

	// Level indicates C4 level (1, 2, 3)
	Level int

	// ParentID is the ID of the parent node (if any)
	ParentID string

	// Data holds reference to the actual entity
	// Type could be *System, *Container, or *Component
	Data any

	// Metadata for additional properties
	Metadata map[string]string
}

// GraphEdge represents a directed relationship between two nodes.
type GraphEdge struct {
	// Source node ID
	Source string

	// Target node ID
	Target string

	// Type of relationship (e.g., "uses", "depends-on", "implements")
	Type string

	// Description of the relationship
	Description string

	// Bidirectional indicates if this is a two-way relationship
	Bidirectional bool

	// Weight for weighted graph algorithms (e.g., coupling strength)
	Weight float64

	// Metadata for additional properties
	Metadata map[string]string
}

// NewArchitectureGraph creates a new empty architecture graph.
func NewArchitectureGraph() *ArchitectureGraph {
	return &ArchitectureGraph{
		Nodes:     make(map[string]*GraphNode),
		Edges:     make(map[string][]*GraphEdge),
		ParentMap: make(map[string]string),
	}
}

// AddNode adds a node to the graph.
func (ag *ArchitectureGraph) AddNode(node *GraphNode) error {
	if node == nil || node.ID == "" {
		return fmt.Errorf("node cannot be nil and must have an ID")
	}

	if _, exists := ag.Nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %q already exists", node.ID)
	}

	ag.Nodes[node.ID] = node

	// Track parent relationship if specified
	if node.ParentID != "" {
		ag.ParentMap[node.ID] = node.ParentID
	}

	return nil
}

// GetNode retrieves a node by ID.
func (ag *ArchitectureGraph) GetNode(id string) *GraphNode {
	return ag.Nodes[id]
}

// AddEdge adds a directed edge between two nodes.
func (ag *ArchitectureGraph) AddEdge(edge *GraphEdge) error {
	if edge == nil || edge.Source == "" || edge.Target == "" {
		return fmt.Errorf("edge must have source and target")
	}

	// Verify nodes exist
	if ag.Nodes[edge.Source] == nil {
		return fmt.Errorf("source node %q not found", edge.Source)
	}

	if ag.Nodes[edge.Target] == nil {
		return fmt.Errorf("target node %q not found", edge.Target)
	}

	// Add forward edge
	ag.Edges[edge.Source] = append(ag.Edges[edge.Source], edge)

	// Add reverse edge if bidirectional
	if edge.Bidirectional {
		reverseEdge := &GraphEdge{
			Source:        edge.Target,
			Target:        edge.Source,
			Type:          edge.Type,
			Description:   edge.Description,
			Bidirectional: false, // Prevent infinite recursion
			Weight:        edge.Weight,
			Metadata:      edge.Metadata,
		}
		ag.Edges[edge.Target] = append(ag.Edges[edge.Target], reverseEdge)
	}

	return nil
}

// GetOutgoingEdges returns all edges from a source node.
func (ag *ArchitectureGraph) GetOutgoingEdges(nodeID string) []*GraphEdge {
	return ag.Edges[nodeID]
}

// GetIncomingEdges returns all edges pointing to a target node.
func (ag *ArchitectureGraph) GetIncomingEdges(nodeID string) []*GraphEdge {
	var incoming []*GraphEdge
	for _, edges := range ag.Edges {
		for _, edge := range edges {
			if edge.Target == nodeID {
				incoming = append(incoming, edge)
			}
		}
	}
	return incoming
}

// GetParent returns the parent node of a given node.
func (ag *ArchitectureGraph) GetParent(nodeID string) *GraphNode {
	if parentID, exists := ag.ParentMap[nodeID]; exists {
		return ag.Nodes[parentID]
	}
	return nil
}

// GetChildren returns all child nodes of a given node.
func (ag *ArchitectureGraph) GetChildren(nodeID string) []*GraphNode {
	var children []*GraphNode
	for childID, parentID := range ag.ParentMap {
		if parentID == nodeID {
			if node := ag.Nodes[childID]; node != nil {
				children = append(children, node)
			}
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

// GetNodesByLevel returns all nodes at a specific C4 level.
func (ag *ArchitectureGraph) GetNodesByLevel(level int) []*GraphNode {
	var nodes []*GraphNode
	for _, node := range ag.Nodes {
		if node.Level == level {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodesByType returns all nodes of a specific type.
func (ag *ArchitectureGraph) GetNodesByType(nodeType string) []*GraphNode {
	var nodes []*GraphNode
	for _, node := range ag.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Size returns the number of nodes in the graph.
func (ag *ArchitectureGraph) Size() int {
	return len(ag.Nodes)
}

// EdgeCount returns the total number of edges in the graph.
func (ag *ArchitectureGraph) EdgeCount() int {
	count := 0
	for _, edges := range ag.Edges {
		count += len(edges)
	}
	return count
}

// IsConnected checks if there's a path between two nodes.
func (ag *ArchitectureGraph) IsConnected(source, target string) bool {
	return ag.GetPath(source, target) != nil
}

// Validate checks the integrity of the graph.
// Returns error if graph has inconsistencies.
func (ag *ArchitectureGraph) Validate() error {
	// Check that all edges reference existing nodes
	for source, edges := range ag.Edges {
		if ag.Nodes[source] == nil {
			return fmt.Errorf("edge source %q not found in nodes", source)
		}

		for _, edge := range edges {
			if ag.Nodes[edge.Target] == nil {
				return fmt.Errorf("edge target %q not found in nodes", edge.Target)
			}
		}
	}

	// Check that parent references exist
	for childID, parentID := range ag.ParentMap {
		if ag.Nodes[childID] == nil {
			return fmt.Errorf("child node %q not found", childID)
		}

		if ag.Nodes[parentID] == nil {
			return fmt.Errorf("parent node %q not found for child %q", parentID, childID)
		}
	}

	return nil
}
