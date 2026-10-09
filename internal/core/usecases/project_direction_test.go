package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestViewDirections is FR-005 / research R5: top-down for container (system-*),
// component (container-*) and declared views; left-right for the landscape and
// deployment views; a declared view's own direction wins.
func TestViewDirections(t *testing.T) {
	t.Parallel()
	ir := declaredIR(
		arch.View{Address: "view.default", Name: "default", Include: []arch.Address{"system.shop"}},
		arch.View{Address: "view.wide", Name: "wide", Include: []arch.Address{"system.shop"}, Direction: "right"},
	)
	ir.Environments = []arch.Environment{{Address: "deployment.prod", Name: "prod",
		Instances: []arch.Instance{{Address: "deployment.prod.instance.api", Name: "api", Of: "container.api"}}}}
	views, _ := ResolveViews(ir, Provenance{})
	got := map[string]string{}
	for _, v := range views {
		got[string(v.ID)] = v.Direction
	}
	want := map[string]string{"landscape": "right", "system-shop": "down", "deployment-prod": "right",
		"default": "down", "wide": "right"}
	for id, dir := range want {
		if got[id] != dir {
			t.Errorf("%s: direction %q, want %q (all: %v)", id, got[id], dir, got)
		}
	}
	for _, v := range views {
		if v.Direction == "" {
			t.Errorf("%s has no direction; the projection always sets one", v.ID)
		}
	}
}
