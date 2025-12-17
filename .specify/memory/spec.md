# loko Specification

> Generated from loko-bootstrap-final.md specification prompt
> See ISSUES.md for implementation tasks

## Overview

loko is a C4 model architecture documentation tool that enables Cloud Solution Architects to design systems conversationally with LLM agents via MCP, while also providing powerful CLI and API interfaces for direct interaction.

## User Stories

### US-1: LLM-Driven Architecture Design (P1)
As a Cloud Solution Architect, I want to use an LLM chatbot connected to loko via MCP to have a conversational workflow where the LLM guides me through designing architecture.

### US-2: Direct File Editing Workflow (P1)
As a developer, I want to edit .md and .d2 files directly in my text editor and have loko automatically rebuild documentation in real-time.

### US-3: Project Scaffolding (P1)
As a developer, I want to quickly scaffold C4 documentation structure using templates.

### US-4: API Integration (P2)
As a DevOps engineer, I want to trigger loko builds via HTTP API in CI/CD pipelines.

### US-5: Multi-Format Export (P2)
As an architect, I want to export documentation to multiple formats (HTML, Markdown, PDF).

### US-6: Token-Efficient Architecture Queries (P1)
As an LLM agent, I want to query architecture with configurable detail levels to minimize token consumption.

## Key Functional Requirements

- FR-001: TOML configuration (loko.toml) with validation
- FR-002: YAML frontmatter parsing in markdown files
- FR-003: Global and project templates via ason library
- FR-004: D2 CLI integration with caching
- FR-005: MCP server with token-efficient query tools
- FR-006: Progressive context loading (summary/structure/full)
- FR-007: CLI commands: init, new, build, serve, watch, validate, mcp, api
- FR-008: Clean Architecture with strict dependency inversion
- FR-009: TOON format support for MCP responses (v0.2.0)

## Architecture

See docs/adr/ for detailed decisions:
- ADR-0001: Clean Architecture
- ADR-0002: Token-Efficient MCP
- ADR-0003: TOON Format Support

## Implementation Plan

See **ISSUES.md** for phased implementation:
- Phase 1: Foundation (project structure, entities, ports)
- Phase 2: First Use Case (CreateSystem end-to-end)
- Phase 3: Build Pipeline (d2, HTML generation, CLI)
- Phase 4: MCP (token-efficient queries, MCP server)
- Phase 5: v0.2.0 (TOON format, HTTP API)

## External Dependencies

| Dependency | Type | Purpose |
|------------|------|---------|
| [d2](https://d2lang.com) | CLI (shell out) | Diagram rendering |
| [veve-cli](https://github.com/madstone-tech/veve-cli) | CLI (shell out) | PDF generation |
| [ason](https://github.com/madstone-tech/ason) | Go library | Template scaffolding |
| [toon-go](https://github.com/toon-format/toon-go) | Go library | Token-efficient encoding |

### ason Documentation
- GitHub: https://github.com/madstone-tech/ason
- LLM Docs: https://context7.com/madstone-tech/ason/llms.txt
