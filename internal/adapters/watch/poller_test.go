package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.ChangeWatcher = (*Poller)(nil)

// harness drives a Poller with a hand-fed tick channel, so no test sleeps
// through real poll intervals.
type harness struct {
	t     *testing.T
	root  string
	ticks chan time.Time
	done  chan struct{}
	out   <-chan struct{}
	stamp int64
}

func start(t *testing.T, extra []string, exclude ...string) *harness {
	t.Helper()
	h := &harness{t: t, root: t.TempDir(), ticks: make(chan time.Time), done: make(chan struct{}), stamp: 1}
	h.write("main.loko.hcl", "project \"x\" {}")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	p := &Poller{
		ticks:    func() (<-chan time.Time, func()) { return h.ticks, func() {} },
		tickDone: func() { h.done <- struct{}{} },
	}
	abs := make([]string, len(exclude))
	for i, e := range exclude {
		abs[i] = filepath.Join(h.root, e)
	}
	out, err := p.Watch(ctx, usecases.WatchSpec{Root: h.root, ExtraFiles: func() []string { return extra }, Exclude: abs})
	if err != nil {
		t.Fatal(err)
	}
	h.out = out
	return h
}

// write creates or changes a file and gives it a distinct modification time,
// so a change is visible regardless of file-system timestamp granularity.
func (h *harness) write(rel, body string) {
	h.t.Helper()
	p := filepath.Join(h.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		h.t.Fatal(err)
	}
	h.stamp++
	ts := time.Unix(1_700_000_000+h.stamp, 0)
	if err := os.Chtimes(p, ts, ts); err != nil {
		h.t.Fatal(err)
	}
}

// tick delivers one tick and waits until the poller has fully processed it,
// so a file written next can never land in this tick's snapshot.
func (h *harness) tick() {
	h.ticks <- time.Now()
	<-h.done
}

// signals ticks n times and counts the signals seen.
func (h *harness) signals(n int) int {
	h.t.Helper()
	got := 0
	for range n {
		h.tick()
		select {
		case <-h.out:
			got++
		case <-time.After(20 * time.Millisecond):
		}
	}
	// A signal may still be sitting in the one-slot buffer.
	select {
	case <-h.out:
		got++
	case <-time.After(50 * time.Millisecond):
	}
	return got
}

func TestPollSettlesThenSignalsOnce(t *testing.T) {
	t.Parallel()
	h := start(t, nil)
	if n := h.signals(2); n != 0 {
		t.Fatalf("no change: got %d signals", n)
	}
	h.write("main.loko.hcl", "project \"x\" { description = \"changed\" }")
	if n := h.signals(3); n != 1 {
		t.Errorf("one change: got %d signals, want 1 after it settles", n)
	}
}

func TestPollDebouncesBursts(t *testing.T) {
	t.Parallel()
	h := start(t, nil)
	h.tick()
	for i, f := range []string{"a.loko.hcl", "b.loko.hcl", "c.loko.hcl"} {
		h.write(f, "system \"s"+string(rune('a'+i))+"\" {}")
		h.tick() // a change on every tick: nothing has settled yet
	}
	if n := h.signals(3); n != 1 {
		t.Errorf("three changes on consecutive ticks: got %d signals, want 1 (FR-033)", n)
	}
}

func TestPollDetectsDeletion(t *testing.T) {
	t.Parallel()
	h := start(t, nil)
	h.write("extra.loko.hcl", "")
	h.signals(2)
	if err := os.Remove(filepath.Join(h.root, "extra.loko.hcl")); err != nil {
		t.Fatal(err)
	}
	if n := h.signals(3); n != 1 {
		t.Errorf("deletion: got %d signals, want 1", n)
	}
}

func TestPollIgnoresNonSourceAndOutput(t *testing.T) {
	t.Parallel()
	h := start(t, nil, "public")
	h.write("README.md", "hello")
	h.write("dist/index.html", "<p>generated</p>")
	h.write("public/x.loko.hcl", "") // inside the excluded output directory
	h.write("notes/todo.txt", "x")
	if n := h.signals(3); n != 0 {
		t.Errorf("non-source changes caused %d signals, want 0 (FR-032)", n)
	}
}

func TestPollWatchesProseAndTheme(t *testing.T) {
	t.Parallel()
	h := start(t, []string{"./docs/a.md"})
	h.write("docs/a.md", "# a")
	if n := h.signals(3); n != 1 {
		t.Errorf("prose change: got %d signals, want 1", n)
	}
	h.write("docs/unreferenced.md", "# b")
	if n := h.signals(3); n != 0 {
		t.Errorf("unreferenced markdown: got %d signals, want 0", n)
	}
	h.write("templates/style.css", "body{}")
	if n := h.signals(3); n != 1 {
		t.Errorf("theme change: got %d signals, want 1", n)
	}
}

func TestPollClosesOnCancel(t *testing.T) {
	t.Parallel()
	ticks := make(chan time.Time)
	p := &Poller{ticks: func() (<-chan time.Time, func()) { return ticks, func() {} }}
	ctx, cancel := context.WithCancel(context.Background())
	out, err := p.Watch(ctx, usecases.WatchSpec{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("unexpected signal")
		}
	case <-time.After(time.Second):
		t.Fatal("the channel did not close after cancellation")
	}
}
