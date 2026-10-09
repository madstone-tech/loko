# ADR-0013: Every output is a projection through one view model

**Status:** Accepted
**Date:** 2026-09-30
**Feature:** [014-viewmodel-renderers](../../specs/014-viewmodel-renderers/spec.md)
**Related:** [ADR-0012](0012-hcl-source-of-truth.md) (HCL as the single authored source)

## Context

ADR-0012 made HCL the only authored artifact and deleted `build` and `serve`, because they drew from
the removed file-tree model. Bringing them back raised several questions:

- how outputs relate to the compiled IR;
- how to render diagrams without the external `d2` binary that v0 required;
- how to keep generated files trustworthy enough to commit;
- how to preview changes, now that fsnotify and the HTTP API are on the constitution's
  "removed in v1.0" list.

## Decision

1. **One intermediate value.** The IR resolves to views: a landscape, one per system with
   containers, one per container with components, one per environment, and every declared `view`.
   Each view projects to a `ViewModel` that has already decided which elements appear, how they
   nest, which connections join them, and how each is styled. The projection also builds one
   `ElementPage` per element. Every backend consumes only the resulting `Projection`, through a
   single `Backend` port, and knows nothing about HCL, C4 or other backends. Styling is a
   convention table in `internal/core/entities/viewmodel`, so every format agrees.
2. **One lifting rule.** Across every view kind, a relationship endpoint lifts to its nearest
   visible ancestor. When one end has no visible ancestor, the connection becomes a *boundary
   connection* to a single synthetic `outside` node rather than being dropped. In declared views,
   an excluded subtree resolves to nothing, so connections to it reach the boundary instead of
   lifting onto its visible parent.
3. **Format dependencies are declared, not imported.** `html` and `md` embed diagrams by path and
   declare that they require `svg`. The build adds a required format and says so. No backend
   imports another, so a backend is removed by deleting its package and its registry line.
4. **d2 runs in-process with the dagre layout.** `oss.terrastruct.com/d2` v0.7.1 is a library
   dependency confined to `internal/adapters/d2`. No process is started. `os/exec` is banned in
   every layer by archcheck and depguard. ELK was the first choice, but d2 v0.7.1 recompiles its
   layout engine on every call, and ELK made a 1,000-element build take 13.7 s against a 10 s
   budget. Dagre takes 2.8 s and draws the nesting just as correctly.
5. **The output directory is owned through a manifest.** `.loko-manifest` lists what the last build
   wrote. Only listed files are ever pruned, files whose bytes are unchanged are not rewritten, and
   staging proves the directory writable before anything is touched. A build with errors never
   reaches the store. Which files to write, skip or prune is decided by a pure function in core
   (`viewmodel.PlanCommit`); the adapter only performs the I/O.
6. **Every file opens with a notice** naming the source files it came from. It carries no
   timestamp, version or host, so output is byte-identical across runs, machines and discovery
   orders. CI checks this on Linux and macOS.
7. **Polling, not fsnotify.** `loko serve` stats the watched set every 200 ms: architecture source,
   the prose it references, and `templates/`. It signals once the set has been stable for one tick.
   This uses only the standard library and behaves the same on network mounts and container bind
   mounts, where inotify events are lost.
8. **A loopback preview server is not the HTTP API.** `internal/adapters/devserver` serves the last
   build from memory on `127.0.0.1`. It pushes reloads over Server-Sent Events and injects the
   reload script into responses only, so built files never contain dev code. When a build fails,
   every page shows the diagnostics instead of stale output. It is an adapter behind the
   `PreviewServer` port, used only by `serve`; it exposes no architecture data as an API.
   *Amended (housekeeping, 2026-10-09):* `--host` can widen the bind for containers. Loopback
   stays the default, and loko warns whenever the address is not loopback.
9. **Themes override by file or by block.** `.css` and `.js` files in `templates/` replace the
   built-in file. `.gohtml` files are parsed after the built-ins, so each `{{define}}` replaces one
   block. Unknown file names are rejected by a pure rule in core (`viewmodel.ValidateTheme`). Parse
   errors and unknown block names are reported by the html adapter, because only the template
   parser can see them. Any malformed override fails the build with `theme_invalid`.

## Consequences

**Positive:**
- A new output format is one backend package plus one registry line; the middle stage does not
  change.
- Generated output can be committed: rebuilding changes nothing, and a one-element edit changes
  only the files that depict it.
- loko ships as a single static binary. The container image carries no other executable.

**Negative:**
- d2 and goldmark add weight to the binary.
- Dagre's curved edges are less tidy than ELK's orthogonal routing.
- Polling costs a few hundred `stat` calls every 200 ms while `serve` runs.

**Mitigations:**
- The layout engine is one line in `svg_backend.go`. Revisit ELK if d2 starts reusing its JS
  runtime between layouts.
- The SVG backend caches renders by D2 source hash for the life of a `serve` process, so an edit
  re-lays-out only the views it touches.
