package entities

import (
	"testing"
)

// TestArchitectureGraphBasics tests basic graph operations.
func TestArchitectureGraphBasics(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	system := &GraphNode{
		ID:          "ecommerce",
		Type:        "system",
		Name:        "E-Commerce System",
		Description: "Online shopping platform",
		Level:       1,
	}

	container := &GraphNode{
		ID:          "api-server",
		Type:        "container",
		Name:        "API Server",
		Description: "REST API",
		Level:       2,
		ParentID:    "ecommerce",
	}

	component := &GraphNode{
		ID:          "auth",
		Type:        "component",
		Name:        "Authentication",
		Description: "User auth handler",
		Level:       3,
		ParentID:    "api-server",
	}

	// Add nodes
	if err := graph.AddNode(system); err != nil {
		t.Fatalf("failed to add system: %v", err)
	}

	if err := graph.AddNode(container); err != nil {
		t.Fatalf("failed to add container: %v", err)
	}

	if err := graph.AddNode(component); err != nil {
		t.Fatalf("failed to add component: %v", err)
	}

	// Verify size
	if graph.Size() != 3 {
		t.Errorf("expected size 3, got %d", graph.Size())
	}

	// Retrieve node
	retrieved := graph.GetNode("auth")
	if retrieved == nil {
		t.Fatal("failed to retrieve node")
	}

	if retrieved.Name != "Authentication" {
		t.Errorf("expected name 'Authentication', got %q", retrieved.Name)
	}
}

// TestHierarchyNavigation tests parent-child relationships.
func TestHierarchyNavigation(t *testing.T) {
	graph := NewArchitectureGraph()

	system := &GraphNode{ID: "ecom", Type: "system", Name: "E-Commerce", Level: 1}
	container := &GraphNode{ID: "api", Type: "container", Name: "API", Level: 2, ParentID: "ecom"}
	comp1 := &GraphNode{ID: "auth", Type: "component", Name: "Auth", Level: 3, ParentID: "api"}
	comp2 := &GraphNode{ID: "payment", Type: "component", Name: "Payment", Level: 3, ParentID: "api"}

	graph.AddNode(system)
	graph.AddNode(container)
	graph.AddNode(comp1)
	graph.AddNode(comp2)

	// Test GetParent
	parent := graph.GetParent("auth")
	if parent == nil || parent.ID != "api" {
		t.Error("failed to get parent")
	}

	// Test GetChildren
	children := graph.GetChildren("api")
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}

	// Test GetAncestors
	ancestors := graph.GetAncestors("auth")
	if len(ancestors) != 2 {
		t.Errorf("expected 2 ancestors, got %d", len(ancestors))
	}

	// Test GetDescendants
	descendants := graph.GetDescendants("ecom")
	if len(descendants) != 3 {
		t.Errorf("expected 3 descendants, got %d", len(descendants))
	}
}

// TestEdgesAndDependencies tests relationship edges.
func TestEdgesAndDependencies(t *testing.T) {
	graph := NewArchitectureGraph()

	comp1 := &GraphNode{ID: "auth", Type: "component", Name: "Auth"}
	comp2 := &GraphNode{ID: "db", Type: "component", Name: "Database"}
	comp3 := &GraphNode{ID: "cache", Type: "component", Name: "Cache"}

	graph.AddNode(comp1)
	graph.AddNode(comp2)
	graph.AddNode(comp3)

	// Add edges
	edge1 := &GraphEdge{
		Source:      "auth",
		Target:      "db",
		Type:        "uses",
		Description: "Auth queries user data",
		Weight:      0.8,
	}

	edge2 := &GraphEdge{
		Source:        "auth",
		Target:        "cache",
		Type:          "uses",
		Bidirectional: true,
		Weight:        0.5,
	}

	if err := graph.AddEdge(edge1); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	if err := graph.AddEdge(edge2); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	// Test GetDependencies
	deps := graph.GetDependencies("auth")
	if len(deps) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(deps))
	}

	// Test GetDependents (should have cache as dependent due to bidirectional)
	dependents := graph.GetDependents("cache")
	if len(dependents) != 1 {
		t.Errorf("expected 1 dependent, got %d", len(dependents))
	}

	// Test outgoing edges
	outgoing := graph.GetOutgoingEdges("auth")
	if len(outgoing) != 2 {
		t.Errorf("expected 2 outgoing edges, got %d", len(outgoing))
	}

	// Test incoming edges
	incoming := graph.GetIncomingEdges("db")
	if len(incoming) != 1 {
		t.Errorf("expected 1 incoming edge, got %d", len(incoming))
	}
}

