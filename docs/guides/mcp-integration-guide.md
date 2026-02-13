# MCP Integration Guide

This guide shows you how to integrate loko with Claude Desktop and other MCP clients for AI-assisted architecture workflows.

## Table of Contents

- [What is MCP?](#what-is-mcp)
- [Quick Start (Claude Desktop)](#quick-start-claude-desktop)
- [Available MCP Tools](#available-mcp-tools)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

## What is MCP?

MCP (Model Context Protocol) is an open protocol for connecting AI assistants to external tools and data sources. With loko's MCP server, Claude Desktop and other AI assistants can:

- **Query architecture** without loading full graphs
- **Search elements** by name, technology, or tags
- **Create and update** systems, containers, and components
- **Build documentation** automatically
- **Validate architecture** for consistency

**Benefits:**
- ✅ Natural language architecture operations
- ✅ Token-efficient queries (TOON format support)
- ✅ No manual CLI commands needed
- ✅ Interactive architecture exploration
- ✅ Automated documentation workflows

## Quick Start (Claude Desktop)

### 1. Install loko

```bash
# macOS/Linux
brew install madstone-tech/tap/loko

# Or build from source
git clone https://github.com/madstone-tech/loko.git
cd loko
go build -o loko .
sudo mv loko /usr/local/bin/
```

### 2. Configure Claude Desktop

Add loko to your Claude Desktop MCP configuration:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%/Claude/claude_desktop_config.json`  
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "loko": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/your/architecture/project"
      }
    }
  }
}
```

**Replace `/path/to/your/architecture/project`** with your actual project path.

### 3. Restart Claude Desktop

Close and reopen Claude Desktop to load the MCP server.

### 4. Verify Connection

In Claude Desktop, ask:

```
Can you list the loko MCP tools?
```

You should see 17 tools (search_elements, find_relationships, query_architecture, etc.).

### 5. Try It Out

```
Can you query the architecture summary for my project?
```

Claude will use the `query_architecture` tool to fetch your architecture and display it.

## Available MCP Tools

loko provides **17 MCP tools** across 5 categories:

### 1. Query Tools (3 tools)

**query_architecture**
- Query architecture with configurable detail levels
- Formats: text (markdown), json, toon (token-optimized)
- Detail levels: summary (~200 tokens), structure (~500 tokens), full
- **Example:** "Show me a summary of the architecture"

**search_elements**
- Search for elements by name pattern, type, technology, or tags
- Supports glob patterns (`*`, `?`)
- Filters: type (system/container/component), technology, tag
- **Example:** "Find all containers using Python"

**find_relationships**
- Find relationships between elements
- Supports glob patterns for source/target
- Filters: relationship type (depends-on, uses, etc.)
- **Example:** "Show relationships where payment* depends on database*"

### 2. Project Management Tools (1 tool)

**query_project**
- Get project metadata and configuration
- Returns: name, description, template, output directory
- **Example:** "What's the project name?"

### 3. Creation Tools (3 tools)

**create_system**
- Create a new system with name, description, responsibilities
- Initializes directory structure in `src/systems/`
- **Example:** "Create a payment system for handling transactions"

**create_container**
- Create a new container within a system
- Specify technology (Go, Python, etc.)
- **Example:** "Create a REST API container in the payment system using Go"

**create_component**
- Create a new component within a container
- Specify type (service, repository, controller, etc.)
- **Example:** "Create a PaymentService component in the API container"

### 4. Update Tools (4 tools)

**update_system**
- Update system name, description, or responsibilities
- **Example:** "Update the payment system description to include fraud detection"

**update_container**
- Update container properties
- **Example:** "Change the API container technology to Node.js"

**update_component**
- Update component properties
- **Example:** "Rename PaymentService to TransactionService"

**update_diagram**
- Update D2 diagram definitions
- **Example:** "Add a connection from API to Database in the payment system diagram"

### 5. Build & Validation Tools (6 tools)

**build_docs**
- Build architecture documentation
- Formats: HTML, Markdown, PDF (requires veve-cli), TOON
- **Example:** "Build HTML documentation for the architecture"

**validate**
- Validate architecture for errors and warnings
- **Example:** "Validate the architecture and report any issues"

**validate_diagram**
- Validate D2 diagram syntax
- **Example:** "Check if the payment system diagram has syntax errors"

**Graph Tools** (query_graph, add_node, add_edge, remove_node, remove_edge)
- Low-level graph operations for advanced workflows
- **Example:** "Add an edge from payment-api to payment-db with label 'queries'"

## Usage Examples

### Example 1: Explore Architecture

```markdown
User: Can you show me a summary of the architecture?

Claude: [Uses query_architecture tool with detail=summary]
```

**Response:**
```
Architecture: E-Commerce Platform
Description: Microservices architecture for online shopping
Systems: 4
Containers: 16
Components: 58

Systems:
- Payment Core: Handles payment processing
- Order Management: Manages customer orders
- Inventory Service: Tracks product inventory
- User Service: Manages user accounts
```

### Example 2: Search for Payment Components

```markdown
User: Find all containers related to payments

Claude: [Uses search_elements tool with query="payment*", type="container"]
```

**Response:**
```
Found 3 containers:
1. payment-api (Go) - REST API for payment operations
2. payment-processor (Python) - Async payment processing
3. payment-db (PostgreSQL) - Payment data store
```

### Example 3: Create New System

```markdown
User: Create a notification system for sending email and SMS alerts

Claude: [Uses create_system tool]
```

**Response:**
```
✓ Created system: notification-system
  Location: src/systems/notification-system/
  Files:
    - system.md (system documentation)
    - context.d2 (C4 context diagram)
```

### Example 4: Build Documentation

```markdown
User: Build HTML documentation for the architecture

Claude: [Uses build_docs tool with format=html]
```

**Response:**
```
✓ Build completed in 2.3s
✓ Output: dist/
  Files:
    - index.html
    - systems/payment-core.html
    - systems/order-management.html
    - ...
```

### Example 5: Token-Efficient Query (TOON)

```markdown
User: Get the architecture structure in token-efficient format

Claude: [Uses query_architecture tool with format=toon, detail=structure]
```

**Response:**
```toon
|n:18:E-Commerce Platform|s:4:|
  |id:12:payment-core|n:12:Payment Core|c:3:|
    |id:15:payment-api|n:11:Payment API|t:3:Go||
    |id:18:payment-processor|n:17:Payment Processor|t:6:Python||
    |id:19:payment-db|n:14:Payment Store|t:10:PostgreSQL||
  |id:15:order-mgmt|n:15:Order Management|c:4:|...
```

**Token Savings:** ~50% fewer tokens vs JSON format

### Example 6: Find Relationships

```markdown
User: Show me which systems depend on the database

Claude: [Uses find_relationships tool with target_pattern="*-db"]
```

**Response:**
```
Found 5 relationships:
1. payment-api → payment-db (queries)
2. order-api → order-db (queries)
3. user-api → user-db (queries)
4. inventory-api → inventory-db (queries)
5. notification-worker → notification-db (queries)
```

## Configuration

### Environment Variables

**LOKO_PROJECT_ROOT** (required)
- Path to your architecture project root
- Must contain `loko.toml` and `src/` directory
- Example: `/Users/yourname/projects/my-architecture`

**LOKO_TEMPLATE_DIR** (optional)
- Custom template directory for scaffolding
- Defaults to built-in templates
- Example: `/Users/yourname/.loko/templates`

### Claude Desktop Config (Complete Example)

```json
{
  "mcpServers": {
    "loko": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/Users/yourname/architecture-project",
        "LOKO_TEMPLATE_DIR": "/Users/yourname/.loko/templates"
      }
    }
  }
}
```

### Multiple Projects

You can configure multiple loko servers for different projects:

```json
{
  "mcpServers": {
    "loko-ecommerce": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/ecommerce-architecture"
      }
    },
    "loko-backend": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/backend-architecture"
      }
    }
  }
}
```

Claude will have access to both projects simultaneously.

## Troubleshooting

### Issue 1: "loko: command not found"

**Problem:** Claude Desktop can't find the loko binary.

**Solution:** Use absolute path in configuration:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/usr/local/bin/loko",  // Full path
      "args": ["mcp"]
    }
  }
}
```

**Find loko path:**
```bash
which loko
# Output: /usr/local/bin/loko
```

### Issue 2: "Project not found" errors

**Problem:** LOKO_PROJECT_ROOT is incorrect or missing.

**Solution:** Verify path and ensure `loko.toml` exists:

```bash
# Check if directory exists
ls -la /path/to/your/project

