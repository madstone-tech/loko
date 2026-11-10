package tools

import (
	"encoding/json"
	"slices"
	"testing"
)

// TestSchemasMatchContract pins each tool's required arguments and properties
// to contracts/mcp-tools.md.
func TestSchemasMatchContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		schema   map[string]any
		required []string
		props    []string
	}{
		"describe":   {DescribeSchema, []string{}, []string{"address", "format", "level", "ranges"}},
		"query":      {QuerySchema, []string{"kind"}, []string{"address", "element_kind", "format", "kind", "limit", "ranges", "to", "transitive"}},
		"validate":   {ValidateSchema, []string{}, []string{"format", "ignore", "ranges", "severity", "summary"}},
		"apply_edit": {ApplyEditSchema, []string{"base_revision", "edits"}, []string{"base_revision", "diagnostics", "diff", "edits", "format", "preview", "ranges"}},
		"move":       {MoveSchema, []string{"from", "to", "base_revision"}, []string{"base_revision", "diagnostics", "diff", "format", "from", "preview", "ranges", "to"}},
		"schema":     {SchemaSchema, []string{}, []string{"block", "format"}},
		"fmt":        {FmtSchema, []string{"base_revision"}, []string{"base_revision", "diagnostics", "format", "preview", "ranges"}},
		"build":      {BuildSchema, []string{}, []string{"format", "formats", "ignore", "inline", "inline_format", "out", "ranges", "strict"}},
	}
	for name, tt := range tests {
		if _, err := json.Marshal(tt.schema); err != nil {
			t.Errorf("%s: schema does not marshal: %v", name, err)
		}
		if got := tt.schema["required"].([]string); !slices.Equal(got, tt.required) {
			t.Errorf("%s: required = %v, want %v", name, got, tt.required)
		}
		var props []string
		for k := range tt.schema["properties"].(map[string]any) {
			props = append(props, k)
		}
		slices.Sort(props)
		if !slices.Equal(props, tt.props) {
			t.Errorf("%s: properties = %v, want %v", name, props, tt.props)
		}
	}
}

// TestWriteResultSchemas is contract W-1: the diagnostics and diff inputs,
// their values and their defaults.
func TestWriteResultSchemas(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		schema map[string]any
		diff   bool
	}{
		"apply_edit": {ApplyEditSchema, true},
		"move":       {MoveSchema, true},
		"fmt":        {FmtSchema, false},
	}
	check := func(name string, prop any, enum []string, def string) {
		p, ok := prop.(map[string]any)
		if !ok {
			t.Errorf("%s: missing", name)
			return
		}
		if got, _ := p["enum"].([]string); !slices.Equal(got, enum) || p["default"] != def {
			t.Errorf("%s: enum %v default %v, want %v %s", name, p["enum"], p["default"], enum, def)
		}
	}
	for name, tt := range tests {
		props := tt.schema["properties"].(map[string]any)
		check(name+".diagnostics", props["diagnostics"], []string{"changed", "all"}, "changed")
		if tt.diff {
			check(name+".diff", props["diff"], []string{"meaning", "exact"}, "meaning")
		} else if _, ok := props["diff"]; ok {
			t.Errorf("%s: has no diff input", name)
		}
	}
}
