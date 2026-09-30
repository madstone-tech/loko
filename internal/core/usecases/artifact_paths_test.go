package usecases

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func TestCheckCollisions(t *testing.T) {
	t.Parallel()
	as := []viewmodel.Artifact{
		{Path: "diagrams/container-Api.svg", Owner: "container.Api"},
		{Path: "diagrams/container-api.svg", Owner: "container.api"},
		{Path: "element/container/Api.html", Owner: "container.Api"},
		{Path: "element/container/api.html", Owner: "container.api"},
		{Path: "index.html"},
		{Path: "index.html"}, // an exact duplicate is a collision too
	}
	diags := checkCollisions(as, Provenance{})
	if len(diags) != 3 {
		t.Fatalf("got %d diagnostics, want 3: %+v", len(diags), diags)
	}
	for _, d := range diags {
		if d.Severity != arch.SeverityError || d.Code != arch.CodeOutputPathCollision {
			t.Errorf("diagnostic = %+v", d)
		}
	}
	if !strings.Contains(diags[0].Detail, "container.Api") || !strings.Contains(diags[0].Detail, "container.api") {
		t.Errorf("a collision must name both sides (FR-028): %q", diags[0].Detail)
	}
	if got := checkCollisions([]viewmodel.Artifact{{Path: "a"}, {Path: "b"}}, Provenance{}); got != nil {
		t.Errorf("distinct paths reported: %+v", got)
	}
}

func TestBuildArtifactsReportsCollisions(t *testing.T) {
	t.Parallel()
	f := newBuildFixture(smallModel())
	f.svg.artifacts = []viewmodel.Artifact{{Path: "diagrams/A.svg", Owner: "x"}, {Path: "diagrams/a.svg", Owner: "y"}}
	res, err := Build(t.Context(), f.deps, BuildRequest{Root: ".", Formats: []string{"svg"}, OutDir: "out"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Diags.HasErrors() || res.Artifacts != nil || f.store.commits != 0 {
		t.Fatalf("a collision must abort the commit: %+v commits=%d", res, f.store.commits)
	}
}
