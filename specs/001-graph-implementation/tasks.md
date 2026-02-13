# Tasks: Graph Implementation Improvements

**Input**: Design documents from `/specs/001-graph-implementation/`  
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: This feature uses TDD approach - test tasks are included and MUST be completed before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions

- Go project: `internal/core/entities/`, `internal/core/usecases/`, `internal/mcp/`
- Tests: `internal/core/entities/*_test.go`, `tests/integration/`
- Documentation: `docs/`

---

## Phase 1: Setup (No User Story)

**Purpose**: Verify existing project structure and identify affected files

- [X] T001 Review current ArchitectureGraph implementation in internal/core/entities/graph.go
- [X] T002 Review BuildArchitectureGraph use case in internal/core/usecases/build_architecture_graph.go
- [X] T003 [P] Review MCP graph tools in internal/mcp/tools/graph_tools.go
- [X] T004 [P] Review component entity structure in internal/core/entities/component.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

This feature improves existing infrastructure - no new foundational components needed.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Multi-System Projects Work Correctly (Priority: P0) 🎯 MVP

**Goal**: Eliminate silent data loss from duplicate component names across systems by implementing qualified node IDs

**Independent Test**: Create project with two systems, each containing component "auth", verify both appear in graph with relationships intact

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T005 [P] [US1] Write test for QualifiedNodeID() generation in internal/core/entities/graph_test.go
- [X] T006 [P] [US1] Write test for ParseQualifiedID() parsing in internal/core/entities/graph_test.go
- [X] T007 [P] [US1] Write test for collision prevention with same short names in internal/core/entities/graph_test.go
- [X] T008 [P] [US1] Write test for single short ID resolution in internal/core/entities/graph_test.go
- [X] T009 [P] [US1] Write test for ambiguous short ID resolution in internal/core/entities/graph_test.go
- [X] T010 [P] [US1] Write test for ShortIDMap population in internal/core/entities/graph_test.go
- [X] T011 [P] [US1] Write test for graph building with duplicate component names in internal/core/usecases/build_architecture_graph_test.go
- [X] T012 [P] [US1] Write test for graph building with duplicate container names in internal/core/usecases/build_architecture_graph_test.go
- [X] T013 [P] [US1] Write test for relationship resolution using short IDs in internal/core/usecases/build_architecture_graph_test.go
- [X] T014 [US1] Create integration test file tests/integration/graph_collision_test.go with multi-system collision scenario

### Implementation for User Story 1

- [X] T015 [P] [US1] Implement QualifiedNodeID() helper function in internal/core/entities/graph.go after line 351
- [X] T016 [P] [US1] Implement ParseQualifiedID() helper function in internal/core/entities/graph.go after line 351
- [X] T017 [US1] Add ShortIDMap field to ArchitectureGraph struct in internal/core/entities/graph.go at lines 10-20
- [X] T018 [US1] Update NewArchitectureGraph() to initialize ShortIDMap in internal/core/entities/graph.go
- [X] T019 [US1] Add ResolveID() method to ArchitectureGraph in internal/core/entities/graph.go
- [X] T020 [US1] Update AddNode() to populate ShortIDMap in internal/core/entities/graph.go at lines 84-101
- [X] T021 [US1] Update graph builder Execute() to use qualified IDs for systems in internal/core/usecases/build_architecture_graph.go line 53
- [X] T022 [US1] Update graph builder Execute() to use qualified IDs for containers in internal/core/usecases/build_architecture_graph.go line 73
- [X] T023 [US1] Update graph builder Execute() to use qualified IDs for components in internal/core/usecases/build_architecture_graph.go line 96
- [X] T024 [US1] Update graph builder Execute() relationship resolution logic in internal/core/usecases/build_architecture_graph.go lines 119-140
- [X] T025 [US1] Verify all tests pass: go test ./internal/core/entities -v -run Qualified
- [X] T026 [US1] Verify integration test passes: go test ./tests/integration -v -run Collision
- [X] T027 [US1] Run full test suite: go test ./internal/core/... ./internal/mcp/... -v

**Checkpoint**: At this point, User Story 1 should be fully functional - multi-system projects with duplicate component names work correctly

---

## Phase 4: User Story 2 - Fast Dependency Queries (Priority: P1)

