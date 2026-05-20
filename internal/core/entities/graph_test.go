package entities

import (
	"fmt"
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

// TestIncomingEdgesInitialization tests IncomingEdges map is initialized.
func TestIncomingEdgesInitialization(t *testing.T) {
	graph := NewArchitectureGraph()

	if graph.IncomingEdges == nil {
		t.Error("IncomingEdges map should be initialized")
	}

	if len(graph.IncomingEdges) != 0 {
		t.Errorf("IncomingEdges should be empty for new graph, got %d entries", len(graph.IncomingEdges))
	}
}

// TestChildrenMapInitialization tests ChildrenMap is initialized.
func TestChildrenMapInitialization(t *testing.T) {
	graph := NewArchitectureGraph()

	if graph.ChildrenMap == nil {
		t.Error("ChildrenMap should be initialized")
	}

	if len(graph.ChildrenMap) != 0 {
		t.Errorf("ChildrenMap should be empty for new graph, got %d entries", len(graph.ChildrenMap))
	}
}

// TestIncomingEdgesSingleEdge tests IncomingEdges is populated for single edge.
func TestIncomingEdgesSingleEdge(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}

	graph.AddNode(node1)
	graph.AddNode(node2)

	// Add edge from node1 to node2
	edge := &GraphEdge{
		Source:      "node1",
		Target:      "node2",
		Type:        "depends-on",
		Description: "test dependency",
	}

	if err := graph.AddEdge(edge); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	// Verify IncomingEdges for node2
	incoming := graph.IncomingEdges["node2"]
	if len(incoming) != 1 {
		t.Errorf("expected 1 incoming edge for node2, got %d", len(incoming))
	}

	if len(incoming) > 0 && incoming[0].Source != "node1" {
		t.Errorf("incoming edge source should be node1, got %q", incoming[0].Source)
	}
}

// TestIncomingEdgesBidirectional tests IncomingEdges for bidirectional edges.
func TestIncomingEdgesBidirectional(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}

	graph.AddNode(node1)
	graph.AddNode(node2)

	// Add bidirectional edge
	edge := &GraphEdge{
		Source:        "node1",
		Target:        "node2",
		Type:          "depends-on",
		Description:   "bidirectional dependency",
		Bidirectional: true,
	}

	if err := graph.AddEdge(edge); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	// Verify IncomingEdges for node2 (forward edge)
	incoming2 := graph.IncomingEdges["node2"]
	if len(incoming2) != 1 {
		t.Errorf("expected 1 incoming edge for node2, got %d", len(incoming2))
	}

	// Verify IncomingEdges for node1 (reverse edge)
	incoming1 := graph.IncomingEdges["node1"]
	if len(incoming1) != 1 {
		t.Errorf("expected 1 incoming edge for node1 (reverse), got %d", len(incoming1))
	}
}

// TestIncomingEdgesMultipleToSameTarget tests multiple edges to same target.
func TestIncomingEdgesMultipleToSameTarget(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}
	node3 := &GraphNode{ID: "node3", Type: "component", Name: "Node 3"}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Add edges from node1 and node2 to node3
	edge1 := &GraphEdge{Source: "node1", Target: "node3", Type: "depends-on"}
	edge2 := &GraphEdge{Source: "node2", Target: "node3", Type: "depends-on"}

	graph.AddEdge(edge1)
	graph.AddEdge(edge2)

	// Verify IncomingEdges for node3
	incoming := graph.IncomingEdges["node3"]
	if len(incoming) != 2 {
		t.Errorf("expected 2 incoming edges for node3, got %d", len(incoming))
	}

	// Verify sources
	sources := make(map[string]bool)
	for _, e := range incoming {
		sources[e.Source] = true
	}

	if !sources["node1"] || !sources["node2"] {
		t.Error("incoming edges should include both node1 and node2 as sources")
	}
}

// TestChildrenMapWithParent tests ChildrenMap is populated when node has parent.
func TestChildrenMapWithParent(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create parent and child nodes
	parent := &GraphNode{ID: "parent", Type: "container", Name: "Parent"}
	child := &GraphNode{ID: "child", Type: "component", Name: "Child", ParentID: "parent"}

	graph.AddNode(parent)
	graph.AddNode(child)

	// Verify ChildrenMap for parent
	children := graph.ChildrenMap["parent"]
	if len(children) != 1 {
		t.Errorf("expected 1 child for parent, got %d", len(children))
	}

	if len(children) > 0 && children[0] != "child" {
		t.Errorf("child ID should be 'child', got %q", children[0])
	}
}

