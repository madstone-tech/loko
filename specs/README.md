# loko Specifications Directory

This directory contains all feature specifications using the structured spec workflow.

## Structure

Each feature gets its own directory with all planning documents:

```
specs/
└── 001-loko-v0.1.0/           # Feature ID and name
    ├── spec.md                 # User stories and requirements
    ├── plan.md                 # (in .specify/memory)
    ├── research.md             # Technology decisions and risks
    ├── data-model.md           # Entity definitions and validation
    ├── contracts/              # API/tool specifications
    │   └── mcp-tools.md        # MCP tool contracts
    └── quickstart.md           # Test scenarios and acceptance tests
```

## Document Purposes

| Document | Purpose | Status |
|----------|---------|--------|
| **spec.md** | User stories, acceptance criteria, functional/non-functional requirements | ✅ Ready |
| **research.md** | Technology choices, architecture decisions, risks, dependencies | ✅ Ready |
| **data-model.md** | Entity definitions, relationships, validation rules, file format | ✅ Ready |
| **contracts/mcp-tools.md** | MCP tool input/output schemas, error handling | ✅ Ready |
| **quickstart.md** | 5-minute walkthrough, acceptance test scenarios, benchmarks | ✅ Ready |

## Reading Order

1. **Start here**: `spec.md` - Understand what we're building
2. **Why**: `research.md` - Understand the decisions made
3. **How**: `data-model.md` - Understand the data structures
4. **Interfaces**: `contracts/mcp-tools.md` - Understand the APIs
5. **Verify**: `quickstart.md` - How to test everything works

## Feature: loko v0.1.0

**Status**: Ready for Implementation  
**Version**: 0.1.0-dev  
**Priority**: P1 (Core MVP)

### What's Being Built

A C4 model architecture documentation tool with:
- 🤖 LLM-driven architecture design via MCP
- 📝 Direct file editing with watch mode
- 🚀 Fast project scaffolding
- 📊 Token-efficient architecture queries
- 🌐 HTTP API for CI/CD (foundation)
- 📄 Multi-format export (HTML, Markdown, PDF)

### 6 User Stories

1. **US-1** (P1): LLM-Driven Architecture Design
2. **US-2** (P1): Direct File Editing Workflow
3. **US-3** (P1): Project Scaffolding
4. **US-4** (P2): API Integration
5. **US-5** (P2): Multi-Format Export
6. **US-6** (P1): Token-Efficient Architecture Queries

### Success Criteria

- ✅ 100 diagrams built in <30 seconds
- ✅ Watch mode rebuilds in <500ms
- ✅ Memory usage <100MB (50 systems)
- ✅ Architecture queries <500 tokens
- ✅ Time from init to docs <2 minutes
- ✅ Claude can design architecture via MCP without human intervention

---

## How to Use These Specs

### For Implementation

1. Read `spec.md` to understand requirements
2. Reference `data-model.md` when writing entities
3. Reference `contracts/mcp-tools.md` when implementing MCP tools
4. Reference `research.md` for technical decisions
5. Use `quickstart.md` for acceptance test scenarios

### For Testing/QA

1. Follow `quickstart.md` 5-minute walkthrough to verify MVP works
2. Run acceptance test scenarios from `quickstart.md`
3. Benchmark performance targets (NFR-001, NFR-002, NFR-003)
4. Validate against success criteria

### For Code Review

1. Check that implementation follows `data-model.md` entity definitions
2. Check that API/MCP tools follow `contracts/mcp-tools.md` schemas
3. Verify that test scenarios from `quickstart.md` pass
4. Check architecture decisions from `research.md` are followed

---

## Navigation Quick Links

### By Concern

**User Stories & Requirements** → `spec.md`  
**Technology & Architecture** → `research.md`  
**Data Structures** → `data-model.md`  
**API Contracts** → `contracts/mcp-tools.md`  
**Testing & Validation** → `quickstart.md`

### By Role

**Product Owner** → `spec.md` (User Stories, Success Criteria)  
**Architect** → `research.md` (Technical decisions, risks)  
**Engineer** → `data-model.md` + `contracts/mcp-tools.md` (Implementation)  
**QA** → `quickstart.md` (Acceptance tests, benchmarks)

---

## Key Numbers

- **6 User Stories** (3 P1, 1 P1, 1 P1, 1 P2, 1 P2, 1 P1)
- **24 Functional Requirements** (FR-001 through FR-024)
- **11 Non-Functional Requirements** (NFR-001 through NFR-011)
- **10 Success Criteria** (SC-001 through SC-010)
- **8 MCP Tools** defined in contracts
- **5 Core Entities** (Project, System, Container, Component, Diagram)
- **5 Implementation Slices** (each deliverable independently)

---

## Next Steps

1. **Read the specs** - Start with `spec.md`
2. **Review architecture** - Check `research.md` for decisions
3. **Plan implementation** - Use `.specify/memory/plan.md` and `tasks.md`
4. **Code** - Reference `data-model.md` for entities, `contracts/mcp-tools.md` for APIs
5. **Test** - Follow `quickstart.md` acceptance tests

---

## 📢 Public Repository Note

This `specs/` directory is **publicly tracked** in git. It contains **technical specifications only**.

**Strategic/business content** (GTM, pricing, market analysis) belongs in the **`research/`** directory (gitignored).

See **`CONTRIBUTING_TO_SPECS.md`** for guidelines on maintaining this separation.

---

**Last Updated**: 2026-02-13  
**Maintainer**: MADSTONE TECHNOLOGY  
**Reference**: `.specify/memory/constitution.md`, `CONTRIBUTING_TO_SPECS.md`
