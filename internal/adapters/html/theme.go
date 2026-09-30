package html

import (
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// assetNames are the non-template theme files, published under assets/.
var assetNames = []string{"custom.css", "site.js", "style.css"}

// theme is a resolved presentation: parsed templates plus asset bytes.
type theme struct {
	tmpl   *template.Template
	assets map[string][]byte
	// overridden lists the template overrides applied, for attributing an
	// execution failure to them.
	overridden []string
}

// loadTheme parses the built-in theme. Overrides are applied by applyOverrides.
func loadTheme(overrides []vm.ThemeFile) (*theme, error) {
	t := &theme{tmpl: template.New("").Funcs(funcs), assets: map[string][]byte{}}
	entries, err := fs.ReadDir(builtinTheme, "theme")
	if err != nil {
		return nil, fmt.Errorf("reading built-in theme: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := fs.ReadFile(builtinTheme, path.Join("theme", name))
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(name, ".gohtml") {
			if _, err := t.tmpl.New(name).Parse(string(data)); err != nil {
				return nil, fmt.Errorf("built-in theme %s: %w", name, err)
			}
			continue
		}
		t.assets[name] = data
	}
	if err := t.applyOverrides(overrides); err != nil {
		return nil, err
	}
	return t, nil
}

// overridable lists every theme file a project may replace
// (contracts/theme.md).
var overridable = []string{"custom.css", "element.gohtml", "index.gohtml", "layout.gohtml",
	"partials.gohtml", "site.js", "style.css", "view.gohtml"}

var parseLine = regexp.MustCompile(`:(\d+):`)

// applyOverrides resolves user overrides against the built-in theme
// (FR-034). A .css or .js override replaces the file; a .gohtml override is
// parsed after the built-ins, so each block it defines replaces the built-in
// block of that name and the rest stays built-in. Anything malformed fails
// with a ThemeError naming the file — a theme that half-applies is harder to
// diagnose than one that refuses (FR-035).
func (t *theme) applyOverrides(files []vm.ThemeFile) error {
	if p := vm.ValidateTheme(files, overridable); len(p) > 0 {
		return &usecases.ThemeError{File: p[0].Origin, Message: p[0].Message}
	}
	blocks := map[string]bool{}
	for _, tmpl := range t.tmpl.Templates() {
		blocks[tmpl.Name()] = true
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".gohtml") {
			t.assets[f.Name] = f.Bytes
			continue
		}
		probe, err := template.New("override:" + f.Name).Funcs(funcs).Parse(string(f.Bytes))
		if err != nil {
			return &usecases.ThemeError{File: f.Origin, Line: lineOf(err), Message: err.Error()}
		}
		for _, d := range probe.Templates() {
			if name := d.Name(); name != probe.Name() && !blocks[name] {
				return &usecases.ThemeError{File: f.Origin,
					Message: fmt.Sprintf("defines unknown block %q; see the overridable blocks in the theme reference", name)}
			}
		}
		if _, err := t.tmpl.New("override:" + f.Name).Parse(string(f.Bytes)); err != nil {
			return &usecases.ThemeError{File: f.Origin, Line: lineOf(err), Message: err.Error()}
		}
		t.overridden = append(t.overridden, f.Origin)
	}
	return nil
}

func lineOf(err error) int {
	m := parseLine.FindStringSubmatch(err.Error())
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

var funcs = template.FuncMap{
	"join":    strings.Join,
	"prose":   renderProse,
	"segment": vm.Segment,
	"diagramData": func(root string, v *vm.ViewModel) map[string]any {
		return map[string]any{"Root": root, "View": v}
	},
	"linkData": func(root string, ref any) map[string]any {
		if p, ok := ref.(*vm.LinkRef); ok {
			ref = *p
		}
		return map[string]any{"Root": root, "Ref": ref}
	},
	"rowsData": func(root string, rows []vm.RelationRow) map[string]any {
		return map[string]any{"Root": root, "Rows": rows}
	},
	"crossing": crossingLines,
}

// crossingLines describes a view's boundary connections, one line each.
func crossingLines(v *vm.ViewModel) []string {
	labels := map[string]string{}
	for _, n := range v.Nodes {
		labels[n.ID] = n.Label
	}
	var out []string
	for _, e := range v.Edges {
		switch {
		case !e.Crossing:
		case e.Target == vm.OutsideNodeID:
			out = append(out, labels[e.Source]+" → outside: "+e.Label)
		default:
			out = append(out, "outside → "+labels[e.Target]+": "+e.Label)
		}
	}
	return out
}
