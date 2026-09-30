package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func viewIDs(vs []viewmodel.View) []string {
	var out []string
	for _, v := range vs {
		out = append(out, string(v.ID))
	}
	return out
}

func TestResolveViewsDerived(t *testing.T) {
	t.Parallel()
	env := arch.NewEnvironmentAddress("prod")
	ir := irOf(
		[]arch.Element{
			el(arch.KindPerson, "p", ""),
			el(arch.KindSystem, "full", ""),
			el(arch.KindSystem, "hollow", ""),
			el(arch.KindContainer, "api", "system.full"),
			el(arch.KindContainer, "db", "system.full"),
			el(arch.KindComponent, "h", "container.api"),
		},
		nil,
		arch.Environment{Address: env, Name: "prod", Instances: []arch.Instance{
			{Address: arch.NewInstanceAddress(env, "api"), Name: "api", Of: "container.api"},
		}},
		arch.Environment{Address: arch.NewEnvironmentAddress("empty"), Name: "empty"},
	)

	views, diags := ResolveViews(ir, Provenance{})
	if len(diags) != 0 {
		t.Fatalf("diags = %v", diags)
	}
	// hollow has no containers, db has no components, and the empty
	// environment has no instances: none of them yields a view (FR-002).
	want := []string{"landscape", "system-full", "container-api", "deployment-prod"}
	if got := viewIDs(views); !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveViews = %v, want %v", got, want)
	}
	for _, v := range views {
		switch v.ID {
		case "landscape":
			if v.Kind != viewmodel.KindLandscapeView || v.Subject != "" || v.Title != "test" {
				t.Errorf("landscape = %+v", v)
			}
		case "system-full":
			if v.Kind != viewmodel.KindSystemView || v.Subject != "system.full" || v.Title != "full" {
				t.Errorf("system view = %+v", v)
			}
		case "container-api":
			if v.Kind != viewmodel.KindContainerView || v.Subject != "container.api" {
				t.Errorf("container view = %+v", v)
			}
		case "deployment-prod":
			if v.Kind != viewmodel.KindDeploymentView || v.Subject != string(env) {
				t.Errorf("deployment view = %+v", v)
			}
		}
	}
}

func TestResolveViewsNothingToDraw(t *testing.T) {
	t.Parallel()
	views, diags := ResolveViews(irOf(nil, nil), Provenance{})
	if len(views) != 0 || len(diags) != 0 {
		t.Fatalf("empty architecture: views=%v diags=%v, want none and no diagnostic", views, diags)
	}
}
