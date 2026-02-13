# Feature Specification: Serverless Architecture Template

**Feature Branch**: `003-serverless-template`
**Created**: 2026-02-05
**Status**: Draft
**Input**: User description: "003-serverless-template"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scaffold Serverless Project Structure (Priority: P1)

As a cloud architect documenting a serverless application, I want to scaffold C4 documentation with serverless-specific terminology and patterns so that my architecture documentation accurately reflects Lambda, API Gateway, and event-driven concepts instead of generic web server patterns.

**Why this priority**: The serverless template is a documented but missing feature (spec FR-007, README, ROADMAP). The `standard-3layer` template generates documentation with web server terminology (REST API, Database, Cache, Ports, gRPC) that doesn't match serverless architectures. This is the core deliverable.

**Independent Test**: User can scaffold a system, container, and component using the serverless template, and the generated markdown and D2 files contain serverless-appropriate terminology (Lambda functions, triggers, event sources) instead of generic web server patterns.

**Acceptance Scenarios**:

1. **Given** loko is installed with the serverless template, **When** user scaffolds a new system using the serverless template, **Then** system.md contains serverless sections (Event Sources, Functions, External Integrations) instead of generic sections (REST API, Database, Cache)
2. **Given** a serverless system exists, **When** user scaffolds a container using the serverless template, **Then** container.md includes trigger type, functions list, and IAM permissions instead of ports, gRPC services, and container type
3. **Given** a serverless container exists, **When** user scaffolds a component using the serverless template, **Then** component.md describes an individual Lambda function with handler, trigger, runtime, memory, and timeout metadata
4. **Given** a project scaffolded with the serverless template, **When** user runs `loko build`, **Then** documentation builds successfully with no errors
5. **Given** a project scaffolded with the serverless template, **When** user runs `loko validate`, **Then** validation passes with no errors

---

### User Story 2 - Template Selection (Priority: P1)

As a loko user, I want to choose between the standard-3layer and serverless templates when scaffolding entities so that I can use the appropriate template for my architecture style.

**Why this priority**: Currently the CLI hardcodes template paths to `standard-3layer` with no selection mechanism. Without a way to select the serverless template, it cannot be used. This is a prerequisite for the template being useful.

**Independent Test**: User can specify which template to use when running `loko new` commands, and the correct template files are applied.

**Acceptance Scenarios**:

1. **Given** both templates are installed, **When** user scaffolds an entity and selects the serverless template, **Then** the generated files use serverless template content
2. **Given** both templates are installed, **When** user scaffolds an entity without specifying a template, **Then** the standard-3layer template is used (backward compatible default)
3. **Given** user specifies a template that doesn't exist, **When** scaffolding, **Then** a clear error message is shown listing available templates

---

### User Story 3 - Event-Driven Diagram Patterns (Priority: P1)

As a serverless architect, I want D2 diagram templates that use event-driven patterns (async flows, cloud service icons, message queues) so my architecture diagrams accurately represent serverless communication patterns.

**Why this priority**: Diagrams are a core value proposition of loko. Serverless architectures use fundamentally different communication patterns (async, event-driven) than traditional architectures (synchronous request/response). Diagrams must reflect this or they misrepresent the architecture.

**Independent Test**: Generated D2 diagrams for serverless entities render correctly and show event-driven patterns with appropriate styling.

**Acceptance Scenarios**:

1. **Given** serverless system.d2 template, **When** rendered, **Then** shows Lambda functions, API Gateway, and event sources with cloud-native iconography
2. **Given** serverless container.d2 template, **When** rendered, **Then** uses dashed lines for async/event flows (not solid lines for synchronous calls)
3. **Given** serverless component.d2 template, **When** rendered, **Then** shows individual function with trigger source and downstream targets

---

### User Story 4 - Serverless Example Project (Priority: P2)

As a new loko user evaluating the tool for serverless documentation, I want a complete example project demonstrating serverless patterns so I can understand how to structure my own documentation.

**Why this priority**: The README already references `examples/serverless/` but it doesn't exist. An example project helps users understand the template output and serves as validation that the template works end-to-end.

**Independent Test**: The example project in `examples/serverless/` can be built with `loko build` and validated with `loko validate` successfully.

**Acceptance Scenarios**:

1. **Given** the serverless example project exists, **When** user runs `loko validate` in its directory, **Then** validation passes
2. **Given** the serverless example project exists, **When** user runs `loko build`, **Then** documentation builds successfully
3. **Given** the example project, **When** user reviews it, **Then** it demonstrates a realistic serverless architecture (e.g., order processing with API Gateway, Lambda functions, SQS queues, DynamoDB)

