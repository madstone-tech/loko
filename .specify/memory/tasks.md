# loko Tasks

> Generated: 2024-12-17
> Spec Version: 0.1.0-dev
> Status: In Progress

## Progress Summary

| Phase | Status | Issues |
|-------|--------|--------|
| Phase 1: Foundation | ✅ 2/3 Complete | #1 ✅, #2 ✅, #3 🔲 |
| Phase 2: First Use Case | 🔲 Not Started | #4-#6 |
| Phase 3: Build Pipeline | 🔲 Not Started | #7-#10 |
| Phase 4: MCP | 🔲 Not Started | #11-#13 |
| Phase 5: v0.2.0 | 🔲 Not Started | #14-#15 |

## Completed Tasks

### ✅ Issue #1: Initialize project with Clean Architecture structure

**Completed:** 2024-12-17

- [x] Initialize Go module (`github.com/madstone-tech/loko`)
- [x] Create directory structure (cmd/, internal/core/, internal/adapters/, etc.)
- [x] Set up GitHub Actions for test/lint/build
- [x] Add Taskfile with common commands
- [x] Configure golangci-lint
- [x] Add goreleaser config
- [x] Create ADR directory with template

### ✅ Issue #2: Implement core domain entities

**Completed:** 2024-12-17

- [x] `project.go` - Project entity with systems collection
- [x] `system.go` - System entity with containers collection
- [x] `container.go` - Container entity with components
- [x] `component.go` - Component entity
- [x] `diagram.go` - Diagram entity (source + rendered)
- [x] `template.go` - Template entity
- [x] `validation.go` - Validation logic for all entities
- [x] `errors.go` - Domain error types

---

## Phase 1: Foundation (Remaining)

### 🔲 Issue #3: Define use case ports (interfaces)

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-019, FR-020, FR-021

Define all interfaces in `internal/core/usecases/ports.go` that adapters must implement.

**Tasks:**

- [ ] `ProjectRepository` - Load/Save project
- [ ] `TemplateRepository` - List/Get templates  
- [ ] `DiagramRenderer` - Render D2 diagrams
- [ ] `TemplateEngine` - Render scaffolding templates (ason)
- [ ] `SiteBuilder` - Generate HTML site
- [ ] `PDFRenderer` - Generate PDFs (optional dep)
- [ ] `FileWatcher` - Watch for file changes
- [ ] `OutputEncoder` - Encode responses (JSON/TOON)
- [ ] `Logger` - Structured logging
- [ ] `ProgressReporter` - Feedback during operations
- [ ] Define input/output types for complex operations

**Acceptance Criteria:**

- All interfaces documented with godoc
- Input/output types defined
- No implementation details leak into interfaces
- Interfaces are minimal (Interface Segregation Principle)

---

## Phase 2: First Use Case End-to-End

### 🔲 Issue #4: Implement CreateSystem use case

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** US-3, FR-020
**Depends On:** #3

First complete use case: creating a new C4 system.

**Tasks:**

- [ ] `internal/core/usecases/create_system.go`
- [ ] Define `CreateSystemInput` and `CreateSystemOutput`
- [ ] Input validation
- [ ] Duplicate detection
- [ ] Template loading and rendering
- [ ] Project saving
- [ ] Unit tests with mocked ports

**Acceptance Criteria:**

- Use case works with mock repositories
- Returns structured output (System + files created)
- Proper error handling with domain errors
- > 90% test coverage

---

### 🔲 Issue #5: Implement file system adapter

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-001, FR-002
**Depends On:** #3

Implement `ProjectRepository` using the file system.

**Tasks:**

- [ ] `internal/adapters/filesystem/project_repo.go`
- [ ] Load project from loko.toml + directory scan
- [ ] Save project (create directories, write files)
- [ ] Parse YAML frontmatter from markdown
- [ ] Handle missing directories gracefully
- [ ] `internal/adapters/config/loader.go` - TOML config loading

**Acceptance Criteria:**

- Implements `usecases.ProjectRepository` interface
- Integration tests with temp directories
- Handles edge cases (missing files, permissions)

---

### 🔲 Issue #5b: Implement ason template engine adapter

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-006, FR-007
**Depends On:** #3

Integrate ason library for template scaffolding.

**Tasks:**

- [ ] Add `github.com/madstone-tech/ason` dependency
- [ ] `internal/adapters/ason/engine.go` - TemplateEngine implementation
- [ ] Template discovery (global ~/.loko/templates/ + project .loko/templates/)
- [ ] Variable interpolation with ason syntax
- [ ] Create starter templates using ason format
- [ ] Unit tests with sample templates

**Acceptance Criteria:**

