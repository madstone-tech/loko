package d2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestEmitDynamicGolden pins the D2 of the dynamic-views fixture's flows in
// both styles, from its committed projection (feature 019).
func TestEmitDynamicGolden(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "projects", "dynamic-views", "expected", "projection.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p vm.Projection
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	for _, id := range []vm.ViewID{"checkout", "async_order", "checkout_sequence"} {
		v, ok := p.View(id)
		if !ok {
			t.Fatalf("no view %s", id)
		}
		v.View.TitleDeclared = true // not encoded; every flow here declares a title
		golden(t, filepath.Join("testdata", "golden", "dynamic_"+string(id)+".d2"), Emit(v))
	}
}
