package usecases

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Diagnostics lets outer layers name the diagnostics type without importing
// entities (Constitution, Dependency Direction).
type Diagnostics = arch.Diagnostics

// ServeDeps are the ports `loko serve` uses.
type ServeDeps struct {
	Build   BuildDeps
	Watcher ChangeWatcher
	Preview PreviewServer
	// Describe renders diagnostics as text with file, line and column. It is
	// supplied by the caller so core stays free of presentation.
	Describe func(Diagnostics) string
}

// ServeRequest names what to serve.
type ServeRequest struct {
	Root         string
	BuildVersion string
	// OutDir is the build output directory; it is never watched (FR-032) and
	// never written.
	OutDir string
}

// Serve builds the site in memory, publishes it, and rebuilds once per
// settled burst of source changes until ctx is cancelled (FR-029, FR-033).
//
// A failed compile is shown, not fatal: the preview switches to its error
// state and the next good build recovers it without a restart (FR-030,
// FR-031). An operational failure of the first build is returned. Serve renders html and what it requires, and never commits.
func Serve(ctx context.Context, deps ServeDeps, req ServeRequest) error {
	outAbs, err := filepath.Abs(req.OutDir)
	if err != nil {
		return fmt.Errorf("serve: resolving %s: %w", req.OutDir, err)
	}
	var mu sync.Mutex
	var prose []string
	changes, err := deps.Watcher.Watch(ctx, WatchSpec{
		Root:    req.Root,
		Exclude: []string{outAbs},
		ExtraFiles: func() []string {
			mu.Lock()
			defer mu.Unlock()
			return append([]string(nil), prose...)
		},
	})
	if err != nil {
		return fmt.Errorf("serve: watching %s: %w", req.Root, err)
	}
	// rebuild returns only operational errors. The first build's is fatal —
	// an unreadable project root cannot be fixed by editing a file, so serve
	// exits (contracts/cli.md § serve); later ones are shown in the browser.
	rebuild := func() error {
		res, err := BuildArtifacts(ctx, deps.Build, BuildRequest{
			Root: req.Root, BuildVersion: req.BuildVersion, Formats: []string{"html"},
		})
		switch {
		case err != nil:
			return err
		case res.Diags.HasErrors():
			deps.Preview.Fail(deps.Describe(res.Diags))
		case res.NothingToDraw:
			deps.Preview.Fail("nothing to draw: the architecture declares no elements")
		default:
			mu.Lock()
			prose = res.ProseFiles
			mu.Unlock()
			deps.Preview.Publish(res.Artifacts)
		}
		return nil
	}
	if err := rebuild(); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-changes:
			if !ok {
				return nil
			}
			if err := rebuild(); err != nil {
				deps.Preview.Fail(err.Error())
			}
		}
	}
}
