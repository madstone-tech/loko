# Implementation Plan: UX Improvements from Real-World Feedback

**Branch**: `007-ux-improvements` | **Date**: 2026-02-13 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/007-ux-improvements/spec.md`

## Summary

Implement critical UX improvements discovered during real-world usage of loko v0.2.0:

1. **Functional Relationship Subsystem**: Parse relationships from frontmatter and D2 files into the architecture graph to make 4 MCP tools (`find_relationships`, `query_dependencies`, `query_related_components`, `analyze_coupling`) functional.

2. **Technology-Aware Templates**: Replace code-centric universal templates with 7 technology-specific templates (Compute, Datastore, Messaging, API, Event, Storage, Generic) selected automatically based on component technology field.

3. **Source of Truth & Drift Detection**: Establish frontmatter as metadata authority, D2 as visual/relationship authority, and add drift detection to catch inconsistencies.

**Technical Approach**:
- Use official D2 Go libraries (`oss.terrastruct.com/d2`) for parsing D2 syntax
- Union merge relationships from both frontmatter and D2 (deduplicate by source+target+type)
- Implement pattern-based template selection with override flag support
- Graceful degradation: skip malformed D2 files with warnings, continue processing
- Target: 100 components @ <200ms validation time

---

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**:
- `oss.terrastruct.com/d2` - Official D2 parsing libraries (NEW)
- `oss.terrastruct.com/d2/d2graph` - D2 graph structures (NEW)
- `oss.terrastruct.com/d2/d2lib` - D2 parsing functions (NEW)
- Existing: Cobra, Viper, Lipgloss, ason templates, d2 CLI

**Storage**: File system (component `.md` frontmatter, `.d2` files) - no changes  
**Testing**: Go test framework (`go test ./...`) with table-driven tests, >80% coverage in `internal/core/`  
**Target Platform**: Linux, macOS, Docker (single binary distribution)  
**Project Type**: Single project (CLI tool with embedded MCP/API servers)  

**Performance Goals**:
- Relationship parsing: <200ms for 100 components total
- Template selection: <1ms per component (pattern matching)
- D2 parse error recovery: graceful skip, no impact on other files
- Union merge deduplication: O(n) where n = relationship count

**Constraints**:
- Clean Architecture: core/ has zero external dependencies (D2 parsing in adapter layer)
- Thin Handlers: CLI < 50 lines, MCP < 30 lines
- Backward compatibility: Existing MCP clients and templates continue working
- Test coverage: Maintain > 80% in `internal/core/`
- Constitution compliance: All 7 gates must pass

**Scale/Scope**:
- Support projects with 100+ components (performance target)
- 7 technology-specific templates (vs 1 universal template currently)
- 17 MCP tools (15 existing + 2 enhanced: find_relationships, query_dependencies)
- Real-world validation: 17-component serverless notification service

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Gate 1: Clean Architecture (NON-NEGOTIABLE)

- [x] **Core has zero external dependencies**: D2 parsing library used in `internal/adapters/d2/`, not `core/`
- [x] **Dependency direction enforced**: Adapters → Core (D2Parser interface in ports.go)
- [x] **Port interfaces in usecases/ports.go**: `D2Parser` interface defined, implemented by adapter
- [x] **Use cases contain business logic**: `BuildArchitectureGraph` orchestrates relationship merging

**Status**: PASS - D2 parsing behind interface, template selection in entities/usecases

### ✅ Gate 2: Interface-First

- [x] **Ports defined before adapters**: `D2Parser` interface in ports.go before implementation
- [x] **No concrete adapter refs in use cases**: Use cases depend on interfaces only
- [x] **Wiring in main.go**: D2Parser adapter wired to use cases at startup

**Status**: PASS - All external dependencies (D2 libs) behind interfaces

### ✅ Gate 3: Thin Handlers

- [x] **CLI < 50 lines**: `cmd/new.go` adds `--template` flag, delegates to use case
- [x] **MCP < 30 lines**: `find_relationships` tool queries graph (graph now populated)
- [x] **Handlers parse/call/format only**: No business logic in handlers

**Status**: PASS - No new handler complexity introduced

### ✅ Gate 4: Entity Validation

- [x] **Validation in entity constructors**: Template selection logic in `TemplateSelector` entity
- [x] **Use cases trust valid entities**: Relationship merging uses validated Component entities
- [x] **No validation in handlers**: Validation remains in entity constructors

**Status**: PASS - Existing validation patterns maintained

### ✅ Gate 5: Test-First

- [x] **Unit tests for use cases**: Template selection, relationship parsing (5+ test cases each)
- [x] **Integration tests for adapters**: D2 parser with real D2 files
- [x] **E2E tests for handlers**: Create component with technology, verify template selected
- [x] **Target > 80% coverage**: Must maintain existing coverage

**Status**: PASS - Comprehensive test strategy documented (26 test cases total)

### ✅ Gate 6: Token Efficiency

- [x] **Progressive context loading**: Existing pattern unchanged
- [x] **TOON format support**: Existing TOON support unchanged
- [x] **JSON default, TOON opt-in**: No impact on encoding

**Status**: PASS - No changes to encoding layer

### ✅ Gate 7: Simplicity & YAGNI

- [x] **Simplest solution**: Pattern matching (not ML/AI), union merge (not conflict resolution)
- [x] **No premature abstraction**: 7 hardcoded templates (not template marketplace)
- [x] **Concrete mocks, no libraries**: Test mocks for D2Parser, TemplateRegistry
- [x] **Single binary**: D2 libraries statically linked, no new runtime dependencies

**Status**: PASS - Simple, pragmatic solutions

**Overall Gate Result**: ✅ **PASS** - No violations

---

## Project Structure

### Documentation (this feature)

```text
specs/007-ux-improvements/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (with clarifications)
├── research.md          # Phase 0 output - D2 parsing, template patterns
├── data-model.md        # Phase 1 output - Relationship, Template entities
├── quickstart.md        # Phase 1 output - Developer contribution guide
├── contracts/           # Phase 1 output - MCP tool contracts, template schemas
│   ├── d2-parser.md     # D2Parser interface contract
│   ├── template-registry.md  # TemplateRegistry interface contract
│   └── drift-detection.md    # Drift detection output schema
└── checklists/
    └── requirements.md  # Quality validation checklist
