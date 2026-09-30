package d2

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

var update = flag.Bool("update", false, "rewrite golden files from current behaviour")

func node(id, role, kind, label, parent string, tags ...string) vm.Node {
	r := vm.NodeRole(role)
	return vm.Node{ID: id, Address: strings.ReplaceAll(id, "__", "."), Role: r, Kind: kind,
		Label: label, Parent: parent, Style: vm.StyleFor(r, kind, tags)}
}

func edge(src, tgt, label, tech string, crossing bool) vm.Edge {
	return vm.Edge{ID: vm.EdgeIDFor(src, tgt), Source: src, Target: tgt, Label: label,
		Technology: tech, Relationships: []string{src + ".uses.x"}, Crossing: crossing,
		Style: vm.EdgeStyle{Dashed: crossing}}
}

// emitCases cover every shape, nesting, self-loops, crossing edges, awkward
// label text, and tags. Nodes and edges are listed pre-sorted, as the
// projection guarantees.
func emitCases() map[string]vm.ViewModel {
	outside := vm.Node{ID: vm.OutsideNodeID, Role: vm.RoleOutside, Label: "outside this view",
		Style: vm.StyleFor(vm.RoleOutside, "", nil)}
	api := node("container__api", "element", "container", "API", "system__shop", "pci", "edge")
	api.Technology, api.Description = "Go", `Says "hi": costs $5\nand \ more`
	return map[string]vm.ViewModel{
		"landscape": {
			View:    vm.View{ID: "landscape", Kind: vm.KindLandscapeView},
			Sources: []string{"main.loko.hcl"},
			Nodes: []vm.Node{
				node("external__bank", "element", "external", "Bank", ""),
				node("person__customer", "element", "person", "Customer", ""),
				node("system__shop", "element", "system", "Shop", ""),
			},
			Edges: []vm.Edge{
				edge("person__customer", "system__shop", "Buys", "HTTPS", false),
				edge("system__shop", "external__bank", "Charges", "", false),
				edge("system__shop", "system__shop", "Calls itself", "", false),
			},
		},
		"subject": {
			View:    vm.View{ID: "system-shop", Kind: vm.KindSystemView, Subject: "system.shop"},
			Sources: []string{"b.loko.hcl", "a.loko.hcl"},
			Nodes: []vm.Node{
				api,
				node("container__db", "element", "container", "DB", "system__shop"),
				outside,
				node("system__shop", "subject", "system", "Shop", ""),
			},
			Edges: []vm.Edge{
				edge("container__api", "container__db", "Reads", "SQL", false),
				edge("container__api", vm.OutsideNodeID, "2 relationships", "", true),
			},
		},
		"deployment": {
			View:    vm.View{ID: "deployment-prod", Kind: vm.KindDeploymentView, Subject: "deployment.prod"},
			Sources: []string{"deploy.loko.hcl"},
			Nodes: []vm.Node{
				node("deployment__prod", "subject", "", "prod", ""),
				node("deployment__prod__instance__api", "instance", "container", "api", "deployment__prod__node__a__b"),
				node("deployment__prod__node__a", "group", "", "a", "deployment__prod"),
				node("deployment__prod__node__a__b", "group", "", "b", "deployment__prod__node__a"),
			},
		},
	}
}

func TestEmitGolden(t *testing.T) {
	for name, m := range emitCases() {
		t.Run(name, func(t *testing.T) {
			got := Emit(m)
			if !bytes.Equal(got, Emit(m)) {
				t.Fatal("Emit is not deterministic")
			}
			first, _, _ := strings.Cut(string(got), "\n")
			if want := "# " + vm.NoticeText(m.Sources); first != want {
				t.Errorf("line 1 = %q, want %q", first, want)
			}
			golden(t, filepath.Join("testdata", "golden", name+".d2"), got)
		})
	}
}

func TestEmitShapesAndStyles(t *testing.T) {
	t.Parallel()
	out := string(Emit(emitCases()["subject"]))
	checks := []string{
		`"system__shop": "Shop\n[System]" {`,
		`fill: transparent`,
		`stroke-dash: 4`,
		`"container__api": "API\n[Container: Go]\n\nSays \"hi\": costs \$5\\nand \\ more" {`,
		`class: [kind-container; tag-edge; tag-pci]`,
		`fill: "#438dd5"`,
		`"outside": "outside this view" {`,
		`shape: oval`,
		`"system__shop"."container__api" -> "outside": "2 relationships" {`,
		"classes: {\n  kind-container: {}\n",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("emitted D2 lacks %q\n---\n%s", c, out)
		}
	}
	land := string(Emit(emitCases()["landscape"]))
	for _, c := range []string{`shape: c4-person`, `"system__shop" -> "system__shop": "Calls itself"`} {
		if !strings.Contains(land, c) {
			t.Errorf("landscape D2 lacks %q", c)
		}
	}
}

func golden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s differs from golden (run with -update after inspecting)\n--- got\n%s", path, got)
	}
}
