# Implementation Plan: Graph Implementation Improvements

**Feature Branch**: `001-graph-implementation`  
**Created**: 2026-02-12  
**Specification**: [spec.md](./spec.md)

## Overview

This plan addresses critical bugs and performance issues in the ArchitectureGraph implementation. The work is organized into phases that follow the priority order from the specification (P0 → P1 → P2 → P3), with each phase building upon the previous one.

**Estimated Total Effort**: 3-4 days  
**Testing Approach**: TDD with integration tests for each phase  
**Risk Level**: Medium (core infrastructure changes with backward compatibility requirements)

---

## Phase 1: P0 - Fix Node ID Collision Bug (Day 1)

**Goal**: Eliminate silent data loss from duplicate component names across systems  
**Success Criteria**: SC-001 (100% component inclusion rate)  
**Estimated Effort**: 6-8 hours

### Tasks

#### 1.1: Add Qualified ID Generation Functions
**File**: `internal/core/entities/graph.go`  
**Lines**: Add new functions after line 351

- Create `QualifiedNodeID(nodeType, systemID, containerID, nodeID string) string` helper
  - System: returns `systemID`
  - Container: returns `systemID/containerID`
  - Component: returns `systemID/containerID/componentID`
- Create `ParseQualifiedID(qualifiedID string) (parts []string, nodeType string)` helper
  - Splits qualified ID and determines node type from structure

**Tests**: `internal/core/entities/graph_test.go`
- Test qualified ID format for each node type
- Test round-trip parsing (generate → parse → validate)
- Test collision prevention (different hierarchy paths with same short names)

#### 1.2: Add ID Resolution Map to ArchitectureGraph
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify struct at lines 10-20

- Add `ShortIDMap map[string]string` field (short ID → qualified ID)
- Update `NewArchitectureGraph()` to initialize the new map
- Add `ResolveID(shortID string) (qualifiedID string, ok bool)` method
  - Returns qualified ID if unique short ID exists
  - Returns empty string if ambiguous or not found

**Tests**:
- Test single short ID resolution (unambiguous case)
- Test multiple components with same short ID (ambiguous case)
- Test resolution after nodes are added to graph

#### 1.3: Update Graph Builder to Use Qualified IDs
**File**: `internal/core/usecases/build_architecture_graph.go`  
**Lines**: Modify Execute() at lines 33-148

- Line 53: Change `systemNode.ID = entities.QualifiedNodeID("system", system.ID, "", "")`
- Line 73: Change `containerNode.ID = entities.QualifiedNodeID("container", system.ID, container.ID, "")`
- Line 96: Change `componentNode.ID = entities.QualifiedNodeID("component", system.ID, container.ID, component.ID)`
- Lines 119-140: Update relationship edge creation
  - Build lookup map of short component IDs to qualified IDs before loop
  - Resolve `relatedID` using the lookup map
  - Add warning log for unresolved relationships

**Tests**: `internal/core/usecases/build_architecture_graph_test.go`
- Test graph building with duplicate component names across systems
- Test graph building with duplicate container names across systems
- Test relationship resolution using short IDs
- Test error reporting for ambiguous short IDs

#### 1.4: Update AddNode to Populate ShortIDMap
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify AddNode() at lines 84-101

- After line 93 (node added to Nodes map), populate ShortIDMap:
  ```go
  // Extract short ID from qualified ID
  parts, _ := ParseQualifiedID(node.ID)
  shortID := parts[len(parts)-1]
  ag.ShortIDMap[shortID] = node.ID
  ```

**Tests**:
- Test ShortIDMap population during AddNode
- Test short ID lookup after graph is built
- Test error when duplicate node IDs are added (existing behavior preserved)

#### 1.5: Integration Test for Multi-System Collision
**File**: `tests/integration/graph_collision_test.go` (new file)

- Create test project with:
  - System "backend" with Container "api" with Component "auth"
  - System "admin" with Container "ui" with Component "auth"
  - Relationship from backend/api/auth to admin/ui/auth
- Build graph and verify:
  - Both components exist as distinct nodes
  - Relationship edge exists between qualified IDs
  - GetNode() works with both qualified and short IDs (via ResolveID)
  - Validation passes without errors

