package usecases

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// switchSource returns whichever model it currently holds, so a test can
// "edit" the architecture between rebuilds.
type switchSource struct {
	mu    sync.Mutex
	model *arch.SourceModel
}

func (s *switchSource) set(m *arch.SourceModel) { s.mu.Lock(); s.model = m; s.mu.Unlock() }

func (s *switchSource) Load(context.Context, string) (*arch.SourceModel, arch.Diagnostics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.model, nil, nil
}

func TestServe(t *testing.T) {
	t.Parallel()
	src := &switchSource{model: brokenModel()}
	html := &fakeBackend{format: viewmodel.FormatHTML, artifacts: []viewmodel.Artifact{{Path: "index.html"}}}
	svg := &fakeBackend{format: viewmodel.FormatSVG, artifacts: []viewmodel.Artifact{{Path: "diagrams/a.svg"}}}
	d2 := &fakeBackend{format: viewmodel.FormatD2}
	store := &fakeStore{}
	watcher := &fakeWatcher{ch: make(chan struct{})}
	preview := &fakePreview{events: make(chan string, 4)}
	prose := &fakeProse{}

	model := smallModel()
	model.Elements[0].Docs = "docs/s.md"

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, ServeDeps{
			Build:    BuildDeps{Source: src, Prose: prose, Backends: []Backend{d2, svg, html}, Store: store},
			Watcher:  watcher,
			Preview:  preview,
			Describe: func(d arch.Diagnostics) string { return "described " + d[0].Code },
		}, ServeRequest{Root: "proj", OutDir: "proj/dist"})
	}()
	next := func() string {
		t.Helper()
		select {
		case e := <-preview.events:
			return e
		case <-time.After(2 * time.Second):
			t.Fatal("no preview event")
			return ""
		}
	}

	// A broken project at start-up is not fatal: the server shows the error.
	if e := next(); e != "fail" || preview.failed[0] != "described unresolved_reference" {
		t.Fatalf("first event = %s %v", e, preview.failed)
	}
	src.set(model)
	watcher.ch <- struct{}{}
	if e := next(); e != "publish" {
		t.Fatalf("recovery event = %s, want publish (FR-031)", e)
	}
	if got := len(preview.published[0]); got != 2 {
		t.Errorf("published %d artifacts, want html + the svg it requires", got)
	}
	if d2.runs != 0 {
		t.Error("serve renders only html and what it requires")
	}
	if store.commits != 0 {
		t.Error("serve must never write the output directory")
	}
	abs, _ := filepath.Abs("proj/dist")
	if !reflect.DeepEqual(watcher.spec.Exclude, []string{abs}) || watcher.spec.Root != "proj" {
		t.Errorf("watch spec = %+v, want the absolute output directory excluded", watcher.spec)
	}
	if got := watcher.spec.ExtraFiles(); !reflect.DeepEqual(got, []string{"docs/s.md"}) {
		t.Errorf("ExtraFiles = %v, want the last good build's prose", got)
	}

	cancel()
	if err := <-done; err != nil {
		t.Errorf("Serve after cancel = %v, want nil", err)
	}
}

// TestServeUnreadableRootIsFatal: an operational failure of the first build,
// such as an unreadable project root, stops serve instead of being shown in
// the browser; only compile diagnostics are non-fatal (contracts/cli.md § serve).
func TestServeUnreadableRootIsFatal(t *testing.T) {
	t.Parallel()
	preview := &fakePreview{}
	err := Serve(context.Background(), ServeDeps{
		Build:    BuildDeps{Source: stubSource{err: errors.New("reading project root nope: no such file")}, Backends: []Backend{&fakeBackend{format: viewmodel.FormatHTML}, &fakeBackend{format: viewmodel.FormatSVG}}},
		Watcher:  &fakeWatcher{ch: make(chan struct{})},
		Preview:  preview,
		Describe: func(Diagnostics) string { return "" },
	}, ServeRequest{Root: "nope", OutDir: "nope/dist"})
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("Serve = %v, want the operational error", err)
	}
	if len(preview.failed) != 0 || len(preview.published) != 0 {
		t.Error("an operational startup failure must not reach the preview")
	}
}
