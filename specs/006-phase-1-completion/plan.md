# Implementation Plan: Production-Ready Phase 1 Release

**Branch**: `006-phase-1-completion` | **Date**: 2026-02-13 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/006-phase-1-completion/spec.md`

## Summary

Complete loko v0.2.0 Phase 1 to production-ready quality by:
1. **TOON v3.0 compliance** - Align output with official spec using `github.com/toon-format/toon-go` library
2. **Handler refactoring** - Extract business logic from CLI/MCP handlers into use cases (CLI < 50 lines, MCP < 30 lines)
3. **Search & filter MCP tools** - Add `search_elements` and `find_relationships` tools leveraging existing graph infrastructure
4. **CI/CD integration** - Provide GitHub Actions, GitLab CI, and Docker Compose examples with `--strict`/`--exit-code` flags
5. **OpenAPI serving** - Embed Swagger UI and serve spec at `/api/docs` using `go:embed`
6. **Documentation polish** - Update README, create MCP integration guide, verify all 4 examples, create demo GIF
7. **PDF graceful degradation** - Handle veve-cli absence with helpful error messages

**Technical Approach**: Leverage existing clean architecture and graph infrastructure. No new entities required - this release polishes and extends existing capabilities.

---

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: 
- `github.com/toon-format/toon-go` (TOON v3.0 encoding)
- `golang.org/x/time/rate` (rate limiting - optional)
- Swagger UI static assets (embedded via `go:embed`)
- Existing: Cobra, Viper, Lipgloss, d2 CLI, veve-cli (optional)

**Storage**: File system (markdown, TOML, D2, YAML frontmatter) - no changes  
**Testing**: Go test framework (`go test ./...`) with table-driven tests, > 80% coverage in `internal/core/`  
**Target Platform**: Linux, macOS, Docker (single binary distribution)  
**Project Type**: Single project (CLI tool with embedded MCP/API servers)  
**Performance Goals**:
- Search tools: < 200ms response time
- Watch mode rebuild: < 500ms
- MCP tool responses: < 100ms (excluding diagram rendering)
- Build time (10-system project): < 5 seconds
- API server startup: < 100ms increase from v0.1.0

**Constraints**:
- Clean Architecture: core/ has zero external dependencies
- Thin Handlers: CLI < 50 lines, MCP < 30 lines, API < 50 lines
- Backward compatibility: Existing MCP clients must continue working
- Binary size: Swagger UI embedding < 5MB increase
- Test coverage: Maintain > 80% in `internal/core/`

**Scale/Scope**: 
- Support projects with 100+ components
- 17 MCP tools (15 existing + 2 new)
- 4 example projects (simple, 3layer, serverless, microservices)
- CI/CD examples for 2 platforms (GitHub Actions, GitLab CI)

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Gate 1: Clean Architecture (NON-NEGOTIABLE)

- [x] **Core has zero external dependencies**: New TOON library usage is in `internal/adapters/encoding/`, not `core/`
- [x] **Dependency direction enforced**: Search tools follow existing pattern (CLI/MCP → use cases → entities)
- [x] **Port interfaces in usecases/ports.go**: No new external dependencies in core; existing ports sufficient
- [x] **Use cases contain business logic**: Handler refactoring moves logic from `cmd/` to `usecases/`

**Status**: PASS - Architecture remains clean; no violations introduced

### ✅ Gate 2: Interface-First

- [x] **Ports defined before adapters**: TOON encoding uses existing `OutputEncoder` port
- [x] **No concrete adapter refs in use cases**: Search uses `ArchitectureGraph` interface
- [x] **Wiring in main.go**: Swagger UI embedded assets served via handler, not core logic

**Status**: PASS - All external dependencies behind interfaces

### ✅ Gate 3: Thin Handlers

- [x] **CLI < 50 lines**: **PRIMARY GOAL** - Refactor `cmd/new.go` (504 lines), `cmd/build.go` (251 lines), etc.
- [x] **MCP < 30 lines**: **PRIMARY GOAL** - Split `tools.go` (1,084 lines), refactor `graph_tools.go` (348 lines)
- [x] **Handlers parse/call/format only**: New search tools will follow thin handler pattern

**Status**: **CURRENTLY VIOLATED** - This release fixes violations (constitution audit in CI)

### ✅ Gate 4: Entity Validation

- [x] **Validation in entity constructors**: No new entities; existing validation unchanged
- [x] **Use cases trust valid entities**: Search tools use existing validated entities
- [x] **No validation in handlers**: Refactoring removes validation from handlers

**Status**: PASS - Existing validation patterns maintained

### ✅ Gate 5: Test-First

- [x] **Unit tests for use cases**: New search use cases will have 100% coverage
- [x] **Integration tests for adapters**: TOON encoder tested against official parser
- [x] **E2E tests for handlers**: CI examples tested in real pipelines
- [x] **Target > 80% coverage**: Must maintain existing coverage after refactoring

**Status**: PASS - Test-first approach enforced; existing tests protect against regressions

### ✅ Gate 6: Token Efficiency

- [x] **Progressive context loading**: Existing pattern maintained
- [x] **TOON format support**: **PRIMARY GOAL** - Align with v3.0 spec, verify 30-40% reduction
- [x] **JSON default, TOON opt-in**: Existing behavior preserved

**Status**: ENHANCEMENT - TOON compliance improves token efficiency

### ✅ Gate 7: Simplicity & YAGNI

- [x] **Simplest solution**: Search tools leverage existing graph infrastructure
- [x] **No premature abstraction**: Swagger UI embedded directly, no CDN/external service
- [x] **Concrete mocks, no libraries**: Existing mock pattern maintained
- [x] **Single binary**: No additional runtime dependencies

**Status**: PASS - No unnecessary complexity introduced

**Overall Gate Result**: ✅ **PASS** (with violations to be fixed)

**Justification for Current Violations**:
- **Thin Handler violations (Gate 3)**: These are *intentional fixes* - the entire purpose of this release is to resolve these technical debt items before Phase 2

---

## Project Structure

### Documentation (this feature)

```text
specs/006-phase-1-completion/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - TOON v3.0 research, handler refactoring patterns
├── data-model.md        # Phase 1 output - Configuration schema changes
├── quickstart.md        # Phase 1 output - Developer quickstart for contributing
├── contracts/           # Phase 1 output - MCP tool schemas, CI workflow schemas
│   ├── mcp-tools.md     # search_elements, find_relationships schemas
│   ├── ci-examples.md   # GitHub Actions, GitLab CI contract validation
│   └── api-openapi.md   # OpenAPI spec validation contract
└── checklists/
    └── requirements.md  # Already created - validation passed
