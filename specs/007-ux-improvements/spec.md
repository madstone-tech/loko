# Feature Specification: UX Improvements from Real-World Feedback

**Feature ID**: `007-ux-improvements`  
**Status**: Draft  
**Created**: 2026-02-13  
**Updated**: 2026-02-13  
**Authors**: Product feedback analysis from real-world usage  
**Reviewers**: TBD  

---

## Clarifications

### Session 2026-02-13

- Q: Should relationship sync be bidirectional (frontmatter ↔ D2) or unidirectional? → A: Unidirectional - D2 is read-only extraction into graph, frontmatter is authoritative write path, no sync back to D2
- Q: When a component has relationships in BOTH frontmatter and D2, how should the graph merge them? → A: Union merge (combine both sources, deduplicate by source+target+type)
- Q: When D2 parsing fails (invalid syntax), what should happen to relationship extraction? → A: Skip file with warning (log parse error, continue with other components)
- Q: What is the maximum number of components loko should support efficiently before performance optimization is required? → A: 100 components, 200ms max
- Q: Can users override the technology-to-template pattern matching? → A: Override flag support (--template compute/datastore/messaging/api/event/storage/generic)
- Q: When drift detection finds a component referenced in D2 but missing from filesystem, what severity? → A: Error (blocks validation, forces cleanup of broken references)

---

## Executive Summary

Implement critical UX improvements discovered during real-world usage of loko v0.2.0 to model a serverless notification service. Two major gaps surfaced:

1. **Broken Relationship Subsystem**: Four MCP tools (`find_relationships`, `query_dependencies`, `query_related_components`, `analyze_coupling`) return empty results because relationships exist in frontmatter and D2 files but are never parsed into the architecture graph.

2. **Code-Centric Templates**: Current templates apply code-component sections (Public Methods, Unit tests, Key Classes) universally, making 70% of sections irrelevant for infrastructure components (DynamoDB, SQS, Lambda, API Gateway, EventBridge).

These improvements address real user pain points before the v0.2.0 public release, ensuring a polished and functional experience.

---

## Problem Statement

### Current State

**Relationship Subsystem (Broken)**:
- Component `.md` files have `relationships` frontmatter field
- Component D2 files include relationship placeholder comments
- Architecture graph model supports edges/relationships
- Four MCP tools query relationships
- **Gap**: No code parses frontmatter or D2 into the graph → tools return empty results
- **Impact**: Users receive "17 isolated components" validation warnings with no way to fix them

**Template System (Misaligned)**:
- All components get identical code-component template
- DynamoDB tables have "Public Methods" sections
- SQS queues have "Unit tests" sections  
- API Gateway endpoints have "Key Classes/Functions" sections
- **Impact**: 70% of template sections are irrelevant, creating friction instead of guidance
- **Evidence**: In a 17-component serverless project, only Lambda functions benefited from current templates

### User Impact

**From MCP Feedback** (`test/loko-mcp-feedback.md`):
> "The relationship/graph features feel like scaffolding for something not yet built. There are four relationship-related tools that all return empty results."

**From Product Feedback** (`test/loko-product-feedback.md`):
> "Templates should reduce friction for the author, not create homework. If a section consistently gets left as placeholder text, it shouldn't be in the template."

### Success Criteria

1. ✅ `find_relationships` returns actual component relationships
2. ✅ `query_dependencies` shows graph connections
3. ✅ `analyze_coupling` provides meaningful metrics
4. ✅ D2 arrows are parsed into architecture graph (read-only)
5. ✅ DynamoDB components get datastore templates (key schema, capacity, access patterns)
6. ✅ Lambda components get compute templates (trigger, runtime, timeout, error handling)
7. ✅ SQS components get messaging templates (queue type, DLQ, retention)
8. ✅ Template preview shows rendered D2 diagram during `loko new component`
9. ✅ Container/system docs auto-generate component/container tables
10. ✅ Validation detects drift between D2 and frontmatter metadata

---

## User Stories

### Epic 1: Functional Relationship Subsystem

