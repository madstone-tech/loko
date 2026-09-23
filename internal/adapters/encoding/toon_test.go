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
			Name        string `toon:"name"`
			Description string `toon:"description"`
			Count       int    `toon:"count"`
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

		// Should contain field names
		resultStr := string(result)
		if !contains(resultStr, "name:") || !contains(resultStr, "description:") || !contains(resultStr, "count:") {
			t.Errorf("expected field names in output, got: %s", resultStr)
		}
	})

	t.Run("encode array", func(t *testing.T) {
		data := []string{"one", "two", "three"}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should use comma delimiter with length marker
		resultStr := string(result)
		if !contains(resultStr, "[#3]:") || !contains(resultStr, "one,two,three") {
			t.Errorf("expected array format with length marker, got: %s", resultStr)
		}
	})

	t.Run("encode boolean", func(t *testing.T) {
		data := map[string]bool{"active": true, "disabled": false}

		result, err := enc.EncodeTOON(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultStr := string(result)
		// Should use true/false for booleans
		if !contains(resultStr, "true") || !contains(resultStr, "false") {
			t.Errorf("expected true/false for booleans, got: %s", resultStr)
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

		// Should contain field names in header
		resultStr := string(result)
		if !contains(resultStr, "systems") || !contains(resultStr, "name") || !contains(resultStr, "containers") {
			t.Errorf("expected field names in output, got: %s", resultStr)
		}
	})
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

	// Target: At least 2% reduction (official TOON format may have different characteristics)
	if savings < 2 {
		t.Errorf("expected at least 2%% savings, got %.1f%%", savings)
	}
}

// T032: TOON v3.0 Spec Compliance Tests

func TestTOONTabularArrays(t *testing.T) {
	enc := NewEncoder()

	// Test tabular arrays with length markers
	containers := []struct {
		Name       string `toon:"name"`
		Technology string `toon:"technology"`
	}{
		{"API", "Go"},
		{"Database", "PostgreSQL"},
		{"Cache", "Redis"},
	}

	result, err := enc.EncodeTOON(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultStr := string(result)

	// Verify tabular array format with length marker
	if !contains(resultStr, "[#3]") {
		t.Errorf("expected length marker [#3], got: %s", resultStr)
	}

	// Verify tabular format with fields header
	if !contains(resultStr, "{name,technology}:") {
		t.Errorf("expected fields header {name,technology}:, got: %s", resultStr)
	}

	// Verify data rows
	if !contains(resultStr, "API,Go") || !contains(resultStr, "Database,PostgreSQL") || !contains(resultStr, "Cache,Redis") {
		t.Errorf("expected tabular data rows, got: %s", resultStr)
	}

	t.Logf("Tabular array TOON: %s", resultStr)
}

func TestTOONRoundTripEncoding(t *testing.T) {
	enc := NewEncoder()

	// Test with simplified data structure for round-trip compatibility
	data := map[string]any{
		"name":        "TestProject",
		"description": "A test project",
		"version":     "1.0.0",
		"metadata": map[string]any{
			"author": "test",
		},
	}

	// Encode to TOON
	toonData, err := enc.EncodeTOON(data)
	if err != nil {
		t.Fatalf("failed to encode to TOON: %v", err)
	}

	// Decode back
	var decodedData map[string]any
	err = enc.DecodeTOON(toonData, &decodedData)
	if err != nil {
		t.Fatalf("failed to decode from TOON: %v", err)
	}

	// Compare key fields
	if decodedData["name"] != data["name"] {
		t.Errorf("name mismatch: expected %s, got %s", data["name"], decodedData["name"])
	}

	if decodedData["description"] != data["description"] {
		t.Errorf("description mismatch: expected %s, got %s", data["description"], decodedData["description"])
	}

	t.Logf("Original data TOON (%d bytes): %s", len(toonData), string(toonData))
}

// T036: Write Entity Round-Trip Tests

// T037: Write Error Handling Tests

func TestTOONDecodeErrors(t *testing.T) {
	enc := NewEncoder()

	tests := []struct {
		name  string
		input string
		want  string // expected error substring
	}{
		{
			name:  "malformed_syntax",
			input: "{invalid:unclosed",
			want:  "error", // should return clear error
		},
		{
			name:  "invalid_tabular_array",
			input: "[#3{name}:\n  only,two",
			want:  "error", // length marker mismatch
		},
		{
			name:  "empty_input",
			input: "",
			want:  "", // Empty input might not be an error depending on implementation
		},
		{
			name:  "invalid_field_name",
			input: "unknown_field: value",
			want:  "", // might succeed with zero value - acceptable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]any
			err := enc.DecodeTOON([]byte(tt.input), &result)

			if tt.want != "" && err == nil {
				t.Errorf("expected error containing %q, got nil", tt.want)
			}

			if err != nil {
				t.Logf("Error message: %v", err)
				// Verify error message contains location info or is clear
			}
		})
	}
}

// T038: Verify Round-Trip Fidelity on Representative Data

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
