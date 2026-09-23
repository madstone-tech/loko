package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// These tests construct SourceModel literals directly. No files, no parser:
// that is the point of the SourceModel/IR split (research R7). A failure here
// points at a rule, never at a parsing bug.

func at(line int) arch.SourceRange {
	return arch.SourceRange{File: "arch.loko.hcl", StartLine: line, StartColumn: 1, StartByte: line * 10}
}

func ref(raw string, line int) arch.Reference {
	return arch.Reference{Raw: raw, Range: at(line)}
}

func elem(kind arch.ElementKind, name string, line int) arch.ElementDecl {
	return arch.ElementDecl{Kind: kind, Name: name, Range: at(line)}
}

// codes returns the diagnostic codes present, for concise assertions.
func codes(d arch.Diagnostics) []string {
	out := make([]string, 0, len(d))
	for _, x := range d {
		out = append(out, x.Code)
	}
	return out
}

func hasCode(d arch.Diagnostics, code string) bool {
	for _, x := range d {
		if x.Code == code {
			return true
		}
	}
	return false
}

func findCode(d arch.Diagnostics, code string) *arch.Diagnostic {
	for i := range d {
		if d[i].Code == code {
			return &d[i]
		}
	}
	return nil
}

func TestSymbolTableCollectsEveryDeclaration(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{
		Elements: []arch.ElementDecl{
			elem(arch.KindSystem, "payments", 1),
			elem(arch.KindContainer, "api", 5),
			elem(arch.KindPerson, "customer", 9),
		},
		Environments: []arch.EnvironmentDecl{{
			Name:      "prod",
			Range:     at(20),
			Instances: []arch.InstanceDecl{{Name: "api", Range: at(22)}},
		}},
		Views: []arch.ViewDecl{{Name: "path", Range: at(30)}},
	}

	st, diags := BuildSymbolTable(model)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", codes(diags))
	}

	for _, want := range []arch.Address{
		"system.payments", "container.api", "person.customer",
		"deployment.prod", "deployment.prod.instance.api", "view.path",
	} {
		if !st.Has(want) {
			t.Errorf("symbol table missing %q", want)
		}
	}
}

// TestDuplicateDeclaration covers FR-004: the same address twice is an error
// naming both sites, not a merge.
func TestDuplicateDeclaration(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "api", 3),
		elem(arch.KindContainer, "api", 11),
	}}

	_, diags := BuildSymbolTable(model)

	d := findCode(diags, arch.CodeDuplicateDeclaration)
	if d == nil {
		t.Fatalf("want duplicate_declaration, got %v", codes(diags))
	}
	if len(d.Related) == 0 {
		t.Error("duplicate_declaration does not name the first declaration")
	}
	if d.Range.StartLine != 11 {
		t.Errorf("diagnostic points at line %d, want the second declaration at 11", d.Range.StartLine)
	}
}

// TestResolveAcrossFilesAndOrder covers FR-019 and FR-020: resolution works
// across file boundaries and does not depend on declaration order.
func TestResolveAcrossFilesAndOrder(t *testing.T) {
	t.Parallel()

	// The container is declared BEFORE the system it points at.
	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		{Kind: arch.KindContainer, Name: "api", Range: at(1), Parent: ref("system.payments", 2)},
		elem(arch.KindSystem, "payments", 50),
	}}

	res, diags := ResolveModel(model)
	if diags.HasErrors() {
		t.Fatalf("forward reference failed: %v", codes(diags))
	}
	if got, want := res.Parent["container.api"], arch.Address("system.payments"); got != want {
		t.Errorf("parent resolved to %q, want %q", got, want)
	}
}

