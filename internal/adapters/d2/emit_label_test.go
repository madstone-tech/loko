package d2

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func roles(rs ...LineRole) []LineRole { return rs }

// nodeLabel and edgeLabel are the label text D2 draws for a node or an edge.
func nodeLabel(n vm.Node) string {
	text, _ := joinLines(nodeLines(n))
	return text
}

func edgeLabel(e vm.Edge) string {
	text, _ := joinLines(edgeLines(e))
	return text
}

// TestNodePlanRoles: each label line gets the role the SVG pass styles it by
// (feature 020, research R2), and only cards are in the plan.
func TestNodePlanRoles(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("Order fulfilment coordinator ", 4) + "service"
	multi := titledNode("")
	multi.Description = "First line\nSecond line"
	noDesc := titledNode("")
	noDesc.Description = ""
	wrapped := slices.Repeat(roles(roleTitle), strings.Count(wrapTitle(long), "\n")+1)
	if len(wrapped) < 3 {
		t.Fatalf("title wraps over %d lines, want at least 3", len(wrapped))
	}
	for _, tt := range []struct {
		name string
		n    vm.Node
		want []LineRole
	}{
		{"untitled", titledNode(""), roles(roleTitle, roleKind, roleBlank, roleDescription)},
		{"wrapped title", titledNode(long), append(wrapped, roleKind, roleBlank, roleDescription)},
		{"no description", noDesc, roles(roleTitle, roleKind)},
		{"multi-line description", multi, roles(roleTitle, roleKind, roleBlank, roleDescription, roleDescription)},
	} {
		_, plan := emitPlanned(vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{tt.n}})
		got := plan.Nodes["container__api"]
		if !slices.Equal(got.Lines, tt.want) {
			t.Errorf("%s: roles = %v, want %v", tt.name, got.Lines, tt.want)
		}
		if text := nodeLabel(tt.n); len(got.Lines) != strings.Count(text, "\n")+1 {
			t.Errorf("%s: %d roles for %d lines", tt.name, len(got.Lines), strings.Count(text, "\n")+1)
		}
	}
}

func TestNodePlanMembers(t *testing.T) {
	t.Parallel()
	_, plan := emitPlanned(emitCases()["subject"])
	api, ok := plan.Nodes["system__shop.container__api"]
	if !ok {
		t.Fatalf("api card missing from plan: %v", sortedKeys(plan.Nodes))
	}
	if api.Icon != "container" || api.Colour != emitCases()["subject"].Nodes[0].Style.FontColor {
		t.Errorf("api plan = %+v", api)
	}
	if got := sortedKeys(plan.Nodes); !slices.Equal(got, []string{"system__shop.container__api", "system__shop.container__db"}) {
		t.Errorf("plan nodes = %v, want the two cards only (no subject, no outside)", got)
	}
	_, dep := emitPlanned(emitCases()["deployment"])
	if got := sortedKeys(dep.Nodes); !slices.Equal(got, []string{
		"deployment__prod.deployment__prod__node__a.deployment__prod__node__a__b.deployment__prod__instance__api",
	}) {
		t.Errorf("deployment plan nodes = %v, want the instance only", got)
	}
}

// TestWrapDescription: a long description wraps into lines of at most
// descWidth characters, keeps existing breaks, and never truncates. The
// model-level report (pcr-competitive-pricing): one 282-character description
// drew a 1,960px card and made the whole canvas a 6,366 × 669 ribbon.
func TestWrapDescription(t *testing.T) {
	t.Parallel()
	long := "Third-party structured retail-data API. Phase 1 source for Amazon, Walmart, " +
		"Best Buy, Costco, Home Depot, Lowe's, Target, Sam's Club, Wayfair and Bed Bath & Beyond."
	wrapped := wrapDescription(long)
	for _, line := range strings.Split(wrapped, "\n") {
		if len([]rune(line)) > descWidth {
			t.Errorf("line %q is %d characters, budget %d", line, len([]rune(line)), descWidth)
		}
	}
	if got, want := strings.Join(strings.Fields(wrapped), " "), strings.Join(strings.Fields(long), " "); got != want {
		t.Errorf("words changed: %q", got)
	}
	if got := wrapDescription("First line\nSecond line"); got != "First line\nSecond line" {
		t.Errorf("existing breaks not kept: %q", got)
	}
}

func TestEdgePlan(t *testing.T) {
	t.Parallel()
	a, b := titledNode(""), titledNode("")
	b.ID = "container__b"
	a.Parent, b.Parent = "system__s", "system__s"
	sys := vm.Node{ID: "system__s", Role: vm.RoleSubject, Kind: vm.KindSystem, Label: "s",
		Style: vm.StyleFor(vm.RoleSubject, vm.KindSystem, nil)}
	tagged := edge("container__api", "container__b", "Publishes", "SQS", false)
	tagged.Tags = []string{"async", "pii"}
	m := vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{a, b, sys},
		Edges: []vm.Edge{
			edge("container__api", "container__b", "Reads", "SQL", false),
			edge("container__api", "container__b", "", "gRPC", false),
			tagged,
			edge("container__b", "container__api", "Line one\nline two", "", false),
			edge("container__b", "container__api", "Plain", "", false),
		}}
	_, plan := emitPlanned(m)
	want := map[string][]LineRole{
		"system__s.(container__api -> container__b)[0]": roles(roleEdgeLabel, roleEdgeTech),
		"system__s.(container__api -> container__b)[1]": roles(roleEdgeTech),
		"system__s.(container__api -> container__b)[2]": roles(roleEdgeLabel, roleEdgeTech, roleEdgeTags),
	}
	if !reflect.DeepEqual(plan.Edges, want) {
		t.Errorf("edge plan = %v\nwant %v", plan.Edges, want)
	}
	if got := edgeLabel(tagged); got != "Publishes\nSQS\n#async #pii" {
		t.Errorf("edge label = %q", got)
	}
	_, again := emitPlanned(m)
	if !reflect.DeepEqual(plan, again) {
		t.Error("the plan is not a pure function of the view model")
	}
}
