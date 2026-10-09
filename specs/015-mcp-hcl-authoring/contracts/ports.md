# Contract: new ports and use cases

**Feature**: `015-mcp-hcl-authoring` | **Status**: normative

New interfaces live in `internal/core/usecases/ports.go` (Principle II).

```go
// OverlaySource compiles the project with some files replaced by in-memory
// contents. overlay maps project-relative paths to bytes; a path absent on
// disk is a new file.
type OverlaySource interface {
    ArchitectureSource // the same adapter compiles with and without an overlay
    LoadOverlay(ctx context.Context, root string, overlay []authoring.FileContent) (*arch.SourceModel, arch.Diagnostics, error)
}

// SourceEditor turns validated edits into new file contents. It never writes
// until Commit, and Commit writes only *.loko.hcl files inside root.
type SourceEditor interface {
    // Plan applies edits in order to in-memory copies of the affected files
    // and returns their new contents. It reads the current disk state.
    Plan(ctx context.Context, root string, edits []authoring.Edit) (authoring.Plan, error)
    // Diff renders a unified diff per changed file, sorted by path.
    Diff(plan authoring.Plan) []authoring.FileChange
    // Revision hashes the current source files.
    Revision(ctx context.Context, root string) (authoring.Revision, error)
    // Commit writes the plan atomically, refusing if any changed file's
    // current hash differs from base.
    Commit(ctx context.Context, root string, plan authoring.Plan, base authoring.Revision) error
}
```

- `authoring.Plan` is `{Files []FileContent{Path, Old, New []byte}, Removed []string}`, sorted by
  path. `Removed` lists cascade removals.
- Plan errors are typed: `*authoring.EditError{Index, Reason, Detail, Dependents}`, which the use
  case maps to a `Refusal`.

## Use cases

| Function / type | File | Responsibility |
|---|---|---|
| `Describe(ctx, deps, req) (*DescribeResult, error)` | `describe.go` | compile, then project a level or scope (R6) |
| `Query(ctx, deps, req) (*QueryResult, error)` | `query.go`, `query_graph.go` | the graph algorithms (R5); shared by MCP and CLI |
| `Validate(ctx, deps) (*ValidateResult, error)` | `validate_project.go` | compile + diagnostics + revision |
| `AuthoringService{mu sync.Mutex; deps}` | `authoring_service.go` | serialises writes (FR-017) |
| `(*AuthoringService).Apply(ctx, batch) (*authoring.EditResult, error)` | `apply_edits.go`, `apply_plan.go` | the state flow in data-model §5 |
| `(*AuthoringService).Move(ctx, from, to, base, preview)` | `apply_edits.go` | one rename edit |
| `checkDangling(ir, edit)`, `cascadeSet(ir, edit)` | `apply_dependents.go` | FR-018/018a, computed on the compiled IR |

Every use-case file is ≤ 200 effective lines and every function ≤ 60.