# Verify loko.toml
cat /path/to/your/project/loko.toml
```

**Correct configuration:**
```json
{
  "env": {
    "LOKO_PROJECT_ROOT": "/Users/yourname/architecture-project"
  }
}
```

### Issue 3: MCP tools not appearing

**Problem:** Claude Desktop doesn't show loko tools.

**Solutions:**
1. **Restart Claude Desktop** (required after config changes)
2. **Check config syntax** (use a JSON validator)
3. **Verify loko version** (ensure v0.2.0+)

```bash
loko version
# Should show: loko version 0.2.0 or higher
```

4. **Test MCP server manually:**

```bash
cd /path/to/your/project
loko mcp
# Should start MCP server and wait for input
```

Press Ctrl+C to stop.

### Issue 4: Permission denied errors

**Problem:** loko can't write to project directory.

**Solution:** Fix permissions:

```bash
# Make project directory writable
chmod -R u+w /path/to/your/project

# Or create new system manually first
cd /path/to/your/project
mkdir -p src/systems/test-system
```

### Issue 5: Slow query responses

**Problem:** Architecture queries take > 5 seconds.

**Possible Causes:**
1. **Large project:** Use `detail: summary` instead of `full`
2. **Format:** Use `toon` instead of `text` for faster parsing
3. **Caching:** Architecture graph is rebuilt on each query

**Optimization:**

```markdown
# Instead of:
"Show me the full architecture details"

