# Tasks: MCP UX Improvements

**Input**: Design documents from `/specs/008-mcp-ux-improvements/`  
**Branch**: `008-mcp-ux-improvements`  
**Tests**: Not explicitly requested — unit tests included per Constitution V (Test-First) since >80% coverage on `internal/core/` is a quality gate.

**Organization**: Tasks grouped by user story. Each phase is independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no competing writes)
- **[Story]**: Maps to user story from spec.md (US1–US5)

## Path Conventions

Go project — single binary layout:

```
internal/core/entities/     ← domain models (zero external deps)
internal/core/usecases/     ← business logic + ports.go
internal/adapters/filesystem/ ← infrastructure implementations
internal/mcp/tools/         ← MCP tool handlers (<100 lines each)
```

---

## Phase 1: Setup

**Purpose**: Confirm baseline is green before any new code lands.

- [X] T001 Verify all existing tests pass: `go test ./...` (must be green before any change)
- [X] T002 Verify lint is clean: `task lint` (must be clean before any change)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: New entity, port, and filesystem adapter that ALL user story phases depend on. No user story work can begin until this phase is complete.

**⚠️ CRITICAL**: US1–US5 all depend on these foundations.

- [X] T003 Add `Relationship` struct and `RelationshipsFile` struct with TOML tags to `internal/core/entities/relationship.go`
- [X] T004 Add `GenerateRelationshipID(source, target, label string) string` (deterministic SHA-256, 8 hex chars, stdlib `crypto/sha256` only) to `internal/core/entities/relationship.go`
- [X] T005 Add `RelationshipOption` func type and option constructors `WithRelType`, `WithRelTechnology`, `WithRelDirection` to `internal/core/entities/relationship.go`
- [X] T006 Add `NewRelationship(source, target, label string, opts ...RelationshipOption) (*Relationship, error)` constructor with all validation rules (empty source/target/label, source==target, invalid type/direction enum) to `internal/core/entities/relationship.go`
- [X] T007 Add `RelationshipToD2Edge(rel Relationship) string` function (sync/async/event syntax, bidirectional arrow) to `internal/core/entities/relationship.go`
- [X] T008 Write unit tests for all of T003–T007 in `internal/core/entities/relationship_test.go` (table-driven; cover all validation branches and edge syntax variants)
- [X] T009 Add `RelationshipRepository` interface (3 methods: `LoadRelationships`, `SaveRelationships`, `DeleteElement`) to `internal/core/usecases/ports.go`
- [X] T010 Implement `FilesystemRelationshipRepository` struct with `LoadRelationships` (reads `src/<systemID>/relationships.toml`; returns empty slice if absent), `SaveRelationships` (atomic write via `.tmp` → `os.Rename`), and `DeleteElement` (load → filter → save) in `internal/adapters/filesystem/relationship_repo.go`
- [X] T011 Write integration tests for `FilesystemRelationshipRepository` using `t.TempDir()` in `internal/adapters/filesystem/relationship_repo_test.go` (round-trip, absent file returns empty, atomic write, DeleteElement removes matching entries)

**Checkpoint**: `go test ./internal/core/entities/ ./internal/adapters/filesystem/` passes — foundation ready.

---

## Phase 3: User Story 1 — Relationship Management (Priority: P1) 🎯 MVP

**Goal**: Architects can create, list, and delete relationships between containers/components via MCP tools. Relationships persist in `relationships.toml`, are reflected in D2 diagrams, and `query_dependencies`/`query_related_components` return live results.

**Independent Test**: Create a system with two containers. Call `create_relationship` → verify response contains `id` and `diagram_updated: true`. Call `list_relationships` → verify relationship appears. Call `query_dependencies` → verify target container returned. Call `delete_relationship` → verify removed from `list_relationships` and `query_dependencies`. All without touching D2 directly.

### Implementation for User Story 1

