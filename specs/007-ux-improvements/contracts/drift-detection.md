# Contract: Drift Detection Output

**Feature**: 007-ux-improvements  
**Use Case**: `DetectDrift`  
**CLI Command**: `loko validate --check-drift`  
**Location**: `internal/core/usecases/detect_drift.go`

---

## Output Schema

### DriftIssue Structure

```go
type DriftIssue struct {
    ComponentID string        // Component where drift detected
    Type        DriftType     // Category of drift
    Severity    DriftSeverity // Warning or Error
    Message     string        // Human-readable description
    Context     string        // Additional context (expected vs actual)
}
```

### DriftType Enum

```go
type DriftType int

const (
    DriftDescriptionMismatch DriftType = iota  // D2 tooltip != frontmatter description
    DriftMissingComponent                      // D2 references non-existent component
    DriftOrphanedRelationship                  // Frontmatter relationship to deleted component
)
```

### DriftSeverity Enum

```go
type DriftSeverity int

const (
    DriftWarning DriftSeverity = iota // Cosmetic inconsistencies
    DriftError                         // Broken references, data integrity issues
)
```

---

## Severity Rules

| Drift Type | Severity | Rationale |
|-----------|----------|-----------|
| `DriftDescriptionMismatch` | `WARNING` | Cosmetic only, doesn't break functionality |
| `DriftMissingComponent` | `ERROR` | Broken reference, graph queries will fail |
| `DriftOrphanedRelationship` | `ERROR` | Data integrity issue, invalid relationship |

---

## CLI Output Format

### Success (No Drift)

```
✅ Validation passed - No drift detected

Summary:
  Components checked: 17
  Drift issues found: 0
```

### Warning-Level Drift

```
⚠️  Validation passed with warnings

Issues found:
  email-queue (WARNING): D2 tooltip differs from frontmatter description
    D2: "SQS queue"
    Frontmatter: "Standard SQS queue for email notifications"
    
  sms-sender (WARNING): D2 tooltip differs from frontmatter description
    D2: "Lambda function"
    Frontmatter: "Lambda function for sending SMS notifications via SNS"

Summary:
  Components checked: 17
  Warnings: 2
  Errors: 0

Exit code: 0
```

### Error-Level Drift

```
❌ Validation failed - Drift errors detected

Issues found:
  email-queue (ERROR): D2 references non-existent component 'old-sender'
    Arrow: email-queue -> old-sender: "triggers"
    
  notification-router (ERROR): Frontmatter relationship to deleted component
    Relationship: uses -> notification-service/processing-layer/deleted-component
    
  sms-queue (WARNING): D2 tooltip differs from frontmatter description
    D2: "Queue"
    Frontmatter: "Standard SQS queue for SMS notifications"

Summary:
  Components checked: 17
  Warnings: 1
  Errors: 2

Exit code: 1
```

---

## JSON Output Format

**Flag**: `loko validate --check-drift --format json`

```json
{
  "status": "failed",
  "components_checked": 17,
  "drift_issues": [
    {
      "component_id": "email-queue",
      "type": "missing_component",
      "severity": "error",
      "message": "D2 references non-existent component 'old-sender'",
      "context": "Arrow: email-queue -> old-sender: \"triggers\""
    },
    {
      "component_id": "notification-router",
      "type": "orphaned_relationship",
      "severity": "error",
      "message": "Frontmatter relationship to deleted component",
      "context": "Relationship: uses -> notification-service/processing-layer/deleted-component"
    },
    {
      "component_id": "sms-queue",
      "type": "description_mismatch",
      "severity": "warning",
      "message": "D2 tooltip differs from frontmatter description",
      "context": "D2: \"Queue\" | Frontmatter: \"Standard SQS queue for SMS notifications\""
    }
  ],
  "summary": {
    "warnings": 1,
    "errors": 2
  }
}
```

---

## Detection Logic

### Description Mismatch Detection

```go
// Compare D2 tooltip to frontmatter description
d2Tooltip := extractTooltipFromD2(d2Source)
frontmatterDesc := component.Description

if d2Tooltip != "" && d2Tooltip != frontmatterDesc {
    issue := NewDriftIssue(
        component.ID,
        DriftDescriptionMismatch,
        "D2 tooltip differs from frontmatter description",
        fmt.Sprintf("D2: %q | Frontmatter: %q", d2Tooltip, frontmatterDesc),
    )
    issues = append(issues, issue)
}
```

### Missing Component Detection

