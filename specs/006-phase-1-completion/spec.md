# Feature Specification: Production-Ready Phase 1 Release

**Feature Branch**: `006-phase-1-completion`  
**Created**: 2026-02-13  
**Status**: Draft  
**Target Release**: v0.2.0  
**Input**: User description: "Complete loko v0.2.0 Phase 1 with production-ready polish, TOON v3.0 compliance, handler refactoring, search MCP tools, and CI/CD examples"

---

## Overview

This specification defines the requirements for completing loko Phase 1 (v0.2.0) to production-ready quality. The release focuses on:

1. **Completing existing work** - TOON v3.0 compliance, handler refactoring, PDF graceful degradation, documentation polish
2. **Adding high-value quick wins** - Search/filter MCP tools, CI/CD examples, OpenAPI serving
3. **Establishing quality foundation** - Constitution compliance, benchmarking, comprehensive testing

**Philosophy**: Ship quality over quantity. v0.2.0 should be "Phase 1 Done Right" + "Phase 2/3 Quick Wins", NOT "Partial Phase 2".

---

## User Scenarios & Testing

### User Story 1 - LLM Agent Discovers Architecture Elements (Priority: P1)

As an LLM agent helping a user understand their architecture, I want to search and filter architecture elements by name, technology, or tags so I can quickly find relevant components without loading the entire architecture graph.

**Why this priority**: Search capability is foundational for LLM usability. Without it, agents must query the entire architecture and filter client-side, wasting tokens and time.

**Independent Test**: LLM can ask "find all Go components" and receive filtered results in under 200ms without loading full architecture.

**Acceptance Scenarios**:

1. **Given** a project with 50 components across 10 systems, **When** LLM searches "payment", **Then** only components/containers/systems matching "payment" are returned
2. **Given** a multi-technology stack, **When** LLM filters by "technology=Go", **Then** only Go-based elements are returned
3. **Given** elements tagged with "critical", **When** LLM searches "tag=critical", **Then** only critical-tagged elements are returned
4. **Given** a search with no results, **When** LLM queries non-existent element, **Then** empty result set is returned with helpful message

---

### User Story 2 - DevOps Engineer Validates Architecture in CI (Priority: P1)

As a DevOps engineer, I want to validate architecture documentation in CI/CD pipelines using standard YAML configuration files so architecture changes are automatically checked before merge.

**Why this priority**: CI/CD integration is essential for adoption in professional teams. Without it, architecture docs become stale as teams forget to validate manually.

**Independent Test**: Copy GitHub Actions example to `.github/workflows/`, push PR with invalid architecture → workflow fails with clear error.

**Acceptance Scenarios**:

1. **Given** GitHub Actions workflow configured, **When** PR contains orphaned references, **Then** workflow fails with exit code 1 and error details
2. **Given** GitLab CI pipeline configured, **When** architecture validates successfully, **Then** docs artifacts are uploaded
3. **Given** `loko validate --strict` flag, **When** architecture has warnings, **Then** command treats warnings as errors and exits non-zero
4. **Given** Docker Compose dev environment, **When** developer edits architecture files, **Then** watch mode rebuilds within 500ms

---

### User Story 3 - Architect Shares Token-Efficient Architecture with LLM (Priority: P1)

As a solution architect, I want to share architecture details with LLMs using TOON v3.0 format so I minimize token consumption and maximize what fits in context.

**Why this priority**: Token efficiency is loko's core differentiator. TOON v3.0 compliance ensures interoperability and validates marketing claims.

**Independent Test**: Export architecture in TOON format → official TOON parser successfully parses output → token count is 30-40% less than JSON.

**Acceptance Scenarios**:

1. **Given** architecture with 10 systems, **When** exported as TOON v3.0, **Then** output validates against official TOON parser
2. **Given** same architecture exported as JSON and TOON, **When** token counts compared, **Then** TOON uses 30-40% fewer tokens
3. **Given** TOON output with tabular arrays, **When** LLM reads format, **Then** structure is clearly parseable (systems, containers, components)
4. **Given** MCP tool requests TOON format, **When** response returned, **Then** format matches spec-compliant TOON v3.0 syntax

---

### User Story 4 - Developer Integrates loko API (Priority: P2)

As a developer building custom tooling, I want to discover and test the loko HTTP API using OpenAPI documentation so I can integrate architecture queries into my applications.

**Why this priority**: API discoverability accelerates integration. Swagger UI provides interactive testing without reading docs.

**Independent Test**: Start `loko api` → navigate to `/api/docs` → use Swagger UI to test GET `/api/v1/systems` → receive valid JSON response.

