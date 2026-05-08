# Feature Specification: Constitution Compliance Refactor

**Feature Branch**: `009-constitution-compliance`
**Created**: 2026-05-08
**Status**: Draft
**Input**: User description: "Refactor the existing CLI handlers and MCP tool implementations to comply with the constitution layer rules and file size limits. Per docs/superpowers/specs/2026-05-08-loko-production-design.md (Constitution Compliance section), CLI handler functions must be under 50 lines, MCP tool handlers under 30 lines, use case files under 200 lines, and entity files under 300 lines. Strict layering must be enforced: internal/core/entities imports nothing from other internal packages; internal/core/usecases imports entities only; internal/adapters/* imports entities and usecases; cmd, internal/mcp, internal/api import adapters and usecases (never directly from entities). Known violations to address: cmd/new.go (504 lines, decompose into internal/core/usecases/scaffold_*.go), cmd/build.go (251 lines, move logic into internal/core/usecases/build_docs.go), and oversized MCP tools. Add an importrule (or equivalent) lint check to CI to prevent regression. No functional changes—pure refactor with full test coverage maintained."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Restore CLI Handler Compliance with Architectural Constitution (Priority: P1)

As a maintainer of the project, I need every command-line entry point to be a thin orchestration layer that delegates real work to the core domain, so that the codebase honours the architectural constitution that governs how this project must be structured. Today, the two largest CLI handlers (the project scaffolding command and the documentation build command) contain the bulk of their logic inline, mixing user-facing concerns with domain logic and infrastructure details, making them hard to read, test, and reuse.

**Why this priority**: The two named handlers are the most severe and most visible violations called out in the production design document. Until they are split, the constitution is openly contradicted at the most-trafficked entry points of the product, and contributors will continue to copy the same pattern when adding new commands. Fixing these unblocks every other compliance task.

**Independent Test**: Run the project's structural compliance checker (file size measurement plus layer-import verification) against the refactored CLI package. The checker reports zero violations for the previously oversized handlers, all behavioural tests for those commands still pass, and end-users observe identical command output, exit codes, and side effects compared to the previous release.

**Acceptance Scenarios**:

1. **Given** the previously oversized scaffolding command, **When** a contributor inspects the handler file, **Then** every handler function fits within the small-handler size budget and all scaffolding logic is reachable from a dedicated core use case rather than from the command file.
2. **Given** the previously oversized documentation build command, **When** a contributor inspects the handler file, **Then** every handler function fits within the small-handler size budget and the build orchestration is reachable from a dedicated core use case rather than from the command file.
3. **Given** a user invoking either of those commands with their full pre-refactor argument set, **When** the command runs to completion, **Then** the produced files, console output, exit code, and error messages are equivalent to the pre-refactor version under the same inputs.
4. **Given** the existing automated test suite, **When** it is run against the refactored code, **Then** every previously passing test still passes and overall test coverage of the touched commands does not decrease.

---

### User Story 2 - Enforce Strict Layer Boundaries Across the Codebase (Priority: P1)

As a maintainer, I need the layered architecture rules described in the production design document to be enforced mechanically and continuously, so that contributors cannot accidentally introduce dependencies that cross layer boundaries the wrong way. Today the rules exist only as prose, and any violation must be caught manually during review.

**Why this priority**: Refactoring the handlers is wasted effort if regressions can immediately reintroduce layer violations. Mechanical enforcement is what turns the constitution from a guideline into a guarantee, and the production design explicitly calls for an automated lint check.

**Independent Test**: Introduce a deliberate cross-layer dependency on a throwaway branch (e.g., importing an entity package directly from a CLI command file) and run the lint check. The check fails with a clear, actionable message that names the offending file, the offending import, and the rule that was broken. Reverting the change makes the check pass again.

**Acceptance Scenarios**:

1. **Given** the entity layer, **When** the lint check runs, **Then** it fails on any entity-layer file that imports any other internal package.
2. **Given** the use-case layer, **When** the lint check runs, **Then** it fails on any use-case file that imports anything from internal packages other than the entity layer.
3. **Given** the adapters layer, **When** the lint check runs, **Then** it fails on any adapter file that imports any internal package other than entities and use cases.
4. **Given** the CLI command package (`cmd/`), **When** the lint check runs, **Then** it fails on any CLI file that imports the entity layer directly instead of going through adapters or use cases. (Note: the MCP server and HTTP API packages remain permitted to import entities directly under v1.1.0; tightening that is a future-feature concern.)
5. **Given** the lint check is wired into the project's continuous integration pipeline, **When** any pull request introduces a layering violation, **Then** the pipeline fails before the change can be merged.