---

## Phase 2: P1 - Performance Optimizations (Day 1-2)

**Goal**: Reduce dependency query time from O(E) to O(1)  
**Success Criteria**: SC-002 (queries under 50ms), SC-003 (validation under 2s)  
**Estimated Effort**: 6-8 hours

### Tasks

#### 2.1: Add Reverse Adjacency Maps to ArchitectureGraph
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify struct at lines 10-20

- Add `IncomingEdges map[string][]*GraphEdge` field
- Add `ChildrenMap map[string][]string` field (parent ID → child IDs)
- Update `NewArchitectureGraph()` to initialize both maps

**Tests**: `internal/core/entities/graph_test.go`
- Test empty graph has initialized maps
- Test maps survive graph validation

#### 2.2: Update AddEdge to Maintain IncomingEdges
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify AddEdge() at lines 109-141

- After line 124 (forward edge added):
  ```go
  ag.IncomingEdges[edge.Target] = append(ag.IncomingEdges[edge.Target], edge)
  ```
- After line 137 (bidirectional reverse edge added):
  ```go
  ag.IncomingEdges[edge.Source] = append(ag.IncomingEdges[edge.Source], reverseEdge)
  ```

**Tests**:
- Test IncomingEdges populated for single edge
- Test IncomingEdges populated for bidirectional edge
- Test multiple edges to same target accumulate correctly

#### 2.3: Update GetIncomingEdges to Use Index
**File**: `internal/core/entities/graph.go`  
**Lines**: Replace GetIncomingEdges() at lines 148-159

- Change implementation to:
  ```go
  func (ag *ArchitectureGraph) GetIncomingEdges(nodeID string) []*GraphEdge {
      return ag.IncomingEdges[nodeID]
  }
  ```

**Tests**:
- Test GetIncomingEdges returns correct edges (same results as before)
- **Benchmark test**: Measure performance on graph with 200 components, 500 edges
  - Target: < 50ms for 100 consecutive GetIncomingEdges calls
  - Compare before/after optimization

#### 2.4: Update AddNode to Maintain ChildrenMap
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify AddNode() at lines 84-101

- After line 97 (ParentMap populated):
  ```go
  ag.ChildrenMap[node.ParentID] = append(ag.ChildrenMap[node.ParentID], node.ID)
  ```

**Tests**:
- Test ChildrenMap populated when node has parent
- Test ChildrenMap not populated when node has no parent (root)
- Test multiple children of same parent accumulate correctly

#### 2.5: Update GetChildren to Use Index
**File**: `internal/core/entities/graph.go`  
**Lines**: Replace GetChildren() at lines 170-180

- Change implementation to:
  ```go
  func (ag *ArchitectureGraph) GetChildren(nodeID string) []*GraphNode {
      var children []*GraphNode
      for _, childID := range ag.ChildrenMap[nodeID] {
          if node := ag.Nodes[childID]; node != nil {
              children = append(children, node)
          }
      }
      return children
  }
  ```

**Tests**:
- Test GetChildren returns correct nodes (same results as before)
- **Benchmark test**: Measure performance on graph with 5-level hierarchy
  - Target: < 100ms for GetDescendants on root node
  - Compare before/after optimization

#### 2.6: Performance Integration Test
**File**: `tests/integration/graph_performance_test.go` (new file)

- Create large test project:
  - 5 systems, 10 containers per system, 4 components per container (200 components total)
  - Average 3 relationships per component (600 edges)
- Benchmark operations:
  - GetIncomingEdges: < 1ms per call
  - GetChildren: < 1ms per call
  - AnalyzeDependencies: < 2s total
- Use `testing.B` for benchmarks, `-benchmem` for memory profiling

---

## Phase 3: P2 - Type Safety & Caching (Day 2-3)

**Goal**: Eliminate `any` types and add graph caching  
**Success Criteria**: SC-004 (constant MCP query time), SC-005 (compile-time type checking), SC-006 (duplicate prevention)  
**Estimated Effort**: 8-10 hours

### Part A: Type Safety

