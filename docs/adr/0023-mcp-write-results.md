# ADR-0023: MCP write results: diagnostic delta and meaning diff

**Status:** Accepted
**Date:** 2026-11-10
**Feature:** [023-v160-theming-mcp](../../specs/023-v160-theming-mcp/spec.md)
**Related:** [ADR-0011](0011-toon-mcp-default.md) (TOON for MCP),
[ADR-0014](0014-hcl-authoring.md) (assistants edit HCL),
[ADR-0016](0016-mcp-ergonomics.md) (compact MCP diagnostics)

## Context

A third-party feedback report (item 3) describes an assistant that edits a model through
`loko mcp` (1.5). Two things made each write cost more context than the change needed:

- Every `apply_edit` and `move` result listed every diagnostic in the model. An edit that set one
  description repeated the whole warning list. A larger model cost more on every call.
- The diff showed lines that changed only in alignment. Canonical formatting realigns the `=`
  signs of the attributes next to a new attribute. A one-attribute edit then looked like a change
  of three or more lines.

`fmt` returned no diagnostics on success, so an assistant could not see the state of the model
after formatting.

## Decision

1. **A diagnostic delta by default.** A successful `apply_edit`, `move` or `fmt` result (applied,
   preview or no-op) has three fields in place of `diagnostics`:
   `diagnostics_added`, `diagnostics_removed` and `diagnostics_unchanged`. The last one is a count
   by severity (`errors`, `warnings`, `info`). `apply_edit` and `move` already compiled the source
   before the change to check the edit. That result is now kept, so the delta costs no extra
   compile.

2. **The matching key.** Two diagnostics are the same when their `code`, `severity`, `address`,
   file and `detail` are the same. Positions are not part of the key, because an inserted line
   moves every diagnostic below it. Diagnostics with equal keys pair in source order. A removed
   diagnostic keeps its position before the change. We kept `detail` in the key. Without it, two
   diagnostics with no address in one file can collide. We did not match by position through a
   line map, because the core would then need the adapter's diff algorithm.

3. **A meaning diff by default.** `apply_edit` and `move` re-diff each run of removed and added
   lines on copies without whitespace outside string literals. A pair that matches there is a line
   that changed only in whitespace. The diff shows it as a context line with its new text, and
   `files[].whitespace` lists its line number in the new file. A hunk with only such lines is not
   shown. The adapter does the classification, and the `SourceEditor.Diff` port takes a mode
   (`DiffMeaning` or `DiffExact`). Files on disk are still written in canonical form. Only the
   reported diff changes.

4. **The way back.** `diagnostics: "all"` returns the 1.5 `diagnostics` list, and `diff: "exact"`
   returns the 1.5 diff. Both together give the 1.5 result. Constitution principle VI requires
   that a default which removes information keeps the old shape one argument away.

5. **Refusals do not change.** A refused write returns `refusal` and the full `diagnostics` list of
   the refused compile, whatever `diagnostics` says. An assistant must see every blocking error.

6. **`fmt` compiles once.** `fmt` now compiles the project after formatting. Formatting changes
   only whitespace, so its delta is empty and only `diagnostics_unchanged` is set. With
   `diagnostics: "all"`, `fmt` lists every diagnostic after formatting. `fmt` always returns the
   exact diff, because every change it makes is whitespace.

## Measurements

On `testdata/projects/first-run` (16 warnings and 23 information diagnostics), an `apply_edit`
preview that sets the `description` of one relationship gives these result sizes
(`TestWriteResultSize`):

| `diagnostics` | `diff` | TOON bytes | JSON bytes |
|---|---|---|---|
| `changed` (default) | `meaning` (default) | 399 | 399 |
| `all` | `meaning` | 7,912 | 9,553 |
| `changed` | `exact` | 455 | 467 |
| `all` | `exact` (1.5) | 7,968 | 9,621 |

The default result is 5% of the 1.5 result. The target was at most 50% (SC-004). The diff of
the default result has exactly one `+` line (SC-005).

## Known misclassifications

- A rename changes an address. The diagnostics of the renamed element show as removed and added,
  not as unchanged.
- A diagnostic whose `detail` changes, for example a different "Did you mean" suggestion, shows
  as removed and added. Such diagnostics are errors, and errors refuse the change, so the
  assistant sees the full list in the refusal.

## Consequences

**Positive:**
- A write result costs context in proportion to the change, not to the model.
- A one-attribute edit shows as a one-line change.
- `fmt` reports the state of the model after formatting.

**Negative:**
- A client that read `diagnostics` from a successful write must now pass `diagnostics: "all"`.
- A client that applied the diff as a patch must now pass `diff: "exact"`.
- `fmt` does one more compile.

**Mitigations:**
- The changelog records the MCP output change, with the arguments that restore the old shape.
- `TestWriteResultSize` and `TestOneAttributeEditDiff` check the size target and the one-line
  diff on every `go test`.
