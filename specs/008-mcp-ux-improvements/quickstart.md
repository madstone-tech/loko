# Developer Quickstart: MCP UX Improvements (008)

**Branch**: `008-mcp-ux-improvements`  
**Date**: 2026-02-19

This guide walks a developer through implementing this feature from scratch, in the order that minimizes rework and keeps tests green at each step.

---

## Prerequisites

```bash
# Confirm you are on the right branch
git branch --show-current
# → 008-mcp-ux-improvements

# Confirm existing tests pass before starting
go test ./...
task lint
```

---

## Implementation Order

Work in this order — each step builds on the previous and can be tested independently.

### Step 1 — `Relationship` entity (core/entities)

**Files**: `internal/core/entities/relationship.go`, `internal/core/entities/relationship_test.go`

Write the test first:
```bash
# Create test, watch it fail
go test -run TestNewRelationship -v ./internal/core/entities/
```

Implement `Relationship` struct, `RelationshipsFile` struct, `NewRelationship()` constructor with validation, `GenerateRelationshipID()` (SHA-256 based), and option funcs (`WithRelType`, `WithRelTechnology`, `WithRelDirection`).

Key validation cases to cover:
- Empty source → error
- Empty target → error
- Source == target → error
- Empty label → error
- Invalid type → error
- Valid defaults (type="sync", direction="forward")
- Idempotent ID generation (same inputs → same ID)

```bash
go test -run TestRelationship -v ./internal/core/entities/
```

---

### Step 2 — `RelationshipRepository` port (core/usecases/ports.go)

**File**: `internal/core/usecases/ports.go` (append only)

Add the `RelationshipRepository` interface with three methods:
- `LoadRelationships(ctx, projectRoot, systemID) ([]entities.Relationship, error)`
- `SaveRelationships(ctx, projectRoot, systemID, rels []entities.Relationship) error`
- `DeleteElement(ctx, projectRoot, systemID, elementPath string) error`

No test needed at this step (interface definition only). Run lint:
```bash
task lint
```

---

### Step 3 — `RelationshipRepository` filesystem adapter

**Files**: `internal/adapters/filesystem/relationship_repo.go`, `internal/adapters/filesystem/relationship_repo_test.go`

Write integration test using `t.TempDir()`:
```bash
go test -run TestRelationshipRepo -v ./internal/adapters/filesystem/
```

Implement:
- `LoadRelationships`: read `src/<systemID>/relationships.toml`; return empty slice if file absent (not error)
- `SaveRelationships`: marshal `RelationshipsFile`; atomic write (`*.tmp` → rename)
- `DeleteElement`: load → filter → save

Verify atomic write is correct: test that a power-loss simulation (write fails mid-stream) leaves no corrupted file.

---

### Step 4 — Three relationship use cases

**Files** (one pair each): `create_relationship.go` + test, `list_relationships.go` + test, `delete_relationship.go` + test

Use a concrete mock for `RelationshipRepository` (no mocking libraries per constitution):

```go
type mockRelationshipRepo struct {
    rels map[string][]entities.Relationship // keyed by systemID
}
// implement LoadRelationships, SaveRelationships, DeleteElement
```

Test the use cases fully:
- `CreateRelationship`: happy path, idempotent (duplicate returns existing), validation errors
- `ListRelationships`: empty project returns `[]`, source filter, target filter
- `DeleteRelationship`: happy path, not-found returns `ErrNotFound`

```bash
go test -run TestCreateRelationship -v ./internal/core/usecases/
go test -run TestListRelationships  -v ./internal/core/usecases/
go test -run TestDeleteRelationship -v ./internal/core/usecases/
```

---

### Step 5 — D2 edge generation in `create_relationship` use case

Extend `CreateRelationship.Execute` to:
1. Determine the correct diagram file path (system.d2 for container→container; container.d2 for same-container component→component)
2. Regenerate the edges section of that file from the full `relationships.toml` contents
3. Write the file (reuse existing `os.WriteFile` pattern from `scaffold_entity.go`)

Add `RelationshipToD2Edge(rel entities.Relationship) string` as a package-level function in `entities/relationship.go` (pure, no I/O, easy to unit test).

```bash
go test -run TestRelationshipToD2Edge -v ./internal/core/entities/
```

---

### Step 6 — Validate suppression fix

**File**: `internal/core/usecases/validate_architecture.go`

Add the two-line guard at the top of `checkIsolatedComponents`:

```go
if graph.EdgeCount() == 0 {
    return
}
```

Update the existing test in `validate_architecture_test.go` to assert that a graph with components but zero edges emits no `isolated_component` findings.

```bash
go test -run TestValidateArchitecture -v ./internal/core/usecases/
```

---

### Step 7 — `suggestSlugID` + `notFoundError` helpers

**File**: `internal/mcp/tools/helpers.go`

Add the two helper functions. Write a test for `suggestSlugID` using a small constructed `ArchitectureGraph`.

Update existing tools to use `notFoundError`: `update_component.go`, `update_container.go`, `update_system.go`, `graph_tools.go`.

```bash
go test -run TestSuggestSlugID -v ./internal/mcp/tools/
```

---

### Step 8 — Three new MCP tool handlers

**Files**: `create_relationship.go`, `list_relationships.go`, `delete_relationship.go` (all in `internal/mcp/tools/`)

Each handler:
1. Parses input from `args map[string]any`
2. Calls the corresponding use case
3. Calls `t.graphCache.Invalidate(projectRoot)` on success
4. Returns response map

Each handler must stay under 100 lines. Add constructors that accept `RelationshipRepository` and `*mcp.GraphCache`.

Register all three in `registry.go`.

```bash
go test -run TestCreateRelationshipTool -v ./internal/mcp/tools/
```

---

### Step 9 — `create_components` batch tool

**File**: `internal/mcp/tools/create_components.go`

Implement the handler that loops over a `components` array and calls `ScaffoldEntity` per item. Return a `created`/`failed`/`results` response. Under 100 lines.

Register in `registry.go`.

```bash
go test -run TestCreateComponentsTool -v ./internal/mcp/tools/
```

---

### Step 10 — `create_container` diagram initialization fix

**File**: `internal/mcp/tools/create_container.go` and `registry.go`

Modify `CreateContainerTool` to accept `DiagramGenerator` as a constructor parameter. Pass it via `WithDiagramGenerator` to `ScaffoldEntity`. Update `registry.go` to wire `d2.NewGenerator()`.

The existing response code at line 105-108 already handles the `result.DiagramPath != ""` branch — no further change needed there.

```bash
go test -run TestCreateContainerTool -v ./internal/mcp/tools/
```

---

## Final Checks

```bash
# All tests pass
go test ./...

# Coverage on core/ > 80%
task coverage

# Lint clean
task lint

# Build succeeds
make build

# Smoke test with real project
./loko mcp  # connect with an MCP client and run create_relationship
```

---

## Common Pitfalls

| Pitfall | Fix |
|---|---|
| `relationships.toml` not created on first `create_relationship` | Ensure `SaveRelationships` calls `os.MkdirAll` on the system dir before write |
| D2 edges duplicated after multiple `create_relationship` calls | Use regeneration (not append) for the edges section |
| `graphCache.Invalidate` not called → stale query results | Always call after a successful use case execution in the handler |
| `isolated_component` suppression fires when D2 has edges but `relationships.toml` is empty | Suppression checks `graph.EdgeCount()` which includes D2-parsed edges — correct behaviour |
| `create_components` handler exceeds 100 lines | Extract per-item processing into a private function in the same file |
