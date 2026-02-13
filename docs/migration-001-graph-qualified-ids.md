# Migration Guide: Qualified Node IDs

**Feature**: Architecture Graph Improvements (001-graph-implementation)  
**Impact**: Breaking change for multi-system projects  
**Date**: 2025-02-13  
**Affected Version**: v0.2.0+

## Overview

This migration guide helps you update existing loko projects from short node IDs to qualified hierarchical IDs. This change prevents node ID collisions in multi-system projects while maintaining backward compatibility through ShortIDMap.

## What Changed

### Before (Short IDs)

```yaml
# Project with collision risk
systems:
  - id: backend
    containers:
      - id: api
        components:
          - id: auth  # Collision!
  - id: admin
    containers:
      - id: ui
        components:
          - id: auth  # Same ID - overwrites backend/auth
```

**Problem**: Both `auth` components used ID `"auth"`, causing silent data loss when stored in graph.

### After (Qualified IDs)

```yaml
# Same project, no collisions
systems:
  - id: backend
    containers:
      - id: api
        components:
          - id: auth  # Stored as "backend/api/auth"
  - id: admin
    containers:
      - id: ui
        components:
          - id: auth  # Stored as "admin/ui/auth" - unique!
```

**Solution**: Qualified IDs encode full hierarchy: `systemID/containerID/componentID`

## Node ID Format

| Entity Type | Old Format | New Format | Example |
|-------------|------------|------------|---------|
| System | `systemID` | `systemID` (unchanged) | `"backend"` |
| Container | `containerID` | `systemID/containerID` | `"backend/api"` |
| Component | `componentID` | `systemID/containerID/componentID` | `"backend/api/auth"` |

## Migration Steps

### Step 1: Assess Impact

**Single-system projects**: No migration needed if you only have one system.

**Multi-system projects**: Check for duplicate component/container names across systems:

```bash
# Find potential collisions in your loko.toml
grep -A 10 "id = " src/*/loko.toml | grep "id = " | sort | uniq -d
```

If you see duplicates, proceed with migration.

### Step 2: Update Component Relationships

**Before**:
```yaml
components:
  - id: auth
    relationships:
      database: "Stores credentials"  # Short ID reference
```

**After** (Option 1 - Recommended):
```yaml
components:
  - id: auth
    relationships:
      database: "Stores credentials"  # Short ID still works via ResolveID()
```

**After** (Option 2 - Explicit):
```yaml
components:
  - id: auth
    relationships:
      backend/api/database: "Stores credentials"  # Qualified ID (prevents ambiguity)
```

**Best Practice**: Use short IDs for within-container relationships, qualified IDs for cross-system relationships.

### Step 3: Update Custom MCP Tool Usage

If you have custom scripts or tools calling MCP graph tools:

**Before**:
```javascript
// Query dependencies using short ID
await mcp.call("query_dependencies", {
  project_root: ".",
  component_id: "auth"  // Ambiguous if multiple systems have "auth"
});
```

**After**:
```javascript
// Query dependencies using qualified ID
await mcp.call("query_dependencies", {
  project_root: ".",
  system_id: "backend",
  container_id: "api",
  component_id: "auth"  // Unambiguous
});
```

### Step 4: Verify Graph Construction

Run your project through the graph builder to ensure no errors:

```bash
# Build the architecture graph
go run . build-docs --project-root .

# Check for validation warnings
go run . validate --project-root .
```

Look for warnings like:
- "Ambiguous short ID: 'auth' matches multiple components"
- "Dangling reference: 'database' could not be resolved"

### Step 5: Update Documentation

If your project documentation references component IDs:

```markdown
<!-- Before -->
The `auth` component handles authentication.

<!-- After -->
The `backend/api/auth` component handles authentication.
```

## Backward Compatibility

### ShortIDMap Resolution

The graph automatically maintains a `ShortIDMap` that maps short IDs to qualified IDs:

```go
// Short ID lookup (works if unambiguous)
qualifiedID, ok := graph.ResolveID("auth")
if ok {
    // "auth" uniquely identifies one component
    deps := graph.GetDependencies(qualifiedID)
} else {
    // "auth" is ambiguous - multiple matches
    // Must use qualified ID
}
```

### Relationship Resolution

Component relationships use smart resolution:

1. **Try as qualified ID**: Check if `"backend/api/database"` exists
2. **Try as short ID**: Use `ResolveID("database")` for backward compatibility
3. **Fail with clear error**: "Ambiguous ID 'database' matches: backend/api/database, admin/ui/database"

## Examples

### Example 1: E-Commerce with Multiple Systems

**Scenario**: You have `backend`, `admin`, and `mobile` systems, each with an `auth` component.

**Before Migration** (BROKEN):
```
Graph nodes:
  auth -> Last one wins, others silently overwritten
```

**After Migration** (FIXED):
```
Graph nodes:
  backend/api/auth
  admin/ui/auth
  mobile/app/auth
```

**Relationships**:
```yaml
# backend/api/auth component
relationships:
  database: "Stores user credentials"  # Resolves to backend/api/database

# admin/ui/auth component  
relationships:
  backend/api/auth: "Delegates to backend"  # Qualified ID for cross-system
```

### Example 2: Microservices Architecture

**Scenario**: 5 services, each with `database`, `cache`, `api` components.

**Migration**:
```yaml
# payment-service/src/loko.toml
systems:
  - id: payment
    containers:
      - id: service
        components:
          - id: database  # Becomes payment/service/database
          - id: cache     # Becomes payment/service/cache
          - id: api       # Becomes payment/service/api
            relationships:
              database: "Stores transactions"  # Short ID OK (same container)
              cache: "Caches balances"        # Short ID OK (same container)
```

No code changes needed - short IDs work within same container!

## Troubleshooting

### Issue: "Ambiguous short ID"

**Error**: `Ambiguous short ID 'auth' matches: backend/api/auth, admin/ui/auth`

**Solution**: Use qualified ID in relationship:
```yaml
relationships:
  backend/api/auth: "Description"  # Explicit system/container/component
```

### Issue: "Dangling reference"

**Error**: `Dangling reference: 'nonexistent' not found in graph`

**Solution**: Check for typos in component IDs or ensure the referenced component exists.

### Issue: MCP tool returns wrong component

**Symptom**: Querying `auth` returns the wrong system's auth component.

**Solution**: Update MCP tool call to use qualified parameters:
```javascript
{
  system_id: "backend",
  container_id: "api",
  component_id: "auth"
}
```

## Testing

Run these checks after migration:

```bash
# 1. Build succeeds
go build ./...

# 2. All tests pass
go test ./...

# 3. Graph validates without errors
go run . validate --project-root .

# 4. MCP tools work correctly
# Test each graph tool with your project

# 5. Documentation builds
go run . build-docs --project-root . --output-dir ./dist
```

## Rollback

If you encounter issues and need to rollback:

1. The feature is backward compatible - single-system projects work unchanged
2. Multi-system projects may need to temporarily merge systems into one
3. Report issues at: https://github.com/madstone-tech/loko/issues

## Support

For questions or issues:
- GitHub Issues: https://github.com/madstone-tech/loko/issues
- Documentation: See ADR-0004 in `docs/adr/0004-graph-conventions.md`
- Examples: Check `tests/integration/graph_collision_test.go`

## References

- ADR-0004: Architecture Graph Conventions
- Spec: `specs/001-graph-implementation/spec.md`
- Implementation: `internal/core/entities/graph.go`
- Tests: `tests/integration/graph_collision_test.go`
