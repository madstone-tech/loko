package html

import (
	"regexp"
	"strings"
	"testing"
)

var (
	figureRe   = regexp.MustCompile(`(?s)<figure[ >].*?</figure>`)
	labelledBy = regexp.MustCompile(`<svg role="img" aria-labelledby="([^"]+)"`)
)

// TestFiguresHaveCaptions is spec FR-036: every diagram sits in a figure with
// a caption, and an inline SVG is labelled by a caption on the same page.
func TestFiguresHaveCaptions(t *testing.T) {
	t.Parallel()
	for path, body := range renderWith(t, loadProjection(t)) {
		if !strings.HasSuffix(path, ".html") {
			continue
		}
		for _, f := range figureRe.FindAllString(body, -1) {
			if !strings.Contains(f, "<figcaption") {
				t.Errorf("%s: a figure has no caption", path)
			}
		}
		for _, m := range labelledBy.FindAllStringSubmatch(body, -1) {
			if !strings.Contains(body, `<figcaption id="`+m[1]+`"`) {
				t.Errorf("%s: the SVG is labelled by %q, which is not on the page", path, m[1])
			}
		}
	}
}

// TestKindNotColourOnly is spec FR-035: every icon tile in a row comes with
// the kind as text.
func TestKindNotColourOnly(t *testing.T) {
	t.Parallel()
	row := regexp.MustCompile(`(?s)<li class="row[^"]*"[^>]*>(.*?)</li>`)
	for path, body := range renderWith(t, loadProjection(t)) {
		for _, m := range row.FindAllStringSubmatch(body, -1) {
			if strings.Contains(m[1], "icon-tile") && !strings.Contains(m[1], `class="row-kind"`) {
				t.Errorf("%s: a row shows a kind tile without the kind in words: %.160s", path, m[1])
			}
		}
	}
}

// TestReducedMotion is spec FR-033: the cursor blinks three times at most and
// not at all for a reader who prefers reduced motion.
func TestReducedMotion(t *testing.T) {
	t.Parallel()
	_, _, css := themeTokens(t)
	if !regexp.MustCompile(`\.cursor \{[^}]*animation: loko-blink 1s steps\(1\) 3;`).MatchString(css) {
		t.Error("the cursor animation must stop after three blinks")
	}
	i := strings.Index(css, "@media (prefers-reduced-motion: reduce)")
	if i < 0 {
		t.Fatal("no reduced-motion block")
	}
	block := css[i : i+strings.Index(css[i:], "\n}")]
	if !strings.Contains(block, ".cursor { animation: none; }") || !strings.Contains(block, "transition: none !important") {
		t.Errorf("reduced-motion block: %s", block)
	}
}

// TestFocusVisible is spec FR-034: focus is always visible, and no rule hides
// the outline.
func TestFocusVisible(t *testing.T) {
	t.Parallel()
	_, _, css := themeTokens(t)
	if !regexp.MustCompile(`(?m)^:focus-visible \{ outline: 2px solid var\(--accent\);`).MatchString(css) {
		t.Error("no universal :focus-visible outline")
	}
	if regexp.MustCompile(`outline:\s*(none|0)\b`).MatchString(css) {
		t.Error("a rule removes the focus outline")
	}
}
