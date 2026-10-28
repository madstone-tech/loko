package html

import (
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestIndexPage is User Story 3 (spec FR-016–FR-018).
func TestIndexPage(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	proj.Overview = vm.Overview{Systems: 3, Externals: 1, Containers: 6, DataStores: 2, Relationships: 9, Async: 4, Warnings: 5, Infos: 7}
	index := renderWith(t, proj)["index.html"]
	if n := strings.Count(index, `<div class="stat-card">`); n != 4 {
		t.Errorf("%d stat cards, want 4", n)
	}
	for _, want := range []string{
		`<p class="stat-number">3</p><p class="stat-label">Systems</p><p class="stat-sub">1 external</p>`,
		`<p class="stat-number">6</p><p class="stat-label">Containers</p><p class="stat-sub">2 data stores</p>`,
		`<p class="stat-number">9</p><p class="stat-label">Relationships</p><p class="stat-sub">4 async</p>`,
		`<p class="stat-number">5</p><p class="stat-label">Warnings</p><p class="stat-sub">7 info</p>`,
		`<a href="view/landscape.html">Open view</a>`,
		`<section class="systems"`, `<section class="views"`,
	} {
		if !strings.Contains(index, want) {
			t.Errorf("index lacks %s", want)
		}
	}
	if strings.Contains(index, `<table class="nodes">`) {
		t.Error("the raw table of ids is gone (FR-018)")
	}
}

// TestViewPage is spec FR-019–FR-023 and data-model "View mode".
func TestViewPage(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	site := renderWith(t, proj, vm.FormatD2, vm.FormatSVG, vm.FormatHTML)
	page := site["view/system-shop.html"]
	for _, want := range []string{
		`<p class="kind-label">SYSTEM VIEW · AUTOMATIC</p>`,
		` elements</span><span class="sep" aria-hidden="true"> · </span><span>`,
		`<figure class="diagram zoomable">`, `data-zoom="reset"`,
		`<a href="../diagrams/system-shop.svg">SVG</a> · <a href="../diagrams/system-shop.d2">D2 source</a>`,
		`aria-describedby="in-this-view"`, `<section class="in-this-view" id="in-this-view"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("view page lacks %s", want)
		}
	}
	if n := strings.Count(page, `<svg role="img"`); n != 1 {
		t.Errorf("%d inline SVGs, want 1", n)
	}
	if strings.Contains(renderWith(t, proj)["view/system-shop.html"], "D2 source") {
		t.Error("the D2 link appears although this build writes no D2")
	}

	for _, tt := range []struct {
		view vm.View
		want string
	}{
		{vm.View{Kind: vm.KindSystemView}, "SYSTEM VIEW · AUTOMATIC"},
		{vm.View{Kind: vm.KindDeclaredView}, "DECLARED VIEW · DECLARED"},
		{vm.View{Kind: vm.KindDeclaredView, Style: "sequence"}, "DECLARED VIEW · SEQUENCE"},
		{vm.View{Kind: vm.KindDeploymentView}, "DEPLOYMENT VIEW · DEPLOYMENT"},
	} {
		s := newSite(&vm.Projection{}, nil, nil)
		if got := s.viewData(&vm.ViewModel{View: tt.view}).Label; got != tt.want {
			t.Errorf("label = %q, want %q", got, tt.want)
		}
	}
}
