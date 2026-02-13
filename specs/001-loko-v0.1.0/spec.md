# Feature Specification: loko v0.1.0

**Feature Branch**: `001-loko-v0.1.0`
**Created**: 2025-12-17
**Status**: Ready for Planning
**Spec Version**: 0.1.0-dev

---

## Overview

**loko** is a C4 model architecture documentation tool that enables Cloud Solution Architects to design systems conversationally with LLM agents via MCP, while also providing powerful CLI and API interfaces for direct interaction.

This specification covers the complete v0.1.0 MVP release, including all P1 priority user stories.

---

## User Stories & Acceptance Criteria

### US-1: LLM-Driven Architecture Design (Priority: P1)

As a Cloud Solution Architect, I want to use an LLM chatbot (Claude, GPT, Gemini) connected to loko via MCP to have a conversational workflow where the LLM guides me through designing architecture, and the end result is well-documented C4 model architecture with diagrams and markdown documentation.

**Why this priority**: Conversational AI-assisted design is the core differentiator for loko. This is the primary user journey and must work flawlessly.

**Independent Test**: Claude Desktop can design a complete 3-system architecture (Context → Systems → Containers) via MCP tools without human intervention

**Acceptance Scenarios:**

1. **Given** a new project, **When** I chat with the LLM and say "I'm building a payment processing system", **Then** the LLM calls loko MCP tools to initialize the project structure
2. **Given** the LLM asks "what containers do you need?", **When** I respond "API and Database", **Then** the LLM scaffolds these containers using loko MCP tools
3. **Given** the conversation progresses, **When** the LLM generates D2 diagram code, **Then** loko renders it to SVG and embeds it in documentation
4. **Given** the architecture is defined, **When** I ask "show me the docs", **Then** loko builds HTML documentation I can view in a browser
5. **Given** I ask "what's the current architecture?", **When** the LLM queries loko, **Then** it returns a token-efficient summary without consuming excessive context

---

### US-2: Direct File Editing Workflow (Priority: P1)

As a developer, I want to edit .md and .d2 files directly in my text editor (VSCode, Vim), and have loko automatically rebuild documentation in real-time, so I can work in my preferred environment without depending on LLMs.

**Why this priority**: Enables traditional developer workflows. Not everyone wants to use LLMs; local file editing is critical for accessibility and control.

**Independent Test**: User can `loko watch`, edit a .d2 file, and see the browser refresh with new diagram within 500ms, without any manual rebuild command

**Acceptance Scenarios:**

1. **Given** loko is running in watch mode, **When** I edit a .d2 file, **Then** it automatically re-renders the diagram within 500ms
2. **Given** I save a markdown file, **When** loko rebuilds, **Then** the HTML output updates and my browser auto-refreshes
3. **Given** I create a new system folder manually, **When** I run loko validate, **Then** it reports any missing required files
4. **Given** I want to preview, **When** I run loko serve, **Then** I get a local web server showing rendered documentation

---

### US-3: Project Scaffolding (Priority: P1)

As a developer, I want to quickly scaffold C4 documentation structure using templates, so I can start with good conventions and consistent structure across systems.

**Why this priority**: First-time user experience is critical. `loko init` and `loko new` must be intuitive and fast.

**Independent Test**: User can run `loko init myproject` → `loko new system PaymentService` → view generated files in file system and see consistent C4 structure

**Acceptance Scenarios:**

1. **Given** I run loko init, **When** I provide project details interactively, **Then** loko creates project structure with loko.toml configuration
2. **Given** an initialized project, **When** I run loko new system PaymentService, **Then** it scaffolds system.md and system.d2 from templates
3. **Given** a system exists, **When** I run loko new container PaymentService API, **Then** it creates container docs under the system
4. **Given** I want custom templates, **When** I place templates in .loko/templates/, **Then** loko uses them instead of global templates

---

### US-4: API Integration (Priority: P2)

As a DevOps engineer, I want to trigger loko builds via HTTP API in CI/CD pipelines, so I can automate documentation generation and validation as part of deployment workflows.

**Why this priority**: P2 - Important for automation but can be added after CLI works. Foundation in v0.1.0, full implementation in v0.2.0.

**Independent Test**: CI/CD script can POST to `http://localhost:8081/api/v1/build` and receive JSON response indicating build success/failure

