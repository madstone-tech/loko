package html

import (
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// titled gives container.api a title everywhere it appears, and tags one of
// web's relationships, as the projection would.
func titled(p *vm.Projection) {
	for i := range p.Pages {
		pg := &p.Pages[i]
		if pg.Address == "container.api" {
			pg.Title = "Orders API"
		}
		for _, rows := range [][]vm.RelationRow{pg.Uses, pg.UsedBy} {
			for j := range rows {
				if rows[j].Other.Address == "container.api" {
					rows[j].Other.Title = "Orders API"
					rows[j].Tags = []string{"https"}
				}
			}
		}
	}
}

func TestSiteShowsTitlesAndTags(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	titled(proj)
	site := renderSite(t, proj, nil)
	api := site["element/container/api.html"]
	if !strings.Contains(api, "<h1>Orders API</h1>") || !strings.Contains(api, `<small class="address"><code>api</code></small>`) {
		t.Errorf("titled page heading:\n%s", api[:min(len(api), 1500)])
	}
	web := site["element/container/web.html"]
	if !strings.Contains(web, ">Orders API</a>") || !strings.Contains(web, "<th>Tags</th>") || !strings.Contains(web, "#https") {
		t.Error("web's uses table must link by title and show the Tags column")
	}
	if shop := site["element/system/shop.html"]; strings.Contains(shop, "<th>Tags</th>") {
		t.Error("a table without tagged rows must not grow a Tags column")
	}
	if !strings.Contains(site["view/system-shop.html"], `<a class="diagram-link" href="../diagrams/system-shop.svg">`) {
		t.Error("diagrams must link to their SVG (FR-011)")
	}
}

func TestViewLegendOnlyWithAsyncEdges(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	plain := renderSite(t, proj, nil)["view/system-shop.html"]
	if strings.Contains(plain, "diagram-legend") {
		t.Error("a view without async edges must not grow a legend")
	}
	for i := range proj.Views {
		if proj.Views[i].View.ID == "system-shop" && len(proj.Views[i].Edges) > 0 {
			proj.Views[i].Edges[0].Style.Async = true
		}
	}
	page := renderSite(t, proj, nil)["view/system-shop.html"]
	if !strings.Contains(page, "Long dashes: asynchronous.") {
		t.Errorf("legend missing:\n%s", page[:min(len(page), 1200)])
	}
}