- [X] T012 [US1] Create concrete mock `MockRelationshipRepository` (implements `RelationshipRepository` port, stores in memory) in `internal/core/usecases/relationship_mock_test.go` for use across all three use case tests
- [X] T013 [US1] Implement `CreateRelationship` use case struct, `CreateRelationshipRequest`, and `Execute` method (validate via `NewRelationship`, dedup by ID, append + save, update D2 file, return entity) in `internal/core/usecases/create_relationship.go`
- [X] T014 [US1] Write unit tests for `CreateRelationship` use case in `internal/core/usecases/create_relationship_test.go` (happy path, idempotent duplicate, validation error, D2 file written)
- [X] T015 [P] [US1] Implement `ListRelationships` use case struct, `ListRelationshipsRequest`, and `Execute` method (load, apply optional source/target filter, return slice) in `internal/core/usecases/list_relationships.go`
- [X] T016 [P] [US1] Write unit tests for `ListRelationships` use case in `internal/core/usecases/list_relationships_test.go` (empty project returns `[]`, source filter, target filter, no filter returns all)
- [X] T017 [P] [US1] Implement `DeleteRelationship` use case struct, `DeleteRelationshipRequest`, and `Execute` method (load, find by ID, error if not found, save, regenerate D2 edges) in `internal/core/usecases/delete_relationship.go`
- [X] T018 [P] [US1] Write unit tests for `DeleteRelationship` use case in `internal/core/usecases/delete_relationship_test.go` (happy path, not-found returns `ErrNotFound`, D2 updated after delete)
- [X] T019 [US1] Extend `BuildArchitectureGraph` use case to load relationships from `RelationshipRepository` and add them as `GraphEdge` entries in `internal/core/usecases/build_architecture_graph.go` (inject `RelationshipRepository` as optional dependency; graph now reflects stored relationships)
- [X] T020 [US1] Write/update unit tests for `BuildArchitectureGraph` with relationships in `internal/core/usecases/build_architecture_graph_test.go` (graph with stored relationship has correct edge count)
- [X] T021 [US1] Implement `CreateRelationshipTool` MCP handler (parse args, call use case, call `graphCache.Invalidate`, return response per contract) in `internal/mcp/tools/create_relationship.go` (≤100 lines)
- [X] T022 [P] [US1] Implement `ListRelationshipsTool` MCP handler (parse args, call use case, return response per contract) in `internal/mcp/tools/list_relationships.go` (≤100 lines)
- [X] T023 [P] [US1] Implement `DeleteRelationshipTool` MCP handler (parse args, call use case, call `graphCache.Invalidate`, return response per contract) in `internal/mcp/tools/delete_relationship.go` (≤100 lines)
- [X] T024 [US1] Register `CreateRelationshipTool`, `ListRelationshipsTool`, `DeleteRelationshipTool` in `cmd/mcp.go` and wire `FilesystemRelationshipRepository` + `GraphCache` into constructors
- [X] T025 [US1] Write MCP handler unit tests for all three relationship tools in `internal/mcp/tools/relationship_tools_test.go` (mock use case responses, verify cache invalidation called on create/delete)

**Checkpoint**: `go test ./internal/core/usecases/ ./internal/mcp/tools/` passes. Smoke test: connect MCP client → `create_relationship` → `list_relationships` → `query_dependencies` all return correct data.

---

## Phase 4: User Story 2 — Container Diagram Initialization (Priority: P2)

**Goal**: `create_container` automatically generates a D2 diagram scaffold, eliminating the "Use 'update_diagram' tool" message and the extra round-trip it causes.

**Independent Test**: Call `create_container` on a new system. Inspect the response — `diagram` field must be `"D2 template created at src/<system>/<container>/container.d2"`. Confirm the file exists on disk. No `update_diagram` call needed.

### Implementation for User Story 2

- [X] T026 [US2] Modify `CreateContainerTool` struct to accept `usecases.DiagramGenerator` as a second constructor parameter in `internal/mcp/tools/create_container.go`
- [X] T027 [US2] Update `CreateContainerTool.Call` to pass `DiagramGenerator` to `ScaffoldEntity` via `usecases.WithDiagramGenerator(t.diagramGenerator)` in `internal/mcp/tools/create_container.go`
- [X] T028 [US2] Update `NewCreateContainerTool` signature in `internal/mcp/tools/registry.go` to pass `d2.NewGenerator()` (concrete `DiagramGenerator`) as the second argument when constructing `CreateContainerTool`
- [X] T029 [US2] Write/update unit tests for `CreateContainerTool` in `internal/mcp/tools/create_container_test.go` asserting: response `diagram` field is a file path (not the old fallback message), and response contains `id` field (FR-015 compliance)

**Checkpoint**: `go test -run TestCreateContainer ./internal/mcp/tools/` passes. `make build` succeeds. Manual smoke test: `create_container` response shows diagram path.

---

## Phase 5: User Story 3 — Batch Component Creation (Priority: P3)

**Goal**: A single `create_components` MCP call creates N components in one round-trip, returning per-item results including slugified IDs. Partial failures do not abort the batch.

**Independent Test**: Call `create_components` with an array of 5 component definitions for an existing container. Confirm response has `created: 5`, `failed: 0`, and each result item has `id` and `status: "created"`. Then call `query_architecture` and confirm all 5 components appear.

### Implementation for User Story 3

- [X] T030 [US3] Implement `CreateComponentsTool` MCP handler with `InputSchema()` (array schema per contract), and `Call()` loop over `components` array calling `ScaffoldEntity` per item, collecting `created`/`failed`/`results` in `internal/mcp/tools/create_components.go` (≤100 lines; extract per-item logic to private `scaffoldOneComponent` helper if needed)
- [X] T031 [US3] Register `CreateComponentsTool` in `internal/mcp/tools/registry.go`
- [X] T032 [US3] Write unit tests for `CreateComponentsTool` in `internal/mcp/tools/create_components_test.go` (all succeed, partial failure continues batch, empty array returns error, each result has `id` field)

