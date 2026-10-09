package devserver

import (
	"context"
	"errors"
	"fmt"
	"html"
	"mime"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// Server implements usecases.PreviewServer.
//
// It serves the last published build from memory — serve never writes the
// output directory — and is in one of two states: ok, serving artifacts; or
// failed, in which every page shows the diagnostics instead of stale output
// (FR-030). Every transition tells connected browsers to reload.
type Server struct {
	mu      sync.RWMutex
	files   map[string][]byte
	failure string
	hub     *hub
}

// New returns a Server with nothing published yet.
func New() *Server { return &Server{files: map[string][]byte{}, hub: newHub()} }

// Publish implements usecases.PreviewServer.
func (s *Server) Publish(artifacts []vm.Artifact) {
	files := make(map[string][]byte, len(artifacts))
	for _, a := range artifacts {
		files[a.Path] = a.Bytes
	}
	s.mu.Lock()
	s.files, s.failure = files, ""
	s.mu.Unlock()
	s.hub.broadcast()
}

// Fail implements usecases.PreviewServer.
func (s *Server) Fail(text string) {
	s.mu.Lock()
	s.failure = text
	s.mu.Unlock()
	s.hub.broadcast()
}

// Handler serves the site, and Server-Sent Events on /_loko/events.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/_loko/events", s.hub.serve)
	mux.HandleFunc("/", s.serveFile)
	return mux
}

// ListenAndServe binds host:port and serves until ctx is cancelled, then shuts
// down. An empty host means 127.0.0.1: loopback is the default, and anything
// wider is the caller's explicit choice (a container, for one). port 0 picks a
// free port; bound, when given, receives the address actually bound.
func (s *Server) ListenAndServe(ctx context.Context, host string, port int, bound func(net.Addr)) error {
	if host == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("binding %s: %w", addr, err)
	}
	if bound != nil {
		bound(ln.Addr())
	}
	srv := &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		s.hub.close()
		shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if p == "" || strings.HasSuffix(r.URL.Path, "/") {
		p = path.Join(p, "index.html")
	}
	s.mu.RLock()
	body, ok := s.files[p]
	failure := s.failure
	s.mu.RUnlock()

	isPage := strings.HasSuffix(p, ".html")
	if isPage && failure != "" {
		writeHTML(w, http.StatusOK, inject([]byte(errorPage(failure))))
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	if isPage {
		writeHTML(w, http.StatusOK, inject(body))
		return
	}
	if ct := mime.TypeByExtension(path.Ext(p)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(body)
}

func writeHTML(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

func errorPage(text string) string {
	return "<!DOCTYPE html>\n<html lang=\"en\"><head><meta charset=\"UTF-8\"><title>Build failed · loko</title>" +
		"<style>body{font-family:system-ui,sans-serif;margin:2rem;color:#1f2937}" +
		"pre{background:#fef2f2;border-left:4px solid #ef4444;padding:1rem;overflow:auto;white-space:pre-wrap}</style>" +
		"</head><body><h1>The architecture does not compile</h1>" +
		"<p>Fix the problems below and save; this page updates by itself. Nothing shown here is stale output.</p>" +
		"<pre>" + html.EscapeString(text) + "</pre></body></html>\n"
}
