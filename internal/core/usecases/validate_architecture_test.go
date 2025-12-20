package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestValidateArchitectureNilGraph(t *testing.T) {
	uc := NewValidateArchitecture()

	report := uc.Execute(nil, nil)

	if report.IsValid {
		t.Error("Expected invalid report for nil graph")
	}
	if len(report.Issues) == 0 {
		t.Error("Expected at least one issue for nil graph")
	}
	if report.Issues[0].Code != "invalid_graph" {
		t.Errorf("Expected invalid_graph code, got %s", report.Issues[0].Code)
	}
}

func TestValidateArchitectureValidArchitecture(t *testing.T) {
	uc := NewValidateArchitecture()

	// Create a valid architecture
	sys, _ := entities.NewSystem("PaymentService")
	graph := entities.NewArchitectureGraph()

	// Add nodes
	authNode := &entities.GraphNode{ID: "auth", Name: "Auth", Type: "component", Level: 3}
	dbNode := &entities.GraphNode{ID: "database", Name: "Database", Type: "component", Level: 3}
	cacheNode := &entities.GraphNode{ID: "cache", Name: "Cache", Type: "component", Level: 3}

	graph.Nodes["auth"] = authNode
	graph.Nodes["database"] = dbNode
	graph.Nodes["cache"] = cacheNode

	// Add edges (auth -> database, auth -> cache) - linear dependency
	graph.Edges["auth"] = []*entities.GraphEdge{
		{Source: "auth", Target: "database"},
		{Source: "auth", Target: "cache"},
	}

	report := uc.Execute(graph, []*entities.System{sys})

	if !report.IsValid {
		t.Errorf("Expected valid report, got invalid with issues: %+v", report.Issues)
	}
	if len(report.Issues) == 0 {
		t.Log("✓ Valid architecture - no issues found")
	}
}

func TestValidateArchitectureCircularDependency(t *testing.T) {
	uc := NewValidateArchitecture()

	sys, _ := entities.NewSystem("Service")
	graph := entities.NewArchitectureGraph()

	// Create circular dependency: A -> B -> C -> A
	aNode := &entities.GraphNode{ID: "comp-a", Name: "Component A", Type: "component", Level: 3}
	bNode := &entities.GraphNode{ID: "comp-b", Name: "Component B", Type: "component", Level: 3}
	cNode := &entities.GraphNode{ID: "comp-c", Name: "Component C", Type: "component", Level: 3}

	graph.Nodes["comp-a"] = aNode
	graph.Nodes["comp-b"] = bNode
	graph.Nodes["comp-c"] = cNode

	graph.Edges["comp-a"] = []*entities.GraphEdge{{Source: "comp-a", Target: "comp-b"}}
	graph.Edges["comp-b"] = []*entities.GraphEdge{{Source: "comp-b", Target: "comp-c"}}
	graph.Edges["comp-c"] = []*entities.GraphEdge{{Source: "comp-c", Target: "comp-a"}}

	report := uc.Execute(graph, []*entities.System{sys})

	if report.IsValid {
		t.Error("Expected invalid report for circular dependency")
	}

	circularIssues := report.GetIssuesByCode("circular_dependency")
	if len(circularIssues) == 0 {
		t.Error("Expected circular_dependency issue")
	}
	if report.Errors == 0 {
		t.Error("Expected error count to be > 0")
	}
}

func TestValidateArchitectureIsolatedComponent(t *testing.T) {
	uc := NewValidateArchitecture()

	sys, _ := entities.NewSystem("Service")
	graph := entities.NewArchitectureGraph()

	// Create two isolated components
	comp1 := &entities.GraphNode{ID: "isolated-1", Name: "Isolated 1", Type: "component", Level: 3}
	comp2 := &entities.GraphNode{ID: "isolated-2", Name: "Isolated 2", Type: "component", Level: 3}
	comp3 := &entities.GraphNode{ID: "connected", Name: "Connected", Type: "component", Level: 3}

	graph.Nodes["isolated-1"] = comp1
	graph.Nodes["isolated-2"] = comp2
	graph.Nodes["connected"] = comp3

	// No edges - all isolated

	report := uc.Execute(graph, []*entities.System{sys})

	isolatedIssues := report.GetIssuesByCode("isolated_component")
	if len(isolatedIssues) == 0 {
		t.Error("Expected isolated_component issue")
	}
	if !containsID(isolatedIssues[0].Affected, "isolated-1") {
		t.Error("Expected isolated-1 in affected list")
	}
	if !containsID(isolatedIssues[0].Affected, "isolated-2") {
		t.Error("Expected isolated-2 in affected list")
	}
}

