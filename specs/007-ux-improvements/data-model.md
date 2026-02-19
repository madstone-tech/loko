# Data Model: UX Improvements

**Feature**: 007-ux-improvements  
**Date**: 2026-02-13  
**Status**: Design Phase

---

## Overview

This document defines the data model changes for UX improvements, including relationship parsing, template selection, and drift detection.

**Key Principles**:
- Frontmatter is the authoritative source for component metadata
- D2 files are the authoritative source for relationships and visual layout
- Union merge combines relationships from both sources (deduplicate by source+target+type)
- Markdown body is free-form documentation (not parsed or validated)

---

## Entity Changes

### 1. Component (Enhanced)

**File**: `internal/core/entities/component.go` (line 22 - existing Relationships map)

**Current State**:
```go
type Component struct {
    ID           string
    Name         string
    Description  string
    Technology   string
    Tags         []string
    Relationships map[string][]string // Existing but unused
    Path         string
}
```

**Enhancement**: Document frontmatter relationship format

**Frontmatter Schema**:
```yaml
---
id: email-queue
name: "Email Queue"
description: "Standard SQS queue for email notifications"
technology: "Amazon SQS Standard"
tags:
  - messaging
relationships:
  uses:
    - notification-service/processing-layer/email-sender
  triggered_by:
    - notification-service/api-layer/notification-api
  publishes_to:
    - notification-service/data-store/delivery-status-table
---
```

**Validation Rules**:
- `id`: Required, alphanumeric + hyphens, unique within container
- `name`: Required, human-readable string
- `description`: Required, 1-500 characters
- `technology`: Required, free-form string (used for template selection)
- `tags`: Optional, array of lowercase strings
- `relationships`: Optional, map of relationship types to target component paths
  - Key: Relationship type (e.g., "uses", "triggered_by", "publishes_to")
  - Value: Array of absolute component paths (system/container/component)

**Relationship Types** (recommended, not enforced):
- `uses` - Component depends on target
- `triggered_by` - Component is invoked by target
- `publishes_to` - Component sends data to target
- `reads_from` - Component reads data from target
- `writes_to` - Component writes data to target
- `subscribes_to` - Component listens to events from target

---

### 2. D2Relationship (New Entity)

**File**: `internal/core/entities/d2_relationship.go` (NEW)

**Definition**:
```go
// D2Relationship represents a relationship extracted from D2 diagram syntax.
type D2Relationship struct {
    Source string // Source component ID (extracted from D2 node)
    Target string // Target component ID (extracted from D2 node)
    Label  string // Arrow label (relationship type)
}

// NewD2Relationship creates a validated D2Relationship.
func NewD2Relationship(source, target, label string) (*D2Relationship, error) {
    if source == "" {
        return nil, errors.New("source cannot be empty")
    }
    if target == "" {
        return nil, errors.New("target cannot be empty")
    }
    return &D2Relationship{
        Source: source,
        Target: target,
        Label:  label, // Label can be empty (unlabeled arrow)
    }, nil
}

// Key returns a unique identifier for deduplication (source+target+label).
func (r *D2Relationship) Key() string {
    return fmt.Sprintf("%s->%s:%s", r.Source, r.Target, r.Label)
}
```

**Validation Rules**:
- `Source`: Required, non-empty string
- `Target`: Required, non-empty string
- `Label`: Optional, can be empty (unlabeled D2 arrow)

**Example D2 Extraction**:
```d2
# Input D2 file
email-queue: {
  shape: rectangle
  tooltip: "SQS queue"
}

email-queue -> email-sender: "triggers"
notification-api -> email-queue: "publishes to"
```

**Output**:
```go
[]D2Relationship{
    {Source: "email-queue", Target: "email-sender", Label: "triggers"},
    {Source: "notification-api", Target: "email-queue", Label: "publishes to"},
}
```

---

### 3. Graph Edge (Enhanced)

**File**: `internal/core/entities/graph.go` (line 92-114 - existing Edge struct)

