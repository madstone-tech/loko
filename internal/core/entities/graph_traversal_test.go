package entities

import "testing"

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
