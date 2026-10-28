package d2

import (
	"context"
	"encoding/base64"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/lib/svg"
	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// rendered is one view through the emitter, d2 and the SVG pass.
type rendered struct {
	raw, doc string // before and after the pass
	plan     LabelPlan
	hash     string
}

func renderPlanned(t *testing.T, m vm.ViewModel) rendered {
	t.Helper()
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	src, plan := emitPlanned(m)
	raw, hash, err := compile(context.Background(), ruler, src, plan.hasDescription())
	if err != nil {
		t.Fatal(err)
	}
	if want := regexp.MustCompile(`<svg class="(d2-\d+) d2-svg"`).FindStringSubmatch(string(raw)); want == nil || want[1] != hash {
		t.Fatalf("font hash %q does not match the SVG's %v", hash, want)
	}
	return rendered{raw: string(raw), doc: string(applyPlan(raw, plan, hash)), plan: plan, hash: hash}
}

// groupOf returns the element group d2 drew for an absolute ID, or "".
func groupOf(doc, id string) string {
	open := `<g class="` + base64.URLEncoding.EncodeToString([]byte(svg.EscapeText(id)))
	i := strings.Index(doc, open+`"`)
	if i < 0 {
		i = strings.Index(doc, open+` `)
	}
	if i < 0 {
		return ""
	}
	depth := 0
	for j := i; j < len(doc); {
		switch {
		case strings.HasPrefix(doc[j:], "<g"):
			depth++
			j += 2
		case strings.HasPrefix(doc[j:], "</g>"):
			depth--
			j += 4
			if depth == 0 {
				return doc[i:j]
			}
		default:
			j++
		}
	}
	return ""
}

func attr(tag, name string) string {
	m := regexp.MustCompile(` ` + name + `="([^"]*)"`).FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	return m[1]
}

func firstTag(s, name string) string {
	i := strings.Index(s, "<"+name+" ")
	if i < 0 {
		return ""
	}
	return s[i : i+strings.Index(s[i:], ">")+1]
}

func tspans(group string) []string {
	return regexp.MustCompile(`<tspan[^>]*>`).FindAllString(group, -1)
}

func num(t *testing.T, s string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("not a number: %q", s)
	}
	return f
}

// checkCards is contracts/svg-output.md items 1 to 3 and 6 for every card in
// the plan, and item 2's "none" for the groups that are not cards.
func checkCards(t *testing.T, r rendered, notCards ...string) {
	t.Helper()
	used := map[string]bool{}
	for id, p := range r.plan.Nodes {
		used[p.Icon] = true
		g := groupOf(r.doc, id)
		if g == "" {
			t.Errorf("%s: no group in the SVG", id)
			continue
		}
		if n := strings.Count(g, "<use "); n != 1 {
			t.Errorf("%s: %d icons, want 1", id, n)
		}
		use, rect := firstTag(g, "use"), firstTag(g, "rect")
		if attr(use, "href") != "#loko-icon-"+p.Icon || attr(use, "class") != "loko-icon" ||
			attr(use, "width") != "16" || attr(use, "height") != "16" || attr(use, "color") != p.Colour {
			t.Errorf("%s: icon = %s", id, use)
		}
		if num(t, attr(use, "x")) != num(t, attr(rect, "x"))+4 || num(t, attr(use, "y")) != num(t, attr(rect, "y"))+4 {
			t.Errorf("%s: icon %s is not 4px inside %s", id, use, rect)
		}
		if num(t, attr(rect, "rx")) != 6 {
			t.Errorf("%s: card corners = %s", id, rect)
		}
		spans := tspans(g)
		if len(spans) != len(p.Lines) {
			t.Fatalf("%s: %d tspans for %d roles", id, len(spans), len(p.Lines))
		}
		for i, role := range p.Lines {
			if got, want := attr(spans[i], "style"), role.inlineStyle(r.hash); got != want {
				t.Errorf("%s line %d: style %q, want %q", id, i, got, want)
			}
		}
	}
	for _, id := range notCards {
		if g := groupOf(r.doc, id); g == "" || strings.Contains(g, "<use") {
			t.Errorf("%s: not a card, but has an icon or no group: %.200s", id, g)
		}
	}
	defs := regexp.MustCompile(`<symbol id="loko-icon-([a-z-]+)"`).FindAllStringSubmatch(r.doc, -1)
	var got []string
	for _, d := range defs {
		got = append(got, d[1])
	}
	if want := slices.Sorted(maps.Keys(used)); !slices.Equal(got, want) {
		t.Errorf("symbols = %v, want %v", got, want)
	}
	if n := strings.Count(r.doc, "<defs>"); len(used) > 0 && n != 1 {
		t.Errorf("%d icon <defs>, want 1", n)
	}
	if strings.Contains(r.doc, "font-weight") {
		t.Error("a line sets font-weight; d2 selects weights by family (R3)")
	}
}

