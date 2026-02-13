---

description: "Task list for serverless architecture template implementation"
---

# Tasks: Serverless Architecture Template

**Input**: Design documents from `/specs/003-serverless-template/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-contract.md, quickstart.md

**Spec Version**: 0.1.0
**Status**: Ready
**Last Updated**: 2026-02-05

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Task can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Exact file paths included for clarity

## Progress Summary

| Phase | Status | Tasks | User Stories |
|-------|--------|-------|--------------|
| Phase 1: Setup | Complete | T001 | Setup |
| Phase 2: Foundational | Complete | T002-T005 | Blocking |
| Phase 3: US1+US3 Templates | Complete | T006-T012 | US-1, US-3 (P1) |
| Phase 4: US2 Template Selection | Complete | T013-T018 | US-2 (P1) |
| Phase 5: US4 Example Project | Complete | T019-T021 | US-4 (P2) |
| Phase 6: Polish | Complete | T022-T025 | Cross-cutting |

---

## Phase 1: Setup

**Purpose**: Verify existing infrastructure supports template changes

- [x] T001 Verify existing template engine and tests pass with `go test ./internal/adapters/ason/... ./internal/adapters/filesystem/...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Wire template engine into markdown and D2 generation so templates are actually used. MUST complete before any user story work.

- [x] T002 Update `SaveSystem()` in `internal/adapters/filesystem/project_repo.go` to try `templateEngine.RenderTemplate(ctx, "system.md", vars)` before falling back to `generateSystemMarkdown()` (follow existing pattern at lines 299-317)
- [x] T003 [P] Update `SaveContainer()` in `internal/adapters/filesystem/project_repo.go` to try `templateEngine.RenderTemplate(ctx, "container.md", vars)` before falling back to `generateContainerMarkdown()`
- [x] T004 [P] Update `SaveComponent()` in `internal/adapters/filesystem/project_repo.go` to try `templateEngine.RenderTemplate(ctx, "component.md", vars)` before falling back to `generateComponentMarkdown()` (D2 already done)
- [x] T005 Verify backward compatibility: run `go test ./...` and confirm all existing tests pass with template-first-with-fallback changes

**Checkpoint**: Template engine now drives all `.md` generation when templates are available. Existing behavior preserved via fallback.

---

## Phase 3: US1 + US3 - Serverless Template Files (Priority: P1)

**Goal**: Create all 6 serverless template files with serverless-specific content and event-driven D2 diagram patterns.

**Independent Test**: Scaffold a system, container, and component using the serverless template directory; verify generated files contain Lambda/event-driven terminology and D2 diagrams use dashed async lines.

**Note**: US1 (Scaffold Serverless Structure) and US3 (Event-Driven Diagram Patterns) are combined here because the template `.md` and `.d2` files are created together and share the same directory.

### Implementation

- [x] T006 [P] [US1] Create `templates/serverless/system.md` with ason variables (`{{SystemName}}`, `{{Description}}`, `{{Language}}`, `{{Framework}}`, `{{Database}}`) and sections: Overview, Event Sources, Lambda Functions, External Integrations, Technology Stack
- [x] T007 [P] [US3] Create `templates/serverless/system.d2` with API Gateway, Lambda icons, event source shapes, and `style.stroke-dash: 5` for async flows using ason variables (`{{SystemName}}`, `{{SystemID}}`, `{{Description}}`)
- [x] T008 [P] [US1] Create `templates/serverless/container.md` with ason variables (`{{ContainerName}}`, `{{Description}}`, `{{Technology}}`) and sections: Purpose, Trigger Type, Functions List, IAM Permissions, Environment Variables
- [x] T009 [P] [US3] Create `templates/serverless/container.d2` with event flow patterns using dashed lines for async communication, cloud service icons, and ason variables (`{{ContainerID}}`, `{{ContainerName}}`, `{{Description}}`, `{{Technology}}`)
- [x] T010 [P] [US1] Create `templates/serverless/component.md` with ason variables (`{{ComponentName}}`, `{{Description}}`, `{{Technology}}`) and sections: Handler, Trigger, Runtime, Memory, Timeout, Environment Variables, IAM Role
- [x] T011 [P] [US3] Create `templates/serverless/component.d2` with function trigger source, downstream targets, and ason variables (`{{ComponentID}}`, `{{ComponentName}}`, `{{Description}}`, `{{Technology}}`)
- [x] T012 [US1] Verify all 6 template files produce valid content by manually rendering with the ason engine test helper and checking for zero generic web server terminology

**Checkpoint**: All 6 serverless template files exist with correct ason syntax, serverless terminology, and event-driven D2 patterns.

---

## Phase 4: US2 - Template Selection Mechanism (Priority: P1)

**Goal**: Users can choose between `standard-3layer` and `serverless` templates when scaffolding entities.

**Independent Test**: Run `loko new system TestSystem -template serverless` and verify generated files use serverless template content. Run without `-template` flag and verify `standard-3layer` is used.

