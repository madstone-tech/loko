package d2

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// sequenceCase is the checkout flow drawn as a sequence diagram: the web
// container sits in a shop boundary, which the sequence style leaves out.
func sequenceCase() vm.ViewModel {
	n := func(id, kind, label, parent string, role vm.NodeRole) vm.Node {
		return vm.Node{ID: id, Address: strings.ReplaceAll(id, "__", "."), Role: role, Kind: kind, Label: label,
			Parent: parent, Style: vm.StyleFor(role, kind, nil)}
	}
	ref := func(id string) vm.LinkRef { return vm.LinkRef{Address: strings.ReplaceAll(id, "__", ".")} }
	row := func(num int, name, from, to, text, tech string, reply bool) vm.StepRow {
		return vm.StepRow{Number: num, Name: name, From: ref(from), To: ref(to), FromNode: from, ToNode: to,
			Text: text, Technology: tech, Reply: reply}
	}
	return vm.ViewModel{
		View: vm.View{ID: "checkout", Kind: vm.KindDeclaredView, Subject: "view.checkout", Style: "sequence",
			Title: "Checkout", TitleDeclared: true, Direction: "down", Layout: "elk"},
		Sources: []string{"main.loko.hcl"},
		Nodes: []vm.Node{
			n("container__web", "container", "web", "system__shop", vm.RoleElement),
			n("external__payments", "external", "payments", "", vm.RoleElement),
			n("person__customer", "person", "customer", "", vm.RoleElement),
			n("system__shop", "system", "shop", "", vm.RoleSubject),
		},
		Steps: []vm.StepRow{
			row(1, "browse", "person__customer", "container__web", "", "HTTPS", false),
			row(2, "charge", "container__web", "external__payments", "Charge card", "HTTPS", false),
			row(3, "paid", "external__payments", "container__web", "Charged", "", true),
		},
	}
}

// TestEmitSequence is research R7: lifelines in first-appearance order,
// arrows in step order, replies dashed, no boundaries, no layout settings.
func TestEmitSequence(t *testing.T) {
	t.Parallel()
	got := string(Emit(sequenceCase()))
	for _, want := range []string{
		"loko-view-title: \"Checkout\" {",
		"loko-flow: \"\" {\n  shape: sequence_diagram\n",
		"  \"person__customer\" -> \"container__web\": \"1\\nHTTPS\" {\n    style.stroke: \"#737373\"\n",
		"  \"container__web\" -> \"external__payments\": \"2. Charge card\\nHTTPS\" {\n",
		"  \"external__payments\" -> \"container__web\": \"3. Charged\" {\n    style.stroke: \"#737373\"\n    style.font-color: \"#3D3D3D\"\n    style.stroke-dash: 4\n  }\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("sequence D2 lacks %q\n%s", want, got)
		}
	}
	order := []string{`  "person__customer": `, `  "container__web": `, `  "external__payments": `}
	for i := 1; i < len(order); i++ {
		if strings.Index(got, order[i-1]) > strings.Index(got, order[i]) || strings.Index(got, order[i]) < 0 {
			t.Errorf("lifelines not in first-appearance order:\n%s", got)
		}
	}
	for _, absent := range []string{"system__shop", "direction:", "layout-engine"} {
		if strings.Contains(got, absent) {
			t.Errorf("sequence D2 contains %q", absent)
		}
	}
	golden(t, filepath.Join("testdata", "golden", "dynamic_sequence.d2"), []byte(got))
}

// TestSequenceRendersInOrder: the SVG draws the steps top to bottom, under
// either layout engine.
func TestSequenceRendersInOrder(t *testing.T) {
	t.Setenv("PATH", "")
	for _, layout := range []string{"", "elk"} {
		v := sequenceCase()
		v.View.Layout = layout
		as, err := NewSVGBackend().Render(context.Background(), projectionOf(v), usecases.RenderOptions{})
		if err != nil {
			t.Fatal(err)
		}
		svg := as[0].Bytes
		if a, b := textY(t, svg, "2. Charge card"), textY(t, svg, "3. Charged"); a >= b {
			t.Errorf("layout %q: step 2 at y=%v is not above step 3 at y=%v", layout, a, b)
		}
	}
}
