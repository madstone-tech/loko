# MCP Tool Contracts: loko v0.1.0

**Created**: 2025-12-17  
**Status**: Ready for Implementation

---

## Overview

This document specifies the interface contracts for all 8 MCP tools exposed by loko. Each tool MUST:
- Accept JSON input with specified schema
- Return JSON output with specified schema
- Include clear error messages on failure
- Take <2 seconds to execute (user-friendly LLM interaction)

---

## Tool: query_project

**Purpose**: Get high-level project metadata

**User Story**: US-1 (LLM can understand project context)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project (optional, defaults to current)"
    }
  },
  "required": []
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "name": { "type": "string" },
    "description": { "type": "string" },
    "version": { "type": "string" },
    "system_count": { "type": "integer" },
    "total_containers": { "type": "integer" },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "systems": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "container_count": { "type": "integer" }
        }
      }
    }
  }
}
```

**Example Usage**:

```json
{
  "name": "payment-service",
  "description": "Payment processing system",
  "version": "0.1.0",
  "system_count": 3,
  "total_containers": 8,
  "systems": [
    {"id": "PaymentAPI", "name": "Payment API", "container_count": 3},
    {"id": "AuthService", "name": "Auth Service", "container_count": 2}
  ]
}
```

**Error Handling**:
- `project_path` invalid → "Project not found at path"
- `loko.toml` missing → "Not a valid loko project"
- `src/` not readable → "Unable to read project structure"

---

## Tool: query_architecture

**Purpose**: Token-efficient architecture queries with detail levels

**User Story**: US-6 (Token-efficient LLM context)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project (optional)"
    },
    "scope": {
      "type": "string",
      "enum": ["project", "system", "container"],
      "description": "Scope of query (default: project)"
    },
    "target": {
      "type": "string",
      "description": "For system/container scope: specific entity ID"
    },
    "detail": {
      "type": "string",
      "enum": ["summary", "structure", "full"],
      "description": "Detail level (default: summary)"
    },
    "format": {
      "type": "string",
      "enum": ["json", "toon"],
      "description": "Output format (default: json, toon in v0.2.0)"
    }
  },
  "required": []
}
```

**Output Examples**:

### summary (scope: project)

```json
{
  "type": "summary",
  "systems": 3,
  "containers": 8,
  "components": 24,
  "system_names": ["PaymentAPI", "AuthService", "DataStore"]
}
```

**Token cost**: ~200 tokens

### structure (scope: project)

```json
{
  "type": "structure",
  "systems": [
    {
      "id": "PaymentAPI",
      "name": "Payment API",
      "description": "REST API for payment operations",
      "containers": [
        {"id": "API", "name": "API Service"},
        {"id": "Cache", "name": "Redis Cache"},
        {"id": "Queue", "name": "Message Queue"}
      ]
    },
    {
      "id": "DataStore",
      "name": "Data Store",
      "containers": [
        {"id": "DB", "name": "PostgreSQL"}
      ]
    }
  ]
}
```

**Token cost**: ~500 tokens

### full (scope: system, target: "PaymentAPI")

```json
{
  "type": "full",
  "system": {
    "id": "PaymentAPI",
    "name": "Payment API",
    "description": "REST API for payment operations",
    "technology": "Node.js + Express",
    "containers": [
      {
        "id": "API",
        "name": "API Service",
        "description": "REST API",
        "technology": "Node.js",
        "components": [
          {
            "id": "PaymentHandler",
            "name": "Payment Handler",
            "description": "Processes payment requests"
          }
        ]
      }
    ]
  }
}
```

**Token cost**: Variable, optimized for single entity

**Error Handling**:
- Invalid `scope` → "scope must be: project, system, or container"
- Invalid `target` → "System/Container '{target}' not found"
- `detail:full` without `target` → "full detail requires target (system or container)"
- Invalid `format` → "format must be: json or toon"

---

## Tool: create_system

**Purpose**: Scaffold a new C4 system

