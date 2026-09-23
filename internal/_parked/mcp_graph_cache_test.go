package mcp

import (
	"sync"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// TestGraphCacheHitMiss tests cache hit and miss scenarios.
func TestGraphCacheHitMiss(t *testing.T) {
	cache := NewGraphCache()

	projectRoot := "/test/project"

	// Test cache miss
	if graph, ok := cache.Get(projectRoot); ok {
		t.Error("expected cache miss, got hit")
		if graph != nil {
			t.Error("expected nil graph on miss")
		}
	}

	// Add graph to cache
	testGraph := entities.NewArchitectureGraph()
	cache.Set(projectRoot, testGraph)

	// Test cache hit
	if graph, ok := cache.Get(projectRoot); !ok {
		t.Error("expected cache hit, got miss")
	} else if graph != testGraph {
		t.Error("cached graph doesn't match original")
	}
}

// TestGraphCacheInvalidation tests cache invalidation.
func TestGraphCacheInvalidation(t *testing.T) {
	cache := NewGraphCache()

	projectRoot := "/test/project"
	testGraph := entities.NewArchitectureGraph()

	// Set cache
	cache.Set(projectRoot, testGraph)

	// Verify it's cached
	if _, ok := cache.Get(projectRoot); !ok {
		t.Fatal("graph should be cached before invalidation")
	}

	// Invalidate cache
	cache.Invalidate(projectRoot)

	// Verify it's gone
	if _, ok := cache.Get(projectRoot); ok {
		t.Error("graph should not be cached after invalidation")
	}

	// Invalidating non-existent entry should not error
	cache.Invalidate("/non/existent")
}

// TestGraphCacheConcurrentAccess tests concurrent access with race detector.
func TestGraphCacheConcurrentAccess(t *testing.T) {
	cache := NewGraphCache()

	projectRoot := "/test/project"

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			graph := entities.NewArchitectureGraph()
			cache.Set(projectRoot, graph)
		}()
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get(projectRoot)
		}()
	}

	// Concurrent invalidations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Invalidate(projectRoot)
		}()
	}

	wg.Wait()

	// No race conditions should occur (verified by -race flag)
}

// TestGraphCacheMultipleProjects tests caching multiple projects.
func TestGraphCacheMultipleProjects(t *testing.T) {
	cache := NewGraphCache()

	project1 := "/test/project1"
	project2 := "/test/project2"

	graph1 := entities.NewArchitectureGraph()
	graph2 := entities.NewArchitectureGraph()

	cache.Set(project1, graph1)
	cache.Set(project2, graph2)

	// Verify both are cached independently
	if g, ok := cache.Get(project1); !ok || g != graph1 {
		t.Error("project1 graph not correctly cached")
	}

	if g, ok := cache.Get(project2); !ok || g != graph2 {
		t.Error("project2 graph not correctly cached")
	}

	// Invalidating one shouldn't affect the other
	cache.Invalidate(project1)

	if _, ok := cache.Get(project1); ok {
		t.Error("project1 should be invalidated")
	}

	if _, ok := cache.Get(project2); !ok {
		t.Error("project2 should still be cached")
	}
}
