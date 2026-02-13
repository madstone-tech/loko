# loko MCP Tools Reference

This document provides a complete reference for all MCP tools exposed by loko, optimized for LLM consumption.

## Overview

loko exposes 12 MCP tools organized into three categories:
- **Query Tools**: Read architecture information
- **Creation Tools**: Create and modify architecture elements
- **Build Tools**: Validate and generate documentation

## Query Tools

### query_project

Get project overview and statistics.

**Parameters**: None required

**Returns**:
```json
{
  "name": "MyProject",
  "description": "Project description",
  "stats": {
    "systems": 4,
    "containers": 12,
    "components": 34
  }
}
```

**When to use**:
- Starting a new conversation to understand the project
- Before creating new elements to avoid duplicates
- When user asks "what's in this project?"

---

### query_architecture

Query architecture with configurable detail levels for token efficiency.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| scope | string | No | `"project"`, `"system"`, or `"container"` |
| target | string | No | Specific entity name to query |
| detail | string | No | `"summary"`, `"structure"`, or `"full"` |
| format | string | No | `"json"` or `"toon"` |
| include_diagrams | boolean | No | Include D2 source code |

**Detail Levels**:

- **summary** (~200 tokens): Names and counts only
  ```
  Project: MyProject (4 systems, 12 containers)
  Systems: AuthService, OrderService, PaymentService, NotificationService
  ```

- **structure** (~500 tokens): Hierarchy without full details
  ```
  AuthService:
    - API (Go)
    - Database (PostgreSQL)
    - Cache (Redis)
  OrderService:
    - API (Go)
    - Worker (Go)
    - Database (PostgreSQL)
  ```

- **full**: Complete details including descriptions

**When to use**:
- `summary`: Initial exploration, large projects
- `structure`: Understanding system organization
- `full`: Deep dive into specific system/container
- Always use `target` parameter to scope queries on large projects

---

### query_dependencies

Analyze dependencies between components.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| entity_id | string | Yes | ID of entity to analyze |
| direction | string | No | `"outgoing"`, `"incoming"`, or `"both"` |
| depth | integer | No | How many levels deep (default: 1) |

**Returns**:
```json
{
  "entity": "OrderService",
  "outgoing": ["PaymentService", "NotificationService"],
  "incoming": ["WebApp", "MobileApp"],
  "transitive": ["EmailProvider", "SMSProvider"]
}
```

**When to use**:
- Understanding impact of changes
- Finding circular dependencies
- Mapping integration points

---

### query_related_components

Find components related to a given component.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| component_id | string | Yes | Component to find relations for |
| relation_type | string | No | Filter by relation type |

**When to use**:
- Understanding component interactions
- Finding affected components before refactoring

---

### analyze_coupling

Analyze coupling between systems or containers.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| source | string | Yes | Source entity |
| target | string | No | Target entity (omit for all) |

**Returns**: Coupling metrics and recommendations

**When to use**:
- Architecture review
- Identifying tightly coupled components
- Planning decomposition

---

## Creation Tools

### create_system

Create a new system in the project.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | System name (PascalCase recommended) |
| description | string | No | What this system does |
| technology | string | No | Primary technology stack |

**Example**:
```json
{
  "name": "NotificationService",
  "description": "Handles email and SMS notifications",
  "technology": "Go, AWS SES, Twilio"
}
```

**Side Effects**:
- Creates `src/NotificationService/` directory
- Creates `system.md` with frontmatter
- Creates `system.d2` with starter diagram

**When to use**:
- User wants to add a new system
- Starting architecture from scratch
- Breaking apart a monolith

---

### create_container

Create a new container within a system.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | Container name |
| parent | string | Yes | Parent system name |
| description | string | No | What this container does |
| technology | string | No | Technology stack |

**Example**:
```json
{
  "name": "API",
  "parent": "NotificationService",
  "description": "REST API for notification management",
  "technology": "Go, Gin"
}
```

**Side Effects**:
- Creates `src/NotificationService/API/` directory
- Creates `container.md` with frontmatter
- Creates `container.d2` with starter diagram
- Updates parent system's D2 diagram

**When to use**:
- Adding a new deployable unit to a system
- Defining APIs, databases, caches, workers

---

### create_component

Create a new component within a container.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | Component name |
| parent | string | Yes | Parent container (format: `System/Container`) |
| description | string | No | What this component does |
| technology | string | No | Implementation details |

