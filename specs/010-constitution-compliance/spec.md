# Feature Specification: Constitution Compliance Refactor

**Feature Branch**: `010-constitution-compliance`
**Created**: 2026-05-20
**Status**: Draft
**Input**: User description: "Refactor the existing CLI handlers and MCP tool implementations to comply with the constitution layer rules and file size limits. Per docs/superpowers/specs/2026-05-08-loko-production-design.md (Constitution Compliance section), CLI handler functions must be under 50 lines, MCP tool handlers under 30 lines, use case files under 200 lines, and entity files under 300 lines. Strict layering must be enforced: internal/core/entities imports nothing from other internal packages; internal/core/usecases imports entities only; internal/adapters/* imports entities and usecases; cmd, internal/mcp, internal/api import adapters and usecases (never directly from entities). Known violations to address: cmd/new.go (504 lines, decompose into internal/core/usecases/scaffold_*.go), cmd/build.go (251 lines, move logic into internal/core/usecases/build_docs.go), and oversized MCP tools. Add an importrule (or equivalent) lint check to CI to prevent regression. No functional changes—pure refactor with full test coverage maintained."

## Clarifications

### Session 2026-05-21

- Q: Which task runner should host the new `audit-constitution` gating command? → A: Taskfile (`task audit-constitution`), not Make. Rationale: the project already uses Taskfile.yml as the canonical entry point for `task test`, `task lint`, `task build`, `task setup`; adding a `make`-based command would split the toolchain. (Plan-level detail; recorded here for traceability — does not change any functional requirement.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Bring Oversized CLI Handlers Within the Architectural Budget (Priority: P1)

As a maintainer, I need every command-line entry point to be a thin orchestration layer that delegates real work to the core domain, so the project's architectural constitution is honoured at the most-trafficked entry points of the product. Today, the project-scaffolding command and the documentation-build command are the two most egregious offenders: each contains the bulk of its logic inline, mixing user-facing concerns with domain logic and infrastructure access. Until they are decomposed, contributors will keep copying that pattern when adding new commands.

**Why this priority**: These two handlers are the largest and most visible violations called out in the production design document. Fixing them establishes the canonical "thin CLI handler → core use case" pattern that the rest of the refactor depends on, and it unblocks reuse of the extracted use cases by other layers (MCP, future API).

**Independent Test**: Run the project's structural compliance checker against the CLI package after the refactor. It reports zero handler-size violations and zero layer-import violations for the two named commands; behavioural tests for those commands all pass; end-users invoking the commands observe identical produced files, console output, exit codes, and error messages.

**Acceptance Scenarios**:

1. **Given** the previously oversized scaffolding command, **When** a contributor inspects the handler file, **Then** every handler function fits within the small-handler size budget and all scaffolding logic is reachable from one or more dedicated use cases in the core domain rather than from the command file.
2. **Given** the previously oversized documentation-build command, **When** a contributor inspects the handler file, **Then** every handler function fits within the small-handler size budget and the build orchestration is reachable from a dedicated use case in the core domain rather than from the command file.
3. **Given** a user invoking either of those commands with their full pre-refactor argument set, **When** the command runs to completion, **Then** the produced files, console output, exit code, and error messages are equivalent to the pre-refactor version under the same inputs.
4. **Given** the existing automated test suite, **When** it is run against the refactored code, **Then** every previously passing test still passes and coverage of the touched commands does not decrease.

---

### User Story 2 - Bring Oversized MCP Tool Handlers Within the Strict Per-Handler Budget (Priority: P1)

As a maintainer, I need every Model-Context-Protocol tool handler to remain a thin adapter that translates between the protocol and a core use case, so protocol-handling code never accumulates business logic. Today, several MCP tool handlers exceed the constitution's strict per-handler size budget, with domain logic inlined into transport code.

**Why this priority**: MCP tools are an externally exposed surface and an active growth area for the product; future tools will copy whatever pattern is set today. They are co-equal in priority with the CLI work because the MCP server is an independent, user-facing entry point — leaving it non-compliant defeats the purpose of fixing the CLI.

**Independent Test**: Run the structural compliance checker against the MCP tool package. Zero handlers exceed the per-handler size budget; every tool's domain logic is reachable from a use case in the core domain; the integration tests that exercise each MCP tool over the protocol still pass with byte-equivalent responses for the same inputs.

**Acceptance Scenarios**:

1. **Given** every MCP tool handler the project exposes today, **When** the handler files are inspected, **Then** every handler function fits within the strict MCP-handler size budget.
2. **Given** an external client calling any MCP tool, **When** the tool is invoked with its previously supported inputs, **Then** the tool returns results equivalent to the pre-refactor responses for the same inputs.
3. **Given** the existing MCP tool tests, **When** they are run against the refactored handlers, **Then** every previously passing test still passes.

---

### User Story 3 - Enforce Layer and Size Rules Mechanically and Continuously (Priority: P1)

As a maintainer, I need the layered architecture rules and file-size budgets to be enforced by an automated check that runs locally and in continuous integration, so contributors cannot accidentally re-introduce violations after the one-off refactor lands. Today the rules exist only as prose, and any violation has to be caught manually during review.

**Why this priority**: Refactoring without a guard is throwaway work — the same violations will reappear within weeks. The mechanical check is what converts the constitution from a guideline into a guarantee. It is co-equal in priority with Stories 1 and 2 because the value of the refactor is fully realised only when regressions are prevented.

**Independent Test**: On a throwaway branch, introduce one deliberate layer violation (e.g., import the entity layer directly from a CLI command file) and one deliberate file-size violation (e.g., balloon a use-case file past its budget). Run the check locally and open a pull request. Both violations are reported with file, offending entity (function or import path), and the specific rule broken; the pull request is blocked from merging until the violations are removed.

**Acceptance Scenarios**:

1. **Given** the entity layer, **When** the check runs, **Then** it fails on any entity-layer file that imports any other internal package of the project.
2. **Given** the use-case layer, **When** the check runs, **Then** it fails on any use-case file that imports anything from internal packages other than the entity layer.
3. **Given** the adapters layer, **When** the check runs, **Then** it fails on any adapter file that imports any internal package other than entities and use cases.
4. **Given** the outer entry-point packages (CLI commands, MCP server, HTTP API server), **When** the check runs, **Then** it fails on any file in those packages that imports the entity layer directly instead of obtaining entity types via adapters or via use-case return values.
5. **Given** any source file in the production codebase, **When** the check runs, **Then** it fails on any file that exceeds the file-size budget for its category.
6. **Given** the project's continuous integration pipeline, **When** a pull request introduces any layer or size violation, **Then** the pipeline fails before the change can be merged into the default branch.

---

### User Story 4 - Keep the Core Layer Compliant After Migration (Priority: P2)

As a maintainer, I need every use-case file and every entity file in the core domain to stay within its file-size budget after CLI and MCP logic is moved into it, so the refactor does not simply relocate the violations from the outer layers to the core. Without deliberate splitting, the use-case files that absorb the extracted logic will themselves cross the budget.

**Why this priority**: This story is sequenced after Stories 1–3 because the file-size pressure on the core arrives only once the outer-layer logic has been moved. It is a P2 — not P1 — because if Stories 1–3 land but the core files are slightly oversized at the moment of cutover, the mechanical check from Story 3 will catch it and Story 4 finishes the job in a follow-up commit. The story still must land in-scope of this feature so the gate from Story 3 is actually green on main.

**Independent Test**: Run the file-size portion of the structural compliance checker against the entity and use-case packages after the refactor. Every use-case file is within the use-case file budget; every entity file is within the entity file budget; reviewers confirm no use case has been chopped so finely that it loses narrative coherence.

**Acceptance Scenarios**:

1. **Given** the use-case package after the refactor, **When** the file-size check runs, **Then** every file is within the use-case file budget.
2. **Given** the entity package after the refactor, **When** the file-size check runs, **Then** every file is within the entity file budget.
3. **Given** a logically grouped use case (e.g., "scaffold a new project"), **When** a contributor opens the related files, **Then** related steps live together in cohesive files rather than being scattered across many trivial files just to satisfy the size budget.

---

### Edge Cases

- A function or file is exactly at its size limit boundary: the rule treats the limit as inclusive (e.g., 50 lines is allowed, 51 is not), and the counting convention used by the check tool (logical lines of code, excluding blank lines and comments) is documented in the project's compliance reference so contributors get the same answer the tool gets.
- A use case is split across multiple files inside one logical package and the package as a whole is large: the file-size budget applies per file, not per package, so well-organised multi-file use cases remain compliant.
- Generated code (e.g., protocol stubs, embedded asset registries) lives in the repository: generated files are exempt from file-size budgets but remain subject to layer-import rules, and the exemption mechanism is explicitly documented so it cannot be abused as a workaround.
- Test files are large because they enumerate many cases: test files are exempt from production-code file-size budgets so growth of test coverage is never penalised.
- A contributor adds a new outer-layer entry point (e.g., a new server or new CLI subcommand) that needs entity types: they must obtain those types via an adapter or a use-case return value, never by importing the entity package directly. The check's failure message tells them this on first attempt.
- The refactor moves a function whose previous name was part of an exported public surface: the externally observable public surface (command names, command-line flags and their help text, MCP tool names, MCP tool input and output schemas, exit codes, log lines that downstream tools may parse) is preserved exactly so external users see no change.
- A test relied on internals that moved during the refactor: the test is updated to call the new internal location, but the behaviour it verifies is unchanged. No test is deleted to make the refactor fit; if a behaviour was tested before, it is still tested after.
- The compliance check itself flags a pre-existing violation that is outside the declared scope of this refactor (e.g., a file owned by a different feature): the check supports a clearly-documented, time-bound suppression mechanism so the gate can be enabled on the main branch without first blocking unrelated work; every active suppression has an owner and a removal deadline.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every command-line handler function MUST fit within the small-handler size budget (no handler function exceeds 50 lines under the project's documented counting convention).
- **FR-002**: Every Model-Context-Protocol tool handler function MUST fit within the strict MCP-handler size budget (no MCP tool handler function exceeds 30 lines under the project's documented counting convention).
- **FR-003**: Every use-case file in the core domain MUST fit within the use-case file budget (no use-case file exceeds 200 lines under the project's documented counting convention).
- **FR-004**: Every entity file in the core domain MUST fit within the entity file budget (no entity file exceeds 300 lines under the project's documented counting convention).
- **FR-005**: The entity layer MUST NOT import from any other internal package of the project.
- **FR-006**: The use-case layer MUST NOT import from any internal package other than the entity layer.
- **FR-007**: The adapters layer MUST NOT import from any internal package other than the entity layer and the use-case layer.
- **FR-008**: The outer entry-point packages (CLI commands, MCP server, HTTP API server) MUST NOT import the entity layer directly; they MUST obtain entity types via adapters or via use-case return values.
- **FR-009**: The previously oversized project-scaffolding CLI handler MUST be decomposed so that its scaffolding logic lives in dedicated, narrowly-scoped use cases in the core domain, and the handler itself only parses inputs, calls the use case(s), and renders output.
- **FR-010**: The previously oversized documentation-build CLI handler MUST be decomposed so that its build orchestration lives in a dedicated use case in the core domain, and the handler itself only parses inputs, calls the use case, and renders output.
- **FR-011**: Every Model-Context-Protocol tool handler that previously exceeded the strict handler budget MUST be reduced to a thin protocol adapter that delegates domain logic to a use case.
- **FR-012**: An automated structural-compliance check (covering both file-size budgets and layer-import rules) MUST exist as part of the project's tooling and be runnable locally with a single documented command.
- **FR-013**: The automated structural-compliance check MUST be wired into the project's continuous integration pipeline as a gating check; failures of this check MUST block pull requests from being merged into the default branch.
- **FR-014**: The automated structural-compliance check MUST produce error messages that name, for each violation, the offending file, the offending entity (function name for size violations, import path for layer violations), and the specific rule that was broken.
- **FR-015**: The refactor MUST be behaviour-preserving: command names, command-line flags and their help text, MCP tool names, MCP tool input and output schemas, exit codes, and the contents of files produced by build and scaffolding commands MUST be equivalent before and after the refactor for the same inputs.
- **FR-016**: The existing automated test suite MUST pass against the refactored code, and overall test coverage of the touched packages MUST NOT decrease.
- **FR-017**: The compliance check MUST support a clearly-documented suppression mechanism for pre-existing violations outside the scope of this refactor, where each active suppression carries an owner and a removal date, so the gate can be enabled on the main branch without blocking unrelated work indefinitely.

### Key Entities *(include if feature involves data)*

- **Source File Category**: A classification (CLI command, MCP tool handler, use case, entity, adapter, generated, test) used by the compliance check to decide which size budget and which import rules apply to a given file.
- **Layer Rule**: A statement of the form "files in layer X may import only from layers Y, Z, …" that the check evaluates against the import graph of every source file.
- **Size Budget**: A maximum line count, per file or per function, associated with a source-file category and measured under the project's documented counting convention.
- **Violation Report**: A machine- and human-readable record naming the offending file, the offending entity (function for size violations, import path for layer violations), the rule broken, and (for size violations) the measured value versus the budget.
- **Suppression Entry**: A scoped, owner-tagged, dated exemption for a single pre-existing violation that allows the gate to be turned on without blocking work outside this feature's scope.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero handler-size violations remain in the CLI command package after the refactor, as reported by the structural compliance check.
- **SC-002**: Zero handler-size violations remain in the MCP tool package after the refactor, as reported by the structural compliance check.
- **SC-003**: Zero file-size violations remain in the core domain (use cases and entities) after the refactor, as reported by the structural compliance check.
- **SC-004**: Zero layer-import violations remain anywhere in the production codebase after the refactor, as reported by the structural compliance check.
- **SC-005**: The continuous integration pipeline blocks every pull request that introduces any structural violation, demonstrated by a deliberately-violating test pull request that is correctly rejected.
- **SC-006**: 100% of the project's previously passing automated tests still pass against the refactored code.
- **SC-007**: Test coverage of every package touched by the refactor is greater than or equal to the pre-refactor coverage for that package.
- **SC-008**: Users invoking any CLI command or MCP tool with its pre-refactor inputs observe equivalent outputs (produced files, console output, exit codes, error messages, MCP tool responses) — verified by a smoke-test run against a representative sample of commands and tools.
- **SC-009**: The structural compliance check completes in under 30 seconds on a typical contributor laptop, so it is cheap enough to be run on every save during development.
- **SC-010**: A contributor unfamiliar with the project can read the compliance check's failure message for any violation and locate the offending file and line, and identify the rule that was broken, without consulting additional documentation.

## Assumptions

- The constitution version that applies to this refactor is the one described in the Constitution Compliance section of `docs/superpowers/specs/2026-05-08-loko-production-design.md`; the size budgets (50 / 30 / 200 / 300) and the layer rules quoted in the user description are authoritative for the duration of this feature.
- The "counting convention" for handler-function and file-size limits is logical lines of code (excluding blank lines and comment-only lines), computed by the compliance check tool itself; whatever the tool measures is the canonical answer, and any drift between manual estimates and the tool's measurement is resolved in favour of the tool.
- The project's continuous integration pipeline already runs the project's existing lint and test commands and can be extended with one additional gating step without architectural change.
- The MCP server and HTTP API server packages are treated as outer entry-point layers and are required to honour the same "no direct entity-layer import" rule as the CLI commands package; if a prior constitution version exempted any of those, this refactor closes that gap.
- The set of pre-existing structural violations outside the named scope (CLI handlers, MCP tool handlers, core file sizes after migration) is small enough to be addressed either inside this feature or via the documented suppression mechanism — i.e., enabling the CI gate does not require a multi-week separate clean-up effort.
- No new third-party runtime dependency is needed to perform the refactor itself; the compliance check may use existing project tooling, the standard language toolchain, or a small custom tool added to the project's build-time dependencies.
- The refactor lands as a single feature branch merged to the default branch in one or more reviewable pull requests; no flag-guarded rollout is required because the change is internal-only and behaviour-preserving.
