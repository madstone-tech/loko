# Feature Specification: TOON Output Format for MCP Query Tools

**Feature Branch**: `011-toon-mcp-output`
**Created**: 2026-06-03
**Status**: Draft
**Input**: User description: "Add TOON (Token-Oriented Object Notation, github.com/toon-format/toon-go) as the default output format for all MCP query tools to reduce LLM token consumption by 30-40% versus JSON. The toon-go dependency is already present in go.mod. All read tools (query_project, query_architecture, query_dependencies, query_related_components, search_elements, list_relationships, analyze_coupling) should default to TOON output but accept an optional format parameter to fall back to JSON for human-readable debugging. Mutation tools and validation tools continue to return JSON since their outputs are typically small. Include a benchmarking suite that measures token counts on representative payloads and gates the 30-40% reduction target as an acceptance criterion. Reference research/token-efficiency-benchmarks.md for the existing methodology."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - LLM receives compact query responses by default (Priority: P1)

An LLM agent connected to loko's MCP server calls a read tool (e.g. `query_architecture`) to understand a project. Without changing how it calls the tool, it receives the answer in a token-efficient notation that conveys the same information in materially fewer tokens, leaving more of its context budget for reasoning.

**Why this priority**: This is the entire point of the feature — the read tools are the highest-volume, most-repeated calls an LLM makes against loko, and their responses dominate token spend. Shrinking them is the single largest token-efficiency lever in the product, and it delivers value even if nothing else in this feature lands.

**Independent Test**: Call each of the seven read tools against the canonical test project with no `format` argument; confirm every response is returned in TOON and that the information content (every element, relationship, count, and field present in the JSON form) is preserved.

**Acceptance Scenarios**:

1. **Given** an LLM calls `query_project` with no format argument, **When** the tool returns, **Then** the response body is TOON-encoded and contains the same project, system, container, and component information the JSON form would have contained.
2. **Given** an LLM calls any of the seven read tools, **When** it inspects the response, **Then** the format is TOON by default without the caller having to opt in.
3. **Given** a representative project payload, **When** the same query is answered in TOON versus JSON, **Then** the TOON form uses at least 30% fewer tokens.

---

### User Story 2 - Developer requests JSON for human-readable debugging (Priority: P2)

A developer (or an MCP client that needs human-readable output) calls a read tool with an explicit `format: "json"` argument and receives the response as JSON, so they can read, diff, or paste it into existing JSON-aware tooling while debugging.

**Why this priority**: TOON is optimized for token density, not human reading. Without an explicit escape hatch, the default change would make interactive debugging and existing JSON-consuming integrations harder. This makes the new default safe to adopt.

**Independent Test**: Call each read tool with `format: "json"` and confirm the response is valid JSON equivalent in content to the TOON response; call with `format: "toon"` and confirm TOON; call with an unrecognized format value and confirm a clear error rather than a silent fallback.

**Acceptance Scenarios**:

1. **Given** a read tool is called with `format: "json"`, **When** it returns, **Then** the response is valid JSON carrying the same information as the default TOON response.
2. **Given** a read tool is called with `format: "toon"` explicitly, **When** it returns, **Then** the response is TOON (identical to the default).
3. **Given** a read tool is called with an unsupported `format` value, **When** it is processed, **Then** the caller receives a clear error naming the offending value and the allowed values.
4. **Given** a mutation tool (e.g. `create_system`) or a validation tool, **When** it returns, **Then** the response is JSON regardless of any `format` argument, and its behavior is unchanged from today.

---

### User Story 3 - Maintainer proves and guards the token-reduction target (Priority: P3)

A maintainer runs a repeatable benchmark that measures token counts for TOON versus JSON across representative payloads and reports the percentage reduction, so the 30–40% target can be verified before release and protected against regression over time.

**Why this priority**: The 30–40% reduction is the feature's headline claim and its acceptance criterion. A reproducible, reviewable benchmark turns that claim from an assertion into an enforced gate, and it fills in the currently-empty results table in the existing methodology doc.

**Independent Test**: Run the benchmark suite against the representative payload set; confirm it emits per-payload JSON-token, TOON-token, and percentage-reduction figures, and that it fails (non-zero) if the aggregate reduction across the read-tool payloads drops below the agreed threshold.

**Acceptance Scenarios**:

1. **Given** the benchmark suite is run, **When** it completes, **Then** it reports JSON tokens, TOON tokens, and percentage reduction for each representative payload and an aggregate figure.
2. **Given** the aggregate reduction is below the minimum acceptance threshold, **When** the benchmark runs in the gating context, **Then** it signals failure.
3. **Given** the benchmark has run, **When** a maintainer opens the methodology results document, **Then** the previously-empty results table is populated with the measured figures.

---

### Edge Cases