**User Story**: US-1, US-3 (Create systems via LLM or CLI)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "system_name": {
      "type": "string",
      "description": "System name (e.g., 'PaymentService')"
    },
    "description": {
      "type": "string",
      "description": "System description"
    },
    "template": {
      "type": "string",
      "description": "Template to use (optional, default: system)"
    }
  },
  "required": ["project_path", "system_name"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "system": {
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" },
        "path": { "type": "string" }
      }
    },
    "files_created": {
      "type": "array",
      "items": { "type": "string" }
    },
    "message": { "type": "string" }
  }
}
```

**Example Usage**:

```json
{
  "success": true,
  "system": {
    "id": "PaymentService",
    "name": "Payment Service",
    "path": "src/PaymentService"
  },
  "files_created": [
    "src/PaymentService/system.md",
    "src/PaymentService/system.d2"
  ],
  "message": "System 'PaymentService' created successfully"
}
```

**Error Handling**:
- System already exists → "System 'PaymentService' already exists"
- Invalid system name → "System name must be alphanumeric"
- Template not found → "Template '{template}' not found"

---

## Tool: create_container

**Purpose**: Scaffold a new C4 container in a system

**User Story**: US-1, US-3 (Create containers)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "system_name": {
      "type": "string",
      "description": "Parent system ID"
    },
    "container_name": {
      "type": "string",
      "description": "Container name (e.g., 'API')"
    },
    "description": {
      "type": "string",
      "description": "Container description"
    },
    "type": {
      "type": "string",
      "description": "Container type (e.g., 'Web Service', 'Database')"
    }
  },
  "required": ["project_path", "system_name", "container_name"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "container": {
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" },
        "path": { "type": "string" }
      }
    },
    "files_created": {
      "type": "array",
      "items": { "type": "string" }
    },
    "message": { "type": "string" }
  }
}
```

**Example Usage**:

```json
{
  "success": true,
  "container": {
    "id": "API",
    "name": "API Service",
    "path": "src/PaymentService/API"
  },
  "files_created": [
    "src/PaymentService/API/container.md",
    "src/PaymentService/API/container.d2"
  ],
  "message": "Container 'API' created in system 'PaymentService'"
}
```

**Error Handling**:
- System not found → "System 'PaymentService' not found"
- Container already exists → "Container 'API' already exists in system"
- Invalid container name → "Container name must be alphanumeric"

---

## Tool: create_component

**Purpose**: Add a component to a container

**User Story**: US-1 (Create components)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "system_name": {
      "type": "string",
      "description": "Parent system ID"
    },
    "container_name": {
      "type": "string",
      "description": "Parent container ID"
    },
    "component_name": {
      "type": "string",
      "description": "Component name"
    },
    "description": {
      "type": "string",
      "description": "Component description"
    }
  },
  "required": ["project_path", "system_name", "container_name", "component_name"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "component": {
      "type": "object",
      "properties": {
        "id": { "type": "string" },
        "name": { "type": "string" }
      }
    },
    "message": { "type": "string" }
  }
}
```

**Error Handling**:
- System/Container not found → Appropriate error message
- Invalid component name → "Component name must be alphanumeric"

---

## Tool: update_diagram

**Purpose**: Update D2 diagram code for system/container

**User Story**: US-1 (LLM generates D2 code)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "scope": {
      "type": "string",
      "enum": ["system", "container"],
      "description": "What to diagram"
    },
    "target": {
      "type": "string",
      "description": "System or container ID"
    },
    "d2_code": {
      "type": "string",
      "description": "D2 diagram source code"
    }
  },
  "required": ["project_path", "scope", "target", "d2_code"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "diagram": {
      "type": "object",
      "properties": {
        "path": { "type": "string" },
        "rendered_to": { "type": "string" }
      }
    },
    "message": { "type": "string" }
  }
}
```

**Example Usage**:

```json
{
  "success": true,
  "diagram": {
    "path": "src/PaymentService/system.d2",
    "rendered_to": "dist/diagrams/system-PaymentService.svg"
  },
  "message": "Diagram updated and rendered successfully"
}
```

**Error Handling**:
- D2 syntax invalid → "D2 syntax error: [specific error]"
- Target not found → "System/Container not found"
- Rendering failed → "Failed to render diagram"

---

## Tool: build_docs

