# Research & Analysis: loko v0.1.0

**Created**: 2025-12-17  
**Status**: Phase 0 Research Complete

---

## Technology Stack Analysis

### Go Language Choice ✅

**Decision**: Go 1.25+ for loko core

**Rationale**:

- Single static binary compilation (no runtime dependencies except d2, veve-cli)
- Fast execution (critical for watch mode <500ms latency)
- Clean Architecture support (interfaces, dependency inversion)
- Excellent CLI frameworks (Cobra)
- Strong JSON support (native encoding/json)
- Minimal dependencies philosophy aligns with loko Constitution

**Risks**:

- Learning curve for contributors unfamiliar with Go
- Smaller ecosystem for some domains vs. Node.js/Python

**Mitigation**: AGENTS.md and CONTRIBUTING.md provide clear guidance

---

### Framework Choices

#### CLI: Cobra ✅

**Why**: Industry standard for Go CLIs (kubectl, docker, github-cli use it)

- Consistent command structure
- Built-in help generation
- Flag parsing
- Subcommand support

**Usage**: Thin wrapper pattern (<50 lines) per command

---

#### Configuration: Viper + TOML ✅

**Why**: Viper handles multiple config sources (TOML, env, flags)

- TOML human-readable format
- Supports both local and global config
- Hierarchical config (loko.toml overrides ~/.loko/config.toml)

**Design**: Adapter pattern (`internal/adapters/config/loader.go`)

---

#### UI: Lipgloss + Bubbletea ✅

**Why**: Lipgloss for styled CLI output, Bubbletea for interactive prompts

