package d2

import (
	"context"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestLayoutEngine(t *testing.T) {
	t.Parallel()
	nodes := []vm.Node{{ID: "a", Label: "a"}, {ID: "b", Label: "b"}}
	for layout, want := range map[string]string{"": "elk", "elk": "elk", "dagre": "dagre"} {
		got := string(Emit(vm.ViewModel{View: vm.View{ID: "v", Direction: "down", Layout: layout}, Nodes: nodes}))
		if !strings.Contains(got, "d2-config: {\n    layout-engine: "+want+"\n  }") {
			t.Errorf("layout %q: want engine %s:\n%s", layout, want, got)
		}
	}
	seq := string(Emit(vm.ViewModel{View: vm.View{ID: "v", Layout: "elk", Style: vm.StyleSequence}, Nodes: nodes}))
	if strings.Contains(seq, "layout-engine") {
		t.Errorf("sequence view names an engine:\n%s", seq)
	}
	src := Emit(vm.ViewModel{View: vm.View{ID: "v", Direction: "down", Layout: "elk"}, Nodes: nodes})
	if !strings.Contains(string(src), "d2-config: {\n    layout-engine: elk\n  }") {
		t.Fatalf("elk view:\n%s", src)
	}
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	svg, _, err := compile(context.Background(), ruler, src, false)
	if err != nil || !strings.Contains(string(svg), "<svg") {
		t.Fatalf("elk compile: %v", err)
	}
	if _, err := layoutEngine("neato"); err == nil {
		t.Error("unknown engine accepted")
	}
}
