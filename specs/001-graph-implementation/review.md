# Graph (Node) Implementation Review

**Date:** 2026-02-12
**Scope:** `internal/core/entities/graph.go`, `internal/core/usecases/build_architecture_graph.go`, `internal/mcp/tools/graph_tools.go`, `internal/mcp/tools/tools.go`, `internal/core/usecases/validate_architecture.go`

---

## Summary

The `ArchitectureGraph` is the core data structure that converts loko's hierarchical C4 model (System → Container → Component) into a directed graph for dependency analysis, cycle detection, pathfinding, and coupling metrics. The implementation is a custom adjacency-list graph in Go.

Overall the design is sound and well-suited to the C4 domain. The issues below are ordered by severity — the first one is a correctness bug, the rest are structural improvements to address before the codebase grows.

---

## P0 — Node ID Collision Across Systems

**Files:** `build_architecture_graph.go` lines 90–113, `entities/component.go` line 51

**Problem:** Component node IDs are derived from `NormalizeName(name)`, which produces flat strings like `"auth"` or `"payment-handler"`. The graph builder uses `component.ID` directly as the graph node key. If two systems (or two containers) each have a component named `Auth`, the second `AddNode` call fails because the ID `"auth"` already exists.

```go
// build_architecture_graph.go — current behavior
componentNode := &entities.GraphNode{
    ID:       component.ID,   // just "auth" — no namespace
    ParentID: container.ID,
    // ...
}
```

`AddNode` rejects duplicates, but the error is swallowed in the build loop, meaning the second component silently disappears from the graph. Its relationships also never get wired up.

**Impact:** Any multi-system project with overlapping component names produces an incomplete and silently incorrect graph. Validation, coupling analysis, and MCP queries all operate on a partial picture.

**Fix:** Use qualified IDs that include the hierarchy path:

```go
componentNodeID := fmt.Sprintf("%s/%s/%s", system.ID, container.ID, component.ID)
```

Apply the same pattern to container IDs (`system.ID + "/" + container.ID`) to prevent collisions there too. The `Component.Relationships` map would need to store qualified target IDs, or the graph builder would need a lookup table from short IDs to qualified IDs.

---

## P1 — `GetIncomingEdges` Is O(E), Called in Loops

**File:** `graph.go` lines 149–159

**Problem:** `GetIncomingEdges` scans every edge list in the graph to find edges pointing at a given node. This makes it O(E) per call. It's then called inside loops in `checkIsolatedComponents` and `checkHighCoupling` (via `GetDependents`), making those validation checks O(N×E).

```go
func (ag *ArchitectureGraph) GetIncomingEdges(nodeID string) []*GraphEdge {
    var incoming []*GraphEdge
    for _, edges := range ag.Edges {       // scan ALL edge lists
        for _, edge := range edges {
            if edge.Target == nodeID {
                incoming = append(incoming, edge)
            }
        }
    }
    return incoming
}
```

For current C4 model sizes (< 200 nodes) this is negligible, but it becomes a problem if loko ever handles larger models or if the validation/analysis is called in a hot loop (e.g., watch mode rebuilds).

**Fix:** Add a reverse adjacency map to `ArchitectureGraph` and maintain it in `AddEdge`:

```go
type ArchitectureGraph struct {
    Nodes          map[string]*GraphNode
    Edges          map[string][]*GraphEdge  // outgoing
    IncomingEdges  map[string][]*GraphEdge  // reverse index
    ParentMap      map[string]string
}
```

```go
func (ag *ArchitectureGraph) AddEdge(edge *GraphEdge) error {
    // ... existing validation ...
    ag.Edges[edge.Source] = append(ag.Edges[edge.Source], edge)
    ag.IncomingEdges[edge.Target] = append(ag.IncomingEdges[edge.Target], edge)
    // ... bidirectional handling ...
}
```

---

## P1 — `GetChildren` Is O(N)

**File:** `graph.go` lines 170–180

**Problem:** `GetChildren` iterates the entire `ParentMap` to find children of a node. `GetDescendants` calls `GetChildren` recursively, compounding the cost.

```go
func (ag *ArchitectureGraph) GetChildren(nodeID string) []*GraphNode {
    var children []*GraphNode
    for childID, parentID := range ag.ParentMap {   // scan ALL entries
        if parentID == nodeID {
            // ...
        }
    }
    return children
}
```

**Fix:** Add a `ChildrenMap` maintained alongside `ParentMap`:

```go
type ArchitectureGraph struct {
    // ...
    ParentMap   map[string]string
    ChildrenMap map[string][]string  // parent ID -> child IDs
}
```

Update `AddNode` to populate both maps when `ParentID` is set.

---

## P2 — No Duplicate Edge Prevention

**File:** `graph.go` lines 109–141

**Problem:** `AddEdge` appends edges without checking whether an identical edge already exists. If `BuildArchitectureGraph` is called twice or if a component's `Relationships` map somehow contains a duplicate entry through a different code path, the graph will contain duplicate edges. This inflates `EdgeCount()`, skews coupling metrics, and produces duplicate entries in dependency queries.