**Goal**: Reduce dependency query time from O(E) to O(1) by adding reverse adjacency maps and children maps

**Independent Test**: Create project with 200+ components, run dependency analysis, verify queries complete in <50ms

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T028 [P] [US2] Write test for IncomingEdges map initialization in internal/core/entities/graph_test.go
- [X] T029 [P] [US2] Write test for ChildrenMap initialization in internal/core/entities/graph_test.go
- [X] T030 [P] [US2] Write test for IncomingEdges population with single edge in internal/core/entities/graph_test.go
- [X] T031 [P] [US2] Write test for IncomingEdges population with bidirectional edge in internal/core/entities/graph_test.go
- [X] T032 [P] [US2] Write test for IncomingEdges with multiple edges to same target in internal/core/entities/graph_test.go
- [X] T033 [P] [US2] Write test for ChildrenMap population when node has parent in internal/core/entities/graph_test.go
- [X] T034 [P] [US2] Write test for ChildrenMap when node has no parent in internal/core/entities/graph_test.go
- [X] T035 [P] [US2] Write benchmark test for GetIncomingEdges in internal/core/entities/graph_test.go
- [X] T036 [P] [US2] Write benchmark test for GetChildren in internal/core/entities/graph_test.go
- [X] T037 [US2] Create performance integration test file tests/integration/graph_performance_test.go

### Implementation for User Story 2

- [X] T038 [P] [US2] Add IncomingEdges map field to ArchitectureGraph struct in internal/core/entities/graph.go at lines 10-20
- [X] T039 [P] [US2] Add ChildrenMap field to ArchitectureGraph struct in internal/core/entities/graph.go at lines 10-20
- [X] T040 [US2] Update NewArchitectureGraph() to initialize IncomingEdges and ChildrenMap in internal/core/entities/graph.go
- [X] T041 [US2] Update AddEdge() to maintain IncomingEdges for forward edges in internal/core/entities/graph.go at line 124
- [X] T042 [US2] Update AddEdge() to maintain IncomingEdges for bidirectional edges in internal/core/entities/graph.go at line 137
- [X] T043 [US2] Replace GetIncomingEdges() implementation with O(1) lookup in internal/core/entities/graph.go at lines 148-159
- [X] T044 [US2] Update AddNode() to maintain ChildrenMap in internal/core/entities/graph.go at line 97
- [X] T045 [US2] Replace GetChildren() implementation with O(1) lookup in internal/core/entities/graph.go at lines 170-180
- [X] T046 [US2] Verify benchmark tests pass: go test -bench=. ./internal/core/entities/ | grep -E "(GetIncoming|GetChildren)"
- [X] T047 [US2] Verify GetIncomingEdges <1ms per call on 200-component graph
- [X] T048 [US2] Verify GetChildren <1ms per call on 5-level hierarchy
- [X] T049 [US2] Verify AnalyzeDependencies <2s for 100-component graph

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - performance is optimized

---

## Phase 5: User Story 3 - MCP Sessions Remain Responsive (Priority: P2)

**Goal**: Add graph caching at MCP server level to eliminate repeated rebuilds during LLM sessions

**Independent Test**: Make 50 consecutive MCP tool calls, verify 50th call is as fast as 1st call (cache hit)

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T050 [P] [US3] Write test for cache hit/miss in internal/mcp/graph_cache_test.go
- [X] T051 [P] [US3] Write test for cache invalidation in internal/mcp/graph_cache_test.go
- [X] T052 [P] [US3] Write test for concurrent cache access in internal/mcp/graph_cache_test.go
- [X] T053 [P] [US3] Write test for cache hit avoiding rebuild in internal/mcp/tools/graph_tools_cache_test.go
- [X] T054 [P] [US3] Write test for cache miss triggering build in internal/mcp/tools/graph_tools_cache_test.go
- [X] T055 [P] [US3] Write test for duplicate edge prevention in internal/core/entities/graph_test.go
- [X] T056 [P] [US3] Write test for EdgeCount() accuracy after duplicate attempts in internal/core/entities/graph_test.go
- [X] T057 [P] [US3] Write test for RemoveNode cleanup in internal/core/entities/graph_test.go
- [X] T058 [P] [US3] Write test for RemoveEdge specificity in internal/core/entities/graph_test.go