# Use:
"Show me an architecture summary" (200 tokens vs 5,000)
```

## Advanced Usage

### 1. Custom Workflows

**Workflow: Create Multi-Tier System**

```markdown
User: Create a three-tier web application with frontend, API, and database

Claude: [Executes workflow]
1. create_system(name="web-app", description="Three-tier web application")
2. create_container(system="web-app", name="frontend", technology="React")
3. create_container(system="web-app", name="api", technology="Node.js")
4. create_container(system="web-app", name="database", technology="PostgreSQL")
5. update_diagram(system="web-app", add connections)
6. build_docs(format="html")
```

Result: Complete system scaffolded and documented in seconds.

### 2. Batch Operations

```markdown
User: Create 5 microservices: user, order, payment, inventory, notification

Claude: [Loops through create_system for each service]
```

### 3. Architecture Validation

```markdown
User: Check the architecture for errors and build documentation if valid

Claude:
1. validate() - Check for errors
2. If valid: build_docs(format="html")
3. If invalid: Report errors to user
```

### 4. Token-Efficient Queries

For high-frequency queries or large architectures:

```markdown
User: Get payment system structure in TOON format

Claude: [Uses query_architecture with format=toon]
# ~50% fewer tokens vs JSON
```

### 5. Relationship Analysis

```markdown
User: Find all circular dependencies in the architecture

Claude:
1. find_relationships(source_pattern="*")
2. Analyze results for cycles
3. Report findings
```

## Best Practices

### 1. Use Appropriate Detail Levels

```markdown
# ✅ Good: Quick overview
"Show me a summary of the architecture"
# Result: ~200 tokens

# ❌ Avoid: Unnecessary detail
"Show me every single component in the architecture"
# Result: ~5,000 tokens (25x more expensive)
```

### 2. Leverage Search Tools

```markdown
# ✅ Good: Targeted search
"Find all Python containers in the payment system"

# ❌ Avoid: Fetching everything
"Show me all containers, I'll filter myself"
```

### 3. Use TOON Format for Repeated Queries

```markdown
# For dashboards or frequent polling
"Get architecture summary in TOON format"
# 50% token savings per query
```

### 4. Batch Related Operations

```markdown
# ✅ Good: Single conversation
"Create payment system, add API container, build docs"

# ❌ Avoid: Multiple conversations
"Create payment system"
[New conversation]
"Add API to payment"
[New conversation]
"Build docs"
```

### 5. Validate Before Building

```markdown
# ✅ Good: Validate first
"Validate architecture, then build if valid"

# ❌ Avoid: Build without validation
"Build documentation"
# May fail if architecture has errors
```

## Next Steps

- **CI/CD Integration:** See [CI/CD Integration Guide](./ci-cd-integration.md) for automated workflows
- **TOON Format:** See [TOON Format Guide](./toon-format-guide.md) for token efficiency
- **Examples:** Explore `examples/` directory for sample architectures

---

**Last updated:** 2025-02-13  
**loko version:** v0.2.0  
**MCP protocol:** Latest  
**Tested with:** Claude Desktop (macOS/Linux/Windows)
