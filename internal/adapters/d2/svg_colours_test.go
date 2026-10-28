package d2

import (
	"context"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// d2DefaultEdge is d2 theme 0's edge colour (B1), which no edge may keep.
const d2DefaultEdge = "#0D32B2"

// TestSVGColours is FR-009: edges and their labels use the brand neutrals,
// and the brand's green accent never appears.
func TestSVGColours(t *testing.T) {
	t.Setenv("PATH", "")
	c := emitCases()
	views := []vm.ViewModel{c["landscape"], c["subject"], c["deployment"]}
	as, err := NewSVGBackend().Render(context.Background(), projectionOf(views...), usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range as {
		doc := string(a.Bytes)
		if strings.Contains(strings.ToUpper(doc), "#18C800") {
			t.Errorf("%s uses the brand accent", a.Path)
		}
		ids := idIndex(views[i].Nodes)
		seen := map[[2]string]int{}
		for _, e := range views[i].Edges {
			pair := [2]string{e.Source, e.Target}
			g := groupOf(doc, edgeID(ids[e.Source], ids[e.Target], seen[pair]))
			seen[pair]++
			if g == "" {
				t.Errorf("%s: no group for edge %s", a.Path, e.ID)
				continue
			}
			if !strings.Contains(g, `stroke="`+vm.EdgePalette.Stroke+`"`) || strings.Contains(g, d2DefaultEdge) {
				t.Errorf("%s: edge %s stroke is not %s: %.300s", a.Path, e.ID, vm.EdgePalette.Stroke, g)
			}
			if !strings.Contains(firstTag(g, "text"), `fill="`+vm.EdgePalette.Font+`"`) {
				t.Errorf("%s: edge %s label is not %s: %s", a.Path, e.ID, vm.EdgePalette.Font, firstTag(g, "text"))
			}
		}
	}
}
