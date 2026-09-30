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
		restore := strings.Replace(body, `<header class="acme">ACME `, `<header class="page-header"><h1>`, 1)
		restore = strings.Replace(restore, "</header>", "</h1></header>", 1)
		if restore != base[path] {
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
