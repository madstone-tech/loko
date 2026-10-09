package d2

import (
	"context"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func titledNode(title string) vm.Node {
	return vm.Node{ID: "container__api", Address: "container.api", Role: vm.RoleElement, Kind: vm.KindContainer,
		Label: "api", Title: title, Technology: "Go", Description: "Serves orders",
		Style: vm.StyleFor(vm.RoleElement, vm.KindContainer, nil)}
}

func TestTitledLabel(t *testing.T) {
	t.Parallel()
	got := nodeLabel(titledNode("Orders API"))
	if want := "Orders API\n[Container: Go] · api\n\nServes orders"; got != want {
		t.Errorf("titled label = %q, want %q", got, want)
	}
	untitled := titledNode("")
	if got := nodeLabel(untitled); got != "api\n[Container: Go]\n\nServes orders" {
		t.Errorf("untitled label changed: %q", got)
	}
	noTech := titledNode("Orders API")
	noTech.Technology = ""
	if got := nodeLabel(noTech); !strings.HasPrefix(got, "Orders API\n[Container] · api") {
		t.Errorf("no technology: %q", got)
	}
}

func TestLongTitleWrapsAndRenders(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("Order fulfilment coordinator ", 4) + "service"
	if len(long) < 120 {
		t.Fatalf("title is %d chars", len(long))
	}
	label := nodeLabel(titledNode(long))
	head := label[:strings.Index(label, "\n[")]
	if strings.Join(strings.Fields(head), " ") != strings.Join(strings.Fields(long), " ") {
		t.Errorf("title was truncated or altered:\n%s", head)
	}
	for _, line := range strings.Split(head, "\n") {
		if len([]rune(line)) > titleWidth {
			t.Errorf("title line %q exceeds %d characters", line, titleWidth)
		}
	}
	m := vm.ViewModel{View: vm.View{ID: "v", Direction: "down"}, Nodes: []vm.Node{titledNode(long)}}
	src := Emit(m)
	if strings.Contains(string(src), "|md") {
		t.Error("labels must be plain text (no markdown labels)")
	}
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	svg, err := compile(context.Background(), ruler, src)
	if err != nil || !strings.Contains(string(svg), "fulfilment") || strings.Contains(string(svg), "foreignObject") {
		t.Errorf("titled node does not render as plain SVG text: %v", err)
	}
}
