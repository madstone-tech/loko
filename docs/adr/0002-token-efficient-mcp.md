# ADR 0002: Token-Efficient MCP Queries

## Status

Accepted

## Context

When an LLM agent designs architecture with loko via MCP, it needs context about the existing project. For large projects (30+ systems), sending everything consumes:
- Excessive tokens (cost)
- Context window space (limiting conversation)
- Processing time (latency)

## Decision

Implement progressive context loading with three detail levels:

### Summary (~200 tokens for 20-system project)
```json
{
  "project": "payment-platform",
  "systems": 4,
  "containers": 12,
  "systems_list": ["PaymentService", "OrderService", ...]
}
```

### Structure (~500 tokens)
```json
{
  "systems": {
    "PaymentService": {
      "containers": ["API", "Database", "Worker"],
      "external_dependencies": ["StripeAPI"]
    }
  }
}
```

### Full (targeted, variable)
Complete details for a specific system or container, optionally including D2 diagram source.

### API Design

```
query_architecture(
  scope: "project" | "system" | "container",
  target: string,           // For specific entity
  detail: "summary" | "structure" | "full",
  include_diagrams: bool    // D2 source code
)
```

## Consequences

**Positive:**
- 10x reduction in token usage for typical queries
- LLM can progressively drill down as needed
- Faster response times
- More room in context window for conversation

**Negative:**
- More complex MCP tool implementation
- LLM must learn to use detail levels effectively
- Multiple round trips for deep exploration

**Mitigations:**
- Clear tool description with examples
- Default to "summary" — always fast
- Compressed notation option for power users
