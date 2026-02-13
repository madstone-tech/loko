# Feature Specification: Graph Implementation Improvements

**Feature Branch**: `001-graph-implementation`  
**Created**: 2026-02-12  
**Status**: Draft  
**Input**: User description: "Comprehensive review and fixes for ArchitectureGraph implementation addressing correctness bugs, performance optimizations, and type safety improvements across P0-P3 priority levels."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Multi-System Projects Work Correctly (Priority: P0)

When a developer works on a project with multiple systems that have components with similar names (e.g., "Auth" component in both "Backend" and "Admin" systems), the architecture graph should correctly represent all components without silent failures or data loss.

**Why this priority**: This is a correctness bug that silently corrupts the graph, leading to incorrect validation results, missing relationships in MCP queries, and unreliable coupling analysis. This affects the core functionality of loko.

**Independent Test**: Can be fully tested by creating a project with two systems, each containing a component with the same name, then verifying that both components appear in the graph with their relationships intact.

**Acceptance Scenarios**:

1. **Given** a project with System A containing Component "auth" and System B containing Component "auth", **When** the architecture graph is built, **Then** both components appear as distinct nodes in the graph with unique identifiers
2. **Given** two components with the same name in different systems each having relationships, **When** querying dependencies via MCP tools, **Then** both components' relationships are correctly reported
3. **Given** a graph with duplicate component names across systems, **When** running validation checks, **Then** no components are silently excluded from validation

---

### User Story 2 - Fast Dependency Queries (Priority: P1)

When a developer queries component dependencies repeatedly during an interactive MCP session (e.g., analyzing coupling, finding isolated components), the system should respond quickly without performance degradation, even for projects with hundreds of components.

**Why this priority**: Performance issues make the tool frustrating to use and block adoption for larger projects. The current O(E) and O(N) lookups compound in validation loops.

**Independent Test**: Can be tested by creating a project with 200+ components, running dependency analysis, and measuring query response time (should be under 100ms per query).

**Acceptance Scenarios**:

1. **Given** a graph with 200 components, **When** querying incoming edges for a component, **Then** the query completes in under 50ms
2. **Given** a graph with deep hierarchies (5+ levels), **When** retrieving all descendants of a container, **Then** the query completes in under 100ms
3. **Given** validation running on a large project, **When** checking for isolated components and high coupling, **Then** the entire validation completes in under 2 seconds

---

### User Story 3 - MCP Sessions Remain Responsive (Priority: P2)

When an LLM agent interacts with the architecture via multiple MCP tool calls in a conversation (e.g., "show me dependencies of X", "analyze coupling", "find isolated components"), the session should remain fast and responsive even after dozens of queries.

**Why this priority**: Repeated graph rebuilding wastes computation and creates a poor user experience during LLM-assisted architecture analysis sessions.

**Independent Test**: Can be tested by making 50 consecutive MCP tool calls on the same project and verifying that the 50th call is as fast as the 1st call.

**Acceptance Scenarios**:

1. **Given** an MCP session with a loaded project, **When** making 50 consecutive dependency queries without file changes, **Then** the average query time remains constant (within 10% variance)
2. **Given** a cached graph and a file change in the src/ directory, **When** making a new MCP query, **Then** the cache is invalidated and graph is rebuilt
3. **Given** multiple projects open in different MCP sessions, **When** querying different projects, **Then** each project's cache is isolated and correct

---

### User Story 4 - Type-Safe Graph Operations (Priority: P2)

When a developer extends the graph functionality or modifies use cases that consume the graph, the compiler should catch type errors and prevent runtime surprises from incorrect data types.

**Why this priority**: Type safety prevents bugs and makes the codebase more maintainable. The current use of `any` and `map[string]any` erodes type safety and makes refactoring risky.

**Independent Test**: Can be tested by attempting to access graph data with incorrect types and verifying that compilation fails (not runtime).

**Acceptance Scenarios**:

1. **Given** a dependency analysis report, **When** a developer accesses isolated components, **Then** the type system guarantees it's a string slice without runtime assertions
2. **Given** a graph node containing entity data, **When** a developer accesses the entity, **Then** the type system provides discoverable methods without type switches
3. **Given** an MCP tool with input arguments, **When** the tool is invoked with incorrect argument types, **Then** validation fails early with a clear type error

---

### User Story 5 - Clear Architecture Documentation (Priority: P3)

When a new contributor reads the codebase or an AI agent analyzes the architecture, they should understand the key design decisions around graph construction, node ID format, and lifecycle without digging through implementation code.

**Why this priority**: Implicit conventions make the codebase harder to understand and increase the risk of introducing bugs when making changes.

**Independent Test**: Can be tested by having a developer unfamiliar with the codebase answer questions like "why don't system dependencies appear as graph edges?" using only the documentation.

**Acceptance Scenarios**:

1. **Given** an ADR document for graph conventions, **When** a developer reads it, **Then** they understand how node IDs are constructed and why
2. **Given** documentation on graph lifecycle, **When** a developer needs to add caching, **Then** they know when the graph is valid vs. stale
3. **Given** documentation on relationship scope, **When** a developer wonders why systems don't have edges, **Then** the ADR explains this design choice

---

### Edge Cases

- What happens when a component references another component that doesn't exist (broken relationship)?
- How does the system handle circular dependencies in component relationships?
- What happens when two components in the same container have the same name (collision within a container)?
- How does the system behave when the graph is queried during concurrent builds (race conditions)?
- What happens when a component has relationships to components in different systems?
- How does caching handle rapid file changes (e.g., during active development)?
- What happens when duplicate edges are added through different code paths?
- How does the system handle removing a node that other nodes depend on?

## Requirements *(mandatory)*