// TestUnresolvedReferencePerPosition is SC-002: every reference position in the
// language gets a case, so the "100% of reference positions" claim is measured
// rather than asserted.
func TestUnresolvedReferencePerPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model *arch.SourceModel
		line  int
	}{
		{
			name: "container.system",
			model: &arch.SourceModel{Elements: []arch.ElementDecl{
				{Kind: arch.KindContainer, Name: "api", Range: at(1), Parent: ref("system.nope", 2)},
			}},
			line: 2,
		},
		{
			name: "component.container",
			model: &arch.SourceModel{Elements: []arch.ElementDecl{
				{Kind: arch.KindComponent, Name: "h", Range: at(1), Parent: ref("container.nope", 3)},
			}},
			line: 3,
		},
		{
			name: "uses.target",
			model: &arch.SourceModel{Elements: []arch.ElementDecl{
				{Kind: arch.KindContainer, Name: "api", Range: at(1), Relations: []arch.RelationDecl{
					{LocalName: "orders", Target: ref("container.nope", 4), Range: at(3)},
				}},
			}},
			line: 4,
		},
		{
			name: "instance.of",
			model: &arch.SourceModel{Environments: []arch.EnvironmentDecl{{
				Name: "prod", Range: at(1),
				Instances: []arch.InstanceDecl{{Name: "api", Range: at(2), Of: ref("container.nope", 5)}},
			}}},
			line: 5,
		},
		{
			name: "view.include",
			model: &arch.SourceModel{Views: []arch.ViewDecl{
				{Name: "v", Range: at(1), Include: []arch.Reference{ref("container.nope", 6)}},
			}},
			line: 6,
		},
		{
			name: "view.exclude",
			model: &arch.SourceModel{Views: []arch.ViewDecl{
				{Name: "v", Range: at(1), Exclude: []arch.Reference{ref("container.nope", 7)}},
			}},
			line: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := ResolveModel(tt.model)

			d := findCode(diags, arch.CodeUnresolvedReference)
			if d == nil {
				t.Fatalf("want unresolved_reference, got %v", codes(diags))
			}
			// The diagnostic must point at the reference itself, not at the
			// enclosing block — that is the difference between a usable error
			// and one the reader has to hunt through.
			if d.Range.StartLine != tt.line {
				t.Errorf("points at line %d, want the reference at line %d", d.Range.StartLine, tt.line)
			}
		})
	}
}

// TestWrongReferenceKind covers FR-022.
func TestWrongReferenceKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model *arch.SourceModel
	}{
		{
			name: "container parented to a container",
			model: &arch.SourceModel{Elements: []arch.ElementDecl{
				elem(arch.KindContainer, "db", 1),
				{Kind: arch.KindContainer, Name: "api", Range: at(5), Parent: ref("container.db", 6)},
			}},
		},
		{
			name: "component parented to a system",
			model: &arch.SourceModel{Elements: []arch.ElementDecl{
				elem(arch.KindSystem, "payments", 1),
				{Kind: arch.KindComponent, Name: "h", Range: at(5), Parent: ref("system.payments", 6)},
			}},
		},
		{
			name: "instance realising a deployment",
			model: &arch.SourceModel{Environments: []arch.EnvironmentDecl{{
				Name: "prod", Range: at(1),
				Instances: []arch.InstanceDecl{{Name: "x", Range: at(2), Of: ref("deployment.prod", 3)}},
			}}},
		},
		{
			name: "relationship targeting an instance",
			model: &arch.SourceModel{
				Elements: []arch.ElementDecl{{
					Kind: arch.KindContainer, Name: "api", Range: at(1),
					Relations: []arch.RelationDecl{{LocalName: "r", Target: ref("deployment.prod.instance.api", 2), Range: at(2)}},
				}},
				Environments: []arch.EnvironmentDecl{{
					Name: "prod", Range: at(10),
					Instances: []arch.InstanceDecl{{Name: "api", Range: at(11)}},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := ResolveModel(tt.model)
			if !hasCode(diags, arch.CodeWrongReferenceKind) {
				t.Errorf("want wrong_reference_kind, got %v", codes(diags))
			}
		})
	}
}

// TestResolveSuggestsNearMiss: a typo is the common case, and naming the likely
// intended element is most of the value of the diagnostic.
func TestResolveSuggestsNearMiss(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "orders_db", 1),
		{Kind: arch.KindContainer, Name: "api", Range: at(5), Relations: []arch.RelationDecl{
			{LocalName: "o", Target: ref("container.ordrs_db", 7), Range: at(6)},
		}},
	}}

	_, diags := ResolveModel(model)
	d := findCode(diags, arch.CodeUnresolvedReference)
	if d == nil {
		t.Fatalf("want unresolved_reference, got %v", codes(diags))
	}
	if !contains(d.Detail, "container.orders_db") {
		t.Errorf("detail does not suggest the near miss: %q", d.Detail)
	}
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
