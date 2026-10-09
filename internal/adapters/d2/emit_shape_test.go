package d2

import (
	"context"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestShapesMapToD2(t *testing.T) {
	t.Parallel()
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	for shape, d2 := range map[string]string{"database": "cylinder", "queue": "queue", "topic": "hexagon",
		"function": "step", "bucket": "stored_data"} {
		n := titledNode("")
		n.Style = vm.WithShape(n.Style, shape)
		src := Emit(vm.ViewModel{View: vm.View{ID: "v", Direction: "down"}, Nodes: []vm.Node{n}})
		if !strings.Contains(string(src), "shape: "+d2+"\n") {
			t.Errorf("%s: emitted\n%s", shape, src)
		}
		if _, err := compile(context.Background(), ruler, src); err != nil {
			t.Errorf("%s does not render: %v", shape, err)
		}
	}
}