**Current State**:
```go
type Edge struct {
    Source       string
    Target       string
    RelationType string
    Label        string
}
```

**Enhancement**: Populate from union merge of frontmatter + D2 relationships

**Source Priority**:
1. Frontmatter relationships → parse into edges
2. D2 relationships → parse into edges
3. Union merge: combine both, deduplicate by `(Source, Target, RelationType)` tuple
4. If same `(Source, Target)` exists in both with different `RelationType`, keep both as separate edges

**Deduplication Example**:
```go
// Frontmatter: email-queue uses email-sender
// D2: email-queue -> email-sender: "triggers"
// Result: 2 edges (different relationship types)
[]Edge{
    {Source: "email-queue", Target: "email-sender", RelationType: "uses"},
    {Source: "email-queue", Target: "email-sender", RelationType: "triggers"},
}

// Frontmatter: email-queue uses email-sender
// D2: email-queue -> email-sender: "uses"
// Result: 1 edge (deduplicated)
[]Edge{
    {Source: "email-queue", Target: "email-sender", RelationType: "uses"},
}
```

---

### 4. TemplateType (New Enum)

**File**: `internal/core/entities/template_selector.go` (NEW)

**Definition**:
```go
// TemplateType represents a category of component templates.
type TemplateType int

const (
    TemplateCompute TemplateType = iota
    TemplateDatastore
    TemplateMessaging
    TemplateAPI
    TemplateEvent
    TemplateStorage
    TemplateGeneric
)

// String returns the template file name for this type.
func (t TemplateType) String() string {
    switch t {
    case TemplateCompute:
        return "compute"
    case TemplateDatastore:
        return "datastore"
    case TemplateMessaging:
        return "messaging"
    case TemplateAPI:
        return "api"
    case TemplateEvent:
        return "event"
    case TemplateStorage:
        return "storage"
    case TemplateGeneric:
        return "generic"
    default:
        return "generic"
    }
}

// Technology patterns mapped to template types
var TechnologyPatterns = map[TemplateType][]string{
    TemplateCompute:    {"lambda", "function", "fargate", "ecs task"},
    TemplateDatastore:  {"dynamodb", "database", "table", "rds", "aurora"},
    TemplateMessaging:  {"sqs", "queue", "sns", "topic", "kinesis"},
    TemplateAPI:        {"api gateway", "rest", "graphql", "endpoint"},
    TemplateEvent:      {"eventbridge", "event", "step functions"},
    TemplateStorage:    {"s3", "bucket", "efs"},
}

// SelectTemplate selects a template type based on technology string.
// If override is provided, it takes precedence.
// Returns TemplateGeneric if no pattern matches.
func SelectTemplate(technology string, override *TemplateType) TemplateType {
    if override != nil {
        return *override
    }
    
    tech := strings.ToLower(technology)
    for tmplType, patterns := range TechnologyPatterns {
        for _, pattern := range patterns {
            if strings.Contains(tech, pattern) {
                return tmplType
            }
        }
    }
    return TemplateGeneric
}
```

---

### 5. DriftIssue (New Entity)

**File**: `internal/core/entities/drift_issue.go` (NEW)

**Definition**:
```go
// DriftSeverity represents the severity of a drift issue.
type DriftSeverity int

const (
    DriftWarning DriftSeverity = iota // Cosmetic inconsistencies
    DriftError                         // Broken references, data integrity issues
)

// DriftType categorizes the kind of drift detected.
type DriftType int

const (
    DriftDescriptionMismatch DriftType = iota  // D2 tooltip != frontmatter description
    DriftMissingComponent                      // D2 references non-existent component
    DriftOrphanedRelationship                  // Frontmatter relationship to deleted component
)

// DriftIssue represents a detected inconsistency between data sources.
type DriftIssue struct {
    ComponentID string        // Component where drift detected
    Type        DriftType     // Category of drift
    Severity    DriftSeverity // Warning or Error
    Message     string        // Human-readable description
    Context     string        // Additional context (e.g., expected vs actual)
}

// NewDriftIssue creates a validated DriftIssue.
func NewDriftIssue(componentID string, driftType DriftType, message string, context string) *DriftIssue {
    severity := DriftWarning
    if driftType == DriftMissingComponent || driftType == DriftOrphanedRelationship {
        severity = DriftError
    }
    
    return &DriftIssue{
        ComponentID: componentID,
        Type:        driftType,
        Severity:    severity,
        Message:     message,
        Context:     context,
    }
}
```

