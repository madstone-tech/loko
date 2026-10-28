package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestDescribeSteps: describe lists a dynamic view's style and steps, and
// nothing new for a view without them (feature 019, FR-016).
func TestDescribeSteps(t *testing.T) {
	t.Parallel()
	ir := dynamicIR(st("order", "container.web", "container.api", "container.web.uses.orders", "Submit"))
	ir.Views[0].Style = arch.StyleSequence
	ir.Views[0].Steps[0].Via = "container.web.uses.orders"
	ir.Views = append(ir.Views, arch.View{Address: "view.plain", Name: "plain"})
	out := &DescribeResult{Level: "summary"}
	describeIR(ir, out, "")
	if len(out.Views) != 2 {
		t.Fatalf("views %+v", out.Views)
	}
	v := out.Views[0]
	if v.Style != "sequence" || len(v.Steps) != 1 {
		t.Fatalf("dynamic view %+v", v)
	}
	s := v.Steps[0]
	if s.Number != 1 || s.Name != "order" || s.From != "container.web" || s.To != "container.api" ||
		s.Via != "container.web.uses.orders" || s.Relationship != "container.web.uses.orders" || s.Description != "Submit" {
		t.Errorf("step %+v", s)
	}
	if p := out.Views[1]; p.Style != "" || p.Steps != nil {
		t.Errorf("plain view %+v", p)
	}
}
