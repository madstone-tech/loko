# Research: ViewModel Renderers

**Feature**: `014-viewmodel-renderers` | **Date**: 2026-09-30 | **Plan**: [plan.md](plan.md)

Each entry records the decision, why it was made, and what was rejected. Every
`NEEDS CLARIFICATION` in the plan's Technical Context is resolved here.

---

## R1. In-process diagram rendering

**Decision**: Reintroduce `oss.terrastruct.com/d2` as a direct dependency, pinned at `v0.7.1` (the
version the v0 tree used before feature 013 dropped it). Compile with `d2lib.Compile`, lay out with
the in-process dagre plugin (`d2layouts/d2dagrelayout`), measure text with `textmeasure.NewRuler()`, and
render with `d2renderers/d2svg`. The existing `internal/adapters/d2/renderer.go`, which shells out
through `exec.LookPath("d2")`, is deleted rather than adapted.

**Rationale**: ELK and dagre both run inside an embedded JavaScript VM (goja) that ships in the d2
module, so no process is started and no executable is looked up (FR-014, US6). The fonts are
embedded in the module, so rendering works with an empty `PATH` and no network access.

**Revised during implementation (2026-09-30)**: the plan originally chose ELK, because the v0
renderer used `--layout elk` and ELK has a reputation for nested containers. It was replaced by
dagre after SC-006 failed under ELK (see R2 for the measurement). Rendered side by side, the two
engines drew the five-deep deployment nesting and the system boundary equally correctly: d2 v0.7.1's
`LayoutNested` lays out each container separately, which removes dagre's historical weakness with
edges into containers. ELK's orthogonal edge routing is somewhat tidier; dagre's edges curve. That
is a cosmetic difference, while the 10 s bound is a requirement.

**Determinism**: d2's compile → layout → render pipeline is deterministic for identical input. SVG
element ids come from a content hash of the diagram. The layout engine does not seed a random number
generator. We treat that as a claim to verify, not an assumption, so a golden test renders each
fixture twice in one process and once with the fixture's files discovered in reverse order, and
asserts the SVG bytes are identical (FR-021, SC-002). Pinning the exact d2 version is part of the
guarantee: a d2 upgrade may change the bytes, and any upgrade must regenerate the SVG goldens in the
same PR.

**Alternatives considered**:
- *Keep the d2 CLI, bundle it.* Still an external process, which FR-014 forbids outright.
- *ELK layout.* Tidier orthogonal edges, but about 5× slower in d2 v0.7.1 (R2), which breaks
  SC-006. Revisit if d2 starts reusing its JS runtime between layouts.
- *TALA layout.* A separate proprietary binary. Rejected for the same reason as the CLI.
- *Write our own SVG layout.* Out of proportion to the feature (Principle VII).

---

## R2. Rendering throughput (SC-006)

**Decision**: Render views concurrently on a bounded worker pool (`runtime.GOMAXPROCS(0)` workers).
Each worker gets its own layout runtime, and results are collected into a slice indexed by view
position, so output order never depends on which worker finishes first. The SVG backend keeps an
in-memory cache keyed by the SHA-256 of the D2 source it lays out. The cache lasts only for the
life of the process and never reaches disk.

**Rationale**: Layout dominates build time, and views are independent, so the work parallelises
cleanly. The per-process cache is what meets the 2-second watch bound: editing one element's
description changes the D2 source of the few views that contain it, and every other view is served
from the cache. A one-shot `loko build` gains nothing from the cache, and pays nothing for it.

**Verification**: `cmd/build_perf_test.go` generates a 1,020-element architecture (20 systems, 200
containers, 800 components: 221 views and 2,930 artifacts). It asserts that the full build is under
10 s and that a single-description rebuild through the serve path is under 2 s. The 200-container
wide view is its own test. It lives in `cmd/` rather than extending feature 013's `usecases`
harness, because it needs the real backends.

