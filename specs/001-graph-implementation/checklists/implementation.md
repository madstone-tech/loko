# Implementation Checklist: Graph Implementation Improvements

**Feature**: `001-graph-implementation`  
**Plan**: [plan.md](../plan.md)  
**Specification**: [spec.md](../spec.md)  
**Status**: Not Started

---

## Phase 1: P0 - Node ID Collision Fix

**Goal**: Eliminate silent data loss from duplicate component names  
**Estimated Effort**: 6-8 hours  
**Status**: ⬜ Not Started

### Tasks

- [ ] **1.1: Add Qualified ID Generation Functions**
  - [ ] Create `QualifiedNodeID()` helper in `internal/core/entities/graph.go`
  - [ ] Create `ParseQualifiedID()` helper
  - [ ] Write unit tests for ID generation (all node types)
  - [ ] Write unit tests for ID parsing (round-trip)
  - [ ] Write unit tests for collision prevention

- [ ] **1.2: Add ID Resolution Map**
  - [ ] Add `ShortIDMap` field to `ArchitectureGraph` struct
  - [ ] Update `NewArchitectureGraph()` to initialize map
  - [ ] Add `ResolveID()` method
  - [ ] Write tests for unambiguous resolution
  - [ ] Write tests for ambiguous cases
  - [ ] Write tests for resolution after node addition

- [ ] **1.3: Update Graph Builder**
  - [ ] Modify system node ID generation (line 53)
  - [ ] Modify container node ID generation (line 73)
  - [ ] Modify component node ID generation (line 96)
  - [ ] Add relationship resolution lookup map (lines 119-140)
  - [ ] Add warning log for unresolved relationships
  - [ ] Write tests for duplicate names across systems
  - [ ] Write tests for relationship resolution

- [ ] **1.4: Update AddNode Method**
  - [ ] Add ShortIDMap population logic (after line 93)
  - [ ] Write tests for map population
  - [ ] Write tests for short ID lookup
  - [ ] Verify duplicate detection still works

- [ ] **1.5: Integration Test**
  - [ ] Create `tests/integration/graph_collision_test.go`
  - [ ] Test multi-system with duplicate component names
  - [ ] Test relationship resolution
  - [ ] Test validation passes
  - [ ] Verify 100% component inclusion

### Validation Gates

- [ ] All unit tests pass: `go test ./internal/core/entities -v -run Qualified`
- [ ] Integration test passes: `go test ./tests/integration -v -run Collision`
- [ ] No silent failures in graph construction
- [ ] Coverage >80% for new code: `go test -cover ./internal/core/entities`

---

## Phase 2: P1 - Performance Optimizations

**Goal**: Reduce query time from O(E) to O(1)  
**Estimated Effort**: 6-8 hours  
**Status**: ⬜ Not Started

### Tasks

- [ ] **2.1: Add Reverse Adjacency Maps**
  - [ ] Add `IncomingEdges` field to struct
  - [ ] Add `ChildrenMap` field to struct
  - [ ] Update `NewArchitectureGraph()` to initialize both
  - [ ] Write tests for empty graph initialization

- [ ] **2.2: Update AddEdge for IncomingEdges**
  - [ ] Add IncomingEdges population (after line 124)
  - [ ] Add IncomingEdges for bidirectional edges (after line 137)
  - [ ] Write tests for single edge
  - [ ] Write tests for bidirectional edge
  - [ ] Write tests for multiple edges to same target

- [ ] **2.3: Optimize GetIncomingEdges**
  - [ ] Replace O(E) implementation with O(1) lookup
  - [ ] Write tests for correctness (same results as before)
  - [ ] Write benchmark test (target <1ms for 100 calls on 200-component graph)

- [ ] **2.4: Update AddNode for ChildrenMap**
  - [ ] Add ChildrenMap population (after line 97)
  - [ ] Write tests with parent
  - [ ] Write tests without parent (root)
  - [ ] Write tests for multiple children

- [ ] **2.5: Optimize GetChildren**
  - [ ] Replace O(N) implementation with O(1) lookup
  - [ ] Write tests for correctness
  - [ ] Write benchmark test (target <1ms for 5-level hierarchy)

- [ ] **2.6: Performance Integration Test**
  - [ ] Create `tests/integration/graph_performance_test.go`
  - [ ] Generate 200-component test graph (5 systems, 10 containers each, 4 components each)
  - [ ] Benchmark GetIncomingEdges (<1ms per call)
  - [ ] Benchmark GetChildren (<1ms per call)
  - [ ] Benchmark AnalyzeDependencies (<2s total)

### Validation Gates

- [ ] All tests pass: `go test ./internal/core/entities -v`
- [ ] Benchmarks meet targets: `go test -bench=. ./internal/core/entities | grep -E "(GetIncoming|GetChildren)"`
- [ ] GetIncomingEdges <1ms per call
- [ ] GetChildren <1ms per call
- [ ] AnalyzeDependencies <2s for 100-component graph

---

## Phase 3A: P2 - Type Safety

**Goal**: Eliminate `any` types and enable compile-time checking  
**Estimated Effort**: 5-6 hours  
**Status**: ⬜ Not Started

