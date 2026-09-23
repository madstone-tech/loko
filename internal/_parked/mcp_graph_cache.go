package mcp

import (
	"sync"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// GraphCache provides thread-safe caching of architecture graphs per project.
// It eliminates the need to rebuild graphs on every MCP tool call during interactive sessions.
type GraphCache struct {
	mu      sync.RWMutex
	entries map[string]*CachedGraph
}

// CachedGraph wraps an architecture graph with metadata.
type CachedGraph struct {
	Graph   *entities.ArchitectureGraph
	BuiltAt time.Time
}

// NewGraphCache creates a new graph cache.
func NewGraphCache() *GraphCache {
	return &GraphCache{
		entries: make(map[string]*CachedGraph),
	}
}

// Get retrieves a cached graph for the given project root.
// Returns the graph and true if found, nil and false otherwise.
func (gc *GraphCache) Get(projectRoot string) (*entities.ArchitectureGraph, bool) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	if entry, ok := gc.entries[projectRoot]; ok {
		return entry.Graph, true
	}
	return nil, false
}

// Set stores a graph in the cache for the given project root.
func (gc *GraphCache) Set(projectRoot string, graph *entities.ArchitectureGraph) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.entries[projectRoot] = &CachedGraph{
		Graph:   graph,
		BuiltAt: time.Now(),
	}
}

// Invalidate removes the cached graph for the given project root.
// This should be called when source files change.
func (gc *GraphCache) Invalidate(projectRoot string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	delete(gc.entries, projectRoot)
}
