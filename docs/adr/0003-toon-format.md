# ADR 0003: TOON Format Support

## Status

Accepted (Implementation: v0.2.0)

## Context

loko's MCP interface sends architecture data to LLMs. Token consumption directly impacts cost and context window usage. We already implement progressive context loading (ADR 0002), but can optimize further.

TOON (Token-Oriented Object Notation) is a compact format designed for LLM input that achieves 30-60% token reduction for uniform arrays — exactly the structure of architecture data.

### Example: JSON vs TOON

**JSON (~380 tokens)**
```json
{
  "systems": [
    {"name": "PaymentService", "containers": ["API", "DB"]},
    {"name": "OrderService", "containers": ["API", "DB"]}
  ]
}
```

**TOON (~220 tokens)**
```
systems[2]{name,containers}:
  PaymentService,API|DB
  OrderService,API|DB
```

## Decision

Support TOON as an optional output format for MCP queries:

1. Add `format: "json" | "toon"` parameter to `query_architecture` tool
2. Default to JSON for maximum compatibility
3. Use official `toon-format/toon-go` library
4. Include format hint in response when TOON is used

### Implementation

```go
// New port
type OutputEncoder interface {
    Encode(data any) ([]byte, error)
    ContentType() string
    FormatHint() string
}

// Adapters
- internal/adapters/encoding/json_encoder.go  (default)
- internal/adapters/encoding/toon_encoder.go  (optional)
```

## Consequences

**Positive:**
- Additional 30-40% token reduction on top of progressive loading
- Official Go library available and maintained
- Aligns with token-efficiency design philosophy
- Compound effect: progressive loading + TOON = 60-85% total reduction

**Negative:**
- Additional dependency
- Not all LLMs familiar with TOON format
- Requires format hint in tool description

**Mitigations:**
- Make TOON opt-in, not default
- Provide clear format hints in responses
- Document when to use each format
- Benchmark and report real-world savings

## References

- https://toonformat.dev/
- https://github.com/toon-format/toon-go
