package usecases

import (
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// dynamicIR: customer → web (shop) → api (shop) → payments; repo (in api) → db.
func dynamicIR(steps ...arch.Step) *arch.IR {
	ir := irOf(
		[]arch.Element{
			el(arch.KindPerson, "customer", ""),
			el(arch.KindExternal, "payments", ""),
			el(arch.KindSystem, "shop", ""),
			el(arch.KindContainer, "web", "system.shop"),
			el(arch.KindContainer, "api", "system.shop"),
			el(arch.KindContainer, "db", "system.shop"),
			el(arch.KindComponent, "repo", "container.api"),
		},
		[]arch.Relationship{
			rl("person.customer", "browse", "container.web", "Browses", "HTTPS"),
			rl("container.web", "orders", "container.api", "Places the order", "JSON"),
			rl("container.api", "charge", "external.payments", "Charges the card", "HTTPS"),
			rl("component.repo", "write", "container.db", "Stores the order", "SQL"),
		},
	)
	for i := range steps {
		steps[i].Number = i + 1
		steps[i].Address = arch.NewStepAddress("view.checkout", steps[i].Name)
	}
	ir.Views = []arch.View{{Address: "view.checkout", Name: "checkout", Steps: steps}}
	return ir
}

func st(name string, from, to, rel arch.Address, desc string) arch.Step {
	return arch.Step{Name: name, From: from, To: to, Relationship: rel, Description: desc}
}

// TestProjectDynamic is research R6: participants once, containers inside
// their system's boundary, one numbered edge per step, steps in order.
func TestProjectDynamic(t *testing.T) {
	t.Parallel()
	reply := st("ack", "external.payments", "container.api", "container.api.uses.charge", "")
	reply.Reply = true
	ir := dynamicIR(
		st("browse", "person.customer", "container.web", "person.customer.uses.browse", ""),
		st("order", "container.web", "container.api", "container.web.uses.orders", "Submit"),
		st("charge", "container.api", "external.payments", "container.api.uses.charge", ""),
		reply,
		st("store", "component.repo", "container.db", "component.repo.uses.write", ""),
		st("again", "container.web", "container.api", "", "Retry"),
	)
	vm := projectOne(t, ir, "checkout")
	idx := nodeIndex(vm)

	shop := idx["system__shop"]
	if shop.Role != viewmodel.RoleSubject || shop.Parent != "" {
		t.Errorf("shop is not a boundary: %+v", shop)
	}
	for id, parent := range map[string]string{
		"container__web": "system__shop", "container__api": "system__shop", "container__db": "system__shop",
		"component__repo": "container__api", "person__customer": "", "external__payments": "",
	} {
		n, ok := idx[id]
		if !ok || n.Parent != parent || n.Role != viewmodel.RoleElement {
			t.Errorf("%s = %+v (present %v), want parent %q", id, n, ok, parent)
		}
	}
	if len(vm.Nodes) != 7 {
		t.Errorf("%d nodes, want the 6 participants and the shop boundary", len(vm.Nodes))
	}

	labels := map[string]viewmodel.Edge{}
	for _, e := range vm.Edges {
		labels[e.ID] = e
	}
	if len(vm.Edges) != 6 {
		t.Fatalf("%d edges, want one per step: %v", len(vm.Edges), edgeIDs(vm))
	}
	for id, want := range map[string]string{
		"step__0001__browse": "1. Browses", "step__0002__order": "2. Submit", "step__0004__ack": "4. Charges the card",
		"step__0006__again": "6. Retry",
	} {
		if got := labels[id].Label; got != want {
			t.Errorf("%s label = %q, want %q", id, got, want)
		}
	}
	if e := labels["step__0002__order"]; e.Technology != "JSON" || e.Source != "container__web" || e.Target != "container__api" {
		t.Errorf("order edge %+v", e)
	}
	unlabelled := dynamicIR(st("bare", "container.web", "container.db", "", ""))
	if got := projectOne(t, unlabelled, "checkout").Edges[0].Label; got != "1" {
		t.Errorf("a step with no text is labelled %q, want the number alone", got)
	}
	if !labels["step__0004__ack"].Style.Dashed || labels["step__0003__charge"].Style.Dashed {
		t.Error("only the reply is dashed")
	}

	var order []string
	for _, r := range vm.Steps {
		order = append(order, r.Name)
	}
	if !slices.Equal(order, []string{"browse", "order", "charge", "ack", "store", "again"}) {
		t.Errorf("steps %v", order)
	}
	first := vm.Steps[0]
	if first.Number != 1 || first.From.Address != "person.customer" || first.FromNode != "person__customer" ||
		first.ToNode != "container__web" || first.Text != "Browses" {
		t.Errorf("first row %+v", first)
	}
	if !vm.Steps[3].Reply {
		t.Error("the reply row is not marked")
	}
}

// TestProjectDynamicNesting: a component sits in its container when the
// container takes part, a participating system is an element, not a
// boundary, and a view with steps is never reported empty.
func TestProjectDynamicNesting(t *testing.T) {
	t.Parallel()
	ir := dynamicIR(
		st("call", "container.api", "component.repo", "", ""),
		st("ext", "person.customer", "system.shop", "", ""),
	)
	vm := projectOne(t, ir, "checkout")
	idx := nodeIndex(vm)
	if idx["component__repo"].Parent != "container__api" {
		t.Errorf("repo parent %q", idx["component__repo"].Parent)
	}
	if idx["system__shop"].Role != viewmodel.RoleElement || idx["container__api"].Parent != "system__shop" {
		t.Errorf("shop %+v, api %+v", idx["system__shop"], idx["container__api"])
	}
}

// TestProjectDynamicStylesAgree: the two styles of one view number and label
// every step identically (SC-003); only the drawing differs.
func TestProjectDynamicStylesAgree(t *testing.T) {
	t.Parallel()
	ir := dynamicIR(st("order", "container.web", "container.api", "container.web.uses.orders", ""),
		st("charge", "container.api", "external.payments", "container.api.uses.charge", "Charge card"))
	edges := projectOne(t, ir, "checkout")
	ir.Views[0].Style = arch.StyleSequence
	seq := projectOne(t, ir, "checkout")
	if seq.View.Style != viewmodel.StyleSequence || edges.View.Style != "" {
		t.Errorf("styles %q / %q", edges.View.Style, seq.View.Style)
	}
	if !slices.Equal(edges.Steps, seq.Steps) {
		t.Errorf("step rows differ:\n%+v\n%+v", edges.Steps, seq.Steps)
	}
}
