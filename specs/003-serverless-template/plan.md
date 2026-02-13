# Implementation Plan: Serverless Architecture Template

**Branch**: `003-serverless-template` | **Date**: 2026-02-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-serverless-template/spec.md`

## Summary

Add a `serverless` starter template to loko alongside the existing `standard-3layer` template. This requires three things: (1) creating 6 template files with serverless-specific content, (2) adding template selection support to the CLI so users can choose which template to use, and (3) wiring the ason template engine into markdown generation (currently only used for component D2 diagrams).

The key technical insight from research is that markdown generation in `project_repo.go` is currently hardcoded in Go methods (`generateSystemMarkdown`, `generateContainerMarkdown`, `generateComponentMarkdown`), bypassing the template files that exist in `templates/standard-3layer/`. These methods must be updated to use the template engine with fallback to hardcoded generation for backward compatibility.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: ason template engine (internal), fsnotify, d2 CLI (external)
**Storage**: Filesystem (template files, project files)
**Testing**: `go test`, integration tests in `tests/integration/`
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: Single Go binary with embedded templates
**Performance Goals**: N/A (template rendering is trivial)
**Constraints**: Backward compatible - existing projects and workflows must not break
**Scale/Scope**: 6 new template files, ~4 modified Go files, 1 new example project

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution file (`.specify/memory/constitution.md`) is a blank template with no project-specific rules defined. Architectural constraints from `CLAUDE.md` apply:

| Gate | Status | Notes |
|------|--------|-------|
| Clean Architecture (core/ imports nothing from adapters/) | PASS | Template files are static assets, no core changes needed |
| Interface-First (use ports.go interfaces) | PASS | Uses existing `TemplateEngine` interface from ports.go |
| Thin Handlers (<50 lines) | PASS | CLI changes add ~5 lines per command (flag parsing) |
| Entity Validation (in entities, not use cases) | PASS | No new entities introduced |
| No over-engineering | PASS | Template-first-with-fallback pattern already exists in codebase |

**Post-Phase 1 Re-check**: All gates still pass. No new patterns introduced - extending existing template engine usage to markdown files.

## Project Structure

### Documentation (this feature)

```text
specs/003-serverless-template/
├── plan.md              # This file
├── research.md          # Phase 0 output - template engine findings
├── data-model.md        # Phase 1 output - entity/file inventory
├── quickstart.md        # Phase 1 output - usage walkthrough
├── contracts/           # Phase 1 output
│   └── cli-contract.md  # CLI flag and config changes
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
templates/
├── standard-3layer/          # Existing (unchanged)
│   ├── system.md
│   ├── system.d2
│   ├── container.md
│   ├── container.d2
│   ├── component.md
│   └── component.d2
└── serverless/               # NEW - 6 template files
    ├── system.md             # Event Sources, Functions, External Integrations
    ├── system.d2             # API Gateway + Lambda context diagram
    ├── container.md          # Trigger Type, Functions List, IAM Permissions
    ├── container.d2          # Event flow with dashed async lines
    ├── component.md          # Handler, Trigger, Runtime, Memory, Timeout
    └── component.d2          # Function with trigger source and targets

cmd/
├── new.go                    # MODIFIED - add -template flag, template-first rendering
└── build.go                  # MODIFIED - read template from config/flag

internal/adapters/
├── filesystem/
│   └── project_repo.go      # MODIFIED - use template engine for all .md generation
└── config/
    └── loader.go             # MODIFIED - add Template field to config

examples/
└── serverless/               # NEW - complete example project
    ├── loko.toml
    └── src/
        └── order-processing/
            ├── system.md
            ├── system.d2
            ├── api-handlers/
            │   ├── container.md
            │   └── container.d2
            └── event-processors/
                ├── container.md
                └── container.d2

main.go                       # MODIFIED - pass -template flag to NewCommand
docs/quickstart.md            # MODIFIED - document -template flag
```

**Structure Decision**: This feature adds files alongside the existing structure. No new directories under `internal/` are needed. Template files are static assets at the repository root. The only Go code changes are in `cmd/` (CLI flag) and `internal/adapters/` (template-first rendering).

## Implementation Approach

### Phase A: Create Serverless Template Files (US-1, US-3)

Create all 6 template files in `templates/serverless/` with serverless-specific content:
- `.md` files use ason `{{Variable}}` syntax matching the existing variable set
- `.d2` files use dashed lines (`style.stroke-dash: 5`) for async flows
- D2 files use cloud-native icons from terrastruct icon library
- Content uses serverless terminology throughout (Lambda, triggers, event sources, IAM permissions)

### Phase B: Template Selection Mechanism (US-2)

1. Add `templateName string` field to `NewCommand` in `cmd/new.go`
2. Add `-template` flag in `main.go`'s `handleNew()` function
3. Replace hardcoded `"standard-3layer"` in search path construction with the flag value
4. Add `LOKO_TEMPLATE_DIR` env var support to `cmd/new.go` (already exists in `cmd/build.go`)
5. Optionally add `template` field to `loko.toml` config

### Phase C: Template-First Markdown Generation (US-1)

Update `internal/adapters/filesystem/project_repo.go`:
1. In `SaveSystem()` (line 144): Try `templateEngine.RenderTemplate(ctx, "system.md", vars)` before calling `generateSystemMarkdown()`
2. In `SaveContainer()` (line 183): Same pattern for `container.md`
3. In `SaveComponent()` (line 289): Already partially done for `component.d2`, extend to `component.md`

Pattern (already exists at lines 299-317 for component.d2):
```
if pr.templateEngine != nil {
    rendered, err := pr.templateEngine.RenderTemplate(ctx, "system.md", variables)
    if err == nil {
        content = rendered
    } else {
        content = pr.generateSystemMarkdown(system)  // fallback
    }
} else {
    content = pr.generateSystemMarkdown(system)  // no engine
}
```

### Phase D: Template-First D2 Generation (US-3)

Update `cmd/new.go`:
1. In `createSystem()`: Try rendering `system.d2` from template before using `D2Generator`
2. In `createContainer()`: Try rendering `container.d2` from template before using hardcoded `createContainerD2Template()`
3. In `createComponent()`: Try rendering `component.d2` from template before using hardcoded `createComponentD2Template()`

### Phase E: Example Project (US-4)

Create `examples/serverless/` with a complete, hand-crafted example project demonstrating:
- System: Order Processing API
- Containers: API Handlers, Event Processors
- Realistic serverless architecture with SQS, DynamoDB, API Gateway

### Phase F: Documentation & Validation

- Update `docs/quickstart.md` with `-template` flag documentation
- Update `README.md` to reflect both templates are now available
- Run end-to-end validation: scaffold with serverless template, validate, build

## Key Risks

| Risk | Mitigation |
|------|------------|
| Template rendering breaks existing projects | Fallback to hardcoded generation if template not found |
| D2 dashed lines not rendering correctly | Verify with D2 CLI during integration testing |
| Variable mismatch between templates and code | Reuse exact same variable names as standard-3layer |
| `loko.toml` config change breaks existing configs | New field is optional with default value |

## Complexity Tracking

No constitution violations. This feature follows established patterns:
- Template-first-with-fallback (exists in `project_repo.go` for `component.d2`)
- CLI flag parsing (exists for all other flags in `main.go`)
- Template file structure (mirrors `standard-3layer` exactly)