**Checkpoint**: `go test -run TestCreateComponents ./internal/mcp/tools/` passes. Smoke test: batch call with 5 components → all appear in `query_architecture`.

---

## Phase 6: User Story 4 — Validate Isolation Suppression (Priority: P4)

**Goal**: `validate` on a freshly initialized project (zero relationships) produces zero `isolated_component` findings.

**Independent Test**: Create a project with 3 components and no relationships. Call `validate`. Assert `report.issues` contains no entry with `code: "isolated_component"`. Then add one relationship and call `validate` again — now unrelated isolated components ARE flagged.

### Implementation for User Story 4

- [X] T033 [US4] Add early-return guard `if graph.EdgeCount() == 0 { return }` at the top of `checkIsolatedComponents` in `internal/core/usecases/validate_architecture.go`
- [X] T034 [US4] Add/update unit tests in `internal/core/usecases/validate_architecture_test.go`: (a) graph with components + zero edges → zero `isolated_component` issues; (b) graph with components + one edge → isolated components ARE flagged for unconnected nodes

**Checkpoint**: `go test -run TestValidateArchitecture ./internal/core/usecases/` passes.

---

## Phase 7: User Story 5 — Slugified ID Suggestions in Error Messages (Priority: P5)

**Goal**: When any MCP tool receives an unrecognized element name, the error message includes the likely slugified ID (`"did you mean 'api-lambda'?"`) or a fallback directing the user to `query_architecture`.

**Independent Test**: Call any tool (e.g., `create_relationship`) with `source: "API Lambda"` where no such element exists. Confirm the error message contains `"did you mean"` and the correct slug. Call with a completely unknown name → confirm the fallback message directs to `query_architecture`.

### Implementation for User Story 5

- [X] T035 [US5] Add `suggestSlugID(input string, graph *entities.ArchitectureGraph) string` helper to `internal/mcp/tools/helpers.go` (normalize with `entities.NormalizeName`, check graph node + ShortIDMap; return empty string if no match or graph nil)
- [X] T036 [US5] Add `notFoundError(entityType, input, suggestion string) error` helper to `internal/mcp/tools/helpers.go` (formats `"<type> \"<input>\" not found — did you mean \"<suggestion>\"?"` or fallback to `query_architecture` message)
- [X] T037 [US5] Write unit tests for `suggestSlugID` and `notFoundError` in `internal/mcp/tools/helpers_test.go` (match found, no match, nil graph, exact slug already correct)
- [X] T038 [P] [US5] Apply `notFoundError` to element lookups in `internal/mcp/tools/update_component.go` replacing bare `fmt.Errorf` not-found errors
- [X] T039 [P] [US5] Apply `notFoundError` to element lookups in `internal/mcp/tools/update_container.go` replacing bare `fmt.Errorf` not-found errors
- [X] T040 [P] [US5] Apply `notFoundError` to element lookups in `internal/mcp/tools/update_system.go` replacing bare `fmt.Errorf` not-found errors
- [X] T041 [P] [US5] Apply `notFoundError` to element lookups in `internal/mcp/tools/graph_tools.go` (`query_dependencies`, `query_related_components`)
- [X] T042 [US5] Apply `notFoundError` to element lookups in all three new relationship tool handlers (`create_relationship.go`, `list_relationships.go`, `delete_relationship.go`) for system/source/target not-found errors

**Checkpoint**: `go test ./internal/mcp/tools/` passes. Smoke test: call any tool with a display-name input → error includes slug hint.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates, wiring verification, and constitution compliance check.

- [X] T043 Run full test suite and confirm >80% coverage on `internal/core/`: `task coverage` (must pass quality gate)
- [X] T044 Run lint: `task lint` — fix any new violations introduced by this feature
- [X] T045 Run build: `make build` — confirm single binary compiles cleanly
- [X] T046 [P] Audit new MCP handler line counts: confirm `create_relationship.go`, `list_relationships.go`, `delete_relationship.go`, `create_components.go` each stay ≤100 lines of handler logic (exclude schema definitions)
- [X] T047 [P] Verify `internal/core/` has zero new external imports: `go list -deps ./internal/core/...` — no new non-stdlib packages
- [X] T048 [P] Verify `FR-015` compliance: inspect response maps in all create tool handlers (`create_system`, `create_container`, `create_component`, `create_relationship`) confirm `id` key is always present
- [ ] T049 Execute quickstart.md 10-step validation sequence end-to-end against a real loko project to confirm all user stories pass their independent tests

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — run immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — **BLOCKS all user stories**
- **Phase 3 (US1)**: Depends on Phase 2 — highest priority, implement first
- **Phase 4 (US2)**: Depends on Phase 2 only — no dependency on US1
- **Phase 5 (US3)**: Depends on Phase 2 only — no dependency on US1 or US2
- **Phase 6 (US4)**: Depends on Phase 2 only — single file change, fast
- **Phase 7 (US5)**: Depends on Phase 2 only; benefits from Phase 3 tools being present for smoke testing
- **Phase 8 (Polish)**: Depends on all user story phases complete

