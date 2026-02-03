---

description: "Task list for loko v0.1.0 implementation"
---

# Tasks: loko v0.1.0 - C4 Architecture Documentation Tool

**Spec Version**: 0.1.0-dev  
**Status**: In Progress  
**Last Updated**: 2025-12-17

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Task can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US-1, US-2, etc.)
- Exact file paths included for clarity

## Progress Summary

| Phase | Status | Tasks | User Stories |
|-------|--------|-------|--------------|
| Phase 1: Foundation | ✅ 2/3 | T001-T003 | Setup/Ports |
| Phase 2: US-3 Scaffolding | 🔲 Not Started | T004-T009 | US-3 (P1) |
| Phase 3: US-2 File Editing | 🔲 Not Started | T010-T018 | US-2 (P1) |
| Phase 4: US-1 MCP Design | 🔲 Not Started | T019-T027 | US-1 (P1) |
| Phase 5: US-6 Token Queries | 🔲 Not Started | T028-T032 | US-6 (P1) |
| Phase 6: US-4 HTTP API | 🔲 Not Started | T033-T040 | US-4 (P2) |
| Phase 7: US-5 Multi-Format | 🔲 Not Started | T041-T047 | US-5 (P2) |
| Phase 8: Polish | 🔲 Not Started | T048-T052 | Cross-cutting |

---

## Phase 1: Foundation (Shared Infrastructure)

**Purpose**: Core interfaces and domain model - unblocks all user stories

### ✅ Completed

- [x] T001 Initialize Go project with Clean Architecture
- [x] T002 Implement core domain entities (Project, System, Container, Component)

### Remaining

- [ ] T003 [P] Define use case port interfaces in `internal/core/usecases/ports.go`
  - ProjectRepository (load/save)
  - TemplateEngine (render templates)
  - DiagramRenderer (render D2)
  - SiteBuilder (generate HTML)
  - FileWatcher (watch for changes)
  - Logger (structured logging)
  - ProgressReporter (feedback)
  - OutputEncoder (JSON/TOON)
  - PDFRenderer (optional)
  - Validation helpers

**Checkpoint**: All ports defined → adapters can be implemented in parallel

---

## Phase 2: US-3 Project Scaffolding (P1)

**Goal**: `loko init` and `loko new` commands work end-to-end  
**Independent Test**: User can run `loko init myproject && loko new system PaymentService && ls src/PaymentService/` and see generated files

### Tests (if requested)

- [ ] T004 [P] [US-3] Unit test for CreateSystem use case in `internal/core/usecases/create_system_test.go`
- [ ] T005 [P] [US-3] Unit test for template engine in `internal/adapters/ason/engine_test.go`
- [ ] T006 [P] [US-3] Integration test for full init→new workflow in `tests/integration/scaffolding_test.go`

### Implementation

- [ ] T007 [US-3] Create CreateSystem use case in `internal/core/usecases/create_system.go` (input validation, template loading, project saving)
- [ ] T008 [US-3] Implement ason template engine adapter in `internal/adapters/ason/engine.go` (template discovery from ~/.loko/templates/ and .loko/templates/)
- [ ] T009 [US-3] Implement filesystem project repository in `internal/adapters/filesystem/project_repo.go` (TOML loading, YAML frontmatter, directory creation)
- [ ] T010 [US-3] Wire up CLI commands in `main.go` with dependency injection (ProjectRepository → TemplateEngine → CreateSystem UC)
- [ ] T011 [US-3] Implement `cmd/init.go` - `loko init` command (interactive prompts, project setup)
- [ ] T012 [US-3] Implement `cmd/new.go` - `loko new system|container|component` commands (thin wrapper, <50 lines)
- [ ] T013 [US-3] Create starter templates in `templates/` directory (standard-3layer, serverless with ason syntax)
- [ ] T014 [US-3] Add TOML config loader in `internal/adapters/config/loader.go` (parse loko.toml, defaults)

**Checkpoint**: User Story 3 complete - scaffolding works independently

---

## Phase 3: US-2 File Editing & Watch Mode (P1)

**Goal**: Direct file editing with hot-reload  
**Independent Test**: User can `loko watch`, edit a .d2 file, and see auto-refresh within 500ms

### Tests (if requested)

- [ ] T015 [P] [US-2] Unit test for BuildDocs use case in `internal/core/usecases/build_docs_test.go`
- [ ] T016 [P] [US-2] Integration test for D2 rendering in `tests/integration/diagram_rendering_test.go`
- [ ] T017 [P] [US-2] Integration test for incremental builds in `tests/integration/incremental_build_test.go`

### Implementation

