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
| `html/` | The site builder and its templates. The renderer stage rewires it to consume ViewModels instead of the filesystem, and splits `templates.go` (1,535 effective lines) while doing so. | Renderer stage (014) |

To bring a file back: move it out of `_parked/`, change its imports from
`internal/core/entities` to `internal/core/entities/arch`, and rework it against
the compiled IR. Do not restore it unchanged — the model it was written for no
longer exists.