func TestSVGPassCards(t *testing.T) {
	t.Parallel()
	subject := emitCases()["subject"]
	subject.Nodes[1].Style = vm.WithShape(subject.Nodes[1].Style, "database")
	r := renderPlanned(t, subject)
	checkCards(t, r, "system__shop", "outside")
	if g := groupOf(r.doc, "system__shop.container__api"); !strings.Contains(g, "kind-container tag-edge tag-pci") {
		t.Errorf("api classes changed: %.120s", g)
	}
	if g := groupOf(r.doc, "system__shop.container__db"); !strings.Contains(g, "kind-container shape-database") ||
		!strings.Contains(g, "#loko-icon-database") {
		t.Errorf("db group: %.200s", g)
	}
	reads := tspans(groupOf(r.doc, "system__shop.(container__api -> container__db)[0]"))
	if len(reads) != 2 || attr(reads[0], "style") != "" || attr(reads[1], "style") != roleEdgeTech.inlineStyle(r.hash) {
		t.Errorf("edge label lines = %v", reads)
	}

	checkCards(t, renderPlanned(t, emitCases()["landscape"]))
	checkCards(t, renderPlanned(t, emitCases()["deployment"]),
		"deployment__prod", "deployment__prod.deployment__prod__node__a")
	// Sequence participants keep their icons: the icon fits inside the card,
	// clear of the name and the lifeline (FR-018, checked by eye in T019).
	seq := renderPlanned(t, sequenceCase())
	checkCards(t, seq)
	if len(seq.plan.Nodes) != 3 || len(seq.plan.Edges) != 2 {
		t.Errorf("sequence plan: %d cards, %d styled steps; want 3 and 2", len(seq.plan.Nodes), len(seq.plan.Edges))
	}
}

// TestSVGPassRegularFace: descriptions use d2's regular face, which the SVG
// must embed even when d2 drew no regular text of its own (research R3).
func TestSVGPassRegularFace(t *testing.T) {
	t.Parallel()
	m := vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{titledNode("")}}
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		t.Fatal(err)
	}
	plain, hash, err := compile(context.Background(), ruler, Emit(m), false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(plain), hash+"-font-regular;") {
		t.Fatal("fixture has regular text already; it no longer tests the added face")
	}
	r := renderPlanned(t, m)
	if !strings.Contains(r.doc, "@font-face {\n\tfont-family: "+r.hash+"-font-regular;\n\tsrc: url(\"data:") {
		t.Error("the regular face is not embedded")
	}
	checkCards(t, r)
}

func TestSVGPassNoCards(t *testing.T) {
	t.Parallel()
	env := node("deployment__prod", "subject", "", "prod", "")
	r := renderPlanned(t, vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{env}})
	if r.doc != r.raw {
		t.Error("a view with no cards was changed by the pass")
	}
}

// TestCardGeometry is FR-012 and User Story 4: the icon adds no padding, sits
// above the name, and the text does not leave a larger band below it than
// above it.
func TestCardGeometry(t *testing.T) {
	t.Parallel()
	r := renderPlanned(t, vm.ViewModel{View: vm.View{ID: "v"}, Nodes: []vm.Node{titledNode("")}})
	box := regexp.MustCompile(`viewBox="[^"]*"`)
	if box.FindString(r.doc) != box.FindString(r.raw) {
		t.Errorf("the pass changed the diagram's size")
	}
	g := groupOf(r.doc, "container__api")
	rect, text := firstTag(g, "rect"), firstTag(g, "text")
	if rect != firstTag(groupOf(r.raw, "container__api"), "rect") {
		t.Errorf("the pass changed the card: %s", rect)
	}
	top, height := num(t, attr(rect, "y")), num(t, attr(rect, "height"))
	baseline := num(t, attr(text, "y"))
	last := baseline
	for _, s := range tspans(g) {
		last += num(t, attr(s, "dy"))
	}
	const ascent = 12 // the name's glyph height above its baseline at 16px
	nameTop := baseline - ascent
	if iconBottom := num(t, attr(firstTag(g, "use"), "y")) + 16; iconBottom > nameTop {
		t.Errorf("icon ends at %.1f, below the name's top at %.1f", iconBottom, nameTop)
	}
	if below, above := top+height-last, nameTop-top; below > above {
		t.Errorf("%.1fpx below the text, more than the %.1fpx above it", below, above)
	}
}

// TestSVGPassAddressAttribute is feature 021's contracts/svg-address.md: every
// card carries its model address right after its class attribute, once, and
// nothing else carries one.
func TestSVGPassAddressAttribute(t *testing.T) {
	t.Parallel()
	cases := emitCases()
	cases["sequence"] = sequenceCase()
	for name, m := range cases {
		r := renderPlanned(t, m)
		for id, p := range r.plan.Nodes {
			if p.Address == "" {
				t.Errorf("%s: card %s has no address in the plan", name, id)
				continue
			}
			attr := `data-loko-address="` + svg.EscapeText(p.Address) + `"`
			g := groupOf(r.doc, id)
			open := g[:strings.IndexByte(g, '>')]
			if !strings.HasSuffix(open, `" `+attr) || strings.Count(open, `class="`) != 1 {
				t.Errorf("%s: %s start tag = %q, want the attribute right after class", name, id, open)
			}
			if n := strings.Count(r.doc, attr); n != 1 {
				t.Errorf("%s: %s appears %d times, want 1", name, attr, n)
			}
		}
		if n := strings.Count(r.doc, "data-loko-address="); n != len(r.plan.Nodes) {
			t.Errorf("%s: %d addressed groups, want %d (cards only)", name, n, len(r.plan.Nodes))
		}
	}
}
