# Contract: command-line surface

**Feature**: `013-hcl-compiler-core` | **Status**: normative

Every command is a thin shell over a shared use case (FR-039, Constitution Principle III: CLI handler
functions ≤ 50 effective lines). Nothing below contains logic that the MCP layer cannot reach by
calling the same function.

---

## Exit codes (FR-038)

| Code | Meaning |
|---|---|
| `0` | Success — no errors, or warnings present without `--strict` |
| `1` | One or more errors |
| `2` | Warnings present and `--strict` was requested |

**Exactly three.** No command in this release introduces a fourth; `fmt --check` reuses `1`
(clarification 2026-09-22). Codes `3` and above are reserved for later stages (`3` = reconcile below
minimum coverage).

---

## `loko validate`

Compiles the project and reports every diagnostic found in one run (FR-031, FR-034).

```
loko validate [--strict] [--format text|json] [--path DIR]
```

| Flag | Default | Behaviour |
|---|---|---|
| `--strict` | false | Warnings escalate the exit code to `2` |
| `--format` | `text` | `text` renders with source snippets; `json` emits the diagnostics contract |
| `--path` | `.` | Project root to discover beneath |

Exit codes are identical whichever `--format` is chosen (FR-034a).

**Text output**

```
Error: Unresolvable reference
  on arch.loko.hcl line 14, column 16:
  14:     target = container.ordrs_db
                   ^^^^^^^^^^^^^^^^^^
  No element "container.ordrs_db" is declared. Did you mean "container.orders_db"?

2 errors, 1 warning
```

**JSON output**: see [diagnostics.schema.json](./diagnostics.schema.json). Ordering is deterministic
(FR-033) and identical to the text ordering.

---

## `loko fmt`

Rewrites source canonically, preserving every comment and the declaration order (FR-035).

```
loko fmt [--check] [--path DIR]
```

| Flag | Behaviour |
|---|---|
| *(none)* | Rewrites non-canonical files in place; prints each path rewritten |
| `--check` | Writes nothing. Lists non-canonical paths. Exit `1` if any, `0` if none (FR-035a) |

Guarantees:

- An already-canonical file is left byte-unchanged (SC-007).
- Running twice is idempotent — the second run reports nothing.
- A file that does not parse is reported with a source location and **left untouched**; the command
  does not partially format a project it cannot fully parse.

---

## `loko export`

Writes the compiled IR (FR-036).

```
loko export [--format json|toon] [--out FILE] [--path DIR]
```

| Flag | Default | Behaviour |
|---|---|---|
| `--format` | `json` | Both encodings carry equivalent information |
| `--out` | stdout | Destination file |
| `--path` | `.` | Project root |

- **No artefact is produced when compilation reports errors** (FR-037). Diagnostics go to stderr and
  the exit code is `1`; any `--out` file is left untouched rather than truncated.
- Output is byte-identical across runs, machines, and file-system orderings (FR-040, SC-003).
- Every artefact carries `schemaVersion` (FR-036a) and contains no timestamp, hostname, tool version,
  or absolute path (FR-036c).

---

## Commands removed in this stage

Consequence of the model replacement (research R9). Restored in the stages noted.

| Command | Status |
|---|---|
| `loko build` | Removed — returns with the renderer stage |
| `loko serve` | Removed — returns with the renderer stage |
| `loko watch` | Removed — folded into `serve` in the renderer stage |
| `loko init` | Removed — returns with the renderer stage (research R10) |
| `loko new system\|container\|component` | Removed permanently; authoring is editing source, or the MCP write tools from the authoring stage |
| `loko api` | Removed permanently (FR-041) |
| `loko validate --check-drift` | Removed permanently (FR-043) |

Surviving after this stage: `validate`, `fmt`, `export`, `version`, `completion`, `mcp` (reduced tool
set until the authoring stage).

---

## Global behaviour

- **stdout carries the artefact; stderr carries diagnostics and progress.** `loko export --format
  json | conftest test -` must work without filtering, which is the zero-cost policy escape hatch the
  design relies on.
- Colour is disabled when stdout is not a terminal or when `NO_COLOR` is set. Colour never appears in
  `--format json`.
- Every command accepts `--path`; none reads configuration from outside the project root, since the
  `project` block replaced the previous configuration file (FR-003).
