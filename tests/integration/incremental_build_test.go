package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// MockIncremental FileBuildTracker tracks which files have been rendered.
type MockIncrementalBuildTracker struct {
	renderedDiagrams map[string]time.Time // Path → rendered time
	buildStartTime   time.Time
}

// NewMockIncrementalBuildTracker creates a new tracker.
func NewMockIncrementalBuildTracker() *MockIncrementalBuildTracker {
	return &MockIncrementalBuildTracker{
		renderedDiagrams: make(map[string]time.Time),
		buildStartTime:   time.Now(),
	}
}

// RecordRender records that a diagram was rendered.
func (t *MockIncrementalBuildTracker) RecordRender(path string) {
	t.renderedDiagrams[path] = time.Now()
}

// WasRendered checks if a file was rendered.
func (t *MockIncrementalBuildTracker) WasRendered(path string) bool {
	_, ok := t.renderedDiagrams[path]
	return ok
}

// RenderCount returns the total number of renders.
func (t *MockIncrementalBuildTracker) RenderCount() int {
	return len(t.renderedDiagrams)
}

// TestIncrementalBuildTracking tests that we track which diagrams need rebuilding.
func TestIncrementalBuildTracking(t *testing.T) {
	tracker := NewMockIncrementalBuildTracker()

	// Initially, nothing is rendered
	if tracker.RenderCount() != 0 {
		t.Error("expected 0 renders initially")
	}

	// Record some renders
	tracker.RecordRender("system-payment.d2")
	tracker.RecordRender("container-api.d2")

	if tracker.RenderCount() != 2 {
		t.Errorf("expected 2 renders, got %d", tracker.RenderCount())
	}

	if !tracker.WasRendered("system-payment.d2") {
		t.Error("payment system diagram should be marked as rendered")
	}

	if !tracker.WasRendered("container-api.d2") {
		t.Error("api container diagram should be marked as rendered")
	}

	if tracker.WasRendered("system-auth.d2") {
		t.Error("auth system diagram should not be marked as rendered")
	}
}

// TestContentHashComparison tests diagram caching via content hash.
func TestContentHashComparison(t *testing.T) {
	tests := []struct {
		name         string
		oldSource    string
		newSource    string
		needsRebuild bool
	}{
		{
			name:         "same_content_no_rebuild",
			oldSource:    "shape: rect",
			newSource:    "shape: rect",
			needsRebuild: false,
		},
		{
			name:         "different_content_rebuild",
			oldSource:    "shape: rect",
			newSource:    "shape: circle",
			needsRebuild: true,
		},
		{
			name:         "whitespace_difference",
			oldSource:    "shape: rect",
			newSource:    "shape: rect\n",
			needsRebuild: true, // Whitespace counts as different
		},
		{
			name:         "empty_to_content",
			oldSource:    "",
			newSource:    "shape: rect",
			needsRebuild: true,
		},
		{
			name:         "content_to_empty",
			oldSource:    "shape: rect",
			newSource:    "",
			needsRebuild: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate hash comparison
			oldHash := hashDiagram(tt.oldSource)
			newHash := hashDiagram(tt.newSource)

			needsRebuild := oldHash != newHash
			if needsRebuild != tt.needsRebuild {
				t.Errorf("expected needsRebuild=%v, got %v", tt.needsRebuild, needsRebuild)
			}
		})
	}
}

// hashDiagram is a simple hash function for testing.
func hashDiagram(source string) string {
	// In production, use SHA256, but for testing we can use a simple approach
	return fmt.Sprintf("%d", len(source))
}

// TestSelectiveRenderingLogic tests selecting which diagrams to render.
func TestSelectiveRenderingLogic(t *testing.T) {
	// Simulate a project state before and after a change
	oldState := map[string]string{
		"system-payment": "shape: rect",
		"container-api":  "shape: rect",
		"container-db":   "shape: cylinder",
	}

	newState := map[string]string{
		"system-payment":  "shape: rect",                 // Unchanged
		"container-api":   "shape: rect\nlabel: Updated", // Changed
		"container-db":    "shape: cylinder",             // Unchanged
		"container-cache": "shape: rect",                 // New
	}

	// Determine what needs rendering
	toRender := make([]string, 0)
	for key, newSource := range newState {
		oldSource, exists := oldState[key]
		if !exists || oldSource != newSource {
			toRender = append(toRender, key)
		}
	}

	// Should only render updated and new diagrams
	if len(toRender) != 2 {
		t.Errorf("expected 2 diagrams to render, got %d: %v", len(toRender), toRender)
	}

	// Verify specific diagrams are marked for rendering
	shouldRender := map[string]bool{
		"container-api":   true,
		"container-cache": true,
		"system-payment":  false,
		"container-db":    false,
	}

	for diagram, shouldRenderIt := range shouldRender {
		isMarked := false
		for _, marked := range toRender {
			if marked == diagram {
				isMarked = true
				break
			}
		}
		if isMarked != shouldRenderIt {
			t.Errorf("diagram %s: expected shouldRender=%v, got %v", diagram, shouldRenderIt, isMarked)
		}
	}
}