// TestChildrenMapWithoutParent tests ChildrenMap when node has no parent.
func TestChildrenMapWithoutParent(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create node without parent
	node := &GraphNode{ID: "root", Type: "system", Name: "Root"}

	graph.AddNode(node)

	// Verify ChildrenMap is empty for root
	if len(graph.ChildrenMap) != 0 {
		t.Errorf("ChildrenMap should be empty when only root node exists, got %d entries", len(graph.ChildrenMap))
	}
}

// BenchmarkGetIncomingEdges benchmarks GetIncomingEdges performance.
func BenchmarkGetIncomingEdges(b *testing.B) {
	// Create graph with 200 components and 500 edges
	graph := NewArchitectureGraph()

	// Create 200 components
	for i := 0; i < 200; i++ {
		node := &GraphNode{
			ID:   QualifiedNodeID("component", "system", "container", fmt.Sprintf("comp%d", i)),
			Type: "component",
			Name: fmt.Sprintf("Component %d", i),
		}
		graph.AddNode(node)
	}

	// Create 500 edges (each component depends on ~2-3 others)
	for i := 0; i < 500; i++ {
		source := QualifiedNodeID("component", "system", "container", fmt.Sprintf("comp%d", i%200))
		target := QualifiedNodeID("component", "system", "container", fmt.Sprintf("comp%d", (i+1)%200))

		edge := &GraphEdge{
			Source: source,
			Target: target,
			Type:   "depends-on",
		}
		graph.AddEdge(edge)
	}

	// Benchmark GetIncomingEdges
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		targetNode := QualifiedNodeID("component", "system", "container", fmt.Sprintf("comp%d", i%200))
		_ = graph.GetIncomingEdges(targetNode)
	}
}

// BenchmarkGetChildren benchmarks GetChildren performance.
func BenchmarkGetChildren(b *testing.B) {
	// Create graph with 5-level hierarchy
	graph := NewArchitectureGraph()

	// Create 1 system
	system := &GraphNode{ID: "system", Type: "system", Name: "System"}
	graph.AddNode(system)

	// Create 5 containers under system
	for i := 0; i < 5; i++ {
		container := &GraphNode{
			ID:       QualifiedNodeID("container", "system", fmt.Sprintf("cont%d", i), ""),
			Type:     "container",
			Name:     fmt.Sprintf("Container %d", i),
			ParentID: "system",
		}
		graph.AddNode(container)

		// Create 10 components under each container
		for j := 0; j < 10; j++ {
			component := &GraphNode{
				ID:       QualifiedNodeID("component", "system", fmt.Sprintf("cont%d", i), fmt.Sprintf("comp%d", j)),
				Type:     "component",
				Name:     fmt.Sprintf("Component %d", j),
				ParentID: QualifiedNodeID("container", "system", fmt.Sprintf("cont%d", i), ""),
			}
			graph.AddNode(component)
		}
	}

	// Benchmark GetChildren on system node
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = graph.GetChildren("system")
	}
}

// TestDuplicateEdgePrevention tests that adding identical edge twice results in single edge.
func TestDuplicateEdgePrevention(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}

	graph.AddNode(node1)
	graph.AddNode(node2)

	// Add edge
	edge := &GraphEdge{
		Source:      "node1",
		Target:      "node2",
		Type:        "depends-on",
		Description: "test dependency",
	}

	if err := graph.AddEdge(edge); err != nil {
		t.Fatalf("failed to add edge first time: %v", err)
	}

	// Add same edge again (duplicate)
	duplicateEdge := &GraphEdge{
		Source:      "node1",
		Target:      "node2",
		Type:        "depends-on",
		Description: "test dependency",
	}

	if err := graph.AddEdge(duplicateEdge); err != nil {
		t.Errorf("adding duplicate edge should not return error, got: %v", err)
	}

	// Verify only one edge exists
	outgoing := graph.GetOutgoingEdges("node1")
	if len(outgoing) != 1 {
		t.Errorf("expected 1 edge after duplicate add, got %d", len(outgoing))
	}
}

