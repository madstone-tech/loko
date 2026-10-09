package usecases

import (
	"sync"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// revisionsKept bounds the memory: a token older than this many reads is
// treated as stale, and the caller reads again.
const revisionsKept = 64

// RevisionMemory maps the short revision tokens handed to callers back to the
// per-file hashes they stand for. It lives only as long as the process; after
// a restart every old token is unknown, which a write treats as stale.
type RevisionMemory struct {
	mu    sync.Mutex
	order []string
	revs  map[string]authoring.Revision
}

// remember records rev. A nil memory remembers nothing.
func (m *RevisionMemory) remember(rev authoring.Revision) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	tok := rev.Token()
	if _, ok := m.revs[tok]; ok {
		return
	}
	if m.revs == nil {
		m.revs = map[string]authoring.Revision{}
	}
	m.revs[tok] = rev
	m.order = append(m.order, tok)
	if len(m.order) > revisionsKept {
		delete(m.revs, m.order[0])
		m.order = m.order[1:]
	}
}

// lookup returns the revision behind a token, or false if it was never issued
// or has been forgotten.
func (m *RevisionMemory) lookup(token string) (authoring.Revision, bool) {
	if m == nil {
		return authoring.Revision{}, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	rev, ok := m.revs[token]
	return rev, ok
}
