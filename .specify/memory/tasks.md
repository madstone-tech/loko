---
description: "Task list for loko v0.1.0 implementation"
---

# Tasks: loko v0.1.0 - C4 Architecture Documentation Tool

**Spec Version**: 0.1.0-dev
**Status**: In Progress
**Last Updated**: 2026-02-06 (Phase 2 tasks generated via /speckit.tasks)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Task can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US-1, US-2, etc.)
- Exact file paths included for clarity

## Progress Summary

| Phase | Status | Tasks | User Stories |
|-------|--------|-------|--------------|
| Phase 1: Foundation | ✅ Complete | T001-T003 + #002, #003 | Setup/Ports |
| Phase 2A: Handler Refactoring | 🟡 Tasks Generated | T001-T024 (feature) | US-6, US-7 (Constitution) |
| Phase 2B: TOON Alignment | 🟡 Tasks Generated | T025-T042 (feature) | US-1, US-2, US-3, US-4, US-5 |
| Phase 3: US-3 Scaffolding | 🔲 Not Started | T004-T014 | US-3 (P1) |
| Phase 4: US-2 File Editing | 🔲 Not Started | T015-T026 | US-2 (P1) |
| Phase 5: US-1 MCP Design | 🔲 Not Started | T027-T034 | US-1 (P1) |
| Phase 6: US-4 HTTP API | 🔲 Not Started | T041-T047 | US-4 (P2) |
| Phase 7: US-5 Multi-Format | 🔲 Not Started | T048-T052 | US-5 (P2) |
| Phase 8: Polish | 🔲 Not Started | T053-T062 | Cross-cutting |

---

## Phase 1: Foundation (COMPLETE)