### Functional Requirements

#### P0 - Node ID Collision Fix

- **FR-001**: System MUST generate unique node IDs for all components across all systems by including the full hierarchy path (system/container/component)
- **FR-002**: System MUST generate unique node IDs for all containers across all systems by including the system path (system/container)
- **FR-003**: System MUST maintain a mapping from short IDs to qualified IDs to support existing relationship references
- **FR-004**: System MUST report errors when duplicate node IDs are detected instead of silently ignoring them
- **FR-005**: Graph builder MUST successfully add all components from all systems to the graph without silent failures

#### P1 - Performance Optimizations

- **FR-006**: System MUST maintain a reverse adjacency map (IncomingEdges) that provides O(1) lookup of incoming edges for any node
- **FR-007**: System MUST update the IncomingEdges map whenever an edge is added to maintain consistency
- **FR-008**: System MUST maintain a children map (ChildrenMap) that provides O(1) lookup of child nodes for any parent
- **FR-009**: System MUST update the ChildrenMap whenever a node with a parent is added
- **FR-010**: GetIncomingEdges MUST execute in O(1) time by returning the pre-indexed incoming edges
- **FR-011**: GetChildren MUST execute in O(1) time by returning the pre-indexed children list

#### P2 - Duplicate Edge Prevention & Graph Caching

- **FR-012**: System MUST check for existing edges before adding new edges to prevent duplicates
- **FR-013**: System MUST consider two edges identical if they have the same source, target, and type
- **FR-014**: MCP server MUST cache built graphs keyed by project root path
- **FR-015**: MCP server MUST invalidate cached graphs when files in the src/ directory change
- **FR-016**: System MUST provide RemoveNode method that removes a node and all edges referencing it
- **FR-017**: System MUST provide RemoveEdge method that removes edges from both Edges and IncomingEdges maps

#### P2 - Type Safety Improvements

- **FR-018**: System MUST define a DependencyReport struct to replace map[string]any return type from AnalyzeDependencies
- **FR-019**: System MUST define input argument structs for each MCP tool with JSON schema tags
- **FR-020**: System MUST generate JSON schemas from struct tags to keep schema and parsing in sync
- **FR-021**: System MUST deserialize MCP tool arguments into typed structs instead of using type assertions
- **FR-022**: System MUST define a C4Entity interface that System, Container, and Component implement
- **FR-023**: GraphNode.Data MUST be typed as C4Entity instead of any to enable compile-time type checking

#### P3 - Quality Improvements

- **FR-024**: Validation checks MUST filter nodes to only include components (exclude systems and containers)
- **FR-025**: System MUST document thread safety requirements for graph operations
- **FR-026**: System MUST provide an ADR document covering node ID conventions, graph lifecycle, and relationship scope
- **FR-027**: System MUST document which operations are safe for concurrent access and which require synchronization

### Key Entities

- **ArchitectureGraph**: Represents the directed graph of C4 model elements with nodes (systems, containers, components) and edges (relationships). Contains maps for nodes, outgoing edges, incoming edges, parent relationships, and children relationships.

- **GraphNode**: Represents a node in the architecture graph. Contains an ID (qualified path), type (system/container/component), name, optional parent ID, and entity data conforming to C4Entity interface.

- **GraphEdge**: Represents a directed relationship between two nodes. Contains source ID, target ID, relationship type, and optional description.

- **DependencyReport**: Structured report of dependency analysis results. Contains counts (systems, containers, components, nodes, edges), isolated component IDs, highly coupled component metrics, and central component metrics.

- **C4Entity**: Interface implemented by System, Container, and Component to provide common entity operations (ID, name, type) without type assertions.

- **CachedGraph**: Represents a cached architecture graph with build timestamp for cache invalidation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Projects with multiple systems containing identically-named components produce complete graphs with all components represented (100% component inclusion rate)

- **SC-002**: Dependency queries on graphs with 200+ components complete in under 50ms per query (2x faster than current O(E) implementation)

- **SC-003**: Validation checks on large projects (100+ components) complete in under 2 seconds

- **SC-004**: MCP tool response times remain constant (within 10% variance) across 50 consecutive queries when project files haven't changed

- **SC-005**: Graph operations compile with full type checking - no runtime type assertions in graph consumer code

- **SC-006**: Duplicate edge detection prevents edge count inflation - EdgeCount() returns accurate count even when AddEdge is called multiple times with identical edges

- **SC-007**: New contributors can answer architecture questions (node ID format, graph lifecycle, relationship scope) using only ADR documentation without reading implementation code

- **SC-008**: Cache invalidation responds to file changes within 1 second - MCP queries reflect updated architecture immediately after src/ file modifications

### Assumptions

- **Graph size**: Typical C4 models contain fewer than 500 nodes and 1000 edges (per Constitution design constraints)
- **Concurrency model**: Graph is built sequentially, then queried concurrently by multiple MCP tool calls (read-heavy workload)
- **Component naming**: Component names are not guaranteed to be unique across systems or containers (hence the need for qualified IDs)
- **Relationship storage**: Component relationships are currently stored as flat ID strings in Component.Relationships map
- **Cache invalidation**: File system watching is available at the MCP server level to detect src/ directory changes
- **Type safety goals**: Balance between full type safety and Go's pragmatic approach - interface{} is acceptable for JSON serialization boundaries
- **Performance targets**: Based on typical LLM interaction patterns (10-50 queries per session) and human perception thresholds (sub-100ms feels instant)
- **Documentation location**: ADRs belong in docs/ directory alongside existing architectural documentation
- **Thread safety approach**: Prefer documentation and careful usage patterns over complex locking (given read-heavy workload and single-threaded build)
