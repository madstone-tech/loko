# Research: MCP & Validation Ergonomics

Decisions for [spec.md](spec.md). Each one names what was chosen, why, and what was rejected.

## R1. One language definition, in `arch`

**Decision**: Add `internal/core/entities/arch/schemas.go`, a pure-data table of the language:
every block (name, labels, where it may nest, attributes, one minimal example) and every attribute
(name, value type, required, allowed values, a one-line note). The allowed values reference the
existing sets (`arch.Shapes`, `arch.Layouts`, `arch.Directions`, `arch.RelationshipKinds`).

- `hclsource/schema.go` stops holding its own attribute name lists (`commonElementAttrs`,
  `projectAttrs`, …) and derives them from the table. The decoder's behaviour is unchanged.
- `authoring.legalAttrs` and `authoring.enumAttrs` stay as copies, because entity packages may not
  import one another. A parity test in `usecases` (next to `TestRenderValueSetParity`) compares
  them with the table per edit target, attribute name, value shape and allowed values.
- The `schema` MCP tool encodes the table through a use case, so the compiler, the refusal messages
  and the tool all come from the same definition (FR-004).

**Rationale**: today the language is described in three places: the decoder's name lists, the edit
schema (split between `legalAttrs` and `enumAttrs`), and `docs/language.md`. A fourth copy for the
tool would be one more place to forget when 018 adds attributes. One table removes the risk.

**Alternatives considered**:
- *Generate the tool's answer from the hcl `BodySchema`*: it has names only, no types, required
  flags or values, and it lives in an adapter the MCP layer must not depend on for content.
- *Make `authoring` import `arch`*: breaks the entity sibling-import rule, which needs an ADR and
  a rule change for a problem the existing parity-test pattern already solves.
- *Hand-write the answer as a markdown resource*: it would drift exactly like `docs/language.md`.

The file is named `schemas.go`, one of the pure-data names exempt from the file-size budget
(Constitution, Principle III), because the table is data with no logic.

## R2. The `schema` tool

**Decision**: a read-only MCP tool named `schema`, with optional `block` (one block name) and the
usual `format`. It returns the blocks, the attribute table, the `apply_edit` address forms, the
`docs` rule, and one example per block. An unknown `block` is refused with the list of names.

The address forms are data in `authoring` (one pattern and one example per `TargetKind`), next to
the code that parses them, so a parity test can parse every example.

**Rationale**: `schema` is the name the report suggested and what clients look for. One tool with a
filter keeps the tool list short. The full answer is measured in a test and must stay within
**3,000 tokens in TOON** (Principle VI); a single block must stay within 500.

**Alternatives considered**: an MCP *resource* instead of a tool: not every client lists resources,
and the report's client did not; `guide` or `language` as names: less discoverable.

## R3. Handshake instructions

**Decision**: the `initialize` result gains an `instructions` string (MCP 2025-06-18
`InitializeResult.instructions`): call `schema` before the first edit, use `external` for systems
the team does not own, and send `preview: true` to see an edit's diff before writing it. Kept under
80 words.

**Rationale**: instructions are the one place every client reads before choosing a tool.

## R4. `planned` on instances

**Decision**: `instance` accepts `planned = true` (bool). A planned instance without a binding draws
no `unbound_instance`; a planned instance with a binding draws a new warning,
`planned_instance_bound`. The IR carries `planned` (exported with `omitempty`, so existing exports
are byte-identical). Rendering ignores it in this feature.

`authoring` allows `ValueBool` as a top-level value for this one attribute; until now booleans
appeared only inside a map.

**Alternatives considered**: `lifecycle = "planned"`: opens a vocabulary (deprecated, retired, …)
that nothing yet needs (Principle VII). A tag such as `tags = ["planned"]`: tags carry no meaning
anywhere else in the compiler, and a magic tag would be invisible in `schema`.

## R5. Ignoring warning codes

**Decision**: suppression happens once, at the end of `CompileArchitecture`, from two sources: the
new `CompileRequest.Ignore` (per run) and the `project` block's new `ignore_warnings` list. Every
consumer (validate, build, serve, MCP validate, MCP build) therefore sees the same diagnostics, and
`ExitCode(strict)` counts only what is left (FR-011, FR-012).

- `arch` gains `WarningCodes` (the warning half of `AllCodes`) and `CheckIgnore(codes)`, which
  refuses an error code or an unknown code and lists the warning codes.
- In the source, an unknown code in `ignore_warnings` is a warning (`unknown_warning_code`), not an
  error, so an upgrade that retires a code cannot break a project. An error code there is an
  `invalid_attribute_value` error.
- Suppressed diagnostics are counted (`ValidateResult.Ignored`), so a summary can say "4 ignored"
  instead of hiding them silently.

**Alternatives considered**: filtering in each handler: four places to keep in step, and strict mode
would disagree between `validate` and `build`. A severity *override* (warning → info): loko has no
info severity, and adding one is a bigger change than the report asked for.

## R6. Severity filter and summary