**Finding (2026-09-30)**: with ELK the full build took **13.7 s**. A CPU profile showed half of all
CPU time inside goja running the layout engine. d2 v0.7.1 creates a new JS runtime and recompiles
the engine for every layout call, once per nesting level. Worker count made it worse, not better
(11 workers 13.7 s, 6 workers 14.4 s, 4 workers 16.8 s), and `GOGC=400` gave only 12.1 s. Switching
the engine to dagre gave **2.8 s**. With dagre, a warm-cache rebuild after one edit takes 0.82 s, and
the 200-container view takes 1.5 s. See R1 for the layout-quality comparison.

**Finding (2026-09-30, first CI run)**: GitHub's hosted runners are about 2.9× slower per core than
a 4-CPU developer container. With the code as committed, the full build took 13.4 s on
`ubuntu-latest` and 8.8 s on `macos-latest`, and the warm rebuild took 2.4 s and 1.8 s. A profile
showed that 59% of a warm rebuild went to re-rendering the site navigation on every page, which is
quadratic in views. The navigation is now rendered once per page depth, and pages render in
parallel. That brought the warm rebuild from 1.12 s to 0.16 s on 4-CPU Linux. The full build is now
67% dagre inside d2's JS engine, which d2 v0.7.1 recreates for every layout, so our code is no
longer the bottleneck. It takes 4.2 s on 4-CPU Linux, and about 12 s estimated on hosted runners.

**Decision**: the 10 s full-build bound is enforced on developer hardware (`task test`). On CI
(`CI=true`) the time is logged, not failed. The edit-to-browser bound, which is the latency users
feel, stays strict on CI. Revisit if d2 starts reusing its JS runtime between layouts.

**Alternatives considered**: An on-disk render cache. Rejected because it is state that outlives
the process. The compiler is stateless by design (ADR-0012), and a stale cache is exactly the kind
of second source of truth this feature exists to remove.

---

## R3. Where views come from

**Decision**: The system derives views in `usecases` from the IR, with no configuration. Their
identities:

| Kind | Produced when | View id | Subject |
|---|---|---|---|
| landscape | at least one element exists | `landscape` | — |
| system | the system has ≥ 1 container | `system-<name>` | `system.<name>` |
| container | the container has ≥ 1 component | `container-<name>` | `container.<name>` |
| deployment | the environment has ≥ 1 instance | `deployment-<env>` | `deployment.<env>` |
| declared | every `view` block, unless empty (FR-005) | `<view label>` | `view.<label>` |

Declared and derived view ids share one namespace. When they collide, the declared view wins and a
`view_shadowed` warning names the derived view it replaced (FR-004). A view id is a function of the
address of what the view depicts, never of position or discovery order (FR-006).

**Rationale**: The ids are readable in a file listing, stable across runs, and cannot collide
between derived kinds, because each carries its kind as a prefix. The roadmap spoke of "one context
view", but the spec calls it *landscape*, because it shows every top-level element rather than one
system's context. This plan follows the spec.

**Alternatives considered**: Using raw addresses as ids (`system.payments`). The dot is
indistinguishable from a file extension in most file browsers, and a declared view could not shadow
one without an awkward label.

---

## R4. What each view contains (projection rules)

**Decision**: One rule serves every view kind. A view has a **visible set** of elements. Every
relationship endpoint is **lifted** to its nearest ancestor-or-self in the visible set. Then:

1. If both lifted ends are visible and distinct, the relationship becomes an edge. Several
   relationships that lift onto the same `(source, target)` pair merge into one edge. The edge
   keeps its constituent relationship addresses, sorted. It is labelled with the single
   relationship's description, or with "`N` relationships" when there are several.
2. If both lifted ends are the *same* element, the edge is kept only when the authored relationship
   was itself a self-connection (`Source == Target`), which becomes a self-loop (edge case). A
   self-edge that exists only because of lifting is internal to that element and is dropped.