**Fix:** Check for existing edges before appending:

```go
func (ag *ArchitectureGraph) AddEdge(edge *GraphEdge) error {
    // ... validation ...
    for _, existing := range ag.Edges[edge.Source] {
        if existing.Target == edge.Target && existing.Type == edge.Type {
            return nil // already exists
        }
    }
    ag.Edges[edge.Source] = append(ag.Edges[edge.Source], edge)
    // ...
}
```

---

## P2 — Graph Rebuilt on Every MCP Tool Call

**File:** `internal/mcp/tools/graph_tools.go`

**Problem:** Every MCP tool (`QueryDependenciesTool`, `QueryRelatedComponentsTool`, `AnalyzeCouplingTool`) loads the full project, lists all systems, and builds the graph from scratch on each invocation. During an LLM conversation the MCP server may receive dozens of queries in sequence — each one rebuilds the same graph.

```go
func (t *QueryDependenciesTool) Call(ctx context.Context, args map[string]any) (any, error) {
    project, err := t.repo.LoadProject(ctx, projectRoot)    // disk I/O
    systems, err := t.repo.ListSystems(ctx, projectRoot)    // disk I/O
    graph, err := graphBuilder.Execute(ctx, project, systems) // rebuild
    // ...
}
```

**Fix:** Cache the built graph at the MCP server level, keyed by project root. Invalidate when the file watcher detects changes under `src/`. This keeps MCP queries fast while still reflecting edits.

```go
type MCPServer struct {
    graphCache map[string]*cachedGraph // projectRoot -> cached
}

type cachedGraph struct {
    graph   *entities.ArchitectureGraph
    builtAt time.Time
}
```

---

## P2 — No `RemoveNode` / `RemoveEdge`

**File:** `graph.go`

**Problem:** The graph is append-only. There is no way to remove a node or edge without rebuilding the entire graph. This is fine for the current build-once-query-many pattern, but becomes limiting if:

- The MCP server supports interactive editing (add/remove components during a conversation)
- Hot-reload needs to incrementally update the graph instead of rebuilding

**Fix:** Add `RemoveNode` and `RemoveEdge` methods. `RemoveNode` should also clean up all edges referencing the node and remove it from `ParentMap`/`ChildrenMap`.

---

## P3 — `GraphNode.Data` Is `any`

**File:** `graph.go` line 44

**Problem:** The `Data` field holds `*System`, `*Container`, or `*Component` as `any`. Every consumer needs a type switch to use it, and the compiler can't catch mistakes.

```go
Data any  // could be *System, *Container, or *Component
```

**Fix (option A — interface):**

```go
type C4Entity interface {
    EntityID() string
    EntityName() string
    EntityType() string  // "system", "container", "component"
}
```

Have `System`, `Container`, and `Component` implement it, then change `Data` to `C4Entity`.

**Fix (option B — keep it simple):** If you don't need to call methods on `Data` from graph code, this is low priority. Just document the expected types.

---

## P3 — Isolated Component Check Includes Systems and Containers

**File:** `validate_architecture.go` lines 184–211

**Problem:** `checkIsolatedComponents` iterates `graph.Nodes` which includes systems and containers. Systems and containers never have dependency edges (only hierarchy), so they always appear "isolated." The current implementation flags them alongside truly isolated components.

```go
for nodeID := range graph.Nodes {   // includes systems and containers
    deps := graph.GetDependencies(nodeID)
    dependents := graph.GetDependents(nodeID)
    if len(deps) == 0 && len(dependents) == 0 {
        isolated = append(isolated, nodeID)
    }
}
```

**Fix:** Filter to components only:

```go
for nodeID, node := range graph.Nodes {
    if node.Type != "component" {
        continue
    }
    // ...
}
```

The same issue applies to `checkHighCoupling` — it should probably only evaluate components.

---

## P3 — Thread Safety

**File:** `graph.go`

**Problem:** The graph struct has no synchronization. With `Parallel: true` and `MaxWorkers: 4` in the project config, concurrent access during build or query phases could cause data races.

Currently the graph is built sequentially and then queried read-only, so this isn't an active bug. But it's an implicit contract that isn't enforced or documented.

**Fix (minimal):** Add a comment documenting that the graph must be built before concurrent reads, and that concurrent writes are not supported.

**Fix (robust):** Add `sync.RWMutex` if you plan to support concurrent reads during builds or allow mutation after construction.

---

## P2 — MCP Tool Schemas Can Drift from Argument Parsing

**Files:** `internal/mcp/tools/tools.go`, `internal/mcp/tools/graph_tools.go`

**Problem:** Every MCP tool defines its input schema as a hand-written `map[string]any` in `InputSchema()`, then parses arguments with separate type assertions in `Call()`. These two surfaces are not connected — the compiler cannot catch when one changes without the other.

For example, `CreateSystemTool.InputSchema()` declares `"responsibilities"` as an array property (line 50), and `Call()` parses it with a separate assertion (line 105). If a dev adds a new field to the schema but forgets the corresponding parse line in `Call()`, or renames a field in one place but not the other, the tool silently ignores input or panics at runtime.