**Severity Rules**:
- **WARNING**: Description mismatch (D2 tooltip != frontmatter description) - cosmetic only
- **ERROR**: Missing component referenced in D2 - broken reference
- **ERROR**: Orphaned relationship in frontmatter to deleted component - data integrity issue

**Example**:
```go
// Description mismatch (WARNING)
&DriftIssue{
    ComponentID: "email-queue",
    Type:        DriftDescriptionMismatch,
    Severity:    DriftWarning,
    Message:     "D2 tooltip differs from frontmatter description",
    Context:     "D2: 'SQS queue' | Frontmatter: 'Standard SQS queue for email notifications'",
}

// Missing component (ERROR)
&DriftIssue{
    ComponentID: "email-queue",
    Type:        DriftMissingComponent,
    Severity:    DriftError,
    Message:     "D2 references non-existent component 'old-sender'",
    Context:     "Arrow: email-queue -> old-sender",
}
```

---

## Interface Definitions (Ports)

### D2Parser Interface

**File**: `internal/core/usecases/ports.go` (ENHANCE)

**Definition**:
```go
// D2Parser parses D2 diagram syntax to extract relationships.
type D2Parser interface {
    // ParseRelationships extracts relationship arrows from D2 source code.
    // Returns a slice of D2Relationship or an error if parsing fails.
    // Graceful degradation: returns error only for catastrophic failures,
    // partial parse errors should be logged and skipped.
    ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error)
}
```

**Error Handling Contract**:
- Parse error (invalid syntax): Return error with descriptive message
- Empty file: Return empty slice (valid state)
- Partial parse success: Return relationships successfully parsed + log warnings for failed portions

---

### TemplateRegistry Interface

**File**: `internal/core/usecases/ports.go` (ENHANCE)

**Definition**:
```go
// TemplateRegistry resolves template types to actual template file paths.
type TemplateRegistry interface {
    // GetTemplatePath returns the absolute path to the template file for the given type.
    // Returns error if template file does not exist.
    GetTemplatePath(templateType entities.TemplateType) (string, error)
    
    // ValidateTemplate checks if the template file exists and is readable.
    ValidateTemplate(templateType entities.TemplateType) error
}
```

**Contract**:
- Template files live in `templates/component/` directory
- File naming: `{template-type}.md` (e.g., `compute.md`, `datastore.md`)
- Returns error if template file missing or unreadable

---

## State Transitions

### Component Creation with Template Selection

```
1. User: loko new component --technology "DynamoDB"
2. CLI: Parse flags → technology = "DynamoDB"
3. Use Case: SelectTemplate("DynamoDB", nil) → TemplateDatastore
4. Use Case: registry.GetTemplatePath(TemplateDatastore) → "templates/component/datastore.md"
5. Use Case: Load template content
6. Use Case: Render template with component metadata
7. Use Case: Write component.md to filesystem
8. Result: Component created with datastore template (key schema, capacity, access patterns)
```

### Relationship Parsing & Union Merge

```
1. Trigger: loko build (or MCP query_architecture)
2. Use Case: Load all component frontmatter → extract relationships map
3. Use Case: Load all component D2 files → parse relationships
4. Use Case: Union merge:
   - For each frontmatter relationship: create Edge(source, target, type)
   - For each D2 relationship: create Edge(source, target, label)
   - Deduplicate by (source, target, type) tuple
5. Use Case: Add edges to architecture graph
6. Result: find_relationships, query_dependencies return populated results
```

