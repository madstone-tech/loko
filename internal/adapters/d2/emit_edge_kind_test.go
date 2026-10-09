package d2

import (
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestEdgeKindsAndTags(t *testing.T) {
	t.Parallel()
	a, b := titledNode(""), titledNode("")
	b.ID, b.Address = "container__b", "container.b"
	m := vm.ViewModel{View: vm.View{ID: "v", Direction: "down"}, Nodes: []vm.Node{a, b, {ID: vm.OutsideNodeID, Role: vm.RoleOutside, Label: "outside this view"}},
		Edges: []vm.Edge{
			{ID: "e1", Source: a.ID, Target: b.ID, Label: "Publishes", Tags: []string{"audit", "write"}, Style: vm.EdgeStyle{Async: true}},
			{ID: "e2", Source: a.ID, Target: vm.OutsideNodeID, Label: "Leaves", Crossing: true, Style: vm.EdgeStyle{Dashed: true}},
			{ID: "e3", Source: b.ID, Target: a.ID, Label: "Calls"},
		}}
	out := string(Emit(m))
	if !strings.Contains(out, `"container__api" -> "container__b": "Publishes\n#audit #write" {`+"\n  style.stroke-dash: 8\n}") {
		t.Errorf("async edge with tags:\n%s", out)
	}
	if !strings.Contains(out, `"container__api" -> "outside": "Leaves" {`+"\n  style.stroke-dash: 4\n}") {
		t.Errorf("crossing edge keeps dash 4:\n%s", out)
	}
	if strings.Contains(out, `"container__b" -> "container__api": "Calls" {`) {
		t.Errorf("a plain edge has no style block:\n%s", out)
	}
}