func TestValidateArchitectureHighCoupling(t *testing.T) {
	uc := NewValidateArchitecture()

	sys, _ := entities.NewSystem("Service")
	graph := entities.NewArchitectureGraph()

	// Create a hub component with many dependencies
	hubNode := &entities.GraphNode{ID: "hub", Name: "Hub", Type: "component", Level: 3}
	graph.Nodes["hub"] = hubNode

	// Add 6 components that hub depends on
	hubEdges := make([]*entities.GraphEdge, 0)
	for i := 1; i <= 6; i++ {
		id := string(rune('a' + i - 1))
		compID := "comp-" + id
		node := &entities.GraphNode{ID: compID, Name: "Component " + id, Type: "component", Level: 3}
		graph.Nodes[compID] = node
		hubEdges = append(hubEdges, &entities.GraphEdge{Source: "hub", Target: compID})
	}
	graph.Edges["hub"] = hubEdges

	report := uc.Execute(graph, []*entities.System{sys})

	couplingIssues := report.GetIssuesByCode("high_coupling")
	if len(couplingIssues) == 0 {
		t.Error("Expected high_coupling issue")
	}
	if !containsID(couplingIssues[0].Affected, "hub") {
		t.Error("Expected hub in affected list")
	}
	if report.Warnings == 0 {
		t.Error("Expected warning count to be > 0")
	}
}

func TestValidateArchitectureDanglingReference(t *testing.T) {
	uc := NewValidateArchitecture()

	// Create system with component that references non-existent component
	sys, _ := entities.NewSystem("Service")
	cont, _ := entities.NewContainer("API")
	comp, _ := entities.NewComponent("Service A")

	// Add relationship to non-existent component
	comp.AddRelationship("non-existent-component", "calls")

	cont.AddComponent(comp)
	sys.AddContainer(cont)

	// Graph only has comp-a, not the referenced "non-existent-component"
	graph := entities.NewArchitectureGraph()
	graph.Nodes["service-a"] = &entities.GraphNode{ID: "service-a", Name: "Service A", Type: "component", Level: 3}

	report := uc.Execute(graph, []*entities.System{sys})

	danglingIssues := report.GetIssuesByCode("dangling_reference")
	if len(danglingIssues) == 0 {
		t.Error("Expected dangling_reference issue")
	}
	if report.Errors == 0 {
		t.Error("Expected error count to be > 0")
	}
}

func TestValidateArchitectureReportSummary(t *testing.T) {
	uc := NewValidateArchitecture()

	sys, _ := entities.NewSystem("Service")
	graph := entities.NewArchitectureGraph()

	// Create valid graph
	graph.Nodes["comp-1"] = &entities.GraphNode{ID: "comp-1", Name: "Component 1", Type: "component", Level: 3}

	report := uc.Execute(graph, []*entities.System{sys})

	if report.Summary == "" {
		t.Error("Expected non-empty summary")
	}
	if !report.IsValid {
		t.Errorf("Expected valid architecture, got summary: %s", report.Summary)
	}
}

func TestValidateArchitectureGetIssuesBySeverity(t *testing.T) {
	report := &ArchitectureReport{
		Issues: []ArchitectureIssue{
			{Severity: "error", Code: "test1"},
			{Severity: "warning", Code: "test2"},
			{Severity: "error", Code: "test3"},
			{Severity: "info", Code: "test4"},
		},
	}

	errors := report.GetIssuesBySeverity("error")
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	warnings := report.GetIssuesBySeverity("warning")
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	infos := report.GetIssuesBySeverity("info")
	if len(infos) != 1 {
		t.Errorf("Expected 1 info, got %d", len(infos))
	}
}

// Helper function to check if a string slice contains a value
func containsID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