---

### User Story 3 - Bring Oversized MCP Tool Handlers Within Budget (Priority: P2)

As a maintainer, I need every Model-Context-Protocol tool handler to remain a thin adapter that translates between the protocol and a core use case, so that protocol-handling code never accumulates business logic. Today, several MCP tool handlers exceed the constitution's strict per-handler size budget, with domain logic inlined into transport code.

**Why this priority**: MCP tools are an externally exposed surface and a major growth area for the product, so future tools will copy whatever pattern is set today. Bringing them into compliance now is high-leverage, but it is sequenced behind the CLI work because the CLI handlers are the most severe violations and because some MCP tools will reuse the use cases extracted in Story 1.

**Independent Test**: Run the structural compliance checker against the MCP tool package. Zero handlers exceed the per-handler size budget, every tool's domain logic is reachable from a use case, and the integration tests that exercise each MCP tool over the protocol still pass with identical responses.

**Acceptance Scenarios**:

1. **Given** the catalogue of MCP tools the project exposes today, **When** their handler files are inspected, **Then** every handler function fits within the strict MCP-handler size budget.
2. **Given** an external client calling any MCP tool, **When** the tool is invoked with its previously supported inputs, **Then** the tool returns results that are equivalent to the pre-refactor responses for the same inputs.
3. **Given** the existing MCP tool tests, **When** they are run against the refactored handlers, **Then** every previously passing test still passes.

---

### User Story 4 - Keep Use-Case and Entity Files Within Their File-Size Budgets (Priority: P2)

As a maintainer, I need every use-case file and every entity file in the core domain to stay within its file-size budget, so that the core remains scannable and reviewable. As CLI and MCP logic is moved into the core, files in those layers risk crossing the budget unless they are split deliberately.

**Why this priority**: Once handler logic is relocated to the core, the file-size pressure shifts there. Without deliberate splitting, the project would simply move the violation from one layer to another. This story makes the migration "honest" by ensuring the destination is also compliant.

**Independent Test**: Run the file-size portion of the structural compliance checker against the entity and use-case packages after the refactor. Every use-case file is within its budget, every entity file is within its (more generous) budget, and no use case has been split so finely that it loses its narrative coherence (subjective check during code review).

**Acceptance Scenarios**:

1. **Given** the use-case package after the refactor, **When** the file-size check runs, **Then** every file is within the use-case file budget.
2. **Given** the entity package after the refactor, **When** the file-size check runs, **Then** every file is within the entity file budget.
3. **Given** a logically grouped use case (e.g., "scaffold a new project"), **When** a contributor opens the related files, **Then** related steps live together in cohesive files rather than being scattered across many trivial files just to satisfy the size budget.

---

### User Story 5 - Prevent Regression Through Continuous Integration Gates (Priority: P3)

As a maintainer, I need the structural compliance checker (layer-import rules and file-size budgets) to run automatically on every pull request and on the main branch, so that the constitution is enforced without manual vigilance.

**Why this priority**: The mechanical lint check from Story 2 only delivers durable value when it blocks merges. This story is sequenced last because it depends on the checks themselves existing and passing on the current code first; otherwise the gate would block all merges from day one.

**Independent Test**: Open a pull request that intentionally introduces both a layering violation and a file-size violation. The continuous integration run reports both failures with clear messages and prevents the pull request from being merged. Removing the violations in a follow-up commit makes the run pass.

**Acceptance Scenarios**:

1. **Given** the project's continuous integration pipeline, **When** a pull request is opened, **Then** the structural compliance check is one of the gating checks and a failure in it blocks merging.
2. **Given** a pull request that violates a layering rule, **When** the pipeline runs, **Then** the failure message names the offending file, the offending import, and the rule that was broken.
3. **Given** a pull request that exceeds a file-size budget, **When** the pipeline runs, **Then** the failure message names the offending file, its current size, and the budget for that file's category.

---

### Edge Cases

