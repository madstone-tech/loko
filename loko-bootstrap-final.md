# loko Bootstrap Package (Final)

> Complete project bootstrap incorporating:
> - Clean Architecture
> - Token-Efficient MCP (Progressive Context Loading)
> - TOON Format Support

---

# Table of Contents

1. [Constitution Prompt](#1-constitution-prompt)
2. [Specification Prompt](#2-specification-prompt)
3. [README.md](#3-readmemd)
4. [CONTRIBUTING.md](#4-contributingmd)
5. [CODE_OF_CONDUCT.md](#5-code_of_conductmd)
6. [ROADMAP.md](#6-roadmapmd)
7. [Architecture Decision Records](#7-architecture-decision-records)
8. [Initial GitHub Issues](#8-initial-github-issues)
9. [Project Structure](#9-project-structure)
10. [Quick Start Commands](#10-quick-start-commands)

---

# 1. Constitution Prompt

Use this with speckit to establish project principles:

```
/speckit.constitution Create principles for loko - an open-source C4 architecture documentation tool built in Go. Focus on:

CORE VALUES:
- Simplicity over complexity - minimize dependencies, clear abstractions
- User empowerment - provide both LLM and human interfaces
- Convention over configuration - smart defaults, minimal setup
- Composability - integrate with existing tools (d2, ason, veve-cli) rather than reimplementing
- Transparency - building in public, clear communication
- Token efficiency - minimize LLM costs through smart data representation

CODE QUALITY:
- Go 1.23+ with modern idioms
- Interfaces for all external dependencies (testability)
- Error handling with context and helpful messages (lipgloss formatting)
- Structured logging in JSON format (production-ready)
- Comprehensive test coverage (unit, integration, golden tests)

ARCHITECTURE PRINCIPLES:
- Clean Architecture with strict dependency inversion
- Core package (internal/core/) has zero external dependencies beyond stdlib
- Interfaces (ports) defined by use cases, implemented by adapters
- Entities are pure Go structs with validation logic
- Use cases orchestrate business logic, adapters handle infrastructure
- CLI, MCP, and API are thin layers calling shared use cases
- Dependency injection at startup (main.go), no DI framework
- File system is the database (no hidden state)
- Immutable builds (same input = same output)
- Shell out to specialized tools (d2, veve-cli) rather than reimplementing

DOCUMENTATION STANDARDS:
- README with quick start in under 5 minutes
- Architecture Decision Records (ADR) for major decisions
- API documentation via godoc
- User documentation in docs/ directory
- Examples that actually work (tested in CI)

COMMUNITY GUIDELINES:
- Welcoming to contributors of all skill levels
- Issues and PRs get response within 48 hours
- Semantic versioning and clear changelogs
- Public roadmap and feature discussions
- Bias toward action - ship MVPs and iterate

RELEASE STANDARDS:
- CI/CD with GitHub Actions
- Automated testing on Linux, macOS, Windows
- goreleaser for multi-platform binaries
- Docker images published to GitHub Container Registry
- Homebrew formula for macOS/Linux

PERFORMANCE REQUIREMENTS:
- Build 100 diagrams in under 30 seconds (with caching)
- Watch mode rebuild latency under 500ms
- Memory usage under 100MB for typical projects
- Support projects with 1000+ documents
- Token consumption for architecture overview < 500 tokens

SECURITY:
- No arbitrary code execution
- Path traversal protection
- Input sanitization for shell commands
- API key authentication (when API enabled)
- Regular dependency updates via Dependabot
```

---

# 2. Specification Prompt

Use this with speckit to define detailed requirements:

```
/speckit.specify Build loko - a C4 model architecture documentation tool that enables Cloud Solution Architects to design systems conversationally with LLM agents via MCP, while also providing powerful CLI and API interfaces for direct interaction.

USER STORY 1 (Priority: P1) - LLM-Driven Architecture Design
As a Cloud Solution Architect, I want to use an LLM chatbot (Claude, GPT, Gemini) connected to loko via MCP to have a conversational workflow where the LLM guides me through designing architecture, and the end result is well-documented C4 model architecture with diagrams and markdown documentation.

ACCEPTANCE SCENARIOS:
1. Given a new project, when I chat with the LLM and say "I'm building a payment processing system", then the LLM calls loko MCP tools to initialize the project structure
2. Given the LLM asks "what containers do you need?", when I respond "API and Database", then the LLM scaffolds these containers using loko MCP tools
3. Given the conversation progresses, when the LLM generates D2 diagram code, then loko renders it to SVG and embeds it in documentation
4. Given the architecture is defined, when I ask "show me the docs", then loko builds HTML documentation I can view in a browser
5. Given I ask "what's the current architecture?", when the LLM queries loko, then it returns a token-efficient summary without consuming excessive context

USER STORY 2 (Priority: P1) - Direct File Editing Workflow
As a developer, I want to edit .md and .d2 files directly in my text editor (VSCode, Vim), and have loko automatically rebuild documentation in real-time, so I can work in my preferred environment without depending on LLMs.

ACCEPTANCE SCENARIOS:
1. Given loko is running in watch mode, when I edit a .d2 file, then it automatically re-renders the diagram within 500ms
2. Given I save a markdown file, when loko rebuilds, then the HTML output updates and my browser auto-refreshes
3. Given I create a new system folder manually, when I run loko validate, then it reports any missing required files
4. Given I want to preview, when I run loko serve, then I get a local web server showing rendered documentation

USER STORY 3 (Priority: P1) - Project Scaffolding
As a developer, I want to quickly scaffold C4 documentation structure using templates, so I can start with good conventions and consistent structure across systems.

ACCEPTANCE SCENARIOS:
1. Given I run loko init, when I provide project details interactively, then loko creates project structure with loko.toml configuration
2. Given an initialized project, when I run loko new system PaymentService, then it scaffolds system.md and system.d2 from templates
3. Given a system exists, when I run loko new container PaymentService API, then it creates container docs under the system
4. Given I want custom templates, when I place templates in .loko/templates/, then loko uses them instead of global templates

USER STORY 4 (Priority: P2) - API Integration
As a DevOps engineer, I want to trigger loko builds via HTTP API in CI/CD pipelines, so I can automate documentation generation and validation as part of deployment workflows.

ACCEPTANCE SCENARIOS:
1. Given loko API server is running, when I POST to /api/v1/build, then it builds documentation and returns status
2. Given I want to query structure, when I GET /api/v1/systems, then I receive JSON listing all systems
3. Given API auth is enabled, when I call without API key, then I get 401 Unauthorized
4. Given a build completes, when I GET /api/v1/validate, then I receive validation report with any issues

USER STORY 5 (Priority: P2) - Multi-Format Export
As an architect, I want to export documentation to multiple formats (HTML, Markdown, PDF), so I can share architecture with different audiences and use cases.

ACCEPTANCE SCENARIOS:
1. Given documentation exists, when I run loko build --format html, then I get a static website I can deploy
2. Given I need a single file, when I run loko build --format markdown, then I get one README.md with all content
3. Given I want PDFs, when I run loko build --format pdf (and veve-cli is installed), then I get PDF documents
4. Given I want all formats, when I run loko build, then it generates HTML, markdown, and PDF based on loko.toml config

USER STORY 6 (Priority: P1) - Token-Efficient Architecture Queries
As an LLM agent, I want to query architecture with configurable detail levels, so I can get context without consuming excessive tokens.

ACCEPTANCE SCENARIOS:
1. Given I need a quick overview, when I call query_architecture with detail:"summary", then I get ~200 tokens with counts and system names
2. Given I need to understand structure, when I call with detail:"structure", then I get ~500 tokens with systems and their containers
3. Given I need full details on one system, when I call with target:"PaymentService" and detail:"full", then I get complete info for only that system
4. Given I want maximum efficiency, when I call with format:"toon", then I get TOON-encoded response with 30-40% fewer tokens

FUNCTIONAL REQUIREMENTS:

Core Configuration:
- FR-001: System MUST support TOML configuration (loko.toml) with validation
- FR-002: System MUST parse YAML frontmatter in markdown files for metadata
- FR-003: System MUST support both global (~/.loko/templates/) and project (.loko/templates/) templates

Diagram Rendering:
- FR-004: System MUST shell out to d2 CLI for diagram rendering with caching
- FR-005: System MUST support parallel D2 rendering for performance

Template System:
- FR-006: System MUST integrate ason as a Go library for template scaffolding
- FR-007: System MUST include two starter templates: standard-3layer and serverless

MCP Interface:
- FR-008: System MUST provide MCP server with tools: query_project, query_architecture, create_system, create_container, create_component, update_diagram, build_docs, validate
- FR-009: System MUST provide progressive context loading via MCP with summary/structure/full detail levels
- FR-010: System MUST support targeted queries (specific system/container) to avoid loading entire project context
- FR-011: System SHOULD provide compressed notation format for architecture relationships

CLI Interface:
- FR-012: System MUST support CLI commands: init, new, build, serve, watch, render, validate, doctor, mcp, api

HTML Generation:
- FR-013: System MUST generate static HTML site with sidebar navigation, breadcrumbs, search, and hot reload

Logging & Validation:
- FR-014: System MUST log in JSON format with structured fields for production observability
- FR-015: System MUST validate architecture for orphaned references, missing files, and C4 hierarchy violations

Build System:
- FR-016: System MUST support incremental builds (only rebuild changed files)
- FR-017: System MUST provide Docker images for containerized usage

PDF Generation:
- FR-018: System MUST shell out to veve-cli for PDF generation when enabled

Clean Architecture:
- FR-019: System MUST implement Clean Architecture with clear separation:
  - internal/core/entities/ - Domain objects with validation
  - internal/core/usecases/ - Application logic and port interfaces
  - internal/adapters/ - Infrastructure implementations
  - cmd/ - CLI commands (thin wrappers)
  - internal/mcp/ - MCP server (thin wrappers)
  - internal/api/ - HTTP API (thin wrappers)
- FR-020: All use cases MUST be callable from CLI, MCP, and API without code duplication
- FR-021: All external dependencies (d2, file system, veve-cli) MUST be accessed through interfaces

Token Efficiency:
- FR-022: System SHOULD support TOON (Token-Oriented Object Notation) as optional output format for MCP queries
- FR-023: When TOON format is requested, system MUST use official toon-format/toon-go library
- FR-024: MCP tool descriptions MUST include format hints when TOON is used

NON-FUNCTIONAL REQUIREMENTS:

Performance:
- NFR-001: Build 100 diagrams in under 30 seconds on typical hardware (with caching)
- NFR-002: Watch mode rebuild latency under 500ms from file change to completion
- NFR-003: Memory usage under 100MB for projects with up to 50 systems

Compatibility:
- NFR-004: Support Linux, macOS, and Windows with identical behavior
- NFR-005: Single binary with no runtime dependencies except d2 (and optionally veve-cli)
- NFR-006: Graceful degradation if optional dependencies (veve-cli) are missing

User Experience:
- NFR-007: Clear, actionable error messages with suggestions (using lipgloss for formatting)
- NFR-008: Comprehensive test coverage (>80%) with CI running on all platforms

Architecture:
- NFR-009: Core package (internal/core/) MUST have zero external dependencies beyond Go standard library
- NFR-010: Token consumption for architecture overview query MUST be under 500 tokens for projects with up to 20 systems
- NFR-011: Adding a new CLI command or MCP tool MUST require under 50 lines of interface code

KEY ENTITIES:
- Project: Root configuration and metadata (loko.toml)
- System: C4 system level (system.md, system.d2)
- Container: C4 container level (container.md, container.d2)
- Component: C4 component level (component.md, component.d2)
- Template: Reusable scaffolding template (template.toml + .tmpl files)
- Diagram: D2 source file (.d2) and rendered output (SVG/PNG)
- Build: Generated documentation artifacts (HTML, markdown, PDF)

TECHNICAL CONSTRAINTS:
- Go 1.23+
- Clean Architecture: core/ has zero external dependencies
- Cobra for CLI framework (thin wrapper over use cases)
- Viper for configuration (adapter layer only)
- Bubbletea and Lipgloss for TUI/styling (UI layer only)
- fsnotify for file watching (adapter layer)
- gomarkdown for markdown parsing (adapter layer)
- Standard library html/template for HTML generation
- Shell out to d2 binary (behind DiagramRenderer interface)
- Shell out to veve-cli binary (behind PDFRenderer interface)
- Import github.com/madstone-tech/ason as library (behind TemplateEngine interface)
- Import github.com/toon-format/toon-go for TOON encoding (adapter layer)
- MCP protocol via stdio transport
- JSON structured logging
- All external dependencies accessed through interfaces in usecases/ports.go

SUCCESS CRITERIA:
- SC-001: Developer can go from loko init to viewing docs in under 2 minutes
- SC-002: LLM can successfully design a 3-system architecture via MCP without human intervention
- SC-003: Watch mode provides sub-500ms feedback loop during documentation editing
- SC-004: Generated HTML documentation is navigable, searchable, and renders correctly on mobile
- SC-005: Validation catches 90%+ of common architecture documentation mistakes
- SC-006: CI/CD pipelines can integrate loko builds with exit codes for failures
- SC-007: Docker image is under 50MB and includes all required dependencies (d2)
- SC-008: Contributors can add a new CLI command or MCP tool with under 50 lines of code
- SC-009: Architecture overview query consumes <500 tokens for 20-system project
- SC-010: TOON format reduces token consumption by 30%+ compared to JSON
```

---

# 3. README.md

```markdown
# 🪇 loko - Guardian of Architectural Wisdom

> *Transform complexity into clarity with C4 model documentation and LLM integration*

**loko** (Papa Loko) is a modern architecture documentation tool that brings the [C4 model](https://c4model.com/) to life through conversational design with LLMs, powerful CLI workflows, and beautiful documentation generation.

[![Go Version](https://img.shields.io/github/go-mod/go-version/madstone-tech/loko)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/madstone-tech/loko)](https://github.com/madstone-tech/loko/releases)
[![License](https://img.shields.io/github/license/madstone-tech/loko)](LICENSE)
[![Tests](https://github.com/madstone-tech/loko/workflows/test/badge.svg)](https://github.com/madstone-tech/loko/actions)
[![Docker](https://img.shields.io/docker/v/madstonetech/loko?label=docker)](https://github.com/madstone-tech/loko/pkgs/container/loko)

---

## ✨ Features

🤖 **LLM-First Design** - Design architecture conversationally with Claude, GPT, or Gemini via [MCP](https://modelcontextprotocol.io)  
📝 **Direct Editing** - Edit markdown and [D2](https://d2lang.com) diagrams in your favorite editor  
⚡ **Real-Time Feedback** - Watch mode rebuilds in <500ms with hot reload  
🎨 **Beautiful Output** - Generate static sites, PDFs, and markdown documentation  
🔧 **Powerful CLI** - Scaffold, build, validate, and serve - all from the terminal  
🐳 **Docker Ready** - Official images with all dependencies included  
🎯 **Zero Config** - Smart defaults with optional customization via TOML  
💰 **Token Efficient** - Progressive context loading + TOON format minimize LLM costs

---

## 🚀 Quick Start

### Installation

**macOS / Linux (Homebrew)**
```bash
brew tap madstone-tech/tap
brew install loko
```

**Go Install**
```bash
go install github.com/madstone-tech/loko@latest
```

**Docker**
```bash
docker pull ghcr.io/madstone-tech/loko:latest
```

### Your First Architecture (2 minutes)

```bash
# Initialize a new project
loko init my-architecture
cd my-architecture

# Scaffold your first system
loko new system PaymentService

# Edit the generated files
vim src/PaymentService/system.md
vim src/PaymentService/system.d2

# Build and preview
loko serve
# Open http://localhost:8080
```

**That's it!** You now have a live-reloading documentation site.

---

## 🎯 Usage Modes

### 1️⃣ Conversational Design (LLM + MCP)

```bash
# Start MCP server
loko mcp

# In your LLM client (Claude, etc):
# "I'm building a payment processing system with an API and database"
# LLM uses loko to scaffold structure and create diagrams
```

### 2️⃣ Manual Editing (Developer Workflow)

```bash
# Watch for changes
loko watch &

# Edit files in your editor
vim src/PaymentService/system.d2

# Automatically rebuilds and refreshes browser
```

### 3️⃣ CI/CD Integration (API)

```bash
# Start API server
loko api

# Trigger builds via HTTP
curl -X POST http://localhost:8081/api/v1/build
```

---

## 📚 Core Concepts

### C4 Model Hierarchy

```
Context
  └── System
       └── Container
            └── Component
```

### Project Structure

```
my-architecture/
├── loko.toml              # Configuration
├── src/                   # Source documentation
│   ├── context.md
│   ├── context.d2
│   └── SystemName/
│       ├── system.md
│       ├── system.d2
│       └── ContainerName/
│           ├── container.md
│           └── container.d2
└── dist/                  # Generated output
    └── index.html
```

### Clean Architecture

loko follows Clean Architecture principles:

```
loko/
├── cmd/                        # CLI commands (thin layer)
│
├── internal/
│   ├── core/                   # THE HEART - zero external dependencies
│   │   ├── entities/           # Domain objects: Project, System, Container
│   │   ├── usecases/           # Application logic + port interfaces
│   │   └── errors/             # Domain-specific errors
│   │
│   ├── adapters/               # Infrastructure implementations
│   │   ├── config/             # TOML configuration
│   │   ├── filesystem/         # File operations
│   │   ├── d2/                 # Diagram rendering
│   │   ├── encoding/           # JSON + TOON encoders
│   │   └── html/               # Site builder
│   │
│   ├── mcp/                    # MCP server + tools
│   ├── api/                    # HTTP API server
│   └── ui/                     # Terminal UI (lipgloss)
│
├── templates/                  # Starter templates
└── docs/                       # Documentation + ADRs
```

---

## 💰 Token Efficiency

loko is designed to minimize LLM token consumption:

### Progressive Context Loading

```bash
# Quick overview (~200 tokens)
query_architecture --detail summary

# System hierarchy (~500 tokens)  
query_architecture --detail structure

# Full details for one system (variable)
query_architecture --detail full --target PaymentService
```

### TOON Format (Optional)

[TOON](https://toonformat.dev) reduces tokens by 30-40% for structured data:

```bash
# JSON: ~380 tokens
{"systems":[{"name":"PaymentService","containers":["API","DB"]},...]}

# TOON: ~220 tokens
systems[4]{name,containers}:
  PaymentService,API|DB
  OrderService,API|DB
  ...
```

---

## 🎨 Features Deep Dive

### Templates

**Global templates** (`~/.loko/templates/`) and **project templates** (`.loko/templates/`) using [ason](https://github.com/madstone-tech/ason):

```bash
loko new system PaymentService
# Uses template: ~/.loko/templates/c4-system/

# Customize per-project
cp -r ~/.loko/templates/c4-system .loko/templates/custom-system
loko new system --template custom-system MyService
```

### Diagram Rendering

Powered by [D2](https://d2lang.com) with caching:

```d2
# src/System/system.d2
User -> API: Uses
API -> Database: Queries
```

```bash
loko render src/System/system.d2
# Generates: dist/diagrams/system.svg
```

### Multi-Format Export

```bash
loko build --format html       # Static website
loko build --format markdown   # Single README.md
loko build --format pdf        # PDF documents (requires veve-cli)
```

### Validation

```bash
loko validate
# Checks for:
# - Orphaned references
# - Missing required files
# - C4 hierarchy violations
# - Broken diagram syntax
```

---

## 🛠️ Configuration

**loko.toml** (TOML format):

```toml
[project]
name = "my-architecture"
description = "System architecture documentation"

[paths]
source = "./src"
output = "./dist"

[d2]
theme = "neutral-default"
layout = "elk"
cache = true

[outputs.html]
enabled = true
navigation = "sidebar"
search = true

[build]
parallel = true
max_workers = 4
```

See [docs/configuration.md](docs/configuration.md) for all options.

---

## 🤝 MCP Integration

loko exposes these tools for LLM interaction:

| Tool | Description |
|------|-------------|
| `query_project` | Get project metadata |
| `query_architecture` | Token-efficient architecture queries |
| `create_system` | Scaffold new system |
| `create_container` | Scaffold container |
| `create_component` | Scaffold component |
| `update_diagram` | Write D2 code to file |
| `build_docs` | Build documentation |
| `validate` | Check architecture consistency |

### Token-Efficient Queries

```json
{
  "name": "query_architecture",
  "parameters": {
    "scope": "project | system | container",
    "target": "specific entity name",
    "detail": "summary | structure | full",
    "format": "json | toon",
    "include_diagrams": false
  }
}
```

---

## 📖 Documentation

- [Installation Guide](docs/installation.md)
- [Quick Start Tutorial](docs/quickstart.md)
- [Configuration Reference](docs/configuration.md)
- [Template System](docs/templates.md)
- [MCP Integration](docs/mcp-integration.md)
- [API Reference](docs/api-reference.md)
- [Architecture Decision Records](docs/adr/)

---

## 🌟 Examples

Check out [examples/](examples/) for complete projects:

- **[simple-project](examples/simple-project/)** - Minimal C4 documentation
- **[3layer-app](examples/3layer-app/)** - Standard web → API → database
- **[serverless](examples/serverless/)** - AWS Lambda architecture

---

## 🐳 Docker

```bash
# Initialize project
docker run -v $(pwd):/workspace ghcr.io/madstone-tech/loko init my-arch

# Build documentation
docker run -v $(pwd):/workspace ghcr.io/madstone-tech/loko build

# Serve with hot reload
docker run -v $(pwd):/workspace -p 8080:8080 ghcr.io/madstone-tech/loko serve
```

---

## 🔧 Development

### Prerequisites

- Go 1.23+
- [d2](https://d2lang.com) (required)
- [veve-cli](https://github.com/madstone-tech/veve-cli) (optional, for PDF)

### Setup

```bash
git clone https://github.com/madstone-tech/loko
cd loko
go mod download
make test
go run main.go --help
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

---

## 🗺️ Roadmap

### v0.1.0 (MVP) - Q1 2025
- ✅ CLI commands (init, new, build, serve, watch)
- ✅ MCP server with token-efficient queries
- ✅ D2 diagram rendering with caching
- ✅ HTML site generation
- ✅ Template system (ason integration)
- ✅ Clean Architecture implementation

### v0.2.0 - Q2 2025
- 🚧 HTTP API server
- 🚧 TOON format support for MCP
- 🚧 PDF export via veve-cli
- 🚧 Advanced validation rules

### v0.3.0 - Q3 2025
- 📋 Architecture graph visualization
- 📋 Diff and changelog generation
- 📋 Plugin system

See [ROADMAP.md](ROADMAP.md) for detailed feature plans.

---

## 🤲 Contributing

We welcome contributions! loko is **building in public** - see our [development progress](https://github.com/madstone-tech/loko/issues).

- 🐛 **Bug reports** → [Open an issue](https://github.com/madstone-tech/loko/issues/new?template=bug_report.md)
- 💡 **Feature requests** → [Start a discussion](https://github.com/madstone-tech/loko/discussions/new?category=ideas)
- 🔧 **Pull requests** → See [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📜 License

[MIT License](LICENSE) - Copyright (c) 2025 MADSTONE TECHNOLOGY

---

## 🙏 Acknowledgments

**loko** builds on excellent open-source tools:

- [D2](https://d2lang.com) - Declarative diagramming
- [ason](https://github.com/madstone-tech/ason) - Template scaffolding
- [TOON](https://toonformat.dev) - Token-efficient notation
- [C4 Model](https://c4model.com) - Architecture visualization approach
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework

---

## 🪇 Why "loko"?

**Papa Loko** is the lwa (spirit) in Haitian Vodou who guards sacred knowledge, maintains tradition, and passes down wisdom to initiates. As the first houngan (priest), he is the keeper of the ritual knowledge that connects the physical and spiritual worlds.

Like Papa Loko, this tool acts as the guardian of your architectural wisdom - organizing knowledge, maintaining documentation traditions, and making complex systems understandable.

*"Papa Loko, you're the wind, pushing us, and we become butterflies."* 🦋

---

<p align="center">
  <strong>Made with ❤️ by <a href="https://github.com/madstone-tech">MADSTONE TECHNOLOGY</a></strong><br>
  Building in public • Join us on the journey
</p>
```

---

# 4. CONTRIBUTING.md

```markdown
# Contributing to loko

Thank you for your interest in contributing to loko! We're building this tool in public and welcome contributions from developers of all experience levels.

## 🎯 Ways to Contribute

- 🐛 **Report bugs** - Help us find and fix issues
- 💡 **Suggest features** - Share ideas for improvements
- 📖 **Improve documentation** - Clarify, expand, or fix docs
- 🔧 **Submit code** - Bug fixes, features, tests
- 🎨 **Design templates** - Create C4 templates for common patterns
- 🧪 **Test and validate** - Try loko on real projects and report findings

## 🚀 Getting Started

### Prerequisites

- **Go 1.23+** ([install](https://go.dev/doc/install))
- **d2** ([install](https://d2lang.com))
- **git**
- Optional: **veve-cli** (for PDF tests)

### Development Setup

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/loko
cd loko

# 2. Install dependencies
go mod download

# 3. Install d2
brew install d2  # macOS
# or download from https://github.com/terrastruct/d2/releases

# 4. Run tests
go test ./...

# 5. Build
go build -o loko .

# 6. Try it out
./loko --help
```

## 🏗️ Architecture Guide

loko uses **Clean Architecture**. Understanding this will help you contribute effectively.

### The Dependency Rule

Dependencies point **inward**. Inner layers never know about outer layers.

```
┌─────────────────────────────────────────────────┐
│           Interfaces (CLI, MCP, API)            │  ← Thin wrappers
├─────────────────────────────────────────────────┤
│         Adapters (d2, filesystem, toon)         │  ← Implements Ports
├─────────────────────────────────────────────────┤
│        Use Cases (CreateSystem, Build)          │  ← Defines Ports
├─────────────────────────────────────────────────┤
│      Entities (Project, System, Container)      │  ← Pure Go
└─────────────────────────────────────────────────┘
```

### Project Structure

```
loko/
├── cmd/                      # CLI commands (thin wrappers)
├── internal/
│   ├── core/                 # THE HEART - zero external deps
│   │   ├── entities/         # Domain objects
│   │   ├── usecases/         # Application logic + ports
│   │   └── errors/           # Domain errors
│   ├── adapters/             # Infrastructure
│   │   ├── config/           # TOML loader
│   │   ├── filesystem/       # File operations
│   │   ├── d2/               # Diagram renderer
│   │   ├── encoding/         # JSON + TOON
│   │   └── html/             # Site builder
│   ├── mcp/                  # MCP server
│   ├── api/                  # HTTP API
│   └── ui/                   # Terminal UI
├── templates/                # Starter templates
└── docs/                     # Documentation
```

### Where to Add Code

| I want to... | Where to add it |
|--------------|-----------------|
| Add a new entity field | `internal/core/entities/` |
| Add validation logic | `internal/core/entities/` (on the entity) |
| Add a new operation | `internal/core/usecases/` (new use case) |
| Add a CLI command | `cmd/` (thin wrapper calling use case) |
| Add an MCP tool | `internal/mcp/tools/` (thin wrapper) |
| Add an API endpoint | `internal/api/handlers/` (thin wrapper) |
| Change how files are stored | `internal/adapters/filesystem/` |
| Change diagram rendering | `internal/adapters/d2/` |
| Add output format | `internal/adapters/encoding/` |

### Adding a New Use Case

1. Define input/output structs in `internal/core/usecases/your_usecase.go`
2. If you need new infrastructure, add interface to `internal/core/usecases/ports.go`
3. Implement the use case
4. Add adapter implementation if needed in `internal/adapters/`
5. Wire it up in `main.go`
6. Add thin wrappers in `cmd/`, `internal/mcp/tools/`, `internal/api/handlers/`

### Example: Adding "Archive System" Feature

```go
// 1. internal/core/usecases/archive_system.go
type ArchiveSystemInput struct {
    SystemName string
}

type ArchiveSystemOutput struct {
    ArchivedAt time.Time
    BackupPath string
}

type ArchiveSystemUseCase struct {
    projects ProjectRepository
    archiver Archiver           // New port
}

func (uc *ArchiveSystemUseCase) Execute(ctx context.Context, input ArchiveSystemInput) (*ArchiveSystemOutput, error) {
    // Business logic here
}
```

```go
// 2. internal/core/usecases/ports.go (add new port)
type Archiver interface {
    Archive(ctx context.Context, path string) (string, error)
}
```

```go
// 3. internal/adapters/filesystem/archiver.go
type ZipArchiver struct{}

func (a *ZipArchiver) Archive(ctx context.Context, path string) (string, error) {
    // Implementation
}
```

```go
// 4. cmd/archive.go (thin CLI wrapper - under 50 lines!)
func archiveCmd(uc *usecases.ArchiveSystemUseCase) *cobra.Command {
    return &cobra.Command{
        Use: "archive [system]",
        RunE: func(cmd *cobra.Command, args []string) error {
            output, err := uc.Execute(ctx, usecases.ArchiveSystemInput{
                SystemName: args[0],
            })
            // Format and display output
        },
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests (Use Cases)

Mock the ports to test business logic in isolation:

```go
func TestArchiveSystemUseCase(t *testing.T) {
    mockRepo := &MockProjectRepo{...}
    mockArchiver := &MockArchiver{...}
    
    uc := usecases.NewArchiveSystemUseCase(mockRepo, mockArchiver)
    
    output, err := uc.Execute(ctx, usecases.ArchiveSystemInput{
        SystemName: "PaymentService",
    })
    
    assert.NoError(t, err)
    assert.True(t, mockArchiver.ArchiveCalled)
}
```

### Integration Tests

Use real adapters with temp directories:

```go
func TestArchiveSystemIntegration(t *testing.T) {
    tmpDir := t.TempDir()
    // Set up real file system
    // Use real adapters
    // Verify actual files created
}
```

### Golden Tests

For output formatting:

```go
func TestBuildHTMLGolden(t *testing.T) {
    got := builder.Build(project)
    golden.Assert(t, got, "testdata/expected.html")
}
```

## 🔧 Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Changes

- Follow Go best practices ([Effective Go](https://go.dev/doc/effective_go))
- Write tests for new functionality
- Update documentation as needed
- Run `go fmt` before committing

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/core/usecases/...

# Run integration tests
go test -tags=integration ./tests/integration/...
```

### 4. Commit

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
feat: add MCP tool for component creation
fix: resolve d2 caching issue on Windows
docs: update installation guide for Homebrew
test: add integration tests for watch mode
chore: update dependencies
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request with:
- Clear title and description
- Reference any related issues (#123)
- Screenshots/demos if applicable

## 📝 Code Style

### General Principles

- **Simplicity** - Prefer clear code over clever code
- **Interfaces** - Use interfaces for dependencies
- **Error handling** - Always handle errors with context
- **Documentation** - Public APIs must have godoc comments

### Example: Good Error Handling

```go
func (e *Engine) RenderDiagram(d2File string) error {
    if !strings.HasSuffix(d2File, ".d2") {
        return &errors.ValidationError{
            Path:    d2File,
            Message: "file must have .d2 extension",
        }
    }
    
    if err := e.renderer.Render(d2File); err != nil {
        return fmt.Errorf("render diagram %s: %w", d2File, err)
    }
    
    return nil
}
```

### Interface Design

```go
// Good - testable and swappable
type DiagramRenderer interface {
    Render(ctx context.Context, opts RenderOptions) (*RenderResult, error)
    Available() bool
}

type Engine struct {
    renderer DiagramRenderer  // Can mock in tests
}
```

## 🐛 Reporting Bugs

Include:
- loko version (`loko --version`)
- Operating system and version
- Go version (`go version`)
- d2 version (`d2 --version`)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (run with `--debug`)

## 💡 Suggesting Features

Consider:
- Does it align with loko's core mission?
- Is it simple to use?
- Can it be composed with existing features?
- Would it benefit most users or is it niche?

## 📦 Adding Dependencies

We minimize dependencies. Before adding a new one:

1. **Check if stdlib can do it** - Go's standard library is excellent
2. **Evaluate maintenance** - Is it actively maintained?
3. **Check size** - Will it bloat the binary?
4. **Discuss first** - Open an issue to discuss necessity

## 🏗️ Architecture Decisions

Major decisions are documented in [ADRs](docs/adr/):

- [ADR 0001: Clean Architecture](docs/adr/0001-clean-architecture.md)
- [ADR 0002: Token-Efficient MCP](docs/adr/0002-token-efficient-mcp.md)
- [ADR 0003: TOON Format Support](docs/adr/0003-toon-format.md)

Discuss in issues before implementing major changes.

## 🎯 Pull Request Checklist

Before submitting:

- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Documentation updated (if needed)
- [ ] Changelog updated (CHANGELOG.md)
- [ ] Commit messages follow convention
- [ ] PR description is clear and complete
- [ ] Interface code is under 50 lines (for new commands/tools)

## 🌟 Recognition

Contributors are recognized in:
- CHANGELOG.md (for each release)
- README.md (top contributors)
- GitHub release notes

## ❓ Questions?

- **General questions** → [GitHub Discussions](https://github.com/madstone-tech/loko/discussions)
- **Bug reports** → [GitHub Issues](https://github.com/madstone-tech/loko/issues)
- **Security issues** → Email security@madstone.tech

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

---

**Thank you for contributing to loko!** 🪇

Every contribution, no matter how small, helps make architecture documentation better for everyone.
```

---

# 5. CODE_OF_CONDUCT.md

```markdown
# Contributor Covenant Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our
community a harassment-free experience for everyone, regardless of age, body
size, visible or invisible disability, ethnicity, sex characteristics, gender
identity and expression, level of experience, education, socio-economic status,
nationality, personal appearance, race, religion, or sexual identity
and orientation.

We pledge to act and interact in ways that contribute to an open, welcoming,
diverse, inclusive, and healthy community.

## Our Standards

Examples of behavior that contributes to a positive environment for our
community include:

* Demonstrating empathy and kindness toward other people
* Being respectful of differing opinions, viewpoints, and experiences
* Giving and gracefully accepting constructive feedback
* Accepting responsibility and apologizing to those affected by our mistakes,
  and learning from the experience
* Focusing on what is best not just for us as individuals, but for the
  overall community

Examples of unacceptable behavior include:

* The use of sexualized language or imagery, and sexual attention or
  advances of any kind
* Trolling, insulting or derogatory comments, and personal or political attacks
* Public or private harassment
* Publishing others' private information, such as a physical or email
  address, without their explicit permission
* Other conduct which could reasonably be considered inappropriate in a
  professional setting

## Enforcement Responsibilities

Project maintainers are responsible for clarifying and enforcing our standards of
acceptable behavior and will take appropriate and fair corrective action in
response to any behavior that they deem inappropriate, threatening, offensive,
or harmful.

Project maintainers have the right and responsibility to remove, edit, or reject
comments, commits, code, wiki edits, issues, and other contributions that are
not aligned to this Code of Conduct, and will communicate reasons for moderation
decisions when appropriate.

## Scope

This Code of Conduct applies within all community spaces, and also applies when
an individual is officially representing the community in public spaces.
Examples of representing our community include using an official e-mail address,
posting via an official social media account, or acting as an appointed
representative at an online or offline event.

## Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported to the project maintainers at conduct@madstone.tech.
All complaints will be reviewed and investigated promptly and fairly.

All project maintainers are obligated to respect the privacy and security of the
reporter of any incident.

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage],
version 2.0, available at
https://www.contributor-covenant.org/version/2/0/code_of_conduct.html.

[homepage]: https://www.contributor-covenant.org
```

---

# 6. ROADMAP.md

```markdown
# loko Roadmap

## Vision

Make C4 architecture documentation delightful through conversational design with LLMs, powerful developer tools, and beautiful output — while minimizing token costs.

---

## v0.1.0 - MVP (Target: Q1 2025)

**Theme:** Core functionality with Clean Architecture foundation

### Features
- ✅ Project initialization (`loko init`)
- ✅ Template scaffolding (`loko new system/container/component`)
- ✅ D2 diagram rendering with caching
- ✅ Watch mode with hot reload (`loko watch`, `loko serve`)
- ✅ HTML site generation (sidebar nav, breadcrumbs, search)
- ✅ MCP server with token-efficient queries
  - `query_architecture` with summary/structure/full detail levels
  - Progressive context loading
- ✅ TOML configuration with validation
- ✅ Frontmatter support in markdown
- ✅ Global and project templates
- ✅ Docker images
- ✅ Two starter templates (3-layer, serverless)
- ✅ Clean Architecture implementation

### Architecture
- Clean separation: entities → use cases → adapters → interfaces
- All use cases testable with mocked ports
- CLI, MCP share same business logic

### Non-Goals for v0.1.0
- ❌ HTTP API (deferred to v0.2.0)
- ❌ TOON format (deferred to v0.2.0)
- ❌ PDF generation (deferred to v0.2.0)

---

## v0.2.0 - Integration & Optimization (Target: Q2 2025)

**Theme:** API integration, TOON format, multiple output formats

### Features
- 🚧 HTTP API server
  - REST endpoints for systems, containers, diagrams
  - API key authentication
  - CORS support
- 🚧 TOON format support for MCP
  - `format: "toon"` parameter
  - 30-40% additional token reduction
  - Official toon-go library integration
- 🚧 PDF export via veve-cli
  - Single PDF (all docs)
  - Per-system PDFs
- 🚧 Enhanced export formats
  - Confluence export
  - Markdown with different navigation styles
- 🚧 CI/CD integrations
  - GitHub Actions example
  - GitLab CI example
  - Exit codes for validation failures

### Token Efficiency Benchmark
- Summary query: <300 tokens (20-system project)
- Structure query: <500 tokens (20-system project)
- TOON format: 30-40% reduction vs JSON

---

## v0.3.0 - Advanced Features (Target: Q3 2025)

**Theme:** Intelligence and visualization

### Features
- 📋 Architecture graph analysis
  - Parse D2 diagrams to build dependency graph
  - Visualize relationships (interactive HTML, DOT output)
- 📋 Advanced validation
  - Detect circular dependencies
  - Validate naming conventions
  - Check for missing documentation
- 📋 Diff and changelog
  - Compare architecture across branches/commits
  - Generate visual diffs for diagrams
  - Automatic changelog generation
- 📋 Search improvements
  - Full-text search across all docs
  - Semantic search (if LLM available)

---

## v1.0.0 - Stable Release (Target: Q4 2025)

**Theme:** Production-ready, stable API, comprehensive documentation

### Features
- 📋 API stability guarantees
- 📋 Comprehensive user documentation
- 📋 Performance optimizations
- 📋 Plugin system (maybe)
- 📋 Custom HTML themes
- 📋 Import from other formats (PlantUML, Structurizr)

---

## Future (Post v1.0)

### Potential Features (Community Driven)

**Collaboration**
- Multi-user editing via operational transforms
- Comments and annotations
- Approval workflows

**Cloud Features**
- Cloud storage (S3, GCS) for large diagrams
- Hosted documentation service
- Team management

**Enhanced Visualization**
- 3D architecture visualization
- Interactive diagram exploration
- Time-based architecture evolution views

**Integrations**
- Confluence plugin
- Notion export
- Slack notifications for changes
- Jira integration for tracking

**AI Enhancements**
- AI-suggested architecture improvements
- Automatic diagram generation from code
- Natural language queries

---

## How to Contribute

See an issue you'd like to work on? Check out:
- [Good first issues](https://github.com/madstone-tech/loko/labels/good%20first%20issue)
- [Help wanted](https://github.com/madstone-tech/loko/labels/help%20wanted)

Want to suggest a feature?
- [Start a discussion](https://github.com/madstone-tech/loko/discussions/new?category=ideas)

---

**Last Updated:** 2025-01-13
**Maintainers:** @yourusername
```

---

# 7. Architecture Decision Records

## docs/adr/0001-clean-architecture.md

```markdown
# ADR 0001: Clean Architecture

## Status

Accepted

## Context

loko exposes functionality through three interfaces:
1. CLI for developers
2. MCP for LLM agents
3. HTTP API for CI/CD integration

Without careful architecture, we risk:
- Duplicating business logic across interfaces
- Tight coupling to infrastructure (file system, d2 binary)
- Difficulty testing without real external dependencies
- Painful changes when swapping components

## Decision

We adopt Clean Architecture with the following structure:

```
internal/
├── core/           # Zero external dependencies
│   ├── entities/   # Domain objects
│   ├── usecases/   # Application logic + ports
│   └── errors/     # Domain errors
├── adapters/       # Infrastructure implementations
├── mcp/            # MCP interface
├── api/            # HTTP interface
└── ui/             # CLI formatting
```

**Key principles:**
1. Core defines interfaces (ports); adapters implement them
2. Use cases contain all business logic
3. CLI, MCP, API are thin wrappers calling use cases (<50 lines each)
4. Dependencies injected at startup in main.go

## Consequences

**Positive:**
- Single implementation of business logic
- Easy to test core without mocking file system
- Can swap d2 for another renderer by changing one adapter
- Clear guidance for where to add new code
- Contributors can add commands/tools with minimal code

**Negative:**
- More files and indirection
- Slightly more boilerplate for simple operations
- Learning curve for contributors unfamiliar with the pattern

**Mitigations:**
- Document the pattern clearly in CONTRIBUTING.md
- Provide examples for common tasks
- Keep adapters thin — don't over-abstract
```

## docs/adr/0002-token-efficient-mcp.md

```markdown
# ADR 0002: Token-Efficient MCP Queries

## Status

Accepted

## Context

When an LLM agent designs architecture with loko via MCP, it needs context about the existing project. For large projects (30+ systems), sending everything consumes:
- Excessive tokens (cost)
- Context window space (limiting conversation)
- Processing time (latency)

## Decision

Implement progressive context loading with three detail levels:

### Summary (~200 tokens for 20-system project)
```json
{
  "project": "payment-platform",
  "systems": 4,
  "containers": 12,
  "systems_list": ["PaymentService", "OrderService", ...]
}
```

### Structure (~500 tokens)
```json
{
  "systems": {
    "PaymentService": {
      "containers": ["API", "Database", "Worker"],
      "external_dependencies": ["StripeAPI"]
    }
  }
}
```

### Full (targeted, variable)
Complete details for a specific system or container, optionally including D2 diagram source.

### API Design

```
query_architecture(
  scope: "project" | "system" | "container",
  target: string,           // For specific entity
  detail: "summary" | "structure" | "full",
  include_diagrams: bool    // D2 source code
)
```

## Consequences

**Positive:**
- 10x reduction in token usage for typical queries
- LLM can progressively drill down as needed
- Faster response times
- More room in context window for conversation

**Negative:**
- More complex MCP tool implementation
- LLM must learn to use detail levels effectively
- Multiple round trips for deep exploration

**Mitigations:**
- Clear tool description with examples
- Default to "summary" — always fast
- Compressed notation option for power users
```

## docs/adr/0003-toon-format.md

```markdown
# ADR 0003: TOON Format Support

## Status

Accepted (Implementation: v0.2.0)

## Context

loko's MCP interface sends architecture data to LLMs. Token consumption directly impacts cost and context window usage. We already implement progressive context loading (ADR 0002), but can optimize further.

TOON (Token-Oriented Object Notation) is a compact format designed for LLM input that achieves 30-60% token reduction for uniform arrays — exactly the structure of architecture data.

### Example: JSON vs TOON

**JSON (~380 tokens)**
```json
{
  "systems": [
    {"name": "PaymentService", "containers": ["API", "DB"]},
    {"name": "OrderService", "containers": ["API", "DB"]}
  ]
}
```

**TOON (~220 tokens)**
```
systems[2]{name,containers}:
  PaymentService,API|DB
  OrderService,API|DB
```

## Decision

Support TOON as an optional output format for MCP queries:

1. Add `format: "json" | "toon"` parameter to `query_architecture` tool
2. Default to JSON for maximum compatibility
3. Use official `toon-format/toon-go` library
4. Include format hint in response when TOON is used

### Implementation

```go
// New port
type OutputEncoder interface {
    Encode(data any) ([]byte, error)
    ContentType() string
    FormatHint() string
}

// Adapters
- internal/adapters/encoding/json_encoder.go  (default)
- internal/adapters/encoding/toon_encoder.go  (optional)
```

## Consequences

**Positive:**
- Additional 30-40% token reduction on top of progressive loading
- Official Go library available and maintained
- Aligns with token-efficiency design philosophy
- Compound effect: progressive loading + TOON = 60-85% total reduction

**Negative:**
- Additional dependency
- Not all LLMs familiar with TOON format
- Requires format hint in tool description

**Mitigations:**
- Make TOON opt-in, not default
- Provide clear format hints in responses
- Document when to use each format
- Benchmark and report real-world savings
```

## docs/adr/template.md

```markdown
# ADR NNNN: Title

## Status

[Proposed | Accepted | Deprecated | Superseded]

## Context

What is the issue that we're seeing that is motivating this decision or change?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

What becomes easier or more difficult to do because of this change?
```

---

# 8. Initial GitHub Issues

## Phase 1: Foundation (Week 1)

### Issue #1: Project Setup and CI

```markdown
Title: Initialize project with Clean Architecture structure
Labels: enhancement, v0.1.0, priority:high

## Description
Set up the project with Clean Architecture directory structure and basic CI.

## Tasks
- [ ] Initialize Go module (`github.com/madstone-tech/loko`)
- [ ] Create directory structure:
  ```
  cmd/
  internal/core/entities/
  internal/core/usecases/
  internal/core/errors/
  internal/adapters/
  internal/mcp/
  internal/api/
  internal/ui/
  templates/
  docs/adr/
  examples/
  tests/
  ```
- [ ] Set up GitHub Actions for test/lint/build
- [ ] Add Makefile with common commands
- [ ] Configure golangci-lint
- [ ] Add goreleaser config (for later)
- [ ] Create ADR directory with template

## Acceptance Criteria
- `go build` succeeds
- `go test ./...` runs (even if no tests yet)
- CI passes on PR
- Directory structure matches Clean Architecture
```

### Issue #2: Core Entities

```markdown
Title: Implement core domain entities
Labels: enhancement, v0.1.0, priority:high

## Description
Create the domain entities in `internal/core/entities/` with validation.

## Tasks
- [ ] `project.go` - Project entity with systems collection
- [ ] `system.go` - System entity with containers collection
- [ ] `container.go` - Container entity with components
- [ ] `component.go` - Component entity
- [ ] `diagram.go` - Diagram entity (source + rendered)
- [ ] `template.go` - Template entity
- [ ] `validation.go` - Validation logic for all entities
- [ ] `errors.go` - Domain error types

## Acceptance Criteria
- All entities have constructors (NewSystem, etc.)
- Validation returns structured errors
- 100% test coverage for validation logic
- Zero external dependencies in this package (`go list -m` shows only stdlib)
```

### Issue #3: Use Case Ports (Interfaces)

```markdown
Title: Define use case ports (interfaces)
Labels: enhancement, v0.1.0, priority:high

## Description
Define all interfaces in `internal/core/usecases/ports.go` that adapters must implement.

## Tasks
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

## Acceptance Criteria
- All interfaces documented with godoc
- Input/output types defined
- No implementation details leak into interfaces
- Interfaces are minimal (don't over-abstract)
```

## Phase 2: First Use Case End-to-End (Week 2)

### Issue #4: CreateSystem Use Case

```markdown
Title: Implement CreateSystem use case
Labels: enhancement, v0.1.0, priority:high

## Description
First complete use case: creating a new C4 system.

## Tasks
- [ ] `internal/core/usecases/create_system.go`
- [ ] Define `CreateSystemInput` and `CreateSystemOutput`
- [ ] Input validation
- [ ] Duplicate detection
- [ ] Template loading and rendering
- [ ] Project saving
- [ ] Unit tests with mocked ports

## Acceptance Criteria
- Use case works with mock repositories
- Returns structured output (System + files created)
- Proper error handling with domain errors
- >90% test coverage
```

### Issue #5: File System Adapter

```markdown
Title: Implement file system adapter
Labels: enhancement, v0.1.0, priority:high

## Description
Implement `ProjectRepository` using the file system.

## Tasks
- [ ] `internal/adapters/filesystem/project_repo.go`
- [ ] Load project from loko.toml + directory scan
- [ ] Save project (create directories, write files)
- [ ] Parse YAML frontmatter from markdown
- [ ] Handle missing directories gracefully
- [ ] `internal/adapters/config/loader.go` - TOML config loading

## Acceptance Criteria
- Implements `usecases.ProjectRepository` interface
- Integration tests with temp directories
- Handles edge cases (missing files, permissions)
```

### Issue #6: CLI Wiring (Basic)

```markdown
Title: Wire up basic CLI with dependency injection
Labels: enhancement, v0.1.0, priority:high

## Description
Create main.go and basic CLI commands.

## Tasks
- [ ] `main.go` with dependency injection (wire adapters → use cases)
- [ ] `cmd/root.go` - Root command with global flags
- [ ] `cmd/init.go` - `loko init` command
- [ ] `cmd/new.go` - `loko new system` command
- [ ] `internal/ui/styles.go` - Lipgloss styles
- [ ] `internal/ui/output.go` - Success/error formatting

## Acceptance Criteria
- `loko init myproject` creates project structure
- `loko new system PaymentService` creates system files
- Commands are thin (<50 lines each)
- Output is nicely formatted with lipgloss
```

## Phase 3: Build Pipeline (Week 3)

### Issue #7: D2 Renderer Adapter

```markdown
Title: Implement D2 diagram renderer adapter
Labels: enhancement, v0.1.0, priority:high

## Description
Implement `DiagramRenderer` using the d2 CLI.

## Tasks
- [ ] `internal/adapters/d2/renderer.go`
- [ ] Shell out to d2 binary
- [ ] Content-based caching (hash → output path)
- [ ] Configurable theme/layout
- [ ] Parallel rendering support
- [ ] Graceful error handling

## Acceptance Criteria
- Implements `usecases.DiagramRenderer` interface
- Cache hit returns immediately without calling d2
- Clear error messages when d2 missing
- Supports SVG and PNG output
```

### Issue #8: BuildDocs Use Case

```markdown
Title: Implement BuildDocs use case
Labels: enhancement, v0.1.0, priority:high

## Description
Use case for building documentation output.

## Tasks
- [ ] `internal/core/usecases/build_docs.go`
- [ ] Support multiple formats (HTML, markdown)
- [ ] Parallel diagram rendering
- [ ] Incremental builds (only changed files)
- [ ] Progress reporting

## Acceptance Criteria
- Builds HTML site from project
- Renders all diagrams (with caching)
- Reports progress via ProgressReporter
- Returns build statistics (files generated, cache hits)
```

### Issue #9: HTML Site Builder

```markdown
Title: Implement HTML site builder adapter
Labels: enhancement, v0.1.0, priority:high

## Description
Generate static HTML documentation site.

## Tasks
- [ ] `internal/adapters/html/builder.go`
- [ ] `internal/adapters/html/templates/` - HTML templates
- [ ] Sidebar navigation
- [ ] Breadcrumbs
- [ ] Search (client-side)
- [ ] Responsive design
- [ ] Hot reload support (WebSocket)

## Acceptance Criteria
- Generates complete static site
- Navigation works correctly
- Mobile-friendly
- Search finds content
```

### Issue #10: Build and Serve Commands

```markdown
Title: Add build, serve, watch CLI commands
Labels: enhancement, v0.1.0, priority:high

## Description
Complete the build pipeline CLI commands.

## Tasks
- [ ] `cmd/build.go` - `loko build` command
- [ ] `cmd/serve.go` - `loko serve` command with local server
- [ ] `cmd/watch.go` - `loko watch` command with file watching
- [ ] `cmd/render.go` - `loko render` for single diagrams
- [ ] `cmd/validate.go` - `loko validate` command

## Acceptance Criteria
- `loko build` generates dist/ directory
- `loko serve` starts server at localhost:8080
- `loko watch` rebuilds on file changes (<500ms)
- All commands under 50 lines
```

## Phase 4: MCP (Week 4)

### Issue #11: QueryArchitecture Use Case

```markdown
Title: Implement token-efficient architecture queries
Labels: enhancement, v0.1.0, mcp, priority:high

## Description
Implement progressive context loading for MCP.

## Tasks
- [ ] `internal/core/usecases/query_architecture.go`
- [ ] Summary level (~200 tokens)
- [ ] Structure level (~500 tokens)
- [ ] Full level (targeted)
- [ ] Compressed notation output option
- [ ] Unit tests

## Acceptance Criteria
- Summary for 20-system project < 300 tokens
- Structure for 20-system project < 600 tokens
- Full returns only requested entity
- Compressed notation parseable by LLMs

## References
- ADR 0002: Token-Efficient MCP Queries
```

### Issue #12: MCP Server

```markdown
Title: Implement MCP server with core tools
Labels: enhancement, v0.1.0, mcp, priority:high

## Description
Create MCP server that exposes use cases as tools.

## Tasks
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

## Acceptance Criteria
- Tools call same use cases as CLI
- Tool handlers < 30 lines each
- JSON schemas for all tool inputs
- Works with Claude Desktop
```

### Issue #13: Documentation and Examples

```markdown
Title: Create documentation and working examples
Labels: documentation, v0.1.0, priority:medium

## Description
Write user documentation and create example projects.

## Tasks
- [ ] `docs/quickstart.md` - 5-minute tutorial
- [ ] `docs/configuration.md` - loko.toml reference
- [ ] `docs/architecture.md` - Clean Architecture explanation
- [ ] `docs/mcp-integration.md` - LLM setup guide
- [ ] `examples/simple-project/` - Minimal example
- [ ] `examples/3layer-app/` - Web/API/DB example
- [ ] CI job that builds examples

## Acceptance Criteria
- Quickstart completable in <5 minutes
- Examples build without errors in CI
- MCP integration tested with Claude
```

## Phase 5: v0.2.0 Features

### Issue #14: TOON Format Support

```markdown
Title: Add TOON format support for MCP responses
Labels: enhancement, v0.2.0, mcp, optimization

## Description
Implement TOON as optional output format for architecture queries.

## Tasks
- [ ] Add `toon-format/toon-go` dependency
- [ ] Create `OutputEncoder` interface in `ports.go`
- [ ] `internal/adapters/encoding/json_encoder.go` (default)
- [ ] `internal/adapters/encoding/toon_encoder.go`
- [ ] Add `format` parameter to `QueryArchitectureInput`
- [ ] Update MCP tool schema and handler
- [ ] Add format hint to TOON responses
- [ ] Benchmark token usage: JSON vs TOON
- [ ] Document usage in MCP integration guide

## Acceptance Criteria
- `query_architecture` accepts `format: "toon"` parameter
- TOON output is valid per toon-format spec
- Token reduction of 30%+ verified
- Format hint included in TOON responses

## References
- https://toonformat.dev/
- https://github.com/toon-format/toon-go
- ADR 0003: TOON Format Support
```

### Issue #15: HTTP API Server

```markdown
Title: Implement HTTP API server
Labels: enhancement, v0.2.0, priority:medium

## Description
Add HTTP API for CI/CD integration.

## Tasks
- [ ] `internal/api/server.go` - HTTP server setup
- [ ] `internal/api/middleware/auth.go` - API key auth
- [ ] `internal/api/middleware/logging.go`
- [ ] `internal/api/handlers/systems.go`
- [ ] `internal/api/handlers/build.go`
- [ ] `internal/api/handlers/validate.go`
- [ ] `internal/api/routes.go`
- [ ] `cmd/api.go` - `loko api` command

## Acceptance Criteria
- REST endpoints work correctly
- API key authentication
- Handlers call same use cases as CLI/MCP
- OpenAPI documentation generated
```

---

# 9. Project Structure

Final directory structure after v0.1.0:

```
loko/
├── main.go                           # Entry point, dependency injection
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── .goreleaser.yaml
├── .github/
│   └── workflows/
│       ├── test.yaml
│       ├── lint.yaml
│       └── release.yaml
│
├── cmd/                              # CLI commands (thin)
│   ├── root.go
│   ├── init.go
│   ├── new.go
│   ├── build.go
│   ├── serve.go
│   ├── watch.go
│   ├── render.go
│   ├── validate.go
│   ├── doctor.go
│   ├── mcp.go
│   └── api.go
│
├── internal/
│   ├── core/                         # Zero external dependencies
│   │   ├── entities/
│   │   │   ├── project.go
│   │   │   ├── system.go
│   │   │   ├── container.go
│   │   │   ├── component.go
│   │   │   ├── diagram.go
│   │   │   ├── template.go
│   │   │   └── validation.go
│   │   │
│   │   ├── usecases/
│   │   │   ├── ports.go              # All interfaces
│   │   │   ├── init_project.go
│   │   │   ├── create_system.go
│   │   │   ├── create_container.go
│   │   │   ├── create_component.go
│   │   │   ├── build_docs.go
│   │   │   ├── render_diagram.go
│   │   │   ├── validate_project.go
│   │   │   ├── query_architecture.go
│   │   │   └── watch_changes.go
│   │   │
│   │   └── errors/
│   │       └── errors.go
│   │
│   ├── adapters/
│   │   ├── config/
│   │   │   ├── loader.go
│   │   │   └── config.go
│   │   │
│   │   ├── filesystem/
│   │   │   ├── project_repo.go
│   │   │   ├── template_repo.go
│   │   │   └── watcher.go
│   │   │
│   │   ├── d2/
│   │   │   ├── renderer.go
│   │   │   └── cache.go
│   │   │
│   │   ├── encoding/                 # v0.2.0
│   │   │   ├── json_encoder.go
│   │   │   └── toon_encoder.go
│   │   │
│   │   ├── ason/
│   │   │   └── engine.go
│   │   │
│   │   ├── veve/                     # v0.2.0
│   │   │   └── renderer.go
│   │   │
│   │   └── html/
│   │       ├── builder.go
│   │       └── templates/
│   │
│   ├── mcp/
│   │   ├── server.go
│   │   └── tools/
│   │       ├── registry.go
│   │       ├── query_project.go
│   │       ├── query_architecture.go
│   │       ├── create_system.go
│   │       ├── create_container.go
│   │       ├── update_diagram.go
│   │       ├── build_docs.go
│   │       └── validate.go
│   │
│   ├── api/                          # v0.2.0
│   │   ├── server.go
│   │   ├── routes.go
│   │   ├── middleware/
│   │   └── handlers/
│   │
│   └── ui/
│       ├── styles.go
│       ├── spinner.go
│       └── output.go
│
├── templates/
│   ├── standard-3layer/
│   │   ├── template.toml
│   │   └── ...
│   └── serverless/
│       └── ...
│
├── docs/
│   ├── installation.md
│   ├── quickstart.md
│   ├── configuration.md
│   ├── templates.md
│   ├── mcp-integration.md
│   ├── api-reference.md
│   └── adr/
│       ├── 0001-clean-architecture.md
│       ├── 0002-token-efficient-mcp.md
│       ├── 0003-toon-format.md
│       └── template.md
│
├── examples/
│   ├── simple-project/
│   ├── 3layer-app/
│   └── serverless/
│
└── tests/
    ├── integration/
    └── testdata/
```

---

# 10. Quick Start Commands

After creating the GitHub repository:

```bash
# 1. Create repository
gh repo create madstone-tech/loko --public --source=. --remote=origin

# 2. Initialize Go module
go mod init github.com/madstone-tech/loko

# 3. Create directory structure
mkdir -p cmd internal/core/{entities,usecases,errors}
mkdir -p internal/adapters/{config,filesystem,d2,encoding,ason,html}
mkdir -p internal/{mcp/tools,api,ui}
mkdir -p templates/{standard-3layer,serverless}
mkdir -p docs/adr examples tests/{integration,testdata}

# 4. Copy bootstrap files
# - README.md
# - CONTRIBUTING.md
# - CODE_OF_CONDUCT.md
# - ROADMAP.md
# - docs/adr/*.md

# 5. Initial commit
git add .
git commit -m "feat: initial project setup with Clean Architecture

- Add README, CONTRIBUTING, CODE_OF_CONDUCT
- Add ROADMAP with phased development plan
- Add ADRs for architecture decisions
- Set up Clean Architecture directory structure"

# 6. Push
git push -u origin main

# 7. Create initial issues (use GitHub CLI or web)
gh issue create --title "Initialize project with Clean Architecture structure" \
  --body "..." --label "enhancement,v0.1.0,priority:high"

# 8. Start building!
```

---

# Summary

This bootstrap package provides everything needed to start loko with:

✅ **Clean Architecture** - Testable, maintainable, single source of truth  
✅ **Token-Efficient MCP** - Progressive context loading (summary → structure → full)  
✅ **TOON Format** - Optional 30-40% additional token reduction (v0.2.0)  
✅ **Clear Roadmap** - Phased development with priorities  
✅ **Contributor-Friendly** - Detailed architecture guide and examples  
✅ **ADRs** - Documented decisions for future reference  

**Build Order:**
1. Foundation (entities, ports, project setup)
2. First use case end-to-end (CreateSystem)
3. Build pipeline (D2, HTML, CLI commands)
4. MCP with token-efficient queries
5. v0.2.0: TOON format, HTTP API

Ready to build! 🪇
