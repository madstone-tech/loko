# CLI Contract: Template Selection

**Feature**: 003-serverless-template
**Date**: 2026-02-05

## Modified Commands

### `loko new system <name> [-template <template-name>]`

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `-template` | string | `standard-3layer` | Template to use for scaffolding |

**Behavior**:
- Loads template files from `templates/<template-name>/`
- If template directory not found: error with list of available templates
- Template affects both `.md` and `.d2` content generation

### `loko new container <name> -parent <system> [-template <template-name>]`

Same `-template` flag as above.

### `loko new component <name> -parent <container> [-template <template-name>]`

Same `-template` flag as above.

## Environment Variables

| Variable | Purpose | Priority |
|----------|---------|----------|
| `LOKO_TEMPLATE_DIR` | Override template search path entirely | Highest (overrides flag) |

## Configuration (`loko.toml`)

```toml
# Optional: set project-wide default template
[project]
template = "serverless"  # or "standard-3layer" (default)
```

## Error Messages

| Condition | Message |
|-----------|---------|
| Template directory not found | `Error: template "X" not found. Available templates: standard-3layer, serverless` |
| Template file missing from directory | `Warning: template file "system.md" not found, using default` |

## Backward Compatibility

- All commands work identically without `-template` flag (default: `standard-3layer`)
- Existing projects are unaffected
- `LOKO_TEMPLATE_DIR` continues to work as before