**Example**:
```json
{
  "name": "EmailSender",
  "parent": "NotificationService/API",
  "description": "Handles email composition and delivery",
  "technology": "AWS SES SDK"
}
```

**When to use**:
- Detailing internal structure of a container
- Documenting major code modules
- Level 3 (component) diagrams

---

### update_diagram

Update or create a D2 diagram file.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| path | string | Yes | Relative path to .d2 file |
| content | string | Yes | D2 source code |
| validate | boolean | No | Validate before saving (default: true) |

**Example**:
```json
{
  "path": "src/NotificationService/system.d2",
  "content": "direction: right\n\nAPI -> Database: \"Queries\"\nAPI -> Queue: \"Publishes\"",
  "validate": true
}
```

**When to use**:
- After creating systems/containers to add relationships
- Customizing auto-generated diagrams
- Adding styling or icons

**Best Practice**: Always validate before saving to catch syntax errors.

---

## Build Tools

### build_docs

Build documentation from source files.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| format | string | No | `"html"`, `"markdown"`, or `"pdf"` |
| clean | boolean | No | Rebuild everything (ignore cache) |
| output | string | No | Output directory |

**Returns**:
```json
{
  "success": true,
  "output_dir": "dist",
  "files_generated": 15,
  "diagrams_rendered": 8,
  "duration_ms": 1250
}
```

**When to use**:
- After making changes to preview results
- User asks to "build" or "generate" documentation
- Before sharing/publishing

---

### validate

Validate architecture for issues.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| scope | string | No | `"project"`, `"system"`, or specific name |
| strict | boolean | No | Fail on warnings (default: false) |

**Returns**:
```json
{
  "valid": false,
  "errors": [
    {
      "code": "empty_system",
      "message": "System 'TestService' has no containers",
      "path": "src/TestService/"
    }
  ],
  "warnings": [
    {
      "code": "missing_description",
      "message": "Container 'Cache' has no description",
      "path": "src/OrderService/Cache/"
    }
  ]
}
```

**Validation Checks**:
- Empty systems (no containers)
- Orphaned references (links to non-existent entities)
- Missing required files (system.md, container.md)
- Invalid C4 hierarchy
- D2 syntax errors

**When to use**:
- After creating/modifying architecture
- Before building documentation
- User asks to "check" or "validate"

---

### validate_diagram

Validate D2 diagram syntax without saving.

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| content | string | Yes | D2 source code to validate |

**Returns**:
```json
{
  "valid": true
}
// or
{
  "valid": false,
  "error": "line 5: unclosed brace",
  "line": 5
}
```

**When to use**:
- Before calling `update_diagram`
- When user provides D2 code to check
- Debugging diagram rendering issues

---

## Workflow Patterns

### Starting a New Project

```
1. query_project          # Check if project exists
2. create_system          # Create first system
3. create_container       # Add containers
4. update_diagram         # Customize diagrams
5. validate               # Check for issues
6. build_docs             # Generate output
```

### Exploring Existing Architecture

```
1. query_project                              # Overview
2. query_architecture(detail: "summary")      # Quick scan
3. query_architecture(target: "X", detail: "full")  # Deep dive
4. query_dependencies(entity_id: "X")         # Understand relationships
```

### Making Changes

```
1. query_architecture(target: "X")   # Current state
2. create_container/create_component # Add new elements
3. update_diagram                    # Update relationships
4. validate                          # Check validity
5. build_docs                        # Rebuild
```

### Token-Efficient Large Projects

```
1. query_architecture(detail: "summary", format: "toon")
2. query_architecture(target: "specific_system", detail: "structure")
3. # Only query full details for the specific area being modified
```

## Error Handling

Common errors and how to handle them:

| Error | Cause | Solution |
|-------|-------|----------|
| `entity_not_found` | Referenced entity doesn't exist | Use `query_project` to list valid entities |
| `duplicate_name` | Entity with name already exists | Choose different name or update existing |
| `invalid_parent` | Parent entity doesn't exist | Create parent first |
| `validation_failed` | D2 syntax error | Use `validate_diagram` to debug |
| `permission_denied` | Cannot write to path | Check file permissions |

## Best Practices

1. **Always query first**: Understand existing structure before creating
2. **Validate before save**: Use `validate_diagram` before `update_diagram`
3. **Use appropriate detail level**: Don't fetch full details unless needed
4. **Target specific entities**: Scope queries to relevant systems
5. **Build after changes**: Generate docs to verify results
6. **Handle errors gracefully**: Check return values for success/failure
