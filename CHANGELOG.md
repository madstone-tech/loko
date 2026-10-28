# Changelog

All notable changes to the loko project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Ordered flows through the model, from item #5 of the first modelling session report
([ADR-0018](docs/adr/0018-dynamic-views.md)).

### Added

- **Dynamic views**: `step "<label>"` blocks inside a `view` (`from`, `to`, optional `via`,
  `description` and `reply`). Steps are numbered by position and checked against the declared
  relationships, through containment and in drawn direction; `reply = true` matches a relationship
  reversed.
- **`view.style`**: `edges` (numbered edges, the default; containers inside their system's
  boundary) or `sequence` (a sequence diagram). Both styles list the steps in a table on the view's
  page.
- **New codes**: `step_mismatch` (error), `step_unmatched` (warning, still drawn) and
  `dynamic_view_selection` (error: a view with steps also selects, or `style` without steps).
- **MCP and CLI**: a `step` edit target in `apply_edit`, with `before` to insert a step before
  another; steps in `describe`; `query` kind `flows` and `loko query flows <address>`.

## [1.2.0] - 2026-10-27

The remaining items of the first modelling session report: deployment nodes and views that say
what they show, docs links that survive moving files, and a hint when a view renders too wide
([ADR-0017](docs/adr/0017-language-polish.md)).

### Added

- **Deployment node attributes**: `title`, `technology`, `description` and `tags` on `node`. The
  deployment diagram labels a node with its title and `[technology]`; its page lists the
  description and tags in a "Nodes" table. Nested nodes inherit nothing.
- **`instance.technology`**: replaces the element's technology on that deployment's diagram, for
  example a local process in `dev` and Lambda in `prod`.
- **View `title` and `description`**: the title is the page heading, browser title and navigation
  entry, and is drawn at the top of the diagram; the description is shown under the heading. The
  name stays the view's address, file name and link.
- **Information diagnostics** (`severity` `info`): advice that never changes an exit code,
  `--strict` included, and is not counted as a warning. `--severity warning` (MCP
  `severity: "warning"`) hides it; `ignore_warnings` and `--ignore` accept information codes.
- **New codes**: `docs_root_relative` (information), `docs_ambiguous` (warning) and `wide_view`
  (information, from `build` when a view renders at least 3,000 px wide and 2.5 times wider than
  tall with no `direction` or `layout` chosen).
- All new attributes are in the `schema` MCP tool, `fmt` and `apply_edit` (node, instance and view
  targets).

### Changed

- **`docs` paths resolve from the declaring file**, with the project root as a fallback. Every
  1.1.0 path still resolves; one found only from the root prints a `docs_root_relative` line with
  the path to write. Exit codes and warning counts are unchanged, and `--severity warning` or
  `--ignore docs_root_relative` restores the 1.1.0 text. If a different file exists at the
  file-relative location, it is now linked and `docs_ambiguous` names both.
- `docs_not_found` names both locations checked.
- The diagnostics JSON contract's `severity` enum gains `info`, and its `summary` an optional `info`
  count.

## [1.1.0] - 2026-10-26

From a first modelling session through `loko mcp`
([report](specs/017-mcp-validation-ergonomics/first-run-report.md)): a new client learns the language
from the server, sees only the warnings that matter, and can render and start a project without a
shell.

### Added

- **`schema` MCP tool**: every block, where it nests, its attributes with types, required flags and
  allowed values, the `apply_edit` address forms and the `docs` path rule, from the same language
  table the compiler uses. The server's handshake instructions point clients at it.
- **`build` and `fmt` MCP tools.** `build` writes what `loko build` writes, into a directory inside
  the project, and can return one view's D2 or SVG inline. `fmt` previews or applies canonical
  formatting with `apply_edit`'s revision check.
- **Starting from an empty directory through MCP**: `apply_edit` has a `project` target, and
  `describe` in an empty directory returns the request that creates the project.
- **`instance.planned`**: an instance whose infrastructure is not built yet raises no
  `unbound_instance`; `planned_instance_bound` flags the marker once a binding exists.
- **Ignoring warnings**: `ignore_warnings` on the `project` block, and `--ignore` on `loko validate`
  and `loko build` (and `ignore` on the MCP `validate` and `build` tools). Ignored warnings do not
  count for `--strict`. Errors cannot be ignored; an unknown code in the source is the
  `unknown_warning_code` warning.
- **`loko validate --summary`** (one line per code) and **`--severity error`**, also on MCP
  `validate`.

### Changed

- **MCP diagnostics are compact by default**: `severity`, `code`, `detail`, `address` and `at`
  (`file:line`). Pass `ranges: "full"` to any tool for the previous form, with `summary`, `range`
  and `related`. On the reported model a `validate` response drops from 13,048 to 5,914 bytes. CLI
  output is unchanged. ([ADR-0016](docs/adr/0016-mcp-ergonomics.md))

