package d2

import (
	"bytes"
	"context"
	"strings"
	"testing"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.Backend = (*SVGBackend)(nil)

// cacheHits reports how many renders the cache served.
func (b *SVGBackend) cacheHits() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.hits
}

var _ usecases.Backend = (*SourceBackend)(nil)

func projectionOf(views ...vm.ViewModel) *vm.Projection {
	for i := range views {
		views[i].DiagramPath = vm.DiagramFile(views[i].View.ID, "svg")
	}
	return &vm.Projection{Views: views}
}

// TestSVGBackend renders in-process with an emptied PATH: no executable is
// looked up and no process is started (FR-014, US6).
func TestSVGBackend(t *testing.T) {
	t.Setenv("PATH", "")
	c := emitCases()
	proj := projectionOf(c["landscape"], c["subject"], c["deployment"])
	b := NewSVGBackend()

	first, err := b.Render(context.Background(), proj, usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 3 {
		t.Fatalf("got %d artifacts, want one per view", len(first))
	}
	for i, a := range first {
		want := vm.DiagramFile(proj.Views[i].View.ID, "svg")
		if a.Path != want || a.Format != vm.FormatSVG {
			t.Errorf("artifact %d = %s (%s), want %s", i, a.Path, a.Format, want)
		}
		s := string(a.Bytes)
		decl, rest, ok := strings.Cut(s, "?>")
		if !ok || !strings.HasPrefix(decl, "<?xml") {
			t.Fatalf("%s: no XML declaration", a.Path)
		}
		notice := "<!-- " + vm.NoticeText(proj.Views[i].Sources) + " -->"
		if !strings.HasPrefix(rest, notice) {
			t.Errorf("%s: the notice must follow the XML declaration immediately; got %.120q", a.Path, rest)
		}
	}
	// R6 guard: d2 writes classes onto the SVG, so user CSS can target tags.
	sub := string(first[1].Bytes)
	for _, class := range []string{"kind-container", "tag-pci", "tag-edge"} {
		if !strings.Contains(sub, class) {
			t.Errorf("SVG lacks class %q: d2 no longer emits classes (research R6)", class)
		}
	}

	hits := b.cacheHits()
	second, err := b.Render(context.Background(), proj, usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for i := range first {
		if !bytes.Equal(first[i].Bytes, second[i].Bytes) {
			t.Errorf("%s differs between renders (FR-021)", first[i].Path)
		}
	}
	if b.cacheHits()-hits != 3 {
		t.Errorf("second render: %d cache hits, want 3", b.cacheHits()-hits)
	}

	fresh, err := NewSVGBackend().Render(context.Background(), proj, usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for i := range first {
		if !bytes.Equal(first[i].Bytes, fresh[i].Bytes) {
			t.Errorf("%s differs across backend instances (FR-021)", first[i].Path)
		}
	}
}

func TestSourceBackend(t *testing.T) {
	t.Parallel()
	c := emitCases()
	as, err := NewSourceBackend().Render(context.Background(), projectionOf(c["landscape"]), usecases.RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(as) != 1 || as[0].Path != "diagrams/landscape.d2" || !bytes.Equal(as[0].Bytes, Emit(c["landscape"])) {
		t.Fatalf("SourceBackend = %+v", as)
	}
}
