package html

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"runtime"
	"strings"
	"sync"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Backend is the "html" backend: a browsable site with an index, a page per
// view and a page per element (FR-025). Diagrams are referenced by path, so
// this backend needs no other backend (FR-012).
type Backend struct{}

// New returns the "html" backend.
func New() *Backend { return &Backend{} }

// Format implements usecases.Backend.
func (*Backend) Format() vm.Format { return vm.FormatHTML }

// pageData is what every template receives: only view-model values, so a
// template cannot reach the file system or the IR (contracts/theme.md).
type pageData struct {
	Kind       string // "index", "view" or "element"
	Title      string
	Root       string // relative href from this page to the site root
	BodyClass  string
	Project    vm.ProjectHeader
	Projection *vm.Projection
	Element    *vm.ElementPage
	View       *vm.ViewModel
	Diagram    *vm.ViewModel
	// Nav is the rendered nav block for this page's depth (renderNavs).
	Nav template.HTML
}

// Render implements usecases.Backend.
func (b *Backend) Render(_ context.Context, in *vm.Projection, opts usecases.RenderOptions) ([]vm.Artifact, error) {
	th, err := loadTheme(opts.Theme)
	if err != nil {
		return nil, err
	}
	jobs := pageJobs(in)
	if err := th.renderNavs(in, jobs); err != nil {
		return nil, err
	}
	out, err := th.renderPages(jobs)
	if err != nil {
		return nil, err
	}
	assets, err := th.assetArtifacts(in)
	if err != nil {
		return nil, err
	}
	return append(out, assets...), nil
}

// pageJob is one page to render.
type pageJob struct {
	path, owner string
	sources     []string
	data        pageData
}

// pageJobs lists every page: the index, one per view, one per element.
func pageJobs(in *vm.Projection) []pageJob {
	views := make(map[vm.ViewID]*vm.ViewModel, len(in.Views))
	for i := range in.Views {
		views[in.Views[i].View.ID] = &in.Views[i]
	}
	job := func(path, owner string, sources []string, d pageData) pageJob {
		d.Root, d.Project, d.Projection = rootOf(path), in.Project, in
		return pageJob{path, owner, sources, d}
	}
	jobs := make([]pageJob, 0, 1+len(in.Views)+len(in.Pages))
	jobs = append(jobs, job(vm.IndexPage("html"), "", in.Sources,
		pageData{Kind: "index", Title: in.Project.Name, BodyClass: "page-index", Diagram: views["landscape"]}))
	for i := range in.Views {
		v := &in.Views[i]
		jobs = append(jobs, job(v.PagePath, ownerOf(*v), v.Sources, pageData{Kind: "view", Title: v.View.Title,
			BodyClass: "page-view view-" + string(v.View.Kind), View: v}))
	}
	for i := range in.Pages {
		p := &in.Pages[i]
		jobs = append(jobs, job(p.PagePath, p.Address, p.Sources, pageData{Kind: "element", Title: p.Name,
			BodyClass: strings.Join(append([]string{"page-element"}, p.Classes...), " "),
			Element:   p, Diagram: views[p.Diagram]}))
	}
	return jobs
}

// renderNavs renders the nav block once per distinct page depth. It lists
// every view, so rendering it per page made site generation quadratic in
// the number of views. The block still comes from the resolved theme, so an
// override of "nav" applies as before.
func (th *theme) renderNavs(in *vm.Projection, jobs []pageJob) error {
	navs := map[string]template.HTML{}
	for i := range jobs {
		root := jobs[i].data.Root
		nav, ok := navs[root]
		if !ok {
			var buf bytes.Buffer
			if err := th.tmpl.ExecuteTemplate(&buf, "nav", pageData{Root: root, Project: in.Project, Projection: in}); err != nil {
				return th.execError("the navigation", err)
			}
			nav = template.HTML(buf.String()) //nolint:gosec // output of the escaping html/template engine
			navs[root] = nav
		}
		jobs[i].data.Nav = nav
	}
	return nil
}

// renderPages executes pages on a bounded worker pool; each result is stored
// at its job's index, so order never depends on completion. html/template is
// safe for concurrent execution once parsed. Every worker exits before
// renderPages returns.
func (th *theme) renderPages(jobs []pageJob) ([]vm.Artifact, error) {
	out := make([]vm.Artifact, len(jobs))
	errs := make([]error, len(jobs))
	next := make(chan int)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), max(len(jobs), 1)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				j := jobs[i]
				out[i], errs[i] = th.page(j.path, j.owner, j.sources, j.data)
			}
		}()
	}
	for i := range jobs {
		next <- i
	}
	close(next)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (th *theme) page(path, owner string, sources []string, d pageData) (vm.Artifact, error) {
	var buf bytes.Buffer
	if err := th.tmpl.ExecuteTemplate(&buf, "layout", d); err != nil {
		return vm.Artifact{}, th.execError(path, err)
	}
	return vm.Artifact{Path: path, Format: vm.FormatHTML, Owner: owner,
		Bytes: withNotice(buf.Bytes(), vm.NoticeText(sources))}, nil
}

// execError attributes a template execution failure to the overrides when
// there are any, so a broken theme fails as theme_invalid (FR-035).
func (th *theme) execError(what string, err error) error {
	if len(th.overridden) > 0 {
		return &usecases.ThemeError{File: strings.Join(th.overridden, ", "),
			Message: fmt.Sprintf("rendering %s: %v", what, err)}
	}
	return fmt.Errorf("rendering %s: %w", what, err)
}

// withNotice keeps <!DOCTYPE html> first, so pages stay in standards mode,
// and places the generated-file notice on line 2 (research R7). The backend
// adds it, not the template, so no theme override can drop it.
func withNotice(page []byte, notice string) []byte {
	comment := "<!-- " + notice + " -->\n"
	if bytes.HasPrefix(page, []byte("<!DOCTYPE")) {
		if i := bytes.IndexByte(page, '\n'); i >= 0 {
			return append([]byte(string(page[:i+1])+comment), page[i+1:]...)
		}
	}
	return append([]byte(comment), page...)
}

func (th *theme) assetArtifacts(in *vm.Projection) ([]vm.Artifact, error) {
	notice := "/* " + vm.NoticeText(in.Sources) + " */\n"
	var out []vm.Artifact
	for _, name := range assetNames {
		out = append(out, vm.Artifact{Path: "assets/" + name, Format: vm.FormatHTML,
			Bytes: append([]byte(notice), th.assets[name]...)})
	}
	type entry struct {
		N string `json:"n"`
		K string `json:"k"`
		P string `json:"p"`
		D string `json:"d,omitempty"`
	}
	index := make([]entry, 0, len(in.Pages))
	for _, p := range in.Pages {
		index = append(index, entry{p.Name, p.Kind, p.PagePath, p.Description})
	}
	data, err := json.Marshal(index)
	if err != nil {
		return nil, err
	}
	out = append(out, vm.Artifact{Path: "assets/search-index.js", Format: vm.FormatHTML,
		Bytes: []byte(notice + "window.LOKO_SEARCH = " + string(data) + ";\n")})
	return out, nil
}

// rootOf is the relative href from the page at path to the site root.
func rootOf(path string) string {
	return strings.TrimSuffix(vm.Rel(path, "x"), "x")
}

func ownerOf(v vm.ViewModel) string {
	if v.View.Subject != "" {
		return v.View.Subject
	}
	return "view " + string(v.View.ID)
}
