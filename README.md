# 🪇 loko - Guardian of Architectural Wisdom

> _Transform complexity into clarity with C4 model documentation and LLM integration_

**loko** (Papa Loko) is a modern architecture documentation tool that brings the [C4 model](https://c4model.com/) to life through conversational design with LLMs, powerful CLI workflows, and beautiful documentation generation.

[![Go Version](https://img.shields.io/github/go-mod/go-version/madstone-tech/loko)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/madstone-tech/loko)](https://github.com/madstone-tech/loko/releases)
[![License](https://img.shields.io/badge/license-BUSL--1.1-blue)](LICENSE)
[![CI](https://github.com/madstone-tech/loko/actions/workflows/ci.yml/badge.svg)](https://github.com/madstone-tech/loko/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/ghcr.io-loko-blue?logo=docker)](https://github.com/madstone-tech/loko/pkgs/container/loko)

---

## ✨ Features

🤖 **LLM-First Design** - 17 MCP tools for conversational architecture with Claude, GPT, or Gemini  
📝 **Direct Editing** - Edit markdown and [D2](https://d2lang.com) diagrams in your favorite editor  
⚡ **Real-Time Feedback** - Watch mode rebuilds in <500ms with hot reload  
🎨 **Beautiful Output** - Generate HTML, Markdown, and TOON formats  
🔧 **Powerful CLI** - Scaffold, build, validate, serve, and query — all from the terminal  
🐳 **Docker Ready** - Official images with all dependencies included  
🎯 **Zero Config** - Smart defaults with optional customization via TOML  
💰 **Token Efficient** - 9.2% token savings with TOON format + progressive context loading  
🔗 **Relationship Graph** - Live dependency graph from frontmatter + D2 arrow syntax (v0.2.0)  
🧩 **Smart Templates** - Technology-aware component scaffolding (7 categories, v0.2.0)  
🔍 **Drift Detection** - Catch inconsistencies between D2 diagrams and frontmatter (v0.2.0)

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

### 3️⃣ CI/CD Integration

```bash
# Validate architecture (with drift detection)
loko validate --strict --exit-code --check-drift

# Build documentation
loko build --format html,markdown,toon
```

See [docs/guides/ci-cd-integration.md](docs/guides/ci-cd-integration.md) for GitHub Actions and GitLab CI examples.

---

## 🛠️ MCP Tools (17 Available)

loko provides 17 MCP tools for LLM-assisted architecture workflows:

| Tool | Description |
|------|-------------|
| `query_project` | Get project metadata |
| `query_architecture` | Token-efficient architecture queries (summary/structure/full) |
| `search_elements` | Search by name, type, technology, or tags |
| `find_relationships` | Find connections between elements |
| `query_dependencies` | Find what a component depends on (direct + transitive) |
| `query_related_components` | Find components related to a given component |
| `analyze_coupling` | Measure coupling metrics across the architecture |
| `create_system` | Scaffold new system |
| `create_container` | Scaffold container |
| `create_component` | Scaffold with technology-aware template + optional D2 preview |
| `update_system` | Update system metadata |
| `update_container` | Update container metadata |
| `update_component` | Update component metadata |
| `update_diagram` | Write D2 code to file |
| `build_docs` | Generate HTML/Markdown/TOON docs (auto-populates component tables) |
| `validate` | Check architecture consistency + optional drift detection |
| `validate_diagram` | Verify D2 syntax |

> **Note**: `find_relationships`, `query_dependencies`, `query_related_components`, and `analyze_coupling` return live graph data as of v0.2.0.

**Setup:** See [docs/guides/mcp-integration-guide.md](docs/guides/mcp-integration-guide.md) for Claude Desktop configuration.

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

### TOON Format

[TOON v3.0](https://github.com/toon-format/toon-go) reduces tokens by 9.2% for structured data:

```bash
# Export architecture in token-efficient format
loko build --format toon

# Query with TOON format via MCP
query_architecture --format toon --detail summary

# Measured savings (5 systems, 15 containers):
# JSON: 4,550 tokens
# TOON: 4,131 tokens
# Savings: 9.2% (419 tokens)
```

See [docs/guides/toon-format-guide.md](docs/guides/toon-format-guide.md) for details.

---

## 🎨 Features Deep Dive

### Technology-Aware Templates (v0.2.0)

`loko new component` auto-selects the best template based on the component's technology:

| Category | Example Technologies |
|----------|---------------------|
| `compute` | AWS Lambda, Azure Functions, Cloud Run |
| `datastore` | PostgreSQL, DynamoDB, Redis, MongoDB |
| `messaging` | Kafka, SQS, RabbitMQ, SNS |
| `api` | REST, GraphQL, gRPC, FastAPI |
| `event` | EventBridge, Pub/Sub, Event Grid |
| `storage` | S3, GCS, Azure Blob |
| `generic` | (default fallback) |

```bash
# Auto-selected template based on technology
loko new component AuthService --parent api-service --technology "AWS Lambda"

# Preview the component's position in the container diagram
loko new component AuthService --parent api-service --preview

# Explicit override
loko new component AuthService --parent api-service --template datastore
```

See [docs/guides/templates.md](docs/guides/templates.md) for the full mapping table and custom templates.

### Relationship Graph (v0.2.0)

loko builds a live dependency graph by merging two sources:

```yaml
# frontmatter (system.md or component.md)
relationships:
  - target: "DatabaseService"
    label: "Reads from"
    technology: "SQL"
```

```d2
# D2 arrow syntax (system.d2)
AuthService -> DatabaseService: Queries
```

```bash
# Query via CLI
loko query --relationships PaymentService

# Query via MCP
find_relationships --source "PaymentService/**"
query_dependencies --target "AuthService"
```

See [docs/guides/relationships.md](docs/guides/relationships.md) for frontmatter syntax and D2 conventions.

### Drift Detection (v0.2.0)

Catch inconsistencies between your D2 diagrams and frontmatter metadata:

```bash
loko validate --check-drift
# WARNING  AuthService: D2 tooltip differs from frontmatter description
# ERROR    PaymentService: D2 arrow references non-existent component "OldService"
# Exit code 1 (errors found)
```

Severity levels: `WARNING` (description mismatches) and `ERROR` (orphaned references, missing components).

See [docs/guides/data-model.md](docs/guides/data-model.md) for the source-of-truth hierarchy.

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
loko build --format toon       # Token-efficient TOON format
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

## 🌟 Examples

Check out [examples/](examples/) for complete projects:

- **[simple-project](examples/simple-project/)** - Minimal C4 documentation
- **[3layer-app](examples/3layer-app/)** - Standard web → API → database
- **[microservices](examples/microservices/)** - Multi-service architecture

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

- Go 1.25+
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

## 📖 Documentation

Comprehensive documentation is available in the [`docs/`](docs/) directory:

### Getting Started
- **[Quick Start Guide](docs/quickstart.md)** - Get running in 5 minutes
- **[CLI Reference](docs/cli-reference.md)** - All commands and flags
- **[Configuration Reference](docs/configuration.md)** - Complete loko.toml options
- **[MCP Integration](docs/mcp-integration.md)** - AI-assisted architecture design

### Guides
- **[Relationships Guide](docs/guides/relationships.md)** - Frontmatter syntax, D2 arrows, union merge
- **[Templates Guide](docs/guides/templates.md)** - Technology-to-template mapping, custom templates
- **[Data Model & Drift Detection](docs/guides/data-model.md)** - Source of truth hierarchy
- **[MCP Setup Guide](docs/guides/mcp-setup.md)** - Detailed MCP configuration
- **[MCP Integration Guide](docs/guides/mcp-integration-guide.md)** - v0.2.0 relationship tools
- **[CI/CD Integration](docs/guides/ci-cd-integration.md)** - GitHub Actions, GitLab CI
- **[TOON Format Guide](docs/guides/toon-format-guide.md)** - Token-efficient notation
- **[Migration Guide v0.2.0](docs/migration-001-graph-qualified-ids.md)** - Qualified node IDs

### Reference
- **[API Reference](docs/api-reference.md)** - HTTP API endpoints
- **[ADR-0001: Clean Architecture](docs/adr/0001-clean-architecture.md)**
- **[ADR-0002: Token-Efficient MCP](docs/adr/0002-token-efficient-mcp.md)**
- **[ADR-0003: TOON Format](docs/adr/0003-toon-format.md)**
- **[ADR-0004: Graph Conventions](docs/adr/0004-graph-conventions.md)**
- **[CHANGELOG](CHANGELOG.md)** - Version history and release notes

See the **[Documentation Index](docs/README.md)** for the complete catalog.

---

## 🗺️ Roadmap

### v0.1.0 (Released) ✅

- Clean Architecture with 18 port interfaces
- Domain entities (Project, System, Container, Component)
- CLI framework (Cobra + Viper) with shell completions
- Template system (standard-3layer, serverless project templates)
- D2 diagram rendering with caching
- HTML site generation with watch mode
- MCP server (15 tools for LLM integration)

### v0.2.0 (Released) ✅

- **Functional Relationship Graph**: `find_relationships`, `query_dependencies`, `query_related_components`, `analyze_coupling` now return live data
- **Technology-Aware Templates**: 7 component templates (compute, datastore, messaging, api, event, storage, generic); `--template` override
- **D2 Diagram Preview**: `loko new component --preview` renders component position; MCP `preview: true` parameter
- **Auto-Generated Tables**: `{{component_table}}` / `{{container_table}}` placeholders auto-populated in docs
- **Drift Detection**: `loko validate --check-drift` with WARNING/ERROR severity and exit code 1 on errors
- **Coverage**: Core package coverage improved from 58.1% to 80.7%

### v0.3.0 (Future)

- OpenAPI serving (Swagger UI at `/api/docs`)
- Architecture diff and changelog generation
- Plugin system
- Multi-project support

---

## 🤝 Contributing

We welcome contributions! loko is **building in public** — see our [development progress](https://github.com/madstone-tech/loko/issues).

- 🐛 **Bug reports** → [Open an issue](https://github.com/madstone-tech/loko/issues/new?template=bug_report.md)
- 💡 **Feature requests** → [Start a discussion](https://github.com/madstone-tech/loko/discussions/new?category=ideas)
- 🔧 **Pull requests** → See [CONTRIBUTING.md](CONTRIBUTING.md)

### Quality gates

loko enforces its [Clean Architecture constitution](.specify/memory/constitution.md)
mechanically. Before opening a PR, run:

```bash
task lint                 # gofmt, vet, golangci-lint (incl. depguard layer rules)
task test                 # full unit + integration suite
task audit-constitution   # structural-compliance gate (file/function-size + layer-import rules)
```

`task audit-constitution` runs in well under a second and is a **required check** on `main`.
It enforces four budgets (CLI handler ≤ 50 lines, MCP handler ≤ 30, use-case file ≤ 200,
entity file ≤ 300) and the layer-import rules. New contributors: start with the one-page
[Constitution Compliance reference](docs/architecture/constitution-compliance.md) and the
feature [quickstart](specs/010-constitution-compliance/quickstart.md).

---

## 📜 License

[Business Source License 1.1](LICENSE) - Copyright (c) 2025-2026 MADSTONE TECHNOLOGY

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

Like Papa Loko, this tool acts as the guardian of your architectural wisdom — organizing knowledge, maintaining documentation traditions, and making complex systems understandable.

_"Papa Loko, you're the wind, pushing us, and we become butterflies."_ 🦋

---
