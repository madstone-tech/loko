# Tasks: TOON Output Format for MCP Query Tools

**Input**: Design documents from `/specs/011-toon-mcp-output/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/mcp-tool-format.md

**Tests**: Tests are REQUIRED per the specification's "Independent Test" criteria and the constitution's Test-First principle. All tests MUST fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing dependencies and project readiness

- [X] T001 Verify toon-go dependency and Encoder implementation in `internal/adapters/encoding/toon.go`
- [X] T002 Verify existing `query_architecture` format support as reference pattern in `internal/mcp/tools/query_architecture.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core helpers and schema updates that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Add `getFormat(args map[string]any) (string, error)` helper in `internal/mcp/tools/helpers.go`
- [X] T004 Add `formatResponse(data map[string]any, format string, encoder usecases.OutputEncoder) (any, error)` helper in `internal/mcp/tools/helpers.go`
- [X] T005 Add `estimateTokenCount(s string) int` helper in `internal/mcp/tools/helpers.go`
- [X] T006 Add unit tests for format helpers: valid formats, invalid format error, empty default, token estimation in `internal/mcp/tools/helpers_test.go`
- [X] T007 Update `schemas.go` to add `format` property with `enum: ["toon", "json"]` and `default: "toon"` for all 7 read tools in `internal/mcp/tools/schemas.go`

**Checkpoint**: Foundation ready — `getFormat()`, `formatResponse()`, and schema updates are tested and working

---

## Phase 3: User Story 1 - TOON Default for Read Tools (Priority: P1) 🎯 MVP

**Goal**: All seven MCP read tools return TOON-encoded responses by default. Information content is preserved.

**Independent Test**: Call each of the seven read tools against the canonical test project with no `format` argument; confirm every response is TOON-encoded and information-equivalent to JSON.

### Tests for User Story 1 (MUST be written FIRST and fail before implementation)

- [X] T008 [P] [US1] Add E2E test: `query_project` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T009 [P] [US1] Add E2E test: `query_architecture` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T010 [P] [US1] Add E2E test: `query_dependencies` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T011 [P] [US1] Add E2E test: `query_related_components` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T012 [P] [US1] Add E2E test: `search_elements` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T013 [P] [US1] Add E2E test: `list_relationships` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T014 [P] [US1] Add E2E test: `analyze_coupling` default returns TOON wrapper in `tests/mcp/tool_format_test.go`
- [X] T015 [US1] Add E2E test: TOON payload decodes to information-equivalent JSON in `tests/mcp/tool_format_test.go`
- [X] T015a [P] [US1] Add E2E test: TOON handles special chars, newlines, nested maps in element names/descriptions in `tests/mcp/tool_format_test.go`

### Implementation for User Story 1

- [X] T016 [P] [US1] Update `query_project` to parse format and wrap response in TOON in `internal/mcp/tools/query_project.go`
- [X] T017 [P] [US1] Update `query_architecture` to standardize on `toon`/`json` enum with TOON default in `internal/mcp/tools/query_architecture.go`
- [X] T018 [P] [US1] Update `query_dependencies` to parse format and wrap response in TOON in `internal/mcp/tools/query_dependencies.go`
- [X] T019 [P] [US1] Update `query_related_components` to parse format and wrap response in TOON in `internal/mcp/tools/query_related_components.go`
- [X] T020 [P] [US1] Update `search_elements` to parse format and wrap response in TOON in `internal/mcp/tools/search_elements.go`
- [X] T021 [P] [US1] Update `list_relationships` to parse format and wrap response in TOON in `internal/mcp/tools/list_relationships.go`
- [X] T022 [P] [US1] Update `analyze_coupling` to parse format and wrap response in TOON in `internal/mcp/tools/analyze_coupling.go`
- [X] T023 [US1] Wire `OutputEncoder` into tool constructors and registry: update each read tool constructor to accept `usecases.OutputEncoder`, update `cmd/mcp.go` to pass `encoding.NewEncoder()` when instantiating tools

**Checkpoint**: All 7 read tools return TOON by default. US1 E2E tests pass.

---

## Phase 4: User Story 2 - JSON Escape Hatch and Error Handling (Priority: P2)

**Goal**: Explicit `format: "json"` returns plain JSON. Invalid format values return clear errors. Mutation/validation tools are unchanged.

**Independent Test**: Call each read tool with `format: "json"`, `format: "toon"`, and `format: "invalid"`; verify JSON/plain/TOON/error behaviors. Call mutation tools and verify no change.

### Tests for User Story 2 (MUST be written FIRST and fail before implementation)

- [X] T024 [P] [US2] Add E2E test: all 7 read tools return plain JSON map when `format: "json"` in `tests/mcp/tool_format_test.go`
- [X] T025 [P] [US2] Add E2E test: all 7 read tools return TOON wrapper when `format: "toon"` explicitly in `tests/mcp/tool_format_test.go`
- [X] T026 [P] [US2] Add E2E test: invalid format returns clear error with allowed values in `tests/mcp/tool_format_test.go`
- [X] T027 [US2] Add E2E test: mutation tools (`create_system`, `update_container`, etc.) ignore format and return JSON in `tests/mcp/tool_format_test.go`
- [X] T028 [US2] Add E2E test: validation tools (`validate`, `validate_diagram`) ignore format and return JSON; verify their `InputSchema()` does NOT expose `format` in `tests/mcp/tool_format_test.go`
- [X] T029 [US2] Update `query_architecture` legacy `"text"` and `"compact"` format values to map to `"toon"` in `internal/mcp/tools/query_architecture.go`

