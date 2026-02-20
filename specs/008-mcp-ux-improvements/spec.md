# Feature Specification: MCP UX Improvements

**Feature Branch**: `008-mcp-ux-improvements`  
**Created**: 2026-02-19  
**Status**: Draft  
**Input**: User feedback from real-world Loko MCP session documenting a Go/AWS serverless project (Agwe) — 6 usability gaps identified across relationship tooling, diagram initialization, validation messaging, batch operations, error surfacing, and graph query reliability.

## Clarifications

### Session 2026-02-19

- Q: Where should relationships be persisted on disk? → A: Dedicated system-level file — a `relationships.toml` per system directory holding all relationships for that system.
- Q: When a container or component is deleted, what happens to relationships referencing it? → A: Auto-remove — all relationships referencing the deleted element are removed from `relationships.toml` automatically.
- Q: What role does the `veve` CLI play in the diagram generation pipeline for relationship-driven diagram updates? → A: PDF export only — `veve` is not involved in D2 diagram writes; diagram files are written directly by loko.
- Q: How should the in-memory ArchitectureGraph cache be invalidated when relationships are written? → A: Eager invalidation — invalidate the cache immediately on every `create_relationship` or `delete_relationship` call; next read triggers a full rebuild.
- Q: Should isolated_component suppression be project-wide or per-container when no relationships exist? → A: Project-wide — suppress all isolated_component findings when relationships.toml contains zero relationships across the entire project.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Architecture Relationships Without Writing Diagrams (Priority: P1)

An architect using an AI agent or the MCP directly wants to express that one container calls another — for example, "the API Lambda enqueues jobs onto the SQS queue using the AWS SDK." Today they must write raw D2 diagram syntax to capture this. They should be able to state the relationship in structured terms and have it stored, queryable, and reflected in diagrams automatically.

**Why this priority**: Relationships are the primary value of a C4 model. Without first-class relationship tooling, the architecture graph is not queryable, `query_dependencies` and `query_related_components` return empty results after a standard setup session, and non-technical users cannot capture data flows at all. This is the single highest-impact gap.

**Independent Test**: Create a system with two containers. Use `create_relationship` to link them. Confirm the relationship appears in `list_relationships`, is reflected in the container diagram, and that `query_dependencies` returns it — all without touching D2 syntax directly.

**Acceptance Scenarios**:

1. **Given** a project with two containers, **When** the user calls `create_relationship` with source, target, and label, **Then** the relationship is stored and returned by `list_relationships`.
2. **Given** a stored relationship, **When** the user calls `query_dependencies` for the source container, **Then** the target container appears in the results.
3. **Given** a stored relationship, **When** the user calls `query_related_components` for elements in either container, **Then** cross-container relationships are surfaced.
4. **Given** a stored relationship, **When** the container diagram is viewed, **Then** an edge connecting source to target appears with the specified label.
5. **Given** a stored relationship, **When** the user calls `delete_relationship` with the relationship ID, **Then** the relationship is removed from storage and no longer appears in queries or diagrams.
6. **Given** a relationship with an invalid source or target ID, **When** `create_relationship` is called, **Then** an error message is returned indicating the element was not found and suggesting the correct slugified ID.

---

### User Story 2 - Initialize Container Diagrams on Creation (Priority: P2)

An architect calls `create_container` to add a new container to their system. Today, the response says "Use 'update_diagram' tool to add D2 diagram," while `create_system` generates a diagram scaffold automatically. The user expects consistent behaviour — a minimal diagram placeholder should exist immediately after creation so no extra round-trips are required before the container is visible.

**Why this priority**: This is a low-effort fix with noticeable impact on first-use experience. The inconsistency is surprising and creates unnecessary extra tool calls in every container setup session.

**Independent Test**: Call `create_container` on a new system. Confirm the container directory contains a diagram file with at least a placeholder template — without any additional `update_diagram` call.

**Acceptance Scenarios**:

1. **Given** a system exists, **When** `create_container` is called, **Then** a D2 diagram file is created alongside the container metadata containing at minimum a scaffold (e.g., container node and placeholder for components).
2. **Given** `create_system` generates a diagram, **When** `create_container` is called, **Then** the response no longer contains the message "Use 'update_diagram' tool to add D2 diagram."
3. **Given** a container is created with known components (via batch create), **When** the diagram is initialized, **Then** the scaffold includes nodes for those components.

---

### User Story 3 - Batch Create Components in One Call (Priority: P3)

An architect documenting a project with many components (e.g., 18 across 9 containers) today must make one `create_component` call per component. This is slow, fills conversation context, and requires the agent to loop over a list manually. They should be able to pass an array of components and have them all created in a single tool call.