### Tasks

- [ ] **3.1: Create C4Entity Interface**
  - [ ] Create `internal/core/entities/c4_entity.go`
  - [ ] Define interface with GetID(), GetName(), GetEntityType()
  - [ ] Write tests for interface contract

- [ ] **3.2: Implement C4Entity**
  - [ ] Add methods to `System` (internal/core/entities/system.go)
  - [ ] Add methods to `Container` (internal/core/entities/container.go)
  - [ ] Add methods to `Component` (internal/core/entities/component.go)
  - [ ] Update existing entity tests to verify implementation

- [ ] **3.3: Update GraphNode.Data Type**
  - [ ] Change `Data any` to `Data C4Entity` in graph.go
  - [ ] Update all GraphNode creation sites in build_architecture_graph.go
  - [ ] Write test for type safety (compile-time check)
  - [ ] Verify GetEntityType() works without type assertion

- [ ] **3.4: Create DependencyReport Struct**
  - [ ] Create `internal/core/entities/dependency_report.go`
  - [ ] Define struct with all fields from map[string]any version
  - [ ] Write tests for JSON marshaling
  - [ ] Write tests for zero value

- [ ] **3.5: Update AnalyzeDependencies**
  - [ ] Change return type to `*entities.DependencyReport`
  - [ ] Replace map assignments with struct field assignments
  - [ ] Update all call sites (MCP tools)
  - [ ] Update tests to use struct fields

- [ ] **3.6: Create MCP Tool Argument Structs**
  - [ ] Add structs to `internal/mcp/tools/schemas.go`
  - [ ] Define QueryDependenciesArgs
  - [ ] Define AnalyzeCouplingArgs
  - [ ] Define all other graph tool arg structs
  - [ ] Write tests for JSON schema generation
  - [ ] Write tests for deserialization
  - [ ] Write tests for required field validation

- [ ] **3.7: Update MCP Tool Call() Methods**
  - [ ] Add `mapToStruct()` helper in tools/helpers.go
  - [ ] Update QueryDependenciesTool.Call()
  - [ ] Update AnalyzeCouplingTool.Call()
  - [ ] Update all other graph tools
  - [ ] Update tool tests to use typed args

- [ ] **3.8: Duplicate Edge Prevention**
  - [ ] Add duplicate check to AddEdge (before line 124)
  - [ ] Write test for adding identical edge twice
  - [ ] Write test for EdgeCount() accuracy
  - [ ] Write test for different types to same target

### Validation Gates

- [ ] Build succeeds: `go build ./...`
- [ ] No `interface{}` in graph consumer code: `grep -r "\.(\*.*)" internal/core/usecases/ | grep -v test | wc -l` → 0
- [ ] No runtime type assertions in core: Manual review
- [ ] All tests pass: `go test ./internal/core/... ./internal/mcp/tools/... -v`

---

## Phase 3B: P2 - Graph Caching

**Goal**: Eliminate repeated graph rebuilds in MCP sessions  
**Estimated Effort**: 3-4 hours  
**Status**: ⬜ Not Started

### Tasks

- [ ] **3.9: Create Graph Cache Structure**
  - [ ] Create `internal/mcp/server/graph_cache.go`
  - [ ] Define `GraphCache` struct with mutex
  - [ ] Define `CachedGraph` struct with timestamp
  - [ ] Implement Get(), Set(), Invalidate() methods
  - [ ] Write tests for cache hit/miss
  - [ ] Write tests for invalidation
  - [ ] Write race detector test: `go test -race`

- [ ] **3.10: Integrate Cache into MCP Server**
  - [ ] Add `graphCache` field to server/registry
  - [ ] Initialize in constructor
  - [ ] Pass cache to graph tool constructors
  - [ ] Write integration test for server init

- [ ] **3.11: Update Graph Tools to Use Cache**
  - [ ] Add `cache` field to QueryDependenciesTool
  - [ ] Add cache lookup before build in Call()
  - [ ] Add cache storage after build
  - [ ] Update all graph tools similarly
  - [ ] Write test for cache hit avoids rebuild
  - [ ] Write test for cache miss triggers build
  - [ ] Write test for multiple queries use cache

- [ ] **3.12: File Watcher Placeholder**
  - [ ] Create `internal/mcp/server/file_watcher.go` (stub)
  - [ ] Define FileWatcher struct
  - [ ] Define Watch() method with TODO comment
  - [ ] Document required behavior
  - [ ] Mark as future work in plan.md

- [ ] **3.13: Add RemoveNode and RemoveEdge**
  - [ ] Implement RemoveNode() in graph.go
  - [ ] Implement RemoveEdge() in graph.go
  - [ ] Add filterEdges() helper
  - [ ] Write test for RemoveNode cleans all edges
  - [ ] Write test for RemoveNode updates all maps
  - [ ] Write test for RemoveEdge specificity
  - [ ] Write test for removal errors

### Validation Gates

- [ ] Manual test: 50 consecutive MCP queries without file changes
- [ ] Verify query time <10ms after cache hit (first query may be slower)
- [ ] Verify cache invalidation rebuilds graph
- [ ] Race detector passes: `go test -race ./internal/mcp/server/...`

