package viewmodel

import (
	"math/rand"
	"testing"
)

func TestSegment(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"api", "api"},
		{"orders-db.v2", "orders-db.v2"},
		{"payments/v2", "payments_2fv2"},
		{"a_b", "a_5fb"},
		{"a b", "a_20b"},
		{"_Odd_name-1", "_5fOdd_5fname-1"},
		{"Ünï", "_c3_9cn_c3_af"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Segment(tt.in); got != tt.want {
			t.Errorf("Segment(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestSegmentInjective is FR-006 / R9: two different names can never share a
// file name, so a collision is only ever the case-folding kind.
func TestSegmentInjective(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	seen := map[string]string{}
	for range 1000 {
		b := make([]byte, 1+r.Intn(6))
		for j := range b {
			b[j] = "aA_-./ %z9"[r.Intn(10)]
		}
		in := string(b)
		out := Segment(in)
		if prev, ok := seen[out]; ok && prev != in {
			t.Fatalf("Segment(%q) == Segment(%q) == %q", in, prev, out)
		}
		seen[out] = in
	}
}

func TestPaths(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, got, want string }{
		{"diagram svg", DiagramFile("landscape", "svg"), "diagrams/landscape.svg"},
		{"diagram d2 escaped", DiagramFile("a_b", "d2"), "diagrams/a_5fb.d2"},
		{"view page", ViewPage("system-shop"), "view/system-shop.html"},
		{"md view page", MarkdownViewPage("system-shop"), "md/view/system-shop.md"},
		{"element html", ElementPath("container.api", "html"), "element/container/api.html"},
		{"element md", ElementPath("container.api", "md"), "md/element/container/api.md"},
		{"element escaped", ElementPath("system._Odd", "html"), "element/system/_5fOdd.html"},
		{"index", IndexPage("html"), "index.html"},
		{"md index", IndexPage("md"), "md/index.md"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestRel(t *testing.T) {
	t.Parallel()
	tests := []struct{ from, to, want string }{
		{"element/container/api.html", "diagrams/x.svg", "../../diagrams/x.svg"},
		{"index.html", "diagrams/x.svg", "diagrams/x.svg"},
		{"view/a.html", "view/b.html", "b.html"},
		{"view/a.html", "index.html", "../index.html"},
		{"md/element/system/s.md", "diagrams/x.svg", "../../../diagrams/x.svg"},
		{"md/element/system/s.md", "md/element/container/c.md", "../container/c.md"},
		{"element/system/s.html", "element/system/s.html", "s.html"},
	}
	for _, tt := range tests {
		if got := Rel(tt.from, tt.to); got != tt.want {
			t.Errorf("Rel(%q, %q) = %q, want %q", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestNodeIDHasNoDots(t *testing.T) {
	t.Parallel()
	// A dot is D2's nesting operator, so a node key must never contain one.
	if got := NodeIDFor("container.api"); got != "container__api" {
		t.Errorf("NodeIDFor = %q", got)
	}
	if got := NodeIDFor("deployment.prod.instance.a_b"); got != "deployment__prod__instance__a_5fb" {
		t.Errorf("NodeIDFor = %q", got)
	}
}
