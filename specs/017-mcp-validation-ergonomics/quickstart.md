# Quickstart: validating feature 017

Each scenario proves one user story end to end. Wire shapes are in [contracts/](contracts/); types
in [data-model.md](data-model.md).

**Prerequisites**: `make build`, so `./loko` is this branch. MCP scenarios use any stdio MCP
client, or the end-to-end harness in `internal/mcp/build_e2e_test.go`.

## 1. Learn the language (Story 1)

1. Start `loko mcp` in `testdata/projects/serverless-reference`.
2. `initialize`: `instructions` mentions `schema`.
3. Call `schema`: every block in `docs/language.md` is listed; `layout` lists `dagre`, `elk`;
   `instance` lists `planned`.
4. Call `schema` with `block: "view"`: one block only. With `block: "sequence"`: refused,
   `unknown_block`, suggestions list the block names.
5. `go test ./internal/core/usecases -run TestLanguageParity` passes; deleting an attribute from
   `arch.Language` makes it fail.

**Expected**: SC-006 holds; the TOON answer is under the R2 token budget (asserted by test).

## 2. Warnings that point at real problems (Story 2)

Fixture: a copy of the reported 12-container model shape (`testdata/projects/first-run`), with
externals as `external` and unbuilt instances `planned = true`.

1. `loko validate`: at most 10 warnings (SC-002); no `unbound_instance` for planned instances.
2. Add a binding to one planned instance: `planned_instance_bound` appears.
3. `loko validate --summary`: one line per code, at most 30 lines (SC-003).
4. `loko validate --strict --ignore missing_docs` on a model whose only warnings are
   `missing_docs`: exits `0`. Without `--ignore`: exits `2`.
5. `loko validate --ignore syntax_error`: exits `1`, lists the warning codes.
6. Put `ignore_warnings = ["missing_docs", "no_such_code"]` on the project: `missing_docs` is gone,
   `unknown_warning_code` names `no_such_code`. `loko build --strict` agrees with
   `loko validate --strict`.
7. MCP `validate`: each diagnostic has `at` and no `range`; the response is at most half the size
   of the same call with `ranges: "full"` (SC-004).

## 3. Render and format without a shell (Story 3)

1. MCP `build` with `formats: ["d2", "svg"]`: `paths` match what `loko build --format d2,svg`
   writes, byte for byte.
2. MCP `build` with `inline: "<a view>"`: `inline.content` equals the file on disk.
3. MCP `build` with `out: "../elsewhere"`: `path_refused`, nothing written.
4. Introduce an unresolved reference; MCP `build`: `ok: false`, no files changed.
5. Misformat a file; MCP `fmt` with `preview: true`: a diff, file unchanged. Without `preview`: the
   file matches `loko fmt` output. Repeat with a stale `base_revision`: `stale_revision`.

## 4. Start from nothing (Story 4)

1. `loko mcp` in an empty directory; `describe`: `ok: false`, `bootstrap.syntax`, and
   `bootstrap.request`.
2. Send `bootstrap.request` with a name filled in and `preview: true`: a diff creating one file.
   Send it again without `preview`: the file exists.
3. `apply_edit` `add project` again: `project_exists`, naming the file.
4. In a second empty directory, `apply_edit` `add person.user` first: `no_project`, with the
   request to send.
5. Add a person and a system, then `build`: SC-005 holds, using MCP only.

## Regression

`make test`, `make lint`, `make audit-constitution`. The determinism job and the export goldens
pass unchanged (SC-007).