**User Story 1.1: Parse Frontmatter Relationships**
```
AS a software architect using loko MCP
I WANT relationships defined in component frontmatter to populate the architecture graph
SO THAT find_relationships and query_dependencies return actual results
```

**Acceptance Criteria**:
- Component frontmatter supports `relationships` map (e.g., `uses: ["system/container/other-component"]`)
- `BuildArchitectureGraph` reads frontmatter and creates graph edges
- `find_relationships` tool returns populated results
- Relationships display in `query_architecture` output

**User Story 1.2: Parse D2 Relationships**
```
AS a software architect authoring D2 diagrams
I WANT arrows in D2 files (component -> target: "label") to sync with the architecture graph
SO THAT my visual diagrams and graph queries stay consistent
```

**Acceptance Criteria**:
- D2 parser extracts arrows (e.g., `email-queue -> email-sender: "triggers"`)
- Arrow labels map to relationship types
- D2 arrows are parsed into graph (read-only extraction)
- Frontmatter is the authoritative write path for relationships
- No sync from frontmatter back to D2 (diagrams remain visual artifacts)
- Union merge: combine frontmatter + D2 relationships, deduplicate by source+target+type

**User Story 1.3: Relationship CRUD Tools (Optional Enhancement)**
```
AS a user of loko MCP
I WANT a create_relationship tool
SO THAT I can model dependencies without manually editing files
```

**Acceptance Criteria**:
- `create_relationship` tool with params: `source`, `target`, `relationship_type`, `label`
- Writes to frontmatter only (D2 diagrams remain manual/visual)
- `delete_relationship` tool for removal
- Validation detects orphaned relationships

---

### Epic 2: Technology-Aware Templates

**User Story 2.1: Template Selection by Technology**
```
AS a user creating a DynamoDB component
I WANT a datastore template with relevant sections (key schema, capacity, access patterns)
SO THAT I don't waste time deleting irrelevant "Public Methods" sections
```

**Acceptance Criteria**:
- Template registry maps technology patterns to template types:
  - `Lambda`, `Function` → Compute template
  - `DynamoDB`, `Database`, `Table` → Datastore template
  - `SQS`, `Queue`, `SNS Topic` → Messaging template
  - `API Gateway`, `REST`, `Endpoint` → API template
  - `EventBridge`, `Event` → Event template
  - `S3`, `Bucket` → Storage template
  - Unknown → Generic minimal template
- `loko new component --technology "DynamoDB"` selects datastore template
- `loko new component --template datastore` overrides pattern matching
- Templates include only relevant sections for that technology type

**Template Specifications**:

**Compute Template** (Lambda, Functions):
```markdown
## Configuration
- **Trigger**: (Event source - API Gateway, SQS, EventBridge)
- **Runtime**: (e.g., Node.js 20, Python 3.12)
- **Timeout**: (seconds)
- **Memory**: (MB)
- **Environment Variables**: (list key variables)

## Implementation
- **Handler**: (entry point function)
- **Dependencies**: (runtime dependencies)
- **Layers**: (Lambda layers used)

## Error Handling
- **Retry Policy**: (max retries, backoff strategy)
- **Dead Letter Queue**: (DLQ configuration)
- **Logging**: (CloudWatch log group)

## Performance Considerations
- **Cold Start**: (optimization strategies)
- **Concurrency**: (reserved/provisioned concurrency)
```

**Datastore Template** (DynamoDB, Databases):
```markdown
## Key Design
- **Partition Key**: (attribute name and type)
- **Sort Key**: (attribute name and type, if applicable)
- **Key Access Pattern**: (primary query pattern)

## Indexes
- **GSIs**: (Global Secondary Indexes - name, keys, projections)
- **LSIs**: (Local Secondary Indexes - name, keys, projections)

## Capacity & Performance
- **Capacity Mode**: (On-Demand or Provisioned)
- **Read/Write Units**: (if provisioned)
- **Auto-Scaling**: (scaling policy, if applicable)

## Data Management
- **TTL**: (Time-to-Live attribute, if enabled)
- **Backup**: (Point-in-time recovery, backup schedule)
- **Encryption**: (at-rest encryption settings)

## Access Patterns
- (List primary query patterns this table supports)
```

