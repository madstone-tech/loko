package cmd

import (
	"slices"
	"testing"
)

// TestMCPToolSet pins the tools loko mcp serves.
func TestMCPToolSet(t *testing.T) {
	t.Parallel()
	var got []string
	for _, tool := range newMCPTools(t.TempDir()) {
		got = append(got, tool.Name())
	}
	slices.Sort(got)
	if want := []string{"apply_edit", "build", "describe", "fmt", "move", "query", "schema", "validate"}; !slices.Equal(got, want) {
		t.Errorf("tools = %v, want %v", got, want)
	}
}
