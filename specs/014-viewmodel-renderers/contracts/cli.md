# Contract: `loko build` and `loko serve`

**Feature**: `014-viewmodel-renderers` | **Status**: normative

Both commands are thin shells over shared use cases (FR-039, Principle III: CLI handler functions
≤ 50 effective lines). They use the three exit codes that feature 013's
[CLI contract](../../013-hcl-compiler-core/contracts/cli.md#exit-codes-fr-038) defines and introduce
no fourth (FR-038).

After this feature the command surface is `validate`, `fmt`, `export`, `build`, `serve`, `mcp`,
`version`, `completion`. `watch` stays removed because it folds into `serve` (SC-010).
`cmd/root_test.go` pins this set.

---

## `loko build`

Compiles the project, projects every view, renders the requested formats, and commits them to the
output directory.

```
loko build [--format d2,svg,md,html] [--out DIR] [--strict]
```

| Flag | Default | Behaviour |
|---|---|---|
| `--format` | `d2,svg,md,html` | A comma-separated list. It may also be repeated. Order and duplicates are irrelevant. |
| `--out` | `<project>/dist` | The output directory. Relative paths resolve against the current working directory, as `loko export --out` does. Created if absent. |
| `--strict` | false | Warnings escalate the exit code to `2`. |

**Behaviour**

1. When compilation reports errors, the diagnostics go to stderr, nothing is written, and the exit
   code is `1`. A previously built output directory is byte-for-byte untouched (FR-024, US1/AC5).
2. An unknown format stops the command before compiling, with exit code `1`:
   `unsupported format "pdf": supported formats are d2, svg, md, html` (FR-015).
3. Required formats are added automatically (`html` and `md` require `svg`), and the summary line
   names them: `added svg (required by html)`.
4. Render-stage findings (`view_empty`, `view_shadowed`, `output_path_collision`, `theme_invalid`)
   are reported with the compile diagnostics in the same renderer and format. Either error aborts
   the commit.
5. On success stdout carries one summary line, `built dist: 42 written, 3 unchanged, 1 removed`, and warnings go to stderr.
6. When there is nothing to draw (no elements), the command writes no diagram, reports
   `nothing to draw: the architecture declares no elements`, and exits `0` (edge case).
7. If the output directory cannot be written, the error names the directory, nothing is partially
   written, and the exit code is `1`.

No external process is started (FR-014, US6/AC3).

---

## `loko serve`

Builds the site in memory, serves it on loopback, and rebuilds when the source changes.

```
loko serve [--port 8080]
```

| Flag | Default | Behaviour |
|---|---|---|
| `--port` | `8080` | The TCP port. The server binds `127.0.0.1` only. |

**Behaviour**

1. Formats are fixed at `html` plus what it requires (`svg`). Nothing is written to disk: serve
   never modifies the output directory.
2. It prints `serving http://127.0.0.1:8080 (ctrl-c to stop)`.
3. A change to the watched set (the `*.loko.hcl` files, the prose they reference, and `templates/`)
   triggers one rebuild per burst (FR-032, FR-033). Browsers reload through SSE (FR-029).
4. A failed compile prints the diagnostics to stderr, switches the server to its error state, and
   causes every HTML request to return a diagnostics page with file, line and column (FR-030). The
   next good build recovers without a restart (FR-031).
5. It exits `0` on SIGINT or SIGTERM. It exits `1` when the port cannot be bound or the initial
   project root is unreadable. A compile error on startup is **not** fatal: the server starts in the
   error state.

---

## Not provided (FR-040, FR-041)

- No `mermaid` or other experimental diagram format is offered.
- No MCP tool reaches any backend. The `mcp` layer is forbidden from importing
  `internal/adapters/{d2,html,markdown,outputdir,devserver}/**` by archcheck.
