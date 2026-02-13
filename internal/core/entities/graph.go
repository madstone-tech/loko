// Package entities defines the core domain models for the loko architecture documentation system.
//
// # Thread Safety
//
// ArchitectureGraph is NOT thread-safe by design. It is intended to be:
//   - Built once during initialization (via BuildArchitectureGraph use case)
//   - Read concurrently by multiple consumers (MCP tools, validation, rendering)
//   - Never modified after construction
//
// The GraphCache in internal/mcp provides thread-safe caching using sync.RWMutex.
// For concurrent access patterns, use GraphCache.Get() to retrieve immutable graphs.
//
// # Graph Lifecycle
//
//  1. Construction: BuildArchitectureGraph creates a new graph from project entities
//  2. Population: AddNode() and AddEdge() populate the graph structure
//  3. Freezing: Once built, the graph is treated as immutable
//  4. Caching: GraphCache stores built graphs for reuse across MCP sessions
//  5. Reading: Multiple goroutines can safely read from cached graphs
//
// # Node ID Format (Qualified IDs)
//
// To prevent collisions in multi-system projects, nodes use hierarchical qualified IDs:
//   - System: "systemID" (e.g., "backend")
//   - Container: "systemID/containerID" (e.g., "backend/api")
//   - Component: "systemID/containerID/componentID" (e.g., "backend/api/auth")
//
// The ShortIDMap enables backward compatibility with short ID lookups while
// maintaining uniqueness guarantees for multi-system architectures.
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

	// ShortIDMap maps short IDs to qualified IDs for resolution
	// Enables backward compatibility with short ID references
	// Maps short ID to list of qualified IDs (may have duplicates across systems)
	ShortIDMap map[string][]string

	// IncomingEdges maps target node ID to list of incoming edges (reverse adjacency)
	// Enables O(1) lookup for GetIncomingEdges and GetDependents
	IncomingEdges map[string][]*GraphEdge

	// ChildrenMap maps parent node ID to list of child node IDs
	// Enables O(1) lookup for GetChildren
	ChildrenMap map[string][]string
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
	// Type must be *System, *Container, or *Component (implements C4Entity)
	Data C4Entity

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
		Nodes:         make(map[string]*GraphNode),
		Edges:         make(map[string][]*GraphEdge),
		ParentMap:     make(map[string]string),
		ShortIDMap:    make(map[string][]string),
		IncomingEdges: make(map[string][]*GraphEdge),
		ChildrenMap:   make(map[string][]string),
	}
}

// AddNode adds a node to the graph using qualified IDs for uniqueness.
//
// Node IDs should use the qualified format to prevent collisions in multi-system projects:
//   - System: "systemID" (e.g., "backend")
//   - Container: "systemID/containerID" (e.g., "backend/api")
//   - Component: "systemID/containerID/componentID" (e.g., "backend/api/auth")
//
// The node is also registered in ShortIDMap for backward-compatible short ID lookups.
//
// Example:
//
//	systemNode := &GraphNode{
//	    ID:   QualifiedNodeID("system", "backend", "", ""),
//	    Type: "system",
//	    Name: "Backend Services",
//	}
//	graph.AddNode(systemNode)
//
//	componentNode := &GraphNode{
//	    ID:       QualifiedNodeID("component", "backend", "api", "auth"),
//	    Type:     "component",
//	    Name:     "Authentication",
//	    ParentID: QualifiedNodeID("container", "backend", "api", ""),
//	}
//	graph.AddNode(componentNode)
//
// AddNode returns an error if the node is nil, has an empty ID, or a node with
// the same ID already exists in the graph.
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

		// Maintain ChildrenMap for O(1) lookup
		ag.ChildrenMap[node.ParentID] = append(ag.ChildrenMap[node.ParentID], node.ID)
	}

	// Populate ShortIDMap for ID resolution
	// Extract short ID (last segment) from qualified ID
	parts, _ := ParseQualifiedID(node.ID)
	if len(parts) > 0 {
		shortID := parts[len(parts)-1]
		ag.ShortIDMap[shortID] = append(ag.ShortIDMap[shortID], node.ID)
	}

	return nil
}

// GetNode retrieves a node by ID.
func (ag *ArchitectureGraph) GetNode(id string) *GraphNode {
	return ag.Nodes[id]
}

// ResolveID attempts to resolve a short ID to a fully qualified ID.
// Returns the qualified ID and true if the short ID uniquely identifies a node.
// Returns empty string and false if the short ID is ambiguous or not found.
func (ag *ArchitectureGraph) ResolveID(shortID string) (qualifiedID string, ok bool) {
	qualifiedIDs, exists := ag.ShortIDMap[shortID]
	if !exists {
		return "", false
	}

	// If there's exactly one match, return it
	if len(qualifiedIDs) == 1 {
		return qualifiedIDs[0], true
	}

	// Multiple matches - ambiguous
	return "", false
}

// AddEdge adds a directed edge between two nodes representing a component relationship.
//
// Edges should use qualified node IDs for source and target to ensure correct routing
// in multi-system projects. The graph automatically maintains both outgoing (Edges map)
// and incoming (IncomingEdges map) adjacency lists for O(1) lookups.
//
// Duplicate edges (same source, target, and type) are silently ignored to support
// idempotent graph construction.
//
// Example:
//
//	edge := &GraphEdge{
//	    Source:      QualifiedNodeID("component", "backend", "api", "auth"),
//	    Target:      QualifiedNodeID("component", "backend", "api", "database"),
//	    Type:        "uses",
//	    Description: "Authenticates users against database",
//	}
//	graph.AddEdge(edge)
//
// AddEdge returns an error if the edge is nil, has missing source/target, or references
// nodes that don't exist in the graph.
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

	// Check for duplicate edge
	for _, existing := range ag.Edges[edge.Source] {
		if existing.Target == edge.Target && existing.Type == edge.Type {
			return nil // Already exists, not an error
		}
	}

	// Add forward edge
	ag.Edges[edge.Source] = append(ag.Edges[edge.Source], edge)

	// Maintain IncomingEdges map for O(1) lookup
	ag.IncomingEdges[edge.Target] = append(ag.IncomingEdges[edge.Target], edge)

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

		// Maintain IncomingEdges map for reverse edge
		ag.IncomingEdges[edge.Source] = append(ag.IncomingEdges[edge.Source], reverseEdge)
	}

	return nil
}