**Checkpoint**: JSON escape hatch works. Invalid format errors are clear. Mutation/validation tools unchanged. US2 E2E tests pass.

---

## Phase 5: User Story 3 - Benchmark Suite with Gating (Priority: P3)

**Goal**: Reproducible benchmark measures token reduction across 7 read-tool payloads and fails if aggregate < 30%.

**Independent Test**: Run `go test ./tests/benchmarks/ -run TestTokenEfficiencyGate -v`; confirm per-payload and aggregate figures reported; confirm failure when threshold not met.

### Tests for User Story 3

- [X] T032 [US3] Create benchmark payload fixtures (5-system, 3-container project) in `tests/benchmarks/token_efficiency_test.go`
- [X] T033 [US3] Implement `TestTokenEfficiencyGate` with per-payload measurement; skip payloads with < 50 tokens (empty/minimal payloads excluded from aggregate per spec edge cases) in `tests/benchmarks/token_efficiency_test.go`
- [X] T034 [US3] Add aggregate reduction calculation and gating logic: fail if aggregate across non-skipped payloads < 30%; report per-payload figures so low-reduction payloads are visible in `tests/benchmarks/token_efficiency_test.go`

### Implementation for User Story 3

- [X] T035 [US3] Generate representative payloads for all 7 read tools using canonical test project in `tests/benchmarks/token_efficiency_test.go`
- [X] T036 [US3] Implement JSON vs TOON encoding measurement helper in `tests/benchmarks/token_efficiency_test.go`
- [X] T037 [US3] Implement benchmark result writer to `research/token-efficiency-benchmarks.md` in `tests/benchmarks/token_efficiency_test.go`
- [X] T038 [US3] Create `research/token-efficiency-benchmarks.md` template with populated results table

**Checkpoint**: Benchmark runs, reports per-payload and aggregate reduction, fails below 30%. US3 tests pass.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify constitution compliance, run full test suite, update docs

- [X] T039 [P] Run full test suite: `go test ./...` — fix any regressions
- [X] T040 [P] Verify handler function effective line counts ≤ 30 for all 7 read tools
- [X] T041 [P] Verify no new external dependencies in `internal/core/`
- [X] T042 Update tool descriptions to mention TOON default in `internal/mcp/tools/` (7 read tool `Description()` methods)
- [X] T043 Update `docs/mcp-integration.md` with TOON default behavior and `format: "json"` escape hatch
- [X] T044 Validate quickstart.md steps against implementation (run through verification checklist)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 must complete before US2 (US2 tests exercise US1 behavior)
  - US3 can run in parallel with US2 once US1 is complete
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories. **This is the MVP.**
- **User Story 2 (P2)**: Can start after US1 — Tests build on US1's TOON default behavior.
- **User Story 3 (P3)**: Can start after US1 — Measures US1's output but is independent of US2.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- For US1: helper foundation (T003-T007) → E2E tests (T008-T015, T015a) → tool implementations (T016-T023)
- For US2: E2E tests (T024-T028) → backward compat (T029)
- For US3: test fixtures (T032-T034) → payload generation (T035-T036) → result writing (T037-T038)

### Parallel Opportunities

- All 7 tool implementations in US1 (T016-T022) can run in parallel once helpers are ready
- All 7 E2E tests in US1 (T008-T014) can run in parallel (different test cases, same test file)
- All US2 E2E tests (T024-T028) can run in parallel
- Polish tasks T039-T041 can run in parallel

---

## Parallel Example: User Story 1

```bash
# After T003-T007 (helpers + schemas) are complete, launch all 7 tool updates together:
Task: "Update query_project to parse format and wrap response in TOON"
Task: "Update query_architecture to standardize on toon/json enum"
Task: "Update query_dependencies to parse format and wrap response in TOON"
Task: "Update query_related_components to parse format and wrap response in TOON"
Task: "Update search_elements to parse format and wrap response in TOON"
Task: "Update list_relationships to parse format and wrap response in TOON"
Task: "Update analyze_coupling to parse format and wrap response in TOON"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 — Write tests first, then implement all 7 tools
4. **STOP and VALIDATE**: Run `go test ./tests/mcp/ -run Test.*Default.*TOON -v`
5. US1 is the MVP — it delivers the core value (30-40% token reduction)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → MVP delivered (TOON default)
3. Add User Story 2 → Test independently → JSON escape hatch + safety
4. Add User Story 3 → Test independently → Benchmark gate + CI protection
5. Polish → Documentation + constitution compliance check

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 tests + 3 tool implementations
   - Developer B: US1 remaining 4 tool implementations
   - Developer C: US2 tests + backward compat
3. Once US1 is complete:
   - Developer A: US3 benchmark suite
   - Developer B/C: US2 implementation + polish

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- The `query_architecture` tool already has partial format support — use it as the reference pattern
- The `Encoder` in `internal/adapters/encoding/` already implements `OutputEncoder` — wire it in, don't recreate it
- Mutation/validation tools intentionally do NOT get `format` parameters — verify they remain unchanged
- Handler function budget: each `Call()` method should delegate format logic to helpers; keep body ≤ 30 effective lines
- Commit after each phase or logical group
- Stop at any checkpoint to validate story independently
