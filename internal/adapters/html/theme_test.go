package html

import (
	"context"
	"errors"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

func override(name, body string) vm.ThemeFile {
	return vm.ThemeFile{Name: name, Bytes: []byte(body), Origin: "templates/" + name}
}

func TestOverrideBlockKeepsTheRest(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	base := renderSite(t, proj, nil)
	themed := renderSite(t, proj, []vm.ThemeFile{override("layout.gohtml", `{{define "header"}}<header class="acme">ACME {{.Title}}</header>{{end}}`)})

	for path, body := range themed {
		if !strings.HasSuffix(path, ".html") {
			if body != base[path] {
				t.Errorf("%s changed although no asset was overridden", path)
			}
			continue
		}
		if !strings.Contains(body, `<header class="acme">ACME `) {
			t.Errorf("%s does not use the overridden header", path)
		}
		// Replacing only the header must leave everything else byte-identical.
		if withoutHeader(body) != withoutHeader(base[path]) {
			t.Errorf("%s differs beyond the overridden block (US7/AC1)", path)
		}
	}
	again := renderSite(t, proj, []vm.ThemeFile{override("layout.gohtml", `{{define "header"}}<header class="acme">ACME {{.Title}}</header>{{end}}`)})
	for p := range themed {
		if again[p] != themed[p] {
			t.Errorf("%s differs between themed builds (US7/AC4)", p)
		}
	}
}

func TestOverrideAssets(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	out := renderSite(t, proj, []vm.ThemeFile{override("style.css", "body{color:red}"), override("custom.css", ".x{}")})
	if !strings.HasSuffix(out["assets/style.css"], "*/\nbody{color:red}") {
		t.Errorf("style.css not replaced:\n%s", out["assets/style.css"])
	}
	if !strings.HasSuffix(out["assets/custom.css"], "*/\n.x{}") {
		t.Errorf("custom.css not replaced:\n%s", out["assets/custom.css"])
	}
}

func TestMalformedOverridesFail(t *testing.T) {
	t.Parallel()
	proj := loadProjection(t)
	tests := []struct {
		name, file, body, want string
		line                   int
	}{
		{"unknown file", "sytle.css", "a{}", "overridable", 0},
		{"parse error", "partials.gohtml", "{{define \"link\"}}\n{{.Ref.Name}\n{{end}}", "", 2},
		{"unknown block", "partials.gohtml", `{{define "nope"}}{{end}}`, `"nope"`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := New().Render(context.Background(), proj, usecases.RenderOptions{Theme: []vm.ThemeFile{override(tt.file, tt.body)}})
			te, ok := errors.AsType[*usecases.ThemeError](err)
			if !ok {
				t.Fatalf("err = %v, want a ThemeError (FR-035)", err)
			}
			if te.File != "templates/"+tt.file || !strings.Contains(te.Message, tt.want) || (tt.line > 0 && te.Line != tt.line) {
				t.Errorf("ThemeError = %+v", te)
			}
		})
	}
}

// withoutHeader drops a page's first <header>…</header>, the block the
// override replaces.
func withoutHeader(page string) string {
	i := strings.Index(page, "<header")
	j := strings.Index(page, "</header>")
	if i < 0 || j < i {
		return page
	}
	return page[:i] + page[j+len("</header>"):]
}

// blocks014 is every block feature 014's theme contract documents.
var blocks014 = []string{"layout", "head", "header", "nav", "footer", "index", "element", "element-tables",
	"relation-table", "element-children", "prose-missing", "view", "tag-chips", "link", "diagram"}

// TestAll014BlocksStillDefined is spec FR-042: no 014 block name is removed,
// and each still reaches the pages, so an override of it takes effect.
func TestAll014BlocksStillDefined(t *testing.T) {
	t.Parallel()
	th, err := loadTheme(nil)
	if err != nil {
		t.Fatal(err)
	}
	proj := loadProjection(t)
	for i := range proj.Pages { // so prose-missing is reached too
		if proj.Pages[i].Address == "container.api" {
			proj.Pages[i].ProseMissing, proj.Pages[i].Prose = true, ""
		}
	}
	for _, name := range blocks014 {
		if th.tmpl.Lookup(name) == nil {
			t.Errorf("block %q is no longer defined", name)
			continue
		}
		if name == "layout" {
			continue
		}
		file := "partials.gohtml"
		switch name {
		case "head", "header", "nav", "footer":
			file = "layout.gohtml"
		case "index":
			file = "index.gohtml"
		case "view":
			file = "view.gohtml"
		case "element", "element-tables", "relation-table", "element-children", "prose-missing":
			file = "element.gohtml"
		}
		marker := "MARK-" + name
		body := `{{define "` + name + `"}}` + marker + `{{end}}`
		site := renderWith(t, proj)
		themed := renderSiteWith(t, proj, []vm.ThemeFile{override(file, body)})
		found := false
		for p, b := range themed {
			if strings.Contains(b, marker) && site[p] != b {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("overriding %q changes no page: the block is no longer called", name)
		}
	}
}

// Test014OverridesStillBuild: overrides written against 014's data keep
// rendering against 021's (spec FR-042, SC-010).
func Test014OverridesStillBuild(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"relation-table":   `{{define "relation-table"}}{{$tags := anyTags .Rows}}<table>{{range .Rows}}<tr><td>{{template "link" (linkData $.Root .Other)}}</td><td>{{.Description}}</td><td>{{.Technology}}</td>{{if $tags}}<td>{{hashTags .Tags}}</td>{{end}}</tr>{{end}}</table>{{end}}`,
		"element-children": `{{define "element-children"}}{{with .Element}}<ul>{{range .Children}}<li>{{template "link" (linkData $.Root .)}} <small>{{.Kind}}</small></li>{{end}}</ul>{{end}}{{end}}`,
		"element-tables":   `{{define "element-tables"}}{{with .Element}}{{if .Uses}}<h2>Uses</h2>{{template "relation-table" (rowsData $.Root .Uses)}}{{end}}{{end}}{{end}}`,
		"diagram":          `{{define "diagram"}}<img src="{{.Root}}{{.View.DiagramPath}}">{{end}}`,
	} {
		file := "element.gohtml"
		if name == "diagram" {
			file = "partials.gohtml"
		}
		site := renderSiteWith(t, loadProjection(t), []vm.ThemeFile{override(file, body)})
		if len(site) == 0 {
			t.Errorf("%s override produced nothing", name)
		}
	}
}

// renderSiteWith renders with SVGs available and the given overrides.
func renderSiteWith(t *testing.T, proj *vm.Projection, theme []vm.ThemeFile) map[string]string {
	t.Helper()
	as, err := New().Render(context.Background(), proj, usecases.RenderOptions{Theme: theme,
		Upstream: func() (map[string][]byte, error) {
			return map[string][]byte{"diagrams/container-api.svg": fakeSVG("component.repo")}, nil
		}})
	if err != nil {
		t.Fatalf("render with overrides: %v", err)
	}
	out := map[string]string{}
	for _, a := range as {
		out[a.Path] = string(a.Bytes)
	}
	return out
}
