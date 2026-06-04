# Quickstart: TOON Output Format for MCP Query Tools

**Feature**: TOON Output Format for MCP Query Tools  
**Date**: 2026-06-04

## What Changes

All MCP **read** tools now return responses in TOON format by default. Pass `format: "json"` to get human-readable JSON for debugging.

**Read tools affected** (7):
- `query_project`
- `query_architecture`
- `query_dependencies`
- `query_related_components`
- `search_elements`
- `list_relationships`
- `analyze_coupling`

**Tools unchanged** (mutation + validation):
- All `create_*`, `update_*`, `delete_*` tools
- `validate`, `validate_diagram`

## Calling a Read Tool

### Default (TOON)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "query_project",
    "arguments": {
      "project_root": "/path/to/project"
    }
  }
}
```

Response:
```json
{
  "payload": "project:[#1]:{name:MyApp,description:...,version:1.0.0} stats:{systems:3,containers:9,components:18}",
  "format": "toon",
  "token_estimate": 42
}
```

### Explicit JSON (for debugging)

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "query_project",
    "arguments": {
      "project_root": "/path/to/project",
      "format": "json"
    }
  }
}
```

Response:
```json
{
  "project": { "name": "MyApp", "description": "...", "version": "1.0.0" },
  "stats": { "systems": 3, "containers": 9, "components": 18 }
}
```

### Explicit TOON (same as default)

```json
{
  "name": "query_project",
  "arguments": {
    "project_root": "/path/to/project",
    "format": "toon"
  }
}
```

## Invalid Format

```json
{
  "name": "query_project",
  "arguments": {
    "project_root": "/path/to/project",
    "format": "xml"
  }
}
```

Response (error):
```json
{
  "error": "invalid format \"xml\": expected \"toon\" or \"json\""
}
```

## Running the Benchmark

Verify the token-reduction claim:

```bash
# Run the benchmark suite
go test ./tests/benchmarks/ -bench=BenchmarkTokenEfficiencyGate -v

# Or use the existing script (measures build output, not MCP tool responses)
./scripts/benchmark-token-efficiency.sh
```

Expected output (benchmark):
```
=== RUN   TestTokenEfficiencyGate
    token_efficiency_test.go:87: query_project: JSON=124 tokens, TOON=78 tokens, reduction=37.1%
    token_efficiency_test.go:87: query_architecture: JSON=512 tokens, TOON=312 tokens, reduction=39.1%
    ...
    token_efficiency_test.go:95: Aggregate reduction: 35.2%
--- PASS: TestTokenEfficiencyGate (0.42s)
```

The benchmark fails (non-zero exit) if the aggregate reduction across all 7 read-tool payloads drops below 30%.

## Updating Existing MCP Clients

If you have an MCP client that depends on JSON responses from read tools:

1. Add `"format": "json"` to all read tool calls, OR
2. Update your client to parse the TOON `payload` field (use `toon.Unmarshal` from `github.com/toon-format/toon-go`)

Migration example (Python):
```python
# Before (implicitly JSON)
result = call_tool("query_project", {"project_root": "."})
print(result["project"]["name"])

# After (explicit JSON for backward compat)
result = call_tool("query_project", {"project_root": ".", "format": "json"})
print(result["project"]["name"])
```

## Verification Checklist

- [ ] Call each read tool without `format` → response has `"format": "toon"` wrapper
- [ ] Call each read tool with `format: "json"` → response is plain JSON map
- [ ] Call a read tool with `format: "invalid"` → clear error returned
- [ ] Call a mutation tool → response is unchanged (JSON, no `format` parameter)
- [ ] Run benchmark → aggregate reduction ≥ 30%
