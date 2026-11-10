# Contract: Listing queries and `redundant_relationship` (1.6.0)

**Feature**: `023-v160-theming-mcp` | **Status**: normative.

## Q-1 Listing queries

New kinds in the existing `query` family. The top-level command set is
unchanged (`TestCommandSurface`).

| CLI | MCP `query` | Rows, sorted | Text columns |
|---|---|---|---|
| `loko query views` | `kind: "views"` | one per resolved view (generated and declared), by name | `NAME KIND TITLE` |
| `loko query elements [--kind <kind>]` | `kind: "elements"`, `element_kind` | one per element, by address | `ADDRESS KIND TITLE` |
| `loko query relationships` | `kind: "relationships"` | one per relationship, by address | `ADDRESS SOURCE TARGET KIND DESCRIPTION` |

- `KIND` of a view is its view kind: `landscape`, `system`, `container`,
  `deployment` or `declared`. A shadowed generated view is replaced by the
  declared one, as in the build.
- A relationship with no `kind` shows `sync`.
- `--kind` and `element_kind` accept `person`, `system`, `container`,
  `component`, `external`. Any other value is refused with `invalid_query`,
  which lists the five kinds.
- `--format json|toon` on the CLI and `format` on MCP give the encoded result:
  `views[]` (`name`, `kind`, `title`), `elements[]` (`address`, `kind`, `name`,
  `title`, `distance`), `relationships[]` (`address`, `source`, `target`, `kind`,
  `description`). Existing query kinds keep their 1.5 output byte for byte.
- These kinds take no address argument.

## Q-2 `redundant_relationship`

| Property | Value |
|---|---|
| Code | `redundant_relationship` |
| Severity | information. It never changes an exit code, `--strict` included (ADR-0017). |
| Ignorable | yes: `ignore_warnings` and `--ignore` |
| Fires when | a container and one of its own components use exactly the same target element. A component that uses a child of the container's target does not fire it. |
| Address and range | the container's relationship |
| Related | each component relationship that repeats it, sorted by address |
| Summary | `Relationship repeats a component relationship` |
| Detail | `<container relationship> and <component relationship> both use <target>. The system view draws them as one edge and shows neither description. Remove the container relationship, or give it a different target.` When the kinds differ, the detail adds `The kinds differ: <kind> and <kind>.` |

- The rule lives in its own use-case file. `CompileArchitecture` calls it right
  after `ValidateWarnings`, so `loko validate`, `loko build` and every MCP tool
  report it, and the existing ignore step covers it.
- `docs/language.md` lists it in the Information table of the diagnostics
  reference.