- [x] T001 Initialize Go project with Clean Architecture
- [x] T002 Implement core domain entities (Project, System, Container, Component)
- [x] T003 Define use case port interfaces in `internal/core/usecases/ports.go` (18 interfaces)
- [x] Cobra/Viper CLI migration with shell completions, config hierarchy, aliases (PR #5)
- [x] Serverless architecture template with `-template` flag (PR #4)

**Checkpoint**: Foundation complete — all ports defined, CLI framework migrated, templates ready

---

## Phase 2: Handler Refactoring + TOON Alignment (#005)

**Purpose**: Pay down handler debt (10 files violating constitution) and align TOON encoder with v3.0 spec.
**Detailed task breakdown**: `specs/005-toon-alignment/tasks.md` (48 tasks, 8 phases)
**Spec**: `specs/005-toon-alignment/spec.md` (7 User Stories)

### Summary (see feature tasks.md for full details)

| Feature Phase | Tasks | User Stories | Description |
|---------------|-------|--------------|-------------|
| Setup | T001-T003 | — | Add toon-go dep, create dirs, baseline tests |
| Foundational | T004-T010 | US5, US6 prereqs | New ports (DiagramGenerator, UserPrompter, ReportFormatter), move adapters |
| US6: Thin CLI | T011-T019 | US6 (P1) | Extract use cases, slim 8 CLI handlers to < 50 lines |
| US7: Thin MCP | T020-T024 | US7 (P1) | Split tools.go, slim MCP handlers to < 30 lines |
| US1+US2: TOON Encoder | T025-T034 | US1, US2, US5 (P1) | Struct tags, toon-go encoder, benchmarks |
| US4: Decoder | T035-T038 | US4 (P2) | Round-trip decode support |
| US3: Deprecation | T039-T042 | US3 (P2) | Backward compat, --format compact deprecation |
| Verification | T043-T048 | Cross-cutting | Line counts, architecture checks, quickstart validation |

**Checkpoint**: All handlers under constitutional line limits, TOON output validates against spec, all tests passing

---

## Phase 3: US-3 Project Scaffolding (P1)

**Goal**: `loko init` and `loko new` commands work end-to-end via proper use cases
**Depends on**: Phase 2A (use cases extracted from handlers)

### Tests (if requested)

- [ ] T004 [P] [US-3] Unit test for CreateSystem use case in `internal/core/usecases/create_system_test.go`
- [ ] T005 [P] [US-3] Unit test for template engine in `internal/adapters/ason/engine_test.go`
- [ ] T006 [P] [US-3] Integration test for full init→new workflow in `tests/integration/scaffolding_test.go`

### Implementation

- [ ] T007 [US-3] Wire CreateSystem use case (from T100) with ason template engine adapter
- [ ] T008 [US-3] Implement ason template engine adapter in `internal/adapters/ason/engine.go`
- [ ] T009 [US-3] Implement filesystem project repository in `internal/adapters/filesystem/project_repo.go`
- [ ] T010 [US-3] Wire up CLI commands in `main.go` with dependency injection
- [ ] T011 [US-3] Verify `cmd/init.go` works as thin handler with use case
- [ ] T012 [US-3] Verify `cmd/new.go` works as thin handler (from T100)
- [ ] T013 [US-3] Verify starter templates work (standard-3layer, serverless)
- [ ] T014 [US-3] Add TOML config loader in `internal/adapters/config/loader.go`

**Checkpoint**: User Story 3 complete - scaffolding works independently

---

## Phase 4: US-2 File Editing & Watch Mode (P1)

**Goal**: Direct file editing with hot-reload
**Depends on**: Phase 2A (BuildDocs use case extracted)

### Tests (if requested)

- [ ] T015 [P] [US-2] Unit test for BuildDocs use case in `internal/core/usecases/build_docs_test.go`
- [ ] T016 [P] [US-2] Integration test for D2 rendering in `tests/integration/diagram_rendering_test.go`
- [ ] T017 [P] [US-2] Integration test for incremental builds in `tests/integration/incremental_build_test.go`

### Implementation

- [ ] T018 [US-2] Implement D2 diagram renderer adapter in `internal/adapters/d2/renderer.go`
- [ ] T019 [US-2] Wire BuildDocs use case (from T102) with D2 and HTML adapters
- [ ] T020 [US-2] Implement HTML site builder adapter in `internal/adapters/html/builder.go`
- [ ] T021 [US-2] Create HTML templates in `internal/adapters/html/templates/`
- [ ] T022 [US-2] Implement file watcher adapter in `internal/adapters/filesystem/watcher.go`
- [ ] T023 [US-2] Verify `cmd/build.go` works as thin handler (from T102)
- [ ] T024 [US-2] Verify `cmd/serve.go` works as thin handler
- [ ] T025 [US-2] Verify `cmd/watch.go` works as thin handler (from T105)
- [ ] T026 [US-2] Verify `cmd/validate.go` works as thin handler (from T106)

**Checkpoint**: User Story 2 complete - file editing + watch mode works

---

## Phase 5: US-1 LLM-Driven Architecture Design (P1)

**Goal**: MCP server with core tools for conversational design
**Depends on**: Phase 2A (MCP tools extracted), Phase 3 (scaffolding use cases)

### Tests (if requested)

- [ ] T027 [P] [US-1] Unit test for QueryArchitecture use case
- [ ] T028 [P] [US-1] Unit test for MCP server request handling

### Implementation

- [ ] T029 [US-1] Enhance QueryArchitecture use case with token-efficient formatting
- [ ] T030 [US-1] Verify MCP server works with thin tool handlers (from T108-T109)
- [ ] T031 [US-1] Verify all MCP tools delegate to shared use cases
- [ ] T032 [US-1] Verify `cmd/mcp.go` works as thin handler
- [ ] T033 [US-1] Generate JSON schemas for all MCP tool inputs
- [ ] T034 [US-1] Add structured logging adapter

**Checkpoint**: User Story 1 complete - MCP integration with Claude works

---

## Phase 6: US-4 HTTP API (P2)

**Goal**: CI/CD teams can trigger builds via HTTP API
**Depends on**: Foundation + BuildDocs use case

- [ ] T041 [US-4] Implement HTTP API server in `internal/api/server.go`
- [ ] T042 [US-4] Create API auth middleware
- [ ] T043 [US-4] Implement API handlers (thin, < 50 lines each, reuse use cases)
- [ ] T044 [US-4] Create API response models
- [ ] T045 [US-4] Verify `cmd/api.go` works as thin handler
- [ ] T046 [US-4] Generate OpenAPI spec
- [ ] T047 [US-4] Add API documentation

**Checkpoint**: User Story 4 complete - API works for CI/CD

---

## Phase 7: US-5 Multi-Format Export (P2)

**Goal**: Users can export to HTML, Markdown, and PDF
**Depends on**: Foundation + BuildDocs use case

- [ ] T048 [US-5] Create MarkdownBuilder adapter
- [ ] T049 [US-5] Create PDFRenderer adapter
- [ ] T050 [US-5] Enhance BuildDocs use case for format selection
- [ ] T051 [US-5] Add `--format` flag to build command
- [ ] T052 [US-5] Add export format configuration to loko.toml

**Checkpoint**: User Story 5 complete - multi-format export works

---

## Phase 8: Polish & Cross-Cutting

- [ ] T053 [P] Write quickstart tutorial in `docs/quickstart.md`
- [ ] T054 [P] Write configuration reference in `docs/configuration.md`
- [ ] T055 [P] Create example projects in `examples/`
- [ ] T056 [P] Write MCP integration guide in `docs/mcp-integration.md`
- [ ] T057 CI job to build and test all examples
- [ ] T058 Comprehensive error messages with lipgloss formatting
- [ ] T059 [P] Additional unit tests to reach > 80% coverage
- [ ] T060 Code cleanup and refactoring based on review feedback
- [ ] T061 Performance optimization
- [ ] T062 Quickstart validation (follow docs, verify they work)

---

## Dependencies & Execution Order

### Critical Path

```
Phase 1 (Foundation) ✅
    ↓
Phase 2A (Handler Refactoring) ──→ Phase 2B (TOON Alignment)
    ↓
Phase 3 (US-3 Scaffolding) ──┐
Phase 4 (US-2 Watch Mode) ───┼─→ Phase 5 (US-1 MCP)
                              │
Phase 6 (US-4 API) ←─────────┘
Phase 7 (US-5 Export) ←───────┘
    ↓
Phase 8 (Polish)
```

### Key Change from Original Plan

Phase 2 (Handler Refactoring) was inserted because:
1. 10 files violate the constitution's thin handler principle
2. New features would compound the debt
3. Extracting use cases now creates the foundation Phase 3-5 need
4. CLI and MCP will share use cases instead of duplicating logic

---

## Notes

- **[P] tasks** = Different files, no dependencies between them
- **[Story] label** = Maps to user story for traceability
- Each user story should be independently completable and testable
- **Test files**: Write tests FIRST, ensure they FAIL before implementation
- **Commit after each task** or logical group
- **Stop at any checkpoint** to validate story independently
- Use `task test`, `task lint` before commits
- No third-party mocking libraries — use concrete mocks (see MockProjectRepo pattern)
- Handler line limits are constitutional: CLI < 50, MCP < 30
