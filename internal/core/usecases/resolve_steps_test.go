package usecases

import (
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// stepModel is web → api (declared), with one dynamic view of two steps.
func stepModel(steps ...arch.StepDecl) *arch.SourceModel {
	web := elem(arch.KindContainer, "web", 2)
	web.Parent = ref("system.shop", 2)
	web.Relations = []arch.RelationDecl{uses("orders", "container.api", 3)}
	api := elem(arch.KindContainer, "api", 4)
	api.Parent = ref("system.shop", 4)
	return &arch.SourceModel{
		Elements: []arch.ElementDecl{elem(arch.KindSystem, "shop", 1), web, api},
		Views:    []arch.ViewDecl{{Name: "checkout", Range: at(10), Steps: steps}},
	}
}

func step(name, from, to string, line int) arch.StepDecl {
	return arch.StepDecl{Name: name, From: ref(from, line), To: ref(to, line), Range: at(line)}
}

// TestResolveSteps: step references resolve like any other; step labels are
// declarations (feature 019).
func TestResolveSteps(t *testing.T) {
	t.Parallel()
	order := step("order", "container.web", "container.api", 11)
	order.Via = ref("container.web.uses.orders", 11)
	m := stepModel(order, step("again", "container.web", "container.api", 12))
	res, diags := ResolveModel(m)
	if len(diags) != 0 {
		t.Fatalf("diags %v", diags)
	}
	ir := BuildIR(m, res)
	v := ir.Views[0]
	if len(v.Steps) != 2 {
		t.Fatalf("steps %+v", v.Steps)
	}
	s := v.Steps[0]
	if s.Name != "order" || s.Number != 1 || s.Address != "view.checkout.step.order" || s.From != "container.web" ||
		s.To != "container.api" || s.Via != "container.web.uses.orders" {
		t.Errorf("first step %+v", s)
	}
	if v.Steps[1].Name != "again" || v.Steps[1].Number != 2 {
		t.Errorf("steps not in declaration order: %+v", v.Steps)
	}
}

func TestResolveStepsRefused(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		steps []arch.StepDecl
		code  string
		hint  string
	}{
		{"unknown from", []arch.StepDecl{step("s", "container.wbe", "container.api", 11)}, arch.CodeUnresolvedReference, "container.web"},
		{"via not a relationship", func() []arch.StepDecl {
			s := step("s", "container.web", "container.api", 11)
			s.Via = ref("container.api", 11)
			return []arch.StepDecl{s}
		}(), arch.CodeWrongReferenceKind, "uses"},
		{"duplicate label", []arch.StepDecl{step("s", "container.web", "container.api", 11),
			step("s", "container.web", "container.api", 12)}, arch.CodeDuplicateDeclaration, "view.checkout.step.s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, diags := ResolveModel(stepModel(tt.steps...))
			if !slices.Contains(codes(diags), tt.code) {
				t.Fatalf("codes %v, want %s", codes(diags), tt.code)
			}
			for _, d := range diags {
				if d.Code == tt.code && !strings.Contains(d.Detail, tt.hint) {
					t.Errorf("detail %q lacks %q", d.Detail, tt.hint)
				}
			}
		})
	}
	m := stepModel(step("s", "container.web", "container.api", 11))
	m.Views = append(m.Views, arch.ViewDecl{Name: "other", Range: at(20),
		Steps: []arch.StepDecl{step("s", "container.web", "container.api", 21)}})
	if _, diags := ResolveModel(m); len(diags) != 0 {
		t.Errorf("one label in two views: %v", diags)
	}
}