---

### Edge Cases

- What if user mixes serverless and standard-3layer templates in the same project? Templates should work alongside each other since they produce the same file types (system.md, container.md, component.md, .d2 files)
- What if the serverless template directory is missing from the installation? Clear error message should be shown, listing available templates
- What if user has a custom template directory via `LOKO_TEMPLATE_DIR`? Custom paths should take precedence over built-in templates

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST include a `serverless` template alongside the existing `standard-3layer` template
- **FR-002**: Serverless template MUST include all 6 template files: `system.md`, `system.d2`, `container.md`, `container.d2`, `component.md`, `component.d2`
- **FR-003**: All template files MUST use ason variable syntax (`{{VariableName}}`) consistent with the standard-3layer template
- **FR-004**: Serverless `system.md` MUST include sections for: Overview, Event Sources, Functions, External Integrations, and Technology Stack
- **FR-005**: Serverless `container.md` MUST include sections for: Purpose, Trigger Type, Functions List, and IAM Permissions
- **FR-006**: Serverless `component.md` MUST include metadata for: Handler, Trigger, Runtime, Memory, Timeout
- **FR-007**: Serverless D2 diagrams MUST use dashed lines for async/event-driven flows
- **FR-008**: Serverless D2 diagrams MUST use cloud-native iconography where appropriate
- **FR-009**: Users MUST be able to select which template to use when scaffolding entities
- **FR-010**: The default template MUST remain `standard-3layer` for backward compatibility
- **FR-011**: All scaffolded projects using the serverless template MUST pass `loko validate`
- **FR-012**: All scaffolded projects using the serverless template MUST build successfully with `loko build`
- **FR-013**: A serverless example project MUST exist at `examples/serverless/`

### Key Entities

- **Template**: A set of 6 files (3 markdown + 3 D2) that define the scaffolding output for a specific architecture style. Uses ason variable syntax for dynamic content substitution.
- **System** (serverless context): A serverless application composed of Lambda function groups, API Gateways, and event sources
- **Container** (serverless context): A logical grouping of related Lambda functions (e.g., API Handlers, Event Processors, Scheduled Tasks)
- **Component** (serverless context): An individual Lambda function with its configuration (handler, trigger, runtime, memory, timeout)
- **Event Source**: The trigger mechanism for a Lambda function (API Gateway, SQS, SNS, S3, EventBridge, CloudWatch Events)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 6 serverless template files are present and contain serverless-specific content (zero generic web server terminology like "REST API port", "gRPC service", "container type")
- **SC-002**: A project scaffolded entirely with the serverless template passes `loko validate` with 0 errors
- **SC-003**: A project scaffolded with the serverless template produces valid documentation via `loko build` with 0 errors
- **SC-004**: D2 diagrams from the serverless template render to valid SVG output
- **SC-005**: Users can select the serverless template when scaffolding without editing configuration files
- **SC-006**: Existing projects using the standard-3layer template continue to work with no changes (backward compatibility)
- **SC-007**: The serverless example project validates and builds successfully

## Scope & Exclusions

### In Scope

- 6 serverless template files in `templates/serverless/`
- Template selection mechanism for `loko new` commands
- Serverless example project in `examples/serverless/`
- AWS Lambda/API Gateway/event-driven patterns
- Updating README to accurately reflect available templates

### Out of Scope (Future)

- Azure Functions / Google Cloud Functions template variants
- Terraform/CDK integration or infrastructure-as-code generation
- Auto-detection of architecture style from existing code
- Cost estimation annotations in templates
- Step Functions / state machine orchestration patterns (may be added later)
- Template selection for `loko init` (project-level template config)

## Dependencies & Assumptions

### Dependencies

- Existing ason template engine (supports `{{Variable}}` substitution and multi-path search)
- Existing standard-3layer template as reference for structure and variable naming conventions
- D2 diagram tool for rendering serverless D2 templates

### Assumptions

- Users are familiar with AWS serverless concepts (Lambda, API Gateway, SQS, SNS, EventBridge)
- D2 can render dashed lines for async flows and supports external icon URLs for cloud service shapes
- The same ason variables used in standard-3layer (SystemName, Description, Technology, etc.) apply to serverless templates with equivalent semantics
- Template selection is per-entity (not per-project), allowing mixed template usage within a single project
