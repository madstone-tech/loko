package html

import (
	"strings"
	"testing"
)

func TestRenderProse(t *testing.T) {
	t.Parallel()
	in := "# Title\n\nSome **bold**.\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n<script>alert(1)</script>\n"
	got := string(renderProse(in))
	for _, want := range []string{"<h1>Title</h1>", "<strong>bold</strong>", "<table>", "<td>1</td>"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered prose lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "<script>") {
		t.Errorf("raw HTML in prose must not pass through:\n%s", got)
	}
	if string(renderProse(in)) != got {
		t.Error("prose rendering is not deterministic")
	}
}
