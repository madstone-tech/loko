package cmd

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestServe is US5 end to end on a real port and the real poller: an edit
// reloads the browser within the 2 s budget (SC-006), a compile error is
// shown with its position (FR-030), fixing it recovers (FR-031), and a
// non-source change does not rebuild (FR-032).
func TestServe(t *testing.T) {
	root := copyFixture(t, "two-systems")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan net.Addr, 1)
	done := make(chan int, 1)
	var stdout, stderr bytes.Buffer
	go func() {
		code, err := runServeWith(ctx, ServeOptions{Root: root, Port: 0, Stdout: &stdout, Stderr: &stderr,
			Ready: func(a net.Addr) { ready <- a }})
		if err != nil {
			t.Errorf("runServeWith: %v", err)
		}
		done <- code
	}()
	base := "http://" + (<-ready).String()

	reloads := subscribe(t, base)
	// The first build is not part of any budget; under -race with coverage on
	// a hosted runner it can take well over 10 s.
	startup := 10 * time.Second
	if raceEnabled {
		startup = 2 * time.Minute
	}
	waitFor(t, startup, func() bool { return strings.Contains(fetch(t, base+"/element/container/api.html"), "Storefront API") })
	// The initial build's reload may still be queued; only later ones count.
	for drained := false; !drained; {
		select {
		case <-reloads:
		case <-time.After(300 * time.Millisecond):
			drained = true
		}
	}

	src := filepath.Join(root, "main.loko.hcl")
	original := readString(t, src)
	write := func(s string) {
		if err := os.WriteFile(src, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// SC-006's 2 s budget is asserted in normal builds. Under -race, wall-clock
	// time is meaningless, so only the behaviour is checked.
	budget := 2 * time.Second
	if raceEnabled {
		budget = 30 * time.Second
	}
	start := time.Now()
	write(strings.Replace(original, `"Storefront API"`, `"Storefront API v2"`, 1))
	expectReloadShowing(t, reloads, base+"/element/container/api.html", "Storefront API v2", budget)
	t.Logf("edit reflected in %v", time.Since(start))

	write(original + "\nsystem \"broken\" {\n  uses \"x\" { target = container.nope }\n}\n")
	page := expectReloadShowing(t, reloads, base+"/index.html", "main.loko.hcl:", budget)
	if strings.Contains(page, "sidebar") {
		t.Errorf("the error page must not show the stale site:\n%.400s", page)
	}

	write(original)
	expectReloadShowing(t, reloads, base+"/index.html", "sidebar", budget)

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("notes"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reloads:
		t.Error("a non-source change triggered a rebuild")
	case <-time.After(time.Second):
	}

	cancel()
	if code := <-done; code != 0 {
		t.Errorf("exit = %d after cancellation", code)
	}
	if !strings.Contains(stdout.String(), "serving http://127.0.0.1:") {
		t.Errorf("no serving line:\n%s", stdout.String())
	}
}

func subscribe(t *testing.T, base string) <-chan struct{} {
	t.Helper()
	resp, err := http.Get(base + "/_loko/events")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	out := make(chan struct{}, 8)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			if sc.Text() == "event: reload" {
				out <- struct{}{}
			}
		}
	}()
	return out
}

// expectReloadShowing waits for reloads until the page at url contains want,
// failing if that does not happen within the budget. An earlier reload still
// in flight (from a slow initial build under -race) is skipped, not trusted.
func expectReloadShowing(t *testing.T, reloads <-chan struct{}, url, want string, within time.Duration) string {
	t.Helper()
	deadline := time.After(within)
	for {
		select {
		case <-reloads:
			if page := fetch(t, url); strings.Contains(page, want) {
				return page
			}
		case <-deadline:
			t.Fatalf("no reload showing %q within %v", want, within)
			return ""
		}
	}
}

func fetch(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func waitFor(t *testing.T, within time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(within)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %v", within)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// TestServeUnreadableRoot: a project root that cannot be read exits 1
// (contracts/cli.md § serve).
func TestServeUnreadableRoot(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	code, err := runServeWith(ctx, ServeOptions{Root: filepath.Join(t.TempDir(), "missing"), Port: 0, Stdout: &stdout, Stderr: &stderr})
	if code != 1 || err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("code=%d err=%v, want 1 and an error naming the root", code, err)
	}
}