```

### Source Code (repository root)

```text
# loko follows Single Project structure (CLI tool with embedded servers)

internal/
├── core/                           # ZERO external dependencies
│   ├── entities/
│   │   ├── project.go              # Existing
│   │   ├── system.go               # Existing
│   │   ├── container.go            # Existing
│   │   ├── component.go            # Existing (Relationships map line 22)
│   │   ├── graph.go                # Existing (Edge support line 92-114)
│   │   ├── template_selector.go    # NEW - TemplateType enum, pattern matching
│   │   └── drift_issue.go          # NEW - DriftIssue with severity levels
│   └── usecases/
│       ├── ports.go                # ENHANCE - add D2Parser, TemplateRegistry interfaces
│       ├── create_component.go     # ENHANCE - select template by technology
│       ├── build_architecture_graph.go  # ENHANCE - parse frontmatter + D2, union merge
│       ├── validate_architecture.go     # ENHANCE - add drift detection
│       ├── render_diagram_preview.go    # NEW - render D2 to SVG/ASCII
│       └── detect_drift.go         # NEW - drift detection use case
│
├── adapters/
│   ├── filesystem/
│   │   └── project_repo.go         # ENHANCE - read relationships from frontmatter (line 917)
│   ├── d2/
│   │   ├── renderer.go             # Existing - diagram rendering
│   │   ├── parser.go               # NEW - D2 relationship parser using oss.terrastruct.com/d2
│   │   └── preview_renderer.go     # NEW - ASCII/SVG preview mode
│   ├── ason/
│   │   ├── engine.go               # Existing
│   │   └── template_registry.go    # NEW - map technology patterns to templates
│   └── encoding/
│       ├── json.go                 # Existing
│       └── toon.go                 # Existing
│
├── mcp/
│   ├── server.go                   # Existing
│   └── tools/
│       ├── find_relationships.go   # WORKS NOW - graph has edges
│       ├── query_dependencies.go   # WORKS NOW - graph has edges
│       ├── create_component.go     # ENHANCE - return diagram preview
│       └── ...                     # Other existing tools
│
├── api/                            # Existing (unchanged)
└── ui/                             # Existing (unchanged)

cmd/
├── root.go                         # Existing
├── new.go                          # ENHANCE - add --template, --preview flags
├── validate.go                     # ENHANCE - add --check-drift flag
└── ...                             # Other existing commands

templates/
└── component/
    ├── default.md                  # Existing (becomes compute.md)
    ├── compute.md                  # NEW - Lambda, Functions
    ├── datastore.md                # NEW - DynamoDB, RDS
    ├── messaging.md                # NEW - SQS, SNS, Kinesis
    ├── api.md                      # NEW - API Gateway, REST
    ├── event.md                    # NEW - EventBridge, Step Functions
    ├── storage.md                  # NEW - S3, EFS
    └── generic.md                  # NEW - Unknown technology minimal template

tests/
├── integration/
│   ├── relationship_parsing_test.go     # NEW - end-to-end relationship tests
│   ├── template_selection_test.go       # NEW - end-to-end template tests
│   └── drift_detection_test.go          # NEW - end-to-end drift tests
└── unit/
    ├── template_selector_test.go        # NEW - pattern matching tests
    └── d2_parser_test.go                # NEW - D2 parsing tests

