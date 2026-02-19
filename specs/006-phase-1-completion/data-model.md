# Data Model: Production-Ready Phase 1 Release

**Feature**: 006-phase-1-completion  
**Date**: 2026-02-13  
**Status**: Complete

---

## Overview

This release does **NOT introduce new entities** - it polishes and extends existing ones. Data model changes are minimal:

1. **Configuration schema** - Extend `loko.toml` with API settings
2. **CLI flags** - Add `--strict`, `--exit-code` to validation
3. **MCP tool schemas** - Add `search_elements`, `find_relationships`

---

## Configuration Schema Changes

### Current: loko.toml

```toml
[project]
name = "my-architecture"
description = "System architecture documentation"
version = "1.0.0"

[paths]
source = "./src"
output = "./dist"

[d2]
theme = "neutral-default"
layout = "elk"
cache = true

[outputs]
html = true
markdown = true
pdf = false

[build]
parallel = true
max_workers = 4

[server]
serve_port = 8080
api_port = 8081
hot_reload = true
```

### Extended: loko.toml (v0.2.0)

```toml
[project]
name = "my-architecture"
description = "System architecture documentation"
version = "1.0.0"

[paths]
source = "./src"
output = "./dist"

[d2]
theme = "neutral-default"
layout = "elk"
cache = true

[outputs]
html = true
markdown = true
pdf = false

[build]
parallel = true
max_workers = 4

[server]
serve_port = 8080
api_port = 8081
hot_reload = true

# NEW: API server configuration (optional)
[api]
rate_limit = 100                      # Requests per minute per IP (0 = disabled)
allowed_origins = ["http://localhost:*"]  # CORS origins
timeout = "30s"                       # Request timeout
enable_swagger = true                 # Serve Swagger UI at /api/docs
```

**Validation Rules**:
- `rate_limit`: Integer >= 0 (0 disables rate limiting)
- `allowed_origins`: Array of strings (glob patterns supported)
- `timeout`: Duration string (e.g., "30s", "1m")
- `enable_swagger`: Boolean (default: true)

**Backward Compatibility**:
- All `[api]` settings optional (defaults shown above)
- Existing configs work without modification
- Rate limiting disabled by default for local dev

---

## CLI Flag Changes

### Command: `loko validate`

**Current**:
```bash
loko validate [flags]
```

**Extended (v0.2.0)**:
```bash
loko validate [flags]

Flags:
  --strict        Treat warnings as errors (exit non-zero)
  --exit-code     Return exit code 1 on any errors (for CI)
  --timeout       Validation timeout (default: 30s)
```

**Behavior**:
- `--strict`: Warnings become errors, fail validation
- `--exit-code`: Ensures non-zero exit for CI detection
- Can be combined: `loko validate --strict --exit-code`

**Exit Codes**:
- `0`: Validation passed (no errors, or warnings if not strict)
- `1`: Validation failed (errors found, or warnings in strict mode)
- `2`: Validation error (timeout, file not found, etc.)

### Command: `loko build`

**Current**:
```bash
loko build [--format html|markdown|pdf]
```

**Extended (v0.2.0)**:
```bash
loko build [--format html|markdown|pdf] [flags]

Flags:
  --format        Output format (html,markdown,pdf) (default: html,markdown)
  --skip-pdf      Skip PDF generation even if enabled in config
  --output        Override output directory
```

**Behavior**:
- `--skip-pdf`: Suppresses PDF warnings if veve-cli missing
- Graceful degradation: HTML/Markdown succeed even if PDF fails

---

## MCP Tool Schemas

### Tool: search_elements

**Name**: `search_elements`  
**Category**: Query  
**Status**: New (v0.2.0)