**Messaging Template** (SQS, SNS, Queues):
```markdown
## Queue Configuration
- **Queue Type**: (Standard or FIFO)
- **Visibility Timeout**: (seconds)
- **Message Retention**: (days)
- **Receive Wait Time**: (long polling setting)

## Dead Letter Queue
- **DLQ**: (target queue for failed messages)
- **Max Receive Count**: (retries before DLQ)

## Message Format
- **Schema**: (message structure/contract)
- **Content Type**: (JSON, XML, etc.)

## Throughput & Limits
- **Max Message Size**: (KB)
- **Throughput**: (messages/second estimate)
- **Batching**: (batch size for consumers)
```

**API Template** (API Gateway, REST):
```markdown
## Endpoint Configuration
- **Method**: (GET, POST, PUT, DELETE)
- **Path**: (resource path)
- **Stage**: (dev, staging, prod)

## Authentication
- **Auth Type**: (IAM, Cognito, API Key, Lambda Authorizer)
- **CORS**: (allowed origins, headers, methods)

## Request/Response
- **Request Schema**: (parameters, body structure)
- **Response Schema**: (success/error formats)
- **Status Codes**: (200, 400, 404, 500, etc.)

## Performance & Limits
- **Rate Limiting**: (throttle settings)
- **Caching**: (cache TTL, if enabled)
- **Timeout**: (integration timeout)
```

**Event Template** (EventBridge, Event-Driven):
```markdown
## Event Configuration
- **Event Bus**: (default or custom bus name)
- **Event Pattern**: (filter pattern for rule)
- **Source**: (event source identifier)

## Rule Configuration
- **Targets**: (Lambda, SQS, SNS, etc.)
- **Input Transformation**: (input path, template)
- **Retry Policy**: (max retries, backoff)

## Event Schema
- **Detail Type**: (event detail type)
- **Schema**: (event payload structure)

## Monitoring
- **Failed Invocations**: (DLQ for failed events)
- **Metrics**: (invocation count, errors)
```

**Storage Template** (S3, Buckets):
```markdown
## Bucket Configuration
- **Bucket Name**: (globally unique name)
- **Region**: (AWS region)
- **Versioning**: (enabled/disabled)

## Access Control
- **Bucket Policy**: (public/private, IAM policies)
- **CORS**: (cross-origin settings)
- **Encryption**: (SSE-S3, SSE-KMS, etc.)

## Lifecycle Management
- **Transition Rules**: (move to Glacier, expire objects)
- **Retention Policy**: (compliance requirements)

## Performance
- **Transfer Acceleration**: (enabled/disabled)
- **Event Notifications**: (trigger Lambda, SNS, SQS)
```

**Generic Template** (Unknown technology):
```markdown
## Overview
(Brief description of this component's purpose)

## Configuration
(Key configuration settings)

## Dependencies
(What this component connects to)

## Operational Notes
(Failure modes, monitoring, troubleshooting)
```

**User Story 2.2: D2 Diagram Preview**
```
AS a user creating a new component
I WANT to see the rendered D2 diagram preview
SO THAT I can verify the visual layout before committing
```

**Acceptance Criteria**:
- `loko new component` renders the scaffolded D2 file to SVG
- Preview displayed in terminal (ASCII art) or saved to temp file
- MCP `create_component` tool returns D2 preview in response
- Preview shows component position in parent container diagram

**User Story 2.3: Auto-Generated Component Lists**
```
AS a user editing container documentation
I WANT component tables to auto-populate
SO THAT I don't manually maintain lists that loko already knows
```

**Acceptance Criteria**:
- `container.md` generates table of components from filesystem
- `system.md` generates table of containers from filesystem
- Tables include: Name, Technology, Description (from frontmatter)
- `build_docs` regenerates tables on every build
- Manual edits preserved in separate "Notes" section

---

### Epic 3: Source of Truth & Drift Detection

**User Story 3.1: Frontmatter as Metadata Source of Truth**
```
AS a loko user
I WANT frontmatter to be the authoritative source for component metadata
SO THAT there's no ambiguity when D2 and markdown diverge
```

