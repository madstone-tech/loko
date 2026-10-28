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

// Render implements usecases.Backend.
func (b *Backend) Render(_ context.Context, in *vm.Projection, opts usecases.RenderOptions) ([]vm.Artifact, error) {
	th, err := loadTheme(opts.Theme)
	if err != nil {
		return nil, err
	}
	jobs := pageJobs(in, opts)
	out, err := th.renderPages(jobs)
	if err != nil {
		return nil, err
	}
	if opts.Upstream != nil {
		upstream, err := opts.Upstream()
		if err != nil {
			return nil, err
		}
		fillInline(out, jobs, upstream)
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
func pageJobs(in *vm.Projection, opts usecases.RenderOptions) []pageJob {
	s := newSite(in, opts.Upstream, opts.Formats)
	job := func(path, owner string, sources []string, d pageData) pageJob {
		d.Root, d.Project, d.Projection = rootOf(path), in.Project, in
		return pageJob{path, owner, sources, d}
	}
	jobs := make([]pageJob, 0, 1+len(in.Views)+len(in.Pages))
	jobs = append(jobs, job(vm.IndexPage("html"), "", in.Sources, s.indexData()))
	for i := range in.Views {
		v := &in.Views[i]
		jobs = append(jobs, job(v.PagePath, ownerOf(*v), v.Sources, s.viewData(v)))
	}
	for i := range in.Pages {
		p := &in.Pages[i]
		sources := p.SiteSources
		if sources == nil {
			sources = p.Sources
		}
		jobs = append(jobs, job(p.PagePath, p.Address, sources, s.elementData(p)))
	}
	return jobs
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
		wg.Go(func() {
			for i := range next {
				j := jobs[i]
				out[i], errs[i] = th.page(j.path, j.owner, j.sources, j.data)
			}
		})
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

// page renders one page. The nav block is rendered first, per page, because
// its tree is pruned to the page (feature 021); layout then places it.
func (th *theme) page(path, owner string, sources []string, d pageData) (vm.Artifact, error) {
	buf := pageBuffers.Get().(*bytes.Buffer)
	defer pageBuffers.Put(buf)
	buf.Reset()
	buf.Grow(64 << 10)
	d.TreeHTML = treeHTML(d.Root, d.Tree)
	views, err := th.renderViewsNav(d)
	if err != nil {
		return vm.Artifact{}, th.execError(path, err)
	}
	d.ViewsNav = views
	if err := th.tmpl.ExecuteTemplate(buf, "nav", d); err != nil {
		return vm.Artifact{}, th.execError(path, err)
	}
	d.Nav = template.HTML(buf.String()) //nolint:gosec // output of the escaping html/template engine
	buf.Reset()
	if err := th.tmpl.ExecuteTemplate(buf, "layout", d); err != nil {
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
// The result is always a fresh slice of exactly the right size: page may be a
// pooled buffer that is reused once this returns.
func withNotice(page []byte, notice string) []byte {
	comment := "<!-- " + notice + " -->\n"
	out := make([]byte, 0, len(page)+len(comment))
	if bytes.HasPrefix(page, []byte("<!DOCTYPE")) {
		if i := bytes.IndexByte(page, '\n'); i >= 0 {
			out = append(out, page[:i+1]...)
			out = append(out, comment...)
			return append(out, page[i+1:]...)
		}
	}
	out = append(out, comment...)
	return append(out, page...)
}

func (th *theme) assetArtifacts(in *vm.Projection) ([]vm.Artifact, error) {
	notice := "/* " + vm.NoticeText(in.Sources) + " */\n"
	out, err := brandArtifacts(vm.NoticeText(in.Sources))
	if err != nil {
		return nil, err
	}
	icons, err := iconsCSS(vm.NoticeText(in.Sources))
	if err != nil {
		return nil, err
	}
	out = append(out, icons)
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

// subtitle is the address name shown under a titled page's heading.
func subtitle(p *vm.ElementPage) string {
	if p.Title == "" {
		return ""
	}
	return p.Name
}

// pageBuffers reuses page buffers across pages: a page with an inlined
// diagram is tens of kilobytes, and growing a fresh buffer for each one made
// a large site's rebuild spend its time returning memory (feature 021).
var pageBuffers = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// renderViewsNav renders the nav-views block once per page depth and current
// view, and reuses it for every other page with the same pair.
func (th *theme) renderViewsNav(d pageData) (template.HTML, error) {
	key := d.Root
	if d.View != nil {
		key += "\x00" + string(d.View.View.ID)
	}
	th.viewsMu.Lock()
	defer th.viewsMu.Unlock()
	if html, ok := th.viewsNav[key]; ok {
		return html, nil
	}
	var buf bytes.Buffer
	if err := th.tmpl.ExecuteTemplate(&buf, "nav-views", d); err != nil {
		return "", err
	}
	if th.viewsNav == nil {
		th.viewsNav = map[string]template.HTML{}
	}
	th.viewsNav[key] = template.HTML(buf.String()) //nolint:gosec // output of the escaping html/template engine
	return th.viewsNav[key], nil
}
