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

// TestTitledLabel is research R8: the kind in capitals, the technology as
// written, and a titled element's name at the end of the kind line.
func TestTitledLabel(t *testing.T) {
	t.Parallel()
	noTech := titledNode("Orders API")
	noTech.Technology = ""
	noDesc := titledNode("")
	noDesc.Description = ""
	bare := titledNode("")
	bare.Kind, bare.Technology, bare.Description = vm.KindSystem, "", ""
	grpc := titledNode("")
	grpc.Technology = "Go, gRPC"
	noKind := titledNode("Orders API")
	noKind.Kind, noKind.Technology, noKind.Description = "", "", ""
	for _, tt := range []struct {
		name string
		n    vm.Node
		want string
	}{
		{"titled", titledNode("Orders API"), "Orders API\nCONTAINER · Go · api\n\nServes orders"},
		{"untitled", titledNode(""), "api\nCONTAINER · Go\n\nServes orders"},
		{"titled without technology", noTech, "Orders API\nCONTAINER · api\n\nServes orders"},
		{"no description", noDesc, "api\nCONTAINER · Go"},
		{"kind only", bare, "api\nSYSTEM"},
		{"technology keeps its casing", grpc, "api\nCONTAINER · Go, gRPC\n\nServes orders"},
		{"titled without kind", noKind, "Orders API\nELEMENT · api"},
	} {
		if got := nodeLabel(tt.n); got != tt.want {
			t.Errorf("%s: label = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestLongTitleWrapsAndRenders(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("Order fulfilment coordinator ", 4) + "service"
	if len(long) < 120 {
		t.Fatalf("title is %d chars", len(long))
	}
	label := nodeLabel(titledNode(long))
	head := label[:strings.Index(label, "\nCONTAINER")]
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
	svg, _, err := compile(context.Background(), ruler, src, false)
	if err != nil || !strings.Contains(string(svg), "fulfilment") || strings.Contains(string(svg), "foreignObject") {
		t.Errorf("titled node does not render as plain SVG text: %v", err)
	}
}