// TestEdgeCountAccuracy tests EdgeCount() returns correct count after duplicate attempts.
func TestEdgeCountAccuracy(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}
	node3 := &GraphNode{ID: "node3", Type: "component", Name: "Node 3"}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Add 2 unique edges
	edge1 := &GraphEdge{Source: "node1", Target: "node2", Type: "depends-on"}
	edge2 := &GraphEdge{Source: "node1", Target: "node3", Type: "depends-on"}

	graph.AddEdge(edge1)
	graph.AddEdge(edge2)

	// Try to add edge1 again (duplicate)
	graph.AddEdge(edge1)

	// EdgeCount should be 2, not 3
	if count := graph.EdgeCount(); count != 2 {
		t.Errorf("expected EdgeCount = 2 after duplicate add, got %d", count)
	}
}

// TestRemoveNodeCleanup tests RemoveNode cleans up all related data structures.
func TestRemoveNodeCleanup(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes with parent-child relationship
	parent := &GraphNode{ID: "parent", Type: "container", Name: "Parent"}
	child := &GraphNode{ID: "child", Type: "component", Name: "Child", ParentID: "parent"}
	other := &GraphNode{ID: "other", Type: "component", Name: "Other"}

	graph.AddNode(parent)
	graph.AddNode(child)
	graph.AddNode(other)

	// Add edges
	edge1 := &GraphEdge{Source: "parent", Target: "child", Type: "contains"}
	edge2 := &GraphEdge{Source: "child", Target: "other", Type: "depends-on"}
	edge3 := &GraphEdge{Source: "other", Target: "child", Type: "depends-on"}

	graph.AddEdge(edge1)
	graph.AddEdge(edge2)
	graph.AddEdge(edge3)

	// Remove child node
	if err := graph.RemoveNode("child"); err != nil {
		t.Fatalf("failed to remove node: %v", err)
	}

	// Verify node is removed from Nodes
	if graph.GetNode("child") != nil {
		t.Error("node should be removed from Nodes map")
	}

	// Verify ParentMap is cleaned up
	if _, exists := graph.ParentMap["child"]; exists {
		t.Error("node should be removed from ParentMap")
	}

	// Verify ChildrenMap is cleaned up
	children := graph.ChildrenMap["parent"]
	for _, c := range children {
		if c == "child" {
			t.Error("node should be removed from parent's ChildrenMap")
		}
	}

	// Verify edges are removed
	outgoing := graph.GetOutgoingEdges("child")
	if len(outgoing) != 0 {
		t.Errorf("outgoing edges should be removed, got %d", len(outgoing))
	}

	incoming := graph.GetIncomingEdges("child")
	if len(incoming) != 0 {
		t.Errorf("incoming edges should be removed, got %d", len(incoming))
	}

	// Verify ShortIDMap is cleaned up
	if qualifiedIDs, exists := graph.ShortIDMap["child"]; exists {
		for _, qid := range qualifiedIDs {
			if qid == "child" {
				t.Error("node should be removed from ShortIDMap")
			}
		}
	}
}

// TestRemoveEdgeSpecificity tests RemoveEdge only removes the specified edge.
func TestRemoveEdgeSpecificity(t *testing.T) {
	graph := NewArchitectureGraph()

	// Create nodes
	node1 := &GraphNode{ID: "node1", Type: "component", Name: "Node 1"}
	node2 := &GraphNode{ID: "node2", Type: "component", Name: "Node 2"}

	graph.AddNode(node1)
	graph.AddNode(node2)

	// Add two edges with different types between same nodes
	edge1 := &GraphEdge{Source: "node1", Target: "node2", Type: "depends-on"}
	edge2 := &GraphEdge{Source: "node1", Target: "node2", Type: "uses"}

	graph.AddEdge(edge1)
	graph.AddEdge(edge2)

	// Remove only the "depends-on" edge
	if err := graph.RemoveEdge("node1", "node2", "depends-on"); err != nil {
		t.Fatalf("failed to remove edge: %v", err)
	}

	// Verify "depends-on" edge is removed
	outgoing := graph.GetOutgoingEdges("node1")
	if len(outgoing) != 1 {
		t.Errorf("expected 1 remaining edge, got %d", len(outgoing))
	}

	if len(outgoing) > 0 && outgoing[0].Type != "uses" {
		t.Errorf("wrong edge removed, expected 'uses', got %q", outgoing[0].Type)
	}

	// Verify IncomingEdges is also updated
	incoming := graph.GetIncomingEdges("node2")
	if len(incoming) != 1 {
		t.Errorf("expected 1 incoming edge, got %d", len(incoming))
	}

	if len(incoming) > 0 && incoming[0].Type != "uses" {
		t.Errorf("wrong incoming edge, expected 'uses', got %q", incoming[0].Type)
	}
}
