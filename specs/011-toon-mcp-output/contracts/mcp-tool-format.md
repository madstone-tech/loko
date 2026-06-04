# Contract: MCP Read Tool Format Parameter

**Feature**: TOON Output Format for MCP Query Tools  
**Date**: 2026-06-04  
**Applies to**: All MCP **read** tools (7 tools)

## Scope

This contract defines the `format` input parameter and the response shape for all MCP read tools.

## Affected Tools

| Tool | Type | Current Format Behavior |
|------|------|------------------------|
| `query_project` | Read | Returns `map[string]any` (JSON-RPC native) |
| `query_architecture` | Read | Already accepts `format`; values are `text`/`json`/`toon`/`compact` |
| `query_dependencies` | Read | Returns `map[string]any` (JSON-RPC native) |
| `query_related_components` | Read | Returns `map[string]any` (JSON-RPC native) |
| `search_elements` | Read | Returns `map[string]any` (JSON-RPC native) |
| `list_relationships` | Read | Returns `map[string]any` (JSON-RPC native) |
| `analyze_coupling` | Read | Returns `map[string]any` (JSON-RPC native) |

**Unaffected tools** (no `format` parameter added):
- All mutation tools: `create_system`, `create_container`, `create_component`, `create_relationship`, `create_components`, `update_system`, `update_container`, `update_component`, `update_diagram`, `delete_relationship`
- All validation tools: `validate`, `validate_diagram`
- All build tools: `build_docs`

## Input Schema Addition

Each affected tool's `InputSchema()` MUST include the `format` property:

```json
{
  "type": "object",
  "properties": {
    "project_root": {
      "type": "string",
      "description": "Root directory of the project (defaults to current)"
    },
    "format": {
      "type": "string",
      "enum": ["toon", "json"],
      "default": "toon",
      "description": "Output format: 'toon' for token-efficient LLM output (default), 'json' for human-readable debugging"
    }
  }
}
```

**`query_architecture` special case**: Its existing schema already has `format` but with values `["summary", "structure", "full"]` for the `detail` parameter and an unadvertised `format` field. The new schema will expose `format` explicitly with `enum: ["toon", "json"]` and retain `detail` separately.

## Response Contract

### When `format == "json"` (or `format` omitted on a tool that does not yet support TOON — backward compat)

The tool returns its data as a `map[string]any` directly. The MCP transport serializes this to JSON automatically.

```json
{
  "project": { "name": "MyApp", "description": "...", "version": "1.0.0" },
  "stats": { "systems": 3, "containers": 9, "components": 18 }
}
```

### When `format == "toon"` (default for all read tools)

The tool returns a wrapper map containing the TOON-encoded payload:

```json
{
  "payload": "<toon-encoded-string>",
  "format": "toon",
  "token_estimate": 247
}
```

The `payload` string is the TOON serialization of the same data that would appear in the JSON response. The `token_estimate` is `len(payload) / 4` (rounded up), using the project's documented approximation.

### Error Contract

**Invalid format value**:

```json
{
  "error": "invalid format \"xml\": expected \"toon\" or \"json\""
}
```

**Tool execution failure** (unchanged from today):

```json
{
  "error": "failed to load project: ..."
}
```

## Backward Compatibility

- **Breaking change**: Read tools that previously returned JSON maps now return TOON-wrapped payloads by default. Existing MCP clients must pass `format: "json"` to restore previous behavior.
- **Mitigation**: This is an intentional, documented change. The spec explicitly states this is the desired behavior.
- **`query_architecture` legacy**: The deprecated `"compact"` and `"text"` format values are still accepted internally but are no longer advertised in the input schema. `"text"` maps to `"toon"` behavior.

## Testing Contract

### Unit Tests (per tool)

Each affected tool MUST have tests covering:
1. Default call (no `format`) → returns TOON-wrapped response
2. `format: "toon"` → returns TOON-wrapped response
3. `format: "json"` → returns JSON map directly
4. `format: "invalid"` → returns error with allowed values

### Integration Tests

`tests/mcp/tool_format_test.go` MUST exercise all 7 tools end-to-end against the canonical test project.

### Benchmark Contract

`tests/benchmarks/token_efficiency_test.go` MUST:
1. Generate representative payloads for all 7 read tools
2. Measure JSON token count and TOON token count per payload
3. Report per-payload and aggregate percentage reduction
4. Fail (non-zero exit) if aggregate reduction < 30%
