# Changelog

All notable changes to the loko project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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

- `loko watch`: it is part of `loko serve`.

## 0.x — proof of concept (v0.1.0 to v0.3.1)

The 0.x releases were a proof of concept built on `loko.toml`, markdown frontmatter and
hand-written D2, with scaffolding (`loko init`, `loko new`), an HTTP API (`loko api`) and drift
detection between the two relationship sources. v1 replaces that model with a compiled HCL source
of truth and shares no file format with it; [ADR-0012](docs/adr/0012-hcl-source-of-truth.md)
explains why. The 0.x tags stay installable, and their notes are in the
[GitHub releases](https://github.com/madstone-tech/loko/releases) and this file's git history.

[Unreleased]: https://github.com/madstone-tech/loko/compare/v0.3.1...HEAD
