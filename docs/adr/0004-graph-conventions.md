# ADR 0004: Architecture Graph Conventions

## Status

Accepted

## Context

The loko architecture documentation system uses a directed graph to model C4 architecture relationships. As projects grew to include multiple systems, we encountered node ID collisions when different systems used components with the same names (e.g., "auth" component in both "backend" and "admin" systems). Additionally, the graph's thread safety model, lifecycle, and relationship scope needed clear documentation to prevent misuse.

Key issues motivating this ADR:
1. **Node ID Collisions**: Component names like "auth" or "database" are common across systems, causing silent data loss when stored in a map with non-unique keys
2. **Thread Safety Confusion**: Unclear whether ArchitectureGraph can be accessed concurrently by MCP tools
3. **Graph Lifecycle Ambiguity**: When is a graph built? When is it cached? When is it invalidated?
4. **Relationship Scope Questions**: Why don't systems have dependency edges in the graph?

## Decision

We establish the following conventions for the ArchitectureGraph implementation:

### 1. Node ID Format: Qualified Hierarchical IDs

**Decision**: Use qualified IDs that encode the full hierarchy path.

**Format**:
- **System**: `systemID` (e.g., `"backend"`)
- **Container**: `systemID/containerID` (e.g., `"backend/api"`)  
- **Component**: `systemID/containerID/componentID` (e.g., `"backend/api/auth"`)

**Example**:
```
Multi-system project:
  backend/api/auth      ← Component in backend system
  admin/ui/auth         ← Different component in admin system
  payment/worker/queue  ← Component in payment system
```

**Implementation**: 
- Helper functions: `QualifiedNodeID()` generates IDs, `ParseQualifiedID()` parses them
- ShortIDMap: Maps short IDs (e.g., "auth") to all qualified IDs for backward compatibility
- ResolveID(): Resolves short IDs when unambiguous (only one match)

### 2. Thread Safety Model

**Decision**: ArchitectureGraph is **NOT thread-safe** by design. Thread safety is provided at a higher level via GraphCache.

**Rationale**:
- Graphs are **immutable after construction** (build once, read many times)
- Synchronization overhead would penalize the common read-heavy workload
- GraphCache (in `internal/mcp`) provides thread-safe caching with `sync.RWMutex`

**Usage Pattern**:
```go
// CORRECT: Use GraphCache for concurrent access
cache := mcp.NewGraphCache()
graph, _ := cache.Get(projectRoot)  // Thread-safe read

// INCORRECT: Don't modify graph after caching
graph.AddNode(newNode)  // ❌ Breaks immutability contract
```

**Guarantees**:
- ✅ Multiple goroutines can safely **read** from the same cached graph
- ✅ GraphCache handles concurrent Get/Set with RWMutex
- ❌ Modifying a graph after construction is undefined behavior

### 3. Graph Lifecycle

**Decision**: Define a clear 5-stage lifecycle for ArchitectureGraph instances.

**Stages**:

1. **Construction**: `NewArchitectureGraph()` creates empty graph with initialized maps
2. **Population**: `BuildArchitectureGraph` use case adds nodes and edges
3. **Freezing**: Graph is treated as immutable after `Execute()` returns
4. **Caching**: `GraphCache.Set()` stores graph for reuse (e.g., during MCP session)
5. **Reading**: Multiple MCP tools read cached graph concurrently

**Invalidation**:
- Manual: `GraphCache.Invalidate(projectRoot)` when project files change
- Automatic: Not implemented (cache entries live for MCP session lifetime)

**Example Flow**:
```
User edits loko.toml
  ↓
Claude Code detects file change
  ↓
MCP server invalidates cache for project
  ↓
Next MCP tool call rebuilds graph
  ↓
New graph cached for subsequent calls
```

### 4. Relationship Scope

**Decision**: Only **component-level relationships** create graph edges. System and container relationships are **hierarchical only** (via ParentMap).

**Why systems don't have dependency edges**:
- C4 model defines relationships at component level (Level 3)
- Systems and containers are **grouping constructs**, not operational units
- System "dependencies" are aggregated from their components' relationships

**Example**:
```
backend system
  ├── api container
  │   └── auth component ──[uses]──> database component
  └── worker container

admin system  
  └── ui container
      └── dashboard component ──[calls]──> backend/api/auth
```

**Graph Edges**:
- ✅ `admin/ui/dashboard` → `backend/api/auth` (component-to-component)
- ✅ `backend/api/auth` → `backend/api/database` (within same system)
- ❌ No edge between `admin` system and `backend` system

**Validation Impact**:
- Isolated component checks (`checkIsolatedComponents`) filter to `Type == "component"`
- High coupling checks (`checkHighCoupling`) filter to `Type == "component"`
- Systems/containers without relationship edges are **not flagged** as issues

**Querying System Dependencies**:
```go
// To find all external systems a system depends on:
systemGraph := graphBuilder.GetSystemGraph(graph, systemID)
for _, component := range systemGraph.Nodes {
    deps := systemGraph.GetDependencies(component.ID)
    for _, dep := range deps {
        if !strings.HasPrefix(dep.ID, systemID) {
            // External dependency to another system
        }
    }
}
```

## Consequences

### Positive

1. **No Silent Data Loss**: Qualified IDs eliminate collisions in multi-system projects
2. **Clear Concurrency Model**: Developers know to use GraphCache for thread safety
3. **Predictable Lifecycle**: Graph immutability enables aggressive caching
4. **Accurate Validation**: Component-only checks prevent false positives for systems/containers
5. **Self-Documenting Code**: Qualified IDs make it obvious which system a component belongs to

### Negative

1. **Migration Required**: Existing projects with short IDs need migration guide
2. **Longer IDs**: `backend/api/auth` vs `auth` increases memory/storage slightly
3. **Breaking Change**: MCP tools using short IDs must handle ambiguous resolution
4. **Cache Invalidation**: Manual invalidation required when files change (no auto-detection)

### Mitigations

1. **ShortIDMap**: Enables backward compatibility with short ID queries
2. **ResolveID()**: Gracefully handles ambiguous IDs with clear error messages
3. **Migration Guide**: Document in `docs/migration-001-graph-qualified-ids.md`
4. **Validation Tests**: Integration tests ensure no regressions in ID handling
5. **GraphCache API**: Provides `Invalidate()` for explicit cache control

## References

- Implementation: `internal/core/entities/graph.go`
- Caching: `internal/mcp/graph_cache.go`
- Use Cases: `internal/core/usecases/build_architecture_graph.go`
- Tests: `internal/core/entities/graph_test.go`, `tests/integration/graph_collision_test.go`
- Related: ADR 0001 (Clean Architecture), ADR 0002 (Token-Efficient MCP)
