package html

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
}

// Render implements usecases.Backend.
func (b *Backend) Render(_ context.Context, in *vm.Projection, opts usecases.RenderOptions) ([]vm.Artifact, error) {
	th, err := loadTheme(opts.Theme)
	if err != nil {
		return nil, err
	}
	var out []vm.Artifact
	page := func(path, owner string, sources []string, d pageData) error {
		d.Root, d.Project, d.Projection = rootOf(path), in.Project, in
		a, err := th.page(path, owner, sources, d)
		if err == nil {
			out = append(out, a)
		}
		return err
	}
	land, _ := in.View("landscape")
	if err := page(vm.IndexPage("html"), "", in.Sources,
		pageData{Kind: "index", Title: in.Project.Name, BodyClass: "page-index", Diagram: ptr(land)}); err != nil {
		return nil, err
	}
	for i := range in.Views {
		v := &in.Views[i]
		if err := page(v.PagePath, ownerOf(*v), v.Sources, pageData{Kind: "view", Title: v.View.Title,
			BodyClass: "page-view view-" + string(v.View.Kind), View: v}); err != nil {
			return nil, err
		}
	}
	for i := range in.Pages {
		p := &in.Pages[i]
		d, _ := in.View(p.Diagram)
		if err := page(p.PagePath, p.Address, p.Sources, pageData{Kind: "element", Title: p.Name,
			BodyClass: strings.Join(append([]string{"page-element"}, p.Classes...), " "),
			Element:   p, Diagram: ptr(d)}); err != nil {
			return nil, err
		}
	}
	assets, err := th.assetArtifacts(in)
	if err != nil {
		return nil, err
	}
	return append(out, assets...), nil
}

func (th *theme) page(path, owner string, sources []string, d pageData) (vm.Artifact, error) {
	var buf bytes.Buffer
	if err := th.tmpl.ExecuteTemplate(&buf, "layout", d); err != nil {
		if len(th.overridden) > 0 {
			return vm.Artifact{}, &usecases.ThemeError{File: strings.Join(th.overridden, ", "),
				Message: fmt.Sprintf("rendering %s: %v", path, err)}
		}
		return vm.Artifact{}, fmt.Errorf("rendering %s: %w", path, err)
	}
	return vm.Artifact{Path: path, Format: vm.FormatHTML, Owner: owner,
		Bytes: withNotice(buf.Bytes(), vm.NoticeText(sources))}, nil
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

func ptr(v vm.ViewModel) *vm.ViewModel {
	if v.View.ID == "" {
		return nil
	}
	return &v
}
