package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestProjectIsPureAndSorted(t *testing.T) {
	t.Parallel()
	ir := deepEnvIR()
	a, diagsA, err := Project(ir, nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	b, diagsB, err := Project(ir, nil, Provenance{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) || !reflect.DeepEqual(diagsA, diagsB) {
		t.Fatal("Project is not deterministic (FR-010)")
	}
	var ids []viewmodel.ViewID
	for _, v := range a.Views {
		ids = append(ids, v.View.ID)
	}
	want := []viewmodel.ViewID{"landscape", "system-core", "container-svc", "deployment-prod"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("views = %v, want %v", ids, want)
	}
	if a.Project.Name != "test" {
		t.Errorf("project = %+v", a.Project)
	}
	if len(a.Environments) != 1 || a.Environments[0].PagePath != "view/deployment-prod.html" {
		t.Errorf("environments = %+v", a.Environments)
	}
}
