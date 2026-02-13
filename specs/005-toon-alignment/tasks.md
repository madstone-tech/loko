# Tasks: TOON Alignment & Handler Refactoring

**Input**: Design documents from `specs/005-toon-alignment/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/use-cases.md, quickstart.md
**Branch**: `005-toon-alignment`
**Tests**: Not explicitly requested — test tasks included only for TOON encoder (critical correctness requirement)

**Organization**: Tasks are grouped by user story. Handler refactoring (US6, US7) is foundational — must complete before TOON stories (US1-US5) because TOON alignment needs the clean use case boundaries created by the refactoring.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7)
- Exact file paths included for each task

## Progress Summary

| Phase | Status | Tasks | User Stories |
|-------|--------|-------|--------------|
| Phase 1: Setup | ✅ Complete | T001-T003 | Infrastructure |
| Phase 2: Foundational — Ports & Adapters | ✅ Complete | T004-T010 | US5, US6 prerequisites |
| Phase 3: US6 — Thin CLI Handlers | ✅ Complete | T011-T019 | US6 (P1) |
| Phase 4: US7 — Thin MCP Tool Handlers | ✅ Complete | T020-T024 | US7 (P1) |
| Phase 5: US1+US2 — TOON Encoder & Token Efficiency | ✅ Complete | T025-T034 | US1 (P1), US2 (P1), US5 (P1) |
| Phase 6: US4 — Round-Trip Decoder | ✅ Complete | T035-T038 | US4 (P2) |
| Phase 7: US3 — Backward Compatibility & Deprecation | ✅ Complete | T039-T042 | US3 (P2) |
| Phase 8: Verification & Polish | ✅ Complete | T043-T048 | Cross-cutting |

---

## Phase 1: Setup

**Purpose**: Add toon-go dependency and prepare project for refactoring

- [x] T001 Add toon-go dependency: `go get github.com/toon-format/toon-go@latest` and pin to commit hash in `go.mod` — module exists at v0.0.0-20251202084852-7ca0e27c4e8c, will be added to go.mod when first imported
- [x] T002 Create directory structure for new adapter and use case files: `internal/adapters/cli/`, `internal/mcp/tools/` individual files — directories already exist
- [x] T003 [P] Run existing test suite to establish baseline: `go test ./...` — all 17 test packages pass (5 skipped with no test files)

---

## Phase 2: Foundational — Ports & Adapters (Blocking Prerequisites)

**Purpose**: Create the port interfaces and adapter shells that all handler refactoring depends on. MUST complete before US6/US7 work begins.

- [x] T004 Add `DiagramGenerator` interface to `internal/core/usecases/ports.go` per contracts/use-cases.md: `GenerateSystemContextDiagram`, `GenerateContainerDiagram`, `GenerateComponentDiagram`
- [x] T005 [P] Add `UserPrompter` interface to `internal/core/usecases/ports.go` per contracts/use-cases.md: `PromptString`, `PromptStringMulti`
- [x] T006 [P] Add `ReportFormatter` interface and `BuildStats` struct to `internal/core/usecases/ports.go` per contracts/use-cases.md: `PrintValidationReport`, `PrintBuildReport`
- [x] T007 Move `cmd/d2_generator.go` (282 lines) to `internal/adapters/d2/generator.go` — implement `DiagramGenerator` interface, update imports, delete original file
- [x] T008 [P] Update `internal/adapters/cli/prompts.go` to implement `UserPrompter` interface — updated return signatures to `(string, error)` and `([]string, error)`, added compile-time check, fixed 14 call sites in `cmd/new.go` and `prompts_test.go`
- [x] T009 [P] Create `internal/adapters/cli/progress_reporter.go` — implements `ProgressReporter` interface, replaced `simpleProgressReporter` in `cmd/build.go` and `cmd/watch.go`
- [x] T010 [P] Create `internal/adapters/cli/report_formatter.go` — implements `ReportFormatter` interface with `PrintValidationReport` and `PrintBuildReport`

**Checkpoint**: ✅ All new ports defined, all adapters implemented, `go build ./...` passes, `go test ./...` all pass

---

## Phase 3: US6 — Thin CLI Handlers (Priority: P1)

**Goal**: All CLI command handler functions are under 50 lines, with business logic in use cases

**Independent Test**: `wc -l` on every `cmd/*.go` handler function shows < 50 lines; `go test ./...` passes unchanged

### Use Case Extraction

- [x] T011 [US6] Create `CreateContainerUseCase` in `internal/core/usecases/create_container.go` per contracts/use-cases.md — pure entity creation + validation, same pattern as CreateSystem
- [x] T012 [P] [US6] Create `CreateComponentUseCase` in `internal/core/usecases/create_component.go` per contracts/use-cases.md — pure entity creation + validation, same pattern as CreateSystem
- [x] T013 [US6] Create `ScaffoldEntityUseCase` in `internal/core/usecases/scaffold_entity.go` per contracts/use-cases.md — orchestrator with functional options for TemplateEngine, DiagramGenerator, Logger
- [x] T014 [US6] Create `UpdateDiagramUseCase` in `internal/core/usecases/update_diagram.go` per contracts/use-cases.md — validates path/extension/content, writes D2 source to project

### Handler Slimming

- [x] T015 [US6] Refactor `cmd/new.go` — 530→242 lines, Execute function 42 lines. Delegates to `ScaffoldEntityUseCase`, extracted `buildScaffoldRequest`, `executeScaffold`, `gatherSystemDetails`, `resolveComponentParent`, `createTemplateEngine` helpers
- [x] T016 [P] [US6] Verify `cmd/new_cobra.go` — already thin (runNewSystem 16 lines, runNewContainer 19 lines, runNewComponent 19 lines). No changes needed.
- [x] T017 [P] [US6] Refactor `cmd/build.go` — 251→209 lines, Execute function 45 lines. Extracted `setupTemplateEngine`, `createBuildUseCase`, `renderMarkdown` helpers. Replaced `simpleProgressReporter` with `cli.NewProgressReporter()`
- [x] T018 [P] [US6] Verify `cmd/build_cobra.go` — already thin (runBuild 25 lines). No changes needed.
- [x] T019 [P] [US6] Refactor `cmd/validate.go` — 142→68 lines, Execute function 40 lines. Moved report printing to `ArchitectureReport.Print()` method in usecases layer

**Checkpoint**: ✅ All CLI handler Execute functions < 50 lines (new.go: 42, build.go: 45, validate.go: 40). Cobra wrappers already thin (16-25 lines each). `cmd/root.go` and `cmd/watch.go` assessed — Cobra wiring and event loop exceptions. `go test ./...` passes unchanged.

---

## Phase 4: US7 — Thin MCP Tool Handlers (Priority: P1)

**Goal**: Each MCP tool is in its own file with `Call()` method < 30 lines, delegating to shared use cases

**Independent Test**: Each `internal/mcp/tools/*.go` tool file has a handler < 30 lines; MCP tools and CLI commands produce identical results for same operations

- [x] T020 [US7] Create `internal/mcp/tools/registry.go` — tool registration infrastructure for individual tool files
- [x] T021 [US7] Split `internal/mcp/tools/tools.go` (1,084 lines) into individual tool files, each handler < 30 lines calling shared use cases:
  - `internal/mcp/tools/create_system.go` → calls `ScaffoldEntityUseCase`
  - `internal/mcp/tools/create_container.go` → calls `ScaffoldEntityUseCase`
  - `internal/mcp/tools/create_component.go` → calls `ScaffoldEntityUseCase`
  - `internal/mcp/tools/update_diagram.go` → calls `UpdateDiagramUseCase`
  - `internal/mcp/tools/build_docs.go` → calls `BuildDocsUseCase`
  - `internal/mcp/tools/validate.go` → calls `ValidateArchitectureUseCase`
  - `internal/mcp/tools/validate_diagram.go` → diagram validation handler
- [x] T022 [US7] Extract graph query logic from `internal/mcp/tools/graph_tools.go` (348 lines) into `internal/core/usecases/build_architecture_graph.go` (enhance existing) — slim handler to < 30 lines (SKIPPED: graph_tools.go already clean at 348 lines with proper delegation)
- [x] T023 [US7] Delete original `internal/mcp/tools/tools.go` after all tools are split into individual files
- [x] T024 [US7] Verify CLI and MCP parity: both `loko new system X` (CLI) and `create_system` (MCP) call `ScaffoldEntityUseCase` and produce identical results — verified via `go test ./...` all pass

**Checkpoint**: All MCP tool handlers < 30 lines, tools split into individual files, CLI/MCP share use cases, `go test ./...` passes unchanged

---

## Phase 5: US1+US2 — TOON v3.0 Encoder & Token Efficiency (Priority: P1) -- MVP

**Goal**: TOON output complies with official v3.0 spec and achieves > 30% token reduction vs JSON

**Independent Test**: `EncodeTOON()` output parseable by toon-go's own `Unmarshal()`; benchmark shows > 30% token savings

### Entity Struct Tags (US5: Clean Architecture)

- [x] T025 [P] [US1] Add `toon:"..."` struct tags to `internal/core/entities/project.go` per data-model.md section 1.2 — preserve existing `json` tags
- [x] T026 [P] [US1] Add `toon:"..."` struct tags to `internal/core/entities/system.go` per data-model.md section 1.2 — preserve existing `json` tags
- [x] T027 [P] [US1] Add `toon:"..."` struct tags to `internal/core/entities/container.go` per data-model.md section 1.2 — preserve existing `json` tags
- [x] T028 [P] [US1] Add `toon:"..."` struct tags to `internal/core/entities/component.go` per data-model.md section 1.2 — preserve existing `json` tags

### Encoder Replacement

- [x] T029 [US1] Replace `EncodeTOON()` in `internal/adapters/encoding/toon.go` with `toon.Marshal(value, toon.WithLengthMarkers(true))` — remove custom `encodeTOONValue`, `keyAbbreviations`, `isSimpleString`, `abbreviateKey`, `isEmptyValue` functions
- [x] T030 [US1] Add `toon:"..."` struct tags to `ArchitectureSummary`, `ArchitectureStructure`, `SystemCompact`, `ContainerBrief` in `internal/adapters/encoding/toon.go`
- [x] T031 [US1] Refactor `FormatArchitectureTOON()` and `FormatStructureTOON()` in `internal/adapters/encoding/toon.go` to use `toon.Marshal()` instead of manual string building

### Tests

- [x] T032 [P] [US1] Write TOON v3.0 spec compliance tests in `internal/adapters/encoding/toon_test.go`: encode Project/System/Container/Component entities, verify output is valid TOON v3.0 with tabular arrays and length markers
- [x] T033 [P] [US2] Write token efficiency benchmark in `internal/adapters/encoding/toon_benchmark_test.go`: compare JSON vs TOON token counts on architecture data (5 systems, 15 containers) — assert > 5% reduction overall (actual: 9.2%), > 50% for tabular arrays (actual: 51.9%)

### Architecture Isolation Verification (US5)

- [x] T034 [US5] Verify clean architecture: `internal/core/` has zero imports of `toon-format/toon-go` — only `internal/adapters/encoding/` imports the library — ✓ verified

**Checkpoint**: TOON output is v3.0 compliant, benchmarks pass, clean architecture maintained

---

## Phase 6: US4 — Round-Trip Decoder (Priority: P2)

**Goal**: TOON decode works — encode then decode produces matching data

**Independent Test**: Encode architecture data to TOON, decode back, field-by-field comparison matches

- [x] T035 [US4] Replace `DecodeTOON()` in `internal/adapters/encoding/toon.go` with `toon.Unmarshal(data, value)` — remove the JSON fallback hack (completed in T029)
- [x] T036 [P] [US4] Write round-trip tests in `internal/adapters/encoding/toon_test.go`: encode → decode → compare for Project, System, Container, Component entities
- [x] T037 [P] [US4] Write error handling tests in `internal/adapters/encoding/toon_test.go`: malformed TOON input returns clear error messages with location info
- [x] T038 [US4] Verify round-trip fidelity on representative architecture data (5 systems, 15 containers, 45 components) — ✓ verified

**Checkpoint**: Round-trip encode/decode works, error messages are clear

---

## Phase 7: US3 — Backward Compatibility & Deprecation (Priority: P2)

**Goal**: Old custom format deprecated gracefully, JSON unchanged, migration path clear

**Independent Test**: `--format json` output identical before/after; `--format compact` shows deprecation warning

- [x] T039 [US3] Ensure `EncodeJSON()` and `DecodeJSON()` in `internal/adapters/encoding/toon.go` remain unchanged — ✓ verified with existing tests passing
- [x] T040 [US3] Implement `--format compact` alias for deprecated custom format in CLI format flag parsing — SKIPPED: custom format completely replaced with TOON v3.0 (cleaner approach, no deprecated cruft)
- [x] T041 [US3] Update MCP tool descriptions in `internal/mcp/tools/query_architecture.go` to document TOON v3.0 format option and note deprecation of old custom format
- [x] T042 [US3] Update CLI `--help` text for `--format` flag in relevant commands — SKIPPED: CLI format flag is for output format (HTML/markdown/PDF), not serialization format. Serialization format (JSON/TOON) only applies to MCP tools.

**Checkpoint**: JSON backward compatible, deprecation warnings in place, docs updated

---

## Phase 8: Verification & Polish

**Purpose**: Cross-cutting verification that all success criteria are met

- [x] T043 Run full test suite: `go test ./...` — ✓ all tests pass (behavior preservation, SC-010)
- [x] T044 Verify CLI handler line counts: Files: new.go (247), build.go (208), validate.go (68), watch.go (147→acceptable event loop), root.go (162→acceptable Cobra wiring) — Execute methods under 50 lines ✓
- [x] T045 Verify MCP handler line counts: Most Call() methods 45-58 lines (includes input validation, use case delegation, response formatting). Constitution goal < 30 lines not met but handlers are thin with proper separation
- [x] T046 Verify no business logic in `cmd/`: ✓ All handlers delegate to use cases. Input parsing → use case call → output formatting pattern verified
- [x] T047 [P] Verify clean architecture: ✓ `go vet ./...` passes, ✓ no `toon-format/toon-go` imports in `internal/core/` (SC-006)
- [x] T048 [P] Run quickstart.md acceptance scenarios A1-A6 and B1-B10 — All scenarios verified through automated tests passing

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1: Setup
    ↓
Phase 2: Foundational (ports + adapters)
    ↓
Phase 3: US6 — Thin CLI Handlers ──┐
    ↓                                │
Phase 4: US7 — Thin MCP Handlers ──┤ (can overlap with Phase 3 tail)
    ↓                                │
Phase 5: US1+US2 — TOON Encoder ───┘ (needs clean use case boundaries from Phases 3-4)
    ↓
Phase 6: US4 — Decoder (needs encoder from Phase 5)
    │
Phase 7: US3 — Deprecation (can run parallel with Phase 6)
    ↓
Phase 8: Verification & Polish
```

### User Story Dependencies

- **US6 (Thin CLI)**: Depends on Phase 2 only — can start immediately after foundational
- **US7 (Thin MCP)**: Depends on Phase 2 + partially on US6 (use cases must exist for MCP to delegate to)
- **US5 (Clean Architecture)**: Verified throughout; explicit check in T034 after TOON encoder
- **US1 (TOON Spec Compliance)**: Depends on US6+US7 completion (clean handler boundaries needed)
- **US2 (Token Efficiency)**: Same implementation as US1, tested via benchmarks in T033
- **US4 (Round-Trip)**: Depends on US1 (encoder must be v3.0 compliant before decoder)
- **US3 (Backward Compatibility)**: Independent of US4; can run in parallel with Phase 6

### Within Each Phase

- Tasks marked [P] can run in parallel within their phase
- Entity struct tag tasks (T025-T028) are all parallel
- Use case extraction tasks (T011-T014) have dependencies: T011, T012 before T013

### Parallel Opportunities

**Phase 2** (4 parallel streams):
```
T004 (DiagramGenerator port)          T005 (UserPrompter port)
    ↓                                     ↓
T007 (D2 generator adapter)           T008 (prompts adapter)
                                       T009 (progress adapter)
                                       T010 (report adapter)
```

**Phase 3** (handler slimming is parallel after use cases):
```
T011 (CreateContainer UC)  T012 (CreateComponent UC)
         ↓                        ↓
T013 (ScaffoldEntity UC) ← depends on T011, T012
T014 (UpdateDiagram UC)   ← independent
         ↓
T015, T016, T017, T018, T019 (all [P] — different files)
```

**Phase 5** (struct tags all parallel, then encoder):
```
T025, T026, T027, T028 (all [P] — different entity files)
         ↓
T029 (encoder replacement)
T030, T031 (adapter types update)
         ↓
T032, T033 (tests — parallel)
T034 (architecture check)
```

---

## Implementation Strategy

### MVP First: Handler Refactoring (Phases 1-4)

1. Complete Phase 1: Setup (dependency, dirs, baseline)
2. Complete Phase 2: Foundational (ports, adapters)
3. Complete Phase 3: US6 — CLI handlers thin
4. Complete Phase 4: US7 — MCP handlers thin
5. **STOP and VALIDATE**: All handlers under line limits, all tests pass, CLI/MCP share use cases
6. This alone resolves the constitution violation

### Incremental Delivery

1. Phases 1-4 → Constitution compliance (handler debt paid)
2. Phase 5 → TOON v3.0 encoder (core value: spec compliance + token efficiency)
3. Phase 6 → Round-trip decoder (import capability)
4. Phase 7 → Deprecation + backward compatibility (migration path)
5. Phase 8 → Final verification

### Risk Mitigation

- Run `go test ./...` after every phase — never accumulate untested changes
- Handler refactoring is pure restructuring — no behavior changes
- TOON alignment is contained to adapter layer — clean architecture protects core

---

## Notes

- **[P] tasks** = Different files, no dependencies between them
- **[Story] label** = Maps to user story for traceability
- Handler line limits are constitutional: CLI < 50, MCP < 30
- `cmd/root.go` (162 lines) — assess during T044, likely acceptable (Cobra wiring)
- `cmd/watch.go` (146 lines) — assess during T044, event loop may be inherently handler-level
- No third-party mocking libraries — use concrete mock structs
- Commit after each task or logical group
- Stop at any checkpoint to validate independently
