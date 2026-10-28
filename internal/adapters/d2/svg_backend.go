package d2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"sync"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2fonts"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SVGBackend is the "svg" backend. It emits D2 source, then compiles, lays
// out (dagre by default, or ELK; both embedded) and renders it inside this
// process (FR-014). One pass over the rendered SVG then adds the cards' icons
// and the label line styles from the emitter's label plan (feature 020).
//
// Rendered SVGs are cached by the SHA-256 of their D2 source and label plan
// for the life of the backend. A one-shot build gains nothing; under `loko
// serve` an edit re-lays-out only the views whose source changed (research
// R2). The plan is part of the key because line roles cannot be read back
// from the source text (feature 020, R10). The cache is never written to disk.
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
	src, plan := emitPlanned(v)
	svg, err := b.cached(cacheKey(src, plan), func() ([]byte, error) {
		raw, fontHash, err := compile(ctx, ruler, src, plan.hasDescription())
		if err != nil {
			return nil, err
		}
		return applyPlan(raw, plan, fontHash), nil
	})
	if err != nil {
		return vm.Artifact{}, err
	}
	w, h := svgSize(svg)
	return vm.Artifact{
		Path:   vm.DiagramFile(v.View.ID, "svg"),
		Format: vm.FormatSVG,
		Bytes:  withNotice(svg, vm.NoticeText(v.Sources)),
		Owner:  ownerOf(v),
		Width:  w,
		Height: h,
	}, nil
}

// viewBox matches the outer svg element's viewBox, which d2svg sets to the
// laid-out diagram's bounding box plus padding.
var viewBox = regexp.MustCompile(`<svg[^>]* viewBox="[-0-9.]+ [-0-9.]+ ([0-9.]+) ([0-9.]+)"`)

// svgSize reads a rendered diagram's size from its SVG, so the wide-view hint
// needs no second layout run (feature 018, R10). It reads the cached bytes,
// so a cache hit reports the size too. Zero when there is no viewBox.
func svgSize(svg []byte) (int, int) {
	m := viewBox.FindSubmatch(svg)
	if m == nil {
		return 0, 0
	}
	w, errW := strconv.ParseFloat(string(m[1]), 64)
	h, errH := strconv.ParseFloat(string(m[2]), 64)
	if errW != nil || errH != nil {
		return 0, 0
	}
	return int(math.Round(w)), int(math.Round(h))
}

// cacheKey identifies a rendered SVG: the D2 source and the label plan.
func cacheKey(src []byte, plan LabelPlan) [sha256.Size]byte {
	h := sha256.New()
	h.Write(src)
	h.Write([]byte{0})
	h.Write(plan.Key())
	return [sha256.Size]byte(h.Sum(nil))
}

func (b *SVGBackend) cached(key [sha256.Size]byte, render func() ([]byte, error)) ([]byte, error) {
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

// compile lays out and renders D2 source. It returns the SVG and the prefix
// of its d2 font families. With regularFont, the SVG embeds d2's regular face
// even when d2 drew no regular text, because descriptions use it (020 R3).
func compile(ctx context.Context, ruler *textmeasure.Ruler, src []byte, regularFont bool) ([]byte, string, error) {
	opts := &d2svg.RenderOpts{
		Pad:         new(int64(d2svg.DEFAULT_PADDING)),
		ThemeID:     new(int64(0)),
		OmitVersion: new(true),
	}
	diagram, _, err := d2lib.Compile(log.WithDefault(ctx), string(src),
		&d2lib.CompileOptions{LayoutResolver: layoutEngine, Ruler: ruler}, opts)
	if err != nil {
		return nil, "", fmt.Errorf("compiling d2: %w", err)
	}
	svg, err := d2svg.Render(diagram, opts)
	if err != nil {
		return nil, "", fmt.Errorf("rendering svg: %w", err)
	}
	m := diagramHash.FindSubmatch(svg)
	if m == nil {
		return nil, "", fmt.Errorf("rendering svg: no diagram hash")
	}
	fontHash := string(m[1])
	if regularFont {
		family := d2fonts.SourceSansPro
		if diagram.FontFamily != nil {
			family = *diagram.FontFamily
		}
		svg = withRegularFace(svg, fontHash, family.Font(0, d2fonts.FONT_STYLE_REGULAR).GetEncodedSubset(diagram.GetCorpus()))
	}
	return svg, fontHash, nil
}

// diagramHash matches the class d2 gives its inner svg, which also prefixes
// its font family names.
var diagramHash = regexp.MustCompile(`<svg class="(d2-[0-9]+) `)

// withRegularFace adds d2's regular @font-face, in d2's own format, to the
// SVG's style sheet when d2 did not embed it.
func withRegularFace(svg []byte, fontHash, dataURL string) []byte {
	if bytes.Contains(svg, []byte("font-family: "+fontHash+"-font-regular;")) {
		return svg
	}
	end := bytes.Index(svg, []byte("]]></style>"))
	if end < 0 {
		return svg
	}
	face := "\n@font-face {\n\tfont-family: " + fontHash + "-font-regular;\n\tsrc: url(\"" + dataURL + "\");\n}"
	return slices.Concat(svg[:end], []byte(face), svg[end:])
}

// layoutEngine resolves the engine a view's D2 source names (its d2-config
// layout-engine; dagre when unset). Dagre is the default because d2 v0.7.1
// builds a fresh JS runtime and recompiles the layout engine on every call,
// once per nesting level, and ELK's engine is the heavier of the two. Measured
// on 1,020 elements: ELK 13.7 s, dagre 2.8 s against a 10 s budget (SC-006,
// research R1/R2), so ELK is opt-in per project or view (ADR-0015).
func layoutEngine(engine string) (d2graph.LayoutGraph, error) {
	switch engine {
	case "dagre":
		return d2dagrelayout.DefaultLayout, nil
	case "elk":
		return d2elklayout.DefaultLayout, nil
	default:
		return nil, fmt.Errorf("unknown layout engine %q", engine)
	}
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
