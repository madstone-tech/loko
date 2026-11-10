# Contract: MCP write results (1.6.0)

**Feature**: `023-v160-theming-mcp` | **Status**: normative. Recorded in ADR-0023.
Applies to the source-changing tools `apply_edit`, `move` and `fmt`.

## W-1 New inputs

| Input | Tools | Values | Default | Meaning |
|---|---|---|---|---|
| `diagnostics` | `apply_edit`, `move`, `fmt` | `changed`, `all` | `changed` | `all` returns the 1.5 shape: one `diagnostics` list with every diagnostic in the model after the change. |
| `diff` | `apply_edit`, `move` | `meaning`, `exact` | `meaning` | `exact` returns the 1.5 unified diff. |

`ranges` (`line`, `full`) still applies to every diagnostic list. `format` is
unchanged.

## W-2 Diagnostic delta (`diagnostics: "changed"`)

A successful result (applied, preview or no-op) carries:

| Field | Type | Content |
|---|---|---|
| `diagnostics_added` | list of diagnostic views | in the model after the change, not before |
| `diagnostics_removed` | list of diagnostic views | before the change, not after. Positions are the pre-change `file:line` |
| `diagnostics_unchanged` | `{errors, warnings, info}` | counts by severity of the rest. Each count omitted when 0. The field is omitted when every count is 0 |

- The `diagnostics` field is absent in this mode.
- Matching key: `code`, `severity`, `address`, `file`, `detail`. Positions are
  not part of the key. Two diagnostics with the same key pair in source order.
- A rename (`move`, or `apply_edit` with `op: rename`) changes an address, so the
  renamed element's diagnostics appear as removed plus added.
- `fmt` compiles once, after formatting. Formatting is lexical, so its delta is
  empty and only `diagnostics_unchanged` is set. With `diagnostics: "all"`, `fmt`
  now lists every diagnostic after formatting. In 1.5 it returned none on success.

## W-3 Refusals

A refused change keeps its 1.5 content whatever `diagnostics` says: `refusal`
and the full `diagnostics` list of the refused compile (FR-015).

## W-4 Meaning diff (`diff: "meaning"`)

- `files[].diff` is a unified diff (3 context lines, GNU hunk headers) of the
  lines that changed in meaning. A line that changed only in whitespace outside
  string literals is shown as a context line with its new text.
- `files[].whitespace` lists the new-file line numbers of those lines. It is
  omitted when empty.
- A hunk that only changes whitespace is not shown. A file is listed when its
  diff is not empty or it has whitespace lines.
- Applied files are written in canonical form, as in 1.5.
- `fmt` always returns the exact diff: every change it makes is whitespace.

## W-5 Size target

On `testdata/projects/first-run` (at least 10 existing diagnostics), an
`apply_edit` preview that sets one description, encoded with the defaults, is at
most half the size of the same call with `diagnostics: "all"` and
`diff: "exact"` (SC-004). A one-attribute edit preview has exactly one `+` line
in its diff (SC-005).