// GetOutgoingEdges returns all edges from a source node.
func (ag *ArchitectureGraph) GetOutgoingEdges(nodeID string) []*GraphEdge {
	return ag.Edges[nodeID]
}

// GetIncomingEdges returns all edges pointing to a target node.
// Optimized to O(1) lookup using IncomingEdges map.
func (ag *ArchitectureGraph) GetIncomingEdges(nodeID string) []*GraphEdge {
	return ag.IncomingEdges[nodeID]
}

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

// QualifiedNodeID generates a qualified hierarchical ID for a node.
// - System: returns systemID
// - Container: returns systemID/containerID
// - Component: returns systemID/containerID/componentID
func QualifiedNodeID(nodeType, systemID, containerID, nodeID string) string {
	switch nodeType {
	case "system":
		return systemID
	case "container":
		if systemID == "" || containerID == "" {
			return containerID // fallback for backward compatibility
		}
		return systemID + "/" + containerID
	case "component":
		if systemID == "" || containerID == "" || nodeID == "" {
			return nodeID // fallback for backward compatibility
		}
		return systemID + "/" + containerID + "/" + nodeID
	default:
		return nodeID
	}
}

// ParseQualifiedID parses a qualified ID into its component parts and determines node type.
// Returns the parts slice and the inferred node type.
func ParseQualifiedID(qualifiedID string) (parts []string, nodeType string) {
	if qualifiedID == "" {
		return []string{}, ""
	}

	parts = splitID(qualifiedID)

	switch len(parts) {
	case 1:
		nodeType = "system"
	case 2:
		nodeType = "container"
	case 3:
		nodeType = "component"
	default:
		nodeType = "unknown"
	}

	return parts, nodeType
}

// splitID splits a qualified ID by '/' separator.
func splitID(id string) []string {
	if id == "" {
		return []string{}
	}

	parts := []string{}
	current := ""

	for _, ch := range id {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
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

// RemoveNode removes a node and all its associated edges from the graph.
func (ag *ArchitectureGraph) RemoveNode(nodeID string) error {
	if ag.Nodes[nodeID] == nil {
		return fmt.Errorf("node %q not found", nodeID)
	}

	// Remove from Nodes
	delete(ag.Nodes, nodeID)

	// Remove from ParentMap
	delete(ag.ParentMap, nodeID)

	// Remove from parent's ChildrenMap
	for parentID, children := range ag.ChildrenMap {
		ag.ChildrenMap[parentID] = filterStrings(children, nodeID)
	}

	// Remove from ShortIDMap
	for shortID, qualifiedIDs := range ag.ShortIDMap {
		ag.ShortIDMap[shortID] = filterStrings(qualifiedIDs, nodeID)
		if len(ag.ShortIDMap[shortID]) == 0 {
			delete(ag.ShortIDMap, shortID)
		}
	}

	// Remove all outgoing edges
	delete(ag.Edges, nodeID)

	// Remove all incoming edges
	delete(ag.IncomingEdges, nodeID)

	// Remove edges where this node is the target
	for source := range ag.Edges {
		ag.Edges[source] = filterEdgesByTarget(ag.Edges[source], nodeID)
	}

	// Remove edges where this node is the source (from IncomingEdges)
	for target := range ag.IncomingEdges {
		ag.IncomingEdges[target] = filterEdgesBySource(ag.IncomingEdges[target], nodeID)
	}

	return nil
}

// RemoveEdge removes a specific edge from the graph.
func (ag *ArchitectureGraph) RemoveEdge(source, target, edgeType string) error {
	// Remove from Edges
	var found bool
	newEdges := make([]*GraphEdge, 0)
	for _, edge := range ag.Edges[source] {
		if edge.Target == target && edge.Type == edgeType {
			found = true
			continue
		}
		newEdges = append(newEdges, edge)
	}
	ag.Edges[source] = newEdges

	if !found {
		return fmt.Errorf("edge not found: %s -> %s (%s)", source, target, edgeType)
	}

	// Remove from IncomingEdges
	newIncoming := make([]*GraphEdge, 0)
	for _, edge := range ag.IncomingEdges[target] {
		if edge.Source == source && edge.Type == edgeType {
			continue
		}
		newIncoming = append(newIncoming, edge)
	}
	ag.IncomingEdges[target] = newIncoming

	return nil
}

// filterStrings returns a new slice excluding the specified value.
func filterStrings(slice []string, exclude string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != exclude {
			result = append(result, s)
		}
	}
	return result
}

// filterEdgesByTarget returns edges that don't have the specified target.
func filterEdgesByTarget(edges []*GraphEdge, target string) []*GraphEdge {
	result := make([]*GraphEdge, 0)
	for _, edge := range edges {
		if edge.Target != target {
			result = append(result, edge)
		}
	}
	return result
}

// filterEdgesBySource returns edges that don't have the specified source.
func filterEdgesBySource(edges []*GraphEdge, source string) []*GraphEdge {
	result := make([]*GraphEdge, 0)
	for _, edge := range edges {
		if edge.Source != source {
			result = append(result, edge)
		}
	}
	return result
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
