package usecases

import (
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