3. If exactly one end has a visible ancestor-or-self, the relationship becomes a **boundary
   edge**: an edge between the visible end and the view's single synthetic `outside` node (FR-009).
   Boundary edges merge per `(inside element, direction)`.
4. If neither end is visible, the relationship is not in this view.

Visible sets per kind:

- **landscape**: every element with no parent (persons, systems, externals). Nothing is outside, so
  there are no boundary edges.
- **system `S`**: `S` drawn as a boundary, its containers nested inside, plus every element that is
  the lifted neighbour of one of those containers, lifted to its top-level ancestor.
- **container `C`**: `C` as a boundary, its components inside, plus neighbours lifted to *sibling
  containers* when they are in the same system and to top-level elements otherwise.
- **deployment `E`**: the environment's placement-group tree, reproduced exactly (edge case: five
  levels deep stays five levels), and its instances placed in their groups. Edges are relationships
  between the logical elements the instances realise, lifted to the nearest instantiated
  ancestor-or-self *in `E`*. A relationship whose other end has no instance in `E` becomes a
  boundary edge.
- **declared `V`**: (`include` ∪ descendants of `include`) ∪ (every element carrying any tag in
  `tags`), minus (`exclude` ∪ descendants of `exclude`). When both `include` and `tags` are absent,
  the base set is every element. Nesting is shown between visible elements. Connections that ended
  on an excluded element become boundary edges (US2/AC2).

**Rationale**: Lifting follows the C4 convention for implied relationships. One algorithm for all
five kinds means one set of tests and one place for the boundary rule to be wrong. Neighbours in
system and container views are drawn outside the subject's boundary box, the same way a C4 container
view draws the external systems a container talks to.

**Alternatives considered**: Dropping cross-view edges, which FR-009 forbids. Drawing a separate
stub per hidden endpoint, which is noisy on wide views and leaks "where it goes", which the spec says
the reader should not see.

---

## R5. The backend interface and format dependencies

**Decision**: There is one port, in `ports.go`:

```go
type Backend interface {
    Format() viewmodel.Format
    Render(ctx context.Context, in *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error)
}
```

A `Projection` is the complete intermediate value: the project header, every `ViewModel`, and every
`ElementPage`. The roadmap sketched `Render(ViewModel) ([]byte, error)`, which fits per-view diagram
formats but not the site or markdown, which produce many files and cross-link them. Per-view
backends iterate over `in.Views` themselves.

Formats that embed another format's output by reference declare the dependency in core
(`viewmodel.Format.Requires()`): `html → svg` and `md → svg`. The build use case expands the
requested set with its requirements before rendering, and the build summary names any format that
was added. Backends never call one another, and no backend package imports another (FR-012). The
HTML page emits `<img src="../diagrams/<id>.svg">` from a path function in core, and whichever
backend produces that file is irrelevant to it.

**Rationale**: FR-011 and FR-012 prohibit a backend from reading anything but the projection and
from depending on another backend. A declared, output-level requirement satisfies both, and
removing a backend then means deleting its package, its registry line, and any `Requires` entry
naming it.

**Alternatives considered**:
- *HTML inlines SVG bytes*: the HTML backend would need the SVG backend, or a layout engine of its
  own, which breaks FR-012.
- *HTML embeds D2 source for client-side rendering*: needs a JavaScript runtime in the browser and
  cannot be link-checked.
- *Error when html is requested without svg*: a zero-configuration user would hit it immediately.

---

## R6. Styling by convention (FR-008, FR-016..019)

**Decision**: Style is decided entirely in `viewmodel` (a pure lookup table plus tag mapping) and
carried on every node and edge:

| Element kind | Shape | Fill | Stroke | Dashed |
|---|---|---|---|---|
| person | `person` | `#08427b` | `#073b6f` | no |
| system | `rectangle` | `#1168bd` | `#0b4884` | no |
| container | `rectangle` | `#438dd5` | `#3c7fc0` | no |
| component | `rectangle` | `#85bbf0` | `#78a8d8` | no |
| external | `rectangle` | `#999999` | `#8a8a8a` | **yes** |
| subject boundary / group / environment | `boundary` | none | `#444444` | yes |
| `outside` node | `oval` | none | `#999999` | yes |