**Purpose**: Search architecture elements by name, description, tags, or technology

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the project",
      "default": "."
    },
    "query": {
      "type": "string",
      "description": "Search query (supports glob patterns like backend.*, *-service)",
      "required": true
    },
    "type": {
      "type": "string",
      "enum": ["system", "container", "component"],
      "description": "Filter by element type (optional)"
    },
    "tag": {
      "type": "string",
      "description": "Filter by tag (exact match, optional)"
    },
    "technology": {
      "type": "string",
      "description": "Filter by technology (exact match, optional)"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum results to return",
      "default": 20,
      "minimum": 1,
      "maximum": 100
    }
  },
  "required": ["query"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "results": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "string", "description": "Qualified element ID"},
          "name": {"type": "string", "description": "Element name"},
          "type": {"type": "string", "enum": ["system", "container", "component"]},
          "description": {"type": "string"},
          "technology": {"type": "string"},
          "tags": {"type": "array", "items": {"type": "string"}},
          "path": {"type": "string", "description": "File path relative to project root"}
        }
      }
    },
    "total_matched": {"type": "integer", "description": "Total matches (may exceed limit)"},
    "query_time_ms": {"type": "integer", "description": "Query execution time"}
  }
}
```

**Example Usage**:
```json
// Request
{
  "name": "search_elements",
  "arguments": {
    "query": "payment",
    "type": "component",
    "limit": 10
  }
}

// Response
{
  "results": [
    {
      "id": "backend/api/payment-handler",
      "name": "Payment Handler",
      "type": "component",
      "description": "Processes payment transactions",
      "technology": "Go",
      "tags": ["critical", "pci-compliant"],
      "path": "src/backend/api/payment-handler/component.md"
    }
  ],
  "total_matched": 1,
  "query_time_ms": 45
}
```

---

### Tool: find_relationships

**Name**: `find_relationships`  
**Category**: Query  
**Status**: New (v0.2.0)

**Purpose**: Find relationships between architecture elements using glob patterns

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the project",
      "default": "."
    },
    "source_pattern": {
      "type": "string",
      "description": "Glob pattern for source elements (e.g., backend.*, *-service)",
      "required": true
    },
    "target_pattern": {
      "type": "string",
      "description": "Glob pattern for target elements (optional, defaults to all)"
    },
    "relationship_type": {
      "type": "string",
      "description": "Filter by relationship type (e.g., depends-on, uses)",
      "enum": ["depends-on", "uses", "calls", "reads-from", "writes-to"]
    },
    "limit": {
      "type": "integer",
      "description": "Maximum relationships to return",
      "default": 50,
      "minimum": 1,
      "maximum": 200
    }
  },
  "required": ["source_pattern"]
}
```

**Output Schema**:
```json
{
  "type": "object",
  "properties": {
    "relationships": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "source_id": {"type": "string"},
          "source_name": {"type": "string"},
          "target_id": {"type": "string"},
          "target_name": {"type": "string"},
          "type": {"type": "string"},
          "description": {"type": "string"}
        }
      }
    },
    "total_matched": {"type": "integer"},
    "query_time_ms": {"type": "integer"}
  }
}
```

**Example Usage**:
```json
// Request
{
  "name": "find_relationships",
  "arguments": {
    "source_pattern": "backend.*",
    "target_pattern": "external.*",
    "relationship_type": "depends-on"
  }
}

// Response
{
  "relationships": [
    {
      "source_id": "backend/api/auth-service",
      "source_name": "Auth Service",
      "target_id": "external/oauth-provider",
      "target_name": "OAuth Provider",
      "type": "depends-on",
      "description": "Delegates authentication to OAuth provider"
    }
  ],
  "total_matched": 1,
  "query_time_ms": 62
}
```

---

## Entity Changes

### No New Entities

This release **does NOT add new entities**. Existing entities remain unchanged:

**Unchanged**:
- `Project` - Architecture project metadata
- `System` - C4 system level
- `Container` - C4 container level
- `Component` - C4 component level
- `ArchitectureGraph` - Dependency graph representation
- `ValidationResult` - Validation errors/warnings

**Minor Extension** (non-breaking):
- `ProjectConfig` - Add optional `APIConfig` field for API server settings

```go
// internal/core/entities/project.go
type ProjectConfig struct {
    Name        string
    Description string
    Version     string
    Paths       PathConfig
    D2          D2Config
    Outputs     OutputConfig
    Build       BuildConfig
    Server      ServerConfig
    API         *APIConfig // NEW (optional pointer)
}

type APIConfig struct {
    RateLimit      int           // 0 = disabled
    AllowedOrigins []string      // CORS origins
    Timeout        time.Duration // Request timeout
    EnableSwagger  bool          // Serve Swagger UI
}
```

---

## Validation Rules

### Constitution Audit Validation

**New Validation** (automated in CI):

