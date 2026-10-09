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
		"describe":   {DescribeSchema, []string{}, []string{"address", "format", "level"}},
		"query":      {QuerySchema, []string{"kind"}, []string{"address", "format", "kind", "limit", "to", "transitive"}},
		"validate":   {ValidateSchema, []string{}, []string{"format"}},
		"apply_edit": {ApplyEditSchema, []string{"base_revision", "edits"}, []string{"base_revision", "edits", "format", "preview"}},
		"move":       {MoveSchema, []string{"from", "to", "base_revision"}, []string{"base_revision", "format", "from", "preview", "to"}},
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
