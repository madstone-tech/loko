package devserver

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/synctest"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

var _ usecases.PreviewServer = (*Server)(nil)

const page = "<!DOCTYPE html>\n<!-- notice -->\n<html><body><p>hi</p></body></html>\n"

func artifacts() []vm.Artifact {
	return []vm.Artifact{
		{Path: "index.html", Bytes: []byte(page)},
		{Path: "assets/style.css", Bytes: []byte("body{}")},
		{Path: "diagrams/landscape.svg", Bytes: []byte("<svg/>")},
	}
}

func get(t *testing.T, ts *httptest.Server, path string) (int, string, string) {
	t.Helper()
	resp, err := ts.Client().Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(b)
}

// The HTTP tests run in a testing/synctest bubble on httptest.NewTestServer's
// in-memory network: no real sockets, and no real-clock timeouts.

func TestServesFromMemoryWithInjectedReload(t *testing.T) {
	synctest.Test(t, testServesFromMemory)
}

func testServesFromMemory(t *testing.T) {
	s := New()
	as := artifacts()
	s.Publish(as)
	ts := httptest.NewTestServer(t, s.Handler())

	code, ctype, body := get(t, ts, "/")
	if code != 200 || !strings.HasPrefix(ctype, "text/html") {
		t.Fatalf("GET / = %d %s", code, ctype)
	}
	if !strings.Contains(body, "/_loko/events") || !strings.Contains(body, "</script>\n</body>") {
		t.Errorf("reload script not injected before </body>:\n%s", body)
	}
	if string(as[0].Bytes) != page {
		t.Error("the artifact itself was modified; injection must happen per response")
	}
	if code, ctype, _ := get(t, ts, "/assets/style.css"); code != 200 || !strings.HasPrefix(ctype, "text/css") {
		t.Errorf("css = %d %s", code, ctype)
	}
	if code, _, _ := get(t, ts, "/nope.html"); code != 404 {
		t.Errorf("missing page = %d, want 404", code)
	}
}

func TestErrorStateAndRecovery(t *testing.T) {
	synctest.Test(t, testErrorState)
}

func testErrorState(t *testing.T) {
	s := New()
	s.Publish(artifacts())
	ts := httptest.NewTestServer(t, s.Handler())

	s.Fail("main.loko.hcl:3:5: Unresolvable reference <container.nope>")
	_, _, body := get(t, ts, "/")
	if !strings.Contains(body, "main.loko.hcl:3:5") || strings.Contains(body, "<p>hi</p>") {
		t.Errorf("error state must show diagnostics, never the stale page (FR-030):\n%s", body)
	}
	if !strings.Contains(body, "&lt;container.nope&gt;") {
		t.Error("diagnostics text must be escaped")
	}
	if code, _, _ := get(t, ts, "/diagrams/landscape.svg"); code != 200 {
		t.Error("assets are still served in the error state")
	}
	s.Publish(artifacts())
	if _, _, body := get(t, ts, "/"); !strings.Contains(body, "<p>hi</p>") {
		t.Error("a later build must clear the error state (FR-031)")
	}
}

func TestEventsBroadcastReload(t *testing.T) {
	synctest.Test(t, testEvents)
}

// testEvents reads the SSE stream on a fake network. synctest.Wait returns
// once the server has written and the reader has consumed everything, so
// each assertion sees exactly the events sent so far.
func testEvents(t *testing.T) {
	s := New()
	ts := httptest.NewTestServer(t, s.Handler())
	resp, err := ts.Client().Get(ts.URL + "/_loko/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content type = %s", resp.Header.Get("Content-Type"))
	}
	var mu sync.Mutex
	var lines []string
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			mu.Lock()
			lines = append(lines, sc.Text())
			mu.Unlock()
		}
	}()
	reloads := func() int {
		synctest.Wait()
		mu.Lock()
		defer mu.Unlock()
		n := 0
		for _, l := range lines {
			if l == "event: reload" {
				n++
			}
		}
		return n
	}
	if n := reloads(); n != 0 {
		t.Fatalf("%d reloads before anything was published", n)
	}
	s.Publish(artifacts())
	if n := reloads(); n != 1 {
		t.Errorf("after Publish: %d reloads, want 1", n)
	}
	s.Fail("boom")
	if n := reloads(); n != 2 {
		t.Errorf("after Fail: %d reloads, want 2", n)
	}
	// End the stream the way shutdown does, so the handler and the reader
	// exit and the bubble can finish.
	s.hub.close()
	synctest.Wait()
}

func TestListenBindsLoopbackOnly(t *testing.T) {
	t.Parallel()
	s := New()
	ctx, cancel := context.WithCancel(context.Background())
	addr := make(chan net.Addr, 1)
	errc := make(chan error, 1)
	go func() { errc <- s.ListenAndServe(ctx, 0, func(a net.Addr) { addr <- a }) }()
	a := <-addr
	if tcp, ok := a.(*net.TCPAddr); !ok || !tcp.IP.IsLoopback() {
		t.Errorf("bound %v, want a loopback address", a)
	}
	cancel()
	if err := <-errc; err != nil {
		t.Errorf("ListenAndServe after cancel = %v, want nil", err)
	}
}
