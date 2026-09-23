package d2_test

import (
	"testing"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2lib"
)

// TestD2LibraryImport verifies D2 library integration (T003)
func TestD2LibraryImport(t *testing.T) {
	// Verify we can import and use basic D2 types
	var _ *d2graph.Graph
	var _ = d2lib.Parse

	t.Log("D2 library imports successful")
}
