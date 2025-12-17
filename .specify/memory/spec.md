# loko Specification

> Generated: 2024-12-17
> Status: Active
> Version: 0.1.0-dev

## Overview

**loko** is a C4 model architecture documentation tool that enables Cloud Solution Architects to design systems conversationally with LLM agents via MCP, while also providing powerful CLI and API interfaces for direct interaction.

## User Stories

### US-1: LLM-Driven Architecture Design (P1)

As a Cloud Solution Architect, I want to use an LLM chatbot (Claude, GPT, Gemini) connected to loko via MCP to have a conversational workflow where the LLM guides me through designing architecture, and the end result is well-documented C4 model architecture with diagrams and markdown documentation.

**Acceptance Scenarios:**

1. Given a new project, when I chat with the LLM and say "I'm building a payment processing system", then the LLM calls loko MCP tools to initialize the project structure
2. Given the LLM asks "what containers do you need?", when I respond "API and Database", then the LLM scaffolds these containers using loko MCP tools
3. Given the conversation progresses, when the LLM generates D2 diagram code, then loko renders it to SVG and embeds it in documentation
4. Given the architecture is defined, when I ask "show me the docs", then loko builds HTML documentation I can view in a browser
5. Given I ask "what's the current architecture?", when the LLM queries loko, then it returns a token-efficient summary without consuming excessive context

### US-2: Direct File Editing Workflow (P1)

As a developer, I want to edit .md and .d2 files directly in my text editor (VSCode, Vim), and have loko automatically rebuild documentation in real-time, so I can work in my preferred environment without depending on LLMs.

**Acceptance Scenarios:**

1. Given loko is running in watch mode, when I edit a .d2 file, then it automatically re-renders the diagram within 500ms
2. Given I save a markdown file, when loko rebuilds, then the HTML output updates and my browser auto-refreshes
3. Given I create a new system folder manually, when I run loko validate, then it reports any missing required files
4. Given I want to preview, when I run loko serve, then I get a local web server showing rendered documentation

### US-3: Project Scaffolding (P1)

As a developer, I want to quickly scaffold C4 documentation structure using templates, so I can start with good conventions and consistent structure across systems.

**Acceptance Scenarios:**

1. Given I run loko init, when I provide project details interactively, then loko creates project structure with loko.toml configuration
2. Given an initialized project, when I run loko new system PaymentService, then it scaffolds system.md and system.d2 from templates
3. Given a system exists, when I run loko new container PaymentService API, then it creates container docs under the system
4. Given I want custom templates, when I place templates in .loko/templates/, then loko uses them instead of global templates

### US-4: API Integration (P2)

As a DevOps engineer, I want to trigger loko builds via HTTP API in CI/CD pipelines, so I can automate documentation generation and validation as part of deployment workflows.

**Acceptance Scenarios:**

1. Given loko API server is running, when I POST to /api/v1/build, then it builds documentation and returns status
2. Given I want to query structure, when I GET /api/v1/systems, then I receive JSON listing all systems
3. Given API auth is enabled, when I call without API key, then I get 401 Unauthorized
4. Given a build completes, when I GET /api/v1/validate, then I receive validation report with any issues

### US-5: Multi-Format Export (P2)

As an architect, I want to export documentation to multiple formats (HTML, Markdown, PDF), so I can share architecture with different audiences and use cases.

**Acceptance Scenarios:**

1. Given documentation exists, when I run loko build --format html, then I get a static website I can deploy
2. Given I need a single file, when I run loko build --format markdown, then I get one README.md with all content
3. Given I want PDFs, when I run loko build --format pdf (and veve-cli is installed), then I get PDF documents
4. Given I want all formats, when I run loko build, then it generates HTML, markdown, and PDF based on loko.toml config

### US-6: Token-Efficient Architecture Queries (P1)

As an LLM agent, I want to query architecture with configurable detail levels, so I can get context without consuming excessive tokens.

**Acceptance Scenarios:**

1. Given I need a quick overview, when I call query_architecture with detail:"summary", then I get ~200 tokens with counts and system names
2. Given I need to understand structure, when I call with detail:"structure", then I get ~500 tokens with systems and their containers
3. Given I need full details on one system, when I call with target:"PaymentService" and detail:"full", then I get complete info for only that system
4. Given I want maximum efficiency, when I call with format:"toon", then I get TOON-encoded response with 30-40% fewer tokens

## Functional Requirements

### Core Configuration

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-001 | System MUST support TOML configuration (loko.toml) with validation | P1 |
| FR-002 | System MUST parse YAML frontmatter in markdown files for metadata | P1 |
| FR-003 | System MUST support both global (~/.loko/templates/) and project (.loko/templates/) templates | P1 |

### Diagram Rendering

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-004 | System MUST shell out to d2 CLI for diagram rendering with caching | P1 |
| FR-005 | System MUST support parallel D2 rendering for performance | P1 |

### Template System

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-006 | System MUST integrate ason as a Go library for template scaffolding | P1 |
| FR-007 | System MUST include two starter templates: standard-3layer and serverless | P1 |

### MCP Interface

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-008 | System MUST provide MCP server with tools: query_project, query_architecture, create_system, create_container, create_component, update_diagram, build_docs, validate | P1 |
| FR-009 | System MUST provide progressive context loading via MCP with summary/structure/full detail levels | P1 |
| FR-010 | System MUST support targeted queries (specific system/container) to avoid loading entire project context | P1 |
| FR-011 | System SHOULD provide compressed notation format for architecture relationships | P2 |

