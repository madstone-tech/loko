# MCP Tool Handler Migration Map (Feature 010)

This file records how each MCP tool handler under `internal/mcp/tools/` was brought within
the 30-effective-line per-handler budget (Constitution Principle III, FR-002, FR-011). Every
handler is now a thin protocol adapter: **unmarshal request → call use case → marshal
response**. Domain logic lives in `internal/core/usecases/`.

This file is categorically exempt from the size budget (documentation, not a handler).

## Decision per tool

`reuse` = handler delegates to a use case shared with the CLI / other tools (often extracted
in US1). `new` = a use case created for this tool in US2. `thin` = handler was already (or
became) a direct repository/port passthrough with no extractable domain logic.

| MCP tool | Use case(s) called | Disposition |
|----------|--------------------|-------------|
| `analyze_coupling` | `BuildArchitectureGraph` (+ rel-repo variant) | new |
| `build_docs` | `BuildDocs` | reuse (US1 build use case) |
| `create_component` | `ScaffoldEntity` | reuse (US1 scaffold) |
| `create_components` | `ProjectRepository` (batch over `ScaffoldEntity`) | reuse |
| `create_container` | `ScaffoldEntity` + `DiagramGenerator` | reuse |
| `create_system` | `ScaffoldEntity` (+ diagram generator) | reuse |
| `create_relationship` | `CreateRelationship` | new |
| `delete_relationship` | `DeleteRelationship` | new |
| `find_relationships` | `FindRelationships` | new |
| `list_relationships` | `ListRelationships` | new |
| `query_architecture` | `QueryArchitecture` | new |
| `query_dependencies` | rel-repo / project-repo query | thin |
| `query_project` | `ProjectRepository` | thin |
| `query_related_components` | rel-repo / project-repo query | thin |
| `search_elements` | `SearchElements` | new |
| `update_component` | `ProjectRepository` | thin |
| `update_container` | `ProjectRepository` | thin |
| `update_diagram` | `UpdateDiagram` | new |
| `update_system` | `ProjectRepository` | thin |
| `validate` | `ValidateArchitecture` + `BuildArchitectureGraph` | new |
| `validate_diagram` | `DiagramRenderer` (port) | thin |

## Schema types

Per-tool JSON-schema-shaped request/response structs were moved into sibling
`*_schemas.go` / the shared `schemas.go` file (categorically exempt from the function-size
budget per the exemptions in `rules.yaml`). Handlers reference these types; they carry no
logic.

## Verification

`task audit-constitution` reports **0 violations** across `internal/mcp/tools/` (and the whole
repo). Every handler function is ≤ 30 effective lines. Behaviour is preserved — MCP tool
input/output schemas are unchanged.
