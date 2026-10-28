package usecases

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// renderAll runs the backends in two stages (feature 021, research R1). The
// backends that require no other format run first; each remaining backend then
// receives the artifacts of the formats it requires as RenderOptions.Upstream.
// Within a stage backends run concurrently, and results are collected by index,
// so the combined order never depends on which backend finishes first.
func renderAll(ctx context.Context, backends []Backend, proj *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error) {
	var first, second []Backend
	for _, b := range backends {
		if len(b.Format().Requires()) == 0 {
			first = append(first, b)
		} else {
			second = append(second, b)
		}
	}
	all, err := renderStage(ctx, first, proj, func(Backend) RenderOptions { return opts })
	if err != nil {
		return nil, err
	}
	rest, err := renderStage(ctx, second, proj, func(b Backend) RenderOptions {
		o := opts
		o.Upstream = upstreamFor(b.Format(), all)
		return o
	})
	if err != nil {
		return nil, err
	}
	return append(all, rest...), nil
}

// renderStage runs backends concurrently, each with the options optsFor gives
// it. Every goroutine finishes before it returns.
func renderStage(ctx context.Context, backends []Backend, proj *viewmodel.Projection,
	optsFor func(Backend) RenderOptions) ([]viewmodel.Artifact, error) {

	results := make([][]viewmodel.Artifact, len(backends))
	errs := make([]error, len(backends))
	var wg sync.WaitGroup
	for i, b := range backends {
		o := optsFor(b)
		wg.Go(func() {
			results[i], errs[i] = b.Render(ctx, proj, o)
		})
	}
	wg.Wait()
	var all []viewmodel.Artifact
	for i, b := range backends {
		if errs[i] != nil {
			return nil, fmt.Errorf("render %s: %w", b.Format(), errs[i])
		}
		all = append(all, results[i]...)
	}
	return all, nil
}

// backendFormats lists the formats of the selected backends, in canonical order.
func backendFormats(backends []Backend) []viewmodel.Format {
	out := make([]viewmodel.Format, 0, len(backends))
	for _, f := range viewmodel.AllFormats {
		if backendFor(backends, f) != nil {
			out = append(out, f)
		}
	}
	return out
}

// upstreamFor collects the artifacts of the formats f requires, keyed by path.
func upstreamFor(f viewmodel.Format, artifacts []viewmodel.Artifact) map[string][]byte {
	required := f.Requires()
	out := map[string][]byte{}
	for _, a := range artifacts {
		if slices.Contains(required, a.Format) {
			out[a.Path] = a.Bytes
		}
	}
	return out
}
