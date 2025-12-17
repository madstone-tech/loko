<!--
  SYNC IMPACT REPORT (v1.0.0)
  Generated: 2025-12-16

  Version Change: Template → v1.0.0 (Initial Ratification)
  Action: Complete constitution created from template with user-supplied principles

  New Sections Added:
  - Core Values (6 principles)
  - Code Quality (5 requirements)
  - Architecture Principles (8 principles)
  - Documentation Standards (5 standards)
  - Community Guidelines (4 guidelines)
  - Release Standards (4 standards)
  - Performance Requirements (5 metrics)
  - Security (4 requirements)

  Template Updates Required:
  ✅ plan-template.md - Verified: Complexity Tracking aligns with Simplicity principle
  ✅ spec-template.md - Verified: Requirements capture aligns with all principles
  ✅ tasks-template.md - Verified: Task categorization aligns with testing discipline

  Follow-up Items: None - all values concrete and actionable
-->

# loko Constitution

## Core Values

### I. Simplicity Over Complexity

Minimize dependencies and maintain clear abstractions. Every addition must justify its
complexity cost. Prefer built-in abstractions to frameworks. MUST avoid unnecessary
layers of indirection or organizational patterns that don't solve concrete problems.

**Rationale**: Reduces maintenance burden, makes codebase approachable for new
contributors, and keeps cognitive load manageable as the project scales.

### II. User Empowerment

Provide both LLM-native and human-friendly interfaces. Users must never feel locked
into a single interaction model. Both conversational design (MCP) and direct CLI usage
MUST be equally supported and discoverable.

**Rationale**: Lowers barriers to adoption; accommodates diverse workflows; ensures
tool remains useful even as AI/LLM landscape evolves.

### III. Convention Over Configuration

Ship with smart defaults that solve 80% of cases. Configuration files are optional,
not required. When users must configure, the rationale MUST be explicit in docs.

**Rationale**: Reduces friction in getting started; keeps loko.toml minimal and focused
on real project variations, not tuning knobs.

### IV. Composability

Integrate with existing specialized tools (d2, ason, veve-cli) rather than
reimplementing their functionality. Loko orchestrates; it does not duplicate.

**Rationale**: Respects the UNIX philosophy; keeps codebase focused; benefits from
upstream improvements in those tools; reduces maintenance burden.

### V. Transparency

Build in public. Document decisions via Architecture Decision Records. Maintain a
visible roadmap and open issue discussions. Rationale for non-obvious choices MUST be
available.

**Rationale**: Builds community trust; makes contribution easier; attracts users who
value clarity; holds the project accountable to its principles.

### VI. Token Efficiency

Minimize LLM token consumption through smart data representation and progressive
context loading. Every feature touching the MCP interface MUST include token cost
considerations.

**Rationale**: Reduces user operational costs; makes loko economically viable for
large projects; encourages sustainable AI-assisted workflows.

## Code Quality Standards

### Language & Modern Idioms

- Go 1.25+ syntax and conventions MUST be followed
- Use generics where they reduce boilerplate (e.g., shared container logic)
- Prefer error wrapping with context (via errors.Join, %w) over bare error returns
- Use functional options pattern for configurable types instead of builder methods

**Rationale**: Keeps code maintainable across Go versions; makes the codebase feel
cohesive to Go developers.

### Interfaces for Testability

All external dependencies (filesystem, HTTP, config, rendering engines) MUST be
defined as Go interfaces in the core package. Implementations are in adapters.
Testing MUST use concrete mocks or in-memory implementations, never third-party
mocking libraries.

**Rationale**: Makes it trivial to test business logic without external I/O; keeps
tests fast and deterministic; ensures dependency inversion is enforced by the type
system.

### Error Handling with Context

Every error MUST include contextual information. Use wrapped errors to preserve
caller-visible stack. Error messages for users MUST be formatted with lipgloss for
readable CLI output (color, bold, indentation). Internal errors MAY use raw text.

**Rationale**: Errors are the primary way users discover what went wrong; good error
messages reduce support burden and build user trust.

### Structured Logging

Logs MUST be emitted in JSON format to stdout (production-ready). Use a structured
logger (zap or slog preferred). Log level MUST be configurable via env var or flag.
Development mode MAY use human-readable format as an alternative.

