# ADR 0011: TOON as Default Format for MCP Read Tools

## Status

Proposed

## Context

Constitution Principle VI (Token Efficiency) states:

> "Default to JSON for compatibility; TOON is opt-in"

This principle was established in ADR 0003 when TOON was first introduced. At that time, TOON was added as an optional format to the `query_architecture` tool with JSON as the default, because:
1. TOON was a new, unfamiliar format
2. Existing MCP clients consumed JSON
3. The change was additive — no client behavior changed

Feature 011 proposes making TOON the default output format for all seven MCP read tools (`query_project`, `query_architecture`, `query_dependencies`, `query_related_components`, `search_elements`, `list_relationships`, `analyze_coupling`). This reverses the default for read tools: TOON becomes default, JSON becomes opt-in via `format: "json"`.

This directly conflicts with Principle VI as currently written.

### Why the conflict is justified

The Principle VI default was scoped to **file output and CLI output** — formats that are consumed by humans, scripts, and external tools that expect stable serialization. MCP read-tool responses, by contrast, are:

1. **Internal to the LLM↔MCP server interaction** — never written to disk, never consumed by external systems
2. **Ephemeral** — each response is consumed immediately by the calling LLM and discarded
3. **Volume-dominant** — read tools are the highest-frequency MCP calls; their responses dominate token spend
4. **Already negotiated** — the MCP transport is JSON-RPC; the *content* of the `result` field is what changes, not the transport encoding

The original opt-in design in ADR 0003 was conservative because TOON was experimental. Since then, TOON has proven stable, the `toon-go` library is a declared dependency, and the `OutputEncoder` interface has been in production use. The risk of TOON as default for MCP read-tool *content* is materially lower than for file/CLI output.

### What does NOT change

This amendment does NOT change the default for:
- CLI output (`loko build --format` defaults remain unchanged)
- File output (architecture documents written to disk)
- Mutation tool responses (still JSON)
- Validation tool responses (still JSON)
- HTTP API responses (JSON default retained)

## Decision

Amend Principle VI to narrow the "JSON default" scope:

> "Default to JSON for compatibility in file output, CLI output, and HTTP API responses. TOON is the default for MCP read-tool responses. TOON is opt-in for all other contexts."

This makes the default context-dependent rather than global. The reasoning:
1. MCP read tools are LLM-facing; token efficiency is the primary optimization target
2. JSON remains available per-call via `format: "json"` for debugging
3. All other interfaces (CLI, file, HTTP API, mutation tools) keep JSON default to preserve external compatibility

## Consequences

**Positive:**
- Read-tool token reduction of 30-40% applies automatically to all LLM interactions
- No configuration change required to realize the benefit
- Context-dependent default is more precise than a global rule

**Negative:**
- Existing MCP clients that consume read-tool responses must update (pass `format: "json"` to restore JSON)
- Principle VI text is now context-dependent, slightly more complex

**Mitigations:**
- Document the breaking change in release notes
- Update `docs/mcp-integration.md` with migration guidance
- Add `format: "json"` example to tool descriptions
- The `format` parameter is discoverable in each tool's `InputSchema()`

## References

- ADR 0002: Token-Efficient MCP Responses
- ADR 0003: TOON Format Support
- Constitution Principle VI: Token Efficiency
- Feature Spec: `specs/011-toon-mcp-output/spec.md`