This pattern repeats across all 9+ MCP tools in the codebase.

```go
// InputSchema() — declares the contract
"responsibilities": map[string]any{
    "type":  "array",
    "items": map[string]any{"type": "string"},
},

// Call() — parses independently, can drift
responsibilitiesIface, _ := args["responsibilities"].([]any)
```

**Fix:** Define a struct per tool that serves as both the schema source and the parse target. Use struct tags to generate the JSON Schema, and `json.Unmarshal` the args into the struct. This keeps the schema and parsing in one place.

```go
type CreateSystemArgs struct {
    ProjectRoot      string   `json:"project_root" jsonschema:"required,description=Root directory"`
    Name             string   `json:"name" jsonschema:"required,description=System name"`
    Description      string   `json:"description,omitempty"`
    Responsibilities []string `json:"responsibilities,omitempty"`
    // ...
}
```

There are lightweight Go libraries for this (e.g., `invopop/jsonschema`), or you can write a small helper that reflects over struct tags. The effort is moderate but it eliminates an entire class of bugs across all MCP tools.

---

## P2 — `AnalyzeDependencies` Returns `map[string]any`

**File:** `build_architecture_graph.go` lines 197–246

**Problem:** `AnalyzeDependencies` returns `map[string]any`, which means every consumer has to know the magic string keys (`"isolated_components"`, `"highly_coupled_components"`, etc.) and type-assert the values. This is the same `any`-erosion problem as `GraphNode.Data` but in a use case return type — a place where a concrete struct costs nothing and prevents runtime surprises.

```go
func (uc *BuildArchitectureGraph) AnalyzeDependencies(
    graph *entities.ArchitectureGraph,
) map[string]any {
    report := make(map[string]any)
    report["systems_count"] = len(systems)
    report["isolated_components"] = isolated
    report["highly_coupled_components"] = highly_coupled
    // ...
}
```

Consumers (like `AnalyzeCouplingTool.Call()`) then have to do:

```go
report["isolated_components"]   // what type is this? []string? []any?
report["highly_coupled_components"]  // map[string]int? map[string]any?
```

**Fix:** Replace with a typed struct:

```go
type DependencyReport struct {
    SystemsCount          int
    ContainersCount       int
    ComponentsCount       int
    TotalNodes            int
    TotalEdges            int
    IsolatedComponents    []string
    HighlyCoupled         map[string]int  // componentID -> dependency count
    CentralComponents     map[string]int  // componentID -> dependent count
}
```

This is a small, low-risk change. The struct can still be serialized to JSON for the MCP tools.

---

## P3 — Cross-Cutting Decisions Are Implicit

**Problem:** Clean Architecture defines clear layers (entities, use cases, adapters), but some decisions cut across those layers and aren't documented anywhere:

- **How node IDs are constructed** — `NormalizeName()` is called in entities, the graph builder, and MCP tools, but the "flat ID" assumption isn't documented. Fixing the P0 collision issue will require changing this convention in multiple places.
- **Graph lifecycle** — the graph is built per-request in MCP tools but could be cached. There's no documented contract about when the graph is valid or stale.
- **Relationship scope** — only components have `Relationships`; systems have `Dependencies` and `ExternalSystems` as metadata strings, not graph edges. This is a design choice but it's not stated anywhere.

**Fix:** Add a short ADR (Architecture Decision Record) in `docs/` covering these conventions. The `docs/` directory already exists, so this fits the project's structure. Something like `docs/adr-003-graph-conventions.md` covering ID format, graph lifecycle, and what gets modeled as edges vs. metadata. This pays for itself the first time a contributor asks "why doesn't the graph include system-level dependencies?"

---

## Not an Issue

A few things that might look concerning but are actually fine given the project's domain:

- **Custom graph instead of a library:** C4 models are small (< 500 nodes typically). A library like `gonum/graph` would add complexity without meaningful benefit.
- **BFS storing full paths in queue:** Memory-inefficient in theory, but for graphs this small it doesn't matter.
- **Relationships only at component level:** This matches the C4 model correctly — components are the units that communicate. System-level dependencies are modeled differently (as metadata, not graph edges).

---

## Suggested Priority Order

| Priority | Issue | Effort |
|----------|-------|--------|
| P0 | Node ID collision across systems | Medium — requires qualified IDs and updating relationship references |
| P1 | Reverse adjacency map for incoming edges | Small — add map + update `AddEdge` |
| P1 | Children map for hierarchy lookups | Small — add map + update `AddNode` |
| P2 | Duplicate edge prevention | Small |
| P2 | MCP tool schema drift from argument parsing | Medium — define arg structs, add schema generation |
| P2 | `AnalyzeDependencies` returns `map[string]any` | Small — replace with typed struct |
| P2 | Graph caching in MCP server | Medium |
| P2 | `RemoveNode` / `RemoveEdge` | Medium |
| P3 | `Data any` → interface | Small |
| P3 | Filter isolated check to components only | Small |
| P3 | Cross-cutting decisions not documented | Small — add ADR in `docs/` |
| P3 | Thread safety documentation or mutex | Small |