- Beautiful terminal output (matches loko's design philosophy)
- Interactive CLI for `loko init` prompts
- Only used in CLI layer (not core business logic)

**Usage**: `internal/ui/` package for formatting and styling

---

### External Tool Integration

#### D2 CLI Shell-Out ✅

**Decision**: Shell out to d2 binary rather than embedding renderer

**Rationale**:

- D2 is actively maintained (not our responsibility)
- Respects tool ownership (Composability principle)
- Users install d2 once globally
- Simpler than embedding D2 engine
- Future: D2 library if needed

**Implementation**: `internal/adapters/d2/renderer.go`

- os/exec with timeout
- Content-based caching (hash → output path)
- Error handling when d2 missing

**Risks**:

- D2 version compatibility
- Platform-specific d2 binary behavior
- Timeout handling for large diagrams

**Mitigation**:

- Pin D2 version in docs
- Test on all platforms (CI)
- Configurable timeout in loko.toml

---

#### veve-cli Shell-Out (Optional) ✅

**Decision**: Shell out to veve-cli for PDF generation (optional dependency)

**Rationale**:

- Owned by MADSTONE TECH (governance aligned)
- PDF is complex; veve-cli handles it
- Optional feature (v0.2.0)
- Graceful degradation if missing

**Implementation**: `internal/adapters/pdf/renderer.go`

- Check if veve-cli available at startup
- Return error if PDF requested but veve-cli missing
- Document requirement in error message

---

### Library Choices

#### ason (Template Scaffolding) ✅

**Library**: `github.com/madstone-tech/ason`

**Why**:

- Lightweight template engine
- Owned by MADSTONE TECH (stable)
- Simple variable substitution
- Perfect for C4 templates

**Context7 Docs**: <https://context7.com/madstone-tech/ason/llms.txt>

**Implementation**: `internal/adapters/ason/engine.go`

- Template discovery (global + project)
- Variable substitution
- Error handling for missing templates

---

#### toon-go (Token Efficiency) ✅

**Library**: `github.com/toon-format/toon-go` (v0.2.0)

**Why**:

- TOON format for 30-40% token reduction
- Official library (maintained)
- Critical for LLM cost efficiency (US-6)

**Usage** (v0.2.0):

- `internal/adapters/encoding/toon_encoder.go`
- Optional output format for MCP queries
- Benchmark token reduction

---

#### fsnotify (File Watching) ✅

**Library**: `github.com/fsnotify/fsnotify`

**Why**:

- Standard Go file watcher
- Cross-platform (Linux, macOS, Windows)
- Mature and stable

**Implementation**: `internal/adapters/filesystem/watcher.go`

- Watch source directory
- Trigger rebuild on file changes
- Debounce rapid changes

---

#### gomarkdown (Markdown Parsing) ✅

**Library**: `github.com/go-echarts/markdown` or similar

**Why**:

- Parse markdown structure
- Extract frontmatter metadata
- Navigation structure for HTML

**Implementation**: `internal/adapters/markdown/parser.go`

- Extract YAML frontmatter
- Parse markdown hierarchy
- Support for D2 code blocks

---

### MCP Integration ✅

**Protocol**: MCP (Model Context Protocol)

**Why**:

- Standardized protocol for LLM tool integration
- Supports Claude Desktop, other clients
- JSON-RPC over stdio
- Safe for untrusted LLM execution

**Design Pattern**:

- `internal/mcp/server.go` - Protocol handler
- `internal/mcp/tools/` - Tool implementations
- Tools call same use cases as CLI (no duplication)

**Tools (8 total)**:

1. `query_project` - Get project metadata
2. `query_architecture` - Token-efficient queries
3. `create_system` - Scaffold system
4. `create_container` - Scaffold container
5. `create_component` - Scaffold component
6. `update_diagram` - Write D2 code
7. `build_docs` - Build documentation
8. `validate` - Check architecture

---

## Architecture Decisions

### Clean Architecture ✅

**Chosen Pattern**: Clean Architecture with strict dependency inversion

**Layers**:

1. **core/** (Business Logic) - Zero external dependencies beyond stdlib
   - entities/ (Domain objects: Project, System, Container, Component)
   - usecases/ (Application logic: CreateSystem, BuildDocs, QueryArchitecture)
   - ports/ (Interfaces: ProjectRepository, DiagramRenderer, etc.)

2. **adapters/** (Infrastructure)
   - filesystem/ (File I/O, ProjectRepository implementation)
   - d2/ (D2 diagram rendering)
   - html/ (Static site generation)
   - ason/ (Template scaffolding)
   - config/ (TOML configuration loading)
   - logging/ (JSON structured logging)
   - markdown/ (Markdown parsing)
   - encoding/ (JSON/TOON encoding)

3. **CLI/MCP/API** (Thin Wrappers)
   - cmd/ (CLI commands - Cobra)
   - internal/mcp/ (MCP server)
   - internal/api/ (HTTP API - future)

**Rationale**:

- Business logic testable without I/O
- Easy to swap implementations (e.g., different file storage)
- Follows Dependency Inversion Principle
- Aligns with loko Constitution (NFR-009: zero external deps in core)

**Risks**:

- Initial setup overhead (more files, more interfaces)
- Team must understand pattern

**Mitigation**: AGENTS.md and CONTRIBUTING.md explain pattern clearly

---

### File System as Database ✅

**Design**: Projects = directories on disk

**Structure**:

```
myproject/
├── loko.toml              # Configuration
├── src/                   # Source files
│   ├── context.md         # Context level
│   ├── context.d2
│   └── PaymentService/    # System
│       ├── system.md
│       ├── system.d2
│       └── API/           # Container
│           ├── container.md
│           └── container.d2
└── dist/                  # Generated output
    ├── index.html
    ├── diagrams/
    │   └── *.svg
    └── ...
```

**Rationale**:

- No database setup friction
- Users can edit files directly
- Version control friendly (git works natively)
- No hidden state

**Advantages**:

- Simple, transparent
- Works with any editor
- Easy to backup/share
- No database schema to migrate

**Challenges**:

- Must handle file system inconsistencies
- Locking for concurrent access (mitigation: warn user)
- Path traversal security (mitigation: validate paths)

---

### Token-Efficient Queries (US-6) ✅

**Problem**: LLM context windows are expensive. Full project JSON could consume 1000+ tokens.

**Solution**: Progressive context loading with detail levels

**Levels**:

1. **summary** (~200 tokens): System names, counts
2. **structure** (~500 tokens): Systems + containers, no components
3. **full** (variable): All details for one entity

**Format Options**:

- JSON (default, readable)
- TOON (v0.2.0, 30-40% token reduction)

**Implementation**: `internal/core/usecases/query_architecture.go`

**Example Outputs**:

Summary (payment system):

```json
{
  "project": "PaymentService",
  "system_count": 3,
  "systems": ["API", "Database", "Payment Gateway"]
}
```

Structure:

```json
{
  "systems": [
    {
      "name": "API",
      "containers": ["Web Service", "Cache"]
    },
    ...
  ]
}
```

---

## Design Decisions

### Incremental Builds ✅

**Challenge**: Building 100 diagrams takes 30 seconds. Watch mode needs <500ms latency.

**Solution**: Incremental builds

1. Track which files changed
2. Only rebuild changed diagrams and affected HTML
3. Cache rendered diagrams by content hash

**Implementation**:

- `internal/core/usecases/build_docs.go` - Orchestration
- `internal/adapters/d2/renderer.go` - Caching logic
- `internal/adapters/filesystem/watcher.go` - Change detection

**Cache Strategy**:

- Input: D2 source code
- Key: SHA256 hash of D2 content
- Value: Rendered SVG file path
- Invalidation: On d2 version change (detected at startup)

---

### Hot Reload UI ✅

**Challenge**: Watch mode <500ms requires fast feedback

**Solution**: Browser hot reload via WebSocket or Server-Sent Events

**Design**:

1. `loko serve` starts HTTP server on :8080
2. Injects small JavaScript into HTML pages
3. JavaScript opens WebSocket connection
4. On file change, server notifies client
5. Browser refreshes (or partial reload if possible)

**Implementation**: `internal/adapters/html/server.go`

---

### MCP Tool Design ✅

**Pattern**: Tools call same use cases as CLI (no duplication)

**Example Flow**:

```
MCP Tool (query_architecture)
  ↓
core/usecases/QueryArchitecture UC
  ↓
Ports (ProjectRepository, OutputEncoder)
  ↓
Adapters (FileSystem, JSON/TOON Encoder)
```

**Constraints**:

- Tool handlers <30 lines (thin wrappers)
- JSON schemas for inputs
- Clear error messages

---

## Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
| --- | --- | --- | --- |
| D2 CLI behavior varies across versions | Medium | Medium | Pin version, test on all platforms |
| MCP protocol changes | Low | High | Use stable SDK, pin version |
| ason API instability | Low | Medium | Owned by MADSTONE, version pinned |
| File system concurrency issues | Medium | Low | Document single-writer assumption |
| Token estimation wrong | Medium | Low | Benchmark, iterate based on real usage |
| Watch mode performance degradation at scale | Low | Medium | Profile, optimize hot paths |

### Schedule Risks

| Risk | Likelihood | Impact | Mitigation |
| --- | --- | --- | --- |
| Scope creep (feature bloat) | High | High | Strict v0.1.0 scope, defer to v0.2.0 |
| MCP integration complexity | Medium | Medium | Build vertical slice early (Slice 3) |
| HTML generation complexity | Medium | Medium | Start simple, iterate on design |
| Template system edge cases | Medium | Low | Use ason's proven implementation |

---

## Dependencies Map

```
Go 1.23+
├── stdlib (no external deps in core/)
├── External Tools
│   ├── d2 (diagram rendering)
│   └── veve-cli (PDF export, optional)
├── Libraries
│   ├── Cobra (CLI framework)
│   ├── Viper (configuration)
│   ├── ason (template scaffolding)
│   ├── fsnotify (file watching)
│   ├── gomarkdown (markdown parsing)
│   ├── Lipgloss (UI styling)
│   ├── Bubbletea (interactive CLI)
│   ├── toon-go (token efficiency, v0.2.0)
│   └── MCP SDK
└── CI/CD
    ├── GitHub Actions
    ├── goreleaser (binary distribution)
    └── golangci-lint (code quality)
```

---

## Assumptions

1. **D2 available**: Users install d2 separately
2. **Go installed**: Development requires Go 1.23+
3. **Text editor available**: Users have VSCode, Vim, etc.
4. **Internet access**: LLM interaction requires network
5. **Single writer**: Projects assume single developer (or careful coordination)
6. **File permissions**: Users have read/write to project directory

---

## Success Metrics (Testable)

1. **Build performance**: 100 diagrams in <30s ✅
2. **Watch latency**: File edit to browser refresh <500ms ✅
3. **Memory**: <100MB for 50-system project ✅
4. **Token efficiency**: <500 tokens for 20-system overview ✅
5. **Scaffold time**: `loko init` → docs viewed in <2 minutes ✅
6. **LLM integration**: Claude can design architecture via MCP ✅

---

## Open Questions & Decisions Needed

| Question | Decision | Owner |
| --- | --- | --- |
| Which HTML template engine (inline vs. external)? | Inline Go templates (`text/template`) | Architecture Review |
| Breadcrumb navigation structure? | Automatic from file hierarchy | Design Review |
| Search implementation (client-side or server-side)? | Client-side (simple.js or similar) | Architecture Review |
| Maximum project size before performance degrades? | Test up to 1000 systems | QA/Performance |

---