**Rationale**: Enables log aggregation in production; makes debugging easier; log
structure enables automated alerting/monitoring.

### Test Coverage

Unit tests MUST cover core business logic (>80% coverage for internal/core/).
Integration tests MUST cover user-facing workflows. Golden tests MUST validate
generated output (HTML, diagrams, markdown). Contract tests MUST validate interfaces
between adapters and core.

**Rationale**: Gives confidence in refactoring; catches regressions early; golden
tests act as specification for output format.

## Architecture Principles

### Clean Architecture with Dependency Inversion

The codebase MUST follow strict Clean Architecture boundaries:

```
internal/
├── core/                     # Business logic, zero external dependencies
│   ├── entities/             # Domain objects (Project, System, Container, etc.)
│   ├── usecases/             # Application logic orchestration
│   └── ports/                # Interfaces (repositories, services, ports)
├── adapters/                 # Infrastructure implementations
│   ├── filesystem/           # File I/O
│   ├── d2/                   # D2 integration
│   ├── encoding/             # JSON, TOON serialization
│   └── html/                 # HTML generation
├── mcp/                      # MCP server (thin layer calling use cases)
├── api/                      # HTTP API server (thin layer)
└── cli/                      # CLI commands (thin layer)
```

**Rule**: core/ MUST have zero imports from adapters, mcp, api, or cli. Adapters MAY
import from core. Thin layers (cli, api, mcp) MAY import from adapters and core.

**Rationale**: Ensures business logic is testable and independent of infrastructure
choices; makes it trivial to swap implementations (e.g., different file storage).

### Entities: Pure Structs with Validation

Domain entities (Project, System, Container, Component) MUST be pure Go structs with
no methods except validation and serialization helpers. MUST NOT embed HTTP handlers,
database connection logic, or I/O concerns.

**Rationale**: Entities remain easy to test and understand; validation logic is
explicit; reduces cognitive load.

### Use Cases Orchestrate Business Logic

A use case (e.g., "Create System", "Build Documentation") MUST:

1. Accept input (validated by caller)
2. Invoke ports (interfaces) to fetch/persist data
3. Call domain logic (entities)
4. Return output or error

Use cases MUST NOT directly access files, HTTP, or render output.

**Rationale**: Business logic is testable without I/O; easy to compose use cases;
clear separation of concerns.

### Adapters Implement Ports

Each port (interface) is implemented by exactly one adapter. Adapters are swappable
without changing business logic.

Examples:

- `ProjectLoader` interface → `FilesystemProjectLoader` adapter
- `DiagramRenderer` interface → `D2DiagramRenderer` adapter
- `DocumentBuilder` interface → `HTMLDocumentBuilder` adapter

**Rationale**: Easy to swap implementations; clear what each adapter is responsible
for; enables testing with fake adapters.

### CLI, MCP, and API as Thin Layers

CLI commands, MCP tools, and API endpoints MUST directly call use cases. They MUST
NOT contain business logic. Their sole job: parse input, validate user permissions,
call use case, format output.

**Rationale**: Business logic lives in one place; easy to add new interfaces (CLI,
API, web UI) by just adding a thin layer.

### Dependency Injection at Startup

All dependencies MUST be wired in main.go at startup. MUST NOT use global state, init()
functions, or DI frameworks. Use constructor injection (pass dependencies as function
parameters).

**Rationale**: Makes dependencies explicit; easy to see what the application needs;
no hidden initialization order issues.

### File System as Database

Loko MUST use the file system as the canonical storage. No hidden state, no caching
without explicit cache invalidation. Projects are directories on disk; all state is
represented as files.

**Rationale**: Users can edit files directly; version control works naturally; no
database setup friction.

### Immutable Builds

For the same input (source files, config, loko version), loko MUST always produce the
same output (HTML, diagrams). Non-determinism (timestamps, random IDs) MUST be
eliminated or mocked in tests.

**Rationale**: Enables caching; makes CI/CD reliable; users can trust that rebuilds
won't introduce spurious changes.

### Shell Out to Specialized Tools

Loko MUST shell out to d2 (diagram rendering) and veve-cli (PDF generation) rather
than reimplementing. Input/output is text (D2 source → SVG or PDF).