- [ ] T018 [US-2] Implement D2 diagram renderer adapter in `internal/adapters/d2/renderer.go` (shell to d2 CLI, caching, error handling)
- [ ] T019 [US-2] Create BuildDocs use case in `internal/core/usecases/build_docs.go` (orchestrate rendering, track progress, incremental logic)
- [ ] T020 [US-2] Implement HTML site builder adapter in `internal/adapters/html/builder.go` (generate static site with sidebar, breadcrumbs, search)
- [ ] T021 [US-2] Create HTML templates in `internal/adapters/html/templates/` (layout.html, index.html, system.html, container.html)
- [ ] T022 [US-2] Implement file watcher adapter in `internal/adapters/filesystem/watcher.go` (fsnotify integration)
- [ ] T023 [US-2] Implement `cmd/build.go` - `loko build` command (call BuildDocs, format output)
- [ ] T024 [US-2] Implement `cmd/serve.go` - `loko serve` command (HTTP server on localhost:8080, serve dist/)
- [ ] T025 [US-2] Implement `cmd/watch.go` - `loko watch` command (file watcher, rebuild on change, <500ms latency)
- [ ] T026 [US-2] Implement `cmd/validate.go` - `loko validate` command (check for orphaned refs, missing files, hierarchy violations)

**Checkpoint**: User Story 2 complete - file editing + watch mode works

---

## Phase 4: US-1 LLM-Driven Architecture Design (P1)

**Goal**: MCP server with core tools for conversational design  
**Independent Test**: Claude Desktop can use loko MCP to design a 3-system architecture end-to-end

### Tests (if requested)

- [ ] T027 [P] [US-1] Unit test for QueryArchitecture use case in `internal/core/usecases/query_architecture_test.go` (token counting)
- [ ] T028 [P] [US-1] Unit test for MCP server request handling in `internal/mcp/server_test.go`

### Implementation

- [ ] T029 [US-1] Create QueryArchitecture use case in `internal/core/usecases/query_architecture.go` (summary ~200 tokens, structure ~500 tokens, full/targeted responses)
- [ ] T030 [US-1] Implement MCP server in `internal/mcp/server.go` (stdio transport, JSON-RPC, protocol handler)
- [ ] T031 [US-1] Create MCP tool handlers in `internal/mcp/tools/` (query_project.go, query_architecture.go, create_system.go, create_container.go, create_component.go, update_diagram.go, build_docs.go, validate.go)
- [ ] T032 [US-1] Implement `cmd/mcp.go` - `loko mcp` command (start MCP server)
- [ ] T033 [US-1] Generate JSON schemas for all MCP tool inputs in `internal/mcp/tools/schemas.go`
- [ ] T034 [US-1] Add structured logging in `internal/adapters/logging/logger.go` (JSON format, configurable level)

**Checkpoint**: User Story 1 complete - MCP integration with Claude works

---

## Phase 5: US-6 Token-Efficient Architecture Queries (P1)

**Goal**: LLM can query architecture without excessive token overhead  
**Independent Test**: Query architecture for 20-system project returns <300 tokens for summary, <600 for structure

### Tests (if requested)

- [ ] T035 [P] [US-6] Benchmark token consumption in `tests/benchmarks/token_efficiency_test.go` (summary, structure, full)

### Implementation

- [ ] T036 [US-6] Enhance QueryArchitecture use case with compression in `internal/core/usecases/query_architecture.go` (add "compressed" detail level option)
- [ ] T037 [US-6] Create compressed notation formatter in `internal/adapters/encoding/notation.go` (custom format for compact representation)
- [ ] T038 [US-6] Add format parameter to QueryArchitectureInput
- [ ] T039 [US-6] Update MCP tool schema for detail levels in `internal/mcp/tools/schemas.go`
- [ ] T040 [US-6] Document token efficiency strategy in `docs/token-efficiency.md` (examples, comparisons)

**Checkpoint**: User Story 6 complete - token-efficient queries verified

---

## Phase 6: US-4 API Integration (P2)

**Goal**: CI/CD teams can trigger builds via HTTP API  
**Independent Test**: CI pipeline can POST to /api/v1/build and get JSON response with build status

### Implementation

- [ ] T041 [US-4] Implement HTTP API server in `internal/api/server.go` (router setup, middleware)
- [ ] T042 [US-4] Create API auth middleware in `internal/api/middleware/auth.go` (bearer token, validate against config)
- [ ] T043 [US-4] Implement API handlers in `internal/api/handlers/` (build.go for POST /build, systems.go for GET /systems, validate.go for GET /validate)
- [ ] T044 [US-4] Create API response models in `internal/api/models.go` (BuildResponse, SystemResponse, ValidateResponse)
- [ ] T045 [US-4] Implement `cmd/api.go` - `loko api` command (start API server on configurable port)
- [ ] T046 [US-4] Generate OpenAPI spec in `internal/api/openapi.yaml` (auto-document all endpoints)
- [ ] T047 [US-4] Add API documentation in `docs/api-reference.md` (auth, endpoints, examples)

**Checkpoint**: User Story 4 complete - API works for CI/CD

---

## Phase 7: US-5 Multi-Format Export (P2)

**Goal**: Users can export to HTML, Markdown, and PDF  
**Independent Test**: User runs `loko build --format markdown` and gets single README.md with all content

### Implementation