// TestPathFinding tests BFS path finding.
func TestPathFinding(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create a chain: A -> B -> C -> D
	nodeA := &GraphNode{ID: "a", Type: "component", Name: "A"}
	nodeB := &GraphNode{ID: "b", Type: "component", Name: "B"}
	nodeC := &GraphNode{ID: "c", Type: "component", Name: "C"}
	nodeD := &GraphNode{ID: "d", Type: "component", Name: "D"}

	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)

	graph.AddEdge(&GraphEdge{Source: "a", Target: "b", Type: "uses"})
	graph.AddEdge(&GraphEdge{Source: "b", Target: "c", Type: "uses"})
	graph.AddEdge(&GraphEdge{Source: "c", Target: "d", Type: "uses"})

	// Test path finding
	path := graph.GetPath("a", "d")
	if len(path) != 4 {
		t.Errorf("expected path of length 4, got %d", len(path))
	}

	if path[0].ID != "a" || path[len(path)-1].ID != "d" {
		t.Error("path endpoints incorrect")
	}

	// Test non-existent path
	path = graph.GetPath("d", "a")
	if path != nil {
		t.Error("expected no path from d to a")
	}

	// Test IsConnected
	if !graph.IsConnected("a", "d") {
		t.Error("expected a to be connected to d")
	}

	if graph.IsConnected("d", "a") {
		t.Error("expected d to not be connected to a")
	}
}

// TestLevelFiltering tests getting nodes by level.
func TestLevelFiltering(t *testing.T) {
	graph := NewArchitectureGraph()

	nodes := []*GraphNode{
		{ID: "sys1", Type: "system", Level: 1},
		{ID: "sys2", Type: "system", Level: 1},
		{ID: "cont1", Type: "container", Level: 2},
		{ID: "cont2", Type: "container", Level: 2},
		{ID: "comp1", Type: "component", Level: 3},
		{ID: "comp2", Type: "component", Level: 3},
		{ID: "comp3", Type: "component", Level: 3},
	}

	for _, node := range nodes {
		graph.AddNode(node)
	}

	// Test level filtering
	level1 := graph.GetNodesByLevel(1)
	if len(level1) != 2 {
		t.Errorf("expected 2 level-1 nodes, got %d", len(level1))
	}

	level2 := graph.GetNodesByLevel(2)
	if len(level2) != 2 {
		t.Errorf("expected 2 level-2 nodes, got %d", len(level2))
	}

	level3 := graph.GetNodesByLevel(3)
	if len(level3) != 3 {
		t.Errorf("expected 3 level-3 nodes, got %d", len(level3))
	}

	// Test type filtering
	systems := graph.GetNodesByType("system")
	if len(systems) != 2 {
		t.Errorf("expected 2 systems, got %d", len(systems))
	}

	components := graph.GetNodesByType("component")
	if len(components) != 3 {
		t.Errorf("expected 3 components, got %d", len(components))
	}
}

// TestGraphValidation tests graph validation.
func TestGraphValidation(t *testing.T) {
	graph := NewArchitectureGraph()

	node1 := &GraphNode{ID: "node1", Type: "system"}
	node2 := &GraphNode{ID: "node2", Type: "container", ParentID: "node1"}

	graph.AddNode(node1)
	graph.AddNode(node2)

	// Should be valid
	if err := graph.Validate(); err != nil {
		t.Fatalf("valid graph failed validation: %v", err)
	}

	// Create invalid graph with missing parent
	invalidGraph := NewArchitectureGraph()
	invalidNode := &GraphNode{ID: "orphan", Type: "component", ParentID: "nonexistent"}
	invalidGraph.AddNode(invalidNode)

	if err := invalidGraph.Validate(); err == nil {
		t.Error("expected validation to fail for missing parent")
	}
}

// TestEdgeCount tests edge counting.
func TestEdgeCount(t *testing.T) {
	graph := NewArchitectureGraph()

	node1 := &GraphNode{ID: "node1", Type: "component"}
	node2 := &GraphNode{ID: "node2", Type: "component"}
	node3 := &GraphNode{ID: "node3", Type: "component"}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Add unidirectional edge
	graph.AddEdge(&GraphEdge{Source: "node1", Target: "node2", Type: "uses"})
	if graph.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", graph.EdgeCount())
	}

	// Add bidirectional edge
	graph.AddEdge(&GraphEdge{Source: "node2", Target: "node3", Type: "uses", Bidirectional: true})
	if graph.EdgeCount() != 3 {
		t.Errorf("expected 3 edges (1 unidirectional + 2 bidirectional), got %d", graph.EdgeCount())
	}
}