**Rationale**: Respects tool ownership; benefits from upstream improvements; keeps
loko's scope focused on orchestration, not rendering.

## Documentation Standards

### README Quick Start

README MUST include a working quick start in under 5 minutes. Commands MUST be
copy-paste ready. Quick start MUST end with a visible, testable result (e.g., URL to
open, file to inspect).

**Rationale**: First-time users must get immediate wins; reduces time-to-hello-world.

### Architecture Decision Records

Major decisions (tech choices, architectural changes, new principles) MUST be
documented in `docs/adr/` using the template in `.specify/templates/adr-template.md`.
ADRs MUST state context, decision, consequences, and rationale.

**Rationale**: Preserves institutional knowledge; explains why decisions were made;
helps future contributors understand tradeoffs.

### API Documentation via Godoc

All exported functions and types in internal/core/ MUST have Godoc comments. Comments
MUST be complete sentences. Examples MAY be included for complex functions.

**Rationale**: Enables `go doc` and pkg.go.dev to auto-generate API docs; makes the
code self-documenting.

### User Documentation in docs/

Feature documentation MUST live in `docs/`. Assume the reader is a user, not a
contributor. Structure: feature overview, use cases, examples, troubleshooting.

**Rationale**: Keeps docs separated from code; makes it easy for new users to find
help without diving into source.

### Examples Must Actually Work

Every example in docs/ MUST be tested in CI. Example projects MUST be runnable
end-to-end. Broken examples erode trust faster than missing docs.

**Rationale**: Prevents docs from becoming a source of confusion; users can follow
along with confidence.

## Community Guidelines

### Welcoming to All Skill Levels

Issues and PRs from new contributors MUST be treated with patience and respect.
PRs MUST NOT be rejected for style or small mistakes without offering to help fix
them. We are building in public; mistakes are learning opportunities.

**Rationale**: Reduces contribution friction; builds long-term community; makes the
project more diverse.

### Response SLA

Issues and PRs MUST receive a response (acknowledgment, question, or feedback) within
48 hours. If a maintainer cannot respond, another maintainer MUST. Abandoning issues
MUST NOT happen.

**Rationale**: Shows the project is alive; prevents contributor frustration; keeps
momentum.

### Semantic Versioning and Clear Changelogs

Loko MUST follow Semantic Versioning (MAJOR.MINOR.PATCH). Changelogs MUST be detailed
and human-readable. Each release MUST include: new features, bug fixes, breaking
changes (if any), and migration instructions (if needed).

**Rationale**: Users can make informed upgrade decisions; changelog is the first place
users check for news.

### Public Roadmap and Feature Discussions

A ROADMAP.md MUST list planned features and estimated timelines. Feature requests and
discussions MUST happen in GitHub Discussions or Issues (not private channels). Users
should never be surprised by major changes.

**Rationale**: Builds trust; lets users plan around releases; attracts contributors
interested in specific features.

### Bias Toward Action

MVPs are preferred to perfect designs. Small, working PRs beat large, complex ones.
Ship early and iterate based on user feedback. Perfectionism is the enemy of progress.

**Rationale**: Real feedback is better than speculation; early releases attract early
adopters; small changes are easier to review and reason about.

## Release Standards

### CI/CD with GitHub Actions

Every commit MUST run tests, linting, and build checks via GitHub Actions. Tests MUST
pass before merging to main. No manual release process.

**Rationale**: Catches issues early; ensures consistency; automates tedious work.

### Multi-Platform Testing

Tests MUST run on Linux, macOS, and Windows (if applicable). Loko MUST build and run
on all three platforms before release.

**Rationale**: Ensures cross-platform users aren't left behind; catches OS-specific
bugs.

### Goreleaser for Multi-Platform Binaries

Loko MUST distribute pre-built binaries for common platforms (linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64, windows/amd64) via goreleaser. Binaries MUST be signed.

**Rationale**: Users don't have to build from source; faster adoption; professional
appearance.

### Docker Images to GitHub Container Registry

Official Docker images MUST be published to ghcr.io for each release. Images MUST
include all dependencies (d2, veve-cli). Base image MUST be minimal (alpine or
distroless).

**Rationale**: Users with Docker get loko with zero setup; enables containerized
workflows.