## [1.0.2] - 2026-10-25

### Fixed

- `apply_edit` refused every file in a subdirectory with `path_refused: … resolves outside the
  project` when the project root was `.`, the default for `loko mcp`. Files at the root itself were
  accepted, and so were subdirectories when `--project` named another directory.
- `missing_docs` no longer fires for an `external`: it is someone else's system and has no prose of
  its own. A `docs` path an external does declare is still checked.
- `docs_not_found` says that `docs` paths are relative to the project root, not to the file that
  declares them.

## [1.0.1] - 2026-10-09

### Fixed

- The MCP server reported version `0.1.0` in its `initialize` handshake; it now reports the build's
  version.

### Documentation

From a first-run test in which an assistant used only the published documentation:

- A **diagnostics reference** in the language doc: every error and warning code, what it means and
  how to fix it, and how `--strict` treats warnings. `missing_docs`, the first warning most new
  models hit, was not documented anywhere.
- `loko query` explains that paths follow dependencies, so an asynchronous flow through a queue is
  traced in two parts, and that `no path` exits `0`.
- The CLI reference has a GitHub Actions example and lists `--project` for `loko mcp`.
- Container image examples say `:latest`; the CI guide pins `:v1.0` and explains why. CI examples
  use the current action versions.

## [1.0.0] - 2026-10-09

The first v1 release: loko becomes a compiler for software architecture. The HCL source is the
only authored artefact, and diagrams, markdown, the site, queries, exports and the MCP server are
projections of it. Nothing carries over from the 0.x format; see the note at the end.

### Added

- **The loko language** (feature 013): `*.loko.hcl` files with one `project` block; `person`,
  `system`, `container`, `component` and `external` elements; `uses` relationships nested in their
  source; a deployment plane (`deployment`, `node`, `instance`, Terraform and CloudFormation
  `binding`s); declared `view`s; `locals` and functions; and a `reconcile` block. References are
  typed and resolved at compile time across files. See [docs/language.md](docs/language.md).
- **`loko validate`** reports every problem in one run with file, line and column, and usually a
  suggestion. `--strict` fails on warnings, and `--format json` emits the diagnostics for CI. Exit
  codes are `0`, `1` and `2`, and nothing else.
- **`loko fmt`** rewrites source in canonical form; `--check` lists non-canonical files and exits
  `1` without writing.
- **`loko export --format json|toon`** writes the compiled architecture with a schema version.
  Output is byte-identical across runs, machines and file orderings.
- **Rendering attributes** (feature 016), all optional, affecting diagrams and pages only:
  - `title` on any element: a display name shown in place of the address name, which stays
    visible beside the kind.
  - `shape` on containers and externals: `database`, `queue`, `topic`, `function` or `bucket`,
    drawn as a cylinder, a queue, a hexagon, a chevron or stored data.
  - `kind` on relationships: `async` (long dashes) or `trigger` (arrow drawn from the trigger).
    Queries are unchanged.
  - `tags` on relationships, shown on edges and in the uses / used-by tables.
  - `direction` on views (`down` or `right`).
  - `layout` on the project and on views: `dagre` (default) or `elk`. ELK draws right-angled
    edges and packs dense views more tightly, at about five times the layout time. A view's
    `layout` overrides the project's; the D2 source names the engine, so the d2 CLI matches.
  - New diagnostics: `invalid_attribute_value`, `shape_not_allowed`, `empty_title`. See ADR-0015.
  - Each diagram on the site links to its full-size SVG.