**Decision**: both are presentation and happen in the validate use case's result, after
suppression. `--severity error` (CLI) or `severity: "error"` (MCP) keeps only errors and leaves the
exit code alone (FR-010). `--summary` / `summary: true` returns groups of
`{code, severity, count, addresses}`, errors first, then by code; addresses sorted and capped at 20
per group with a `more` count.

**Rationale**: the cap keeps SC-003 (one screen) true for large models.

## R7. Compact MCP diagnostics

**Decision**: every MCP tool that returns diagnostics (`validate`, `describe`, `apply_edit`, `move`,
`build`, `fmt`) returns them in a compact form by default:
`{severity, code, summary, detail, address, at}`, with `at` written as `file:line`. `ranges: "full"`
returns the 1.0.2 shape. The CLI's `--format json` keeps full ranges.

The mapping is a use-case function (`BriefDiagnostics`) returning a use-case DTO, so MCP handlers
stay thin and do not import entities.

**Measured**: for the reported 33-warning model, the 1.0.2 TOON response carried seven position
fields per diagnostic. Compact output drops six of them and the `range` key, a reduction the
SC-004 test measures against a fixture with the same shape.

**Recorded in**: ADR-0016 and the changelog, as the one deliberate output change (SC-007).

## R8. MCP `build`

**Decision**: a `build` tool wired in `cmd/wiring.go` with `newBuildDeps()`, so `internal/mcp` never
imports a render adapter (Constitution FR-041). Inputs: `formats` (default: the CLI default set),
`out` (default `dist`), `inline` (a view name), `inline_format` (`d2` default, or `svg`), `strict`,
`ignore`. Output: written paths, counts, compact diagnostics and, when asked, the inline view's
bytes as text.

- `out` must resolve inside the project root. Resolving it touches the file system (symlinks, real
  paths), which Principle II puts behind a port, so the check is a new `ArtifactStore` method,
  `Confine(ctx, root, out) (string, error)`, implemented in `adapters/outputdir` with the same
  absolute-root rule 1.0.2 gave `hclsource.checkPath`. It runs before rendering starts. The CLI
  keeps accepting any `--out`.
- The tool calls one use case, `BuildInProject` (new file `build_in_project.go`): confine, build,
  then pick the inline view. The inline view is picked from the artefacts the build already
  produced, by view name: no second render. An unknown name is refused with the list of view names.
  Keeping all three steps in the use case keeps the handler to parse, call, encode (Principle III).
- `strict` behaves exactly like `loko build --strict`: files are written, and remaining warnings
  make the result `ok: false` (the CLI exits `2` after writing, `cmd/build.go`).
- Builds take a read snapshot and do not queue on the authoring mutex; they write only the output
  directory, which `ArtifactStore` already owns through `.loko-manifest`.

**Alternatives considered**: returning SVG as an MCP image content block: most text clients cannot
use it, and SVG is text anyway.

## R9. MCP `fmt`

**Decision**: a `fmt` tool on the `AuthoringService`, so it queues with edits and takes the same
inputs as `apply_edit`: a required `base_revision`, and `preview` (default `false`, as in
`apply_edit`). `preview: true` returns the diffs and writes nothing; otherwise the files are written
through `SourceEditor.Commit` (atomic, refused as `stale_revision` if anything changed since
`base_revision`).

`SourceFormatter` gains one method, `Canonical(ctx, root) ([]authoring.FileContent, arch.Diagnostics,
error)`, returning the canonical bytes of each file that would change, so formatting becomes a
plan the editor commits. `Format` and `Check` stay for the CLI.

**Alternatives considered**: calling `Format` directly: it writes without a revision check and could
overwrite an edit another client just made (FR-019).

## R10. The `project` target and empty-project bootstrap

**Decision**:
- `authoring` gains `TargetProject`, address `project`. `add` creates the declaration and requires
  `name` and `file`; `update` sets `description`, `layout`, `loko_version`, `ignore_warnings`.
  `add` when a project exists is refused (`project_exists`), naming its file.
- Adding any element when the project has no `project` block is refused (`no_project`), with the
  minimal syntax and a ready-to-send `apply_edit` request in the refusal.
- `describe` on an empty directory returns `ok: false`, the existing `no_source_found` diagnostic,
  and a new `bootstrap` field: the minimal syntax and a ready-to-send request
  (`{"op":"add","target":"project","address":"project","file":"architecture.loko.hcl",
  "set":{"name":"<name>","description":"<description>"}}`).
- Creating the project's file in an empty directory goes through the same `checkPath`, plan and
  `Commit` as any new file.

**Alternatives considered**: an `init` tool (rejected in clarification Q3).

## R11. New diagnostic codes

| Code | Severity | When |
|---|---|---|
| `planned_instance_bound` | warning | a `planned` instance has a binding |
| `unknown_warning_code` | warning | `ignore_warnings` names a code loko does not emit |

Both get fixtures, rows in the diagnostics reference, and entries in `AllCodes`, so the existing
rule-coverage test enforces them.

## R12. Language-change governance

`planned` and `ignore_warnings` are permanent v1.x attributes. Following feature 016's precedent,
ADR-0016 records them with the compact-diagnostics default, and `docs/language.md` documents them.
