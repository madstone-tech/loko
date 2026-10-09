# Parked code

Nothing here is compiled. Go ignores directories whose names begin with `_`,
which is what keeps this out of the build without a build tag on every file.

These files are **retained deliberately** (FR-045 of feature
013-hcl-compiler-core). They were written against the v0 model, which feature
013 deleted, so they no longer compile — but each is the starting point for a
later stage and rewriting from scratch would throw away working logic.

| Path | Why it is here | Comes back in |
|---|---|---|
| `d2_parser.go`, `d2_parser_test.go`, `d2_import_test.go` | Parses D2 arrow syntax into relationships. The v1.2 diagram-import adapter is this code pointed at the new model. | v1.2 diagram import |

To bring a file back: move it out of `_parked/`, change its imports from
`internal/core/entities` to `internal/core/entities/arch`, and rework it against
the compiled IR. Do not restore it unchanged — the model it was written for no
longer exists.

The site builder that used to live in `html/` returned in feature 014
(`internal/adapters/html`), reworked to render view models; its presentation
moved into embedded theme files rather than being restored.

The v0 MCP tools (`mcp_tools/`, `mcp_graph_cache.go`) were deleted in feature
015 rather than restored: they wrote the old file-tree model, and nothing in
them carried over. Their replacements — `describe`, `query`, `validate`,
`apply_edit` and `move` — read the compiled IR and write HCL only.
