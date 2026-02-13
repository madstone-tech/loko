# Contributing to Feature Specifications

This guide explains how to maintain the `specs/` directory for a **public-facing repository**.

## 📂 Directory Purpose

The `specs/` directory contains **technical feature specifications** and is **publicly tracked** in git. This demonstrates our engineering rigor and helps the community understand our development process.

---

## ✅ What Belongs in `specs/`

**Public Technical Content** (safe to share):
- User stories and acceptance criteria
- Technical specifications and requirements
- Data models and entity definitions
- API/tool contracts and schemas
- Implementation plans and tasks
- Architecture Decision Records (ADRs)
- Migration guides
- Test scenarios and benchmarks
- Code structure and package layout

**Example**: Feature 001's spec files are excellent examples of public technical content.

---

## ❌ What Belongs in `research/` (Gitignored)

**Private Strategic Content** (NOT for public):
- Go-to-market (GTM) strategies
- Market analysis and competitive research
- Revenue models and pricing strategies
- Customer acquisition plans
- Business roadmaps and timelines
- Partnership discussions
- Internal metrics and KPIs
- Proprietary methodologies

**Location**: `/research/` directory (already in `.gitignore`)

---

## 🔍 Review Checklist Before Committing Specs

Before committing any spec file, ask yourself:

### 1. **Does it contain business strategy?**
   - ❌ If yes → Move to `research/`
   - ✅ If no → Safe for `specs/`

### 2. **Does it reveal competitive advantages?**
   - ❌ If yes → Move to `research/`
   - ✅ If no → Safe for `specs/`

### 3. **Does it discuss pricing or revenue?**
   - ❌ If yes → Move to `research/`
   - ✅ If no → Safe for `specs/`

### 4. **Is it purely technical (code, architecture, tests)?**
   - ✅ If yes → Safe for `specs/`
   - ❌ If no → Review above questions

### 5. **Would you be comfortable with competitors reading it?**
   - ✅ If yes → Safe for `specs/`
   - ❌ If no → Move to `research/`

---

## 📝 Spec Document Templates

Use these templates from `.specify/memory/`:

**For `specs/` (Public)**:
- `spec.md` - User stories, requirements, acceptance criteria
- `plan.md` - Technical implementation plan
- `research.md` - Technology choices and architecture decisions (technical only)
- `data-model.md` - Entity definitions and validation
- `contracts/*.md` - API/tool schemas
- `tasks.md` - Implementation task list
- `quickstart.md` - Test scenarios and acceptance tests

**For `research/` (Private)**:
- `gtm-strategy.md` - Go-to-market planning
- `market-analysis.md` - Competitive landscape
- `business-roadmap.md` - Strategic milestones
- `metrics-kpis.md` - Business metrics

---

## 🚀 Example: Feature Spec Split

### ✅ Public (`specs/001-new-feature/`)
```markdown
# spec.md

## User Story
As a developer, I want to export architecture to PDF
so that I can share documentation offline.

## Technical Requirements
- Use veve-cli for PDF generation
- Support A4 and Letter paper sizes
- Include diagrams as embedded images
- Maintain hyperlink references
```

### ❌ Private (`research/feature-001-business-case/`)
```markdown
# business-case.md

## Market Opportunity
- 60% of enterprise users request PDF export
- Willing to pay $X/month premium tier
- Competitors charge $Y/year for similar feature
- Target: 1,000 paid conversions in Q2

## Revenue Impact
- Expected MRR: $Z
- Customer acquisition cost: $W
```

---

## 🔄 Workflow

1. **Start Feature**:
   ```bash
   # Create spec in public directory
   mkdir specs/00X-feature-name
   
   # Create business docs in private directory
   mkdir research/feature-00X-business
   ```

2. **Write Specs**:
   - Technical content → `specs/00X-feature-name/`
   - Business content → `research/feature-00X-business/`

3. **Review Before Commit**:
   ```bash
   # Check what you're about to commit
   git diff specs/
   
   # Run the checklist (see above)
   # Remove any business/strategic content
   ```

4. **Commit**:
   ```bash
   git add specs/00X-feature-name/
   git commit -m "Add spec for Feature 00X: <name>"
   
   # Business docs are gitignored automatically
   ```

---

## 🛡️ Sensitive Term Scanner

Before committing, scan for sensitive terms:

```bash
# Check for business-sensitive keywords
grep -r "GTM\|revenue\|pricing\|competitive\|market share\|acquisition cost" specs/00X-feature-name/

# If any matches found, review and move to research/
```

---

## 📊 Public Benefits

Keeping technical specs public provides:

1. **Transparency** - Shows rigorous development process
2. **Community Trust** - Open-source project credibility
3. **Contributor Onboarding** - Easy to understand feature design
4. **Historical Context** - Why decisions were made
5. **Educational Value** - Other projects learn from our methodology

---

## ❓ FAQ

**Q: Can I reference business metrics in technical specs?**  
A: Only generic targets (e.g., "support 10,000 users"), not specific revenue/pricing.

**Q: What about roadmap timelines?**  
A: Technical milestones are OK (e.g., "Phase 1: Q1 2025"). Business KPIs are not (e.g., "Target: 500 paid users by Q2").

**Q: Can I mention competitor features?**  
A: Yes, if discussing technical implementation (e.g., "Similar to GitLab's approach"). No, if discussing competitive positioning (e.g., "We're 2x faster than X for $Y less").

**Q: What about internal tool names?**  
A: Generic names are fine (e.g., "internal MCP server"). Proprietary codenames are not (e.g., "Project Falcon").

---

## 📞 Questions?

If unsure whether content belongs in `specs/` or `research/`, ask:
- "Would this help an open-source contributor understand the feature?"
  - ✅ Yes → `specs/`
  - ❌ No → `research/`

---

**Last Updated**: 2026-02-13  
**Maintainer**: MADSTONE TECHNOLOGY  
**Related**: `.gitignore`, `research/README.md`, `.specify/memory/constitution.md`
