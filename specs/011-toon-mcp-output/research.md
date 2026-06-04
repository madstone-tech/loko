# Research: TOON Output Format for MCP Query Tools

**Feature**: TOON Output Format for MCP Query Tools  
**Date**: 2026-06-04

## Decisions

### Decision 1: Default Format = TOON for Read Tools Only

**Decision**: All seven MCP read tools default to TOON. Mutation and validation tools remain JSON-only and ignore any `format` argument.

**Rationale**:
- Read tools return large, repetitive payloads (arrays of systems/containers/components) where TOON's compact delimiters and abbreviated keys provide maximum benefit.
- Mutation/validation tools return small confirmation/error payloads where JSON is already minimal and is relied upon by downstream integrations.
- The spec explicitly scopes format changes to read tools (FR-005).

**Alternatives considered**:
- Global default toggle (env var / config) — Rejected: per-call granularity matches MCP tool semantics and lets a single client mix formats.
- Default JSON with opt-in TOON — Rejected: contradicts the feature's purpose; the savings only materialize if TOON is the default for the high-volume read path.

### Decision 2: Reuse Existing `Encoder` in `adapters/encoding/`

**Decision**: Use the existing `Encoder` struct (already implements `usecases.OutputEncoder`) rather than creating a new adapter.

**Rationale**:
- `Encoder.EncodeJSON()` and `Encoder.EncodeTOON()` already exist and are tested.
- `toon-go` is already a dependency in `go.mod`.
- The adapter pattern (interface in `ports.go`, implementation in `adapters/`) already matches the architecture.
- No new dependency or abstraction needed.

**Alternatives considered**:
- New `MCPResponseEncoder` interface — Rejected: YAGNI. The existing `OutputEncoder` is sufficient.
- Inline TOON formatting in each tool — Rejected: violates DRY and makes format switching brittle.

### Decision 3: Handler-Level Format Dispatch (Not Use-Case Level)

**Decision**: Each MCP tool handler parses the `format` argument and decides whether to wrap the response in TOON. Use cases remain unchanged.

**Rationale**:
- Format selection is a presentation concern, not a business-logic concern.
- Use cases already return `map[string]any` (or structs that map to `map[string]any`).
- Keeping format logic in the thin handler layer preserves Clean Architecture separation.
- `query_architecture` already follows this pattern with `ExecuteWithFormat`.

**Alternatives considered**:
- Add format parameter to every use case — Rejected: bloats core layer with presentation concerns.
- Middleware/wrapper around all tools — Rejected: adds indirection for a simple 3-line dispatch.

### Decision 4: TOON Response Wrapper Shape

**Decision**: When `format == "toon"`, the tool returns:

```go
map[string]any{
    "payload":        string(toonBytes),
    "format":         "toon",
    "token_estimate": len(toonBytes) / 4,
}
```

**Rationale**:
- The MCP transport requires JSON-RPC framing. The actual TOON payload must be a JSON string value inside the RPC response.
- Including `"format": "toon"` makes the response self-describing for the client.
- Including `"token_estimate"` helps LLM clients budget context.
- This matches the existing `query_architecture` response shape (which already has `"text"` and `"token_estimate"`).

**Alternatives considered**:
- Return raw TOON bytes as the top-level response — Rejected: breaks JSON-RPC transport contract.
- Embed TOON in a `"text"` key (like `query_architecture` currently does) — Rejected: `"payload"` is clearer and more generic across tools.

### Decision 5: Benchmark Suite as Go Test with Gating

**Decision**: Implement the benchmark as a Go test (`tests/benchmarks/token_efficiency_test.go`) that exercises all 7 read tools against representative payloads and fails if aggregate reduction < 30%.

**Rationale**:
- Go tests are the project's existing test framework.
- `testing.T` failure naturally gates CI (`go test ./...`).
- Reuses existing test fixtures (`createTestProject` in `toon_benchmark_test.go`).
- The existing `scripts/benchmark-token-efficiency.sh` measures build output, not MCP tool responses; the new benchmark fills that gap.

**Alternatives considered**:
- Standalone benchmark binary — Rejected: `go test -bench` is sufficient and integrates with CI.
- Shell script wrapping MCP tool calls — Rejected: slower, harder to maintain, no access to in-memory fixtures.

### Decision 6: Token Estimation = Character Count / 4

**Decision**: Continue using the project's existing approximation: 1 token ≈ 4 characters for structured data.

**Rationale**:
- Already documented in `scripts/benchmark-token-efficiency.sh` and `research/token-efficiency-benchmarks.md`.
- The gate is expressed as a percentage reduction, so the absolute estimator choice cancels out.
- Adding a real tokenizer (e.g., tiktoken) would add a Python dependency or a large Go module.

**Alternatives considered**:
- Use actual GPT tokenizer (tiktoken/cl100k_base) — Rejected: adds heavy dependency for a gate that only needs relative comparison.

## Open Questions Resolved

1. **Should `query_architecture`'s existing `format` field be changed?**
   - Resolved: Yes. It currently accepts `"text"`, `"json"`, `"toon"`, `"compact"`. The advertised schema will be updated to `enum: ["toon", "json"]` with `"toon"` as default. `"text"` and `"compact"` remain accepted internally for backward compatibility but map to `"toon"` behavior.

2. **Should mutation tools silently ignore `format` or error?**
   - Resolved: Silently ignore. The `format` parameter is not in their input schemas, so an MCP client cannot pass it without a schema violation. If passed anyway (protocol-level), it is ignored.

3. **Should the TOON payload be pretty-printed or minified?**
   - Resolved: Minified. `toon.Marshal` produces compact output. Pretty-printing would increase token count and defeat the purpose.
