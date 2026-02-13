# Research: Serverless Architecture Template

**Feature**: 003-serverless-template
**Date**: 2026-02-05

## R1: Template Engine Architecture

### Decision
The ason template engine resolves templates by **filename only** within registered search paths. Templates are NOT namespaced by template type. The search path determines which template set is used.

### Rationale
The existing `findTemplate()` in `internal/adapters/ason/engine.go` iterates search paths in registration order and returns the first match. This means switching from `standard-3layer` to `serverless` only requires changing the search path from `templates/standard-3layer/` to `templates/serverless/`. No engine modifications needed.

### Alternatives Considered
- **Namespaced templates** (e.g., `serverless/system.md`): Would require engine changes to support subdirectory resolution. Rejected because the current simple path-based approach works.
- **Template registry pattern**: Over-engineered for 2 templates. YAGNI.

## R2: Current Markdown Generation Gap

### Decision
Markdown generation must be migrated from hardcoded Go methods to ason template rendering for the serverless template to work.

### Rationale
Currently, `project_repo.go` has hardcoded methods:
- `generateSystemMarkdown()` (line 535) - builds markdown via `strings.Builder`
- `generateContainerMarkdown()` (line 628) - same pattern
- `generateComponentMarkdown()` (line 722) - same pattern

These produce `standard-3layer`-style content regardless of which template is selected. Only `component.d2` uses the template engine (line 307). The pattern for template-first-with-fallback already exists in `SaveComponent` (lines 299-317).

### Implementation Approach
Apply the same try-template-then-fallback pattern used for `component.d2` to all 6 file types:
1. Try `templateEngine.RenderTemplate(ctx, "system.md", variables)` first
2. Fall back to existing hardcoded `generateSystemMarkdown()` if template not found
3. This preserves backward compatibility while enabling template-driven content

### Alternatives Considered
- **Keep hardcoded markdown, only template D2**: Would mean serverless template `.md` files are ignored. Rejected because it defeats the purpose.
- **Remove hardcoded fallbacks entirely**: Risky - breaks if template directory is missing. Keep fallbacks for robustness.

## R3: D2 Generation Gap

### Decision
D2 diagram generation in `cmd/new.go` (hardcoded methods `createContainerD2Template`, `createComponentD2Template`) and `cmd/d2_generator.go` (system diagrams) must also be migrated to template-first rendering.

### Rationale
Currently there are 3 separate D2 generation paths:
1. `cmd/d2_generator.go`: `GenerateSystemContextDiagram()`, `UpdateSystemD2File()`, `UpdateContainerD2File()` - used by `cmd/new.go` for system D2
2. `cmd/new.go`: `createContainerD2Template()`, `createComponentD2Template()` - inline D2 generation
3. `project_repo.go`: Template engine for `component.d2` only

For the serverless template, the D2 templates need to produce event-driven patterns (dashed lines, cloud icons) which the hardcoded generators cannot produce.

### Implementation Approach
In `cmd/new.go`, before falling back to the hardcoded D2 generators, try to render from the template engine. The D2Generator in `cmd/d2_generator.go` is only used for *updating* existing diagrams when new containers/components are added, and can remain as-is for standard-3layer. Serverless D2 files will be generated from templates initially.

## R4: Template Selection Mechanism

### Decision
Add a `-template` flag to `cmd/new.go` (and corresponding field on `NewCommand`). Default value: `standard-3layer`. The build command should read the template from `loko.toml` config or environment variable.

### Rationale
- Per-entity template selection allows mixed architectures in one project
- Default preserves backward compatibility
- `loko.toml` can specify a project-wide default template
- Environment variable `LOKO_TEMPLATE_DIR` already exists in `build.go` as precedent

### Alternatives Considered
- **Project-level only** (in `loko.toml`): Too rigid - projects may have both serverless and traditional systems
- **Interactive prompt**: Adds friction for automation/scripting
- **Auto-detection**: No reliable way to detect architecture style

## R5: D2 Dashed Lines for Async Flows

### Decision
D2 supports dashed stroke patterns via `style.stroke-dash: 5` attribute. This will be used in serverless D2 templates for async/event-driven flows.

### Rationale
D2 documentation confirms stroke-dash support on connections:
```d2
a -> b: "async event" {
  style.stroke-dash: 5
}
```

### Verification
Confirmed in D2 language specification. No external icon URL issues - D2 supports `icon` property for external images (already used in standard-3layer templates, e.g., `icon: "https://icons.terrastruct.com/..."`).

## R6: Template Variable Consistency

### Decision
Reuse the same ason variable names from `standard-3layer` templates. Serverless-specific content goes in the template body text, not in new variables.

### Rationale
The existing variables (`SystemName`, `SystemID`, `Description`, `Technology`, `ContainerName`, `ContainerID`, `ComponentName`, `ComponentID`) are generic enough for any architecture style. Serverless-specific sections (Event Sources, Triggers, IAM Permissions) are static template text, not variable-substituted.

### Variables Used
| Variable | Used In | Source |
|----------|---------|--------|
| `SystemName` | system.md, system.d2 | `entities.System.Name` |
| `SystemID` | system.d2 | `entities.System.ID` |
| `Description` | all files | entity `.Description` field |
| `Technology` | container/component files | entity `.Technology` field |
| `ContainerName` | container.md, container.d2 | `entities.Container.Name` |
| `ContainerID` | container.d2 | `entities.Container.ID` |
| `ComponentName` | component.md, component.d2 | `entities.Component.Name` |
| `ComponentID` | component.d2 | `entities.Component.ID` |
| `Language` | system.md | from `CreateSystemRequest` |
| `Framework` | system.md, component.md | from request or entity |
| `Database` | system.md, container.md | from request or entity |