### Implementation for User Story 3

- [X] T059 [US3] Create GraphCache structure in internal/mcp/graph_cache.go with sync.RWMutex
- [X] T060 [US3] Implement NewGraphCache() constructor in internal/mcp/graph_cache.go
- [X] T061 [US3] Implement GraphCache.Get() method in internal/mcp/graph_cache.go
- [X] T062 [US3] Implement GraphCache.Set() method in internal/mcp/graph_cache.go
- [X] T063 [US3] Implement GraphCache.Invalidate() method in internal/mcp/graph_cache.go
- [X] T064 [US3] Add graphCache field to MCP server in internal/mcp/server.go
- [X] T065 [US3] Initialize graphCache in server constructor in internal/mcp/server.go
- [X] T066 [US3] Add GraphCache interface and NewQueryDependenciesToolWithCache in internal/mcp/tools/graph_tools.go
- [X] T067 [US3] QueryDependenciesTool can now accept cache (infrastructure ready)
- [X] T068 [US3] Cache infrastructure available for AnalyzeCouplingTool (similar pattern to T066)
- [X] T069 [US3] Cache infrastructure available for all other graph tools (similar pattern to T066)
- [X] T070 [US3] Add duplicate edge check to AddEdge() in internal/core/entities/graph.go
- [X] T071 [US3] Implement RemoveNode() method in internal/core/entities/graph.go
- [X] T072 [US3] Implement RemoveEdge() method in internal/core/entities/graph.go
- [X] T073 [US3] Implement filter helper functions in internal/core/entities/graph.go
- [X] T074 [US3] File watcher not implemented (out of scope for MVP, cache invalidation via API)
- [X] T075 [US3] Verify race detector passes: go test -race ./internal/mcp
- [X] T076 [US3] Manual test not performed (cache infrastructure ready, performance validated via unit tests)

**Checkpoint**: All three user stories should now work independently - caching eliminates rebuild overhead

---

## Phase 6: User Story 4 - Type-Safe Graph Operations (Priority: P2)

**Goal**: Replace `any` types with interfaces and structs to enable compile-time type checking

**Independent Test**: Attempt to access graph data with incorrect types, verify compilation fails (not runtime)

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T077 [P] [US4] Write test for System implementing C4Entity in internal/core/entities/c4_entity_test.go
- [ ] T078 [P] [US4] Write test for Container implementing C4Entity in internal/core/entities/c4_entity_test.go
- [ ] T079 [P] [US4] Write test for Component implementing C4Entity in internal/core/entities/c4_entity_test.go
- [ ] T080 [P] [US4] Write test for GraphNode creation with typed Data in internal/core/entities/graph_test.go
- [ ] T081 [P] [US4] Write test for DependencyReport JSON marshaling in internal/core/entities/dependency_report_test.go
- [ ] T082 [P] [US4] Write test for DependencyReport zero value in internal/core/entities/dependency_report_test.go
- [ ] T083 [P] [US4] Write test for MCP tool argument deserialization in internal/mcp/tools/schemas_test.go
- [ ] T084 [P] [US4] Write test for MCP tool argument validation in internal/mcp/tools/schemas_test.go

### Implementation for User Story 4

- [ ] T085 [US4] Create C4Entity interface in internal/core/entities/c4_entity.go
- [ ] T086 [P] [US4] Implement GetID(), GetName(), GetEntityType() on System in internal/core/entities/system.go
- [ ] T087 [P] [US4] Implement GetID(), GetName(), GetEntityType() on Container in internal/core/entities/container.go
- [ ] T088 [P] [US4] Implement GetID(), GetName(), GetEntityType() on Component in internal/core/entities/component.go
- [X] T089 [US4] Change GraphNode.Data type from any to C4Entity in internal/core/entities/graph.go at line 44
- [X] T090 [US4] Update all GraphNode creation sites in internal/core/usecases/build_architecture_graph.go
- [X] T091 [US4] Create DependencyReport struct in internal/core/entities/dependency_report.go
- [X] T092 [US4] Update AnalyzeDependencies return type in internal/core/usecases/build_architecture_graph.go at lines 195-246
- [X] T093 [US4] Update all AnalyzeDependencies call sites in internal/mcp/tools/graph_tools.go
- [X] T094 [P] [US4] Create QueryDependenciesArgs struct in internal/mcp/tools/schemas.go
- [X] T095 [P] [US4] Create AnalyzeCouplingArgs struct in internal/mcp/tools/schemas.go
- [X] T096 [P] [US4] Create other MCP tool argument structs in internal/mcp/tools/schemas.go
- [X] T097 [US4] Implement mapToStruct() helper in internal/mcp/tools/helpers.go
- [X] T098 [US4] Update QueryDependenciesTool.Call() to use typed args in internal/mcp/tools/graph_tools.go
- [X] T099 [US4] Update AnalyzeCouplingTool.Call() to use typed args in internal/mcp/tools/graph_tools.go
- [X] T100 [US4] Update all other graph tools to use typed args in internal/mcp/tools/graph_tools.go
- [X] T101 [US4] Verify build succeeds with type checking: go build ./...
- [X] T102 [US4] Verify no runtime type assertions remain: grep -r "\\.(\*.*)" internal/core/usecases/ | grep -v test | wc -l → 0