```bash
# Check CLI handlers (< 50 lines)
for file in cmd/*.go; do
    lines=$(grep -v '^import\|^/\|^$\|^package\|^)$' "$file" | wc -l)
    if [ "$lines" -gt 50 ]; then
        echo "❌ VIOLATION: $file has $lines lines (limit: 50)"
        exit 1
    fi
done

# Check MCP tools (< 30 lines)
for file in internal/mcp/tools/*.go; do
    lines=$(grep -v '^import\|^/\|^$\|^package\|^)$' "$file" | wc -l)
    if [ "$lines" -gt 30 ]; then
        echo "❌ VIOLATION: $file has $lines lines (limit: 30)"
        exit 1
    fi
done
```

**Exclusions**:
- Import statements
- Comments (single and multi-line)
- Blank lines
- Package declarations

### TOON Format Validation

**New Validation** (automated in tests):

```go
// Validate TOON output against official parser
func TestTOONCompliance(t *testing.T) {
    arch := loadTestArchitecture()
    toonData, err := encodeTOON(arch)
    require.NoError(t, err)
    
    // Parse with official TOON parser
    var result ArchitectureData
    err = toon.Unmarshal(toonData, &result)
    require.NoError(t, err, "TOON output must parse with official parser")
    
    // Verify data integrity
    assert.Equal(t, arch.Name, result.Name)
    assert.Len(t, result.Systems, len(arch.Systems))
}
```

---

## State Transitions

### Build Process (Enhanced)

**Current State Machine**:
```
[Init] → [Load Config] → [Render Diagrams] → [Build HTML] → [Complete]
```

**Extended State Machine** (v0.2.0):
```
[Init] 
  → [Load Config]
  → [Render Diagrams]
  → [Build HTML]
  → [Build Markdown]
  → [Build PDF] (optional, may skip if veve-cli missing)
  → [Complete]
```

**New Transitions**:
- `Build PDF` can transition to `Complete` with warning (graceful degradation)
- `Build PDF` skipped if `--skip-pdf` flag present

### Validation Process (Enhanced)

**Current State Machine**:
```
[Init] → [Load Project] → [Validate Hierarchy] → [Check References] → [Report]
```

**Extended State Machine** (v0.2.0):
```
[Init]
  → [Load Project]
  → [Validate Hierarchy]
  → [Check References]
  → [Run Constitution Audit] (optional)
  → [Classify Results] (errors vs warnings)
  → [Apply Strict Mode] (if --strict flag)
  → [Report]
  → [Exit] (code 0 or 1 based on --exit-code flag)
```

**New Transitions**:
- `Apply Strict Mode`: Warnings → Errors if `--strict` flag
- `Exit`: Returns appropriate code if `--exit-code` flag

---

## API Contracts (OpenAPI)

### Endpoint: GET /api/v1/openapi.json

**Status**: New (v0.2.0)

**Response**:
```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "loko Architecture API",
    "version": "0.2.0",
    "description": "HTTP API for C4 architecture documentation"
  },
  "servers": [
    {"url": "http://localhost:8081/api/v1"}
  ],
  "paths": {
    "/project": {...},
    "/systems": {...},
    "/systems/{id}": {...},
    "/build": {...},
    "/validate": {...}
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer"
      }
    }
  }
}
```

### Endpoint: GET /api/docs

**Status**: New (v0.2.0)

**Purpose**: Serve Swagger UI for interactive API testing

**Response**: HTML page with embedded Swagger UI

**Features**:
- Interactive testing of all API endpoints
- Bearer token authentication support
- OpenAPI spec auto-loaded from `/api/v1/openapi.json`
- Offline-capable (no CDN dependencies)

---

## Summary

**Data Model Changes**: Minimal and backward-compatible

| Change Type | Description | Breaking? |
|-------------|-------------|-----------|
| **Config Extension** | Add optional `[api]` section to loko.toml | No |
| **CLI Flags** | Add `--strict`, `--exit-code` to validate command | No |
| **MCP Tools** | Add `search_elements`, `find_relationships` tools | No |
| **API Endpoints** | Add `/api/v1/openapi.json`, `/api/docs` | No |
| **Validation Rules** | Add constitution audit checks | No |

**Migration Required**: None - all changes are additive and backward-compatible

**Next Phase**: Generate contract validation tests and quickstart guide