**Acceptance Scenarios**:

1. **Given** API server running, **When** developer navigates to `/api/docs`, **Then** Swagger UI loads with full API documentation
2. **Given** Swagger UI interactive testing, **When** developer tests authenticated endpoint, **Then** can add Bearer token and test successfully
3. **Given** OpenAPI spec at `/api/v1/openapi.json`, **When** developer validates spec, **Then** `openapi-generator validate` passes
4. **Given** API server help text, **When** developer runs `loko api --help`, **Then** OpenAPI/Swagger UI endpoints are documented

---

### User Story 5 - Contributor Follows Clean Architecture (Priority: P2)

As a loko contributor, I want all handlers to follow the constitution's thin-handler principle so I can understand the codebase quickly and know where business logic lives.

**Why this priority**: Code quality and maintainability. Technical debt compounds if not addressed before adding Phase 2 features (ADRs, Quality Attributes).

**Independent Test**: Run constitution audit script → all CLI handlers < 50 lines → all MCP tools < 30 lines → business logic in use cases only.

**Acceptance Scenarios**:

1. **Given** refactored `cmd/new.go`, **When** counted (excluding imports/comments), **Then** file is < 50 lines
2. **Given** business logic in use cases, **When** contributor adds new CLI command, **Then** handler only parses args, calls use case, formats response
3. **Given** MCP tools split into separate files, **When** contributor adds new tool, **Then** tool handler is < 30 lines following existing patterns
4. **Given** constitution audit in CI, **When** PR violates thin-handler principle, **Then** CI fails with clear violation message

---

### User Story 6 - User Builds PDF Without veve-cli (Priority: P3)

As a user without veve-cli installed, I want to build HTML and Markdown documentation without errors so I can use loko immediately without installing optional dependencies.

**Why this priority**: First-run experience. Users should get value immediately; PDF is nice-to-have, not required.

**Independent Test**: Fresh install without veve-cli → `loko build` → HTML/Markdown succeed → clear message about optional PDF with install link.

**Acceptance Scenarios**:

1. **Given** veve-cli not installed, **When** user runs `loko build`, **Then** HTML/Markdown succeed with warning about skipped PDF
2. **Given** veve-cli not installed, **When** user runs `loko build --format pdf`, **Then** helpful error message with installation instructions
3. **Given** `--skip-pdf` flag, **When** user runs `loko build --skip-pdf`, **Then** no PDF warnings shown
4. **Given** Docker image, **When** user runs loko container, **Then** veve-cli is pre-installed and PDF works

---

### User Story 7 - New User Completes MCP Setup (Priority: P2)

As a new loko user, I want step-by-step MCP integration documentation with Claude Desktop so I can start designing architecture conversationally without trial and error.

**Why this priority**: MCP is loko's hero feature. Poor onboarding means users miss the core value proposition.

**Independent Test**: New user follows MCP guide → configures Claude Desktop → chats "create a payment system" → loko scaffolds architecture → user views in browser.

**Acceptance Scenarios**:

1. **Given** MCP integration guide, **When** new user configures Claude Desktop, **Then** configuration works first try with provided JSON
2. **Given** example conversation flow, **When** user follows steps, **Then** completes init → scaffold → build → serve workflow
3. **Given** troubleshooting section, **When** user encounters common issue, **Then** finds solution in docs without support
4. **Given** MCP tool reference table, **When** user wants to know tool capabilities, **Then** can see all 17 tools with examples

---

### Edge Cases

- **Empty architecture**: What happens when `search_elements` is called on project with zero systems? → Returns empty result set with message "No elements found"
- **TOON format breaking changes**: What if TOON v3.0 spec changes after release? → Version TOON output, document migration path
- **CI pipeline timeout**: What if architecture validation takes longer than CI timeout? → Add `--timeout` flag, optimize validation performance
- **Rate limiting in local dev**: What if developer hits rate limit testing API locally? → Rate limiting disabled by default, only enabled in production config
- **Binary size bloat from Swagger UI**: How to prevent embedding Swagger UI from bloating binary? → Compress assets, use minimal Swagger UI build, document size increase
- **Multiple search filters**: What if user applies conflicting filters (e.g., "type=system" and "tag=component-level")? → Return empty set, document filter semantics
- **Glob pattern performance**: What if glob pattern matches thousands of elements? → Apply `limit` parameter (default 20), paginate results
- **Constitution audit false positives**: What if legitimate handler needs > 50 lines? → Document exceptions, require explicit approval in PR review

