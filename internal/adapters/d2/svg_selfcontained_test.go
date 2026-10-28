package d2

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// TestSVGSelfContained is User Story 2: a rendered SVG fetches nothing, runs
// nothing, and keeps its labels as selectable text (FR-004, FR-007).
func TestSVGSelfContained(t *testing.T) {
	t.Setenv("PATH", "")
	c := emitCases()
	as, err := NewSVGBackend().Render(context.Background(),
		projectionOf(c["landscape"], c["subject"], c["deployment"], sequenceCase()), usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	hrefs := regexp.MustCompile(`href="([^"]*)"`)
	uses := regexp.MustCompile(`<use [^>]*href="#(loko-icon-[a-z-]+)"`)
	urls := regexp.MustCompile(`url\(([^)]*)\)`)
	for _, a := range as {
		s := string(a.Bytes)
		for _, bad := range []string{"<foreignObject", "<script", "<image", "@import"} {
			if strings.Contains(s, bad) {
				t.Errorf("%s contains %s", a.Path, bad)
			}
		}
		for _, h := range hrefs.FindAllStringSubmatch(s, -1) {
			if !strings.HasPrefix(h[1], "#") {
				t.Errorf("%s references %q", a.Path, h[1])
			}
		}
		for _, u := range urls.FindAllStringSubmatch(s, -1) {
			if strings.Contains(u[1], "http") {
				t.Errorf("%s loads %.60q", a.Path, u[1])
			}
		}
		for _, u := range uses.FindAllStringSubmatch(s, -1) {
			if !strings.Contains(s, `<symbol id="`+u[1]+`"`) {
				t.Errorf("%s draws %s, which it does not define", a.Path, u[1])
			}
		}
	}
	sub := string(as[1].Bytes)
	for _, text := range []string{">API</tspan>", ">CONTAINER · Go</tspan>", `>Says &#34;hi&#34;: costs $5\nand \ more</tspan>`} {
		if !strings.Contains(sub, text) {
			t.Errorf("label text %q is not a selectable tspan", text)
		}
	}
}
