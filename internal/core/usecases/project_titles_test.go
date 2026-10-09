package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// titledIR is sampleIR with a title on system s1 and container c1, and tags
// on the relationship p → c1.
func titledIR() *arch.IR {
	ir := sampleIR()
	for i := range ir.Elements {
		switch ir.Elements[i].Address {
		case aS1:
			ir.Elements[i].Title = "Storefront"
		case aC1:
			ir.Elements[i].Title = "Web app"
		}
	}
	for i := range ir.Relationships {
		if ir.Relationships[i].Source == aP {
			ir.Relationships[i].Tags = []string{"https", "public"}
		}
	}
	return ir
}

func TestTitlesReachNodesAndPages(t *testing.T) {
	t.Parallel()
	ir := titledIR()
	views, _ := ResolveViews(ir, Provenance{})
	var sys viewmodel.ViewModel
	for _, v := range views {
		if v.ID == "system-s1" {
			sys = ProjectView(ir, v, Provenance{})
		}
	}
	found := false
	for _, n := range sys.Nodes {
		if n.Address == string(aC1) {
			found = true
			if n.Title != "Web app" || n.Label != "c1" {
				t.Errorf("titled node: Title %q Label %q, want Web app / c1", n.Title, n.Label)
			}
		}
		if n.Address == string(aC2) && n.Title != "" {
			t.Errorf("untitled node has Title %q", n.Title)
		}
	}
	if !found {
		t.Fatal("container.c1 not in system-s1")
	}
	for _, p := range pagesFor(t, ir, nil) {
		if p.Address == string(aS1) && p.Title != "Storefront" {
			t.Errorf("page title = %q", p.Title)
		}
		if p.Address == string(aP) {
			if len(p.Uses) == 0 || p.Uses[0].Other.Title != "Web app" || len(p.Uses[0].Tags) != 2 {
				t.Errorf("person.p uses row: %+v", p.Uses)
			}
		}
	}
}

func TestShapesReachNodesPagesAndInstances(t *testing.T) {
	t.Parallel()
	ir := titledIR()
	for i := range ir.Elements {
		if ir.Elements[i].Address == aC1 {
			ir.Elements[i].Shape = "database"
		}
	}
	views, _ := ResolveViews(ir, Provenance{})
	for _, v := range views {
		if v.ID != "system-s1" {
			continue
		}
		for _, n := range ProjectView(ir, v, Provenance{}).Nodes {
			if n.Address == string(aC1) && n.Style.Shape != viewmodel.ShapeDatabase {
				t.Errorf("c1 shape = %s", n.Style.Shape)
			}
			if n.Address == string(aC2) && n.Style.Shape != viewmodel.ShapeRectangle {
				t.Errorf("unshaped c2 = %s", n.Style.Shape)
			}
		}
	}
	if p := pagesFor(t, ir, nil)[string(aC1)]; p.Shape != "database" {
		t.Errorf("page shape = %q", p.Shape)
	}
}