---

## Requirements

### Functional Requirements - Search & Filter

- **FR-001**: System MUST provide `search_elements` MCP tool that searches by name, description, tags, and technology
- **FR-002**: System MUST support glob patterns (e.g., `backend.*`, `*-service`) for element name matching
- **FR-003**: `search_elements` MUST filter by element type (system, container, component)
- **FR-004**: `search_elements` MUST limit results to prevent token overflow (default: 20, configurable)
- **FR-005**: System MUST provide `find_relationships` MCP tool that queries graph edges
- **FR-006**: `find_relationships` MUST support glob patterns for source and target filtering
- **FR-007**: Both tools MUST return results in under 200ms (excluding network latency)
- **FR-008**: Empty result sets MUST include helpful messages indicating why no matches found

### Functional Requirements - TOON v3.0 Compliance

- **FR-009**: TOON output MUST validate against official TOON v3.0 parser without errors
- **FR-010**: TOON format MUST use tabular array syntax for repeated structures (systems, containers, components)
- **FR-011**: TOON token count MUST be 30-40% less than equivalent JSON representation
- **FR-012**: System MUST provide benchmarks comparing JSON vs TOON token efficiency
- **FR-013**: README MUST document TOON format with real spec-compliant examples
- **FR-014**: TOON output MUST maintain backward compatibility with existing MCP clients OR provide migration guide
- **FR-015**: System MUST use `github.com/toon-format/toon-go` library (no custom parser)

### Functional Requirements - Handler Refactoring

- **FR-016**: All CLI command files MUST be < 50 lines (excluding imports/comments/blank lines)
- **FR-017**: All MCP tool handlers MUST be < 30 lines (excluding imports/comments/blank lines)
- **FR-018**: Business logic MUST reside exclusively in `internal/core/usecases/`
- **FR-019**: Domain services MUST be moved to appropriate adapters (not in `cmd/`)
- **FR-020**: Handlers MUST only parse arguments, call use cases, format responses, handle errors
- **FR-021**: System MUST eliminate runtime type assertions in `core/usecases` (use typed interfaces)
- **FR-022**: Constitution audit MUST pass in CI with zero violations
- **FR-023**: Test coverage MUST remain > 80% after refactoring

### Functional Requirements - CI/CD Integration

- **FR-024**: System MUST provide GitHub Actions workflow example in `examples/ci/`
- **FR-025**: System MUST provide GitLab CI pipeline example in `examples/ci/`
- **FR-026**: System MUST provide Docker Compose dev environment example
- **FR-027**: `loko validate` MUST support `--strict` flag that treats warnings as errors
- **FR-028**: `loko validate` MUST support `--exit-code` flag that returns non-zero on errors
- **FR-029**: CI examples MUST work on free tiers (GitHub Actions Free, GitLab Free)
- **FR-030**: CI examples MUST be tested in real pipelines before release
- **FR-031**: System MUST document CI/CD integration in `docs/guides/ci-cd-integration.md`

### Functional Requirements - OpenAPI Serving

- **FR-032**: API server MUST serve OpenAPI 3.0 spec at `/api/v1/openapi.json`
- **FR-033**: API server MUST serve OpenAPI YAML spec at `/api/v1/openapi.yaml`
- **FR-034**: API server MUST serve Swagger UI at `/api/docs`
- **FR-035**: OpenAPI spec MUST be accurate (validated against actual handlers)
- **FR-036**: OpenAPI spec MUST document Bearer token authentication
- **FR-037**: Swagger UI MUST work offline (no CDN dependencies)
- **FR-038**: Swagger UI static assets MUST be embedded using `go:embed`
- **FR-039**: API server startup time MUST not increase by more than 100ms
- **FR-040**: `loko api --help` MUST mention OpenAPI/Swagger UI endpoints

### Functional Requirements - PDF Graceful Degradation

- **FR-041**: `loko build` without veve-cli MUST build HTML/Markdown successfully
- **FR-042**: `loko build` without veve-cli MUST show warning about skipped PDF with installation link
- **FR-043**: `loko build --format pdf` without veve-cli MUST show error with veve-cli installation instructions
- **FR-044**: System MUST support `--skip-pdf` flag to suppress PDF warnings
- **FR-045**: Docker image MUST include veve-cli by default
- **FR-046**: Tests MUST cover both veve-cli present and absent scenarios
- **FR-047**: Error messages MUST follow loko UI style (lipgloss formatting)

### Functional Requirements - Documentation Polish

