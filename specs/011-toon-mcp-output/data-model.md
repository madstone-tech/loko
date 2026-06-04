# Data Model: TOON Output Format for MCP Query Tools

**Feature**: TOON Output Format for MCP Query Tools  
**Date**: 2026-06-04

## Overview

This feature introduces no new domain entities. It changes how existing entity data is serialized and returned from MCP tools. The "data model" here is the response envelope and the format selector.

## Format Selector

The format selector is a per-call parameter, not a persistent entity.

| Field | Type | Allowed Values | Default | Description |
|-------|------|----------------|---------|-------------|
| `format` | `string` | `"toon"`, `"json"` | `"toon"` | Output serialization format for read tools |

**Validation Rules**:
- Must be exactly `"toon"` or `"json"` (case-sensitive)
- Empty string or missing key → defaults to `"toon"`
- Any other value → returns error: `invalid format "{value}": expected "toon" or "json"`

## Response Envelope

All MCP tool responses are wrapped in a `map[string]any` (for JSON-RPC content). The format change affects the **values** inside this map, not the envelope structure.

### Read Tool Response Structure (Format-Agnostic)

Each read tool returns a `map[string]any` with tool-specific keys. Examples:

**`query_project`**:
```go
map[string]any{
    "project": map[string]any{"name": ..., "description": ..., "version": ...},
    "stats":   map[string]any{"systems": ..., "containers": ..., "components": ...},
}
```

**`query_architecture`**:
```go
map[string]any{
    "text":           string,        // human-readable or TOON payload
    "token_estimate": int,         // approximate token count
    "format":         string,        // "toon" or "json"
}
```

**`query_dependencies`**:
```go
map[string]any{
    "container_id": string,
    "dependencies": []map[string]any,
    "paths":        []map[string]any,
}
```

**`query_related_components`**:
```go
map[string]any{
    "component_id":     string,
    "dependencies":     []map[string]any,
    "dependents":       []map[string]any,
    "dependency_count": int,
    "dependent_count":  int,
}
```

**`search_elements`**:
```go
map[string]any{
    "query":   string,
    "results": []map[string]any,
    "count":   int,
}
```

**`list_relationships`**:
```go
map[string]any{
    "system":        string,
    "count":         int,
    "relationships": []map[string]any,
}
```

**`analyze_coupling`**:
```go
map[string]any{
    "system":         string,
    "metrics":        map[string]any,
    "highly_coupled": []map[string]any,
    "central_nodes":  []map[string]any,
}
```

### Format Transformation

When `format == "toon"`, the inner `map[string]any` value is serialized via `Encoder.EncodeTOON()` and the result is returned as a string inside the response envelope:

```go
// JSON path (format == "json"): returns the map directly
return resultMap, nil

// TOON path (format == "toon"): returns the map serialized as TOON string
payload, _ := encoder.EncodeTOON(resultMap)
return map[string]any{
    "payload":        string(payload),
    "format":         "toon",
    "token_estimate": estimateTokenCount(string(payload)),
}, nil
```

## Benchmark Payload Set

The benchmark operates on representative payloads drawn from:

1. **Canonical test project** — the project used by existing tests (`internal/core/usecases/testdata/` or in-memory fixtures)
2. **Example projects** referenced in `scripts/benchmark-token-efficiency.sh`:
   - `simple-project` (1 system, 2 containers, 4 components)
   - `3layer-app` (3 systems, 9 containers, 18 components)
   - `serverless` (5 systems, 15 containers, 30 components)
   - `microservices` (10 systems, 30 containers, 60 components)

For each payload, the benchmark:
1. Builds the response map as the tool would
2. Encodes to JSON → measures token count
3. Encodes to TOON → measures token count
4. Computes percentage reduction

## Relationships

- `Format Selector` → `OutputEncoder` (adapters/encoding): the encoder implements the serialization
- `Format Selector` → `MCP Tool Handler` (internal/mcp/tools): the handler parses the selector and dispatches
- `Benchmark Payload Set` → `Test Fixtures`: payloads are generated from existing test fixtures or example projects
