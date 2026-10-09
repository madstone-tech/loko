# Data Model: Rendering Fidelity

**Feature**: `016-rendering-fidelity` | **Research**: [research.md](research.md)

All additions are optional and omitted when unset. No existing field changes meaning.

## 1. Source model (`internal/core/entities/arch/source_model.go`)

| Type | New field | Type | Notes |
|---|---|---|---|
| `ElementDecl` | `Title` | string | any element |
| `ElementDecl` | `Shape` | string | validated in core (R7) |
| `RelationDecl` | `Kind` | string | `""` means `sync` |
| `RelationDecl` | `Tags` | `[]string` | as authored |
| `ViewDecl` | `Direction` | string | `""` means the view's default (R5) |

Attribute ranges are recorded in the existing `AttrRanges` map, so diagnostics point at the value.

## 2. IR (`internal/core/entities/arch`)

| Type | New field | JSON/TOON | Notes |
|---|---|---|---|
| `Element` | `Title` | `title,omitempty` | |
| `Element` | `Shape` | `shape,omitempty` | |
| `Relationship` | `Kind` | `kind,omitempty` | omitted for `sync` and unset, so sync never appears in exports |
| `Relationship` | `Tags` | `tags,omitempty` | sorted, de-duplicated |
| `View` | `Direction` | `direction,omitempty` | as authored; the default is applied at projection |

New constants: `ShapeDatabase…ShapeBucket`, `RelSync`, `RelAsync`, `RelTrigger`, `DirDown`,
`DirRight`, with membership helpers used by validation.

## 3. View model (`internal/core/entities/viewmodel`)

| Type | Change |
|---|---|
| `Shape` | adds `ShapeDatabase`, `ShapeQueue`, `ShapeTopic`, `ShapeFunction`, `ShapeBucket` |
| `Node` | adds `Title` (string, `omitempty`); `Label` stays the name |
| `EdgeStyle` | adds `Async` bool (`omitempty`); `Dashed` keeps meaning "crossing" |
| `Edge` | adds `Tags` (`[]string`, `omitempty`) |
| `View` | adds `Direction` (`down` \| `right`), always set by the projection |
| `ElementPage` | adds `Title` (`omitempty`) |
| `StyleFor` | gains an optional element shape; a non-empty shape overrides the kind's default |

## 4. Authoring (`internal/core/entities/authoring`)

| Target | New legal attributes |
|---|---|
| element | `title` (string); `shape` (string) on `container` and `external` only |
| relationship | `kind` (string), `tags` (string list) |
| view | `direction` (string) |

`NewEdit` rejects values outside each set with a field error that lists the allowed values.

## 5. Use-case DTOs (`internal/core/usecases/describe.go`)

| DTO | New fields (all `omitempty`) |
|---|---|
| `ElementView` | `Title` (every level), `Shape` (`structure` and `full`) |
| `RelationshipView` | `Kind`, `Tags` |
| `ViewInfo` | `Direction` |
