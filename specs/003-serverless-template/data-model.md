# Data Model: Serverless Architecture Template

**Feature**: 003-serverless-template
**Date**: 2026-02-05

## Entities

This feature does not introduce new domain entities. It adds template files and modifies the template selection flow. The existing C4 entities (System, Container, Component) are reused with serverless-specific content in the templates.

### Existing Entities (Unchanged)

| Entity | Location | Key Fields |
|--------|----------|------------|
| `System` | `internal/core/entities/system.go` | Name, ID, Description, KeyUsers, Responsibilities, Dependencies, ExternalSystems, Technology |
| `Container` | `internal/core/entities/container.go` | Name, ID, Description, Technology, Path |
| `Component` | `internal/core/entities/component.go` | Name, ID, Description, Technology, Path |

### Template Files (New)

| File | Location | Purpose |
|------|----------|---------|
| `system.md` | `templates/serverless/system.md` | Serverless system documentation with Event Sources, Functions, External Integrations sections |
| `system.d2` | `templates/serverless/system.d2` | System context diagram with API Gateway, Lambda, event source shapes |
| `container.md` | `templates/serverless/container.md` | Function group documentation with Trigger Type, Functions List, IAM Permissions sections |
| `container.d2` | `templates/serverless/container.d2` | Container diagram with event flow patterns using dashed lines |
| `component.md` | `templates/serverless/component.md` | Individual Lambda function documentation with Handler, Trigger, Runtime, Memory, Timeout |
| `component.d2` | `templates/serverless/component.d2` | Component diagram with trigger source and downstream targets |

### Configuration (Modified)

| Entity | Location | Change |
|--------|----------|--------|
| `NewCommand` | `cmd/new.go` | Add `templateName string` field |
| `BuildCommand` | `cmd/build.go` | Read template name from config or flag |
| `ProjectConfig` | `internal/adapters/config/loader.go` | Add `Template string` field to config (optional) |

## Relationships

```
templates/
├── standard-3layer/    (existing, 6 files)
│   ├── system.md
│   ├── system.d2
│   ├── container.md
│   ├── container.d2
│   ├── component.md
│   └── component.d2
└── serverless/         (new, 6 files - same names, different content)
    ├── system.md
    ├── system.d2
    ├── container.md
    ├── container.d2
    ├── component.md
    └── component.d2

Selection Flow:
  CLI flag (-template) → NewCommand.templateName → search path construction → ason engine
  loko.toml (template) → BuildCommand → search path construction → ason engine
  LOKO_TEMPLATE_DIR env var → search path override (highest priority)
```

## Validation Rules

- Template name must be a valid directory name under `templates/`
- Template directory must contain all 6 required files
- Template files must produce valid YAML frontmatter in `.md` files
- Template D2 files must produce valid D2 syntax (verified by `loko build`)
- Template variables must match the variables provided by the rendering code

## State Transitions

N/A - Template files are static assets. No runtime state transitions.
