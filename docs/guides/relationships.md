# Relationships Guide

This guide explains how to define and query architectural relationships in loko using both frontmatter syntax and D2 diagram arrows.

## Table of Contents

- [Overview](#overview)
- [Frontmatter Syntax](#frontmatter-syntax)
- [D2 Arrow Syntax](#d2-arrow-syntax)
- [Union Merge](#union-merge)
- [Querying Relationships](#querying-relationships)
- [Troubleshooting](#troubleshooting)

---

## Overview

loko builds a unified relationship graph from **two sources of truth** that are merged automatically:

1. **Frontmatter relationships** — declared in `component.md` YAML front matter
2. **D2 diagram arrows** — declared in `component.d2` diagram files

Both sources are merged at graph-build time. Duplicate edges (same source → target) are deduplicated automatically.

---

## Frontmatter Syntax

Define relationships directly in a component's markdown file under the `relationships` key.

**File**: `src/payment-service/api-gateway/auth-handler.md`

```yaml
---
name: "Auth Handler"
description: "Validates JWT tokens and enforces RBAC"
technology: "Go"
relationships:
  "payment-service/api-gateway/user-service": "validates tokens via"
  "payment-service/api-gateway/audit-logger": "logs auth events to"
---
```

### Relationship Target Format

Targets use the **qualified component ID**: `system/container/component`

```
<system-name>/<container-name>/<component-name>
```

All names are lowercased and hyphenated automatically:
- System "Payment Service" → `payment-service`
- Container "API Gateway" → `api-gateway`
- Component "Auth Handler" → `auth-handler`

### Relationship Value

The value is a free-text description of the relationship:

```yaml
relationships:
  "payment-service/api/database": "reads/writes user records"
  "external/stripe/billing-api": "charges payments via REST"
```

---

## D2 Arrow Syntax

Define relationships visually in a component's `.d2` diagram file.

**File**: `src/payment-service/api-gateway/auth-handler.d2`

```d2
auth-handler -> user-service: validates tokens
auth-handler -> audit-logger: logs events
auth-handler -> rate-limiter: checks limits
```

### Arrow Format

```
source -> target: optional label
```

- `source` — the source component ID (short or qualified)
- `target` — the target component ID (short or qualified)
- `label` — optional relationship description

### Using Short IDs

Within the same container, you can use short component IDs:

```d2
# These are equivalent if components are in the same container:
auth-handler -> user-service: calls
payment-service/api-gateway/auth-handler -> payment-service/api-gateway/user-service: calls
```

---

## Union Merge

When loko builds the architecture graph, it reads **both** frontmatter and D2 sources and merges them:

```
frontmatter edges ∪ D2 edges = unified graph (deduplicated)
```

**Deduplication key**: `sourceQualifiedID -> targetQualifiedID`

If the same edge appears in both sources:
- The edge is included once
- The frontmatter description takes precedence (frontmatter is parsed first)

### Worker Pool

D2 files are parsed concurrently using a 10-worker pool. For 100 components, parsing completes in approximately 450ms.

---

## Querying Relationships

Use MCP tools or the CLI to query the relationship graph:

### MCP Tools

```
find_relationships          — Search edges by source/target pattern
query_dependencies          — Find what a component depends on
query_related_components    — Find components related to a given component
analyze_coupling            — Measure coupling metrics across the architecture
```

**Example** (via Claude):
```
"Show me all relationships from auth-handler"
"What does payment-processor depend on?"
"Which components are tightly coupled?"
```

### CLI

```bash
# Build architecture graph and view relationships
loko validate --check-drift

# Query via MCP
loko mcp
```

---

## Troubleshooting

### "No relationships found" from MCP tools

**Cause**: The architecture graph was built without D2 parsing enabled, or no relationships are defined.

**Fix**: Ensure your components have either:
1. A `relationships:` map in frontmatter, OR
2. Arrow syntax (`->`) in their `.d2` files

Verify with:
```bash
go test -run TestBuildArchitectureGraph -v ./internal/core/usecases/
```

### D2 parse warnings

You may see `WARN missing slog.Logger in context` — this is harmless noise from the upstream D2 library. Suppress it by redirecting stderr if needed.

### Qualified ID not resolving

If a relationship target can't be resolved, check:
1. The target uses the correct format: `system/container/component`
2. All names are lowercase and hyphenated
3. The target component exists in `src/`

### Duplicate edges

If you define the same relationship in both frontmatter and D2, loko deduplicates automatically. You won't see duplicates in query results.
