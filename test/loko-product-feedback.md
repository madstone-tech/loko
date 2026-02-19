# Loko Product Feedback (Non-MCP)

**Date:** 2026-02-13
**Context:** Modeled a serverless Notification Service — 1 system, 6 containers, 17 components (mix of Lambda functions, DynamoDB tables, SQS queues, API Gateway endpoints, EventBridge rules, SES/SNS services)

---

## D2 Diagram Experience

**Positive:** D2 is a great choice — text-based, diffable, version-controllable. The C4 model alignment (system → container → component) gives clear structure.

**Gap:** There's no diagram rendering in the workflow. I authored 7 D2 diagrams but never saw what they look like. `build_docs` generates HTML, but it's unclear whether diagrams are rendered to SVG/PNG or shipped as raw D2 source. A user would need to separately install and run the `d2` CLI to visualize their architecture. Consider embedding rendering in `build_docs` or providing a `render_diagram` tool.

---

## File Structure

**Positive:** The `src/{system}/{container}/{component}/` hierarchy is intuitive and maps cleanly to C4 levels.

**Concern:** Every component gets both a `component.md` and a `{component-name}.d2`. For 17 components, that's 34 files plus container and system files. This is fine for a real project that grows over time, but the ratio of boilerplate to actual content in a freshly-scaffolded project is high.

---

## Templates: The Core Problem

The current templates are **code-component templates applied universally**. Every `component.md` gets the same sections regardless of what the component actually is:

```markdown
## Interfaces
### Public Methods
- `Method1()` - Description of method 1

## Implementation Details
### Key Classes/Functions
- `Class1` - Description

### Data Structures
- (List important data structures)

## Testing
- Unit tests: (specify framework)
- Integration tests: (specify framework)
- Coverage: (target %)

## Performance Considerations
- (Note any performance-critical aspects)
```

This makes sense for a Go Lambda function. It makes no sense for a DynamoDB table, an SQS queue, an EventBridge rule, or an API Gateway endpoint. These are infrastructure resources, not code units. They don't have "Public Methods" or "Unit tests."

The same issue exists at the container level — `container.md` includes sections like "Port: (e.g., 8080)" and "Health checks: /health" which are meaningless for a DynamoDB data store or an SQS message queue layer.

### What Would Be More Useful

**1. Technology-aware templates**

Loko already captures `technology` in the component metadata. Use it to select an appropriate template:

| Technology pattern | Template type | Key sections |
|---|---|---|
| `Lambda`, `Function` | Compute | Trigger, Runtime, Timeout, Memory, Environment vars, Error handling |
| `DynamoDB`, `Database`, `Table` | Data store | Key schema, GSIs/LSIs, Capacity mode, TTL, Access patterns |
| `SQS`, `Queue`, `SNS Topic` | Messaging | Queue type, Visibility timeout, Retention, DLQ config, Message format |
| `API Gateway`, `REST`, `Endpoint` | API | Method, Path, Auth, Request/Response schema, Rate limits |
| `EventBridge`, `Event` | Event | Bus name, Event pattern, Rule targets, Input transformation |
| `S3`, `Bucket` | Storage | Bucket policy, Lifecycle rules, Versioning, Encryption |
| Default / unknown | Generic | Description, Responsibilities, Dependencies (minimal) |

**2. Prompt-based sections instead of fixed sections**

Instead of dictating structure, use short prompts that guide the author toward what matters:

```markdown
# Email Queue

Standard SQS queue for email notifications.

<!-- What configuration choices matter for this component? -->

<!-- What does this component connect to, and how? -->

<!-- What failure modes should operators know about? -->
```

This works for any component type. An SQS queue author writes about visibility timeout and DLQ policy. A Lambda author writes about cold starts and concurrency. A DynamoDB table author writes about key design and access patterns. The prompt steers without dictating.

**3. Minimal defaults with opt-in depth**

Start with the bare minimum — just the frontmatter and a one-liner description:

```markdown
---
id: email-queue
name: "Email Queue"
description: "Standard SQS queue for email notifications"
technology: "Amazon SQS Standard"
tags:
  - "messaging"
---

# Email Queue

Standard SQS queue for email notifications.
```

Then provide a command like `loko enrich <component>` that asks targeted questions based on the technology and adds relevant sections. This way the scaffolded project is clean, and detail is added intentionally rather than by template.

**4. Don't duplicate what the D2 file already expresses**

The component `.md` currently has sections for "Dependencies" and "Interfaces" — but these are exactly what the D2 diagram models with arrows. If the D2 file says `email-queue -> email-sender: "triggers"`, the markdown shouldn't also need a manually-maintained dependencies list. Either generate the dependencies section from D2, or drop it from the template.

**5. Container templates should reflect their purpose**

A "Processing Layer" container running Lambda functions should surface different information than a "Data Store" container running DynamoDB. Suggested container template variants:

- **Compute container:** Functions list, runtime, IAM role, VPC config, scaling
- **Data container:** Tables/collections, key design, capacity, backup, TTL
- **Messaging container:** Queues/topics, throughput, DLQ config, ordering guarantees
- **API container:** Endpoints, auth method, rate limits, CORS, stages
- **Event container:** Bus, rules, patterns, targets, retry policy

### The Principle

Templates should **reduce friction for the author**, not create homework. The best template is one where every section feels relevant to fill in. If a section consistently gets left as placeholder text, it shouldn't be in the template.

A good test: after scaffolding a component, can someone fill in every section without thinking "this doesn't apply to me"? If not, the template is too generic.

---

## Source of Truth Ambiguity

There's no clear hierarchy for where information lives:

| Information | `.md` frontmatter | `.md` body | D2 file |
|---|---|---|---|
| Name | `name` field | `# heading` | node label |
| Description | `description` field | prose paragraph | `tooltip` |
| Relationships | — | Dependencies section | arrows (`->`) |
| Technology | `technology` field | Technology Stack section | comment header |

When these diverge (and over time they will), which one is authoritative? Suggestions:
- **Frontmatter is the source of truth** for metadata (name, description, technology, tags)
- **D2 files are the source of truth** for relationships and visual layout
- **Markdown body is free-form documentation** — not parsed, not validated
- Make this explicit in the docs and enforce it in `validate`

---

## Auto-Generation Gaps

Several things loko knows but doesn't reflect in generated files:

1. **Container component tables** — `container.md` has `(Add your components here)` but loko knows every component in the container from the filesystem. This table should auto-populate.

2. **System container tables** — same issue in `system.md` with `(Add your containers here)`.

3. **Cross-references in docs** — `build_docs` could link component pages to their container, link containers to their system, and generate a dependency graph from D2 relationships.

---

## Missing: Change Detection / Drift

After editing 17 D2 files and 0 `.md` files, the descriptions in D2 tooltips and `.md` frontmatter could drift apart. There's no way to ask:
- "What changed since last validation?"
- "Are my diagrams consistent with my metadata?"
- "Which components have stale documentation?"

A `lint` or `drift` command comparing diagram content against metadata would catch inconsistencies early — for example, a D2 file that references a component that doesn't exist, or a component whose D2 description doesn't match its `.md` description.

---

## Summary

Loko's core model (C4-aligned, filesystem-based, D2 diagrams) is strong. The main improvement area is **context-awareness** — templates, validation, and documentation generation should adapt to what kind of component they're dealing with, rather than applying a one-size-fits-all code-component lens to infrastructure resources.