### Homebrew Formula

A Homebrew formula MUST be maintained in a tap (madstone-tech/tap). Formula MUST
install the latest release.

**Rationale**: macOS/Linux users expect `brew install`; reduces friction.

## Performance Requirements

### Build Latency

Building 100 diagrams MUST complete in under 30 seconds (with disk caching). Watch
mode rebuild latency MUST be under 500ms. Rebuild MUST be incremental (not full
rebuild on every change).

**Rationale**: Keeps developer feedback loop tight; watch mode must feel snappy.

### Memory Usage

Loko MUST consume less than 100MB RAM for typical projects (100-200 documents). Memory
MUST not grow unbounded when watching for changes.

**Rationale**: Loko should be lightweight; no resource hogging on shared dev machines.

### Document Scale

Loko MUST support projects with 1000+ documents without degradation. Queries and
builds MUST remain fast at scale.

**Rationale**: No artificial limits; loko should grow with users' projects.

### Token Consumption

Architecture overview queries via MCP MUST consume < 500 tokens for a typical project.
Progressive context loading MUST be implemented to avoid ballpark estimates.

**Rationale**: Keeps LLM costs predictable; encourages AI-assisted workflows.

## Security

### No Arbitrary Code Execution

Loko MUST NOT eval(), exec(), or parse Go templates dynamically. Shell commands to
external tools (d2, veve-cli) MUST be constructed with fixed argument lists; user
input MUST NOT be interpolated into commands directly.

**Rationale**: Prevents shell injection attacks; users cannot be tricked into running
malicious code.

### Path Traversal Protection

All file operations MUST validate paths are within the project root. Symbolic links
MUST be resolved to check they don't escape the project. Users cannot reference files
outside their project directory.

**Rationale**: Multi-tenant safety; prevents users from accidentally (or maliciously)
reading/writing files outside their project.

### Input Sanitization for Shell Commands

D2 source code and other user inputs passed to shell commands MUST be quoted or
escaped. Use os/exec with slice arguments (not string interpolation).

**Rationale**: Prevents shell injection; d2 source might include backticks or shell
metacharacters.

### API Key Authentication

If HTTP API is enabled (optional feature), authentication MUST use bearer tokens (not
Basic auth or API keys in URLs). Tokens MUST be hashed before storage. HTTPS MUST be
enforced in production.

**Rationale**: Prevents credentials from being exposed in logs or URLs; API is secure
by default.

### Dependency Updates

Dependencies MUST be regularly updated (monthly Dependabot PRs minimum). Security
patches MUST be applied immediately. Dependency audit MUST run in CI.

**Rationale**: Reduces attack surface; catches CVEs early; users can trust loko is
actively maintained.

## Governance

### Constitution Supersedes All Practices

This Constitution is the source of truth for loko's governance. All code, design, and
process decisions MUST comply with these principles. Exceptions require documented
amendment to this Constitution.

### Amendment Procedure

To amend this Constitution:

1. **Proposal**: Open an issue or discussion explaining the change and rationale
2. **Discussion**: Community feedback period (minimum 7 days)
3. **Approval**: Maintainer approval (consensus preferred, majority acceptable)
4. **Documentation**: ADR created explaining the change
5. **Implementation**: Constitution updated with new version; CHANGELOG updated
6. **Propagation**: Affected templates and docs updated to reflect new principles

### Version Policy

Constitution follows Semantic Versioning:

- **MAJOR**: Backward-incompatible changes to core principles or architecture
- **MINOR**: New principles or sections added; expanded guidance
- **PATCH**: Clarifications, wording refinements, non-semantic corrections

### Compliance Review

At each major release (quarterly minimum), maintainers MUST review the codebase for
Constitution compliance. Non-compliant code MUST be refactored or justified via ADR.

### Guidance at Runtime

This Constitution is the governance layer. Runtime development guidance (workflows,
coding practices, how to set up a dev environment) lives in:

- `CONTRIBUTING.md` - How to contribute to loko itself
- `docs/` - How to use loko as a user
- `.specify/` - Project specification and planning templates
- ADRs in `docs/adr/` - Why decisions were made

**Version**: 1.0.0 | **Ratified**: 2025-12-16 | **Last Amended**: 2025-12-16