**Acceptance Scenarios:**

1. **Given** loko API server is running, **When** I POST to /api/v1/build, **Then** it builds documentation and returns status
2. **Given** I want to query structure, **When** I GET /api/v1/systems, **Then** I receive JSON listing all systems
3. **Given** API auth is enabled, **When** I call without API key, **Then** I get 401 Unauthorized
4. **Given** a build completes, **When** I GET /api/v1/validate, **Then** I receive validation report with any issues

---

### US-5: Multi-Format Export (Priority: P2)

As an architect, I want to export documentation to multiple formats (HTML, Markdown, PDF), so I can share architecture with different audiences and use cases.

**Why this priority**: P2 - HTML is MVP; markdown and PDF can follow. Users can work with HTML in v0.1.0.

**Independent Test**: User can run `loko build --format markdown` and get a single README.md file with complete architecture content

**Acceptance Scenarios:**

1. **Given** documentation exists, **When** I run loko build --format html, **Then** I get a static website I can deploy
2. **Given** I need a single file, **When** I run loko build --format markdown, **Then** I get one README.md with all content
3. **Given** I want PDFs, **When** I run loko build --format pdf (and veve-cli is installed), **Then** I get PDF documents
4. **Given** I want all formats, **When** I run loko build, **Then** it generates HTML, markdown, and PDF based on loko.toml config

---

### US-6: Token-Efficient Architecture Queries (Priority: P1)

As an LLM agent, I want to query architecture with configurable detail levels, so I can get context without consuming excessive tokens.

**Why this priority**: P1 - Core to MCP performance. Excessive token usage makes LLM integration expensive and unusable.

**Independent Test**: Query a 20-system project: summary returns <300 tokens, structure returns <600 tokens, TOON format returns 30%+ fewer tokens than JSON

**Acceptance Scenarios:**

1. **Given** I need a quick overview, **When** I call query_architecture with detail:"summary", **Then** I get ~200 tokens with counts and system names
2. **Given** I need to understand structure, **When** I call with detail:"structure", **Then** I get ~500 tokens with systems and their containers
3. **Given** I need full details on one system, **When** I call with target:"PaymentService" and detail:"full", **Then** I get complete info for only that system
4. **Given** I want maximum efficiency, **When** I call with format:"toon", **Then** I get TOON-encoded response with 30-40% fewer tokens

---

## Functional Requirements

### Core Configuration (FR-001 to FR-003)

| ID     | Requirement                                                                                   | Priority | Status |
| ------ | --------------------------------------------------------------------------------------------- | -------- | ------ |
| FR-001 | System MUST support TOML configuration (loko.toml) with validation                            | P1       | Draft  |
| FR-002 | System MUST parse YAML frontmatter in markdown files for metadata                             | P1       | Draft  |
| FR-003 | System MUST support both global (~/.loko/templates/) and project (.loko/templates/) templates | P1       | Draft  |

### Diagram Rendering (FR-004 to FR-005)

| ID     | Requirement                                                        | Priority | Status |
| ------ | ------------------------------------------------------------------ | -------- | ------ |
| FR-004 | System MUST shell out to d2 CLI for diagram rendering with caching | P1       | Draft  |
| FR-005 | System MUST support parallel D2 rendering for performance          | P1       | Draft  |

### Template System (FR-006 to FR-007)

| ID     | Requirement                                                               | Priority | Status |
| ------ | ------------------------------------------------------------------------- | -------- | ------ |
| FR-006 | System MUST integrate ason as a Go library for template scaffolding       | P1       | Draft  |
| FR-007 | System MUST include two starter templates: standard-3layer and serverless | P1       | Draft  |

### MCP Interface (FR-008 to FR-011)

| ID     | Requirement                                                                                                                                                           | Priority | Status |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ |
| FR-008 | System MUST provide MCP server with tools: query_project, query_architecture, create_system, create_container, create_component, update_diagram, build_docs, validate | P1       | Draft  |
| FR-009 | System MUST provide progressive context loading via MCP with summary/structure/full detail levels                                                                     | P1       | Draft  |
| FR-010 | System MUST support targeted queries (specific system/container) to avoid loading entire project context                                                              | P1       | Draft  |
| FR-011 | System SHOULD provide compressed notation format for architecture relationships                                                                                       | P2       | Draft  |