**Acceptance Criteria**:
- Documentation explicitly states: frontmatter = source of truth for metadata
- D2 files = source of truth for relationships and layout
- Markdown body = free-form documentation (not parsed)
- `validate` command enforces this hierarchy

**User Story 3.2: Drift Detection**
```
AS a user editing D2 diagrams
I WANT validation to detect when D2 descriptions diverge from frontmatter
SO THAT I catch inconsistencies early
```

**Acceptance Criteria**:
- `loko validate --check-drift` compares D2 tooltips to frontmatter descriptions
- Reports components where D2 references non-existent components (ERROR severity - broken references)
- Reports orphaned relationships (component deleted but relationship remains in frontmatter - ERROR severity)
- Reports description mismatches between D2 tooltips and frontmatter (WARNING severity - cosmetic only)

---

## Technical Design

### Architecture Changes

**No new layers or violations** - all changes fit within existing Clean Architecture:

```
internal/core/entities/
  component.go              # Relationships map already exists (line 22)
  graph.go                  # Edge support already exists (line 92-114)
  template.go               # NEW - TemplateType enum, TemplateRegistry

internal/core/usecases/
  build_architecture_graph.go  # ENHANCE - wire relationships from frontmatter/D2
  create_component.go          # ENHANCE - select template by technology
  validate_architecture.go     # ENHANCE - add drift detection
  render_diagram_preview.go    # NEW - render D2 to SVG/ASCII for preview

internal/adapters/
  filesystem/
    project_repo.go         # ENHANCE - parse D2 relationships (line 917 already reads them)
  d2/
    relationship_parser.go  # NEW - parse D2 arrow syntax
    renderer.go             # ENHANCE - add ASCII preview mode
  ason/
    template_registry.go    # NEW - map technology patterns to templates
    
cmd/
  new.go                    # ENHANCE - add --preview and --template flags
  validate.go               # ENHANCE - add --check-drift flag

internal/mcp/tools/
  create_component.go       # ENHANCE - return diagram preview in response
  find_relationships.go     # WORKS NOW - graph has edges
```

### Data Model Changes

**Component Frontmatter** (already supported, just document it):
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
---
```

**Template Registry Configuration** (`loko.toml`):
```toml
[templates]
compute = ["Lambda", "Function", "Fargate", "ECS Task"]
datastore = ["DynamoDB", "Database", "Table", "RDS", "Aurora"]
messaging = ["SQS", "Queue", "SNS", "Topic", "Kinesis"]
api = ["API Gateway", "REST", "GraphQL", "Endpoint"]
event = ["EventBridge", "Event", "Step Functions"]
storage = ["S3", "Bucket", "EFS"]
```

### D2 Relationship Parsing

**Input** (component D2 file):
```d2
email-queue: {
  shape: rectangle
  tooltip: "Standard SQS queue for email notifications"
}

email-queue -> email-sender: "triggers" {
  style.animated: true
}