#### 3.1: Create C4Entity Interface
**File**: `internal/core/entities/c4_entity.go` (new file)

```go
package entities

// C4Entity is the common interface for all C4 model entities.
type C4Entity interface {
    GetID() string
    GetName() string
    GetEntityType() string // "system", "container", "component"
}
```

**Tests**: `internal/core/entities/c4_entity_test.go`
- Verify System implements C4Entity
- Verify Container implements C4Entity
- Verify Component implements C4Entity

#### 3.2: Implement C4Entity on Entities
**Files**: 
- `internal/core/entities/system.go`
- `internal/core/entities/container.go`
- `internal/core/entities/component.go`

Add methods to each:
```go
func (s *System) GetID() string { return s.ID }
func (s *System) GetName() string { return s.Name }
func (s *System) GetEntityType() string { return "system" }
```

**Tests**: Update existing entity tests to verify interface implementation

#### 3.3: Update GraphNode.Data Type
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify GraphNode struct at lines 22-48

- Change line 44: `Data C4Entity` (was `Data any`)
- Update all GraphNode creation sites in build_architecture_graph.go

**Tests**:
- Test GraphNode creation with System, Container, Component
- Test compile error when attempting to assign non-C4Entity to Data
- Verify GetEntityType() works without type assertion

#### 3.4: Create DependencyReport Struct
**File**: `internal/core/entities/dependency_report.go` (new file)

```go
package entities

type DependencyReport struct {
    SystemsCount      int            `json:"systems_count"`
    ContainersCount   int            `json:"containers_count"`
    ComponentsCount   int            `json:"components_count"`
    TotalNodes        int            `json:"total_nodes"`
    TotalEdges        int            `json:"total_edges"`
    IsolatedComponents []string      `json:"isolated_components"`
    HighlyCoupled     map[string]int `json:"highly_coupled_components"`
    CentralComponents map[string]int `json:"central_components"`
}
```

**Tests**: `internal/core/entities/dependency_report_test.go`
- Test JSON marshaling/unmarshaling
- Test zero value initialization

#### 3.5: Update AnalyzeDependencies Return Type
**File**: `internal/core/usecases/build_architecture_graph.go`  
**Lines**: Modify AnalyzeDependencies() at lines 195-246

- Change return type from `map[string]any` to `*entities.DependencyReport`
- Replace `report["key"] = value` with struct field assignments
- Update all call sites (MCP tools)

**Tests**: Update existing tests to use struct fields instead of map keys

#### 3.6: Create MCP Tool Argument Structs
**File**: `internal/mcp/tools/schemas.go` (existing file, add structs)

Define structs for each MCP tool:
```go
type QueryDependenciesArgs struct {
    ProjectRoot  string `json:"project_root" jsonschema:"required,description=Project root directory"`
    ComponentID  string `json:"component_id" jsonschema:"description=Component to analyze"`
}

type AnalyzeCouplingArgs struct {
    ProjectRoot string `json:"project_root" jsonschema:"required"`
    Threshold   int    `json:"threshold,omitempty" jsonschema:"description=Coupling threshold"`
}
// ... etc for all graph tools
```

**Tests**: `internal/mcp/tools/schemas_test.go`
- Test JSON schema generation from struct tags
- Test deserialization from MCP tool args
- Test validation of required fields

#### 3.7: Update MCP Tool Call() Methods
**File**: `internal/mcp/tools/graph_tools.go`  
**Lines**: Update each tool's Call() method

Replace manual type assertions with:
```go
var args QueryDependenciesArgs
if err := mapToStruct(rawArgs, &args); err != nil {
    return nil, fmt.Errorf("invalid arguments: %w", err)
}
```

Add `mapToStruct` helper in tools/helpers.go using json.Marshal/Unmarshal

**Tests**: Update existing MCP tool tests to use typed args

#### 3.8: Duplicate Edge Prevention
**File**: `internal/core/entities/graph.go`  
**Lines**: Modify AddEdge() at lines 109-141

Add after line 121 (before appending edge):
```go
// Check for duplicate edge
for _, existing := range ag.Edges[edge.Source] {
    if existing.Target == edge.Target && existing.Type == edge.Type {
        return nil // Already exists, not an error
    }
}
```