- **FR-048**: README MUST state accurate MCP tool count (17: 15 existing + 2 new)
- **FR-049**: README roadmap MUST reflect actual v0.1.0 shipped state
- **FR-050**: README MUST include spec-compliant TOON examples (not custom notation)
- **FR-051**: README MUST include verified token efficiency claims with benchmarks
- **FR-052**: System MUST provide MCP integration guide with Claude Desktop configuration
- **FR-053**: MCP guide MUST include example conversation flow (init → scaffold → build)
- **FR-054**: MCP guide MUST include troubleshooting section for common issues
- **FR-055**: MCP guide MUST include tool reference table with all 17 tools
- **FR-056**: All 4 examples MUST build successfully (simple-project, 3layer-app, serverless, microservices)
- **FR-057**: Each example MUST have README explaining what it demonstrates
- **FR-058**: README MUST include 2-3 minute demo GIF (< 5MB) showing full workflow

### Functional Requirements - Rate Limiting & CORS (Optional)

- **FR-059**: API server MAY provide rate limiting middleware (default: 100 req/min per IP)
- **FR-060**: Rate limiting MAY be configurable via `loko.toml`: `[api] rate_limit = 100`
- **FR-061**: CORS MAY be configurable via `loko.toml`: `[api] allowed_origins = [...]`
- **FR-062**: Request timeout MAY be configurable via `loko.toml`: `[api] timeout = "30s"`
- **FR-063**: Rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`) MAY be included in responses
- **FR-064**: 429 Too Many Requests response MAY include Retry-After header
- **FR-065**: Rate limiting MUST be disabled by default for local development

### Key Entities

**No new entities** - This release polishes existing entities:

- **ArchitectureGraph**: Enhanced with search/filter capabilities via new MCP tools
- **ProjectConfig**: Extended with optional API configuration (rate_limit, allowed_origins, timeout)
- **ValidationResult**: May include constitution audit results in addition to architecture validation

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can find architecture elements via search in under 200ms for projects with up to 100 components
- **SC-002**: TOON format reduces token count by 30-40% compared to JSON (measured via benchmark)
- **SC-003**: All CLI handlers are verifiable < 50 lines via automated constitution audit
- **SC-004**: All MCP tool handlers are verifiable < 30 lines via automated constitution audit
- **SC-005**: CI/CD examples work first-try on GitHub Actions and GitLab CI free tiers
- **SC-006**: New users complete MCP integration following guide in under 10 minutes
- **SC-007**: Watch mode rebuilds complete in under 500ms for 10-system projects
- **SC-008**: All 4 example projects build successfully without errors
- **SC-009**: README demo GIF loads in under 2 seconds on standard broadband
- **SC-010**: API server startup time remains under 100ms (< 10% increase from v0.1.0)
- **SC-011**: Test coverage remains > 80% in `internal/core/` after refactoring
- **SC-012**: Zero constitution violations reported by CI audit
- **SC-013**: TOON output validates against official parser with 100% success rate
- **SC-014**: OpenAPI spec validates with `openapi-generator validate` with zero errors
- **SC-015**: 90% of users can self-serve using documentation without support

### Qualitative Outcomes

- **SC-016**: Documentation accurately represents shipped functionality (no false claims)
- **SC-017**: Codebase architecture facilitates Phase 2 implementation (ADRs, Quality Attributes)
- **SC-018**: Error messages are clear and actionable for common failure scenarios
- **SC-019**: First-time user experience is smooth from install to first architecture build
- **SC-020**: Release is launch-ready for HN/Reddit with polished demo and accurate claims

---

## Assumptions

1. **TOON v3.0 specification is stable** - Assume official TOON v3.0 spec won't change significantly during development
2. **veve-cli remains optional** - PDF generation stays as nice-to-have feature, not required
3. **MCP protocol stability** - Assume MCP protocol changes are backward compatible or documented
4. **CI/CD platform availability** - Assume GitHub Actions and GitLab CI free tiers remain available
5. **Existing test suite quality** - Assume current > 80% coverage is comprehensive enough to catch regressions
6. **Constitution audit tooling** - Assume simple line-counting script is sufficient for handler size validation
7. **Search performance** - Assume graph-based search scales to 100 components without optimization beyond limiting results
8. **Swagger UI size** - Assume embedded Swagger UI adds < 5MB to binary (compressed assets)
9. **Docker base image** - Assume veve-cli can be installed in Docker image without size/complexity issues
10. **MCP client compatibility** - Assume Claude Desktop and other MCP clients handle TOON format changes gracefully

---

## Non-Goals (Out of Scope for v0.2.0)

The following features are explicitly **NOT** included in v0.2.0:

- ❌ **Architecture Decision Records (ADRs)** - Deferred to v0.3.0
- ❌ **Quality Attributes frontmatter** - Deferred to v0.3.0
- ❌ **Hosted Publishing** - Deferred to v0.4.0
- ❌ **License System** - Deferred to v0.4.0
- ❌ **Structurizr DSL Import** - Deferred to v1.0.0
- ❌ **Architecture Graph Visualization UI** - Deferred to v0.3.0
- ❌ **Diff and Changelog Generation** - Deferred to v0.3.0
- ❌ **Plugin System** - Deferred to v0.3.0
- ❌ **Advanced rate limiting** (per-user, token bucket) - Simple IP-based rate limiting only
- ❌ **SSO/SAML integration** - Deferred to v1.0.0
- ❌ **Audit logging** - Deferred to v1.0.0
- ❌ **Multi-workspace support** - Deferred to v0.4.0

---

## Dependencies & Constraints

### Technical Dependencies

- **External Libraries**:
  - `github.com/toon-format/toon-go` (TOON v3.0 encoding)
  - `golang.org/x/time/rate` (rate limiting - optional)
  - Swagger UI static assets (embedded via `go:embed`)

- **External Tools**:
  - `veve-cli` (optional, for PDF generation)
  - `d2` (required, for diagram rendering - already dependency)
  - `openapi-generator` (development only, for spec validation)

- **CI/CD Platforms**:
  - GitHub Actions (free tier)
  - GitLab CI (free tier)

### Architecture Constraints

From Constitution and AGENTS.md:

1. **Clean Architecture**: Core has zero external dependencies; dependency direction: CLI/MCP/API → Adapters → Use Cases → Entities
2. **Thin Handlers**: CLI < 50 lines, MCP < 30 lines, business logic in use cases only
3. **Interface Testability**: No global state, no `init()` functions, all dependencies injected
4. **File System is Database**: Markdown, TOML, D2, YAML frontmatter; no hidden state
5. **Immutable Builds**: Same input → same output; mock non-determinism in tests

### Performance Constraints

- Watch mode rebuild: < 500ms
- MCP tool response: < 100ms (excluding diagram rendering)
- Search tools response: < 200ms
- Graph operations: < 1ms
- Build time (10-system project): < 5 seconds
- API server startup: < 100ms increase from v0.1.0

### Test Coverage Constraints

- `internal/core/`: > 80% coverage
- All use cases: 100% coverage
- Critical paths: 100% coverage
- Integration tests for MCP/CLI/API

---

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| TOON v3.0 library doesn't support required features | Medium | High | Evaluate library early; implement custom formatter if needed; document as known limitation |
| Handler refactoring introduces regressions | Medium | High | Maintain 100% test coverage; refactor incrementally; run full test suite between changes |
| Swagger UI bloats binary size significantly | Low | Medium | Use compressed minimal build; document size increase; make embedding optional if > 10MB |
| CI examples don't work on all platforms | Medium | Medium | Test on actual GitHub Actions and GitLab CI before release; provide troubleshooting docs |
| Search performance degrades with large architectures | Low | Medium | Implement result limiting (default 20); add pagination if needed; benchmark with 100+ component project |
| TOON format changes break MCP clients | Low | High | Version TOON output; provide backward compatibility mode; document migration path |
| Constitution audit produces false positives | Low | Low | Allow documented exceptions; require PR review approval for violations |
| veve-cli installation fails in Docker | Low | Medium | Use well-tested base image; provide fallback instructions; make PDF fully optional |

---

## Timeline Estimate

**Total Estimated Effort**: 11-18 days (2-3 weeks)

**Week 1** (Complete Existing Work):
- Days 1-2: TOON v3.0 compliance + benchmarking
- Days 3-4: Handler refactoring (CLI + MCP)
- Day 5: PDF graceful degradation + initial docs polish

**Week 2** (High-Value Quick Wins):
- Days 1-2: Search & filter MCP tools implementation + tests
- Day 3: CI/CD examples + `--strict`/`--exit-code` flags
- Day 4: OpenAPI serving + Swagger UI embedding
- Day 5: Rate limiting/CORS (if time permits) + buffer

**Week 3** (Polish & Release):
- Days 1-2: Documentation polish (README, MCP guide, examples)
- Day 3: Demo GIF creation + final example verification
- Day 4: Integration testing + performance benchmarking
- Day 5: Release preparation + launch materials
