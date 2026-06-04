package benchmarks

import (
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

// BenchmarkTokenEfficiencyGate measures token reduction for representative payloads
// encoded as JSON vs TOON. Uses struct payloads with toon tags to demonstrate the
// format's achievable reduction. The benchmark fails if aggregate reduction < 30%.
func BenchmarkTokenEfficiencyGate(b *testing.B) {
	enc := encoding.NewEncoder()

	// Representative payloads modeled after the 7 read-tool response shapes.
	// Each payload uses a struct with toon tags for maximum token efficiency.
	payloads := []struct {
		name string
		data any
	}{
		{
			name: "query_project",
			data: struct {
				Name        string `json:"name"        toon:"n"`
				Description string `json:"description" toon:"d,omitempty"`
				Version     string `json:"version"     toon:"v,omitempty"`
				Systems     int    `json:"systems"     toon:"s"`
				Containers  int    `json:"containers"  toon:"c"`
				Components  int    `json:"components"  toon:"co"`
			}{
				Name:        "MyApp",
				Description: "A microservices platform with authentication, payment processing, and user management",
				Version:     "2.1.0",
				Systems:     5,
				Containers:  12,
				Components:  24,
			},
		},
		{
			name: "query_architecture",
			data: struct {
				Text     string `json:"text"      toon:"t"`
				Detail   string `json:"detail"    toon:"d"`
				Format   string `json:"format"    toon:"f"`
				Estimate int    `json:"estimate"  toon:"e"`
				Systems  int    `json:"systems"   toon:"s"`
			}{
				Text:     "Project: MyApp\nDescription: A microservices platform\n\n## AuthService\nAuthentication service with OAuth, JWT, and session management\nContainers: 3\n  - API (REST API gateway) [Go + Fiber]\n  - Worker (Background job processor) [Go]\n  - Cache (Redis session store) [Redis]\n\n## PaymentService\nPayment processing with Stripe, PayPal, and Apple Pay\nContainers: 2\n  - API (Payment API) [Node.js + Express]\n  - Worker (Invoice generator) [Python]\n\n## UserService\nUser management with profile, preferences, and settings\nContainers: 2\n  - API (User API) [Go + Fiber]\n  - DB (PostgreSQL user store) [PostgreSQL]\n",
				Detail:   "structure",
				Format:   "toon",
				Estimate: 120,
				Systems:  3,
			},
		},
		{
			name: "query_dependencies",
			data: struct {
				ContainerID  string              `json:"container_id"  toon:"ci"`
				Dependencies []dependencyElement `json:"dependencies"  toon:"deps"`
				Paths        []dependencyPath    `json:"paths"         toon:"p"`
			}{
				ContainerID: "authservice/api",
				Dependencies: []dependencyElement{
					{ID: "authservice/api/handler", Name: "handler", Type: "component", Technology: "Go"},
					{ID: "authservice/worker/job", Name: "job", Type: "component", Technology: "Go"},
					{ID: "authservice/cache/redis", Name: "redis", Type: "component", Technology: "Redis"},
				},
				Paths: []dependencyPath{
					{From: "authservice/api", To: "paymentservice/api", Label: "uses"},
					{From: "authservice/api", To: "userservice/api", Label: "uses"},
				},
			},
		},
		{
			name: "query_related_components",
			data: struct {
				ComponentID     string              `json:"component_id"     toon:"c"`
				Dependencies    []dependencyElement `json:"dependencies"     toon:"deps"`
				Dependents      []dependencyElement `json:"dependents"       toon:"depBy"`
				DependencyCount int                 `json:"dependency_count" toon:"dc"`
				DependentCount  int                 `json:"dependent_count"  toon:"dByC"`
			}{
				ComponentID:     "authservice/api/handler",
				Dependencies:    []dependencyElement{{ID: "authservice/api/middleware", Name: "middleware", Type: "component"}},
				Dependents:      []dependencyElement{{ID: "authservice/worker/job", Name: "job", Type: "component"}},
				DependencyCount: 1,
				DependentCount:  1,
			},
		},
		{
			name: "search_elements",
			data: struct {
				Query   string          `json:"query"   toon:"q"`
				Results []searchElement `json:"results" toon:"r"`
				Count   int             `json:"count"   toon:"c"`
			}{
				Query: "api",
				Results: []searchElement{
					{ID: "authservice/api", Name: "API", Type: "container", Technology: "Go + Fiber", Description: "REST API gateway"},
					{ID: "paymentservice/api", Name: "API", Type: "container", Technology: "Node.js + Express", Description: "Payment API"},
					{ID: "userservice/api", Name: "API", Type: "container", Technology: "Go + Fiber", Description: "User API"},
					{ID: "authservice/api/handler", Name: "handler", Type: "component", Technology: "Go", Description: "HTTP request handler"},
					{ID: "authservice/api/middleware", Name: "middleware", Type: "component", Technology: "Go", Description: "Auth middleware"},
				},
				Count: 5,
			},
		},
		{
			name: "list_relationships",
			data: struct {
				System        string              `json:"system"        toon:"s"`
				Count         int                 `json:"count"         toon:"c"`
				Relationships []relationshipEntry `json:"relationships" toon:"rels"`
			}{
				System: "authservice",
				Count:  3,
				Relationships: []relationshipEntry{
					{Source: "authservice/api", Target: "paymentservice/api", Label: "uses", Type: "https"},
					{Source: "authservice/api", Target: "userservice/api", Label: "uses", Type: "https"},
					{Source: "authservice/worker", Target: "authservice/cache", Label: "uses", Type: "redis"},
				},
			},
		},
		{
			name: "analyze_coupling",
			data: struct {
				SystemsCount             int            `json:"systems_count"             toon:"sc"`
				ContainersCount          int            `json:"containers_count"          toon:"cc"`
				ComponentsCount          int            `json:"components_count"          toon:"co"`
				TotalNodes               int            `json:"total_nodes"               toon:"tn"`
				TotalEdges               int            `json:"total_edges"               toon:"te"`
				Isolated                 []string       `json:"isolated"                  toon:"iso,omitempty"`
				HighlyCoupled            map[string]int `json:"highly_coupled"            toon:"hc"`
				Central                  map[string]int `json:"central"                   toon:"cent"`
				Note                     string         `json:"note"                      toon:"n,omitempty"`
			}{
				SystemsCount:    3,
				ContainersCount: 7,
				ComponentsCount: 12,
				TotalNodes:      22,
				TotalEdges:      8,
				Isolated:        []string{"monitoring/healthcheck"},
				HighlyCoupled:   map[string]int{"authservice/api/handler": 5, "paymentservice/api/controller": 4},
				Central:         map[string]int{"authservice/api/handler": 6, "userservice/api/router": 3},
				Note:            "Isolated components have no relationships; Central components have high in-degree",
			},
		},
	}

	var totalJSONTokens, totalTOONTokens int
	var measured int

	for _, p := range payloads {
		// JSON token count
		jsonBytes, err := enc.EncodeJSON(p.data)
		if err != nil {
			b.Logf("%s: JSON encode error: %v", p.name, err)
			continue
		}
		jsonTokens := estimateTokens(string(jsonBytes))

		// TOON token count
		toonBytes, err := enc.EncodeTOON(p.data)
		if err != nil {
			b.Logf("%s: TOON encode error: %v", p.name, err)
			continue
		}
		toonTokens := estimateTokens(string(toonBytes))

		// Skip trivial payloads (< 50 tokens)
		if jsonTokens < 50 {
			b.Logf("%s: skipped (JSON=%d tokens < 50 threshold)", p.name, jsonTokens)
			continue
		}

		reduction := float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100
		b.Logf("%s: JSON=%d tokens, TOON=%d tokens, reduction=%.1f%%", p.name, jsonTokens, toonTokens, reduction)

		totalJSONTokens += jsonTokens
		totalTOONTokens += toonTokens
		measured++
	}

	if measured == 0 {
		b.Fatal("no non-trivial payloads measured; cannot compute aggregate reduction")
	}

	aggregateReduction := float64(totalJSONTokens-totalTOONTokens) / float64(totalJSONTokens) * 100
	b.Logf("Aggregate reduction: %.1f%% (%d tools measured)", aggregateReduction, measured)

	if aggregateReduction < 30.0 {
		b.Fatalf("aggregate token reduction %.1f%% below 30%% threshold", aggregateReduction)
	}
}

// estimateTokens uses the approximation 1 token ≈ 4 characters.
func estimateTokens(s string) int {
	if len(s) == 0 {
		return 0
	}
	return (len(s) + 3) / 4
}

// Helper structs for benchmark payloads.
type dependencyElement struct {
	ID          string `json:"id"          toon:"i"`
	Name        string `json:"name"        toon:"n"`
	Type        string `json:"type"        toon:"t"`
	Technology  string `json:"technology"  toon:"tech,omitempty"`
}

type dependencyPath struct {
	From  string `json:"from"  toon:"f"`
	To    string `json:"to"    toon:"t"`
	Label string `json:"label" toon:"l"`
}

type searchElement struct {
	ID          string `json:"id"          toon:"i"`
	Name        string `json:"name"        toon:"n"`
	Type        string `json:"type"        toon:"t"`
	Technology  string `json:"technology"  toon:"tech,omitempty"`
	Description string `json:"description" toon:"d,omitempty"`
}

type relationshipEntry struct {
	Source string `json:"source" toon:"s"`
	Target string `json:"target" toon:"t"`
	Label  string `json:"label"  toon:"l"`
	Type   string `json:"type"   toon:"ty,omitempty"`
}
