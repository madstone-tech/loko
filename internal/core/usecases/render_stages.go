package usecases

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// renderAll runs every backend concurrently (feature 021, research R1). A
// backend whose format requires others gets RenderOptions.Upstream, which
// blocks until the backends that require nothing have finished and returns
// their artifacts in the formats it requires. Results are collected by index:
// those backends first, then the rest, each in registry order, so the
// combined order never depends on which backend finishes first.
func renderAll(ctx context.Context, backends []Backend, proj *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error) {
	var first, second []Backend
	for _, b := range backends {
		if len(b.Format().Requires()) == 0 {
			first = append(first, b)
		} else {
			second = append(second, b)
		}
	}
	var (
		upstream    []viewmodel.Artifact
		upstreamErr error
		done        = make(chan struct{})
		rest        []viewmodel.Artifact
		restErr     error
		wg          sync.WaitGroup
	)
	wg.Go(func() {
		rest, restErr = renderStage(ctx, second, proj, func(b Backend) RenderOptions {
			o := opts
			o.Upstream = sync.OnceValues(func() (map[string][]byte, error) {
				<-done
				if upstreamErr != nil {
					return nil, upstreamErr
				}
				return upstreamFor(b.Format(), upstream), nil
			})
			return o
		})
	})
	upstream, upstreamErr = renderStage(ctx, first, proj, func(Backend) RenderOptions { return opts })
	close(done)
	wg.Wait()
	if upstreamErr != nil {
		return nil, upstreamErr
	}
	if restErr != nil {
		return nil, restErr
	}
	return append(upstream, rest...), nil
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