**Why this priority**: High friction for any real-world project. The fix reduces round-trips significantly and makes initial documentation sessions faster and less noisy in the conversation context.

**Independent Test**: Call `create_components` (plural) with an array of 5 component definitions for a single container. Confirm all 5 components are created and retrievable via `query_architecture`.

**Acceptance Scenarios**:

1. **Given** a container exists, **When** `create_components` is called with an array of component definitions, **Then** all components in the array are created and returned in the response.
2. **Given** a batch create call where one component name conflicts with an existing component, **When** the call is processed, **Then** the conflicting component returns an error but all non-conflicting components are still created.
3. **Given** a batch create call with an empty array, **When** the call is processed, **Then** a clear error is returned indicating at least one component is required.
4. **Given** a batch create call with 20 components, **When** the call completes, **Then** all 20 are queryable and the response includes the generated slugified ID for each.

---

### User Story 4 - Validate Produces Actionable Isolation Messages (Priority: P4)

An architect runs `validate` immediately after setting up a new project — before any relationships have been defined. Today, every component is flagged as `isolated_component` at info severity, creating the impression that 18 things are wrong before the user has even started adding relationships. The output should distinguish between "you may have a problem" and "here is the logical next step."

**Why this priority**: False-alarm noise on new projects erodes trust in the validation tool. It is a low-effort fix to the message wording and/or suppression logic that improves the first-run experience.

**Independent Test**: Create a project with two components and no relationships. Run `validate`. Confirm the isolation finding is framed as a next-step prompt rather than an error, or is suppressed when no relationships exist at all.

**Acceptance Scenarios**:

1. **Given** a project with components but zero relationships in `relationships.toml`, **When** `validate` is called, **Then** no `isolated_component` findings are emitted — the output contains zero isolation entries.
2. **Given** a project where some components have relationships and others do not, **When** `validate` is called, **Then** only components without any relationship are flagged, and the finding is still framed constructively.
3. **Given** a project with at least one relationship defined, **When** `validate` is called, **Then** components genuinely isolated from all relationships are flagged at the appropriate severity without suppression.

---

### User Story 5 - Surface Slugified IDs in Error Messages (Priority: P5)

An agent or user passes a display name (e.g., "Authorizer Lambda") where a slugified ID (e.g., "authorizer-lambda") is expected. Today the error message says the element was not found, with no hint about the correct format. The error should suggest the likely correct ID.

**Why this priority**: Low effort, directly reduces recovery time when ID format mistakes occur — especially for AI agents that may guess the wrong slug format.

**Independent Test**: Call any tool with a container or component name in display format (with spaces and mixed case). Confirm the error message includes the likely slugified ID and/or a clear suggestion.

**Acceptance Scenarios**:

1. **Given** a container "API Lambda" exists (ID: `api-lambda`), **When** a tool is called with `container_name: "API Lambda"`, **Then** the error message includes the likely correct ID: "did you mean 'api-lambda'?"
2. **Given** a tool call with an entirely unknown element name, **When** the name cannot be slugified to a known ID, **Then** the error message still explains that IDs are slugified and advises the user to run `query_architecture` to retrieve the correct ID.
3. **Given** a successful create call, **When** the response is returned, **Then** the generated slugified ID is always included in the response body.

---

### Edge Cases

- What happens when `create_relationship` is called with source equal to target (self-reference)?
- What happens when `delete_relationship` is called with a non-existent relationship ID?
- What happens when a batch `create_components` call is partially valid — some components have duplicate names and some are new?
- What happens when container diagram initialization fails (e.g., filesystem write error)?
- What happens when `list_relationships` is called on a system with no relationships defined — does it return an empty list or an error?
- When a relationship source or target element is deleted, all relationships referencing it are automatically removed from `relationships.toml` — no dangling references are left.

## Requirements *(mandatory)*

### Functional Requirements

**Relationship Management**

- **FR-001**: The system MUST provide a `create_relationship` tool that accepts source element path, target element path, label, and optional relationship type (sync/async/event) and direction.
- **FR-002**: The system MUST provide a `list_relationships` tool that returns all relationships for a given system, optionally filtered by source or target element.
- **FR-003**: The system MUST provide a `delete_relationship` tool that removes a stored relationship by ID.
- **FR-004**: Stored relationships MUST be reflected in the relevant container and/or system diagram without requiring a separate `update_diagram` call.
- **FR-005**: `query_dependencies` and `query_related_components` MUST use stored relationships (from `create_relationship`) as their primary data source, not only D2 edge parsing.
- **FR-006**: When `create_relationship` is called with an invalid source or target, the error MUST include the expected slugified ID format or suggest running `query_architecture` to retrieve valid IDs.
- **FR-016**: When a container or component is deleted, the system MUST automatically remove all relationships in `relationships.toml` that reference the deleted element as source or target — no dangling references may remain after a delete operation.

