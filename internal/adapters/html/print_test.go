package html

import (
	"strings"
	"testing"
)

func mediaBlock(t *testing.T, css, query string) string {
	t.Helper()
	i := strings.Index(css, "@media "+query+" {")
	if i < 0 {
		t.Fatalf("style.css has no @media %s block", query)
	}
	return css[i : i+strings.Index(css[i:], "\n}")]
}

// TestPrintStyles is spec FR-037: light palette, no chrome, the rail after the
// prose, nothing split across pages.
func TestPrintStyles(t *testing.T) {
	t.Parallel()
	_, _, css := themeTokens(t)
	p := mediaBlock(t, css, "print")
	for _, want := range []string{
		`:root, :root[data-theme="dark"] {`, "--bg: #FFFFFF", "--link: #2D66AE",
		".sidebar, .topbar, .menu-toggle, #theme-toggle, nav.pager, .diagram-toolbar",
		".element-layout { display: block; }", // block flow: header → main → rail
		".rail-card, .card, .stat-card, .row, tr, figure { break-inside: avoid; }",
		".diagram-panel svg { width: 100%; max-height: none; }",
		"background-image: none", "--shadow: none",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("print block lacks %q", want)
		}
	}
}

// TestNarrowLayout is spec FR-038: one column with the rail before the prose,
// the sidebar behind the menu button, long ids wrapping.
func TestNarrowLayout(t *testing.T) {
	t.Parallel()
	_, _, css := themeTokens(t)
	n := mediaBlock(t, css, "(max-width: 760px)")
	for _, want := range []string{
		`grid-template-areas: "rail" "main";`,
		"html.js body.nav-open .sidebar",
		".meta, .row-head, .breadcrumb { overflow-wrap: anywhere; }",
		"html.js .menu-toggle { display: inline-block; }",
		"html:not(.js) .sidebar { position: static;",
	} {
		if !strings.Contains(n, want) {
			t.Errorf("narrow block lacks %q", want)
		}
	}
}