- Implements `usecases.TemplateEngine` interface
- Loads templates from both global and project directories
- Renders templates with provided variables
- Starter templates (standard-3layer, serverless) work correctly

**References:**

- https://github.com/madstone-tech/ason
- https://context7.com/madstone-tech/ason/llms.txt

---

### 🔲 Issue #6: Wire up basic CLI with dependency injection

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-012, NFR-011
**Depends On:** #4, #5, #5b

Create main.go and basic CLI commands.

**Tasks:**

- [ ] `main.go` with dependency injection (wire adapters → use cases)
- [ ] `cmd/root.go` - Root command with global flags
- [ ] `cmd/init.go` - `loko init` command
- [ ] `cmd/new.go` - `loko new system` command
- [ ] `internal/ui/styles.go` - Lipgloss styles
- [ ] `internal/ui/output.go` - Success/error formatting

**Acceptance Criteria:**

- `loko init myproject` creates project structure
- `loko new system PaymentService` creates system files
- Commands are thin (<50 lines each)
- Output is nicely formatted with lipgloss

---

## Phase 3: Build Pipeline

### 🔲 Issue #7: Implement D2 diagram renderer adapter

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-004, FR-005
**Depends On:** #3

Implement `DiagramRenderer` using the d2 CLI.

**Tasks:**

- [ ] `internal/adapters/d2/renderer.go`
- [ ] Shell out to d2 binary
- [ ] Content-based caching (hash → output path)
- [ ] Configurable theme/layout
- [ ] Parallel rendering support
- [ ] Graceful error handling

**Acceptance Criteria:**

- Implements `usecases.DiagramRenderer` interface
- Cache hit returns immediately without calling d2
- Clear error messages when d2 missing
- Supports SVG and PNG output

---

### 🔲 Issue #8: Implement BuildDocs use case

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** US-2, US-5, FR-016
**Depends On:** #7

Use case for building documentation output.

**Tasks:**

- [ ] `internal/core/usecases/build_docs.go`
- [ ] Support multiple formats (HTML, markdown)
- [ ] Parallel diagram rendering
- [ ] Incremental builds (only changed files)
- [ ] Progress reporting

**Acceptance Criteria:**

- Builds HTML site from project
- Renders all diagrams (with caching)
- Reports progress via ProgressReporter
- Returns build statistics (files generated, cache hits)

---

### 🔲 Issue #9: Implement HTML site builder adapter

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-013
**Depends On:** #3

Generate static HTML documentation site.

**Tasks:**

- [ ] `internal/adapters/html/builder.go`
- [ ] `internal/adapters/html/templates/` - HTML templates
- [ ] Sidebar navigation
- [ ] Breadcrumbs
- [ ] Search (client-side)
- [ ] Responsive design
- [ ] Hot reload support (WebSocket)

**Acceptance Criteria:**

- Generates complete static site
- Navigation works correctly
- Mobile-friendly
- Search finds content

---

### 🔲 Issue #10: Add build, serve, watch CLI commands

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`
**Implements:** FR-012, NFR-002
**Depends On:** #8, #9

Complete the build pipeline CLI commands.

**Tasks:**

- [ ] `cmd/build.go` - `loko build` command
- [ ] `cmd/serve.go` - `loko serve` command with local server
- [ ] `cmd/watch.go` - `loko watch` command with file watching
- [ ] `cmd/render.go` - `loko render` for single diagrams
- [ ] `cmd/validate.go` - `loko validate` command

**Acceptance Criteria:**

- `loko build` generates dist/ directory
- `loko serve` starts server at localhost:8080
- `loko watch` rebuilds on file changes (<500ms)
- All commands under 50 lines

---

## Phase 4: MCP

### 🔲 Issue #11: Implement token-efficient architecture queries

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`, `mcp`
**Implements:** US-6, FR-009, FR-010, NFR-010
**Depends On:** #3

Implement progressive context loading for MCP.

**Tasks:**

- [ ] `internal/core/usecases/query_architecture.go`
- [ ] Summary level (~200 tokens)
- [ ] Structure level (~500 tokens)
- [ ] Full level (targeted)
- [ ] Compressed notation output option
- [ ] Unit tests

**Acceptance Criteria:**

- Summary for 20-system project < 300 tokens
- Structure for 20-system project < 600 tokens
- Full returns only requested entity
- Compressed notation parseable by LLMs

**References:**

- ADR 0002: Token-Efficient MCP Queries

---

### 🔲 Issue #12: Implement MCP server with core tools

**Priority:** High | **Labels:** `enhancement`, `v0.1.0`, `mcp`
**Implements:** US-1, FR-008, NFR-011
**Depends On:** #4, #8, #11