**Tests**:
- Test adding identical edge twice results in single edge
- Test EdgeCount() returns correct count after duplicate attempts
- Test different edge types to same target are both added

### Part B: Graph Caching

#### 3.9: Create Graph Cache Structure
**File**: `internal/mcp/server/graph_cache.go` (new file)

```go
package server

import (
    "sync"
    "time"
    "github.com/madstone-tech/loko/internal/core/entities"
)

type GraphCache struct {
    mu      sync.RWMutex
    entries map[string]*CachedGraph
}

type CachedGraph struct {
    Graph   *entities.ArchitectureGraph
    BuiltAt time.Time
}

func NewGraphCache() *GraphCache {
    return &GraphCache{
        entries: make(map[string]*CachedGraph),
    }
}

func (gc *GraphCache) Get(projectRoot string) (*entities.ArchitectureGraph, bool) {
    gc.mu.RLock()
    defer gc.mu.RUnlock()
    
    if entry, ok := gc.entries[projectRoot]; ok {
        return entry.Graph, true
    }
    return nil, false
}

func (gc *GraphCache) Set(projectRoot string, graph *entities.ArchitectureGraph) {
    gc.mu.Lock()
    defer gc.mu.Unlock()
    
    gc.entries[projectRoot] = &CachedGraph{
        Graph:   graph,
        BuiltAt: time.Now(),
    }
}

func (gc *GraphCache) Invalidate(projectRoot string) {
    gc.mu.Lock()
    defer gc.mu.Unlock()
    
    delete(gc.entries, projectRoot)
}
```

**Tests**: `internal/mcp/server/graph_cache_test.go`
- Test cache hit/miss
- Test cache invalidation
- Test concurrent access (race detector)

#### 3.10: Integrate Cache into MCP Server
**File**: `internal/mcp/server/server.go` or tools registry

- Add `graphCache *GraphCache` field to server/registry struct
- Initialize in constructor
- Pass cache to graph tool constructors

**Tests**: Integration test for MCP server initialization

#### 3.11: Update Graph Tools to Use Cache
**File**: `internal/mcp/tools/graph_tools.go`  
**Lines**: Update each graph tool

Modify tool structs to include cache:
```go
type QueryDependenciesTool struct {
    repo         ports.ProjectRepository
    graphBuilder *usecases.BuildArchitectureGraph
    cache        *server.GraphCache // new field
}
```

Update Call() methods:
```go
func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]any) (any, error) {
    // Try cache first
    if graph, ok := t.cache.Get(projectRoot); ok {
        // Use cached graph
        return analyzeGraph(graph), nil
    }
    
    // Cache miss - build graph
    project, err := t.repo.LoadProject(ctx, projectRoot)
    // ... build graph ...
    
    // Store in cache
    t.cache.Set(projectRoot, graph)
    
    return analyzeGraph(graph), nil
}
```

**Tests**:
- Test cache hit avoids rebuild
- Test cache miss triggers build and caches result
- Test multiple queries use cached graph

#### 3.12: Add File Watcher Integration (Placeholder)
**File**: `internal/mcp/server/file_watcher.go` (new file, stub implementation)

```go
package server

// FileWatcher monitors src/ directory for changes
type FileWatcher struct {
    cache *GraphCache
}

// Watch starts monitoring projectRoot/src for changes
// On change detected: cache.Invalidate(projectRoot)
func (fw *FileWatcher) Watch(projectRoot string) error {
    // TODO: Implement using fsnotify or similar
    // For now, just document the contract
    return nil
}
```

**Tests**: Document required behavior, mark as TODO

**Note**: Full file watcher implementation is out of scope for initial PR. The cache invalidation API is in place for future integration.

#### 3.13: Add RemoveNode and RemoveEdge Methods
**File**: `internal/core/entities/graph.go`  
**Lines**: Add after EdgeCount() at line 315