### CLI Interface

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-012 | System MUST support CLI commands: init, new, build, serve, watch, render, validate, doctor, mcp, api | P1 |

### HTML Generation

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-013 | System MUST generate static HTML site with sidebar navigation, breadcrumbs, search, and hot reload | P1 |

### Logging & Validation

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-014 | System MUST log in JSON format with structured fields for production observability | P1 |
| FR-015 | System MUST validate architecture for orphaned references, missing files, and C4 hierarchy violations | P1 |

### Build System

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-016 | System MUST support incremental builds (only rebuild changed files) | P1 |
| FR-017 | System MUST provide Docker images for containerized usage | P2 |

### PDF Generation

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-018 | System MUST shell out to veve-cli for PDF generation when enabled | P2 |

### Clean Architecture

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-019 | System MUST implement Clean Architecture with clear separation | P1 |
| FR-020 | All use cases MUST be callable from CLI, MCP, and API without code duplication | P1 |
| FR-021 | All external dependencies (d2, file system, veve-cli) MUST be accessed through interfaces | P1 |

### Token Efficiency

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-022 | System SHOULD support TOON as optional output format for MCP queries | P2 |
| FR-023 | When TOON format is requested, system MUST use official toon-format/toon-go library | P2 |
| FR-024 | MCP tool descriptions MUST include format hints when TOON is used | P2 |

## Non-Functional Requirements

### Performance

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-001 | Build 100 diagrams | < 30 seconds (with caching) |
| NFR-002 | Watch mode rebuild latency | < 500ms |
| NFR-003 | Memory usage (50 systems) | < 100MB |

### Compatibility

| ID | Requirement |
|----|-------------|
| NFR-004 | Support Linux, macOS, and Windows with identical behavior |
| NFR-005 | Single binary with no runtime dependencies except d2 (and optionally veve-cli) |
| NFR-006 | Graceful degradation if optional dependencies (veve-cli) are missing |

### User Experience

| ID | Requirement |
|----|-------------|
| NFR-007 | Clear, actionable error messages with suggestions (using lipgloss) |
| NFR-008 | Comprehensive test coverage (>80%) with CI running on all platforms |

### Architecture

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-009 | Core package external dependencies | Zero (stdlib only) |
| NFR-010 | Architecture overview query (20 systems) | < 500 tokens |
| NFR-011 | New CLI command or MCP tool code | < 50 lines |

## Key Entities

| Entity | Description | Files |
|--------|-------------|-------|
| Project | Root configuration and metadata | loko.toml |
| System | C4 system level | system.md, system.d2 |
| Container | C4 container level | container.md, container.d2 |
| Component | C4 component level | component.md, component.d2 |
| Template | Reusable scaffolding template | template.toml + .tmpl files |
| Diagram | D2 source and rendered output | .d2 → SVG/PNG |
| Build | Generated documentation artifacts | HTML, markdown, PDF |

## Technical Constraints

### Language & Framework

- Go 1.23+
- Clean Architecture: core/ has zero external dependencies
- Cobra for CLI framework (thin wrapper over use cases)
- Viper for configuration (adapter layer only)
- Bubbletea and Lipgloss for TUI/styling (UI layer only)

### External Tools (Shell Out)

- d2 binary for diagram rendering (behind DiagramRenderer interface)
- veve-cli binary for PDF generation (behind PDFRenderer interface)

### Libraries (Go Import)

- github.com/madstone-tech/ason - Template scaffolding
  - Docs: https://context7.com/madstone-tech/ason/llms.txt
- github.com/toon-format/toon-go - TOON encoding (v0.2.0)
- fsnotify - File watching (adapter layer)
- gomarkdown - Markdown parsing (adapter layer)

### Protocols

- MCP via stdio transport
- JSON structured logging

### Architecture Rules

- All external dependencies accessed through interfaces in usecases/ports.go
- CLI commands are thin wrappers (<50 lines)
- MCP tool handlers are thin wrappers (<30 lines)

## Success Criteria

| ID | Criterion | Target |
|----|-----------|--------|
| SC-001 | Time from `loko init` to viewing docs | < 2 minutes |
| SC-002 | LLM designs 3-system architecture via MCP | Without human intervention |
| SC-003 | Watch mode feedback loop | < 500ms |
| SC-004 | HTML documentation | Navigable, searchable, mobile-friendly |
| SC-005 | Validation catches mistakes | > 90% |
| SC-006 | CI/CD integration | Exit codes for failures |
| SC-007 | Docker image size | < 50MB |
| SC-008 | New CLI/MCP tool code | < 50 lines |
| SC-009 | Architecture overview (20 systems) | < 500 tokens |
| SC-010 | TOON vs JSON token reduction | > 30% |

## Architecture References

- ADR-0001: Clean Architecture
- ADR-0002: Token-Efficient MCP Queries
- ADR-0003: TOON Format Support

## External References

- [C4 Model](https://c4model.com/)
- [D2 Language](https://d2lang.com/)
- [MCP Protocol](https://modelcontextprotocol.io/)
- [ason Documentation](https://context7.com/madstone-tech/ason/llms.txt)
- [TOON Format](https://toonformat.dev/)
