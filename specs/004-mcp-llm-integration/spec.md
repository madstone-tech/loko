# Feature Specification: MCP LLM Integration Enhancement

**Feature Branch**: `004-mcp-llm-integration`
**Created**: 2026-02-06
**Status**: Draft
**Spec Version**: 0.1.0

---

## Overview

Enhance loko's MCP tool descriptions and documentation to provide better LLM guidance. The current tool descriptions are functional but minimal — they tell LLMs *what* tools do but not *when* to use them, *how* to sequence them, or *why* certain patterns matter. This results in inconsistent LLM behavior when working with C4 architecture.

**Core Insight**: MCP tool descriptions are the *active* interface — they're always in context when tools are loaded. Enhanced descriptions = better LLM behavior = loko's core value proposition.

---

## Problem Statement

### Current State

| Tool | Current Description | Issue |
|------|---------------------|-------|
| `create_system` | "Create a new system in the project with name, description, and optional tags" | No C4 context, no workflow guidance |
| `create_container` | "Create a new container in a system" | Doesn't explain when to use vs create_system |
| `query_project` | "Query current project metadata..." | Doesn't explain this should be called first |
| `validate` | "Validate the project architecture..." | Doesn't explain what gets validated |

### Desired State

Tool descriptions should guide LLMs to:
1. **Understand C4 hierarchy** (System → Container → Component)
2. **Know the right workflow** (query first, then create, then validate)
3. **Avoid common mistakes** (creating containers without systems)
4. **Use efficient patterns** (TOON format for token savings)

---

## User Scenarios & Testing

### User Story 1 - First-Time Architecture Creation (Priority: P1)

As an LLM working with loko for the first time, I want tool descriptions to guide me through the correct workflow so I can create valid C4 architecture without trial and error.

**Why this priority**: New users/LLMs benefit most from clear guidance. Poor first impressions reduce adoption.

**Independent Test**: LLM can correctly create a system with containers and components on first attempt by following tool description guidance.

**Acceptance Scenarios**:

1. **Given** an LLM receives loko tools, **When** it reads `create_system` description, **Then** it understands to call `query_project` first to check existing state
2. **Given** an LLM needs to add a container, **When** it reads `create_container` description, **Then** it knows to specify the parent system
3. **Given** an LLM creates entities, **When** it reads tool descriptions, **Then** it knows to call `validate` and `build_docs` afterward

---

### User Story 2 - C4 Level Understanding (Priority: P1)

As an LLM, I want tool descriptions to explain C4 model concepts so I can create architecturally correct documentation without external references.

**Why this priority**: C4 model knowledge is essential for correct usage. LLMs may not have consistent C4 training.

**Independent Test**: LLM can explain the difference between System, Container, and Component based only on tool descriptions.

**Acceptance Scenarios**:

1. **Given** `create_system` description, **When** LLM reads it, **Then** it understands System = C4 Level 1 = software system boundary
2. **Given** `create_container` description, **When** LLM reads it, **Then** it understands Container = C4 Level 2 = deployable unit (API, database, queue)
3. **Given** `create_component` description, **When** LLM reads it, **Then** it understands Component = C4 Level 3 = code module within a container

---

### User Story 3 - Query Before Mutate Pattern (Priority: P1)

As an LLM, I want tool descriptions to establish the query-before-mutate pattern so I don't create duplicate or orphaned entities.

**Why this priority**: Prevents common errors that corrupt architecture state.

**Independent Test**: LLM naturally queries existing architecture before creating new entities.

**Acceptance Scenarios**:

1. **Given** user asks to add a new system, **When** LLM reads `create_system` description, **Then** it calls `query_project` first
2. **Given** user asks to add a container, **When** LLM reads `create_container` description, **Then** it verifies parent system exists first
3. **Given** architecture already has the requested entity, **When** LLM queries first, **Then** it informs user rather than creating duplicate

---

### User Story 4 - Token-Efficient Workflows (Priority: P2)

As an LLM with context limits, I want tool descriptions to guide me toward token-efficient patterns so I can work with large architectures.

**Why this priority**: Large projects can exceed context limits. TOON format and summary detail levels help.

**Independent Test**: LLM uses `detail: "summary"` for initial exploration, switches to `detail: "full"` only when needed.

**Acceptance Scenarios**:

