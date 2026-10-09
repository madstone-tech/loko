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
	dagre := string(Emit(vm.ViewModel{View: vm.View{ID: "v", Direction: "down"}, Nodes: nodes}))
	if strings.Contains(dagre, "layout-engine") {
		t.Errorf("dagre view names an engine:\n%s", dagre)
	}
	src := Emit(vm.ViewModel{View: vm.View{ID: "v", Direction: "down", Layout: "elk"}, Nodes: nodes})
	if !strings.Contains(string(src), "d2-config: {\n    layout-engine: elk\n  }") {
		t.Fatalf("elk view:\n%s", src)
	}
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	svg, err := compile(context.Background(), ruler, src)
	if err != nil || !strings.Contains(string(svg), "<svg") {
		t.Fatalf("elk compile: %v", err)
	}
	if _, err := layoutEngine("neato"); err == nil {
		t.Error("unknown engine accepted")
	}
}
