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

**Last Updated:** 2025-12-15
**Maintainers:** @andhijeannot