These are the standard C4 palette values. Instances take their logical element's style. Every node
carries CSS classes `kind-<kind>`, `external` where it applies, and `tag-<slug>` for each tag
(sorted), so user CSS can target them (FR-018). D2 emits the classes through the `class` keyword,
declaring each one as an empty `classes {}` entry, and d2 ≥ v0.6 writes them onto the SVG `<g>`
element. A golden test pins that attribute, so a d2 upgrade that drops it fails loudly. The HTML
backend puts the same classes on element cards and page bodies.

**Rationale**: A backend that chose colours would be two sources of styling truth. A table in core
is the only way the D2, SVG, markdown and HTML outputs can agree (FR-016).

---

## R7. Generated-file notice (FR-020, SC-004)

**Decision**: One notice text, built in core: `Code generated by loko from <sources>. DO NOT EDIT.`
`<sources>` is the sorted, de-duplicated list of project-relative `*.loko.hcl` files that declared
anything the artifact depicts, joined with `, `. Each backend supplies only the comment syntax for
its format:

| Format | Placement |
|---|---|
| `.d2` | `# …` as line 1 |
| `.svg` | `<!-- … -->` immediately after the XML declaration (a declaration must come first) |
| `.md` | `<!-- … -->` as line 1 |
| `.html` | `<!-- … -->` on line 2, after `<!DOCTYPE html>`, so the doctype remains first and pages stay in standards mode |
| `.css`, `.js` | `/* … */` as line 1 |
| `.loko-manifest` | `# …` as line 1 |

No timestamp, version, host, or absolute path appears. A scanner test walks every file in every
fixture build and asserts the notice is present within the first two lines (SC-004).

Provenance is taken from source ranges. Elements carry `Range.File` in the IR. Environments, groups
and views carry no range in the IR, so the build use case derives a provenance table from the
`SourceModel` it already holds after compiling. The IR shape does not change, and export goldens
from feature 013 stay byte-identical.

---

## R8. Output ownership, pruning and atomicity (FR-022..024)

**Decision**: The output directory holds a manifest, `.loko-manifest`, listing every path the last
build wrote (sorted, one per line, with the generated-file notice). The `ArtifactStore` adapter
implements `Commit`:

1. Verify the directory is writable by creating a staging directory `<out>/.loko-staging-<rand>`.
   If that fails, the build reports an error naming `<out>` and writes nothing.
2. Write each artifact whose bytes differ from what is on disk into staging. Byte-identical files
   are skipped, so their modification times survive, which keeps `make`-style tooling and editors
   quiet (FR-022).
3. Rename staged files into place, remove every path in the *old* manifest that is absent from the
   new set (FR-023), write the new manifest last, and remove the staging directory.
4. Anything not named in the old manifest is never touched: user files in the output directory are
   left alone (edge case).

Compilation errors stop the pipeline before `Commit` is reached, so a broken project leaves the last
good output exactly as it was (FR-024, US1/AC5). The staging directory name is random, but it never
appears in any artifact or manifest.

**Alternatives considered**: Wiping and rewriting the directory, which destroys user files and
defeats mtime stability. Building into a sibling directory and swapping it, which the spec rules out
because the output directory may hold files the tool does not own.

---

## R9. File names: safety, uniqueness, case collisions (FR-006, FR-028)

**Decision**: Paths are computed in core by one function family (`viewmodel.Paths`). Name segments
are escaped injectively: `[A-Za-z0-9.-]` pass through, and every other byte, including `_`, becomes
`_` followed by two lowercase hex digits. `payments/v2` becomes `payments_2fv2`. After every path in
a projection is planned, a case-folded comparison (`strings.ToLower` over the full path) finds pairs
that would collide on a case-insensitive file system. Each pair is reported as an
`output_path_collision` **error** naming both addresses, and nothing is written (FR-028).

