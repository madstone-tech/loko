package devserver

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func get(t *testing.T, base, path string) (int, string, string) {
	t.Helper()
	resp, err := http.Get(base + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(b)
}

func TestServesFromMemoryWithInjectedReload(t *testing.T) {
	t.Parallel()
	s := New()
	as := artifacts()
	s.Publish(as)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	code, ctype, body := get(t, ts.URL, "/")
	if code != 200 || !strings.HasPrefix(ctype, "text/html") {
		t.Fatalf("GET / = %d %s", code, ctype)
	}
	if !strings.Contains(body, "/_loko/events") || !strings.Contains(body, "</script>\n</body>") {
		t.Errorf("reload script not injected before </body>:\n%s", body)
	}
	if string(as[0].Bytes) != page {
		t.Error("the artifact itself was modified; injection must happen per response")
	}
	if code, ctype, _ := get(t, ts.URL, "/assets/style.css"); code != 200 || !strings.HasPrefix(ctype, "text/css") {
		t.Errorf("css = %d %s", code, ctype)
	}
	if code, _, _ := get(t, ts.URL, "/nope.html"); code != 404 {
		t.Errorf("missing page = %d, want 404", code)
	}
}

func TestErrorStateAndRecovery(t *testing.T) {
	t.Parallel()
	s := New()
	s.Publish(artifacts())
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	s.Fail("main.loko.hcl:3:5: Unresolvable reference <container.nope>")
	_, _, body := get(t, ts.URL, "/")
	if !strings.Contains(body, "main.loko.hcl:3:5") || strings.Contains(body, "<p>hi</p>") {
		t.Errorf("error state must show diagnostics, never the stale page (FR-030):\n%s", body)
	}
	if !strings.Contains(body, "&lt;container.nope&gt;") {
		t.Error("diagnostics text must be escaped")
	}
	if code, _, _ := get(t, ts.URL, "/diagrams/landscape.svg"); code != 200 {
		t.Error("assets are still served in the error state")
	}
	s.Publish(artifacts())
	if _, _, body := get(t, ts.URL, "/"); !strings.Contains(body, "<p>hi</p>") {
		t.Error("a later build must clear the error state (FR-031)")
	}
}

func TestEventsBroadcastReload(t *testing.T) {
	t.Parallel()
	s := New()
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/_loko/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content type = %s", resp.Header.Get("Content-Type"))
	}
	lines := make(chan string, 16)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	waitFor := func(want string) {
		t.Helper()
		deadline := time.After(2 * time.Second)
		for {
			select {
			case l, ok := <-lines:
				if !ok {
					t.Fatalf("stream closed before %q", want)
				}
				if l == want {
					return
				}
			case <-deadline:
				t.Fatalf("no %q within 2s", want)
			}
		}
	}
	waitFor(": connected")
	s.Publish(artifacts())
	waitFor("event: reload")
	s.Fail("boom")
	waitFor("event: reload")
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
