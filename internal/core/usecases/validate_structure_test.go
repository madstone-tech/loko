package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// child builds an element with a parent reference.
func child(kind arch.ElementKind, name string, parent string, line int) arch.ElementDecl {
	return arch.ElementDecl{Kind: kind, Name: name, Range: at(line), Parent: ref(parent, line+1)}
}

// uses builds a relationship declaration.
func uses(local, target string, line int) arch.RelationDecl {
	return arch.RelationDecl{LocalName: local, Target: ref(target, line), Range: at(line)}
}

func TestWrongParentKind(t *testing.T) {
	t.Parallel()

	// A component whose parent is a system, and a container whose parent is a
	// component: both are containment errors (FR-028).
	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "s", 1),
		elem(arch.KindComponent, "c", 5),
		child(arch.KindComponent, "bad_component", "system.s", 10),
		child(arch.KindContainer, "bad_container", "component.c", 20),
	}}

	_, diags := ResolveModel(model)

	// Resolution already rejects these by kind, which is the earliest and most
	// precise place to catch them.
	n := 0
	for _, d := range diags {
		if d.Code == arch.CodeWrongReferenceKind {
			n++
		}
	}
	if n != 2 {
		t.Errorf("got %d wrong_reference_kind diagnostics, want 2: %v", n, codes(diags))
	}
}

// TestMissingParent covers the other half of FR-010: a container with no
// system at all, not merely one pointing somewhere wrong.
func TestMissingParent(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "orphan", 1),
		elem(arch.KindComponent, "loose", 5),
	}}

	res, diags := ResolveModel(model)
	diags = append(diags, ValidateStructure(model, res)...)

	n := 0
	for _, d := range diags {
		if d.Code == arch.CodeWrongParentKind {
			n++
		}
	}
	if n != 2 {
		t.Errorf("got %d wrong_parent_kind diagnostics, want 2 (container and component): %v", n, codes(diags))
	}
}

// TestContainmentCycle covers FR-028. Containment must be a single-parent
// acyclic tree.
func TestContainmentCycle(t *testing.T) {
	t.Parallel()

	// Two containers each claiming the other as its system is impossible, but
	// the kind check would catch that. Use components, where a cycle is
	// expressible with the right kinds throughout.
	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "s", 1),
		{Kind: arch.KindContainer, Name: "a", Range: at(5), Parent: ref("system.s", 6)},
		{Kind: arch.KindComponent, Name: "x", Range: at(10), Parent: ref("container.a", 11)},
	}}

	res, diags := ResolveModel(model)
	diags = append(diags, ValidateStructure(model, res)...)
	if hasCode(diags, arch.CodeContainmentCycle) {
		t.Errorf("a legal containment tree reported a cycle: %v", codes(diags))
	}

	// Now make it cyclic by hand: the resolver cannot produce this from valid
	// kinds, but a future language change could, and the check must hold.
	cyclic := &Resolved{
		Table: res.Table,
		Parent: map[arch.Address]arch.Address{
			"container.a": "container.b",
			"container.b": "container.a",
		},
	}
	cyclicModel := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "a", 1),
		elem(arch.KindContainer, "b", 5),
	}}
	if !hasCode(ValidateStructure(cyclicModel, cyclic), arch.CodeContainmentCycle) {
		t.Error("a containment cycle was not reported")
	}
}

// TestSelfContainmentCycle: an element inside itself.
func TestSelfContainmentCycle(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{elem(arch.KindContainer, "a", 1)}}
	res := &Resolved{Parent: map[arch.Address]arch.Address{"container.a": "container.a"}}

	if !hasCode(ValidateStructure(model, res), arch.CodeContainmentCycle) {
		t.Error("an element contained by itself was not reported")
	}
}

// TestRelationshipCyclesAreLegal is FR-032 and the single most likely design
// error in this feature. `api -> queue -> worker -> api` is a normal
// architecture. The containment check must never be applied to relationships.
func TestRelationshipCyclesAreLegal(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "s", 1),
		{Kind: arch.KindContainer, Name: "api", Range: at(5), Parent: ref("system.s", 6),
			Relations: []arch.RelationDecl{uses("q", "container.queue", 7)}},
		{Kind: arch.KindContainer, Name: "queue", Range: at(10), Parent: ref("system.s", 11),
			Relations: []arch.RelationDecl{uses("w", "container.worker", 12)}},
		{Kind: arch.KindContainer, Name: "worker", Range: at(15), Parent: ref("system.s", 16),
			Relations: []arch.RelationDecl{uses("a", "container.api", 17)}},
	}}

	res, diags := ResolveModel(model)
	diags = append(diags, ValidateStructure(model, res)...)

	if diags.HasErrors() {
		t.Fatalf("a legal dependency cycle produced errors: %v", codes(diags))
	}
	if hasCode(diags, arch.CodeContainmentCycle) {
		t.Fatal("dependency cycle reported as a containment cycle — the containment " +
			"traversal has been wrongly applied to relationships")
	}
}

func TestVersionConstraintDiagnostic(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{
		Project: arch.ProjectDecl{
			Name: "p", Declared: true, Version: ">= 99.0", VersionRange: at(2), Range: at(1),
		},
	}

	diags := ValidateProjectVersion(model, "1.0.0")
	d := findCode(diags, arch.CodeVersionUnsatisfied)
	if d == nil {
		t.Fatalf("want version_unsatisfied, got %v", codes(diags))
	}
	for _, want := range []string{">= 99.0", "1.0.0"} {
		if !contains(d.Detail, want) {
			t.Errorf("detail %q does not name %q", d.Detail, want)
		}
	}

	// A satisfied constraint, and an absent one, are both silent.
	ok := &arch.SourceModel{Project: arch.ProjectDecl{Name: "p", Declared: true, Version: "~> 1.0"}}
	if d := ValidateProjectVersion(ok, "1.4.2"); len(d) != 0 {
		t.Errorf("satisfied constraint produced %v", codes(d))
	}
	none := &arch.SourceModel{Project: arch.ProjectDecl{Name: "p", Declared: true}}
	if d := ValidateProjectVersion(none, "1.0.0"); len(d) != 0 {
		t.Errorf("absent constraint produced %v", codes(d))
	}
}
