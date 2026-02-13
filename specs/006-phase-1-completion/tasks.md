# Tasks: Production-Ready Phase 1 Release

**Input**: Design documents from `/specs/006-phase-1-completion/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/  
**Target**: loko v0.2.0  
**Timeline**: 11-18 days (2-3 weeks)

---

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5, US6, US7)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependency updates, and tooling setup

- [X] T001 [P] Add `github.com/toon-format/toon-go` dependency to `go.mod`
- [X] T002 [P] Add Swagger UI minimal build assets to `internal/api/static/swagger-ui/` via `go:embed`
- [X] T003 [P] Update `.gitignore` to exclude generated OpenAPI specs (if not committed)
- [X] T004 Create `examples/ci/` directory structure for GitHub Actions, GitLab CI, Docker Compose examples
- [X] T005 Create `docs/guides/` directory for MCP integration guide

**Checkpoint**: ✅ Dependencies installed, directories created

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 [P] Create constitution audit script `scripts/audit-constitution.sh` (grep-based line counting, excludes imports/comments/blank lines)
- [X] T007 [P] Create TOON encoder adapter in `internal/adapters/encoding/toon_encoder.go` implementing `OutputEncoder` port
- [X] T008 [P] Add TOON encoder tests in `internal/adapters/encoding/toon_encoder_test.go` (validate against official TOON parser)
- [X] T009 Create baseline constitution audit results (document current violations: cmd/new.go 504 lines, cmd/build.go 251 lines, tools.go 1,084 lines)
- [X] T010 Add constitution audit to `.github/workflows/ci.yml` (run on PR, fail on new violations)
- [X] T011 Create token efficiency benchmarking script in `scripts/benchmark-token-efficiency.sh` (compare JSON vs TOON output)

**Checkpoint**: ✅ Foundation ready - audit script working, TOON encoder validated, CI infrastructure in place

---

## Phase 3: User Story 1 - Search & Filter MCP Tools (Priority: P1) 🎯 MVP

**Goal**: LLM agents can search architecture elements by name, technology, tags without loading full graph

**Independent Test**: MCP client calls `search_elements` with query "payment" → receives filtered results < 200ms

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T012 [P] [US1] Contract test for `search_elements` tool in `tests/mcp/test_search_elements.go` (29 test scenarios from contracts/mcp-tools.md)
- [ ] T013 [P] [US1] Contract test for `find_relationships` tool in `tests/mcp/test_find_relationships.go` (test scenarios from contracts/mcp-tools.md)
- [ ] T014 [P] [US1] Integration test for search performance in `tests/integration/test_search_performance.go` (verify < 200ms response time)

### Implementation for User Story 1

- [ ] T015 [P] [US1] Create `SearchElementsRequest` and `SearchElementsResponse` structs in `internal/core/entities/search.go`
- [ ] T016 [P] [US1] Create `FindRelationshipsRequest` and `FindRelationshipsResponse` structs in `internal/core/entities/search.go`
- [ ] T017 [US1] Implement `SearchElements` use case in `internal/core/usecases/search_elements.go` (uses existing ArchitectureGraph port)
- [ ] T018 [US1] Implement `FindRelationships` use case in `internal/core/usecases/find_relationships.go` (uses existing ArchitectureGraph port)
- [ ] T019 [US1] Add glob pattern matching helper in `internal/core/entities/glob_matcher.go` (support wildcards: *, ?)
- [ ] T020 [US1] Add result limiting logic (default: 20, max: 100) to prevent token overflow
- [ ] T021 [US1] Create thin MCP tool handler for `search_elements` in `internal/mcp/tools/search_elements.go` (< 30 lines: parse → call use case → format)
- [ ] T022 [US1] Create thin MCP tool handler for `find_relationships` in `internal/mcp/tools/find_relationships.go` (< 30 lines)
- [ ] T023 [US1] Register new MCP tools in `internal/mcp/server.go` (add to tool registry)
- [ ] T024 [US1] Add empty result set handling with helpful messages ("No elements found matching 'X'")
- [ ] T025 [US1] Update MCP tool count in `README.md` (15 → 17 tools)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently (MCP clients can search architecture)

---

## Phase 4: User Story 2 - CI/CD Integration (Priority: P1) 🎯 MVP

**Goal**: DevOps engineers can validate architecture in CI/CD pipelines with standard YAML config

**Independent Test**: Copy GitHub Actions example to `.github/workflows/` → push PR with invalid architecture → workflow fails with exit code 1

### Tests for User Story 2 ⚠️

- [ ] T026 [P] [US2] Contract test for GitHub Actions workflow in `tests/ci/test_github_actions.go` (15 test scenarios from contracts/ci-examples.md)
- [ ] T027 [P] [US2] Contract test for GitLab CI pipeline in `tests/ci/test_gitlab_ci.go` (test scenarios from contracts/ci-examples.md)
- [ ] T028 [P] [US2] Contract test for Docker Compose watch mode in `tests/ci/test_docker_compose.go` (verify < 500ms rebuild)

### Implementation for User Story 2

- [ ] T029 [P] [US2] Add `--strict` flag to `cmd/validate.go` (treats warnings as errors)
- [ ] T030 [P] [US2] Add `--exit-code` flag to `cmd/validate.go` (returns non-zero on errors)
- [ ] T031 [P] [US2] Create GitHub Actions workflow example in `examples/ci/github-actions.yml` (uses loko validate --strict --exit-code)
- [ ] T032 [P] [US2] Create GitLab CI pipeline example in `examples/ci/.gitlab-ci.yml` (uploads artifacts on success)
- [ ] T033 [P] [US2] Create Docker Compose dev environment in `examples/ci/docker-compose.yml` (watch mode, volume mounts)
- [ ] T034 [P] [US2] Create Dockerfile for loko with veve-cli pre-installed in `examples/ci/Dockerfile`
- [ ] T035 [US2] Test GitHub Actions workflow in real repository (create test PR, verify failure on invalid architecture)
- [ ] T036 [US2] Test GitLab CI pipeline in real repository (verify artifact upload)
- [ ] T037 [US2] Create CI/CD integration guide in `docs/guides/ci-cd-integration.md` (setup instructions for each platform)
- [ ] T038 [US2] Update `cmd/validate.go` help text to mention `--strict` and `--exit-code` flags

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently (search works, CI integration works)

---

## Phase 5: User Story 3 - TOON v3.0 Compliance (Priority: P1) 🎯 MVP

**Goal**: Architects can share architecture with LLMs in spec-compliant TOON format with 30-40% token savings

**Independent Test**: Export architecture as TOON → official TOON parser validates → token count 30-40% less than JSON

### Tests for User Story 3 ⚠️

- [ ] T039 [P] [US3] Token efficiency benchmark test in `tests/benchmarks/test_token_efficiency.go` (compare JSON vs TOON for 4 example projects)
- [ ] T040 [P] [US3] TOON v3.0 spec compliance test in `tests/integration/test_toon_compliance.go` (validate against official parser)
- [ ] T041 [P] [US3] TOON backward compatibility test in `tests/integration/test_toon_backward_compat.go` (verify existing MCP clients still work)

### Implementation for User Story 3

- [ ] T042 [US3] Wire TOON encoder in `cmd/build.go` (add `--format toon` flag)
- [ ] T043 [US3] Wire TOON encoder in `internal/mcp/tools/` for MCP tool responses (conditional based on client capabilities)
- [ ] T044 [US3] Run token efficiency benchmarks on 4 example projects (simple, 3layer, serverless, microservices)
- [ ] T045 [US3] Document benchmark results in `research/token-efficiency-benchmarks.md` (include tables with token counts)
- [ ] T046 [US3] Update README with spec-compliant TOON examples (replace custom notation)
- [ ] T047 [US3] Add TOON format documentation to `docs/guides/toon-format-guide.md` (syntax, use cases, token savings)
- [ ] T048 [US3] Verify all 4 example projects build successfully with TOON format
- [ ] T049 [US3] Create migration guide for existing users (if format breaking changes)

**Checkpoint**: All MVP user stories (US1, US2, US3) should now be independently functional

---

## Phase 6: User Story 4 - OpenAPI Serving (Priority: P2)

**Goal**: Developers can discover and test loko HTTP API using Swagger UI at `/api/docs`

**Independent Test**: Start `loko api` → navigate to `/api/docs` → use Swagger UI to test GET `/api/v1/systems` → receive valid JSON response

### Tests for User Story 4 ⚠️

- [ ] T050 [P] [US4] Contract test for OpenAPI spec accuracy in `tests/api/test_openapi_spec.go` (validate spec against actual handlers)
- [ ] T051 [P] [US4] Integration test for Swagger UI serving in `tests/api/test_swagger_ui.go` (verify /api/docs returns 200)
- [ ] T052 [P] [US4] API server startup performance test in `tests/benchmarks/test_api_startup.go` (verify < 100ms increase)

### Implementation for User Story 4

- [ ] T053 [P] [US4] Download and compress Swagger UI minimal build to `internal/api/static/swagger-ui/` (aim for < 400KB gzipped)
- [ ] T054 [P] [US4] Create `go:embed` directive in `internal/api/static/swagger.go` for embedding Swagger UI assets
- [ ] T055 [US4] Create OpenAPI 3.0 spec generator in `internal/api/openapi/generator.go` (auto-generate from handlers)
- [ ] T056 [US4] Add `/api/v1/openapi.json` endpoint in `internal/api/handlers/openapi.go` (serve JSON spec)
- [ ] T057 [US4] Add `/api/v1/openapi.yaml` endpoint in `internal/api/handlers/openapi.go` (serve YAML spec)
- [ ] T058 [US4] Add `/api/docs` endpoint in `internal/api/handlers/swagger_ui.go` (serve embedded Swagger UI with spec URL)
- [ ] T059 [US4] Document Bearer token authentication in OpenAPI spec (security schemes)
- [ ] T060 [US4] Verify Swagger UI works offline (no CDN dependencies)
- [ ] T061 [US4] Run `openapi-generator validate` on generated spec (ensure spec validity)
- [ ] T062 [US4] Update `cmd/api.go` help text to mention `/api/docs` and `/api/v1/openapi.json` endpoints
- [ ] T063 [US4] Verify binary size increase < 5MB after Swagger UI embedding

**Checkpoint**: User Story 4 should be independently functional (API docs accessible, Swagger UI works)

---

## Phase 7: User Story 5 - Handler Refactoring (Priority: P2)

**Goal**: Contributors can understand codebase quickly; all handlers follow thin-handler principle (CLI < 50 lines, MCP < 30 lines)

**Independent Test**: Run constitution audit script → all CLI handlers < 50 lines → all MCP tools < 30 lines → business logic in use cases only

### Implementation for User Story 5

> **NOTE**: Each refactoring task should maintain test coverage and avoid breaking changes

#### Refactor cmd/new.go (504 lines → < 50 lines)

- [ ] T064 [US5] Create `CreateProjectRequest` and `CreateProjectResponse` structs in `internal/core/entities/project.go`
- [ ] T065 [US5] Extract business logic from `cmd/new.go` to `internal/core/usecases/create_project.go` (< 4 hours)
- [ ] T066 [US5] Refactor `cmd/new.go` to thin handler (parse flags → call CreateProject use case → format output) (< 50 lines)
- [ ] T067 [US5] Run tests to verify `cmd/new.go` refactoring (ensure no regressions)

#### Refactor cmd/build.go (251 lines → < 50 lines)

- [ ] T068 [US5] Create `BuildArchitectureRequest` and `BuildArchitectureResponse` structs in `internal/core/entities/build.go`
- [ ] T069 [US5] Extract business logic from `cmd/build.go` to `internal/core/usecases/build_architecture.go` (< 4 hours)
- [ ] T070 [US5] Refactor `cmd/build.go` to thin handler (parse flags → call BuildArchitecture use case → format output) (< 50 lines)
- [ ] T071 [US5] Run tests to verify `cmd/build.go` refactoring (ensure no regressions)

#### Refactor internal/mcp/tools/tools.go (1,084 lines → split into separate files)

- [ ] T072 [US5] Split `tools.go` into separate files per tool in `internal/mcp/tools/` (one file per tool, < 30 lines each)
- [ ] T073 [US5] Extract business logic from MCP tool handlers to use cases in `internal/core/usecases/` (< 4 hours)
- [ ] T074 [US5] Refactor each MCP tool handler to thin handler pattern (parse params → call use case → format response) (< 30 lines)
- [ ] T075 [US5] Run tests to verify MCP tools refactoring (ensure no regressions)

#### Refactor internal/mcp/tools/graph_tools.go (348 lines → < 30 lines per tool)

- [ ] T076 [US5] Extract graph operations from `graph_tools.go` to use cases in `internal/core/usecases/graph_operations.go` (< 4 hours)
- [ ] T077 [US5] Refactor graph tool handlers to thin handler pattern (< 30 lines each)
- [ ] T078 [US5] Run tests to verify graph tools refactoring (ensure no regressions)

#### Constitution Audit Integration

- [ ] T079 [US5] Run constitution audit script on refactored handlers (verify all pass)
- [ ] T080 [US5] Update constitution audit CI workflow to fail on violations
- [ ] T081 [US5] Document exceptions (if any legitimate handler needs > 50 lines)
- [ ] T082 [US5] Verify test coverage remains > 80% in `internal/core/` after refactoring
- [ ] T083 [US5] Create refactoring guide in `docs/guides/handler-refactoring-guide.md` (Extract Use Case pattern)

**Checkpoint**: User Story 5 should be complete (all handlers follow constitution, audit passes in CI)

---

## Phase 8: User Story 6 - PDF Graceful Degradation (Priority: P3)

**Goal**: Users without veve-cli can build HTML/Markdown immediately without errors

**Independent Test**: Fresh install without veve-cli → `loko build` → HTML/Markdown succeed → clear message about optional PDF with install link

### Tests for User Story 6 ⚠️

- [ ] T084 [P] [US6] Integration test for build without veve-cli in `tests/integration/test_build_no_veve.go` (verify HTML/Markdown succeed)
- [ ] T085 [P] [US6] Integration test for explicit PDF build without veve-cli in `tests/integration/test_pdf_no_veve.go` (verify helpful error message)

### Implementation for User Story 6

- [ ] T086 [US6] Add veve-cli detection logic in `internal/adapters/diagram/veve_adapter.go` (check PATH for veve-cli binary)
- [ ] T087 [US6] Add `--skip-pdf` flag to `cmd/build.go` (suppress PDF warnings)
- [ ] T088 [US6] Implement graceful degradation: if veve-cli absent, skip PDF, show warning with installation link
- [ ] T089 [US6] Implement error handling: if `--format pdf` and veve-cli absent, show helpful error with installation instructions
- [ ] T090 [US6] Style error/warning messages using lipgloss (follow loko UI consistency)
- [ ] T091 [US6] Add veve-cli installation to Docker image in `examples/ci/Dockerfile`
- [ ] T092 [US6] Update `README.md` to document optional veve-cli dependency for PDF export
- [ ] T093 [US6] Test both scenarios: (1) veve-cli present, (2) veve-cli absent

**Checkpoint**: User Story 6 should be complete (graceful degradation works, helpful error messages)

---

## Phase 9: User Story 7 - MCP Integration Guide (Priority: P2)

**Goal**: New users can complete MCP setup with Claude Desktop without trial and error

**Independent Test**: New user follows MCP guide → configures Claude Desktop → chats "create a payment system" → loko scaffolds architecture → user views in browser

### Implementation for User Story 7

- [ ] T094 [P] [US7] Create MCP integration guide in `docs/guides/mcp-integration-guide.md` (Claude Desktop setup instructions)
- [ ] T095 [P] [US7] Add Claude Desktop configuration JSON example to MCP guide (copy-paste ready)
- [ ] T096 [P] [US7] Create example conversation flow in MCP guide (init → scaffold → build → serve)
- [ ] T097 [P] [US7] Create troubleshooting section in MCP guide (common issues: permission denied, port conflicts, invalid config)
- [ ] T098 [P] [US7] Create MCP tool reference table in MCP guide (all 17 tools with descriptions and example usage)
- [ ] T099 [US7] Test MCP guide with fresh Claude Desktop installation (verify configuration works first try)
- [ ] T100 [US7] Add screenshot/GIF of MCP tool usage in Claude Desktop to MCP guide
- [ ] T101 [US7] Link MCP guide from README.md (prominent placement in Getting Started section)

**Checkpoint**: User Story 7 should be complete (MCP guide comprehensive, tested with fresh install)

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T102 [P] Update README roadmap to reflect actual v0.1.0 shipped state (what's done vs. planned)
- [ ] T103 [P] Update README with accurate MCP tool count (17: 15 existing + 2 new search tools)
- [ ] T104 [P] Add verified token efficiency claims to README (link to benchmarks in research/)
- [ ] T105 [P] Create 2-3 minute demo GIF showing full workflow (init → scaffold → build → serve → MCP integration) (< 5MB)
- [ ] T106 [P] Verify all 4 example projects build successfully (simple-project, 3layer-app, serverless, microservices)
- [ ] T107 [P] Create README for each example explaining what it demonstrates
- [ ] T108 [P] Run quickstart.md validation (verify all developer setup steps work)
- [ ] T109 Code cleanup: Remove dead code, fix typos, improve comments
- [ ] T110 Performance optimization: Profile search tools, optimize hot paths
- [ ] T111 Security hardening: Review error messages for information leakage, validate all inputs
- [ ] T112 [P] Run full test suite: `go test ./...` (verify > 80% coverage in internal/core/)
- [ ] T113 [P] Run linting: `task lint` (verify golangci-lint passes)
- [ ] T114 [P] Run formatting: `task fmt` (verify gofmt + goimports passes)
- [ ] T115 Final constitution audit: Run audit script, verify zero violations
- [ ] T116 Create release notes for v0.2.0 in `CHANGELOG.md` (summarize all user stories, breaking changes, migration guide)

**Checkpoint**: All polish complete, ready for release

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-9)**: All depend on Foundational phase completion
  - US1 (Search Tools) - Can start immediately after Foundational
  - US2 (CI/CD) - Can start immediately after Foundational
  - US3 (TOON) - Depends on TOON encoder from Foundational
  - US4 (OpenAPI) - Can start immediately after Foundational
  - US5 (Handler Refactoring) - Can start after US1/US2/US3 complete (to avoid merge conflicts)
  - US6 (PDF Degradation) - Can start immediately after Foundational
  - US7 (MCP Guide) - Should wait for US1 complete (to document new search tools)
- **Polish (Phase 10)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P1)**: Depends on TOON encoder from Foundational (Phase 2)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 5 (P2)**: Should wait for US1/US2/US3 complete to avoid merge conflicts in handlers
- **User Story 6 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 7 (P2)**: Should wait for US1 complete to document new search tools

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Use cases before handlers
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1: All tasks marked [P] can run in parallel (T001-T005)
- Phase 2: Tasks T006-T008 can run in parallel (audit script, TOON encoder, tests)
- Phase 3 (US1): Tests T012-T014 in parallel, entities T015-T016 in parallel, use cases T017-T018 in parallel, handlers T021-T022 in parallel
- Phase 4 (US2): Tests T026-T028 in parallel, flags T029-T030 in parallel, examples T031-T034 in parallel
- Phase 5 (US3): Tests T039-T041 in parallel
- Phase 6 (US4): Tests T050-T052 in parallel, Swagger assets T053-T054 in parallel, endpoints T056-T058 can be done in parallel
- Phase 7 (US5): Refactoring of different handlers can be done in parallel by different developers
- Phase 8 (US6): Tests T084-T085 in parallel
- Phase 9 (US7): Tasks T094-T098 can run in parallel (different documentation files)
- Phase 10: Most tasks marked [P] can run in parallel (T102-T108, T112-T114)

---

## Parallel Example: MVP (Phase 1-5)

```bash
# After Foundational (Phase 2) completes, launch MVP user stories in parallel:

# Developer A: User Story 1 (Search Tools)
Task: "Contract test for search_elements tool"
Task: "Implement SearchElements use case"

# Developer B: User Story 2 (CI/CD Integration)
Task: "Add --strict flag to cmd/validate.go"
Task: "Create GitHub Actions workflow example"

# Developer C: User Story 3 (TOON Compliance)
Task: "Wire TOON encoder in cmd/build.go"
Task: "Run token efficiency benchmarks"
```

---

## Implementation Strategy

### MVP First (User Stories 1, 2, 3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Search Tools)
4. Complete Phase 4: User Story 2 (CI/CD Integration)
5. Complete Phase 5: User Story 3 (TOON Compliance)
6. **STOP and VALIDATE**: Test all MVP stories independently
7. Deploy/demo if ready

This delivers the core value: LLM agents can search architecture (US1), DevOps can validate in CI (US2), and token efficiency is validated (US3).

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (2 days)
2. Add User Story 1 → Test independently → Deploy/Demo (2-3 days)
3. Add User Story 2 → Test independently → Deploy/Demo (2-3 days)
4. Add User Story 3 → Test independently → Deploy/Demo (2-3 days)
5. Add User Story 4 → Test independently → Deploy/Demo (1-2 days)
6. Add User Story 5 → Test independently → Deploy/Demo (3-4 days - largest effort)
7. Add User Story 6 → Test independently → Deploy/Demo (1 day)
8. Add User Story 7 → Test independently → Deploy/Demo (1 day)
9. Polish → Release v0.2.0 (1-2 days)

**Total**: 11-18 days (2-3 weeks)

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (2 days)
2. Once Foundational is done:
   - Developer A: User Story 1 (Search Tools) (2-3 days)
   - Developer B: User Story 2 (CI/CD) (2-3 days)
   - Developer C: User Story 3 (TOON) (2-3 days)
3. After MVP complete:
   - Developer A: User Story 4 (OpenAPI) (1-2 days)
   - Developer B: User Story 6 (PDF Degradation) (1 day)
   - Developer C: User Story 7 (MCP Guide) (1 day)
4. All developers: User Story 5 (Handler Refactoring) together (3-4 days - large effort)
5. All developers: Polish (1-2 days)

**Total with 3 developers**: 7-10 days (1-2 weeks)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution audit must pass in CI before merging any PR
- Maintain > 80% test coverage in internal/core/ throughout