**Purpose**: Build documentation (HTML, markdown, PDF)

**User Story**: US-1, US-2, US-5 (Build complete docs)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "format": {
      "type": "string",
      "enum": ["html", "markdown", "pdf", "all"],
      "description": "Output format (default: all)"
    },
    "incremental": {
      "type": "boolean",
      "description": "Only build changed files (default: true)"
    }
  },
  "required": ["project_path"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "build": {
      "type": "object",
      "properties": {
        "output_dir": { "type": "string" },
        "duration_ms": { "type": "integer" },
        "files_generated": { "type": "integer" },
        "diagrams_rendered": { "type": "integer" },
        "cache_hits": { "type": "integer" }
      }
    },
    "message": { "type": "string" }
  }
}
```

**Example Usage**:

```json
{
  "success": true,
  "build": {
    "output_dir": "dist/",
    "duration_ms": 2847,
    "files_generated": 12,
    "diagrams_rendered": 5,
    "cache_hits": 3
  },
  "message": "Build complete: HTML, Markdown generated (PDF skipped - veve-cli not found)"
}
```

**Error Handling**:
- No systems defined → "Project has no systems. Create at least one system first"
- Diagram rendering failed → "Failed to render diagrams (see details)"
- Output directory permission denied → "Cannot write to output directory"

---

## Tool: validate

**Purpose**: Validate architecture for consistency and completeness

**User Story**: US-2 (Validate architecture)

**Input Schema**:

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to loko project"
    },
    "strict": {
      "type": "boolean",
      "description": "Treat warnings as errors (default: false)"
    }
  },
  "required": ["project_path"]
}
```

**Output Schema**:

```json
{
  "type": "object",
  "properties": {
    "success": { "type": "boolean" },
    "validation": {
      "type": "object",
      "properties": {
        "errors": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "code": { "type": "string" },
              "message": { "type": "string" },
              "location": { "type": "string" }
            }
          }
        },
        "warnings": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "code": { "type": "string" },
              "message": { "type": "string" },
              "location": { "type": "string" }
            }
          }
        }
      }
    },
    "message": { "type": "string" }
  }
}
```

**Example Usage**:

```json
{
  "success": true,
  "validation": {
    "errors": [],
    "warnings": [
      {
        "code": "ORPHANED_CONTAINER",
        "message": "Container 'OldCache' exists but not referenced",
        "location": "src/PaymentService/OldCache/"
      }
    ]
  },
  "message": "Validation passed with 1 warning"
}
```

**Validation Rules**:

**Errors** (block build):
- System ID not unique within project
- Container ID not unique within system
- Container references non-existent system
- Component references non-existent container
- Required markdown files missing (system.md, container.md)
- Invalid YAML frontmatter
- D2 syntax errors in diagrams
- Circular dependencies detected

**Warnings** (non-blocking):
- Orphaned files/directories
- Unreferenced systems/containers
- Missing descriptions
- Empty systems (no containers)
- Outdated timestamps

---

## MCP Tool Handler Code Pattern

All MCP tool handlers MUST follow this pattern:

```go
// internal/mcp/tools/query_project.go

type QueryProjectInput struct {
    ProjectPath string `json:"project_path"`
}

type QueryProjectOutput struct {
    Name             string `json:"name"`
    SystemCount      int    `json:"system_count"`
    TotalContainers  int    `json:"total_containers"`
    Systems          []SystemSummary `json:"systems"`
}

func (h *ToolHandler) QueryProject(ctx context.Context, input QueryProjectInput) (*QueryProjectOutput, error) {
    // 1. Parse input
    // 2. Call use case
    // 3. Format output
    // 4. Return structured response
    
    // Handler must be <30 lines
}
```

---

## Error Response Format

All tools MUST return errors in this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Optional additional context"
  }
}
```

**Error Codes**:
- `INVALID_INPUT` - Input validation failed
- `NOT_FOUND` - Resource doesn't exist
- `ALREADY_EXISTS` - Resource already exists
- `PERMISSION_DENIED` - File access denied
- `INVALID_SYNTAX` - D2/Markdown syntax error
- `INTERNAL_ERROR` - Unexpected server error

---
