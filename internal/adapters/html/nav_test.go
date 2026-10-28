package html

import (
	"regexp"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

func treeOf(page string) string {
	i := strings.Index(page, `<nav class="tree"`)
	if i < 0 {
		return ""
	}
	return page[i : i+strings.Index(page[i:], "</nav>")]
}

// TestPrunedNavPerPage is spec FR-001, FR-002 and SC-002.
func TestPrunedNavPerPage(t *testing.T) {
	t.Parallel()
	site := renderWith(t, loadProjection(t))
	handler := treeOf(site["element/component/handler.html"])
	for _, want := range []string{
		`href="../../element/system/shop.html"`,               // top level
		`href="../../element/container/api.html"`,             // the system's containers
		`href="../../element/component/repo.html"`,            // the container's components
		`<li class="tree-item active"><div class="tree-row">`, // the current element
		`href="../../element/component/handler.html" title="handler" aria-current="page">handler</a>`,
	} {
		if !strings.Contains(handler, want) {
			t.Errorf("handler's tree lacks %s", want)
		}
	}
	if strings.Contains(handler, "component/authorizer.html") {
		t.Error("handler's tree carries another system's branch")
	}
	index := treeOf(site["index.html"])
	if strings.Contains(index, "element/container/") || !strings.Contains(index, "element/system/shop.html") {
		t.Error("the index tree is the top level only")
	}
	if regexp.MustCompile(`>[a-z]+\.[a-z_]+</a>`).MatchString(handler) {
		t.Error("the tree shows an address instead of a name (FR-005)")
	}
}

func TestBreadcrumb(t *testing.T) {
	t.Parallel()
	site := renderWith(t, loadProjection(t))
	page := site["element/component/handler.html"]
	want := `<nav class="breadcrumb" aria-label="Breadcrumb"><ol><li><a href="../../index.html">two-systems</a></li>` +
		`<li><a href="../../element/system/shop.html">shop</a></li><li><a href="../../element/container/api.html">api</a></li>` +
		`<li><span aria-current="page">handler</span></li></ol></nav>`
	if !strings.Contains(page, want) {
		t.Errorf("breadcrumb:\n%s", page[strings.Index(page, `<nav class="breadcrumb"`):][:400])
	}
	if strings.Contains(site["index.html"], `class="breadcrumb"`) {
		t.Error("the index has no breadcrumb")
	}
	if !strings.Contains(site["view/container-api.html"], `<li><a href="../element/container/api.html">api</a></li><li><span aria-current="page">`) {
		t.Error("a view's breadcrumb runs through its subject")
	}
}

func TestPager(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	site := renderWith(t, proj)
	for _, p := range proj.Pages {
		body := site[p.PagePath]
		if got := strings.Contains(body, `rel="prev"`); got != (p.Prev != nil) {
			t.Errorf("%s: prev link %v, Prev %v", p.PagePath, got, p.Prev)
		}
		if got := strings.Contains(body, `rel="next"`); got != (p.Next != nil) {
			t.Errorf("%s: next link %v, Next %v", p.PagePath, got, p.Next)
		}
	}
}

// TestNavOverrideStillWorks: a 014 nav override that lists the views still
// renders (spec FR-042).
func TestNavOverrideStillWorks(t *testing.T) {
	t.Parallel()
	site := renderSite(t, loadProjection(t), []vm.ThemeFile{override("layout.gohtml",
		`{{define "nav"}}<aside class="acme">{{range .Projection.Views}}<a href="{{$.Root}}{{.PagePath}}">{{.View.Title}}</a>{{end}}</aside>{{end}}`)})
	if !strings.Contains(site["element/component/handler.html"], `<aside class="acme"><a href="../../view/landscape.html">`) {
		t.Error("a 014 nav override no longer renders")
	}
}
