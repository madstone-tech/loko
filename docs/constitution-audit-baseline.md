# Constitution Audit Baseline - v0.2.0

**Generated**: 2026-02-13  
**Purpose**: Document current handler violations before Phase 1 refactoring  
**Target**: Phase 5 (User Story 5) will address all violations

This baseline captures the current state of handler compliance with the constitution's thin-handler principle:
- **CLI handlers**: < 50 lines (excluding imports/comments/blank lines)
- **MCP tools**: < 30 lines (excluding imports/comments/blank lines)

**Summary**:
- CLI handlers: 10 violations
- MCP tools: 16 violations
- **Total**: 26 violations

> **Historical record.** This baseline was taken on the v0.2 code base. Several files below
> (`cmd/api.go`, `cmd/new*.go`, `cmd/watch.go`, the `create_*`/`update_*` MCP tools) belong to
> commands and tools removed in v1.0. The current rules and how to run the check are in
> [architecture/constitution-compliance.md](architecture/constitution-compliance.md).

---

Constitution Audit: Handler Thin-Line Validation
==================================================

Checking CLI handlers (cmd/*.go, limit: 50 lines)...
- cmd/api.go: 57 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/build_cobra.go: 78 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/build.go: 149 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/export_cobra.go: 51 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/mcp.go: 56 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/new_cobra.go: 155 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/new.go: 184 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/root.go: 110 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/serve.go: 58 lines (limit: 50) - DOCUMENTED VIOLATION
- cmd/watch.go: 95 lines (limit: 50) - DOCUMENTED VIOLATION

Checking MCP tool handlers (internal/mcp/tools/*.go, limit: 30 lines)...
- internal/mcp/tools/build_docs.go: 78 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/create_component.go: 94 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/create_container.go: 91 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/create_system.go: 142 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/graph_tools.go: 288 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/helpers.go: 86 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/query_architecture.go: 76 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/query_project.go: 49 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/registry.go: 54 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/schemas.go: 185 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/update_component.go: 96 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/update_container.go: 87 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/update_diagram.go: 85 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/update_system.go: 130 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/validate_diagram.go: 69 lines (limit: 30) - DOCUMENTED VIOLATION
- internal/mcp/tools/validate.go: 56 lines (limit: 30) - DOCUMENTED VIOLATION

==================================================
Baseline mode: All violations documented for tracking
Run without --baseline to enforce constitution in CI
