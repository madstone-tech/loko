# ADR-0017: Information diagnostics, docs paths from the declaring file, and the wide-view hint

**Status:** Accepted
**Date:** 2026-10-27
**Feature:** [018-language-polish](../../specs/018-language-polish/spec.md)
**Related:** [ADR-0015](0015-rendering-attributes.md) (layout engines),
[ADR-0016](0016-mcp-ergonomics.md) (warning controls)

## Context

The first modelling session ([017 first-run report](../../specs/017-mcp-validation-ergonomics/first-run-report.md))
left three findings that need more than a new attribute:

- `docs` paths resolved from the project root. A source file in `architecture/payments/` had to
  write `architecture/payments/docs/overview.md`, and moving the file with its docs folder broke
  every link. The reporter found the rule by trial and error.
- A container view of about ten containers rendered 3875 × 1441 px, unreadable at page width, and
  nothing said that `layout = "elk"` (3046 × 1695 px) would fix it.
- Both findings are advice. A warning would turn `--strict` red on models that are not wrong.

## Decision

1. **An information severity.** `info` joins `error` and `warning`. Information is printed
   (`Info:`), included in JSON, TOON and MCP output and in `--summary` after warnings, and can be
   ignored like a warning (`ignore_warnings`, `--ignore`, MCP `ignore`). It never changes an exit
   code, `--strict` included, and is not counted as a warning. `--severity warning` (MCP
   `severity: "warning"`) hides it. The text summary adds `, N info` only when there is some, so
   output without information is unchanged. The diagnostics schema's severity enum gains `info`.

2. **Docs paths resolve from the declaring file, with the root as a fallback.** A path is looked
   for next to the file that declares the element first, then from the project root (the 1.1.0
   rule). Found only from the root, it is used with a `docs_root_relative` information line giving
   the path to write. Found at both as different files, the file-relative one is used and
   `docs_ambiguous` (a warning, since the linked document changes) names both. Found at neither,
   `docs_not_found` names both locations. Paths stay confined to the project root. Every 1.1.0
   path still resolves, and the export's `docs` is the resolved path relative to the root, so
   models whose paths resolve from the root export byte-identically.

   Resolution needs the declaring file and the file system, so it moves into the source adapter
   (`hclsource`), which records the outcome on the element. The use cases only report it, which
   removes their one `os` import.

3. **A wide-view hint.** A view whose rendered SVG is at least 3,000 px wide and 2.5 times wider
   than tall gets a `wide_view` information line from `build` (CLI and MCP), naming its size and
   ratio and what to try. Neither threshold alone separates the report's view from readable
   ones: four example views are wider in shape but narrow enough to read, and the report's view
   is the narrowest that is unreadable at page width. A `direction` or `layout` on the view, or a
   `layout` on the project, means the author chose, so there is no hint. The size comes from the
   SVG already laid out (its viewBox), so there is no second layout run, and `validate`, `serve`
   and a D2-only build give no hint. loko never changes a layout on its own.

## Consequences

**Positive:**
- A source file and its docs folder move together without breaking links, and 1.1.0 models keep
  working.
- Authors learn which setting fixes an unreadable diagram from the build itself.
- Advice has a place that does not fail CI.

**Negative:**
- Models with root-relative docs paths from subdirectories print new information lines, a change
  to existing text output.
- A 1.1.0 model with a different file at the file-relative location links that file instead, with
  a warning.
- The hint thresholds are fixed and measured on this repository's examples.

**Mitigations:**
- `--severity warning` or `--ignore docs_root_relative` restores the 1.1.0 text; exit codes and
  warning counts are unchanged.
- `docs_ambiguous` names both files, so the change of document is never silent.
- The hint is information only and can be ignored per project.
