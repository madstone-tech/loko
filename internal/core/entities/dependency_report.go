package entities

// DependencyReport contains analysis of dependency patterns in the architecture graph.
// Replaces the previous map[string]any return type with a strongly-typed struct.
type DependencyReport struct {
	// Node counts by C4 level
	SystemsCount    int `json:"systems_count"`
	ContainersCount int `json:"containers_count"`
	ComponentsCount int `json:"components_count"`
	TotalNodes      int `json:"total_nodes"`
	TotalEdges      int `json:"total_edges"`

	// Isolated components have no incoming or outgoing dependencies
	IsolatedComponents []string `json:"isolated_components"`

	// Highly coupled components depend on many other components (>2 dependencies)
	// Maps component ID to dependency count
	HighlyCoupledComponents map[string]int `json:"highly_coupled_components"`

	// Central components have many dependents (>2 components depend on them)
	// Maps component ID to dependent count
	CentralComponents map[string]int `json:"central_components"`
}

// NewDependencyReport creates an empty dependency report with initialized maps.
func NewDependencyReport() *DependencyReport {
	return &DependencyReport{
		IsolatedComponents:      []string{},
		HighlyCoupledComponents: make(map[string]int),
		CentralComponents:       make(map[string]int),
	}
}