- [ ] T048 [US-5] Create MarkdownBuilder adapter in `internal/adapters/markdown/builder.go` (generate single README.md, proper hierarchy)
- [ ] T049 [US-5] Create PDFRenderer adapter in `internal/adapters/pdf/renderer.go` (shell to veve-cli, error if missing)
- [ ] T050 [US-5] Enhance BuildDocs use case to support format selection in `internal/core/usecases/build_docs.go` (HTML, Markdown, PDF)
- [ ] T051 [US-5] Add `--format` flag to `cmd/build.go` (default to all enabled formats from loko.toml)
- [ ] T052 [US-5] Add export format configuration to loko.toml in `internal/adapters/config/loader.go` (outputs.html, outputs.markdown, outputs.pdf)

**Checkpoint**: User Story 5 complete - multi-format export works

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements affecting multiple user stories

- [ ] T053 [P] Write quickstart tutorial in `docs/quickstart.md` (5-minute walkthrough)
- [ ] T054 [P] Write configuration reference in `docs/configuration.md` (loko.toml all options)
- [ ] T055 [P] Create example projects in `examples/` (simple-project, 3layer-app, serverless)
- [ ] T056 [P] Write MCP integration guide in `docs/mcp-integration.md` (setup Claude, use loko tools)
- [ ] T057 CI job to build and test all examples
- [ ] T058 Add comprehensive error messages with `lipgloss` formatting in `internal/ui/output.go`
- [ ] T059 [P] Additional unit tests to reach >80% coverage in `internal/core/`
- [ ] T060 Code cleanup and refactoring based on review feedback
- [ ] T061 Performance optimization for diagram rendering and builds
- [ ] T062 Run quickstart.md validation (follow docs, verify they work)

---

## Dependencies & Execution Order

### Critical Path (Unblocked by each phase)

```
Phase 1 (Foundation)
    ↓
Phase 2 (US-3 Scaffolding) ──┐
Phase 3 (US-2 Watch Mode) ───┼─→ Phase 4 (US-1 MCP)
Phase 5 (US-6 Queries) ──────┘
    ↓
Phase 6 (US-4 API) [depends on Foundation + BuildDocs]
Phase 7 (US-5 Export) [depends on Foundation + BuildDocs]
    ↓
Phase 8 (Polish)
```

### User Story Dependencies

- **US-3 (Scaffolding)**: Depends on Foundation only → Can start immediately after T003
- **US-2 (Watch Mode)**: Depends on Foundation only → Can start immediately after T003
- **US-1 (MCP)**: Depends on US-3 (CreateSystem reuse) + US-2 (BuildDocs reuse)
- **US-6 (Queries)**: Depends on Foundation only → Can parallelize with others
- **US-4 (API)**: Depends on Foundation + BuildDocs (from US-2) → Can start after US-2
- **US-5 (Export)**: Depends on Foundation + BuildDocs → Can start after US-2

### Within-Story Parallelization

#### US-3 Scaffolding
```
T004, T005, T006 → T007, T008, T009 → T010, T011, T012, T013, T014
Tests in parallel  →  Models/Services in parallel  →  CLI wiring
```

#### US-2 Watch Mode
```
T015, T016, T017 → T018 (D2), T019 (BuildDocs), T020 (HTML) → T021, T022 → T023-T026 (CLI)
Tests            →  Core adapters (parallel)              →  Templates  →  Commands
```

#### US-1 MCP
```
T027, T028 → T029 (QueryArch), T030 (Server) → T031-T034 (tools + schemas)
Tests      →  Core logic (parallel)          →  Handlers (parallel)
```

---

## Implementation Strategies

### MVP First (User Story 3 Only)

1. Complete Phase 1: Foundation (T003)
2. Complete Phase 2: US-3 Scaffolding (T004-T014)
3. **STOP and VALIDATE**: User can scaffold projects
4. Deploy/demo scaffolding as MVP

### Incremental Delivery (Recommended)

1. **Slice 1**: US-3 Scaffolding (Phase 2) → `loko init` and `loko new` work
2. **Slice 2**: US-2 Watch Mode (Phase 3) → `loko build` and `loko watch` work
3. **Slice 3**: US-1 MCP (Phase 4) → Claude can design architecture
4. **Slice 4**: US-6 Queries (Phase 5) → Token efficiency verified
5. **Slice 5**: US-4 API + US-5 Export (Phases 6-7) → CI/CD and multi-format
6. **Polish**: Phase 8 → Docs, examples, error handling

### Parallel Team Strategy

With 3 developers, after Phase 1:

- **Dev A**: Phase 2 (US-3 Scaffolding)
- **Dev B**: Phase 3 (US-2 Watch Mode)
- **Dev C**: Phase 5 (US-6 Queries - simpler, parallelizable)

Once Phases 2-3 complete:

- **Dev A**: Phase 4 (US-1 MCP - uses results from A and B)
- **Dev B**: Phase 6 (US-4 API - uses results from B)
- **Dev C**: Phase 7 (US-5 Export - uses results from B)

---

## Notes

- **[P] tasks** = Different files, no dependencies between them
- **[Story] label** = Maps to user story for traceability
- Each user story should be independently completable and testable
- **Test files**: Write tests FIRST, ensure they FAIL before implementation
- **Commit after each task** or logical group
- **Stop at any checkpoint** to validate story independently
- Use `make test`, `make lint`, `make coverage` before commits
- No third-party mocking libraries - use concrete mocks (see MockProjectRepo pattern)