```

### Source Code (repository root)

```text
# loko follows Single Project structure (CLI tool with embedded servers)

internal/
├── core/                           # ZERO external dependencies (unchanged)
│   ├── entities/                   # Domain objects (no changes)
│   │   ├── project.go
│   │   ├── system.go
│   │   ├── container.go
│   │   ├── component.go
│   │   └── graph.go
│   └── usecases/                   # Application logic + ports
│       ├── ports.go                # Interface definitions (no new ports needed)
│       ├── create_system.go
│       ├── query_architecture.go   # TOON formatting moved here
│       ├── search_elements.go      # NEW - search use case
│       ├── find_relationships.go   # NEW - relationship filtering
│       ├── validate_architecture.go
│       └── build_docs.go
│
├── adapters/                       # Infrastructure implementations
│   ├── filesystem/                 # ProjectRepository
│   ├── d2/                         # DiagramRenderer
│   ├── encoding/                   # OutputEncoder (JSON, TOON)
│   │   ├── json.go
│   │   ├── toon.go                 # REFACTOR - align with TOON v3.0
│   │   └── toon_test.go            # Validation against official parser
│   ├── html/                       # SiteBuilder
│   ├── pdf/                        # PDFRenderer
│   │   └── renderer.go             # ENHANCE - graceful degradation
│   ├── ason/                       # TemplateEngine
│   └── config/                     # ConfigLoader
│       └── loader.go               # EXTEND - API config (rate_limit, CORS)
│
├── mcp/                            # MCP server (thin layer)
│   ├── server.go
│   └── tools/                      # MCP tool handlers
│       ├── query_project.go        # Existing (unchanged)
│       ├── query_architecture.go   # Existing (unchanged)
│       ├── create_system.go        # Existing (unchanged)
│       ├── graph_tools.go          # REFACTOR - split into separate files
│       ├── search_elements.go      # NEW - < 30 lines
│       ├── find_relationships.go   # NEW - < 30 lines
│       └── ...                     # Other existing tools
│
├── api/                            # HTTP API (thin layer)
│   ├── server.go                   # ENHANCE - serve OpenAPI + Swagger UI
│   ├── handlers/
│   │   └── handlers.go
│   ├── middleware/
│   │   └── middleware.go           # OPTIONAL - rate limiting, CORS
│   ├── openapi.yaml                # Existing
│   └── static/                     # NEW - embedded Swagger UI assets
│       └── swagger-ui/             # go:embed target
│
└── ui/                             # Lipgloss styles (unchanged)

