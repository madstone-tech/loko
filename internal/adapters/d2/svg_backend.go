package d2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"
	"sync"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SVGBackend is the "svg" backend. It emits D2 source, then compiles, lays
// out (dagre, embedded) and renders it inside this process (FR-014).
//
// Rendered SVGs are cached by the SHA-256 of their D2 source for the life of
// the backend. A one-shot build gains nothing; under `loko serve` an edit
// re-lays-out only the views whose source changed (research R2). The cache is
// never written to disk.
type SVGBackend struct {
	mu    sync.Mutex
	cache map[[sha256.Size]byte][]byte
	hits  int
}

// NewSVGBackend returns the "svg" backend.
func NewSVGBackend() *SVGBackend {
	return &SVGBackend{cache: map[[sha256.Size]byte][]byte{}}
}

// Format implements usecases.Backend.
func (*SVGBackend) Format() vm.Format { return vm.FormatSVG }

// Render implements usecases.Backend. Views are laid out on a bounded worker
// pool; each result is stored at its view's index, so output order never
// depends on which worker finishes first. Every worker exits before Render
// returns.
func (b *SVGBackend) Render(ctx context.Context, in *vm.Projection, _ usecases.RenderOptions) ([]vm.Artifact, error) {
	out := make([]vm.Artifact, len(in.Views))
	errs := make([]error, len(in.Views))
	jobs := make(chan int)
	var wg sync.WaitGroup
	workers := min(runtime.GOMAXPROCS(0), max(len(in.Views), 1))
	for range workers {
		wg.Go(func() {
			var ruler *textmeasure.Ruler
			for i := range jobs {
				if ruler == nil {
					r, err := textmeasure.NewRuler()
					if err != nil {
						errs[i] = fmt.Errorf("text measurement: %w", err)
						continue
					}
					ruler = r
				}
				out[i], errs[i] = b.renderView(ctx, ruler, in.Views[i])
			}
		})
	}
	for i := range in.Views {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("view %s: %w", in.Views[i].View.ID, err)
		}
	}
	return out, nil
}

func (b *SVGBackend) renderView(ctx context.Context, ruler *textmeasure.Ruler, v vm.ViewModel) (vm.Artifact, error) {
	src := Emit(v)
	svg, err := b.cached(src, func() ([]byte, error) { return compile(ctx, ruler, src) })
	if err != nil {
		return vm.Artifact{}, err
	}
	return vm.Artifact{
		Path:   vm.DiagramFile(v.View.ID, "svg"),
		Format: vm.FormatSVG,
		Bytes:  withNotice(svg, vm.NoticeText(v.Sources)),
		Owner:  ownerOf(v),
	}, nil
}

func (b *SVGBackend) cached(src []byte, render func() ([]byte, error)) ([]byte, error) {
	key := sha256.Sum256(src)
	b.mu.Lock()
	if svg, ok := b.cache[key]; ok {
		b.hits++
		b.mu.Unlock()
		return svg, nil
	}
	b.mu.Unlock()
	svg, err := render()
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.cache[key] = svg
	b.mu.Unlock()
	return svg, nil
}

func compile(ctx context.Context, ruler *textmeasure.Ruler, src []byte) ([]byte, error) {
	pad := int64(d2svg.DEFAULT_PADDING)
	theme := int64(0)
	omitVersion := true
	opts := &d2svg.RenderOpts{Pad: &pad, ThemeID: &theme, OmitVersion: &omitVersion}
	// dagre, not ELK: d2 v0.7.1 builds a fresh JS runtime and recompiles the
	// layout engine on every call, once per nesting level, and ELK's engine is
	// the heavier of the two. Measured on 1,020 elements: ELK 13.7 s, dagre
	// 2.8 s against a 10 s budget (SC-006, research R1/R2).
	layout := func(string) (d2graph.LayoutGraph, error) { return d2dagrelayout.DefaultLayout, nil }
	diagram, _, err := d2lib.Compile(log.WithDefault(ctx), string(src),
		&d2lib.CompileOptions{LayoutResolver: layout, Ruler: ruler}, opts)
	if err != nil {
		return nil, fmt.Errorf("compiling d2: %w", err)
	}
	svg, err := d2svg.Render(diagram, opts)
	if err != nil {
		return nil, fmt.Errorf("rendering svg: %w", err)
	}
	return svg, nil
}

// withNotice places the generated-file notice immediately after the XML
// declaration, which must stay first (research R7).
func withNotice(svg []byte, notice string) []byte {
	comment := []byte("<!-- " + notice + " -->")
	if i := bytes.Index(svg, []byte("?>")); i >= 0 && bytes.HasPrefix(svg, []byte("<?xml")) {
		out := make([]byte, 0, len(svg)+len(comment))
		out = append(out, svg[:i+2]...)
		out = append(out, comment...)
		return append(out, svg[i+2:]...)
	}
	return append(comment, svg...)
}
