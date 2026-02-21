# MCP Tool Contracts: Relationship Management

**Feature**: 008-mcp-ux-improvements  
**Date**: 2026-02-19

These contracts define the input schema, output shape, and error behaviour for the three new relationship MCP tools.

---

## `create_relationship`

### Input Schema

```json
{
  "type": "object",
  "required": ["project_root", "system_name", "source", "target", "label"],
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the loko project (default: '.')"
    },
    "system_name": {
      "type": "string",
      "description": "Name of the system owning this relationship (slugified or display name)"
    },
    "source": {
      "type": "string",
      "description": "Source element path, e.g. 'agwe/api-lambda' or 'agwe/api-lambda/request-validator'"
    },
    "target": {
      "type": "string",
      "description": "Target element path, e.g. 'agwe/sqs-queue'"
    },
    "label": {
      "type": "string",
      "description": "Human-readable description of the relationship, e.g. 'Enqueue email job'"
    },
    "type": {
      "type": "string",
      "enum": ["sync", "async", "event"],
      "description": "Communication type (default: 'sync')"
    },
    "technology": {
      "type": "string",
      "description": "Technology used (e.g., 'AWS SDK SQS', 'gRPC')"
    },
    "direction": {
      "type": "string",
      "enum": ["forward", "bidirectional"],
      "description": "Arrow direction (default: 'forward')"
    }
  }
}
```

### Success Response

```json
{
  "relationship": {
    "id": "a1b2c3d4",
    "source": "agwe/api-lambda",
    "target": "agwe/sqs-queue",
    "label": "Enqueue email job",
    "type": "async",
    "technology": "AWS SDK SQS",
    "direction": "forward"
  },
  "diagram_updated": true,
  "diagram_path": "src/agwe/system.d2"
}
```

### Error Cases

| Condition | Error message |
|---|---|
| `source` missing or empty | `"source is required"` |
| `target` missing or empty | `"target is required"` |
| `label` missing or empty | `"label is required"` |
| `source == target` | `"source and target must be different elements"` |
| `type` invalid value | `"type must be one of: sync, async, event"` |
| `direction` invalid value | `"direction must be one of: forward, bidirectional"` |
| `source` not found in project | `"container \"api-lambda\" not found — did you mean \"api-lambda\"?"` (slug suggestion applied) |
| `target` not found in project | `"container \"SQS Queue\" not found — did you mean \"sqs-queue\"?"` |
| Duplicate relationship (same source+target+label) | Returns existing relationship (idempotent, no error) |

---

## `list_relationships`

### Input Schema

```json
{
  "type": "object",
  "required": ["project_root", "system_name"],
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the loko project"
    },
    "system_name": {
      "type": "string",
      "description": "System to list relationships for"
    },
    "source": {
      "type": "string",
      "description": "Optional: filter to relationships where source matches this path"
    },
    "target": {
      "type": "string",
      "description": "Optional: filter to relationships where target matches this path"
    }
  }
}
```

### Success Response

```json
{
  "system": "agwe",
  "count": 3,
  "relationships": [
    {
      "id": "a1b2c3d4",
      "source": "agwe/api-lambda",
      "target": "agwe/sqs-queue",
      "label": "Enqueue email job",
      "type": "async",
      "technology": "AWS SDK SQS",
      "direction": "forward"
    },
    {
      "id": "e5f6a7b8",
      "source": "agwe/worker-lambda",
      "target": "agwe/ses",
      "label": "Send email via SES",
      "type": "sync",
      "direction": "forward"
    }
  ]
}
```

**Empty project** (no `relationships.toml` yet):
```json
{
  "system": "agwe",
  "count": 0,
  "relationships": []
}
```

### Error Cases

| Condition | Error message |
|---|---|
| `system_name` not found | `"system \"Agwe\" not found — did you mean \"agwe\"?"` |
| Project root invalid | `"failed to load project: ..."` |

---

## `delete_relationship`

### Input Schema

```json
{
  "type": "object",
  "required": ["project_root", "system_name", "relationship_id"],
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the loko project"
    },
    "system_name": {
      "type": "string",
      "description": "System owning this relationship"
    },
    "relationship_id": {
      "type": "string",
      "description": "ID of the relationship to delete (from create_relationship or list_relationships response)"
    }
  }
}
```

### Success Response

```json
{
  "deleted": true,
  "relationship_id": "a1b2c3d4",
  "diagram_updated": true,
  "diagram_path": "src/agwe/system.d2"
}
```

### Error Cases

| Condition | Error message |
|---|---|
| `relationship_id` not found | `"relationship \"a1b2c3d4\" not found"` |
| `system_name` not found | `"system \"Agwe\" not found — did you mean \"agwe\"?"` |

---

## Cache Invalidation (all three tools)

All three tools call `graphCache.Invalidate(projectRoot)` immediately after a successful operation. The next call to `query_dependencies` or `query_related_components` will trigger a graph rebuild from both `relationships.toml` and D2 files, reflecting the change within the same MCP session.

---

## D2 Diagram Update Behaviour

| Relationship level | Diagram updated |
|---|---|
| Container → Container | `src/<system-id>/system.d2` — edges section regenerated |
| Component → Component (same container) | `src/<system-id>/<container-id>/container.d2` — edges section regenerated |
| Component → Component (cross-container) | `src/<system-id>/system.d2` — edges section regenerated |

Edge syntax generated per type:
- `sync`: `source -> target: "label"`
- `async`: `source -> target: { label: "label"; style.animated: true }`
- `event`: `source -> target: { label: "label"; style.stroke-dash: 5 }`
- `bidirectional`: replace `->` with `<->`
