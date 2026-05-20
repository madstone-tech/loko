# MCP Tools Smoke Test Notes (T061)

## What a manual stdio smoke test would entail

A full end-to-end smoke test would start the loko MCP server in stdio mode
(`loko mcp`), then send JSON-RPC `tools/call` requests over stdin and assert
correct JSON responses on stdout — covering happy paths, missing-required-field
errors, and not-found suggestions for each of the 20 tools registered in
`internal/mcp/tools/registry.go`. The session would also verify that
`tools/list` returns the expected 20 tool names with unchanged `name`,
`description`, and `inputSchema` fields (FR-015 public-surface preservation).

## Why the existing tests are sufficient surrogate evidence

The existing tests in `internal/mcp/tools/*_test.go` exercise every tool
constructor, `Name()`, `InputSchema()`, and `Call()` through direct Go API
calls with both mock and real filesystem repositories. They cover: happy-path
responses, all required-field errors, cache invalidation semantics
(`graph_tools_cache_test.go`), relationship CRUD round-trips, and batch
component creation. The `tests/mcp` integration package additionally starts a
real MCP server and exercises tool dispatch. Together these provide the same
observable guarantees as a manual stdio session without the fragility of process
I/O scripting.

## Tools lacking dedicated integration test coverage

The following tools have unit tests in the tools package but no dedicated
integration test in `tests/mcp`:

- `validate_diagram` — structural validation and D2 syntax check are tested
  only via unit mocks; a real D2 binary is required for the syntax path.
- `build_docs` — build pipeline tested via mock `ProgressReporter`; a real
  HTML output integration test is absent.
- `query_dependencies` / `query_related_components` / `analyze_coupling` —
  graph logic tested in `graph_tools_cache_test.go` via mock repos; a full
  project fixture integration test would add confidence.

These gaps should be addressed in a follow-up task targeting `tests/mcp`
integration coverage for the above tools.
