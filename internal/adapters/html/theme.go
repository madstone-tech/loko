package html

import (
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	// viewsNav caches the rendered nav-views block per page depth and current
	// view: it lists every view, and rendering it on every page made a large
	// site's build quadratic (feature 021).
	viewsMu  sync.Mutex
	viewsNav map[string]template.HTML
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
		if !e.IsDir() { // fonts/, brand/, icons/ are read by brandArtifacts
			names = append(names, e.Name())
		}
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
	"partials.gohtml", "rail.gohtml", "site.js", "style.css", "view.gohtml"}

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
	"anyTags": anyTags,
	// anyAsync and anyCrossing drive the view-page legend, shown only when a
	// view has async edges, so views without them render exactly as before.
	"anyAsync": func(v *vm.ViewModel) bool {
		return slices.ContainsFunc(v.Edges, func(e vm.Edge) bool { return e.Style.Async })
	},
	"anyCrossing": func(v *vm.ViewModel) bool {
		return slices.ContainsFunc(v.Edges, func(e vm.Edge) bool { return e.Crossing })
	},
	"hasLegend": func(v *vm.ViewModel) bool {
		return slices.ContainsFunc(v.Edges, func(e vm.Edge) bool { return e.Style.Async })
	},
	"hashTags": hashTags,
	"prose":    renderProse,
	"segment":  vm.Segment,
	// diagramData's optional arguments are the inline SVG, its caption id and
	// whether D2 sources exist; 014 themes call it with two (feature 021).
	"diagramData": func(root string, v *vm.ViewModel, extra ...any) map[string]any {
		d := map[string]any{"Root": root, "View": v, "Inline": template.HTML(""), "CaptionID": "", "HasD2": false}
		for i, key := range []string{"Inline", "CaptionID", "HasD2"} {
			if i < len(extra) {
				d[key] = extra[i]
			}
		}
		return d
	},
	"linkData": func(root string, ref any) map[string]any {
		if p, ok := ref.(*vm.LinkRef); ok {
			ref = *p
		}
		return map[string]any{"Root": root, "Ref": ref}
	},
	// rowsData takes relation rows (014) or neighbour rows (021), which share
	// the field names a relation-table override reads.
	"rowsData": func(root string, rows any) map[string]any {
		return map[string]any{"Root": root, "Rows": rows}
	},
	"tile": func(kind, icon string) map[string]string {
		return map[string]string{"Kind": kind, "Icon": icon}
	},
	"kindIcon": func(kind string) string { return vm.IconFor("", kind) },
	"nodeIcon": func(n vm.Node) string { return vm.IconFor(n.Style.Shape, n.Kind) },
	"hasProse": func(p *vm.ElementPage) bool {
		return p != nil && p.Prose != "" && !p.ProseInherited && !p.ProseMissing
	},
	// rowDir and rowSibling read a neighbour row; a 014 relation row (passed
	// by an element-tables override) counts as outgoing, not a sibling.
	"rowDir": func(row any) string {
		if n, ok := row.(vm.NeighbourRow); ok {
			return n.Direction
		}
		return vm.DirectionOut
	},
	"rowSibling": func(row any) bool {
		n, ok := row.(vm.NeighbourRow)
		return ok && n.Sibling
	},
	"arrow":      func(dir string) string { return arrows[dir][0] },
	"dirLabel":   func(dir string) string { return arrows[dir][1] },
	"systemRows": systemRows,
	"d2Path":     func(svg string) string { return strings.TrimSuffix(svg, ".svg") + ".d2" },
	"crossing":   crossingLines,
	"nodeRows":   nodeRows,
}

// nodeRow is one line of a deployment page's "Nodes" table (feature 018).
type nodeRow struct {
	Path        string
	Title       string
	Technology  string
	Description string
	Tags        []string
}

// nodeRows lists the deployment nodes that say what they are, each by its
// path from the environment ("aws / vpc"). It is empty when none does, so a
// deployment page without node attributes renders as in 1.1.0.
func nodeRows(v *vm.ViewModel) []nodeRow {
	byID := make(map[string]vm.Node, len(v.Nodes))
	for _, n := range v.Nodes {
		byID[n.ID] = n
	}
	var rows []nodeRow
	for _, n := range v.Nodes {
		if n.Role != vm.RoleGroup || n.Title == "" && n.Technology == "" && n.Description == "" && len(n.Tags) == 0 {
			continue
		}
		path := []string{n.Label}
		for p := byID[n.Parent]; p.Role == vm.RoleGroup; p = byID[p.Parent] {
			path = append([]string{p.Label}, path...)
		}
		rows = append(rows, nodeRow{Path: strings.Join(path, " / "), Title: n.Title,
			Technology: n.Technology, Description: n.Description, Tags: n.Tags})
	}
	return rows
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

// arrows are the Talks to direction glyphs and their spoken form.
var arrows = map[string][2]string{
	vm.DirectionOut:  {"→", "calls"},
	vm.DirectionIn:   {"←", "called by"},
	vm.DirectionBoth: {"↔", "calls and is called by"},
}

// anyTags reports whether any row of a relation table carries tags; the Tags
// column appears only then, so untagged projects render exactly as before.
// It accepts 014's relation rows and 021's neighbour rows.
func anyTags(rows any) bool {
	switch rs := rows.(type) {
	case []vm.RelationRow:
		return slices.ContainsFunc(rs, func(r vm.RelationRow) bool { return len(r.Tags) > 0 })
	case []vm.NeighbourRow:
		return slices.ContainsFunc(rs, func(r vm.NeighbourRow) bool { return len(r.Tags) > 0 })
	}
	return false
}

// hashTags renders tags as "#a #b".
func hashTags(tags []string) string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = "#" + t
	}
	return strings.Join(out, " ")
}
