package devserver

import (
	"net/http"
	"sync"
)

// hub fans reload notifications out to every connected browser.
type hub struct {
	mu      sync.Mutex
	clients map[chan struct{}]bool
	closed  bool
}

func newHub() *hub { return &hub{clients: map[chan struct{}]bool{}} }

// broadcast queues a reload for every client. A client that already has one
// queued needs no second one, so the send never blocks.
func (h *hub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}

// close ends every stream, so a shutdown is not held open by browsers.
func (h *hub) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for c := range h.clients {
		close(c)
		delete(h.clients, c)
	}
}

// serve streams Server-Sent Events until the browser disconnects or the hub
// closes. Browsers reconnect by themselves if the stream drops.
func (h *hub) serve(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	c := make(chan struct{}, 1)
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.clients[c] = true
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(": connected\n\n"))
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case _, open := <-c:
			if !open {
				return
			}
			if _, err := w.Write([]byte("event: reload\ndata: {}\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