**Rationale**: The escape is reversible and so unique by construction. Because it depends only on
the name, it is stable across runs. Checking collisions on planned paths rather than at write time
catches them identically on Linux, where the file system would not complain, and on macOS, where it
would silently overwrite.

---

## R10. Prose in the site and markdown (FR-025, FR-026)

**Decision**: Prose is read before projection through a new `ProseReader` port. It is implemented in
the new `adapters/projectfs`, which resolves `docs` relative to the project root with the same
containment rule as the `docs_not_found` warning. The text is carried on `ElementPage.Prose`. A
missing file sets `ProseMissing` and the page is still produced, with a notice (FR-026). The
compiler has already emitted `docs_not_found` as a warning, so the build does not emit a second one.

The markdown backend inlines the prose verbatim. The HTML backend converts it with
`github.com/yuin/goldmark` (CommonMark), with raw HTML disabled and the GFM table extension on.
goldmark was already an indirect dependency through d2 at `v1.7.4`, so promoting it adds no new
module. It is confined to `internal/adapters/**` by the same archcheck rule that confines d2.

**Alternatives considered**: Restoring the parked hand-written markdown parser
(`_parked/html/markdown_renderer.go`, 403 lines). It is not CommonMark, it mishandles nesting, and
it would be the largest file in the adapter for no gain over a module already in the build.

---

## R11. Theme overrides (FR-034, FR-035, US7)

**Decision**: The built-in site presentation lives as embedded files in
`internal/adapters/html/theme/`: `layout.gohtml`, `index.gohtml`, `element.gohtml`, `view.gohtml`,
`partials.gohtml`, `style.css`, `custom.css` (empty), and `site.js`. Templates use `html/template`,
not the parked `text/template`, so author-supplied descriptions are escaped. Overrides are read from
`<project root>/templates/` through a `ThemeSource` port, and the use case passes them to the
backend as data (`RenderOptions.Theme`), so the backend does not read the file system (FR-011).

Resolution depends on the file type. A `.css` or `.js` override replaces the built-in file of the
same name, byte for byte. A `.gohtml` override is not a whole-file replacement: it is parsed *after*
the built-ins, so each `{{define}}` block it contains (`header`, `footer`, …) replaces the built-in
block of that name, and blocks it omits stay built-in. `custom.css` is loaded after `style.css`, so
additive styling needs no copying. Failures are errors naming the file (FR-035): a template parse
error, a `{{define}}` of a name that no built-in template defines, or a `.gohtml`/`.css`/`.js` file
whose name is not a built-in name. Other file types in `templates/` are ignored.

**Rationale**: `templates/` is the directory the roadmap names. Rejecting unknown names inside it
turns a typo (`sytle.css`) into an error instead of a silently ignored override. Parsing templates
is deterministic, so overrides keep the byte-identical guarantee (US7/AC4).

**Split of `templates.go`**: the 1,753-line parked file becomes embedded `.gohtml`/`.css`/`.js`
assets, which archcheck does not measure, plus Go files each under the 400-line adapter budget. The
Constitution v1.3.0 amendment requires the split to happen during the rework, not after.

---

## R12. Watch without fsnotify (FR-029..033, US5)

**Decision**: A polling watcher in `adapters/watch` using only the standard library. Every 200 ms it
stats the **watched set**: every discovered `*.loko.hcl` (using hclsource's discovery rules, which
already skip `dist/` and dot-directories), the prose files the last good compile referenced, and
`templates/`. It explicitly excludes the output directory. A change is reported once the set has
been stable for one further tick (≈200 ms), so a burst of saves produces a single rebuild (FR-033).
Any path outside the watched set, including generated output, cannot trigger a rebuild (FR-032).
The time from saving to seeing the change is at most ≈400 ms plus the rebuild itself, well inside
the 2 s budget.