### CLI Interface (FR-012)

| ID     | Requirement                                                                                          | Priority | Status |
| ------ | ---------------------------------------------------------------------------------------------------- | -------- | ------ |
| FR-012 | System MUST support CLI commands: init, new, build, serve, watch, render, validate, doctor, mcp, api | P1       | Draft  |

### HTML Generation (FR-013)

| ID     | Requirement                                                                                        | Priority | Status |
| ------ | -------------------------------------------------------------------------------------------------- | -------- | ------ |
| FR-013 | System MUST generate static HTML site with sidebar navigation, breadcrumbs, search, and hot reload | P1       | Draft  |

### Logging & Validation (FR-014 to FR-015)

| ID     | Requirement                                                                                           | Priority | Status |
| ------ | ----------------------------------------------------------------------------------------------------- | -------- | ------ |
| FR-014 | System MUST log in JSON format with structured fields for production observability                    | P1       | Draft  |
| FR-015 | System MUST validate architecture for orphaned references, missing files, and C4 hierarchy violations | P1       | Draft  |

### Build System (FR-016 to FR-017)

| ID     | Requirement                                                         | Priority | Status |
| ------ | ------------------------------------------------------------------- | -------- | ------ |
| FR-016 | System MUST support incremental builds (only rebuild changed files) | P1       | Draft  |
| FR-017 | System MUST provide Docker images for containerized usage           | P2       | Draft  |

### PDF Generation (FR-018)

| ID     | Requirement                                                       | Priority | Status |
| ------ | ----------------------------------------------------------------- | -------- | ------ |
| FR-018 | System MUST shell out to veve-cli for PDF generation when enabled | P2       | Draft  |

### Clean Architecture (FR-019 to FR-021)

| ID     | Requirement                                                                               | Priority | Status |
| ------ | ----------------------------------------------------------------------------------------- | -------- | ------ |
| FR-019 | System MUST implement Clean Architecture with clear separation                            | P1       | Draft  |
| FR-020 | All use cases MUST be callable from CLI, MCP, and API without code duplication            | P1       | Draft  |
| FR-021 | All external dependencies (d2, file system, veve-cli) MUST be accessed through interfaces | P1       | Draft  |

### Token Efficiency (FR-022 to FR-024)

| ID     | Requirement                                                                         | Priority | Status |
| ------ | ----------------------------------------------------------------------------------- | -------- | ------ |
| FR-022 | System SHOULD support TOON as optional output format for MCP queries                | P2       | Draft  |
| FR-023 | When TOON format is requested, system MUST use official toon-format/toon-go library | P2       | Draft  |
| FR-024 | MCP tool descriptions MUST include format hints when TOON is used                   | P2       | Draft  |

---

## Non-Functional Requirements

### Performance (NFR-001 to NFR-003)

| ID      | Requirement                | Target                      | Status |
| ------- | -------------------------- | --------------------------- | ------ |
| NFR-001 | Build 100 diagrams         | < 30 seconds (with caching) | Draft  |
| NFR-002 | Watch mode rebuild latency | < 500ms                     | Draft  |
| NFR-003 | Memory usage (50 systems)  | < 100MB                     | Draft  |

### Compatibility (NFR-004 to NFR-006)

| ID      | Requirement                                                                    | Status |
| ------- | ------------------------------------------------------------------------------ | ------ |
| NFR-004 | Support Linux, macOS, and Windows with identical behavior                      | Draft  |
| NFR-005 | Single binary with no runtime dependencies except d2 (and optionally veve-cli) | Draft  |
| NFR-006 | Graceful degradation if optional dependencies (veve-cli) are missing           | Draft  |

### User Experience (NFR-007 to NFR-008)

| ID      | Requirement                                                         | Status |
| ------- | ------------------------------------------------------------------- | ------ |
| NFR-007 | Clear, actionable error messages with suggestions (using lipgloss)  | Draft  |
| NFR-008 | Comprehensive test coverage (>80%) with CI running on all platforms | Draft  |

### Architecture (NFR-009 to NFR-011)