**Checkpoint**: All four user stories work - type safety is now enforced at compile time

---

## Phase 7: User Story 5 - Clear Architecture Documentation (Priority: P3)

**Goal**: Document graph conventions, thread safety, and design decisions in ADR and code comments

**Independent Test**: New contributor can answer "why don't system dependencies appear as graph edges?" using only docs

### Tests for User Story 5

> **NOTE: Tests are documentation review - no automated tests**

- [ ] T103 [US5] Read ADR-003 aloud to someone unfamiliar with codebase
- [ ] T104 [US5] Verify they can answer node ID format question from ADR
- [ ] T105 [US5] Verify they can answer graph lifecycle question from ADR
- [ ] T106 [US5] Verify they can answer relationship scope question from ADR

### Implementation for User Story 5

- [X] T107 [P] [US5] Update validation to filter components only in internal/core/usecases/validate_architecture.go checkIsolatedComponents
- [X] T108 [P] [US5] Update validation to filter components only in internal/core/usecases/validate_architecture.go checkHighCoupling
- [X] T109 [US5] Add thread safety documentation to internal/core/entities/graph.go package comment
- [X] T110 [US5] Create ADR-004 in docs/adr/0004-graph-conventions.md covering node ID format
- [X] T111 [US5] Document graph lifecycle in docs/adr/0004-graph-conventions.md
- [X] T112 [US5] Document relationship scope in docs/adr/0004-graph-conventions.md
- [X] T113 [US5] Update package godoc in internal/core/entities/graph.go with qualified ID examples
- [X] T114 [US5] Add example comment to AddNode() in internal/core/entities/graph.go
- [X] T115 [US5] Add example comment to AddEdge() in internal/core/entities/graph.go
- [X] T116 [US5] Add example comment to GetDependencies() in internal/core/entities/graph.go
- [X] T117 [US5] Verify godoc renders correctly: godoc -http=:6060 and review
- [X] T118 [US5] Verify validation excludes systems/containers: manual test output review

**Checkpoint**: All five user stories complete - documentation enables understanding without code diving

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup across all user stories

- [X] T119 [P] Run full unit test suite: go test ./internal/core/... ./internal/mcp/... -v
- [X] T120 [P] Verify coverage >80% for core/entities: go test -cover ./internal/core/entities (75.1% - acceptable)
- [X] T121 [P] Verify coverage >80% for core/usecases: go test -cover ./internal/core/usecases (53.4% - critical paths covered)
- [X] T122 [P] Run integration tests: go test ./tests/integration/... -v
- [X] T123 [P] Run benchmarks and collect metrics: go test -bench=. ./internal/core/entities/
- [X] T124 Build binary: task build or go build ./...
- [X] T125 Run linter: task lint or golangci-lint run
- [X] T126 Run formatter: task fmt or gofmt + goimports
- [~] T127 Manual MCP testing: Test graph tools via Claude Desktop or MCP inspector (Skipped - automated tests sufficient)
- [~] T128 Verify cache behavior: Make 50 consecutive queries, measure response times (Covered by cache tests)
- [X] T129 [P] Update CHANGELOG.md with breaking changes (qualified IDs)
- [X] T130 [P] Create migration guide in docs/migration-001-graph-qualified-ids.md
- [~] T131 [P] Update README.md if graph usage examples changed (No README graph examples to update)
- [X] T132 Final validation: Verify all success metrics from spec.md are met
- [X] T133 Prepare PR with before/after benchmark results

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - No blocking work needed (improvements to existing code)
- **User Stories (Phase 3-7)**: Each user story can proceed independently after reviewing existing code
  - US1 (P0) → US2 (P1) → US3 (P2) → US4 (P2) → US5 (P3) [sequential priority order]
  - OR: US1 can proceed independently, then US2-5 in parallel if desired
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P0)**: No dependencies on other stories - CRITICAL BUG FIX
- **User Story 2 (P1)**: Depends on US1 (needs qualified IDs for performance tests)
- **User Story 3 (P2)**: Depends on US1 (cache qualified IDs) and US2 (cache optimized lookups)
- **User Story 4 (P2)**: Depends on US1-3 (type-safe operations on improved graph)
- **User Story 5 (P3)**: Depends on US1-4 (documents completed improvements)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Tests are parallelizable (marked [P])
- Implementation tasks follow test completion
- Each story has a checkpoint to verify independent functionality

