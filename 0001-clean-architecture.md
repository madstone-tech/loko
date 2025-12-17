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
