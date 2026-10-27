package d2

import (
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

// textY returns the y of the first SVG <text> whose content starts with s.
func textY(t *testing.T, svg []byte, s string) float64 {
	t.Helper()
	m := regexp.MustCompile(`<text[^>]* y="([0-9.\-]+)"[^>]*>(?:<tspan[^>]*>)?` + regexp.QuoteMeta(s)).FindSubmatch(svg)
	if m == nil {
		t.Fatalf("no text %q in the SVG", s)
	}
	y, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		t.Fatal(err)
	}
	return y
}

// TestViewTitleAboveDiagram: under dagre and ELK alike, the declared title is
// drawn above every element (feature 018, R7).
func TestViewTitleAboveDiagram(t *testing.T) {
	t.Setenv("PATH", "")
	for _, layout := range []string{"", "elk"} {
		v := emitCases()["titled_view"]
		v.View.Layout = layout
		as, err := NewSVGBackend().Render(context.Background(), projectionOf(v), usecases.RenderOptions{})
		if err != nil {
			t.Fatal(err)
		}
		svg := as[0].Bytes
		if title, api := textY(t, svg, "Phase 1: ingestion only"), textY(t, svg, "API"); title >= api {
			t.Errorf("layout %q: title at y=%v is not above the API box at y=%v", layout, title, api)
		}
	}
}
