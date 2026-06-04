package tools

import (
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

func TestGetFormat(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "empty defaults to toon",
			args: map[string]any{},
			want: "toon",
		},
		{
			name: "empty string defaults to toon",
			args: map[string]any{"format": ""},
			want: "toon",
		},
		{
			name: "explicit toon",
			args: map[string]any{"format": "toon"},
			want: "toon",
		},
		{
			name: "explicit json",
			args: map[string]any{"format": "json"},
			want: "json",
		},
		{
			name:    "invalid format",
			args:    map[string]any{"format": "xml"},
			wantErr: true,
		},
		{
			name:    "invalid format with suggestion",
			args:    map[string]any{"format": "compact"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFormat(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("getFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatResponse(t *testing.T) {
	encoder := encoding.NewEncoder()

	tests := []struct {
		name    string
		data    map[string]any
		format  string
		wantErr bool
		check   func(t *testing.T, got any)
	}{
		{
			name:   "json returns map directly",
			data:   map[string]any{"key": "value"},
			format: "json",
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", got)
				}
				if m["key"] != "value" {
					t.Fatalf("expected key=value, got %v", m["key"])
				}
			},
		},
		{
			name:   "toon returns wrapper",
			data:   map[string]any{"key": "value"},
			format: "toon",
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", got)
				}
				if m["format"] != "toon" {
					t.Fatalf("expected format=toon, got %v", m["format"])
				}
				payload, ok := m["payload"].(string)
				if !ok || payload == "" {
					t.Fatalf("expected non-empty payload, got %v", m["payload"])
				}
				estimate, ok := m["token_estimate"].(int)
				if !ok || estimate <= 0 {
					t.Fatalf("expected positive token_estimate, got %v", m["token_estimate"])
				}
			},
		},
		{
			name:   "toon handles nested maps",
			data:   map[string]any{"nested": map[string]any{"a": 1}},
			format: "toon",
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", got)
				}
				if m["format"] != "toon" {
					t.Fatalf("expected format=toon")
				}
			},
		},
		{
			name:    "invalid format error",
			data:    map[string]any{"key": "value"},
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatResponse(tt.data, tt.format, encoder)
			if (err != nil) != tt.wantErr {
				t.Fatalf("formatResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{"abcdefghijklmnopqrstuvwxyz", 7}, // 26/4 = 6.5 -> 7
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := estimateTokenCount(tt.input)
			if got != tt.want {
				t.Fatalf("estimateTokenCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
