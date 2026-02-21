# QA Test Suite — MCP Tools
**Branch:** `008-mcp-ux-improvements` | **Date:** 2026-02-20

All tests use the notification-service fixture project under `test/src/`.  
Run the MCP server with: `loko mcp --project ./test`

---

## Setup

Before running any test, verify the server starts cleanly:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"qa","version":"1.0"}}}' \
  | loko mcp --project ./test
```

**Expected:** JSON response with `"result"` containing `"serverInfo": {"name":"loko"}` and no error. A blank line on stderr signals ready.

---

## T01 — query_project

**Tool:** `query_project`  
**Purpose:** Returns project metadata and element counts.

```json
{
  "jsonrpc": "2.0", "id": 1, "method": "tools/call",
  "params": {
    "name": "query_project",
    "arguments": { "project_root": "./test" }
  }
}
```

**Pass criteria:**
- `result.project.name` is non-empty
- `result.stats.systems` ≥ 1
- `result.stats.containers` ≥ 1
- `result.stats.components` ≥ 1
- No `"error"` key in response

---

## T02 — query_architecture (summary)

**Tool:** `query_architecture`

```json
{
  "jsonrpc": "2.0", "id": 2, "method": "tools/call",
  "params": {
    "name": "query_architecture",
    "arguments": {
      "project_root": "./test",
      "detail": "summary"
    }
  }
}
```

**Pass criteria:**
- `result.system_count` ≥ 1
- `result.text` is a non-empty string
- `result.token_estimate` is a positive integer

---

## T03 — query_architecture (full + target_system)

```json
{
  "jsonrpc": "2.0", "id": 3, "method": "tools/call",
  "params": {
    "name": "query_architecture",
    "arguments": {
      "project_root": "./test",
      "detail": "full",
      "target_system": "notification-service"
    }
  }
}
```

**Pass criteria:**
- `result._target_system` = `"notification-service"`
- `result.text` contains `"notification-service"` or `"Notification"` (case-insensitive)

---

## T04 — create_system

**Tool:** `create_system`

```json
{
  "jsonrpc": "2.0", "id": 4, "method": "tools/call",
  "params": {
    "name": "create_system",
    "arguments": {
      "project_root": "./test",
      "name": "Order Service",
      "description": "Handles order lifecycle",
      "responsibilities": ["Create orders", "Track fulfilment"],
      "primary_language": "Go",
      "tags": ["backend", "orders"]
    }
  }
}
```

**Pass criteria:**
- `result.system.id` = `"order-service"` (normalized)
- `result.system.name` = `"Order Service"`
- `result.system.tags` contains `"orders"`
- File `test/src/order-service/system.md` exists on disk after the call

**Cleanup:** Remove `test/src/order-service/` after the test run.

---

## T05 — create_container

**Tool:** `create_container`  
**Prerequisite:** T04 (order-service exists)

```json
{
  "jsonrpc": "2.0", "id": 5, "method": "tools/call",
  "params": {
    "name": "create_container",
    "arguments": {
      "project_root": "./test",
      "system_name": "Order Service",
      "name": "API Gateway",
      "description": "HTTP entry point",
      "technology": "Go + Fiber"
    }
  }
}
```

**Pass criteria:**
- `result.container.id` = `"api-gateway"`
- `result.container.technology` = `"Go + Fiber"`
- `result.container.diagram` is non-null (D2 file created)
- File `test/src/order-service/api-gateway/container.d2` exists on disk (implementation uses `container.d2` as the canonical name)

---

## T06 — create_component (single)

**Tool:** `create_component`  
**Prerequisite:** T05

```json
{
  "jsonrpc": "2.0", "id": 6, "method": "tools/call",
  "params": {
    "name": "create_component",
    "arguments": {
      "project_root": "./test",
      "system_name": "Order Service",
      "container_name": "API Gateway",
      "name": "Create Order Handler",
      "description": "Validates and persists new orders",
      "technology": "Go"
    }
  }
}
```

**Pass criteria:**
- `result.component.id` = `"create-order-handler"`
- File `test/src/order-service/api-gateway/create-order-handler/component.md` exists

---

## T07 — create_components (batch)

**Tool:** `create_components`  
**Prerequisite:** T05

```json
{
  "jsonrpc": "2.0", "id": 7, "method": "tools/call",
  "params": {
    "name": "create_components",
    "arguments": {
      "project_root": "./test",
      "system_name": "Order Service",
      "container_name": "API Gateway",
      "components": [
        { "name": "Auth Middleware", "technology": "Go" },
        { "name": "Rate Limiter", "technology": "Go" },
        { "name": "Request Logger", "technology": "Go" }
      ]
    }
  }
}
```

**Pass criteria:**
- `result.created` = 3
- `result.failed` = 0
- All three IDs present in `result.results`: `auth-middleware`, `rate-limiter`, `request-logger`

---

## T08 — create_relationship (container-level path)

**Tool:** `create_relationship`  
**Prerequisite:** T04 + T05  
**Purpose:** Tests container-level relationship creation (2-segment source/target paths).

```json
{
  "jsonrpc": "2.0", "id": 8, "method": "tools/call",
  "params": {
    "name": "create_relationship",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "source": "notification-service/api-gateway",
      "target": "notification-service/message-queue",
      "label": "Enqueue notification job",
      "type": "async",
      "technology": "AWS SQS"
    }
  }
}
```

**Pass criteria:**
- `result.relationship.id` is an 8-hex-char string
- `result.relationship.source` = `"notification-service/api-gateway"`
- `result.relationship.target` = `"notification-service/message-queue"`
- `result.relationship.type` = `"async"`
- `result.diagram_updated` = `true`
- File `test/src/notification-service/system.d2` contains an edge referencing `api-gateway` and `message-queue`

**Save the returned `relationship.id` for T15 (delete test).**

---

## T09 — create_relationship (component-level path)

```json
{
  "jsonrpc": "2.0", "id": 9, "method": "tools/call",
  "params": {
    "name": "create_relationship",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "source": "notification-service/processing-layer/email-sender",
      "target": "notification-service/delivery-channels/ses-email-service",
      "label": "Send email via SES",
      "type": "sync",
      "technology": "AWS SDK"
    }
  }
}
```

**Pass criteria:**
- `result.relationship.source` contains `"email-sender"`
- `result.relationship.target` contains `"ses-email-service"`
- `result.diagram_updated` = `true`
- `result.diagram_path` contains `"notification-service"` (written to system or container D2)

---

## T10 — create_relationship (invalid path — uppercase)

**Purpose:** Validates that invalid element paths are rejected.

```json
{
  "jsonrpc": "2.0", "id": 10, "method": "tools/call",
  "params": {
    "name": "create_relationship",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "source": "Notification-Service/API-Gateway",
      "target": "notification-service/message-queue",
      "label": "Should fail"
    }
  }
}
```

**Pass criteria:**
- Response contains `"error"` (tool error, not JSON-RPC parse error)
- Error message mentions `"did you mean"` or `"valid slug"` or `"lowercase"`

---

## T11 — list_relationships

**Tool:** `list_relationships`  
**Prerequisite:** T08 or pre-existing `relationships.toml` in fixture

```json
{
  "jsonrpc": "2.0", "id": 11, "method": "tools/call",
  "params": {
    "name": "list_relationships",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service"
    }
  }
}
```

**Pass criteria:**
- `result.count` ≥ 1
- `result.relationships` is a non-empty array
- Each entry has `id`, `source`, `target`, `label`
- `result.system` = `"notification-service"`

---

## T12 — query_dependencies (container level — THE KEY BUG FIX)

**Tool:** `query_dependencies`  
**Prerequisite:** T08 (relationship exists in TOML)  
**Purpose:** Verifies `dependency_count > 0` at container level — this was the primary bug.

```json
{
  "jsonrpc": "2.0", "id": 12, "method": "tools/call",
  "params": {
    "name": "query_dependencies",
    "arguments": {
      "project_root": "./test",
      "system_id": "notification-service",
      "container_id": "api-gateway"
    }
  }
}
```

**Pass criteria:**
- `result.container.id` = `"api-gateway"`
- `result.dependency_count` ≥ 1 — **this was 0 before the fix**
- `result.dependencies` is a non-empty array
- Each dependency has `id`, `name`, `type`, `level`

---

## T13 — query_dependencies (component level)

**Tool:** `query_dependencies`  
**Prerequisite:** T09

```json
{
  "jsonrpc": "2.0", "id": 13, "method": "tools/call",
  "params": {
    "name": "query_dependencies",
    "arguments": {
      "project_root": "./test",
      "system_id": "notification-service",
      "container_id": "processing-layer",
      "component_id": "email-sender"
    }
  }
}
```

**Pass criteria:**
- `result.component.id` = `"email-sender"`
- `result.relationship_count` ≥ 1 — **this was 0 before the fix**
- `result.dependencies` is a non-empty array

---

## T14 — query_related_components

**Tool:** `query_related_components`  
**Prerequisite:** T08 or T09

```json
{
  "jsonrpc": "2.0", "id": 14, "method": "tools/call",
  "params": {
    "name": "query_related_components",
    "arguments": {
      "project_root": "./test",
      "system_id": "notification-service",
      "container_id": "processing-layer",
      "component_id": "email-sender"
    }
  }
}
```

**Pass criteria:**
- `result.dependency_count` ≥ 1 OR `result.dependent_count` ≥ 1
- `result.dependencies` and `result.dependents` are arrays (may be empty individually)

---

## T15 — analyze_coupling

**Tool:** `analyze_coupling`  
**Prerequisite:** relationships exist in `relationships.toml` (T08 / T09 or fixture data)

```json
{
  "jsonrpc": "2.0", "id": 15, "method": "tools/call",
  "params": {
    "name": "analyze_coupling",
    "arguments": {
      "project_root": "./test",
      "system_id": "notification-service"
    }
  }
}
```

**Pass criteria:**
- `result.total_components` ≥ 1
- `result.isolated_components` is an array (length 0 is acceptable if all are connected)
- `result.central_components` is a non-empty object when relationships exist
- `result.highly_coupled_components` is an object (may be empty)

---

## T16 — analyze_coupling (whole project)

```json
{
  "jsonrpc": "2.0", "id": 16, "method": "tools/call",
  "params": {
    "name": "analyze_coupling",
    "arguments": {
      "project_root": "./test"
    }
  }
}
```

**Pass criteria:**
- `result.total_systems` ≥ 1
- `result.total_components` ≥ 1
- No error

---

## T17 — validate

**Tool:** `validate`

```json
{
  "jsonrpc": "2.0", "id": 17, "method": "tools/call",
  "params": {
    "name": "validate",
    "arguments": { "project_root": "./test" }
  }
}
```

**Pass criteria:**
- No JSON-RPC error
- `result.valid` is a boolean (either value is acceptable)
- When `relationships.toml` is non-empty: `result.report` does NOT list any isolated component findings (FR-012 suppression)

---

## T18 — validate_diagram

**Tool:** `validate_diagram`

```json
{
  "jsonrpc": "2.0", "id": 18, "method": "tools/call",
  "params": {
    "name": "validate_diagram",
    "arguments": {
      "d2_source": "api -> queue: \"Enqueue job\"\nqueue -> worker: \"Process job\"",
      "level": "container"
    }
  }
}
```

**Pass criteria:**
- `result.syntax_valid` = `true`
- `result.errors` is empty or null
- `result.suggestions` is an array

---

## T19 — validate_diagram (invalid syntax)

```json
{
  "jsonrpc": "2.0", "id": 19, "method": "tools/call",
  "params": {
    "name": "validate_diagram",
    "arguments": {
      "d2_source": "{ broken ::: syntax"
    }
  }
}
```

**Pass criteria:**
- `result.syntax_valid` = `false`
- `result.errors` is non-empty

---

## T20 — update_system

**Tool:** `update_system`  
**Prerequisite:** T04

```json
{
  "jsonrpc": "2.0", "id": 20, "method": "tools/call",
  "params": {
    "name": "update_system",
    "arguments": {
      "project_root": "./test",
      "system_name": "Order Service",
      "description": "Handles complete order lifecycle including fulfilment",
      "tags": ["backend", "orders", "updated"]
    }
  }
}
```

**Pass criteria:**
- `result.system.description` matches the new description
- `result.system.tags` contains `"updated"`
- `result.message` is non-empty

---

## T21 — update_diagram

**Tool:** `update_diagram`

```json
{
  "jsonrpc": "2.0", "id": 21, "method": "tools/call",
  "params": {
    "name": "update_diagram",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "d2_source": "# Notification System\ndirection: right\napi-gateway -> message-queue: \"Enqueue\"\nmessage-queue -> processing-layer: \"Process\""
    }
  }
}
```

**Pass criteria:**
- `result.success` = `true`
- `result.type` = `"system"`
- File `test/src/notification-service/system.d2` content matches the submitted D2 source

---

## T22 — search_elements

**Tool:** `search_elements`

```json
{
  "jsonrpc": "2.0", "id": 22, "method": "tools/call",
  "params": {
    "name": "search_elements",
    "arguments": {
      "project_root": "./test",
      "query": "*email*"
    }
  }
}
```

**Pass criteria:**
- Returns at least one result matching `email` in the ID or name
- Each result has `id`, `name`, `type`

---

## T23 — search_elements (by type filter)

```json
{
  "jsonrpc": "2.0", "id": 23, "method": "tools/call",
  "params": {
    "name": "search_elements",
    "arguments": {
      "project_root": "./test",
      "query": "*",
      "type": "container"
    }
  }
}
```

**Pass criteria:**
- All results have `type` = `"container"`
- Result count ≥ 1

---

## T24 — delete_relationship

**Tool:** `delete_relationship`  
**Prerequisite:** T08 (save the returned `relationship.id`)

```json
{
  "jsonrpc": "2.0", "id": 24, "method": "tools/call",
  "params": {
    "name": "delete_relationship",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "relationship_id": "<id-from-T08>"
    }
  }
}
```

**Pass criteria:**
- `result.deleted` = `true`
- `result.relationship_id` matches the submitted ID
- Subsequent `list_relationships` call no longer returns that ID

---

## T25 — delete_relationship (non-existent ID)

```json
{
  "jsonrpc": "2.0", "id": 25, "method": "tools/call",
  "params": {
    "name": "delete_relationship",
    "arguments": {
      "project_root": "./test",
      "system_name": "Notification Service",
      "relationship_id": "00000000"
    }
  }
}
```

**Pass criteria:**
- Response contains a tool error (non-existent ID rejected)
- Server does not crash

---

## T26 — build_docs

**Tool:** `build_docs`

```json
{
  "jsonrpc": "2.0", "id": 26, "method": "tools/call",
  "params": {
    "name": "build_docs",
    "arguments": {
      "project_root": "./test",
      "output_dir": "./test/dist-qa"
    }
  }
}
```

**Pass criteria:**
- `result.success` = `true`
- `result.systems` ≥ 1
- Directory `test/dist-qa/` exists and contains at least `index.html`

**Cleanup:** Remove `test/dist-qa/` after the test run.

---

## T27 — error handling: system not found with suggestion

**Purpose:** Verifies "did you mean?" suggestion on typo.

```json
{
  "jsonrpc": "2.0", "id": 27, "method": "tools/call",
  "params": {
    "name": "query_dependencies",
    "arguments": {
      "project_root": "./test",
      "system_id": "Notification Service",
      "container_id": "api-gateway"
    }
  }
}
```

**Pass criteria:**
- Response contains a tool error
- Error message contains `"did you mean"` and `"notification-service"` (slug suggestion)

---

## T28 — idempotent create_relationship

**Purpose:** Creating the same relationship twice returns the same ID (SHA-256 is deterministic).

Run T08 twice with identical arguments.

**Pass criteria:**
- Both calls succeed (no error)
- Both `result.relationship.id` values are identical
- `list_relationships` shows only one entry for that source/target/label triple

---

## Test Execution Order (Recommended)

```
T01 → T02 → T03                          # read-only: project & architecture queries
T04 → T05 → T06 → T07                   # create hierarchy
T08 → T09 → T10                          # create relationships (+ invalid)
T11                                       # list relationships
T12 → T13 → T14 → T15 → T16             # graph query tools (THE KEY FIX)
T17 → T18 → T19                          # validate
T20 → T21                                # update
T22 → T23                                # search
T24 → T25                                # delete relationships
T26                                       # build docs
T27 → T28                                # error handling + idempotency
```
