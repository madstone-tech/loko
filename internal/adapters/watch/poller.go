package watch

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// Interval is the production poll interval. With one settling tick, a save is
// noticed within about 2×Interval, well inside the 2 s budget (SC-006).
const Interval = 200 * time.Millisecond

// Poller implements usecases.ChangeWatcher by polling (research R12).
type Poller struct {
	// ticks returns the tick channel and its stop function; tests replace it.
	ticks func() (<-chan time.Time, func())
	// tickDone, when set, runs after each tick is fully processed, so a test
	// can step the poller one tick at a time.
	tickDone func()
}

// New returns a Poller ticking at interval.
func New(interval time.Duration) *Poller {
	return &Poller{ticks: func() (<-chan time.Time, func()) {
		t := time.NewTicker(interval)
		return t.C, t.Stop
	}}
}

type stamp struct {
	size int64
	mod  int64
}

// Watch snapshots the watched set on every tick and sends one signal once a
// change has been followed by a tick with no further change, so a burst of
// saves yields one rebuild (FR-033). Only architecture source, the prose from
// ExtraFiles and theme overrides are watched; anything else, and anything
// under an excluded directory, can never trigger a rebuild (FR-032).
//
// The goroutine exits, and the channel closes, when ctx is cancelled.
func (p *Poller) Watch(ctx context.Context, spec usecases.WatchSpec) (<-chan struct{}, error) {
	ticks, stop := p.ticks()
	// One slot: a signal waiting to be consumed already means "rebuild", so a
	// further change is absorbed by it rather than blocking the poll loop.
	out := make(chan struct{}, 1)
	prev := snapshot(spec)
	go func() {
		defer close(out)
		defer stop()
		pending := false
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticks:
			}
			prev, pending = step(spec, prev, pending, out)
			if p.tickDone != nil {
				p.tickDone()
			}
		}
	}()
	return out, nil
}

// step handles one tick: a change arms the signal, and the first quiet tick
// after a change fires it.
func step(spec usecases.WatchSpec, prev map[string]stamp, pending bool, out chan<- struct{}) (map[string]stamp, bool) {
	cur := snapshot(spec)
	if !equal(prev, cur) {
		return cur, true
	}
	if pending {
		select {
		case out <- struct{}{}:
		default:
		}
	}
	return prev, false
}

func snapshot(spec usecases.WatchSpec) map[string]stamp {
	snap := map[string]stamp{}
	add := func(path string) {
		if excluded(path, spec.Exclude) {
			return
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			snap[path] = stamp{info.Size(), info.ModTime().UnixNano()}
		}
	}
	if files, _, err := hclsource.Discover(spec.Root); err == nil {
		for _, f := range files {
			add(f.Abs)
		}
	}
	if spec.ExtraFiles != nil {
		for _, rel := range spec.ExtraFiles() {
			add(filepath.Join(spec.Root, filepath.FromSlash(rel)))
		}
	}
	_ = filepath.WalkDir(filepath.Join(spec.Root, "templates"), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			add(path)
		}
		return nil
	})
	return snap
}

func excluded(path string, dirs []string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, d := range dirs {
		if abs == d || strings.HasPrefix(abs, d+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func equal(a, b map[string]stamp) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