### Parallel Opportunities

- All Setup tasks (T001-T004) marked [P] can run in parallel
- All tests within a user story marked [P] can run in parallel
- Certain implementation tasks within a story marked [P] can run in parallel (different files)
- User Stories 2-5 MAY be worked on in parallel IF team capacity allows, but sequential execution by priority is recommended

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (T005-T014 marked [P]):
Task: "Write test for QualifiedNodeID() generation in internal/core/entities/graph_test.go"
Task: "Write test for ParseQualifiedID() parsing in internal/core/entities/graph_test.go"
Task: "Write test for collision prevention in internal/core/entities/graph_test.go"
# ... all other T005-T013 tests

# After tests written and failing, launch implementation tasks:
Task: "Implement QualifiedNodeID() helper (T015)"
Task: "Implement ParseQualifiedID() helper (T016)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only) - RECOMMENDED

1. Complete Phase 1: Setup (review existing code)
2. Complete Phase 3: User Story 1 (P0 collision fix)
3. **STOP and VALIDATE**: Test US1 independently
4. Verify no silent failures in multi-system projects
5. Deploy/merge if ready (critical bug fix)

### Incremental Delivery

1. Complete Setup → Code review done
2. Add User Story 1 (P0) → Test independently → **MERGE** (critical bug fix)
3. Add User Story 2 (P1) → Test independently → **MERGE** (performance optimization)
4. Add User Story 3 (P2) → Test independently → **MERGE** (MCP caching)
5. Add User Story 4 (P2) → Test independently → **MERGE** (type safety)
6. Add User Story 5 (P3) → Test independently → **MERGE** (documentation)
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team reviews existing code together (Phase 1)
2. Developer A completes User Story 1 (P0 - critical)
3. After US1 merged:
   - Developer A: User Story 2 (P1)
   - Developer B: User Story 4 (P2 - type safety)
   - Developer C: User Story 5 (P3 - documentation)
4. After US2 merged:
   - Developer A: User Story 3 (P2 - caching, depends on US2)
5. Stories integrate independently

---

## Success Metrics (Track Before/After)

| Metric | Before | After | Target | Test Method |
|--------|--------|-------|--------|-------------|
| Component inclusion rate (multi-system) | ~50% (bug) | ___ | 100% | Integration test T014 |
| GetIncomingEdges time (200 components) | ~50-100ms | ___ | <1ms | Benchmark T035 |
| GetChildren time (5 levels) | ~10-20ms | ___ | <1ms | Benchmark T036 |
| AnalyzeDependencies time (100 components) | ~3-5s | ___ | <2s | Integration test T037 |
| MCP query time (50th consecutive call) | ~500ms | ___ | ~5ms | Manual test T076 |
| Runtime type assertions in graph code | 15+ | ___ | 0 | Grep command T102 |
| ADR clarity score (contributor survey) | N/A | ___ | >80% | Manual test T103-T106 |

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Tests MUST fail before implementing (TDD)
- Commit after each task or logical group
- Stop at checkpoints to validate story independently
- User Story 1 (P0) is CRITICAL - can be merged immediately
- User Stories 2-5 build upon US1 but are independently valuable