```go
// RemoveNode removes a node and all edges referencing it.
func (ag *ArchitectureGraph) RemoveNode(nodeID string) error {
    if ag.Nodes[nodeID] == nil {
        return fmt.Errorf("node %q not found", nodeID)
    }
    
    // Remove from Nodes
    delete(ag.Nodes, nodeID)
    
    // Remove from ParentMap if child
    delete(ag.ParentMap, nodeID)
    
    // Remove from ChildrenMap if parent
    delete(ag.ChildrenMap, nodeID)
    
    // Remove from ShortIDMap
    parts, _ := ParseQualifiedID(nodeID)
    shortID := parts[len(parts)-1]
    delete(ag.ShortIDMap, shortID)
    
    // Remove all edges involving this node
    // Outgoing edges
    delete(ag.Edges, nodeID)
    delete(ag.IncomingEdges, nodeID)
    
    // Incoming edges (where this node is target)
    for source := range ag.Edges {
        ag.Edges[source] = filterEdges(ag.Edges[source], func(e *GraphEdge) bool {
            return e.Target != nodeID
        })
    }
    
    // Outgoing edges from other nodes (where this node is source)
    for target := range ag.IncomingEdges {
        ag.IncomingEdges[target] = filterEdges(ag.IncomingEdges[target], func(e *GraphEdge) bool {
            return e.Source != nodeID
        })
    }
    
    return nil
}

// RemoveEdge removes an edge from source to target.
func (ag *ArchitectureGraph) RemoveEdge(source, target, edgeType string) error {
    if ag.Nodes[source] == nil || ag.Nodes[target] == nil {
        return fmt.Errorf("source or target node not found")
    }
    
    // Remove from Edges
    ag.Edges[source] = filterEdges(ag.Edges[source], func(e *GraphEdge) bool {
        return !(e.Target == target && e.Type == edgeType)
    })
    
    // Remove from IncomingEdges
    ag.IncomingEdges[target] = filterEdges(ag.IncomingEdges[target], func(e *GraphEdge) bool {
        return !(e.Source == source && e.Type == edgeType)
    })
    
    return nil
}

// filterEdges is a helper to filter edge slices
func filterEdges(edges []*GraphEdge, keep func(*GraphEdge) bool) []*GraphEdge {
    result := make([]*GraphEdge, 0, len(edges))
    for _, edge := range edges {
        if keep(edge) {
            result = append(result, edge)
        }
    }
    return result
}
```

**Tests**: `internal/core/entities/graph_test.go`
- Test RemoveNode removes all edges
- Test RemoveNode updates all maps
- Test RemoveEdge removes specific edge
- Test removal of non-existent node/edge returns error

---

## Phase 4: P3 - Documentation & Quality (Day 3-4)

**Goal**: Document conventions and improve validation  
**Success Criteria**: SC-007 (ADR enables understanding), SC-008 (cache invalidation responsiveness)  
**Estimated Effort**: 4-6 hours

### Tasks

#### 4.1: Filter Validation to Components Only
**File**: `internal/core/usecases/validate_architecture.go`  
**Lines**: Find checkIsolatedComponents and checkHighCoupling functions

Update both functions to filter by node type:
```go
for nodeID, node := range graph.Nodes {
    if node.Type != "component" {
        continue
    }
    // ... existing logic ...
}
```

**Tests**: Update validation tests to verify systems/containers are excluded

#### 4.2: Add Thread Safety Documentation
**File**: `internal/core/entities/graph.go`  
**Lines**: Add package-level comment before ArchitectureGraph struct

```go
// Thread Safety:
//
// ArchitectureGraph is NOT safe for concurrent writes. The graph must be fully
// built before concurrent reads. Typical usage pattern:
//
//   1. Build graph sequentially (AddNode, AddEdge)
//   2. Validate graph (graph.Validate())
//   3. Query graph concurrently (GetNode, GetDependencies, etc.)
//
// If concurrent writes are needed, external synchronization is required.
// For read-heavy workloads (MCP queries), caching is preferred over locking.
```

#### 4.3: Create ADR for Graph Conventions
**File**: `docs/adr-003-graph-conventions.md` (new file)