1. **Given** `query_architecture` description, **When** LLM reads it, **Then** it understands to start with `summary` detail level
2. **Given** LLM needs efficient output, **When** it reads description, **Then** it knows TOON format saves 30-40% tokens
3. **Given** large architecture, **When** LLM queries, **Then** it uses `target_system` to scope responses

---

### User Story 5 - D2 Diagram Guidance (Priority: P2)

As an LLM creating diagrams, I want tool descriptions to explain D2 syntax patterns for C4 so I can generate valid, well-styled diagrams.

**Why this priority**: D2 syntax errors are common. Guidance prevents invalid diagrams.

**Independent Test**: LLM can generate valid D2 diagram syntax with appropriate C4 styling.

**Acceptance Scenarios**:

1. **Given** `update_diagram` description, **When** LLM reads it, **Then** it understands D2 syntax basics
2. **Given** LLM creates system diagram, **When** following description, **Then** uses solid lines for sync, dashed for async
3. **Given** `validate_diagram` description, **When** LLM reads it, **Then** it knows to validate before building docs

---

### Edge Cases

- What if LLM ignores description guidance? → Validation tools catch errors with clear messages
- What if project has no systems yet? → `query_project` returns empty state, descriptions guide creation
- What if LLM generates invalid D2? → `validate_diagram` catches errors before `build_docs`
- What if architecture is very large? → Descriptions guide use of scoped queries and TOON format

---

## Requirements

### Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-001 | All tool descriptions MUST include a "When to use" section | P1 |
| FR-002 | All tool descriptions MUST include C4 level context where applicable | P1 |
| FR-003 | All creation tools MUST mention the query-first pattern | P1 |
| FR-004 | `query_architecture` description MUST explain detail levels and token estimates | P1 |
| FR-005 | `query_architecture` description MUST explain TOON format benefits | P2 |
| FR-006 | `update_diagram` description MUST include D2 syntax examples | P2 |
| FR-007 | All tools MUST include example invocations in descriptions | P2 |
| FR-008 | Tool descriptions MUST guide toward validation after mutations | P1 |
| FR-009 | `validate_diagram` MUST explain C4 compliance checks | P2 |
| FR-010 | llms.txt reference documentation MUST be discoverable via tool descriptions | P3 |

### Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-001 | Tool description length | 100-300 words (enough guidance without overwhelming) |
| NFR-002 | Example invocation clarity | Copy-pasteable JSON with realistic values |
| NFR-003 | C4 terminology consistency | Use exact C4 model terms (System, Container, Component) |
| NFR-004 | Token overhead of descriptions | < 500 tokens per tool (reasonable context cost) |

---

## Tool Description Templates

### Creation Tools Template

```
[Brief description of what this creates]

**C4 Level**: [1/2/3] - [System/Container/Component]
**[Level Name]s are**: [Definition in C4 terms]

**When to use**:
- [Primary use case]
- [Secondary use case]

**Workflow**:
1. Call `query_project` to check existing architecture
2. [This tool] to create the entity
3. Call `validate` to check consistency
4. Call `build_docs` to generate output

**Example**:
```json
{
  "project_root": ".",
  "name": "[Realistic example]",
  "description": "[Helpful description]"
}
```
```

### Query Tools Template

```
[Brief description of what this queries]

**When to use**:
- [Primary use case]
- [When NOT to use - suggest alternative]

**Detail levels** (if applicable):
- summary: [token estimate] - [what's included]
- structure: [token estimate] - [what's included]
- full: [token estimate] - [what's included]

**Recommended workflow**:
[Guidance on when to use this vs other query tools]

**Example**:
```json
{
  "project_root": ".",
  "detail": "summary"
}
```
```

---

## Specific Tool Enhancements

### create_system (Current: 1 line → Target: ~150 words)

```
Create a new system (C4 Level 1) representing a software system boundary.

**C4 Level**: 1 - System Context
**Systems are**: The highest level of abstraction - a software system that delivers
value to users. Examples: "E-Commerce Platform", "Payment Gateway", "CRM System".

**When to use**:
- Starting a new architecture documentation project
- Adding a major software system to an existing project
- Documenting a bounded context in domain-driven design

**Workflow**:
1. Call `query_project` first to see existing systems
2. Call `create_system` with name and description
3. Add containers with `create_container`
4. Call `validate` to check consistency

**Example**:
{
  "project_root": ".",
  "name": "Order Management",
  "description": "Handles order lifecycle from creation to fulfillment",
  "primary_language": "Go",
  "framework": "Fiber"
}
```