- **Empty or minimal payloads** (e.g. `query_project` on a project with no systems): TOON must still be produced and parseable; very small payloads may show little or no reduction — these are excluded from the gated aggregate (the gate is about representative, non-trivial payloads).
- **Unsupported `format` value**: returns a clear error, never a silent default to the wrong format.
- **Special characters / nesting** in element names, descriptions, narratives, or tags: TOON output must remain well-formed and unambiguously decodable.
- **A read tool whose payload is dominated by long free-text** (e.g. narratives): reduction may be lower than for list-heavy payloads; the suite reports per-payload figures so this is visible rather than hidden in the aggregate.
- **Existing MCP clients that assumed JSON**: behavior changes by default (now TOON); they must opt back into JSON via `format: "json"`. This is an intentional, documented change for read tools (see Assumptions).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All seven read tools — `query_project`, `query_architecture`, `query_dependencies`, `query_related_components`, `search_elements`, `list_relationships`, `analyze_coupling` — MUST return their responses in TOON when no format is specified.
- **FR-002**: Each read tool MUST accept an optional `format` argument with the allowed values `toon` (default) and `json`.
- **FR-003**: When `format: "json"` is supplied, a read tool MUST return valid JSON whose information content is equivalent to its TOON response (same elements, relationships, counts, and fields).
- **FR-004**: When an unsupported `format` value is supplied, the tool MUST return a clear error that names the rejected value and lists the allowed values, and MUST NOT silently substitute a format.
- **FR-005**: Mutation tools (create/update/delete) and validation tools MUST continue to return JSON, and their outputs MUST be unchanged by this feature regardless of any `format` argument.
- **FR-006**: TOON responses MUST preserve the full information of the equivalent JSON response — no field, element, relationship, or count may be dropped solely because of the format.
- **FR-007**: TOON responses MUST be well-formed and unambiguously decodable for all valid model content, including names, descriptions, narratives, and tags containing special characters or nested structure.
- **FR-008**: Each read tool's declared input schema MUST document the `format` argument, its allowed values, and its default, so MCP clients and LLMs can discover it.
- **FR-009**: A benchmark suite MUST measure and report, per representative payload, the JSON token count, the TOON token count, and the percentage reduction, plus an aggregate across the read-tool payloads, using the methodology documented in `research/token-efficiency-benchmarks.md`.
- **FR-010**: The benchmark suite MUST signal failure when the aggregate token reduction across representative read-tool payloads falls below the minimum acceptance threshold (see Success Criteria), so it can act as a release/CI gate.
- **FR-011**: Running the benchmark MUST populate the results table in `research/token-efficiency-benchmarks.md` (currently empty) with the measured figures.
- **FR-012**: The change MUST be reversible per call — any client can obtain JSON for any read tool by passing `format: "json"` without server reconfiguration.

### Key Entities *(include if feature involves data)*

- **Tool response payload**: the data a read tool returns about the architecture model (project, systems, containers, components, relationships, dependency/coupling reports, search results). Format-independent in content; rendered as either TOON or JSON.
- **Format selector**: the per-call choice of output notation (`toon` or `json`), defaulting to `toon`.
- **Benchmark payload set**: the representative model payloads used to measure token efficiency, drawn from the example projects already referenced by the existing methodology and the canonical test project.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Across the representative read-tool payloads, TOON responses use **at least 30%** fewer tokens than the equivalent JSON responses (target band 30–40%), measured by the documented methodology.
- **SC-002**: All seven named read tools return TOON by default and JSON on request; 100% of them support both formats.
- **SC-003**: Mutation and validation tool outputs are unchanged from before the feature (0 regressions in their responses).
- **SC-004**: An invalid `format` value yields an actionable error in 100% of cases, with no silent format substitution.
- **SC-005**: The benchmark suite is reproducible (same inputs produce the same reported reduction figures) and is wired into the project's automated gate so a drop below the threshold fails the build.
- **SC-006**: The results table in the methodology document is populated with current measured figures rather than empty.
- **SC-007**: A maintainer can verify the headline reduction claim end-to-end by running a single benchmark command and reading its output, without inspecting implementation internals.

## Assumptions

- **The `format` argument is a per-call tool parameter**, not a session-level or environment-level setting. This matches how MCP tools expose options and lets a single client mix TOON and JSON across calls. (Candidate for `/speckit.clarify` if a global/session toggle is preferred.)
- **TOON becomes the default for read tools, which is an intentional behavior change** for existing MCP clients; clients that require JSON opt in via `format: "json"`. The feature description states this explicitly.
- **A TOON encoding capability already exists** in the project (an output-encoder abstraction with a TOON implementation, already used by the build pipeline). This feature applies it to MCP read-tool responses; it does not introduce a new TOON library (`toon-go` is already a dependency).
- **The representative benchmark payload set** is the example projects referenced by `research/token-efficiency-benchmarks.md` plus the canonical test project, exercised through the seven read tools. (Exact dataset composition is a `/speckit.clarify` candidate.)
- **Token counting uses the approximation already documented** in the methodology doc (≈4 characters per token for structured data) unless a more precise tokenizer is later chosen; the gate is expressed as a percentage reduction so it is robust to the estimator choice.
- **The 30–40% figure is a target band**; the hard gate is the lower bound (≥30%) to avoid flapping when a payload happens to land at the band's edge.
- **Mutation/validation tools are out of scope** for format changes because their payloads are small and their JSON outputs are relied upon elsewhere.
- **No change to the MCP transport, tool names, or call signatures** beyond adding the optional `format` argument to the seven read tools.