```markdown
# ADR-003: Architecture Graph Conventions

**Status**: Accepted  
**Date**: 2026-02-12  
**Context**: Graph implementation improvements (001-graph-implementation)

## Decision

### Node ID Format

Node IDs use qualified hierarchical paths:
- **System**: `system-id` (unchanged from short ID)
- **Container**: `system-id/container-id`
- **Component**: `system-id/container-id/component-id`

**Rationale**: Prevents ID collisions when multiple systems/containers have
components with the same name (e.g., "auth" component in multiple systems).

**Migration**: Existing code using short IDs is supported via `ResolveID()` 
lookup map. Graph builder maintains backward compatibility for relationship
references using short component IDs.

### Graph Lifecycle

1. **Build**: Graph is constructed sequentially from project structure
2. **Cache**: Built graph is cached at MCP server level, keyed by project root
3. **Query**: Cached graph serves read-only queries (GetDependencies, etc.)
4. **Invalidate**: Cache invalidated when src/ files change (future: file watcher)
5. **Rebuild**: Next query triggers rebuild and re-caching

**Rationale**: Build-once-query-many pattern optimizes for LLM interaction
sessions (10-50 queries per session). Cache eliminates repeated disk I/O and
graph construction overhead.

### Relationship Scope

Only **components** have dependency edges in the graph. Systems and containers
have hierarchical parent-child relationships, but not dependency edges.

**Rationale**: Matches C4 model semantics - components are the units of code
that communicate. System-level "dependencies" and "external systems" are metadata
for documentation, not runtime dependencies.

**Implication**: Validation checks (isolated components, coupling analysis) 
filter to component nodes only. Systems and containers never appear "isolated"
because they don't participate in dependency edges.

## Consequences

- **Node ID change is breaking**: Existing saved graphs/diagrams using short IDs
  will need migration or regeneration
- **Cache requires invalidation strategy**: Without file watching, cache becomes
  stale on file changes. Manual invalidation or TTL fallback needed short-term.
- **Component-only validation**: Clearer separation between hierarchy (all nodes)
  and dependencies (components only)
```

**Tests**: (Documentation - no automated tests)

#### 4.4: Update Graph Package Godoc
**File**: `internal/core/entities/graph.go`  
**Lines**: Update package comment at top of file

Add comprehensive package documentation:
```go
// Package entities provides core domain models for loko's C4 architecture.
//
// Architecture Graph
//
// The ArchitectureGraph converts the hierarchical C4 model (System → Container
// → Component) into a directed graph for dependency analysis, coupling metrics,
// and architectural querying.
//
// Node IDs use qualified paths to prevent collisions (see ADR-003):
//   - System: "backend"
//   - Container: "backend/api"
//   - Component: "backend/api/auth"
//
// Graph operations are O(1) for lookups using pre-computed indexes:
//   - GetIncomingEdges: O(1) via IncomingEdges map
//   - GetChildren: O(1) via ChildrenMap
//   - GetDependencies: O(1) via Edges map
//
// See docs/adr-003-graph-conventions.md for design rationale.
```

#### 4.5: Add Examples to Key Methods
**File**: `internal/core/entities/graph.go`  
**Lines**: Add example comments to AddNode, AddEdge, GetDependencies

```go
// AddNode adds a node to the graph.
//
// Example:
//   graph := NewArchitectureGraph()
//   node := &GraphNode{
//       ID:   "backend/api/auth",
//       Type: "component",
//       Name: "Authentication Service",
//       // ...
//   }
//   err := graph.AddNode(node)
```

**Tests**: (Documentation - examples can be tested with go test -run Example)

---

## Testing Strategy

### Unit Tests
- **Coverage Target**: >80% for internal/core/entities/graph.go
- **Coverage Target**: >80% for internal/core/usecases/build_architecture_graph.go
- **Tools**: `go test -cover ./internal/core/...`
- **Run After**: Each task completion

### Integration Tests
- **Location**: `tests/integration/`
- **Scenarios**:
  - Multi-system collision (Phase 1)
  - Large graph performance (Phase 2)
  - MCP cache behavior (Phase 3)
- **Run After**: Each phase completion

### Benchmark Tests
- **Location**: `*_test.go` files with `func BenchmarkXxx(b *testing.B)`
- **Metrics**:
  - GetIncomingEdges: < 1ms per call on 200-node graph
  - GetChildren: < 1ms per call on 5-level hierarchy
  - AnalyzeDependencies: < 2s on 100-component graph