### Implementation

- [x] T013 [US2] Add `templateName string` field to `NewCommand` struct and `WithTemplate(name string) *NewCommand` method in `cmd/new.go`
- [x] T014 [US2] Add `-template` flag parsing in `handleNew()` in `main.go`, pass to `NewCommand` via `WithTemplate()`
- [x] T015 [US2] Replace hardcoded `"standard-3layer"` in `NewCommand.Execute()` search path construction in `cmd/new.go` (lines 80-81) with `nc.templateName` (default: `"standard-3layer"`)
- [x] T016 [P] [US2] Replace hardcoded `"standard-3layer"` in `handleBuild()` search path construction in `cmd/build.go` (lines 76, 80) with configurable template name
- [x] T017 [US2] Update D2 generation in `cmd/new.go` to try template engine rendering (`system.d2`, `container.d2`, `component.d2`) before falling back to hardcoded D2Generator and inline D2 methods
- [x] T018 [US2] Add template validation: if specified template directory doesn't exist, print error listing available templates (scan `templates/` directory)

**Checkpoint**: Users can select templates via `-template` flag. Default is backward compatible.

---

## Phase 5: US4 - Serverless Example Project (Priority: P2)

**Goal**: Complete example project at `examples/serverless/` demonstrating a realistic serverless architecture.

**Independent Test**: Run `loko validate` and `loko build` in the example directory; both pass.

### Implementation

- [x] T019 [US4] Create `examples/serverless/loko.toml` with project config for a serverless order processing system
- [x] T020 [US4] Create example system and container files in `examples/serverless/src/` with hand-crafted serverless content: order-processing system with api-handlers and event-processors containers, including `.md` and `.d2` files
- [x] T021 [US4] Validate example project: run `loko validate` and `loko build` against `examples/serverless/`, fix any issues

**Checkpoint**: Example project exists, validates, and builds.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and cleanup

- [x] T022 [P] Update `docs/quickstart.md` to document the `-template` flag with serverless examples
- [x] T023 [P] Update `README.md` to reflect that both `standard-3layer` and `serverless` templates are available (fix the dangling `examples/serverless/` reference)
- [x] T024 Run end-to-end quickstart validation: follow `specs/003-serverless-template/quickstart.md` commands and verify all steps work
- [x] T025 Run `go test ./...`, `go vet ./...`, and `golangci-lint run` to verify no regressions

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
    |
Phase 2 (Foundational - template engine wiring)
    |
    ├── Phase 3 (US1+US3 - Template files) ─┐
    |                                         ├── Phase 5 (US4 - Example project)
    └── Phase 4 (US2 - Template selection) ──┘
                                              |
                                         Phase 6 (Polish)
```

### User Story Dependencies

- **US-1 (Scaffold Structure)**: Depends on Phase 2 (template engine wiring). Template files can be created in parallel.
- **US-2 (Template Selection)**: Depends on Phase 2. Can run in parallel with US-1/US-3. The `-template` flag is independent of the template file content.
- **US-3 (Event-Driven Diagrams)**: Combined with US-1 since `.d2` files are created alongside `.md` files.
- **US-4 (Example Project)**: Depends on US-1+US-3 (template files must exist) and US-2 (selection mechanism must work for end-to-end validation).

### Within-Story Parallelization

#### US1+US3 Template Files
```
T006, T007, T008, T009, T010, T011 → T012
All 6 template files in parallel → Verification
```

#### US2 Template Selection
```
T013 → T014 → T015 → T017 → T018
                T016 (parallel with T015 - different file)
NewCommand field → flag parsing → search path → D2 rendering → validation
```

---

## Implementation Strategy

### MVP First (US1+US2 Only)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: Foundational (template engine wiring)
3. Complete Phase 3: US1+US3 (template files)
4. Complete Phase 4: US2 (template selection)
5. **STOP and VALIDATE**: Scaffold a full serverless project end-to-end
6. Deploy/demo if ready

### Incremental Delivery

1. Phase 2 (Foundational) → Template engine drives all generation
2. Phase 3 (US1+US3) → Serverless template files exist
3. Phase 4 (US2) → Users can select templates
4. Phase 5 (US4) → Example project for reference
5. Phase 6 (Polish) → Documentation and final validation

### Parallel Execution

Phases 3 and 4 can run in parallel after Phase 2 completes:
- **Track A**: Create all 6 template files (T006-T012)
- **Track B**: Wire CLI flag and search path changes (T013-T018)

---

## Notes

- **[P] tasks** = Different files, no dependencies between them
- **[Story] label** = Maps to user story for traceability
- All 6 template files (T006-T011) are fully parallel - different files in the same directory
- No new Go packages needed - all changes are in existing files
- Template variable names must exactly match the variables in `standard-3layer` templates
- Use `make test`, `make lint` or `task test`, `task lint` before commits
- No third-party mocking libraries needed - existing test patterns work