### User Story Dependencies

- **US1 (P1)**: No dependency on other user stories. Foundational phase must be complete.
- **US2 (P2)**: No dependency on US1. Can start after Foundational phase.
- **US3 (P3)**: No dependency on US1 or US2. Can start after Foundational phase.
- **US4 (P4)**: No dependency on any user story. Two-task phase, fast.
- **US5 (P5)**: No dependency on US1–US4. Benefits from relationship tools being available for integration testing.

### Within Each Phase

- Entity/struct tasks before use case tasks
- Use case tasks before MCP handler tasks
- Handler tasks before registry wiring
- Tests written alongside (or immediately before) implementation per Constitution V

### Parallel Opportunities

- T015, T017 (ListRelationships, DeleteRelationship use cases) are parallel with each other after T013 foundation
- T022, T023 (List/Delete MCP handlers) are parallel after T021 pattern established
- T026–T028 (US2) entirely parallel with US1 work (different files)
- T030–T032 (US3) entirely parallel with US1 and US2 (different files)
- T033–T034 (US4) entirely parallel with all other user stories (single function change)
- T035–T037 (US5 helpers) parallel with US1–US4
- T038–T041 (US5 tool updates) all parallel with each other (different files)
- T043–T048 (Polish) marked [P] tasks run in parallel

---

## Parallel Execution Examples

### Phase 2 (Foundational) — sequential within entity, then adapter

```
T003 → T004 → T005 → T006 → T007 → T008 (entity file — sequential)
T009 (port — parallel with T008 once T006 complete)
T010 → T011 (adapter — after T009)
```

### Phase 3 (US1) — parallel use cases after mock

```
T012 (mock — shared dependency)
T013 → T014 (CreateRelationship)
T015 → T016 (ListRelationships)    ← parallel with T013/T014
T017 → T018 (DeleteRelationship)   ← parallel with T013/T014 and T015/T016
T019 → T020 (BuildArchitectureGraph extension — after T012)
T021 (create handler — after T013)
T022 (list handler — after T015)   ← parallel with T021
T023 (delete handler — after T017) ← parallel with T021, T022
T024 (registry wiring — after T021, T022, T023)
T025 (handler tests — after T024)
```

### US2, US3, US4, US5 — all parallel with each other after Phase 2

```
Phase 4 (T026–T029): independent of Phase 3
Phase 5 (T030–T032): independent of Phase 3 and 4
Phase 6 (T033–T034): independent of all user stories
Phase 7 (T035–T042): independent of all user stories; T038–T041 parallel within phase
```

---

## Implementation Strategy

### MVP First (US1 Only — Phases 1–3)

1. Complete Phase 1: Setup (green baseline)
2. Complete Phase 2: Foundational (entity + port + adapter)
3. Complete Phase 3: US1 (create/list/delete relationship tools)
4. **STOP and VALIDATE**: `create_relationship` → `list_relationships` → `query_dependencies` all work end-to-end
5. Merge or demo — relationships are now first-class

### Incremental Delivery

1. Phases 1–2: Foundation → `go test ./...` green
2. Phase 3 (US1): Relationship tools → demo to user, confirm graph queries work
3. Phase 4 (US2): Container diagram init → eliminates extra round-trip on create
4. Phase 5 (US3): Batch components → validate with 20-component project
5. Phase 6 (US4): Validate suppression → zero false alarms on fresh projects
6. Phase 7 (US5): Slug suggestions → better DX for agents and users
7. Phase 8: Polish → `task coverage`, `task lint`, `make build` all green

### Single Developer Order (priority order, phases sequential)

```
Phase 1 → Phase 2 → Phase 3 → Phase 6 → Phase 4 → Phase 5 → Phase 7 → Phase 8
```

Phase 6 (US4, 2 tasks) is inserted early because it is tiny and immediately improves the first-run experience when testing Phase 3 work.

---

## Notes

- `[P]` tasks operate on different files — no write conflicts
- Each user story phase delivers independently verifiable value
- Constitution V (Test-First): tests are included per quality gate requirement, not as optional extras
- Constitution III (Thin Handlers): all new MCP handlers must stay ≤100 lines — enforce at T046
- Constitution I (Clean Architecture): `internal/core/` must have zero new external imports — enforce at T047
- All new files follow existing naming conventions: one concern per file, `_test.go` alongside source