- **Run**: `go test -bench=. -benchmem ./internal/core/entities/`

### Manual Testing
- **MCP Tools**: Test via MCP inspector or Claude Desktop
- **Cache Validation**: Verify query time consistency across 50 calls
- **File Changes**: Manually invalidate cache and verify rebuild

---

## Rollout Plan

### Step 1: Merge Phases Incrementally
- **PR 1**: Phase 1 (P0 collision fix) - Critical, merge ASAP
- **PR 2**: Phase 2 (P1 performance) - High priority, merge within week
- **PR 3**: Phase 3 (P2 type safety + caching) - Can be split into 3A and 3B
- **PR 4**: Phase 4 (P3 documentation) - Can merge independently

### Step 2: Backward Compatibility
- **Short ID Support**: Keep ResolveID() method indefinitely for MCP tool flexibility
- **Deprecation Path**: 
  - v0.2: Qualified IDs default, short IDs supported
  - v0.3: Warn on short ID usage
  - v0.4: Remove short ID support (breaking change)

### Step 3: Migration Guide
Create `docs/migration-001-graph-qualified-ids.md`:
- How to update existing MCP tool calls
- How to regenerate saved graphs/diagrams
- How to update custom scripts

---

## Risk Mitigation

### Risk 1: Qualified IDs Break Existing Diagrams
**Likelihood**: High  
**Impact**: Medium (diagrams need regeneration)  
**Mitigation**: 
- Add migration script to update diagram references
- Document regeneration process in migration guide
- Keep short ID resolution for transition period

### Risk 2: Cache Invalidation Without File Watcher
**Likelihood**: High  
**Impact**: Low (stale cache shows outdated data)  
**Mitigation**:
- Add manual invalidation API in Phase 3
- Document cache behavior in MCP tool descriptions
- Add TTL fallback (cache expires after 5 minutes)

### Risk 3: Performance Regression from Index Maintenance
**Likelihood**: Low  
**Impact**: Low (slightly slower AddNode/AddEdge)  
**Mitigation**:
- Benchmark before/after Phase 2
- Build-time performance is less critical than query-time
- Index updates are O(1) append operations

### Risk 4: Type Safety Changes Require Broad Refactoring
**Likelihood**: Medium  
**Impact**: Medium (many files touched)  
**Mitigation**:
- Make changes compile-time breaking (not runtime)
- Refactor incrementally (C4Entity first, then DependencyReport, then MCP args)
- Comprehensive test coverage catches breaks early

---

## Dependencies

### External Dependencies
- No new external libraries required
- Uses existing Go stdlib (encoding/json, sync, time)

### Internal Dependencies
- `internal/core/entities`: Core domain models (System, Container, Component)
- `internal/core/usecases`: Graph builder and analysis use cases
- `internal/core/usecases/ports`: Repository interfaces
- `internal/mcp/tools`: MCP tool implementations
- `internal/mcp/server`: MCP server (for cache integration)

### Constitution Compliance
- ✅ No arbitrary code execution
- ✅ Clean Architecture (entities → usecases → adapters)
- ✅ No external dependencies in core/
- ✅ Immutable builds (deterministic graph construction)
- ✅ Testable interfaces (repository ports, graph operations)

---

## Success Metrics

Track these metrics before/after implementation:

| Metric | Before | Target | How to Measure |
|--------|--------|--------|----------------|
| Component inclusion rate (multi-system) | ~50% (collision bug) | 100% | Integration test |
| GetIncomingEdges time (200 components) | ~50-100ms (O(E)) | <1ms | Benchmark |
| GetChildren time (5 levels) | ~10-20ms (O(N)) | <1ms | Benchmark |
| AnalyzeDependencies time (100 components) | ~3-5s | <2s | Benchmark |
| MCP query time (50th consecutive call) | ~500ms (rebuild) | ~5ms (cache hit) | Manual test |
| Runtime type assertions in graph code | 15+ | 0 | `grep -r "\.(\*.*)" internal/core/usecases/` |
| ADR clarity score | N/A | >80% | Survey new contributors |

