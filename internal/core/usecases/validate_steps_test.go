package usecases

import (
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// matchModel: web → api (sync), api → payments, repo (in api) → db,
// worker → queue (trigger, drawn queue → worker), loop → loop (self).
func matchModel(steps ...arch.StepDecl) *arch.SourceModel {
	c := func(name string, line int, rels ...arch.RelationDecl) arch.ElementDecl {
		e := elem(arch.KindContainer, name, line)
		e.Parent, e.Relations = ref("system.shop", line), rels
		return e
	}
	repo := elem(arch.KindComponent, "repo", 9)
	repo.Parent = ref("container.api", 9)
	repo.Relations = []arch.RelationDecl{uses("write", "container.db", 9)}
	trigger := uses("triggered", "container.queue", 7)
	trigger.Kind = arch.RelTrigger
	return &arch.SourceModel{
		Elements: []arch.ElementDecl{
			elem(arch.KindSystem, "shop", 1),
			elem(arch.KindExternal, "payments", 2),
			c("web", 3, uses("orders", "container.api", 3)),
			c("api", 4, uses("charge", "external.payments", 4)),
			c("db", 5), c("queue", 6),
			c("worker", 7, trigger),
			c("loop", 8, uses("again", "container.loop", 8)),
			repo,
		},
		Views: []arch.ViewDecl{{Name: "flow", Range: at(20), Steps: steps}},
	}
}

func checkSteps(t *testing.T, m *arch.SourceModel) (arch.Diagnostics, *Resolved) {
	t.Helper()
	res, diags := ResolveModel(m)
	if diags.HasErrors() {
		t.Fatalf("resolve: %v", diags)
	}
	return ValidateSteps(m, res), res
}

// TestValidateStepsSelection is research R2.
func TestValidateStepsSelection(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		mutate func(*arch.ViewDecl)
		line   int
		detail string
	}{
		{"include", func(v *arch.ViewDecl) { v.Include, v.IncludeRange = []arch.Reference{ref("system.shop", 21)}, at(21) }, 21, "include"},
		{"exclude", func(v *arch.ViewDecl) { v.Exclude, v.ExcludeRange = []arch.Reference{ref("container.db", 22)}, at(22) }, 22, "exclude"},
		{"tags", func(v *arch.ViewDecl) { v.Tags, v.TagsRange = []string{"pci"}, at(23) }, 23, "tags"},
		{"style without steps", func(v *arch.ViewDecl) { v.Steps, v.Style, v.StyleRange = nil, "sequence", at(24) }, 24, "style"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := matchModel(step("order", "container.web", "container.api", 25))
			tt.mutate(&m.Views[0])
			d, _ := checkSteps(t, m)
			if len(d) != 1 || d[0].Code != arch.CodeDynamicViewSelection || d[0].Severity != arch.SeverityError ||
				d[0].Range.StartLine != tt.line || !strings.Contains(d[0].Detail, tt.detail) {
				t.Errorf("diagnostics %+v", d)
			}
		})
	}
}

// TestValidateStepsMatching is research R3.
func TestValidateStepsMatching(t *testing.T) {
	t.Parallel()
	reply := func(s arch.StepDecl) arch.StepDecl { s.Reply = true; return s }
	via := func(s arch.StepDecl, rel string) arch.StepDecl { s.Via = ref(rel, s.Range.StartLine); return s }
	for _, tt := range []struct {
		name string
		step arch.StepDecl
		code string
		rel  arch.Address
	}{
		{"direct", step("s", "container.web", "container.api", 30), "", "container.web.uses.orders"},
		{"container step over a component relationship", step("s", "container.api", "container.db", 30), "", "component.repo.uses.write"},
		{"component step over a container relationship", step("s", "container.web", "component.repo", 30), "", "container.web.uses.orders"},
		{"trigger in drawn direction", step("s", "container.queue", "container.worker", 30), "", "container.worker.uses.triggered"},
		{"reply reversed", reply(step("s", "external.payments", "container.api", 30)), "", "container.api.uses.charge"},
		{"against the direction without reply", step("s", "external.payments", "container.api", 30), arch.CodeStepUnmatched, ""},
		{"no relationship", step("s", "container.web", "external.payments", 30), arch.CodeStepUnmatched, ""},
		{"via between other elements", via(step("s", "container.web", "container.db", 30), "container.api.uses.charge"), arch.CodeStepMismatch, ""},
		{"via that matches", via(step("s", "container.web", "container.api", 30), "container.web.uses.orders"), "", "container.web.uses.orders"},
		{"self with a self relationship", step("s", "container.loop", "container.loop", 30), "", "container.loop.uses.again"},
		{"self without one", step("s", "container.web", "container.web", 30), arch.CodeStepUnmatched, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d, res := checkSteps(t, matchModel(tt.step))
			if got := codes(d); (tt.code == "") != (len(got) == 0) || (tt.code != "" && !slices.Equal(got, []string{tt.code})) {
				t.Fatalf("codes %v, want %q", got, tt.code)
			}
			if got := res.StepRelationship["view.flow.step.s"]; got != tt.rel {
				t.Errorf("relationship %q, want %q", got, tt.rel)
			}
			if tt.code == "" {
				return
			}
			if d[0].Range.StartLine != 30 || d[0].Address != "view.flow.step.s" {
				t.Errorf("diagnostic %+v", d[0])
			}
			if tt.code == arch.CodeStepUnmatched && (d[0].Severity != arch.SeverityWarning || !strings.Contains(d[0].Detail, tt.step.From.Raw)) {
				t.Errorf("unmatched %+v", d[0])
			}
			if tt.code == arch.CodeStepMismatch && (d[0].Severity != arch.SeverityError ||
				!strings.Contains(d[0].Detail, "container.api") || !strings.Contains(d[0].Detail, "external.payments") ||
				!strings.Contains(d[0].Detail, "container.db")) {
				t.Errorf("mismatch %+v", d[0])
			}
		})
	}
}

// TestValidateStepsFirstMatch: with several matching relationships and no
// via, the first in address order is used.
func TestValidateStepsFirstMatch(t *testing.T) {
	t.Parallel()
	m := matchModel(step("s", "container.api", "container.db", 30))
	m.Elements[3].Relations = append(m.Elements[3].Relations, uses("audit", "container.db", 4))
	if _, res := checkSteps(t, m); res.StepRelationship["view.flow.step.s"] != "component.repo.uses.write" {
		t.Errorf("relationship %q", res.StepRelationship["view.flow.step.s"])
	}
}