**Rationale**: Constitution v1.3.0 records fsnotify under "Removed in v1.0 … so they are not
reintroduced by habit". Polling a few hundred files every 200 ms costs almost nothing and behaves the
same on every OS and file system, including network mounts and Docker bind mounts where inotify
events are lost. It also makes the debounce trivial and testable with an injected clock.

**Alternatives considered**: Reintroducing fsnotify, which would need a constitution amendment for
no user-visible gain at this scale. Having the use case poll, which puts time and I/O in core.

---

## R13. Serve and live reload (FR-029..031)

**Decision**: `adapters/devserver` is a `net/http` server bound to `127.0.0.1:<port>` (default
8080). It serves the site from the **in-memory** artifact set of the last build, and never writes
the output directory. Live reload uses Server-Sent Events on `/_loko/events`. The server injects a
small reload script into HTML responses *at serve time*, so built artifacts never contain dev-only
code and `loko build` output stays byte-identical to what `serve` shows.

When a rebuild fails to compile, the server switches to *error state*. Every HTML request returns a
diagnostics page (file, line and column, rendered by the existing hclsource diagnostic renderer
without colour), and connected browsers are told to reload. Stale pages are therefore never shown
as current (FR-030). The next successful build clears the state and triggers a reload, so the
command recovers without a restart (FR-031).

**Rationale**: SSE is a standard, one-way, stdlib-only mechanism that browsers reconnect
automatically. Serving from memory keeps serve from mutating the committed output directory.

**Note**: ADR-0012 removed the *HTTP API* (`internal/api`). A loopback preview server is a different
thing: an adapter behind a port, used only by `serve`. ADR-0013 records the distinction.

---

## R14. Commands and exit codes (FR-036..039)

**Decision**:

- `loko build [--format d2,svg,md,html] [--out dist] [--strict]`. The default format set is all four
  formats. An unknown format is rejected by name together with the supported list (FR-015).
- `loko serve [--port 8080]`. Watches, rebuilds, and serves on loopback.

Exit codes are the existing three (`0`, `1`, `2` under `--strict`), mapped through
`CompileResult.ExitCode` exactly as `validate` and `export` do (FR-038). The build use case turns
render-stage findings (`view_empty`, `view_shadowed` warnings, the `output_path_collision` error, and
the `theme_invalid` error) into ordinary diagnostics. Each handler parses flags, calls
`usecases.BuildArtifacts` and `store.Commit`, renders diagnostics, and returns: ≤ 50 effective lines
(FR-039, Principle III). The command-set pin test in `cmd/root_test.go` moves `build` and `serve`
from "removed" to "present". `watch` stays removed, because it folds into `serve` (SC-010).

---

## R15. Enforcement additions

**Decision**: Extend `tools/archcheck/rules.yaml` and the `depguard` mirror:

- `github.com/yuin/goldmark/**` joins `hcl`, `cty` and `d2` in every non-adapter layer's
  `forbiddenExternalImports`.
- `os/exec` is forbidden in `internal/**` and `cmd/**` (FR-014's static half; the runtime half is
  the empty-`PATH` test).
- The `mcp` layer forbids `internal/adapters/{d2,html,markdown,outputdir,devserver}/**` (FR-041).

The constitution needs a **MINOR** amendment (v1.4.0, task T023a). Principles I and II name
`main.go` and `cmd/` as the composition root, because `cmd/` already wires adapters for `validate`
and `export` (analysis finding C3). The same amendment fixes stale text: Principle VII still
says "no runtime dependencies except d2 (and optionally veve-cli)", which is now simply "no runtime
dependencies". The adapter tree and dependency table gain `markdown`, `outputdir`, `projectfs`,
`watch`, `devserver`, and goldmark. The entity layer table gains `viewmodel/`, and the Quality Gates entity budget is corrected from 300 to 200.
