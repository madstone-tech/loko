package usecases

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestConnectivityPropagatesToAncestors: a system whose containers are wired
// up is not an orphan. In C4 an edge is normally drawn between containers, so
// without this every system in every project would warn — and a warning that
// always fires is one users learn to ignore.
func TestConnectivityPropagatesToAncestors(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "payments", 1),
		{Kind: arch.KindContainer, Name: "api", Range: at(5), Parent: ref("system.payments", 6),
			Relations: []arch.RelationDecl{uses("db", "container.orders_db", 7)}},
		{Kind: arch.KindContainer, Name: "orders_db", Range: at(10), Parent: ref("system.payments", 11)},
		{Kind: arch.KindComponent, Name: "handler", Range: at(15), Parent: ref("container.api", 16)},
	}}

	res, _ := ResolveModel(model)
	diags := ValidateWarnings(model, res, "")

	for _, d := range diags {
		if d.Code == arch.CodeOrphanElement && d.Address == "system.payments" {
			t.Error("a system whose containers are connected was reported as an orphan")
		}
	}

	// The component is genuinely unconnected and should still warn.
	var sawHandler bool
	for _, d := range diags {
		if d.Code == arch.CodeOrphanElement && d.Address == "component.handler" {
			sawHandler = true
		}
	}
	if !sawHandler {
		t.Error("a genuinely unconnected component was not reported; the propagation is too broad")
	}
}

// TestExternalNeedsNoDocs: an external is a system the team does not own, so
// it has no prose of its own to point at. Prose it does declare must exist.
func TestExternalNeedsNoDocs(t *testing.T) {
	t.Parallel()

	withDocs := elem(arch.KindExternal, "vendor", 5)
	withDocs.Docs = "docs/vendor.md"
	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindExternal, "stripe", 1),
		elem(arch.KindSystem, "shop", 3),
		withDocs,
	}}

	res, _ := ResolveModel(model)
	diags := ValidateWarnings(model, res, t.TempDir())

	got := map[arch.Address]string{}
	for _, d := range diags {
		if d.Code == arch.CodeMissingDocs || d.Code == arch.CodeDocsNotFound {
			got[d.Address] = d.Code
		}
	}
	want := map[arch.Address]string{
		"system.shop":     arch.CodeMissingDocs,
		"external.vendor": arch.CodeDocsNotFound,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("docs warnings = %v, want %v", got, want)
	}
}

// TestDocsNotFoundStatesTheRule: docs resolve from the project root, which a
// reader of a file in a subdirectory does not expect, so the warning says so.
func TestDocsNotFoundStatesTheRule(t *testing.T) {
	t.Parallel()

	sys := elem(arch.KindSystem, "s", 1)
	sys.Docs = "docs/s.md"
	model := &arch.SourceModel{Elements: []arch.ElementDecl{sys}}

	res, _ := ResolveModel(model)
	for _, d := range ValidateWarnings(model, res, t.TempDir()) {
		if d.Code == arch.CodeDocsNotFound {
			if !strings.Contains(d.Detail, "relative to the project root") {
				t.Errorf("detail %q does not say how docs paths resolve", d.Detail)
			}
			return
		}
	}
	t.Error("no docs_not_found warning")
}

// TestPropagationTerminatesOnCycle: a containment cycle is an error reported
// elsewhere, but the warning pass must not hang on one.
func TestPropagationTerminatesOnCycle(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		{Kind: arch.KindContainer, Name: "a", Range: at(1),
			Relations: []arch.RelationDecl{uses("b", "container.b", 2)}},
		elem(arch.KindContainer, "b", 5),
	}}
	res := &Resolved{
		Target: map[arch.Address]arch.Address{"container.a.uses.b": "container.b"},
		Parent: map[arch.Address]arch.Address{
			"container.a": "container.b",
			"container.b": "container.a",
		},
	}

	done := make(chan struct{})
	go func() {
		ValidateWarnings(model, res, "")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ValidateWarnings did not terminate on a containment cycle")
	}
}

// TestPlannedInstances is feature 017's instance.planned: no binding is
// expected yet, and a marker that outlives the infrastructure is reported.
func TestPlannedInstances(t *testing.T) {
	t.Parallel()
	model := &arch.SourceModel{Environments: []arch.EnvironmentDecl{{
		Name: "prod", Range: at(1),
		Instances: []arch.InstanceDecl{
			{Name: "planned", Range: at(2), Planned: true, PlannedRange: at(3)},
			{Name: "bound", Range: at(4), Planned: true, PlannedRange: at(5),
				Claims: []arch.ClaimDecl{{Kind: arch.ClaimTerraform, Address: "m.x", Range: at(6)}}},
			{Name: "unbound", Range: at(7)},
		},
	}}}
	res, _ := ResolveModel(model)
	got := map[string]arch.Diagnostic{}
	for _, d := range ValidateWarnings(model, res, "") {
		if d.Code == arch.CodeUnboundInstance || d.Code == arch.CodePlannedInstanceBound {
			got[string(d.Address)+" "+d.Code] = d
		}
	}
	if _, ok := got["deployment.prod.instance.planned "+arch.CodeUnboundInstance]; ok {
		t.Error("a planned instance without a binding draws unbound_instance")
	}
	if _, ok := got["deployment.prod.instance.unbound "+arch.CodeUnboundInstance]; !ok {
		t.Error("an unplanned instance without a binding draws no unbound_instance")
	}
	d, ok := got["deployment.prod.instance.bound "+arch.CodePlannedInstanceBound]
	if !ok || d.Range != at(5) {
		t.Errorf("planned and bound: %+v, want planned_instance_bound at the planned attribute", d)
	}
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}