| ID      | Requirement                              | Target             | Status |
| ------- | ---------------------------------------- | ------------------ | ------ |
| NFR-009 | Core package external dependencies       | Zero (stdlib only) | Draft  |
| NFR-010 | Architecture overview query (20 systems) | < 500 tokens       | Draft  |
| NFR-011 | New CLI command or MCP tool code         | < 50 lines         | Draft  |

---

## Key Entities

| Entity    | Description                       | Files                       |
| --------- | --------------------------------- | --------------------------- |
| Project   | Root configuration and metadata   | loko.toml                   |
| System    | C4 system level                   | system.md, system.d2        |
| Container | C4 container level                | container.md, container.d2  |
| Component | C4 component level                | component.md, component.d2  |
| Template  | Reusable scaffolding template     | template.toml + .tmpl files |
| Diagram   | D2 source and rendered output     | .d2 → SVG/PNG               |
| Build     | Generated documentation artifacts | HTML, markdown, PDF         |

---

## Success Criteria

| ID     | Criterion                                 | Target                                 | Status |
| ------ | ----------------------------------------- | -------------------------------------- | ------ |
| SC-001 | Time from `loko init` to viewing docs     | < 2 minutes                            | Draft  |
| SC-002 | LLM designs 3-system architecture via MCP | Without human intervention             | Draft  |
| SC-003 | Watch mode feedback loop                  | < 500ms                                | Draft  |
| SC-004 | HTML documentation                        | Navigable, searchable, mobile-friendly | Draft  |
| SC-005 | Validation catches mistakes               | > 90%                                  | Draft  |
| SC-006 | CI/CD integration                         | Exit codes for failures                | Draft  |
| SC-007 | Docker image size                         | < 50MB                                 | Draft  |
| SC-008 | New CLI/MCP tool code                     | < 50 lines                             | Draft  |
| SC-009 | Architecture overview (20 systems)        | < 500 tokens                           | Draft  |
| SC-010 | TOON vs JSON token reduction              | > 30%                                  | Draft  |

---

## External References

- [C4 Model](https://c4model.com/)
- [D2 Language](https://d2lang.com/)
- [MCP Protocol](https://modelcontextprotocol.io/)
- [ason Documentation](https://context7.com/madstone-tech/ason/llms.txt)
- [TOON Format](https://toonformat.dev/)

---

## Edge Cases & Exclusions

### In Scope for v0.1.0

- ✅ 6 core user stories (3 P1 + 1 P1 + 1 P1 + 1 P2 + 1 P2 + 1 P1)
- ✅ CLI commands for scaffolding, building, serving, watching
- ✅ MCP server with core tools
- ✅ HTML documentation generation
- ✅ File watching and hot reload
- ✅ Token-efficient queries

### Out of Scope for v0.1.0 (v0.2.0+)

- ❌ Markdown/PDF export (v0.2.0)
- ❌ HTTP API full implementation (foundation in v0.1.0, complete in v0.2.0)
- ❌ TOON format support (v0.2.0)
- ❌ Advanced validation rules beyond basic C4 hierarchy
- ❌ Plugin system
- ❌ Database backend (file system only in v0.1.0)
- ❌ Web UI editor (CLI + LLM only in v0.1.0)

---

## Assumptions & Dependencies

### Assumptions

- Users have a text editor (VSCode, Vim, etc.) available
- D2 binary is installed (graceful degradation if missing)
- LLM clients have MCP support (Claude Desktop tested)
- Users have internet for LLM interaction

### Dependencies

- Go 1.23+
- D2 binary (required)
- veve-cli (optional, for PDF)
- MCP SDK
- ason library
- toon-go library (v0.2.0)

---

## Implementation Scope & Slices

### Slice 1: Scaffolding (US-3)
**Duration**: ~1-2 weeks
**Deliverable**: Users can scaffold projects with `loko init` and `loko new`

### Slice 2: File Editing (US-2)
**Duration**: ~2-3 weeks
**Deliverable**: Users can edit files and see live updates with `loko watch`

### Slice 3: MCP Integration (US-1)
**Duration**: ~2-3 weeks
**Deliverable**: Claude can design architecture via MCP

### Slice 4: Token Efficiency (US-6)
**Duration**: ~1-2 weeks
**Deliverable**: MCP queries verified token efficient

### Slice 5: API + Export (US-4, US-5)
**Duration**: ~2-3 weeks
**Deliverable**: HTTP API and multi-format export working

---
