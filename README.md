# 🪇 loko - Guardian of Architectural Wisdom

> _Transform complexity into clarity with C4 model documentation and LLM integration_

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

loko includes built-in templates powered by [ason](https://github.com/madstone-tech/ason):

| Template | Use Case |
|----------|----------|
| `standard-3layer` | Traditional web apps (API → Service → Database) |
| `serverless` | AWS Lambda architectures (API Gateway, SQS, DynamoDB) |

```bash
# Use default template (standard-3layer)
loko new system PaymentService

# Use serverless template for AWS Lambda architectures
loko new system "Order Processing API" -template serverless
loko new container "API Handlers" -parent order-processing-api -template serverless
```

Templates use ason's variable interpolation syntax for scaffolding. See [ason documentation](https://context7.com/madstone-tech/ason/llms.txt) for template authoring.

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

| Tool                 | Description                          |
| -------------------- | ------------------------------------ |
| `query_project`      | Get project metadata                 |
| `query_architecture` | Token-efficient architecture queries |
| `create_system`      | Scaffold new system                  |
| `create_container`   | Scaffold container                   |
| `create_component`   | Scaffold component                   |
| `update_diagram`     | Write D2 code to file                |
| `build_docs`         | Build documentation                  |
| `validate`           | Check architecture consistency       |

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
- **[MCP Integration](docs/mcp-integration.md)** - AI-assisted architecture design
- **[Configuration Reference](docs/configuration.md)** - Complete loko.toml options

### Guides
- **[MCP Setup Guide](docs/guides/mcp-setup.md)** - Detailed MCP configuration
- **[Migration Guide v0.2.0](docs/migration-001-graph-qualified-ids.md)** - Upgrade to qualified node IDs

### Architecture
- **[ADR-0001: Clean Architecture](docs/adr/0001-clean-architecture.md)** - Dependency inversion
- **[ADR-0002: Token-Efficient MCP](docs/adr/0002-token-efficient-mcp.md)** - Minimizing LLM costs
- **[ADR-0003: TOON Format](docs/adr/0003-toon-format.md)** - Compact architecture notation
- **[ADR-0004: Graph Conventions](docs/adr/0004-graph-conventions.md)** - Node IDs and thread safety

### Reference
- **[API Reference](docs/api-reference.md)** - HTTP API endpoints
- **[CHANGELOG](CHANGELOG.md)** - Version history and release notes

See the **[Documentation Index](docs/README.md)** for the complete catalog.

---

## 🗺️ Roadmap

### v0.1.0 (MVP) - In Progress

**Foundation (Complete)**
- ✅ Clean Architecture with 18 port interfaces
- ✅ Domain entities (Project, System, Container, Component) with tests
- ✅ CLI framework (Cobra + Viper) with shell completions
- ✅ Template system with standard-3layer and serverless templates

**Current: Handler Refactoring + TOON Alignment (#005)**
- 🟡 Extract business logic from CLI/MCP handlers into use cases
- 🟡 Align TOON encoder with official TOON v3.0 specification

**Remaining**
- 🔲 Scaffolding use cases with ason template engine adapter
- 🔲 D2 diagram rendering with caching
- 🔲 HTML site generation with watch mode
- 🔲 MCP server with token-efficient queries

### v0.2.0

- 📋 HTTP API server for CI/CD integration
- 📋 PDF export via veve-cli
- 📋 Multi-format export (HTML, Markdown, PDF)
- 📋 Advanced validation rules

### v0.3.0

- 📋 Architecture graph visualization
- 📋 Diff and changelog generation
- 📋 Plugin system

See [specs/](specs/) for detailed feature specifications.

---

## 🤲 Contributing

We welcome contributions! loko is **building in public** - see our [development progress](https://github.com/madstone-tech/loko/issues).

- 🐛 **Bug reports** → [Open an issue](https://github.com/madstone-tech/loko/issues/new?template=bug_report.md)
- 💡 **Feature requests** → [Start a discussion](https://github.com/madstone-tech/loko/discussions/new?category=ideas)
- 🔧 **Pull requests** → See [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📜 License

[MIT License](LICENSE) - Copyright (c) 2025-2026 MADSTONE TECHNOLOGY

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

_"Papa Loko, you're the wind, pushing us, and we become butterflies."_ 🦋

---
