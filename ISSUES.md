# GitHub Issues to Create

Create these issues in order to track the development of loko v0.1.0 and v0.2.0.

## Phase 1: Foundation (Week 1)

### Issue #1: Initialize project with Clean Architecture structure

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
Set up the project with Clean Architecture directory structure and basic CI.

**Tasks:**

- [ ] Initialize Go module (`github.com/madstone-tech/loko`)
- [ ] Create directory structure (cmd/, internal/core/, internal/adapters/, etc.)
- [ ] Set up GitHub Actions for test/lint/build
- [ ] Add Makefile with common commands
- [ ] Configure golangci-lint
- [ ] Add goreleaser config (for later)
- [ ] Create ADR directory with template

**Acceptance Criteria:**

- `go build` succeeds
- `go test ./...` runs (even if no tests yet)
- CI passes on PR
- Directory structure matches Clean Architecture

---

### Issue #2: Implement core domain entities

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
Create the domain entities in `internal/core/entities/` with validation.

**Tasks:**

- [ ] `project.go` - Project entity with systems collection
- [ ] `system.go` - System entity with containers collection
- [ ] `container.go` - Container entity with components
- [ ] `component.go` - Component entity
- [ ] `diagram.go` - Diagram entity (source + rendered)
- [ ] `template.go` - Template entity
- [ ] `validation.go` - Validation logic for all entities
- [ ] `errors.go` - Domain error types

**Acceptance Criteria:**

- All entities have constructors (NewSystem, etc.)
- Validation returns structured errors
- 100% test coverage for validation logic
- Zero external dependencies in this package

---

### Issue #3: Define use case ports (interfaces)

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
Define all interfaces in `internal/core/usecases/ports.go` that adapters must implement.

**Tasks:**

- [ ] `ProjectRepository` - Load/Save project
- [ ] `TemplateRepository` - List/Get templates
- [ ] `DiagramRenderer` - Render D2 diagrams
- [ ] `TemplateEngine` - Render scaffolding templates
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
- Interfaces are minimal (don't over-abstract)

---

## Phase 2: First Use Case End-to-End (Week 2)

### Issue #4: Implement CreateSystem use case

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

### Issue #5: Implement file system adapter

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

### Issue #6: Wire up basic CLI with dependency injection

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

## Phase 3: Build Pipeline (Week 3)

### Issue #7: Implement D2 diagram renderer adapter

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

### Issue #8: Implement BuildDocs use case

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

### Issue #9: Implement HTML site builder adapter

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

### Issue #10: Add build, serve, watch CLI commands

**Labels:** `enhancement`, `v0.1.0`, `priority:high`

**Description:**
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

## Phase 4: MCP (Week 4)

### Issue #11: Implement token-efficient architecture queries

**Labels:** `enhancement`, `v0.1.0`, `mcp`, `priority:high`

**Description:**
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

### Issue #12: Implement MCP server with core tools

**Labels:** `enhancement`, `v0.1.0`, `mcp`, `priority:high`

**Description:**
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

### Issue #13: Create documentation and working examples

**Labels:** `documentation`, `v0.1.0`, `priority:medium`

**Description:**
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

### Issue #14: Add TOON format support for MCP responses

**Labels:** `enhancement`, `v0.2.0`, `mcp`, `optimization`

**Description:**
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

- <https://toonformat.dev/>
- <https://github.com/toon-format/toon-go>
- ADR 0003: TOON Format Support

---

### Issue #15: Implement HTTP API server

**Labels:** `enhancement`, `v0.2.0`, `priority:medium`

**Description:**
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

## Quick Create Script

Use GitHub CLI to create all issues:

```bash
# Phase 1
gh issue create --title "Initialize project with Clean Architecture structure" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Implement core domain entities" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Define use case ports (interfaces)" --label "enhancement,v0.1.0,priority:high"

# Phase 2
gh issue create --title "Implement CreateSystem use case" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Implement file system adapter" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Wire up basic CLI with dependency injection" --label "enhancement,v0.1.0,priority:high"

# Phase 3
gh issue create --title "Implement D2 diagram renderer adapter" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Implement BuildDocs use case" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Implement HTML site builder adapter" --label "enhancement,v0.1.0,priority:high"
gh issue create --title "Add build, serve, watch CLI commands" --label "enhancement,v0.1.0,priority:high"

# Phase 4
gh issue create --title "Implement token-efficient architecture queries" --label "enhancement,v0.1.0,mcp,priority:high"
gh issue create --title "Implement MCP server with core tools" --label "enhancement,v0.1.0,mcp,priority:high"
gh issue create --title "Create documentation and working examples" --label "documentation,v0.1.0,priority:medium"

# Phase 5 (v0.2.0)
gh issue create --title "Add TOON format support for MCP responses" --label "enhancement,v0.2.0,mcp,optimization"
gh issue create --title "Implement HTTP API server" --label "enhancement,v0.2.0,priority:medium"
```
