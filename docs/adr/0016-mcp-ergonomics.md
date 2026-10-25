# ADR-0016: One language table, compact MCP diagnostics, and warning controls

**Status:** Accepted
**Date:** 2026-10-25
**Feature:** [017-mcp-validation-ergonomics](../../specs/017-mcp-validation-ergonomics/spec.md)
**Related:** [ADR-0011](0011-toon-mcp-default.md) (TOON for MCP),
[ADR-0014](0014-hcl-authoring.md) (assistants edit HCL),
[ADR-0015](0015-rendering-attributes.md) (attributes added in 016)

## Context

An assistant modelled a 12-container system through `loko mcp` (1.0.1) and reported its session.
The model compiled, but:

- the assistant learned the language from refusals: about eight refused `apply_edit` previews
  before it knew which attributes each block takes, and it never found `external`;
- 33 warnings came out, most of them expected: third-party systems with no containers or prose,
  and instances whose infrastructure did not exist yet;
- every MCP diagnostic repeated a seven-field source range, which assistants do not use;
- in an empty directory, nothing could create the first file through MCP.

## Decision

1. **One language table.** `arch/schemas.go` holds every block, its labels, where it nests, its
   attributes (type, required, allowed values) and a minimal example. The decoder derives its
   attribute lists from it. `authoring` keeps its own copy, because entity packages may not import
   one another, and a parity test fails when the two disagree. The read-only `schema` MCP tool
   serves the same table, so the compiler, the refusal messages and the tool cannot drift apart.
   The handshake instructions tell clients to call it first.

2. **Compact MCP diagnostics by default.** Every MCP tool returns diagnostics as severity, code,
   summary, detail, address and `file:line`. `ranges: "full"` returns the 1.0.2 shape. CLI JSON
   output is unchanged. This changes MCP output within 1.x. We accept that because the only MCP
   clients are assistants, they act on a file and a line, and the full shape stays one argument
   away. Constitution v1.5.0 (Principle VI) records the rule.

3. **`instance.planned`.** A bool. A planned instance without a binding raises no
   `unbound_instance`; a planned instance with a binding raises `planned_instance_bound`, so the
   marker cannot outlive the infrastructure it describes. We chose a bool over a `lifecycle`
   vocabulary, which nothing yet needs, and over a magic tag, which `schema` could not describe.

4. **Ignoring warning codes.** `project.ignore_warnings` and a per-run `--ignore` (or `ignore` in
   MCP) remove named warning codes from output and from `--strict`. Suppression is the last step of
   compilation, so `validate`, `build`, `serve` and MCP all see the same diagnostics. Errors cannot
   be ignored. A per-run unknown code is refused; an unknown code in the source is a warning
   (`unknown_warning_code`), so a release that retires a code does not break a project.

5. **A `project` edit target instead of an `init` tool.** Creating the project declaration is an
   `apply_edit` with target `project`, so it gets the same preview, revision check and atomic
   write as every other edit. In an empty directory, `describe` returns the request to send.

`instance.planned` and `project.ignore_warnings` are permanent v1.x attributes.

## Consequences

**Positive:**
- A client can learn the whole language in one call, and the answer cannot drift from the compiler.
- Expected warnings can be stated in the model, so the warnings that remain are worth reading.
- MCP responses carrying diagnostics are less than half their 1.0.2 size.

**Negative:**
- A client that read `range` from MCP diagnostics must now pass `ranges: "full"`.
- Two copies of the attribute table exist (`arch` and `authoring`).

**Mitigations:**
- The changelog records the MCP output change, with the argument that restores the old shape.
- The parity test runs on every `go test`, so the two copies cannot disagree unnoticed.