notification-api -> email-queue: "publishes to"
```

**Output** (graph edges):
```go
// In internal/core/entities/graph.go
type Edge struct {
    Source         string // "notification-service/messaging-layer/email-queue"
    Target         string // "notification-service/processing-layer/email-sender"
    RelationType   string // "triggers"
    Label          string // "triggers"
}
```

### Template Selection Logic

```go
// In internal/adapters/ason/template_registry.go
func SelectTemplate(technology string) TemplateType {
    patterns := map[string]TemplateType{
        "Lambda|Function|Fargate": TemplateCompute,
        "DynamoDB|Database|Table|RDS": TemplateDatastore,
        "SQS|Queue|SNS|Topic|Kinesis": TemplateMessaging,
        "API Gateway|REST|GraphQL": TemplateAPI,
        "EventBridge|Event|Step Functions": TemplateEvent,
        "S3|Bucket|EFS": TemplateStorage,
    }
    
    for pattern, tmpl := range patterns {
        if regexp.MustCompile(pattern).MatchString(technology) {
            return tmpl
        }
    }
    return TemplateGeneric
}
```

---

## Implementation Plan

### Sprint 1: Relationship Parsing (6 hours)

**Phase 1: Parse Frontmatter Relationships** (2 hours)
- T200: Add relationship parsing to `ProjectRepository.LoadComponent()` (internal/adapters/filesystem/project_repo.go:917)
- T201: Update `BuildArchitectureGraph` to create edges from component relationships
- T202: Unit tests for relationship parsing (5 test cases)

**Phase 2: Parse D2 Relationships** (3 hours)
- T203: Create `d2.RelationshipParser` to extract arrows from D2 source (graceful degradation: skip file with warning on parse error)
- T204: Wire D2 parsing into `BuildArchitectureGraph` with union merge logic (deduplicate by source+target+type)
- T205: Integration test: create component with frontmatter + D2, verify graph has edges from both sources; test parse error handling

**Phase 3: Validation** (1 hour)
- T206: Test `find_relationships` returns results
- T207: Test `query_dependencies` shows graph
- T208: Verify "isolated components" warning disappears

### Sprint 2: Technology-Aware Templates (8 hours)

**Phase 1: Template Registry** (3 hours)
- T220: Create `TemplateType` enum (Compute, Datastore, Messaging, API, Event, Storage, Generic)
- T221: Implement `SelectTemplate(technology string, override *TemplateType) TemplateType` with pattern matching and override support
- T222: Create 7 template files in `templates/component/` directory
- T223: Update `CreateComponent` use case to call `SelectTemplate()` with optional override
- T224: Unit tests for template selection (9 test cases: 7 types + override + unknown)

**Phase 2: D2 Preview** (3 hours)
- T225: Add `RenderDiagramPreview(d2Source string) (string, error)` use case
- T226: Implement ASCII art renderer using d2 CLI `--sketch` flag
- T227: Update `loko new component --preview` to show rendered diagram
- T228: Update MCP `create_component` to return preview in response
- T229: Integration test: create component, verify preview rendered

**Phase 3: Auto-Generated Lists** (2 hours)
- T230: Add `GenerateComponentTable(containerPath string) (markdown string)` helper
- T231: Update `container.md` template with `{{component_table}}` placeholder
- T232: Update `system.md` template with `{{container_table}}` placeholder
- T233: Wire generation into `build_docs` use case
- T234: Integration test: verify tables populated in built docs

### Sprint 3: Drift Detection (Optional, 6 hours)

**Phase 1: Drift Detection Logic** (3 hours)
- T240: Add `DetectDrift(component Component) []DriftIssue` use case (supports ERROR and WARNING severity)
- T241: Compare D2 tooltip vs frontmatter description (WARNING severity)
- T242: Detect orphaned relationships (component deleted, relationship remains - ERROR severity)
- T243: Detect missing components referenced in D2 (ERROR severity)

**Phase 2: CLI Integration** (2 hours)
- T244: Add `--check-drift` flag to `loko validate`
- T245: Format drift warnings in validation output
- T246: Integration test: introduce drift, verify detection

**Phase 3: Documentation** (1 hour)
- T247: Update docs with source of truth hierarchy
- T248: Add drift detection examples to validation guide

---

## Testing Strategy

### Unit Tests
- Template selection logic (9 test cases: 7 types + override + unknown)
- Relationship parsing from frontmatter (5 test cases: none, single, multiple, invalid, circular)
- D2 relationship parsing (7 test cases: simple arrow, labeled, multi-target, invalid syntax with graceful skip, union merge with frontmatter, deduplication, empty file)
- Drift detection (5 test cases: no drift, description mismatch WARNING, orphaned relationship ERROR, missing component ERROR, multiple drift types)

### Integration Tests
- End-to-end: Create DynamoDB component, verify datastore template used
- End-to-end: Create component with relationships, verify graph edges created
- End-to-end: Update D2 diagram, verify relationships sync to graph
- End-to-end: Build docs, verify component tables auto-generated
- End-to-end: Introduce drift, verify `--check-drift` detects it

### Real-World Validation
- Re-run serverless notification service project creation (17 components)
- Verify all 4 relationship tools return results
- Verify DynamoDB/SQS/Lambda components get appropriate templates
- Verify validation no longer shows "17 isolated components" warning

---

## Documentation Updates

### New Guides
1. **Relationship Modeling Guide** (`docs/guides/relationships.md`)
   - Frontmatter relationship syntax (authoritative write path)
   - D2 arrow syntax (read-only extraction)
   - How both sources merge into the graph
   - Troubleshooting orphaned relationships

2. **Template Guide** (`docs/guides/templates.md`)
   - Technology-to-template mapping (automatic pattern matching)
   - Template selection override (`--template compute|datastore|messaging|api|event|storage|generic`)
   - Custom template creation (for future: custom file paths)
   - Contributing new templates

3. **Source of Truth Hierarchy** (`docs/guides/data-model.md`)
   - Frontmatter = metadata source of truth
   - D2 = relationships & layout source of truth
   - Markdown body = free-form docs
   - Drift detection workflow

### Updated Guides
- **MCP Integration Guide**: Document relationship tools now return results
- **CLI Reference**: Add `--preview`, `--check-drift` flags
- **Validation Guide**: Add drift detection section

---

## Success Metrics

### Functionality
- [ ] `find_relationships` returns non-empty results for test project
- [ ] `query_dependencies` shows graph with 10+ edges
- [ ] `analyze_coupling` provides coupling metrics
- [ ] DynamoDB components have 0 irrelevant template sections
- [ ] Lambda components have 0 irrelevant template sections

### User Experience
- [ ] "Isolated components" warnings only for genuinely isolated components
- [ ] Template selection happens automatically (no manual editing)
- [ ] Component/container tables auto-populate (no manual maintenance)
- [ ] Drift detection catches 100% of test divergence cases
- [ ] Validation completes in <200ms for 100-component projects

### Code Quality
- [ ] > 80% test coverage in `internal/core/`
- [ ] Zero Clean Architecture violations
- [ ] Thin handlers maintained (CLI < 50 lines, MCP < 30 lines)
- [ ] Constitution compliance (all gates pass)

---

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| D2 parsing breaks on complex diagrams | High | Medium | Extensive test suite with real-world D2 examples |
| Template proliferation (users want 20+ templates) | Medium | High | Start with 7 core templates, document custom template creation |
| Conflicting relationship sources (frontmatter vs D2) | Medium | Low | D2 read-only extraction, frontmatter is write path, merge both sources into graph |
| Performance degradation (parsing D2 on every build) | Medium | Low | Cache parsed D2, invalidate on file change; target 100 components @ 200ms max |
| Drift detection false positives | Low | Medium | Make drift warnings (not errors), allow intentional divergence |

---

## Future Enhancements (Out of Scope)

### Deferred to v0.3.0
- `create_relationship` MCP tool (CRUD interface)
- Bulk relationship creation tool
- Relationship visualization in `build_docs` (cross-container dependency graph)
- `loko enrich <component>` command (prompt-based section expansion)
- Custom template repository support
- Template marketplace/registry

### Deferred to v0.4.0
- SQLite relationship store (if filesystem proves insufficient)
- Visual diagram editor (web-based D2 editor)
- AI-powered template suggestions based on description
- Dependency cycle detection in validation
- Component compliance scoring (documentation completeness)

---

## References

### Source Documents
- `test/loko-mcp-feedback.md` - Real-world MCP usage feedback
- `test/loko-product-feedback.md` - Real-world product feedback
- `specs/006-phase-1-completion/spec.md` - Phase 1 MVP specification
- `.specify/memory/constitution.md` - Architecture constraints (v1.0.0)

### Related ADRs
- ADR-0001: Clean Architecture
- ADR-0002: Token-Efficient MCP
- ADR-0003: TOON Format
- ADR-0004: Graph Conventions

### External References
- [TOON Format v3.0 Specification](https://github.com/toon-format/toon-spec)
- [D2 Language Reference](https://d2lang.com/tour/intro)
- [C4 Model](https://c4model.com/)

---

## Approval

**Product Owner**: TBD  
**Technical Lead**: TBD  
**Reviewers**: TBD  

**Status**: ✅ Ready for implementation (Path C - comprehensive UX improvement)

---

## Changelog

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2026-02-13 | 1.0.0 | AI Agent | Initial specification from real-world feedback |