---

## Checklist

Use this checklist to track progress during implementation:

### Phase 1: P0 - Node ID Collision
- [ ] Task 1.1: Qualified ID generation functions
- [ ] Task 1.2: ID resolution map
- [ ] Task 1.3: Update graph builder
- [ ] Task 1.4: Update AddNode
- [ ] Task 1.5: Integration test
- [ ] **Validation**: Run `go test ./internal/core/... -v -run Collision`
- [ ] **Validation**: Verify zero silent failures in test output

### Phase 2: P1 - Performance
- [ ] Task 2.1: Add reverse adjacency maps
- [ ] Task 2.2: Update AddEdge
- [ ] Task 2.3: Update GetIncomingEdges
- [ ] Task 2.4: Update AddNode (ChildrenMap)
- [ ] Task 2.5: Update GetChildren
- [ ] Task 2.6: Performance integration test
- [ ] **Validation**: Run `go test -bench=. ./internal/core/entities/ | grep -E "(GetIncoming|GetChildren)"`
- [ ] **Validation**: Verify <1ms benchmarks pass

### Phase 3A: P2 - Type Safety
- [ ] Task 3.1: Create C4Entity interface
- [ ] Task 3.2: Implement on entities
- [ ] Task 3.3: Update GraphNode.Data
- [ ] Task 3.4: Create DependencyReport struct
- [ ] Task 3.5: Update AnalyzeDependencies
- [ ] Task 3.6: MCP tool argument structs
- [ ] Task 3.7: Update MCP Call() methods
- [ ] Task 3.8: Duplicate edge prevention
- [ ] **Validation**: Run `go build ./...` (should compile without type errors)
- [ ] **Validation**: Verify no `interface{}` in graph consumer code

### Phase 3B: P2 - Caching
- [ ] Task 3.9: Create graph cache structure
- [ ] Task 3.10: Integrate into MCP server
- [ ] Task 3.11: Update graph tools
- [ ] Task 3.12: File watcher placeholder
- [ ] Task 3.13: RemoveNode/RemoveEdge
- [ ] **Validation**: Manual test 50 consecutive MCP queries
- [ ] **Validation**: Verify consistent <10ms response time after cache hit

### Phase 4: P3 - Documentation
- [ ] Task 4.1: Filter validation to components
- [ ] Task 4.2: Thread safety docs
- [ ] Task 4.3: ADR-003 creation
- [ ] Task 4.4: Graph package godoc
- [ ] Task 4.5: Method examples
- [ ] **Validation**: Read ADR aloud to someone unfamiliar with codebase
- [ ] **Validation**: Verify they can answer "why no system edges?" without code

### Final Validation
- [ ] All unit tests pass: `go test ./internal/core/... ./internal/mcp/... -v`
- [ ] All integration tests pass: `go test ./tests/integration/... -v`
- [ ] Coverage meets targets: `task coverage` (>80% core/)
- [ ] Benchmarks meet targets: Check success metrics table
- [ ] Lint passes: `task lint`
- [ ] Build succeeds: `task build`
- [ ] Manual MCP testing via Claude Desktop / MCP inspector
- [ ] Update CHANGELOG.md with breaking changes (qualified IDs)
- [ ] Create migration guide in docs/
- [ ] Notify team of breaking changes in PR description

---

## Notes

- **File Watcher**: Marked as TODO in Phase 3. API is in place for future integration with `fsnotify` or similar. Short-term fallback: manual cache invalidation or TTL.

- **Qualified ID Migration**: Short ID support is maintained for backward compatibility. Removal requires coordinated migration across all MCP tools and diagram generators.

- **Performance Targets**: Based on human perception thresholds (<100ms feels instant) and typical LLM session patterns (10-50 queries). Adjust if usage patterns differ.

- **Type Safety Trade-offs**: Full type safety at MCP boundaries (JSON) would require code generation. Current approach balances safety (compile-time checks in core) with pragmatism (runtime checks at adapter boundaries).

- **Constitution Alignment**: All changes preserve Clean Architecture, avoid external dependencies in core/, and maintain testability. Graph caching is opt-in at adapter level (MCP server), not in core use cases.