- A handler function is exactly at the size limit boundary: the rule treats the boundary as inclusive of the limit (i.e., 50 lines is allowed, 51 lines is not), and the chosen counting convention (logical lines of code, excluding blank lines and comments, computed by the lint tool itself) is documented in the project's compliance reference so contributors get the same answer the tool gets.
- A use case is split across multiple files inside one logical package and the package as a whole is large: the budget applies per file, not per package, so well-organised multi-file use cases remain compliant.
- Generated code (e.g., protocol stubs) lives in the repository: generated files are exempt from file-size budgets but are still subject to layer-import rules, and the exemption mechanism is explicitly documented so contributors do not abuse it.
- Test files are large because they enumerate many cases: test files are exempt from the production-code file-size budgets, with a separate (more generous) budget if any, so growth of test coverage is never penalised.
- A contributor adds a new outer-layer entry point (e.g., a new server) that needs entity types: they must obtain those types via an adapter or a use-case return value, never by importing the entity package directly. The lint message tells them this on first attempt.
- The refactor moves a function whose previous name was part of an exported public surface: the public surface (command names, command-line flags, MCP tool names, MCP tool input/output schemas, exit codes, log message wording where externally consumed) is preserved exactly so external users see no change.
- A test relied on internals that moved during the refactor: the test is updated to call the new internal location, but the behaviour it verifies is unchanged. No test is deleted to make the refactor fit; if a behaviour was tested before, it is still tested after.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every command-line handler function in the project MUST fit within the small-handler size budget (no handler function exceeds 50 lines under the project's documented counting convention).
- **FR-002**: Every Model-Context-Protocol tool handler function MUST fit within the strict MCP-handler size budget (no MCP tool handler function exceeds 30 lines under the project's documented counting convention).
- **FR-003**: Every use-case file in the core domain MUST fit within the use-case file budget (no use-case file exceeds 200 lines under the project's documented counting convention).
- **FR-004**: Every entity file in the core domain MUST fit within the entity file budget (no entity file exceeds 300 lines under the project's documented counting convention).
- **FR-005**: The entity layer MUST NOT import from any other internal package of the project.
- **FR-006**: The use-case layer MUST NOT import from any internal package other than the entity layer.
- **FR-007**: The adapters layer MUST NOT import from any internal package other than the entity layer and the use-case layer.
- **FR-008**: The CLI command package (`cmd/`) MUST NOT import the entity layer directly; it MUST obtain entity types via adapters or via use-case return values. Under constitution v1.1.0, the MCP server package and the HTTP API server package MAY still import the entity layer directly; tightening that restriction is deferred to a future feature so as not to expand this refactor's scope by ~18 additional files.
- **FR-009**: The previously oversized project-scaffolding CLI handler MUST be decomposed so that its core scaffolding logic lives in dedicated, narrowly-scoped use cases in the core domain, and the handler itself only parses inputs, calls the use case(s), and renders output.
- **FR-010**: The previously oversized documentation-build CLI handler MUST be decomposed so that its build orchestration logic lives in a dedicated use case in the core domain, and the handler itself only parses inputs, calls the use case, and renders output.
- **FR-011**: All Model-Context-Protocol tool handlers that previously exceeded the strict handler budget MUST be reduced to thin protocol adapters that delegate domain logic to use cases.
- **FR-012**: An automated structural-compliance check (covering both file-size budgets and layer-import rules) MUST exist as part of the project's tooling and be runnable locally with a single documented command.
- **FR-013**: The automated structural-compliance check MUST be wired into the project's continuous integration pipeline as a gating check; failures of this check MUST block pull requests from being merged into the default branch.
- **FR-014**: The automated structural-compliance check MUST produce error messages that name, for each violation, the offending file, the offending entity (function name for size violations, import path for layer violations), and the specific rule that was broken.
- **FR-015**: The refactor MUST be functionally non-observable to external consumers: command names, command-line flags and their semantics, command exit codes, MCP tool names, MCP tool input and output schemas, and the contents and structure of files produced by the project (e.g., scaffolded files, built documentation) MUST be identical to the pre-refactor behaviour for the same inputs.
- **FR-016**: The pre-existing automated test suite MUST pass against the refactored code without weakening assertions or removing tests; tests MAY be updated to point at relocated internal symbols, but only when the behaviour they verify is preserved.
- **FR-017**: Test coverage of the refactored CLI handlers and MCP tool handlers MUST NOT decrease relative to the pre-refactor baseline.
- **FR-018**: Generated files and test files MAY be exempt from the file-size budgets but MUST remain subject to the layer-import rules; any exemption mechanism MUST be documented in the project's compliance reference so contributors understand it.
- **FR-019**: The project's documentation MUST record the counting convention used by the structural-compliance check (i.e., what counts as a "line" for size budgets) so contributors get the same answer the tool gets.

### Key Entities *(include if feature involves data)*

- **Architectural Layer**: A named tier in the project's onion-style architecture. The layers, from innermost to outermost, are: entity layer, use-case layer, adapters layer, and the outer layer (CLI commands, the MCP server, and the API server). Each layer has explicit rules about which other layers it may depend on.
- **Layer Dependency Rule**: A constraint that says "files in layer X may only import from layers Y and Z." The rule set as a whole defines the project's architectural constitution and is enforced mechanically.
- **File-Size Budget**: A maximum permitted length, expressed in lines under a documented counting convention, that applies to a category of file (CLI handler file, MCP handler file, use-case file, entity file). Different categories have different budgets.
- **Handler-Function Size Budget**: A maximum permitted length, expressed in lines under a documented counting convention, that applies to an individual handler function within the CLI or MCP layer. Different handler categories have different budgets.
- **Structural Compliance Check**: The automated tool that evaluates the codebase against the layer dependency rules and the file-size and handler-size budgets and produces a pass/fail result with per-violation messages.
- **Compliance Violation**: A specific instance of a file or function that exceeds a budget or imports across layers in a forbidden direction. Each violation has a rule, a location (file path), and a subject (function name or import path).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After the refactor, the structural-compliance check reports zero violations across the entire codebase.
- **SC-002**: The two CLI handlers identified as the most severe violations (the project-scaffolding handler and the documentation-build handler) are each reduced to handler functions that all fit within the small-handler size budget, with no handler function in either file exceeding the budget.
- **SC-003**: 100% of Model-Context-Protocol tool handlers in the project fit within the strict MCP-handler size budget after the refactor.
- **SC-004**: The structural-compliance check runs as a gating step in the continuous integration pipeline on every pull request, and a deliberately-introduced violation on a test branch causes the pipeline to fail and prevents merging on at least one verification run.
- **SC-005**: The pre-existing automated test suite passes at the same or higher pass rate after the refactor as it did before, with no removed tests and no weakened assertions.
- **SC-006**: External behaviour (command outputs, exit codes, MCP tool responses, scaffolded files, built documentation) is equivalent between a sample of pre- and post-refactor runs over a representative set of inputs, with no observable behavioural difference.
- **SC-007**: A contributor can run the structural-compliance check locally with a single, documented command and obtain the same pass/fail result as the continuous integration pipeline.
- **SC-008**: When the compliance check fails, the message names the offending file, the offending entity (function name or import path), and the broken rule clearly enough that a contributor unfamiliar with the codebase can locate and understand the violation without further explanation.

## Assumptions

- The size budgets named in the feature description are the authoritative budgets for this feature: 50 lines for CLI handler functions, 30 lines for MCP tool handler functions, 200 lines for use-case files, 300 lines for entity files. The counting convention is whatever the chosen lint tool natively measures, documented in the project's compliance reference.
- The layered architecture described in the feature description is the authoritative architecture for this project: entity layer at the centre, use-case layer wrapping it, adapters layer wrapping that, and CLI/MCP/API as outer-layer entry points. No new layers are introduced by this feature.
- "No functional changes" is taken to mean no observable change to external behaviour: command names, flags, exit codes, MCP tool names and schemas, and files produced. Internal symbol locations, package layouts, function signatures of unexported helpers, and log lines that are not part of any external contract are free to change as needed by the refactor.
- The existing automated test suite is the baseline for behavioural equivalence; no new behavioural tests are required by this feature beyond what the existing suite covers, although tests may be added voluntarily where the refactor surfaces a gap.
- A suitable layer-enforcement lint tool exists in the project's ecosystem and can be configured declaratively. The specific tool is an implementation choice and is out of scope for this specification.
- Generated files (e.g., protocol stubs) and test files exist in the repository and need a documented exemption from the file-size budget while remaining subject to layer-import rules. If neither category exists in practice, the exemption mechanism is still documented for future use but does not exempt anything today.
- The continuous integration pipeline is the canonical enforcement point. If multiple pipelines exist (e.g., for different branches or for releases), the structural-compliance check is added to all of them that gate merges into the default branch.
- The two named CLI handler files (the project-scaffolding command and the documentation-build command) and the unspecified "oversized MCP tools" are the known violations as of the start of this feature; the refactor will discover the precise list of MCP tool violations during execution by running the size check, and will fix all of them.
