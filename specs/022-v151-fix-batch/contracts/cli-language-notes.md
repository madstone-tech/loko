# Contract: CLI help documentation pointer and language notes

**Feature**: `022-v151-fix-batch` | **Status**: normative

## Root help (user-facing text)

`loko --help` must let a new reader find the documentation in one step.

**Rule CLI-1**: The root help body (`rootCmd.Long`) carries one sentence pointing
at the repository `docs/` directory and naming `docs/language.md` as the full
`*.loko.hcl` language reference.

**Rule CLI-2**: No subcommand's help repeats the pointer. The root help is the
single home (`loko build --help` etc. stay unchanged).

**Rule CLI-3**: The help text follows the ASD-STE100 style used elsewhere in the
CLI: short sentences, active voice, no contractions.

## Source-language notes (error text and schema)

**Rule LANG-1**: The unknown-block error for `dynamic` keeps its existing
diagnostic code, position and first sentence. A second sentence in the same
detail string names the construct the author most likely wants. HCL `dynamic`
generates nested blocks, but an author who types it in loko usually means a C4
dynamic view (`docs/language.md`, "Dynamic views"):

```
"dynamic" is not supported: dynamic block generation is deliberately excluded
from v1.0. For a C4 dynamic view, use a view block with step blocks.
```

**Rule LANG-2**: The relationship `kind` note in the language schema (MCP `schema`
tool) agrees with `docs/language.md` ("Relationships", the `trigger` bullet). It
states:

- the default kind (`sync`),
- the default arrow direction: from the declaring element to its target,
- that `trigger` reverses the arrow, and where to write a trigger.

Note text (single string in `internal/core/entities/arch/schemas.go`), with no
semicolons (STE style):

```
sync by default. The arrow points from the declaring element to its target.
trigger reverses the arrow: write it on the invoked element, targeting its trigger
```

**Rule LANG-3**: Neither rule changes rendering behavior. Edge direction, severity,
exit codes and error structure are unchanged from 1.5.0 (patch discipline, spec
FR-008).
