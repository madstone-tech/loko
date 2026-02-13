# loko Documentation

Welcome to the loko documentation! This directory contains all technical documentation, guides, and architecture decision records (ADRs) for the loko project.

## 📚 Documentation Structure

### Getting Started

- **[Quick Start Guide](quickstart.md)** - Get up and running in 5 minutes
- **[Configuration Reference](configuration.md)** - Complete loko.toml configuration options
- **[MCP Integration](mcp-integration.md)** - AI-assisted architecture design with Claude

### Guides

- **[MCP Setup Guide](guides/mcp-setup.md)** - Detailed MCP server configuration
- **[Migration Guide: Qualified IDs](migration-001-graph-qualified-ids.md)** - Upgrade to v0.2.0 with qualified node IDs

### API Reference

- **[API Reference](api-reference.md)** - HTTP API endpoints and usage
- **[MCP Tools](mcp-integration.md)** - Available MCP tools for conversational design

### Architecture Decision Records (ADRs)

- **[ADR-0001: Clean Architecture](adr/0001-clean-architecture.md)** - Dependency inversion and layered design
- **[ADR-0002: Token-Efficient MCP](adr/0002-token-efficient-mcp.md)** - Minimizing token costs for LLM interactions
- **[ADR-0003: TOON Format](adr/0003-toon-format.md)** - Tree-Oriented Object Notation for compact architecture representation
- **[ADR-0004: Graph Conventions](adr/0004-graph-conventions.md)** - Node ID format, thread safety, and graph lifecycle

### Project Planning

- **[Roadmap](roadmap.md)** - Feature roadmap and future plans

### Development

- **[Claude AI Guide](development/claude-guide.md)** - Guide for AI assistants working on loko
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute to loko
- **[Code of Conduct](../CODE_OF_CONDUCT.md)** - Community guidelines

### Release Notes

- **[CHANGELOG](../CHANGELOG.md)** - Version history and release notes

## 🎯 Quick Navigation by Task

### I want to...

**...get started with loko**
→ [Quick Start Guide](quickstart.md)

**...use loko with Claude/AI**
→ [MCP Integration](mcp-integration.md) → [MCP Setup Guide](guides/mcp-setup.md)

**...upgrade to v0.2.0**
→ [Migration Guide](migration-001-graph-qualified-ids.md)

**...understand architecture decisions**
→ [ADR Directory](adr/)

**...configure my project**
→ [Configuration Reference](configuration.md)

**...use the HTTP API**
→ [API Reference](api-reference.md)

**...contribute code**
→ [Contributing Guide](../CONTRIBUTING.md)

**...understand the roadmap**
→ [Roadmap](roadmap.md)

## 📖 Documentation by Role

### For Users

1. [Quick Start](quickstart.md) - Install and create first project
2. [Configuration](configuration.md) - Customize project settings
3. [MCP Integration](mcp-integration.md) - Use with AI assistants
4. [API Reference](api-reference.md) - HTTP API for integrations

### For Contributors

1. [Contributing Guide](../CONTRIBUTING.md) - How to contribute
2. [ADRs](adr/) - Understand architectural decisions
3. [Claude Guide](development/claude-guide.md) - AI-assisted development
4. [Code of Conduct](../CODE_OF_CONDUCT.md) - Community standards

### For AI Assistants

1. [Claude Guide](development/claude-guide.md) - Instructions for AI coding
2. [ADRs](adr/) - Architecture context
3. [AGENTS.md](../AGENTS.md) - Build, test, and development commands

## 🏗️ Architecture Overview

loko follows **Clean Architecture** with strict dependency inversion:

```
┌─────────────────────────────────────────┐
│         Adapters (External)             │
│  CLI, MCP, API, Filesystem, D2          │
├─────────────────────────────────────────┤
│         Use Cases (Application)         │
│  BuildDocs, CreateSystem, Validate      │
├─────────────────────────────────────────┤
│         Entities (Domain)               │
│  Project, System, Container, Component  │
└─────────────────────────────────────────┘
```

- **Entities**: Pure domain models (no external dependencies)
- **Use Cases**: Application logic (depends only on entities)
- **Adapters**: External interfaces (depend on use cases via ports)

See [ADR-0001](adr/0001-clean-architecture.md) for details.

## 🔍 Key Features Documented

### Architecture Graph (v0.2.0)

- **Qualified Node IDs**: Prevent collisions in multi-system projects
- **O(1) Performance**: IncomingEdges and ChildrenMap for fast queries
- **Thread-Safe Caching**: GraphCache for MCP session optimization
- **Type Safety**: C4Entity interface with compile-time checks

See [ADR-0004](adr/0004-graph-conventions.md) and [Migration Guide](migration-001-graph-qualified-ids.md)

### MCP Integration

- **Conversational Design**: Create architectures through natural language
- **8 MCP Tools**: query_project, query_architecture, create_system, etc.
- **TOON Format**: Token-efficient architecture representation
- **Session Caching**: Fast responses for repeated queries

See [MCP Integration](mcp-integration.md) and [ADR-0002](adr/0002-token-efficient-mcp.md)

### C4 Model Support

- **4 Levels**: Context, Containers, Components, Code
- **D2 Diagrams**: Modern, text-based diagram generation
- **HTML Output**: Beautiful, navigable documentation
- **Markdown Export**: Text-based documentation for wikis

See [Quick Start](quickstart.md)

## 📝 Documentation Standards

All documentation in this directory follows these standards:

- **Markdown Format**: GitHub-flavored markdown (.md)
- **Clear Headers**: Use H1 for title, H2 for main sections
- **Code Examples**: Include working, tested examples
- **Links**: Use relative paths for internal docs
- **ADRs**: Follow [MADR](https://adr.github.io/madr/) format
- **Updates**: Keep CHANGELOG.md in sync with docs

## 🤝 Contributing to Documentation

Documentation improvements are welcome! To contribute:

1. Check existing docs in this directory
2. Follow the standards above
3. Test all code examples
4. Update cross-references if needed
5. Submit a PR with clear description

See [CONTRIBUTING.md](../CONTRIBUTING.md) for full guidelines.

## 📧 Support

- **GitHub Issues**: https://github.com/madstone-tech/loko/issues
- **Discussions**: https://github.com/madstone-tech/loko/discussions
- **Email**: support@madstone.tech

---

**Version**: Documentation for loko v0.2.0  
**Last Updated**: 2025-02-13  
**Maintained by**: MADSTONE TECHNOLOGY