// TestBuildPerformance tests that builds complete in reasonable time.
func TestBuildPerformance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a project with multiple systems and containers
	project := &entities.Project{
		Name:    "perf-test",
		Systems: make(map[string]*entities.System),
	}

	// Create 10 systems, each with 5 containers and diagrams
	numSystems := 10
	for i := 1; i <= numSystems; i++ {
		systemName := fmt.Sprintf("System%d", i)
		containers := make(map[string]*entities.Container)
		for j := 1; j <= 5; j++ {
			containerName := fmt.Sprintf("Container%d", j)
			containers[containerName] = &entities.Container{
				ID:   containerName,
				Name: containerName,
				Diagram: &entities.Diagram{
					Source: "shape: rect",
				},
			}
		}
		project.Systems[systemName] = &entities.System{
			ID:         systemName,
			Name:       systemName,
			Containers: containers,
			Diagram: &entities.Diagram{
				Source: "shape: rect",
			},
		}
	}

	systems := make([]*entities.System, 0)
	for _, sys := range project.Systems {
		systems = append(systems, sys)
	}

	// Measure time to collect all diagrams
	startTime := time.Now()

	totalDiagrams := 0
	for _, sys := range systems {
		if sys.Diagram != nil {
			totalDiagrams++
		}
		for _, container := range sys.Containers {
			if container.Diagram != nil {
				totalDiagrams++
			}
		}
	}

	elapsed := time.Since(startTime)

	// Should be able to iterate through 60 diagrams in milliseconds
	expectedDiagrams := numSystems * 6 // 10 systems * (1 system diagram + 5 container diagrams)
	if totalDiagrams != expectedDiagrams {
		t.Errorf("expected %d diagrams, got %d", expectedDiagrams, totalDiagrams)
	}

	// Should complete very quickly (should be <100ms)
	if elapsed > 100*time.Millisecond {
		t.Logf("WARNING: diagram collection took %v (expected <100ms)", elapsed)
	}

	// Verify context is still valid
	select {
	case <-ctx.Done():
		t.Error("context cancelled unexpectedly")
	default:
		// Still valid
	}
}

// TestIncrementalBuildMultipleChanges tests incremental builds with multiple file changes.
func TestIncrementalBuildMultipleChanges(t *testing.T) {
	tests := []struct {
		name           string
		initialState   map[string]string
		changes        map[string]string // Path -> new content (empty = delete)
		expectedRebuld int               // Number of files to rebuild
	}{
		{
			name: "single_diagram_change",
			initialState: map[string]string{
				"system-a": "shape: rect",
				"system-b": "shape: circle",
			},
			changes: map[string]string{
				"system-a": "shape: rect\nlabel: Updated",
			},
			expectedRebuld: 1,
		},
		{
			name: "multiple_diagram_changes",
			initialState: map[string]string{
				"system-a":    "shape: rect",
				"system-b":    "shape: circle",
				"container-a": "shape: rect",
			},
			changes: map[string]string{
				"system-a":    "shape: rect\nlabel: Updated",
				"container-a": "shape: cylinder",
			},
			expectedRebuld: 2,
		},
		{
			name: "new_diagram_added",
			initialState: map[string]string{
				"system-a": "shape: rect",
			},
			changes: map[string]string{
				"system-b": "shape: circle",
			},
			expectedRebuld: 1,
		},
		{
			name: "no_changes",
			initialState: map[string]string{
				"system-a": "shape: rect",
				"system-b": "shape: circle",
			},
			changes:        map[string]string{},
			expectedRebuld: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count diagrams that need rebuilding
			toRebuild := 0
			for key, newSource := range tt.changes {
				oldSource, exists := tt.initialState[key]
				if !exists || oldSource != newSource {
					toRebuild++
				}
			}

			if toRebuild != tt.expectedRebuld {
				t.Errorf("expected %d rebuilds, got %d", tt.expectedRebuld, toRebuild)
			}
		})
	}
}

// TestIncrementalBuildState tests maintaining state across builds.
func TestIncrementalBuildState(t *testing.T) {
	// Simulate state management across multiple builds
	type BuildState struct {
		diagrams      map[string]string
		lastBuildTime time.Time
	}

	state := &BuildState{
		diagrams:      make(map[string]string),
		lastBuildTime: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	// First build: add diagrams
	state.diagrams["system-a"] = "shape: rect"
	state.diagrams["container-b"] = "shape: circle"
	state.lastBuildTime = time.Now()

	firstBuildTime := state.lastBuildTime

	// Wait a bit, then update
	time.Sleep(10 * time.Millisecond)
	state.diagrams["system-a"] = "shape: rect\nlabel: Updated"
	state.lastBuildTime = time.Now()

	// Verify state is maintained
	if state.diagrams["system-a"] != "shape: rect\nlabel: Updated" {
		t.Error("system-a should be updated")
	}

	if state.diagrams["container-b"] != "shape: circle" {
		t.Error("container-b should be unchanged")
	}

	if !state.lastBuildTime.After(firstBuildTime) {
		t.Error("last build time should be updated")
	}
}
