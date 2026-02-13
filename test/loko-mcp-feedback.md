# Loko MCP - Experience Feedback & Suggestions

**Date:** 2026-02-13
**Context:** Used loko MCP to model a serverless Notification Service (1 system, 6 containers, 17 components, 7 D2 diagrams)

---

## What Worked Well

### Element CRUD
- `create_system`, `create_container`, `create_component` — clean inputs, predictable outputs, easy to batch.
- `update_component`, `update_system` — straightforward metadata updates.

### Querying
- `query_architecture` with `detail: full` gives a great overview of the entire project in one call.
- `query_project` and `search_elements` are useful for discovery.

### Diagrams
- `update_diagram` accepting raw D2 source and routing it to the correct file path is an excellent abstraction.
- All 5 container-level diagrams were created in a single parallel batch with no issues.

### Documentation
- `build_docs` — one call, clean output, generates browsable HTML with diagrams, component pages, and search index.

### Validation
- `validate` — fast feedback loop, clear severity levels (error/warning/info), actionable suggestions.

---

## What Was Frustrating

### 1. Relationships Are a Dead End

There are **four** relationship-related tools that all return empty results:
- `find_relationships`
- `query_related_components`
- `query_dependencies`
- `analyze_coupling`

The component D2 files include placeholder comments (`# Add component relationships here using the format: component -> other_component: "relationship_type"`), but loko never parses these into the graph model. There is **no `add_relationship` or `create_relationship` tool**, so the entire relationship/coupling subsystem is effectively read-only with nothing to read.

### 2. Unclear Data Model Boundary

It took multiple rounds of trial-and-error to determine where relationships are supposed to live:
- Component D2 files? (visual only, not parsed into graph)
- Component `.md` frontmatter? (no relationship fields supported)
- A separate relationship store? (doesn't exist)

The answer is "nowhere yet." This boundary between visual diagrams and the architecture graph model is not documented or surfaced by the tools.

### 3. Misleading "Isolated Components" Info Message

Every `validate` call returns:
> "17 isolated component(s) found — Review if these components should have relationships with other components"

This suggests the user should fix something, but provides **no mechanism to actually do it**. This creates noise on every validation run.

### 4. Tool Discoverability Gaps

- `update_component` only supports `description`, `technology`, and `tags` — no hint that relationships aren't part of the component metadata.
- No `help`, `capabilities`, or `schema` tool to explain the overall data model and what each tool can/cannot do.

---

## Suggestions

### High Priority

1. **Add a `create_relationship` tool**
   - Parameters: `source` (system/container/component path), `target` (path), `relationship_type` (e.g., `uses`, `depends-on`, `triggers`), `label` (description)
   - This would make `find_relationships`, `query_dependencies`, `query_related_components`, and `analyze_coupling` actually useful.

2. **Or parse relationships from D2 files into the graph model**
   - The component D2 files already have relationship syntax (`component -> target: "label"`).
   - If loko parsed these arrows into the graph, relationships would be populated automatically from diagrams.

3. **Suppress "isolated components" info if no relationship mechanism exists**
   - Or downgrade it to a "note" that explains the current limitation.

### Medium Priority

4. **Add a `capabilities` or `help` tool**
   - Explain the data model: what's stored in `.md` frontmatter vs. D2 files vs. the graph.
   - List which tools are CRUD vs. read-only.
   - Clarify the relationship between visual diagrams and the architecture graph.

5. **Enrich `update_component` with relationship fields**
   - Allow `depends_on: ["system/container/component"]` in component metadata.

### Nice to Have

6. **Add a `delete_relationship` tool** (complement to `create_relationship`).
7. **Support bulk relationship creation** for modeling entire data flows in one call.
8. **Add relationship visualization in `build_docs`** — generate a cross-container dependency graph in the HTML output.

---

## Overall Rating

| Aspect | Rating | Notes |
|--------|--------|-------|
| Element CRUD | Great | Clean, predictable, batchable |
| Diagram authoring | Great | `update_diagram` with raw D2 is excellent |
| Documentation generation | Great | One call, complete output |
| Validation | Good | Fast but noisy with unfixable info messages |
| Relationship modeling | Not functional | 4 tools, 0 results, no way to populate data |
| Discoverability | Needs work | Trial-and-error to understand data model boundaries |

**Summary:** The CRUD and diagram authoring workflow is solid and productive. The relationship/graph features feel like scaffolding for something not yet built.