**Container Diagram Initialization**

- **FR-007**: When `create_container` is called, the system MUST automatically generate a D2 diagram file in the container directory, containing at minimum a named placeholder node for the container.
- **FR-008**: The `create_container` response MUST NOT contain the message "Use 'update_diagram' tool to add D2 diagram" — it MUST reference the auto-generated diagram file path instead.

**Batch Component Creation**

- **FR-009**: The system MUST provide a `create_components` tool that accepts a system name, container name, and an array of component definitions (name, description, technology, and optional tags/responsibilities).
- **FR-010**: Batch component creation MUST process all valid components even if one or more definitions in the array conflict with existing components; per-item errors MUST be returned alongside successful results.
- **FR-011**: The response from `create_components` MUST include the generated slugified ID for each successfully created component.

**Validation Messaging**

- **FR-012**: When `validate` is called on a project where `relationships.toml` contains zero relationships across the entire project, ALL `isolated_component` findings MUST be suppressed entirely — no isolation findings are emitted.
- **FR-013**: When `validate` is called on a project with at least one relationship defined, components genuinely isolated from all relationships MUST still be flagged at appropriate severity without suppression.

**Graph Cache Consistency**

- **FR-017**: The in-memory architecture graph cache MUST be eagerly invalidated immediately after any `create_relationship` or `delete_relationship` call completes, so that subsequent `query_dependencies` and `query_related_components` calls always reflect the current state of `relationships.toml`.

**ID Surfacing in Error Messages**

- **FR-014**: When any tool receives an element name that does not match a known ID, the error message MUST include the likely slugified form of the provided name (e.g., "did you mean 'api-lambda'?").
- **FR-015**: Every create tool response (system, container, component, relationship) MUST include the generated slugified ID of the created element.

### Key Entities

- **Relationship**: A directed or undirected connection between two architecture elements (containers or components), with a label, optional technology, optional type (sync/async/event), and a system-generated ID. Persisted in a dedicated `relationships.toml` file at the system directory level (e.g., `src/<system-name>/relationships.toml`); D2 diagram edges are generated from this file as a side effect.
- **Element Path**: A slash-separated string identifying an element within the architecture hierarchy (e.g., `system-name/container-name` or `system-name/container-name/component-name`). Used as the canonical identifier in relationship source/target fields.

### Assumptions

- Relationships are stored in a `relationships.toml` file per system directory (not in component frontmatter and not derived solely from D2 parsing); D2 is updated as a side effect of relationship writes.
- Diagram auto-initialization for containers uses a minimal scaffold — it does not attempt to infer component layout before components exist.
- Batch create processes components sequentially within a single call but returns all results (success and error) together.
- The `isolated_component` suppression logic is project-wide: findings are suppressed entirely when `relationships.toml` contains zero entries across the project. Once any relationship exists, individual isolated components are flagged normally per FR-013.
- Slugification follows the existing algorithm already used for system/container/component IDs (lowercase, spaces to hyphens, special characters stripped).
- The `veve` CLI (binary: `veve`) is used exclusively for PDF export and is not invoked during D2 diagram writes triggered by relationship operations. Diagram file updates (FR-004) are direct filesystem writes, consistent with the existing `update_diagram` pattern.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After calling `create_relationship`, the relationship appears in `list_relationships` and is returned by `query_dependencies` in under 2 seconds, with no additional tool calls required. The graph cache is invalidated eagerly so no stale results are served.
- **SC-002**: An architect can document a system with 5 containers and 20 components, including all inter-container relationships, using 50% fewer tool calls than the current approach requires.
- **SC-003**: `validate` on a freshly initialized project (zero entries in `relationships.toml`) produces zero `isolated_component` findings — the isolation check is fully suppressed at the project-wide level.
- **SC-004**: When a tool call fails due to an unrecognized element name, the error message alone is sufficient for the user or agent to correct the call without querying the architecture — confirmed by user testing or agent replay without additional lookups.
- **SC-005**: Creating a container results in a valid, immediately usable diagram file without any additional tool calls — same observable behaviour as creating a system.
- **SC-006**: 100% of create tool responses include the generated slugified ID of the created element, verifiable by inspecting response payloads across all create operations.