```go
// Parse D2 relationships and check if targets exist
d2Relationships := parseD2Relationships(d2Source)
for _, rel := range d2Relationships {
    if !componentExists(rel.Target) {
        issue := NewDriftIssue(
            component.ID,
            DriftMissingComponent,
            fmt.Sprintf("D2 references non-existent component '%s'", rel.Target),
            fmt.Sprintf("Arrow: %s -> %s: %q", rel.Source, rel.Target, rel.Label),
        )
        issues = append(issues, issue)
    }
}
```

### Orphaned Relationship Detection

```go
// Check frontmatter relationships for deleted components
for relType, targets := range component.Relationships {
    for _, target := range targets {
        if !componentExists(target) {
            issue := NewDriftIssue(
                component.ID,
                DriftOrphanedRelationship,
                "Frontmatter relationship to deleted component",
                fmt.Sprintf("Relationship: %s -> %s", relType, target),
            )
            issues = append(issues, issue)
        }
    }
}
```

---

## Test Cases

### Unit Tests

1. **No drift**: Component with matching D2 tooltip and frontmatter → no issues
2. **Description mismatch**: D2 tooltip != frontmatter → 1 WARNING issue
3. **Missing component in D2**: Arrow to non-existent component → 1 ERROR issue
4. **Orphaned relationship**: Frontmatter relationship to deleted component → 1 ERROR issue
5. **Multiple drift types**: Component with all 3 drift types → 3 issues (1 WARNING, 2 ERROR)
6. **Empty D2 tooltip**: Empty tooltip in D2 → no description mismatch issue

### Integration Tests

1. **Real-world project with 17 components**: Detect all drift issues
2. **Performance**: 100 components drift detection <200ms
3. **CLI output formatting**: Verify human-readable output
4. **JSON output formatting**: Verify valid JSON schema
5. **Exit code behavior**: Exit 0 for warnings, exit 1 for errors

---

## Performance Requirements

| Scenario | Requirement | Measurement |
|----------|------------|-------------|
| Single component drift check | <2ms | Benchmark test |
| 100 components drift check | <200ms | Integration test |
| CLI output formatting | <10ms | Unit test |
| JSON serialization | <5ms | Unit test |

---

## Examples

### Example 1: No Drift

**Input**:
```yaml
# component.md frontmatter
description: "Standard SQS queue for email notifications"
```

```d2
# component.d2
email-queue: {
  tooltip: "Standard SQS queue for email notifications"
}
```

**Output**:
```go
issues := DetectDrift(component, d2Source, allComponents)
len(issues) == 0
```

### Example 2: Description Mismatch (WARNING)

**Input**:
```yaml
# component.md frontmatter
description: "Standard SQS queue for email notifications"
```

```d2
# component.d2
email-queue: {
  tooltip: "SQS queue"
}
```

**Output**:
```go
issues := DetectDrift(component, d2Source, allComponents)
len(issues) == 1
issues[0].Type == DriftDescriptionMismatch
issues[0].Severity == DriftWarning
```

### Example 3: Missing Component (ERROR)

**Input**:
```d2
# component.d2
email-queue -> deleted-component: "triggers"
```

```go
// deleted-component does not exist in allComponents
```

**Output**:
```go
issues := DetectDrift(component, d2Source, allComponents)
len(issues) == 1
issues[0].Type == DriftMissingComponent
issues[0].Severity == DriftError
issues[0].Message contains "deleted-component"
```

### Example 4: Multiple Drift Types

**Input**:
```yaml
# component.md frontmatter
description: "Email queue"
relationships:
  uses:
    - notification-service/processing-layer/deleted-sender
```

```d2
# component.d2
email-queue: {
  tooltip: "SQS queue"  # Description mismatch
}

email-queue -> missing-component: "triggers"  # Missing component
```

**Output**:
```go
issues := DetectDrift(component, d2Source, allComponents)
len(issues) == 3

// Description mismatch
issues[0].Type == DriftDescriptionMismatch
issues[0].Severity == DriftWarning

// Missing component in D2
issues[1].Type == DriftMissingComponent
issues[1].Severity == DriftError

// Orphaned relationship in frontmatter
issues[2].Type == DriftOrphanedRelationship
issues[2].Severity == DriftError
```

---

## Validation Checklist

- [ ] WARNING severity for description mismatches
- [ ] ERROR severity for missing components
- [ ] ERROR severity for orphaned relationships
- [ ] Context field provides actionable debugging info
- [ ] CLI output is human-readable
- [ ] JSON output is valid JSON schema
- [ ] Exit code 0 for warnings, 1 for errors
- [ ] Performance: <200ms for 100 components
- [ ] Empty D2 tooltips don't trigger false positives

---

**Status**: ✅ Contract complete - Ready for implementation
