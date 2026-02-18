# Data Model & Drift Detection Guide

This guide explains loko's source-of-truth hierarchy and how to use drift detection to keep your architecture documentation consistent.

## Table of Contents

- [Source of Truth Hierarchy](#source-of-truth-hierarchy)
- [What is Drift?](#what-is-drift)
- [Drift Types and Severity](#drift-types-and-severity)
- [Running Drift Detection](#running-drift-detection)
- [Drift Detection Workflow](#drift-detection-workflow)
- [Fixing Drift Issues](#fixing-drift-issues)

---

## Source of Truth Hierarchy

loko uses a **dual source of truth** model for architecture data:

```
┌─────────────────────────────────────────────────┐
│  FRONTMATTER (.md files)         Priority: HIGH  │
│  ─────────────────────────────────────────────   │
│  • Component metadata (name, description, tech)  │
│  • Relationship declarations                     │
│  • Tags and annotations                          │
└──────────────────────┬──────────────────────────┘
                       │ Union Merge
┌──────────────────────▼──────────────────────────┐
│  D2 DIAGRAMS (.d2 files)         Priority: LOW   │
│  ─────────────────────────────────────────────   │
│  • Visual relationship arrows                    │
│  • Diagram-specific labels                       │
│  • Component position and styling                │
└─────────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│  UNIFIED ARCHITECTURE GRAPH                      │
│  ─────────────────────────────────────────────   │
│  • Deduplicated edges                            │
│  • Qualified node IDs                            │
│  • Traversal & query API                         │
└─────────────────────────────────────────────────┘
```

**Frontmatter is authoritative**: When frontmatter and D2 conflict, frontmatter wins.

**D2 extends frontmatter**: D2 arrows add relationships not declared in frontmatter; they don't replace them.

---

## What is Drift?

**Drift** occurs when the frontmatter and D2 sources become inconsistent with each other. Common causes:

- Renaming a component in frontmatter without updating D2 arrows
- Deleting a component while its ID remains in other components' `relationships:` maps
- Updating a component's description in frontmatter without updating the D2 tooltip/label

Drift does not prevent the project from building, but it indicates your documentation is diverging from reality.

---

## Drift Types and Severity

| Drift Type | Severity | Description |
|-----------|----------|-------------|
| `DriftDescriptionMismatch` | **WARNING** | D2 tooltip/label differs from frontmatter `description` |
| `DriftMissingComponent` | **ERROR** | D2 arrow references a component ID not found in frontmatter |
| `DriftOrphanedRelationship` | **ERROR** | Frontmatter `relationships:` map references a deleted component |

### WARNING vs ERROR

- **WARNING**: Cosmetic only — doesn't break graph queries or documentation build
- **ERROR**: Data integrity issue — graph queries may return incorrect results

---

## Running Drift Detection

```bash
# Run drift detection as part of validation
loko validate --check-drift

# Run validation only (no drift check)
loko validate
```

### Sample Output — No Drift

```
✅ Validation passed - No drift detected

Summary:
  Components checked: 17
  Drift issues found: 0
```

### Sample Output — Warnings Only

```
⚠️  Validation passed with warnings

Issues found:
  email-queue (WARNING): D2 tooltip differs from frontmatter description
    Expected: "Standard SQS queue for email notifications"
    Got: "SQS queue"

Summary:
  Components checked: 17
  Drift issues found: 1 (1 warning, 0 errors)
```

### Sample Output — Errors

```
❌ Validation failed - Critical drift detected

Issues found:
  auth-handler (ERROR): Orphaned relationship - target component 'old-service' no longer exists
  payment-processor (ERROR): Orphaned relationship - target component 'legacy-db' no longer exists

Summary:
  Components checked: 17
  Drift issues found: 2 (0 warnings, 2 errors)

Exit code: 1
```

### Exit Codes

| Scenario | Exit Code |
|----------|-----------|
| No drift | `0` |
| Warnings only | `0` |
| Any ERROR-level drift | `1` |

This makes `loko validate --check-drift` safe to use in CI/CD pipelines.

---

## Drift Detection Workflow

Recommended workflow for maintaining consistent architecture documentation:

```
1. Edit frontmatter (rename/delete component)
        │
        ▼
2. Run `loko validate --check-drift`
        │
        ├── No issues → ✅ Done
        │
        └── Issues found
                │
                ▼
        3. Fix issues (see below)
                │
                ▼
        4. Re-run validation → ✅ Done
```

### CI/CD Integration

Add to your CI pipeline:

```yaml
# .github/workflows/architecture.yml
- name: Validate architecture
  run: loko validate --check-drift
```

This ensures drift is caught before it reaches production documentation.

---

## Fixing Drift Issues

### DriftOrphanedRelationship

A component's `relationships:` map references a component that no longer exists.

**Find the issue**:
```
auth-handler (ERROR): Orphaned relationship - target 'old-payment-service' not found
```

**Fix**: Remove the stale entry from the frontmatter:
```yaml
# Before (broken)
relationships:
  "payment-service/api/old-payment-service": "called legacy API"
  "payment-service/api/user-service": "validates tokens"

# After (fixed)
relationships:
  "payment-service/api/user-service": "validates tokens"
```

### DriftMissingComponent

A D2 arrow targets a component that doesn't exist in frontmatter.

**Fix**: Either:
1. Create the missing component: `loko new component --name "Missing Service"`
2. Remove the arrow from the D2 file

### DriftDescriptionMismatch

The D2 tooltip/label doesn't match the frontmatter `description` field.

**Fix**: Update either the frontmatter description or the D2 tooltip to match:

```d2
# Update the D2 label to match frontmatter description
auth-handler: {
  tooltip: "Standard SQS queue for email notifications"
}
```