docs/guides/
├── relationships.md                # NEW - Frontmatter + D2 relationship modeling
├── templates.md                    # NEW - Technology-aware templates guide
└── data-model.md                   # NEW - Source of truth hierarchy
```

**Structure Decision**: Single project structure maintained. loko is a CLI tool with embedded MCP/API servers. New D2 parsing adapter fits within `internal/adapters/d2/`, template selection in `internal/core/entities/`, following existing patterns.

---

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - table not needed. All gates pass.

---

## Phase 0: Research & Unknowns

**Status**: Research complete via agent tasks

**Deliverables**:
1. ✅ D2 Parsing Strategy - Use `oss.terrastruct.com/d2` official libraries
2. ✅ Template Selection Pattern - Pattern matching with TemplateSelector entity
3. ✅ Error Handling - Graceful degradation: skip file with warning, continue
4. ✅ Performance Strategy - Worker pool for concurrent parsing, target 100 components @ 200ms

**Output**: See `research.md` (to be generated in Phase 1)

---

## Phase 1: Design & Contracts

**Prerequisites**: Phase 0 research complete ✅

**Tasks**:
1. Generate `data-model.md` with:
   - D2Relationship entity (source, target, label)
   - TemplateType enum (Compute, Datastore, Messaging, API, Event, Storage, Generic)
   - DriftIssue entity (type, severity, message)
   - Component.Relationships map (already exists, document format)

2. Generate contracts in `contracts/`:
   - `d2-parser.md` - D2Parser interface contract (input: d2Source string, output: []D2Relationship)
   - `template-registry.md` - TemplateRegistry interface contract
   - `drift-detection.md` - DetectDrift output schema

3. Generate `quickstart.md`:
   - How to add new technology patterns to TemplateSelector
   - How to create custom templates
   - How to test relationship parsing

4. Update agent context:
   - Run `.specify/scripts/bash/update-agent-context.sh opencode`
   - Add D2 parsing libraries to tech stack
   - Add template selection patterns

**Output**: `data-model.md`, `contracts/`, `quickstart.md`, agent context updated

---

## Phase 2: Task Breakdown

**Status**: Deferred to `/speckit.tasks` command

**Output**: `tasks.md` (not created by this command)

---

## Implementation Timeline

**Sprint 1: Relationship Parsing** (6 hours):
- Days 1-2: D2 parser adapter + frontmatter parsing
- Days 2-3: Union merge logic in BuildArchitectureGraph
- Day 3: Integration tests, verify MCP tools work

**Sprint 2: Technology-Aware Templates** (8 hours):
- Days 1-2: TemplateSelector entity + pattern matching
- Days 2-3: Create 7 template files + TemplateRegistry adapter
- Days 3-4: CLI integration (--template flag) + MCP preview
- Day 4: Auto-generated component/container tables

**Sprint 3: Drift Detection** (6 hours, optional):
- Days 1-2: DetectDrift use case + severity logic
- Days 2-3: CLI integration (--check-drift flag)
- Day 3: Integration tests + documentation

**Total**: 14-20 hours (2-3 weeks with testing/polish)

---

## Risk Mitigation

| Risk | Mitigation Strategy |
|------|---------------------|
| D2 library API changes | Pin to specific version, add integration tests with real D2 files |
| Complex D2 syntax edge cases | Extensive test suite with 20+ real-world D2 examples |
| Template selection ambiguity | Explicit priority order in patterns, fallback to Generic |
| Performance degradation | Worker pool pattern, benchmark tests, 200ms target gate |
| Drift detection false positives | WARNING severity for cosmetic drift, ERROR for broken refs |

---

## Next Steps

1. ✅ **Constitution Check** - Passed all 7 gates
2. ✅ **Phase 0 Research** - Completed via agent tasks
3. ⏭️ **Phase 1 Design** - Generate data-model.md, contracts/, quickstart.md
4. ⏭️ **Phase 2 Tasks** - Run `/speckit.tasks` to generate detailed task breakdown

**Command to continue**: Run agent to generate Phase 1 artifacts (data-model.md, contracts/, quickstart.md)

---

## References

- **Feature Spec**: `specs/007-ux-improvements/spec.md`
- **Constitution**: `.specify/memory/constitution.md` v1.0.0
- **Feedback Sources**:
  - `test/loko-mcp-feedback.md` - Real-world MCP usage
  - `test/loko-product-feedback.md` - Real-world product feedback
- **Related Features**:
  - `specs/006-phase-1-completion/` - MVP Phase 1 (completed)
- **External Dependencies**:
  - D2 Libraries: https://github.com/terrastruct/d2
  - D2 Language Reference: https://d2lang.com/tour/intro

---

## Approval

**Status**: ✅ Ready for Phase 1 (Design & Contracts)

**Next Command**: Generate Phase 1 artifacts
