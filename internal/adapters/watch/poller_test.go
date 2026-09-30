package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.ChangeWatcher = (*Poller)(nil)

const tick = 200 * time.Millisecond

// harness runs the real Poller inside a synctest bubble: its ticker uses the
// bubble's fake clock, and advance moves it forward exactly one tick, then
// returns once the poller has finished that tick and blocked again. Nothing
// sleeps in real time, and a write can never race a snapshot.
type harness struct {
	t     *testing.T
	root  string
	out   <-chan struct{}
	stamp int64
}

func start(t *testing.T, extra []string, exclude ...string) *harness {
	t.Helper()
	h := &harness{t: t, root: t.TempDir(), stamp: 1}
	h.write("main.loko.hcl", `project "x" {}`)
	abs := make([]string, len(exclude))
	for i, e := range exclude {
		abs[i] = filepath.Join(h.root, e)
	}
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	out, err := New(tick).Watch(ctx, usecases.WatchSpec{Root: h.root, ExtraFiles: func() []string { return extra }, Exclude: abs})
	if err != nil {
		t.Fatal(err)
	}
	h.out = out
	return h
}

// write creates or changes a file with a distinct modification time, so the
// change is visible whatever the file system's timestamp granularity.
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

func (h *harness) advance() { synctest.Sleep(tick) }

// signals advances n ticks and counts the signals delivered.
func (h *harness) signals(n int) int {
	got := 0
	for range n {
		h.advance()
		select {
		case <-h.out:
			got++
		default:
		}
	}
	return got
}

func TestPollSettlesThenSignalsOnce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, nil)
		if n := h.signals(2); n != 0 {
			t.Fatalf("no change: got %d signals", n)
		}
		h.write("main.loko.hcl", `project "x" { description = "changed" }`)
		if n := h.signals(3); n != 1 {
			t.Errorf("one change: got %d signals, want 1 after it settles", n)
		}
	})
}

func TestPollDebouncesBursts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, nil)
		h.advance()
		for i, f := range []string{"a.loko.hcl", "b.loko.hcl", "c.loko.hcl"} {
			h.write(f, "system \"s"+string(rune('a'+i))+"\" {}")
			h.advance() // a change on every tick: nothing has settled yet
		}
		if n := h.signals(3); n != 1 {
			t.Errorf("three changes on consecutive ticks: got %d signals, want 1 (FR-033)", n)
		}
	})
}

func TestPollDetectsDeletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, nil)
		h.write("extra.loko.hcl", "")
		h.signals(2)
		if err := os.Remove(filepath.Join(h.root, "extra.loko.hcl")); err != nil {
			t.Fatal(err)
		}
		if n := h.signals(3); n != 1 {
			t.Errorf("deletion: got %d signals, want 1", n)
		}
	})
}

func TestPollIgnoresNonSourceAndOutput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, nil, "public")
		h.write("README.md", "hello")
		h.write("dist/index.html", "<p>generated</p>")
		h.write("public/x.loko.hcl", "") // inside the excluded output directory
		h.write("notes/todo.txt", "x")
		if n := h.signals(3); n != 0 {
			t.Errorf("non-source changes caused %d signals, want 0 (FR-032)", n)
		}
	})
}

func TestPollWatchesProseAndTheme(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
	})
}

// TestPollSignalWithinBudget pins the latency the serve budget relies on: a
// save is signalled within two ticks (research R12), measured on the
// bubble's exact clock rather than asserted against wall time.
func TestPollSignalWithinBudget(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, nil)
		h.write("main.loko.hcl", `project "x" { description = "v2" }`)
		saved := time.Now()
		<-h.out
		if got := time.Since(saved); got != 2*tick {
			t.Errorf("signalled after %v, want exactly 2 ticks (%v)", got, 2*tick)
		}
	})
}

func TestPollClosesOnCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		out, err := New(tick).Watch(ctx, usecases.WatchSpec{Root: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		cancel()
		if _, ok := <-out; ok {
			t.Error("unexpected signal")
		}
	})
}