### create_container (Current: 1 line → Target: ~150 words)

```
Create a new container (C4 Level 2) within a system.

**C4 Level**: 2 - Container
**Containers are**: Deployable units that make up a system - APIs, databases,
message queues, web apps. Each container is separately deployable.

**When to use**:
- Adding a deployable service to a system (API, worker, database)
- Documenting infrastructure components (queues, caches, storage)
- Breaking down a system into its runtime units

**Requires**: Parent system must exist. Call `query_project` to verify.

**Workflow**:
1. Verify parent system exists via `query_project`
2. Call `create_container` with system_name and container details
3. Add components with `create_component` if needed
4. Call `validate` to check consistency

**Example**:
{
  "project_root": ".",
  "system_name": "Order Management",
  "name": "Order API",
  "description": "REST API for order operations",
  "technology": "Go + Fiber"
}
```

### query_project (Current: 1 line → Target: ~100 words)

```
Query project metadata and architecture summary. **Call this first** before
creating or modifying entities.

**When to use**:
- Starting any architecture task (always call first)
- Checking what systems exist before creating new ones
- Getting project-wide statistics (system/container/component counts)

**When to use query_architecture instead**:
- Need detailed structure, not just counts
- Need specific format (JSON, TOON)
- Need to scope to a specific system

**Example**:
{
  "project_root": "."
}

Returns: project name, description, version, and counts of systems/containers/components.
```

---

## Success Criteria

### Measurable Outcomes

| ID | Criterion | Target | Measurement |
|----|-----------|--------|-------------|
| SC-001 | LLM first-attempt success rate | > 80% | Manual testing with Claude/GPT |
| SC-002 | Query-before-create pattern adoption | > 90% | Observed in LLM interactions |
| SC-003 | Validation usage after mutations | > 80% | Observed in LLM interactions |
| SC-004 | Token-efficient format usage | > 50% | TOON/summary usage rate |
| SC-005 | Valid D2 generation rate | > 90% | `validate_diagram` pass rate |

---

## Scope & Exclusions

### In Scope

- Enhanced Description() methods for all 12 MCP tools
- C4 model context in relevant descriptions
- Workflow guidance and sequencing
- Example invocations with realistic values
- D2 syntax guidance in diagram tools
- Integration with llms.txt reference docs

### Out of Scope (Future)

- Interactive tool tutorials
- Multi-turn conversation guidance
- Automatic workflow enforcement
- Tool chaining/composition primitives
- Custom tool aliases for LLMs

---

## Implementation Notes

### Files to Modify

1. `internal/mcp/tools/tools.go` - Creation tools (create_system, create_container, create_component, update_diagram, build_docs, validate)
2. `internal/mcp/tools/query_project.go` - query_project tool
3. `internal/mcp/tools/query_architecture.go` - query_architecture tool (already good, minor enhancements)
4. `internal/mcp/tools/graph_tools.go` - Graph query tools (query_dependencies, query_related_components, analyze_coupling)
5. `internal/mcp/tools/schemas.go` - validate_diagram tool (if separate)

### Testing Approach

1. **Manual LLM Testing**: Test with Claude and GPT-4 using standardized prompts
2. **Prompt Suite**: Create test prompts that verify behavior:
   - "Create a new e-commerce system" → Should query first
   - "Add a database to the Order system" → Should verify system exists
   - "Show me the architecture" → Should use appropriate detail level
3. **Before/After Comparison**: Document LLM behavior before and after changes

---

## Dependencies & Assumptions

### Dependencies

- None (this is documentation-only change to existing tools)

### Assumptions

- LLMs read and follow tool descriptions
- Longer descriptions (100-300 words) are acceptable overhead
- C4 model terminology is sufficiently standardized
- Examples in descriptions improve LLM performance

---

## External References

- [C4 Model](https://c4model.com/) - Official C4 documentation
- [D2 Language](https://d2lang.com/) - D2 diagramming language
- [MCP Specification](https://modelcontextprotocol.io/) - Model Context Protocol
- [llms.txt Proposal](https://llmstxt.org/) - LLM-friendly documentation standard
- loko docs/llm/*.md - Reference documentation created for LLM consumption
