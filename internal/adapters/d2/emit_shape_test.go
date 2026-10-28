package d2

import (
	"context"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestShapesAreCards: a declared shape is drawn as a rounded card with the
// shape's icon (FR-001, R13), and stays in the D2 source as a class (R12).
func TestShapesAreCards(t *testing.T) {
	t.Parallel()
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range arch.Shapes {
		n := titledNode("")
		n.Style = vm.WithShape(n.Style, shape)
		src, plan := emitPlanned(vm.ViewModel{View: vm.View{ID: "v", Direction: "down"}, Nodes: []vm.Node{n}})
		s := string(src)
		for _, want := range []string{"shape: rectangle\n", "border-radius: 6\n",
			"class: [kind-container; shape-" + shape + "]\n", "  shape-" + shape + ": {}\n"} {
			if !strings.Contains(s, want) {
				t.Errorf("%s: emitted D2 lacks %q\n%s", shape, want, s)
			}
		}
		for _, old := range []string{"c4-person", "cylinder", "shape: queue", "hexagon", "shape: step", "stored_data"} {
			if strings.Contains(s, old) {
				t.Errorf("%s: emitted D2 still draws %q", shape, old)
			}
		}
		if got, want := plan.Nodes["container__api"].Icon, shapeIcons[vm.Shape(shape)]; got != want {
			t.Errorf("%s: icon = %q, want %q", shape, got, want)
		}
		if _, _, err := compile(context.Background(), ruler, src, false); err != nil {
			t.Errorf("%s does not render: %v", shape, err)
		}
	}
	plain := string(Emit(vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{titledNode("")}}))
	if strings.Contains(plain, "shape-") {
		t.Errorf("an unshaped container has a shape class:\n%s", plain)
	}
}