cmd/                                # CLI commands (thin layer)
├── root.go
├── init.go
├── new.go                          # REFACTOR - reduce from 504 to < 50 lines
├── build.go                        # REFACTOR - reduce from 251 to < 50 lines
├── validate.go                     # ENHANCE - add --strict, --exit-code flags
├── serve.go
├── watch.go
├── mcp.go
├── api.go
└── completion.go

examples/                           # Example projects
├── simple-project/                 # VERIFY - builds successfully
├── 3layer-app/                     # VERIFY - builds successfully
├── serverless/                     # VERIFY - builds successfully
├── microservices/                  # VERIFY - builds successfully
└── ci/                             # NEW - CI/CD examples
    ├── github-actions.yml          # GitHub Actions workflow
    ├── gitlab-ci.yml               # GitLab CI pipeline
    └── docker-compose.dev.yml      # Docker Compose dev environment

docs/                               # Documentation
├── guides/
│   ├── mcp-integration.md          # NEW - Step-by-step MCP setup
│   └── ci-cd-integration.md        # NEW - CI/CD examples guide
├── adr/                            # Architecture Decision Records (meta)
│   ├── 0001-clean-architecture.md
│   ├── 0002-token-efficient-mcp.md
│   ├── 0003-toon-format.md
│   └── 0004-graph-conventions.md
└── README.md

tests/
├── integration/                    # Integration tests
│   ├── mcp_search_test.go          # NEW - test search tools
│   ├── ci_validation_test.go       # NEW - test CI examples
│   └── ...
└── benchmarks/                     # Performance benchmarks
    ├── token_efficiency_test.go    # ENHANCE - JSON vs TOON benchmarks
    └── ...

.github/
└── workflows/
    └── loko-validate.yml           # NEW - CI example (copied from examples/ci/)

.specify/
├── scripts/
│   └── bash/
│       └── constitution-audit.sh   # NEW - line counting audit script
└── memory/
    └── constitution.md             # Existing
```

**Structure Decision**: Single project structure is maintained. loko is a CLI tool that embeds MCP and API servers, following existing conventions. No new top-level directories needed - enhancements fit within existing structure.

---

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| **Current thin handler violations** (Gate 3) | This release exists to *fix* these violations via refactoring | These are not new violations - they are existing technical debt being paid down |
| `cmd/new.go` (504 lines) | Business logic embedded in handler during rapid prototyping | Cannot be simpler - this release extracts logic to use cases |
| `cmd/build.go` (251 lines) | Adapter instantiation and orchestration in handler | Cannot be simpler - this release moves orchestration to use cases |
| `internal/mcp/tools/tools.go` (1,084 lines) | 10+ tools in single file | Simpler to split now rather than grow further - this release splits into individual files |

**Justification**: These violations are the *target of this release* - they will be eliminated, not added. No new complexity is being introduced that requires justification.

---

## Phase 0: Research & Unknowns

**Status**: Research needed for:
1. TOON v3.0 tabular array format specifics
2. Handler refactoring patterns (extracting 500+ lines to < 50)
3. Swagger UI embedding best practices (binary size optimization)
4. Constitution audit script implementation

**Output**: `research.md` (to be generated)

---

## Phase 1: Design & Contracts

**Status**: To be generated after research complete

**Outputs**:
- `data-model.md` - Configuration schema changes (API settings, CI flags)
- `contracts/` - MCP tool schemas, CI workflow contracts, OpenAPI validation
- `quickstart.md` - Developer contribution guide

**Agent Context Update**: Will run `.specify/scripts/bash/update-agent-context.sh opencode` after design complete

---

## Phase 2: Task Breakdown

**Status**: Deferred to `/speckit.tasks` command

**Output**: `tasks.md` (not created by this command)

---

## Implementation Timeline

**Week 1** (Complete Existing Work):
- Days 1-2: TOON v3.0 compliance + benchmarking
- Days 3-4: Handler refactoring (CLI + MCP)
- Day 5: PDF graceful degradation + initial docs polish

**Week 2** (High-Value Quick Wins):
- Days 1-2: Search & filter MCP tools
- Day 3: CI/CD examples + `--strict`/`--exit-code` flags
- Day 4: OpenAPI serving + Swagger UI
- Day 5: Rate limiting/CORS (optional) + buffer

**Week 3** (Polish & Release):
- Days 1-2: Documentation polish
- Day 3: Demo GIF + example verification
- Day 4: Integration testing + benchmarking
- Day 5: Release preparation

**Total**: 11-18 days (2-3 weeks)