### Drift Detection

```
1. User: loko validate --check-drift
2. Use Case: For each component:
   a. Load frontmatter description
   b. Parse D2 file tooltip
   c. Compare: if different → DriftIssue(WARNING, DescriptionMismatch)
   d. Check D2 arrows: if target doesn't exist → DriftIssue(ERROR, MissingComponent)
   e. Check frontmatter relationships: if target deleted → DriftIssue(ERROR, OrphanedRelationship)
3. Use Case: Aggregate all DriftIssues
4. CLI: Format as validation output (errors + warnings)
5. Result: Exit code 1 if any ERROR-severity drift, 0 if only WARNINGs
```

---

## Data Flow Diagrams

### Relationship Parsing Flow

```
┌─────────────────┐
│ Component Files │
└────────┬────────┘
         │
    ┌────┴─────┐
    │          │
┌───▼────┐ ┌──▼────┐
│ .md    │ │  .d2  │
│ (YAML) │ │(D2)   │
└───┬────┘ └──┬────┘
    │         │
    │         │ ParseRelationships()
    │         ▼
    │    ┌─────────────┐
    │    │ D2Parser    │
    │    │ (adapter)   │
    │    └──────┬──────┘
    │           │
    │     []D2Relationship
    │           │
    ▼           ▼
┌───────────────────────┐
│ BuildArchitectureGraph│
│    (use case)         │
│  - Union merge        │
│  - Deduplicate        │
└──────────┬────────────┘
           │
           ▼
    ┌──────────────┐
    │Architecture  │
    │   Graph      │
    │  (edges)     │
    └──────────────┘
```

### Template Selection Flow

```
┌──────────────┐
│ User Input   │
│ --technology │
└──────┬───────┘
       │
       │ "DynamoDB"
       ▼
┌──────────────────┐
│ TemplateSelector │
│   (entity)       │
│  SelectTemplate()│
└────────┬─────────┘
         │
         │ TemplateDatastore
         ▼
┌──────────────────┐
│TemplateRegistry  │
│   (adapter)      │
│ GetTemplatePath()│
└────────┬─────────┘
         │
         │ "templates/component/datastore.md"
         ▼
┌──────────────────┐
│ TemplateEngine   │
│   (adapter)      │
│   Render()       │
└────────┬─────────┘
         │
         │ Rendered content
         ▼
┌──────────────────┐
│ ProjectRepository│
│   (adapter)      │
│   WriteFile()    │
└──────────────────┘
```

---

## Validation Rules Summary

| Entity | Field | Rule |
|--------|-------|------|
| Component | relationships | Optional map, values must be valid component paths |
| D2Relationship | source | Required, non-empty |
| D2Relationship | target | Required, non-empty |
| D2Relationship | label | Optional, can be empty |
| DriftIssue | componentID | Required, must reference existing component |
| DriftIssue | severity | Auto-assigned based on drift type |
| TemplateType | pattern match | Case-insensitive, first match wins |

---

## Performance Characteristics

| Operation | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|-------|
| Parse single D2 file | O(n) | O(m) | n = file size, m = relationship count |
| Parse 100 D2 files | O(n) | O(m) | Concurrent with 10 workers → ~10ms wall time |
| Union merge | O(r log r) | O(r) | r = total relationships, sort + deduplicate |
| Template selection | O(p) | O(1) | p = pattern count (~20 patterns) |
| Drift detection | O(c × r) | O(d) | c = components, r = relationships per component, d = drift issues |

**Target**: 100 components @ <200ms total validation time

---

## References

- Component Entity: `internal/core/entities/component.go` (existing)
- Graph Entity: `internal/core/entities/graph.go` (existing)
- loko Constitution: `.specify/memory/constitution.md` v1.0.0
- D2 Language Reference: https://d2lang.com/tour/intro

---

**Status**: ✅ Data model complete - Ready for contracts generation