---

## Phase 4: P3 - Documentation & Quality

**Goal**: Document conventions and improve validation  
**Estimated Effort**: 4-6 hours  
**Status**: ⬜ Not Started

### Tasks

- [ ] **4.1: Filter Validation to Components**
  - [ ] Update checkIsolatedComponents in validate_architecture.go
  - [ ] Update checkHighCoupling similarly
  - [ ] Write test verifying systems excluded
  - [ ] Write test verifying containers excluded

- [ ] **4.2: Thread Safety Documentation**
  - [ ] Add package-level comment to graph.go
  - [ ] Document build-query-many pattern
  - [ ] Document synchronization requirements
  - [ ] Document MCP cache as concurrency strategy

- [ ] **4.3: Create ADR-003**
  - [ ] Create `docs/adr-003-graph-conventions.md`
  - [ ] Document node ID format decision
  - [ ] Document graph lifecycle
  - [ ] Document relationship scope
  - [ ] Document rationale and consequences
  - [ ] Review with team/AI for clarity

- [ ] **4.4: Update Graph Package Godoc**
  - [ ] Add comprehensive package comment to graph.go
  - [ ] Document qualified ID format
  - [ ] Document O(1) lookup guarantees
  - [ ] Reference ADR-003
  - [ ] Run `go doc internal/core/entities` to verify

- [ ] **4.5: Add Method Examples**
  - [ ] Add example comment to AddNode
  - [ ] Add example comment to AddEdge
  - [ ] Add example comment to GetDependencies
  - [ ] Verify examples are valid Go code

### Validation Gates

- [ ] ADR is understandable: Read aloud to unfamiliar person
- [ ] They can answer "why no system edges?" from ADR alone
- [ ] Godoc renders correctly: `godoc -http=:6060` and review
- [ ] Validation excludes systems/containers: Test output verification

---

## Final Validation

**Status**: ⬜ Not Started

- [ ] **Unit Tests**
  - [ ] All pass: `go test ./internal/core/... ./internal/mcp/... -v`
  - [ ] Coverage >80% for core/entities: `go test -cover ./internal/core/entities`
  - [ ] Coverage >80% for core/usecases: `go test -cover ./internal/core/usecases`

- [ ] **Integration Tests**
  - [ ] All pass: `go test ./tests/integration/... -v`
  - [ ] Collision test passes
  - [ ] Performance test passes

- [ ] **Benchmarks**
  - [ ] GetIncomingEdges <1ms: `go test -bench=GetIncomingEdges ./internal/core/entities/`
  - [ ] GetChildren <1ms: `go test -bench=GetChildren ./internal/core/entities/`
  - [ ] AnalyzeDependencies <2s: Manual test with 100-component graph

- [ ] **Build & Lint**
  - [ ] Build succeeds: `task build` or `go build ./...`
  - [ ] Lint passes: `task lint` or `golangci-lint run`
  - [ ] Format check: `task fmt`

- [ ] **Manual Testing**
  - [ ] MCP tools work via Claude Desktop / MCP inspector
  - [ ] Cache behavior verified (50 consecutive queries)
  - [ ] Qualified IDs appear correctly in MCP responses
  - [ ] No regression in existing functionality

- [ ] **Documentation**
  - [ ] CHANGELOG.md updated with breaking changes
  - [ ] Migration guide created: `docs/migration-001-graph-qualified-ids.md`
  - [ ] ADR-003 reviewed and approved
  - [ ] README.md updated if graph usage changed

- [ ] **Pull Request**
  - [ ] PR description includes breaking changes summary
  - [ ] PR references spec and plan
  - [ ] PR includes before/after benchmark results
  - [ ] Team notified of qualified ID migration
  - [ ] Reviewers assigned

---

## Success Metrics Tracking

| Metric | Before | After | Target | Status |
|--------|--------|-------|--------|--------|
| Component inclusion rate (multi-system) | ~50% | ___ | 100% | ⬜ |
| GetIncomingEdges time (200 components) | ~50-100ms | ___ | <1ms | ⬜ |
| GetChildren time (5 levels) | ~10-20ms | ___ | <1ms | ⬜ |
| AnalyzeDependencies time (100 components) | ~3-5s | ___ | <2s | ⬜ |
| MCP query time (50th consecutive call) | ~500ms | ___ | ~5ms | ⬜ |
| Runtime type assertions in graph code | 15+ | ___ | 0 | ⬜ |
| ADR clarity score (new contributor survey) | N/A | ___ | >80% | ⬜ |

**Instructions**: Fill in "After" column values during final validation. Status: ✅ = met target, ⚠️ = close but not met, ❌ = not met.

---

## Notes & Blockers

**Date**: 2026-02-12

### Blockers
- None currently

### Notes
- File watcher integration deferred to future PR (API in place)
- Short ID support maintained for backward compatibility
- Migration guide needed before removing short ID support

### Decisions
- (Add decisions made during implementation here)

---

**Last Updated**: 2026-02-12  
**Status Legend**: ⬜ Not Started | 🟦 In Progress | ✅ Complete | ❌ Blocked
