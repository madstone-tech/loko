package encoding

import (
	"encoding/json"
	"testing"
)

func TestEncoderJSON(t *testing.T) {
	enc := NewEncoder()

	t.Run("encode simple struct", func(t *testing.T) {
		data := struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}{
			Name:  "test",
			Count: 42,
		}

		result, err := enc.EncodeJSON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := `{"name":"test","count":42}`
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("decode JSON", func(t *testing.T) {
		input := `{"name":"decoded","count":100}`
		var result struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		err := enc.DecodeJSON([]byte(input), &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Name != "decoded" || result.Count != 100 {
			t.Errorf("unexpected result: %+v", result)
		}
	})
}

func TestEncoderTOON(t *testing.T) {
	enc := NewEncoder()

	t.Run("encode simple struct", func(t *testing.T) {
		data := struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Count       int    `json:"count"`
		}{
			Name:        "PaymentService",
			Description: "Handles payments",
			Count:       5,
		}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// TOON should be shorter than JSON
		jsonResult, _ := enc.EncodeJSON(data)
		if len(result) >= len(jsonResult) {
			t.Errorf("TOON should be shorter: TOON=%d, JSON=%d", len(result), len(jsonResult))
		}

		t.Logf("TOON: %s", string(result))
		t.Logf("JSON: %s", string(jsonResult))
	})

	t.Run("encode with abbreviated keys", func(t *testing.T) {
		data := map[string]any{
			"name":        "Test",
			"description": "A test system",
			"technology":  "Go",
		}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultStr := string(result)
		// Should use abbreviated keys
		if !contains(resultStr, "n:") || !contains(resultStr, "d:") || !contains(resultStr, "t:") {
			t.Errorf("expected abbreviated keys, got: %s", resultStr)
		}
	})

	t.Run("encode array", func(t *testing.T) {
		data := []string{"one", "two", "three"}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should use semicolons
		resultStr := string(result)
		if resultStr != "[one;two;three]" {
			t.Errorf("expected [one;two;three], got: %s", resultStr)
		}
	})

	t.Run("encode boolean", func(t *testing.T) {
		data := map[string]bool{"active": true, "disabled": false}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultStr := string(result)
		// Should use T/F for booleans
		if !contains(resultStr, "T") || !contains(resultStr, "F") {
			t.Errorf("expected T/F for booleans, got: %s", resultStr)
		}
	})

	t.Run("encode nested structure", func(t *testing.T) {
		data := map[string]any{
			"systems": []map[string]any{
				{"name": "Auth", "containers": 3},
				{"name": "API", "containers": 2},
			},
		}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		jsonResult, _ := json.Marshal(data)
		t.Logf("TOON (%d bytes): %s", len(result), string(result))
		t.Logf("JSON (%d bytes): %s", len(jsonResult), string(jsonResult))

		// TOON should be more compact
		if len(result) >= len(jsonResult) {
			t.Errorf("TOON should be shorter than JSON")
		}
	})
}

func TestFormatArchitectureTOON(t *testing.T) {
	summary := ArchitectureSummary{
		Name:        "MyProject",
		Description: "A sample project",
		Systems:     3,
		Containers:  8,
		Components:  24,
		SystemNames: []string{"Auth", "API", "Database"},
	}

	result := FormatArchitectureTOON(summary)

	// Should be very compact
	if len(result) > 100 {
		t.Errorf("summary TOON should be <100 chars, got %d: %s", len(result), result)
	}

	// Should contain key information
	if !contains(result, "MyProject") {
		t.Error("should contain project name")
	}
	if !contains(result, "S3") {
		t.Error("should contain system count")
	}

	t.Logf("Summary TOON (%d chars): %s", len(result), result)
}

func TestFormatStructureTOON(t *testing.T) {
	structure := ArchitectureStructure{
		Name: "MyProject",
		Systems: []SystemCompact{
			{
				ID:          "auth",
				Name:        "Auth",
				Description: "Authentication service",
				Containers: []ContainerBrief{
					{ID: "api", Name: "API", Technology: "Go"},
					{ID: "db", Name: "Database", Technology: "PostgreSQL"},
				},
			},
			{
				ID:   "web",
				Name: "Web",
				Containers: []ContainerBrief{
					{ID: "frontend", Name: "Frontend", Technology: "React"},
				},
			},
		},
	}

	result := FormatStructureTOON(structure)

	// Should be compact but readable
	t.Logf("Structure TOON (%d chars):\n%s", len(result), result)

	// Should contain systems
	if !contains(result, "Auth") || !contains(result, "Web") {
		t.Error("should contain system names")
	}

	// Should contain containers with technology
	if !contains(result, "[Go]") || !contains(result, "[React]") {
		t.Error("should contain technology in brackets")
	}
}

func TestTOONTokenEfficiency(t *testing.T) {
	// Test with realistic architecture data
	data := map[string]any{
		"name":        "E-Commerce Platform",
		"description": "Multi-service e-commerce system",
		"version":     "1.0.0",
		"systems": []map[string]any{
			{
				"name":        "Payment Service",
				"description": "Handles payment processing",
				"technology":  "Go + gRPC",
				"containers":  []string{"API", "Worker", "Database"},
			},
			{
				"name":        "User Service",
				"description": "User management and auth",
				"technology":  "Node.js",
				"containers":  []string{"API", "Cache", "Database"},
			},
			{
				"name":        "Order Service",
				"description": "Order processing",
				"technology":  "Python",
				"containers":  []string{"API", "Queue", "Database"},
			},
		},
	}

	enc := NewEncoder()

	jsonResult, _ := enc.EncodeJSON(data)
	toonResult, _ := enc.EncodeTOON(data)

	jsonLen := len(jsonResult)
	toonLen := len(toonResult)

	savings := float64(jsonLen-toonLen) / float64(jsonLen) * 100

	t.Logf("JSON: %d bytes", jsonLen)
	t.Logf("TOON: %d bytes", toonLen)
	t.Logf("Savings: %.1f%%", savings)

	// Target: 30-40% reduction
	if savings < 20 {
		t.Errorf("expected at least 20%% savings, got %.1f%%", savings)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