- **`loko serve --host`** picks the interface to listen on, so the preview works from a container
  (`--host 0.0.0.0` with the port published on the host's loopback). The default stays
  `127.0.0.1`, and loko warns whenever the address is not loopback.
- **Tool edits realign the edited block**: after `apply_edit` changes a block, its `=` signs
  line up as `loko fmt` would. Comments are kept and other blocks are untouched.
- **MCP tools return** (feature 015), rebuilt on the compiled HCL. `loko mcp` registers five:
  - `describe` (summary, structure or full, optionally scoped to one element), `query` and
    `validate` read the architecture. TOON is the default output.
  - `apply_edit` adds, updates and removes elements, relationships, environments, node groups,
    instances, bindings and declared views, singly or in batches of up to 100, with preview and
    cascading removal.
  - `move` renames an element, changing its name, its kind or both.
  - Every change compiles before anything is written and saves all-or-nothing. A write based on a
    stale revision is refused. Comments, ordering and spacing outside the edited declaration
    survive byte for byte, and the tools write nothing but `*.loko.hcl` files.
  - See [docs/mcp-integration.md](docs/mcp-integration.md) and ADR-0014.
- **The `moved` block** records a rename so an element's history survives it. New diagnostics:
  `moved_from_declared`, `moved_to_unresolved`, `moved_duplicate_from`. The IR gains `moves`,
  omitted when empty, so existing exports are unchanged.
- **`loko query`** (`dependents`, `dependencies`, `path`, `orphans`, `coupling`) gives the same
  answers as the MCP `query` tool; `--format json` output is identical.
- **`loko build` returns** (feature 014), rendering entirely from the compiled HCL:
  - D2 and SVG diagrams for a landscape, every system with containers, every container with
    components, and every environment, with no configuration.
  - Declared `view` blocks, rendered alongside the automatic views. A view that selects nothing
    warns with `view_empty`; one that replaces an automatic view warns with `view_shadowed`.
  - Markdown documents and a browsable site with a page per element: prose, uses and used-by
    tables, children, and the relevant diagram.
  - Byte-identical output across runs, machines and discovery orders. Every file carries a
    generated-file notice naming its sources. Owned files are tracked in `dist/.loko-manifest`, so
    only they are pruned.
  - Theme overrides from `<project>/templates/`. A malformed override fails with `theme_invalid`.
- **`loko serve` returns**: an in-memory preview on `127.0.0.1` that rebuilds as you save and
  reloads the browser. When the source does not compile, pages show the diagnostics instead of
  stale output.

### Changed

- **Container, component and declared views are laid out top-down by default** (previously
  left-to-right). Wide container views are much narrower: the reference system's container view went
  from 7407 × 1693 (ratio 4.4) to 3184 × 2454 (ratio 1.3). Set `direction = "right"` on a view to
  keep the old layout. The landscape and deployment views are unchanged.
- Output file names and site URLs keep underscores: `orders_db` is now
  `element/container/orders_db.html`, not `orders_5fdb.html`. Other unsafe bytes are escaped as
  `~` plus two hex digits. A rebuild prunes the old files automatically.
- **Go 1.27.** Building loko now requires Go 1.27 (`go 1.27.0`, toolchain go1.27.2); CI, releases
  and the container image build with 1.27. Tests that exercise concurrency run on
  `testing/synctest`'s fake clock, and every package that starts goroutines fails on a leaked one
  (Go 1.27 `goroutineleak` profile).
- **No `d2` binary is needed.** Diagrams render in-process with the d2 library (v0.7.1, dagre
  layout by default, ELK on request). No layer may start a process (`os/exec` is banned by archcheck).
- **Breaking (container image):** the image's entrypoint is now `/usr/local/bin/loko`. Run
  `docker run <image> build`, not `docker run <image> loko build`. The image no longer contains a
  `d2` binary or scaffold templates.
- **Dependencies**: `golang.org/x/net` v0.60.0 fixes five HTTP/2 advisories (GO-2026-6603,
  6610, 6611, 6612 and 6617). loko never called the affected code, but release scans now report
  no vulnerabilities at all.
- **Releases**: the container image is one multi-platform manifest (`linux/amd64`, `linux/arm64`)
  with an SBOM, and prereleases no longer move `latest`. Homebrew installs a cask,
  `brew install --cask madstone-tech/tap/loko`, with bash, zsh and fish completions.
- **Documentation** describes v1 only. The v0 guides, examples and references (`loko.toml`,
  frontmatter, hand-written D2, the HTTP API) are removed, and the examples are HCL projects.
- Constitution v1.4.0: wiring may live in `main.go` or `cmd/`.

### Removed

- The 0.x project format (`loko.toml`, markdown frontmatter relationships, hand-written `.d2`
  files) and the commands built on it: `loko init`, `loko new`, `loko api` and
  `loko validate --check-drift`. Drift cannot occur with one source of truth.
- `loko watch`: it is part of `loko serve`.

## 0.x — proof of concept (v0.1.0 to v0.3.1)

The 0.x releases were a proof of concept built on `loko.toml`, markdown frontmatter and
hand-written D2, with scaffolding (`loko init`, `loko new`), an HTTP API (`loko api`) and drift
detection between the two relationship sources. v1 replaces that model with a compiled HCL source
of truth and shares no file format with it; [ADR-0012](docs/adr/0012-hcl-source-of-truth.md)
explains why. The 0.x tags stay installable, and their notes are in the
[GitHub releases](https://github.com/madstone-tech/loko/releases) and this file's git history.

[Unreleased]: https://github.com/madstone-tech/loko/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/madstone-tech/loko/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/madstone-tech/loko/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/madstone-tech/loko/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/madstone-tech/loko/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/madstone-tech/loko/compare/v0.3.1...v1.0.0
