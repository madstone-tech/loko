# Data Model: Dynamic Views

Fields added per layer. Everything new in exported or encoded types is `omitempty`, so models
without steps encode exactly as in 1.2.0.

## `arch` (source model)

| Type | Field | Type | Note |
|---|---|---|---|
| `ViewDecl` | `Style`, `StyleRange` | string, `SourceRange` | `edges` or `sequence` |
| `ViewDecl` | `Steps` | `[]StepDecl` | declaration order |
| `ViewDecl` | `IncludeRange`, `ExcludeRange`, `TagsRange` | `SourceRange` | to report `dynamic_view_selection` at the attribute |
| `StepDecl` | `Name` | string | the label |
| `StepDecl` | `From`, `To`, `Via` | `Reference` | `Via` zero when unset |
| `StepDecl` | `Description` | string | |
| `StepDecl` | `Reply` | bool | |
| `StepDecl` | `Range` | `SourceRange` | the block |

Address: `NewStepAddress(view, label)` → `view.<view>.step.<label>`, declared in the symbol table.

## `arch` (IR)

| Type | Field | JSON | Note |
|---|---|---|---|
| `View` | `Style` | `style,omitempty` | |
| `View` | `Steps` | `steps,omitempty` | declaration order |
| `Step` | `Address` | `address` | |
| `Step` | `Name` | `name` | |
| `Step` | `Number` | `number` | 1-based position |
| `Step` | `From`, `To` | `from`, `to` | resolved addresses |
| `Step` | `Via` | `via,omitempty` | as written, resolved |
| `Step` | `Relationship` | `relationship,omitempty` | `via` or the match; empty when unmatched |
| `Step` | `Reply` | `reply,omitempty` | |
| `Step` | `Description` | `description,omitempty` | |
| `Step` | `Range` | `range` | |

## `arch` (diagnostics)

`CodeStepMismatch` (error), `CodeStepUnmatched` (warning, in `IgnorableCodes`),
`CodeDynamicViewSelection` (error). `AllCodes`: 37.

## `authoring`

| Item | Change |
|---|---|
| `TargetStep` | new target `step`; `legalAttrs`: `from`, `to`, `via` (`ValueRef`), `description` (`ValueString`), `reply` (`ValueBool`) |
| `requiredOnAdd[TargetStep]` | `from`, `to` |
| `TargetView` | gains `style`; `enumAttrs[TargetView]["style"] = {"edges", "sequence"}` |
| `Edit.Before` | string, step label; only valid on `add` of a `step` |

## `viewmodel`

| Type | Field | JSON | Note |
|---|---|---|---|
| `View` | `Style` | `style,omitempty` | `sequence`, or empty for edges |
| `ViewModel` | `Steps` | `steps,omitempty` | `[]StepRow`, in order |
| `StepRow` | `Number` | `number` | |
| `StepRow` | `Name` | `name` | |
| `StepRow` | `From`, `To` | `from`, `to` | `LinkRef` to the element's page |
| `StepRow` | `FromNode`, `ToNode` | `fromNode`, `toNode` | node IDs, for the sequence emitter |
| `StepRow` | `Text` | `text,omitempty` | description or the relationship's |
| `StepRow` | `Technology` | `technology,omitempty` | |
| `StepRow` | `Reply` | `reply,omitempty` | |

## `usecases`

| Item | Change |
|---|---|
| `ViewInfo` (describe) | `Style`, `Steps []StepInfo` |
| `QueryRequest.Kind` | adds `flows` |
| `QueryResult` | `Flows []FlowHit` (view, step number, label, role `from`/`to`) |
