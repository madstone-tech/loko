# Contract: `loko query`

**Feature**: `015-mcp-hcl-authoring` | **Status**: normative

```
loko query dependents   <address> [--transitive] [--format text|json|toon]
loko query dependencies <address> [--transitive] [--format text|json|toon]
loko query path         <from> <to>              [--format text|json|toon]
loko query orphans                               [--format text|json|toon]
loko query coupling     [--limit 20]             [--format text|json|toon]
```

- `--project/-p` selects the project root, as for every command.
- Each subcommand calls `usecases.Query`, the same function the MCP `query` tool calls. The
  `--format json` output is byte-identical to the MCP tool's JSON for the same inputs (FR-029,
  SC-006).
- `text` (the default) is a compact table for people; `json` and `toon` are for tools and pipes.
- An unknown address prints `unknown address "container.ap": did you mean container.api?` and
  exits `1` (FR-030).
- A project that does not compile prints the diagnostics as `validate` does and exits `1` (FR-007).
- Exit codes are the existing three; `2` is not used, because queries have no warnings mode
  (FR-031).
- Handler functions are ≤ 50 effective lines (FR-032).