Create MCP server that exposes use cases as tools.

**Tasks:**

- [ ] `internal/mcp/server.go` - Protocol handler (stdio)
- [ ] `internal/mcp/tools/registry.go` - Tool registration
- [ ] `internal/mcp/tools/query_project.go`
- [ ] `internal/mcp/tools/query_architecture.go`
- [ ] `internal/mcp/tools/create_system.go`
- [ ] `internal/mcp/tools/create_container.go`
- [ ] `internal/mcp/tools/update_diagram.go`
- [ ] `internal/mcp/tools/build_docs.go`
- [ ] `internal/mcp/tools/validate.go`
- [ ] `cmd/mcp.go` - `loko mcp` command

**Acceptance Criteria:**

- Tools call same use cases as CLI
- Tool handlers < 30 lines each
- JSON schemas for all tool inputs
- Works with Claude Desktop

---

### 🔲 Issue #13: Create documentation and working examples

**Priority:** Medium | **Labels:** `documentation`, `v0.1.0`
**Implements:** SC-001
**Depends On:** #10, #12

Write user documentation and create example projects.

**Tasks:**

- [ ] `docs/quickstart.md` - 5-minute tutorial
- [ ] `docs/configuration.md` - loko.toml reference
- [ ] `docs/architecture.md` - Clean Architecture explanation
- [ ] `docs/mcp-integration.md` - LLM setup guide
- [ ] `examples/simple-project/` - Minimal example
- [ ] `examples/3layer-app/` - Web/API/DB example
- [ ] CI job that builds examples

**Acceptance Criteria:**

- Quickstart completable in <5 minutes
- Examples build without errors in CI
- MCP integration tested with Claude

---

## Phase 5: v0.2.0 Features

### 🔲 Issue #14: Add TOON format support for MCP responses

**Priority:** Medium | **Labels:** `enhancement`, `v0.2.0`, `mcp`, `optimization`
**Implements:** FR-022, FR-023, FR-024, SC-010
**Depends On:** #11

Implement TOON as optional output format for architecture queries.

**Tasks:**

- [ ] Add `toon-format/toon-go` dependency
- [ ] Create `OutputEncoder` interface in `ports.go`
- [ ] `internal/adapters/encoding/json_encoder.go` (default)
- [ ] `internal/adapters/encoding/toon_encoder.go`
- [ ] Add `format` parameter to `QueryArchitectureInput`
- [ ] Update MCP tool schema and handler
- [ ] Add format hint to TOON responses
- [ ] Benchmark token usage: JSON vs TOON
- [ ] Document usage in MCP integration guide

**Acceptance Criteria:**

- `query_architecture` accepts `format: "toon"` parameter
- TOON output is valid per toon-format spec
- Token reduction of 30%+ verified
- Format hint included in TOON responses

**References:**

- https://toonformat.dev/
- https://github.com/toon-format/toon-go
- ADR 0003: TOON Format Support

---

### 🔲 Issue #15: Implement HTTP API server

**Priority:** Medium | **Labels:** `enhancement`, `v0.2.0`
**Implements:** US-4
**Depends On:** #4, #8

Add HTTP API for CI/CD integration.

**Tasks:**

- [ ] `internal/api/server.go` - HTTP server setup
- [ ] `internal/api/middleware/auth.go` - API key auth
- [ ] `internal/api/middleware/logging.go`
- [ ] `internal/api/handlers/systems.go`
- [ ] `internal/api/handlers/build.go`
- [ ] `internal/api/handlers/validate.go`
- [ ] `internal/api/routes.go`
- [ ] `cmd/api.go` - `loko api` command

**Acceptance Criteria:**

- REST endpoints work correctly
- API key authentication
- Handlers call same use cases as CLI/MCP
- OpenAPI documentation generated

---

## Dependency Graph

```
#1 ─────────────────────────────────────────────────────────┐
#2 ─────────────────────────────────────────────────────────┼─► Foundation Complete
#3 ─────────────────────────────────────────────────────────┘
    │
    ├──► #4 CreateSystem ──┬──► #6 CLI ──► #10 CLI Commands
    ├──► #5 FileSystem ────┤
    ├──► #5b ason ─────────┘
    │
    ├──► #7 D2 Renderer ──► #8 BuildDocs ──► #10
    │                                    └──► #15 HTTP API
    ├──► #9 HTML Builder ──► #10
    │
    └──► #11 QueryArch ──► #12 MCP Server ──► #13 Docs
                       └──► #14 TOON (v0.2.0)
```

## Next Action

**Start with Issue #3: Define use case ports (interfaces)**

This unblocks all adapter implementations and use cases.
