# Quickstart: validating 1.6.0

**Branch**: `023-v160-theming-mcp` | **Date**: 2026-11-09

Prerequisites: a binary built from the branch (`go build -o /tmp/lk .`) and a
clean `go test ./...`. Each scenario names the spec story and success criterion
it proves. Contracts hold the exact rules.

## 1. Theming without copied blocks (story 1, SC-001)

```sh
/tmp/lk build -p examples/custom-branding
```

Expected:
- The build prints `theme: templates/ (custom.css; fonts.css; layout.gohtml: brand)`
  plus an `assets/` entry ([theming T-6](contracts/theming.md)).
- `templates/layout.gohtml` defines only `brand`. In the browser the sidebar
  shows the example's mark, the search box, tree and views list are built-in,
  interface text uses the example's sans font, code blocks stay monospace, and
  every page shows the project's `footer` text.
- `loko build -p examples/simple-project` prints no theme line.

## 2. Theme files and fonts (story 2, SC-002, SC-003)

```sh
ls examples/custom-branding/dist/assets/theme/
grep -c "base64" examples/custom-branding/templates/custom.css   # → 0
```

Expected: the mark and font files are in `dist/assets/theme/`. `custom.css` has
no embedded data. The page loads `assets/fonts.css` from the theme, and no
built-in font face (check the browser's network panel or `grep -c woff2
dist/assets/fonts.css`). Removing one asset and rebuilding prunes it.

Negative case: add `templates/assets/../x` through a symbolic link, or two files
that differ only in case. The build fails with `theme_invalid` that names the
file, and `dist/` is unchanged.

## 3. MCP write results (story 3, SC-004, SC-005)

Run `loko mcp -p testdata/projects/first-run` and call `apply_edit` with
`preview: true` to set one description.

Expected ([mcp-write-results](contracts/mcp-write-results.md)):
- The result has `diagnostics_unchanged` counts, no `diagnostics` list, and its
  size is at most half of the same call with `diagnostics: "all"` and
  `diff: "exact"`.
- `files[0].diff` has exactly one `+` line. The realigned sibling attributes are
  in `files[0].whitespace`.
- The same call with `diagnostics: "all"` and `diff: "exact"` matches 1.5.
- `move` and `fmt` return the same delta fields. A refused edit still lists every
  blocking error.

## 4. Default layout (story 4, SC-006, SC-007)

```sh
/tmp/lk build -p testdata/projects/first-run --out /tmp/fr
grep -l 'layout-engine: elk' /tmp/fr/diagrams/*.d2 | wc -l   # every view except sequences
```

Expected: every view without a layout uses elk. Setting `layout = "dagre"` on the
project gives the 1.5 diagrams. The wide-view hint uses the text in
[rendering R-2](contracts/rendering-defaults.md).

## 5. Redundant relationship (story 5)

Add a component to a container where both use the same target. Run
`/tmp/lk validate`. Expected: one `Info: redundant_relationship` on the container's
relationship, naming the component's. `--strict` keeps the exit code. Adding the
code to `ignore_warnings` removes it.

## 6. Listing (story 6)

```sh
/tmp/lk query views -p examples/simple-project
/tmp/lk query elements --kind container -p examples/simple-project
/tmp/lk query relationships --format json -p examples/simple-project
```

Expected: the tables and fields in [Q-1](contracts/query-and-diagnostics.md),
sorted. MCP `query` with `kind: "views"` returns the same rows.

## 7. Outside endpoints (story 7)

```sh
/tmp/lk build -p testdata/projects/two-systems --out /tmp/ts
grep -c '"outside__' /tmp/ts/diagrams/deployment-prod.d2   # 4 outside nodes
```

Expected: one outside node per outside element, labelled with its title. With
`outside = "collapsed"` on the project or the `deployment` block, the view
matches 1.5.

## Gates

```sh
go test ./...
golangci-lint run
task audit-constitution
go test ./cmd/ ./internal/adapters/{d2,hclsource,html,markdown}/ -update && git diff --stat testdata internal/adapters/*/testdata
```

Every golden change must trace to a task: `assets/style.css` (font token, T009),
the 6 `projection.json` (layout, T062, and outside endpoints for 3 of them,
T086), and the 4 html and 2 markdown deployment-view goldens (T086).
