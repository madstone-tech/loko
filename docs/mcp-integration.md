# MCP Integration Guide

loko includes a Model Context Protocol (MCP) server that enables AI assistants like Claude to help design and document software architecture.

## What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io/) is an open protocol that allows AI assistants to interact with external tools and data sources. loko implements an MCP server that exposes architecture design and documentation tools.

## Setup with Claude Desktop

### 1. Install loko

```bash
go install github.com/madstone-tech/loko@latest
```

### 2. Configure Claude Desktop

Add loko to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "loko": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/your/project"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

After adding the configuration, restart Claude Desktop to load the loko MCP server.

## Available MCP Tools

loko exposes 12 tools through MCP:

### Query Tools

| Tool | Description |
|------|-------------|
| `query_project` | Get project overview and statistics |
| `query_architecture` | Query architecture with progressive detail levels |
| `query_dependencies` | Analyze dependencies between components |
| `query_related_components` | Find related components |
| `analyze_coupling` | Analyze coupling between systems |

### Creation Tools

| Tool | Description |
|------|-------------|
| `create_system` | Create a new system |
| `create_container` | Create a new container in a system |
| `create_component` | Create a new component in a container |
| `update_diagram` | Update a D2 diagram |

### Build Tools

| Tool | Description |
|------|-------------|
| `build_docs` | Build documentation |
| `validate` | Validate architecture |
| `validate_diagram` | Validate D2 diagram syntax |

## Usage Examples

### Query Architecture

Ask Claude to explore your architecture:

> "What systems are in this project?"

Claude will use `query_project` to get an overview:

```
Project: E-Commerce Platform
Systems: 4 (AuthService, ProductCatalog, OrderService, PaymentGateway)
Containers: 12
Components: 34
```

### Progressive Detail

Request different detail levels:

> "Show me the structure of the AuthService system"

Claude uses `query_architecture` with detail level:

- **summary** (~200 tokens): High-level overview
- **structure** (~500 tokens): Systems and containers
- **full**: Complete architecture details

### Create Architecture

Design new systems conversationally:

> "Create a new notification system with email and SMS containers"

Claude will:
1. Use `create_system` to create "NotificationService"
2. Use `create_container` to add "EmailService" and "SMSService"
3. Use `update_diagram` to create the system diagram

### Validate Architecture

Check for issues:

> "Validate the architecture for any problems"

Claude uses `validate` to check:
- Empty systems
- Missing descriptions
- Orphaned references
- Invalid hierarchy

### Analyze Dependencies

Understand relationships:

> "What does the OrderService depend on?"

Claude uses `query_dependencies` to trace:
- Direct dependencies
- Transitive dependencies
- Potential circular dependencies

## Token-Efficient Queries

loko supports TOON (Token-Optimized Object Notation) for efficient context usage:

```
# JSON format (more tokens)
{"name": "AuthService", "description": "Handles authentication"}

# TOON format (fewer tokens)
{n:AuthService,d:Handles authentication}
```

Request TOON format for large architectures:

> "Show me the full architecture in TOON format"

This reduces token usage by 40-90% compared to JSON.

## Best Practices

### 1. Start with Queries

Before creating new elements, query the existing architecture:

> "What's the current structure of this project?"

### 2. Use Progressive Detail

Start with summary, drill down as needed:

> "Give me a summary of the project"
> "Now show me details of the PaymentService"

### 3. Validate After Changes

Always validate after making changes:

> "I just created the new system. Can you validate everything?"

### 4. Build Documentation

After design sessions, build updated documentation:

> "Build the HTML documentation"

## Troubleshooting

### Server Not Starting

Check if loko is in your PATH:

```bash
which loko
```

### Connection Issues

Verify the configuration path is correct:

```bash
# Test MCP server directly
echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | loko mcp
```

### Permission Errors

Ensure the project directory is writable:

```bash
ls -la /path/to/your/project
```

### Debug Mode

Enable verbose logging:

```bash
LOKO_LOG_LEVEL=debug loko mcp
```

## Advanced Configuration

### Multiple Projects

Configure multiple loko instances for different projects:

```json
{
  "mcpServers": {
    "loko-frontend": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/frontend-project"
      }
    },
    "loko-backend": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/backend-project"
      }
    }
  }
}
```

### Custom Templates

Point to custom template directories:

```json
{
  "mcpServers": {
    "loko": {
      "command": "loko",
      "args": ["mcp"],
      "env": {
        "LOKO_PROJECT_ROOT": "/path/to/project",
        "LOKO_TEMPLATE_DIR": "/path/to/custom/templates"
      }
    }
  }
}
```

## Example Session

Here's a complete example of designing architecture with Claude:

1. **Initialize Project**
   > "Create a new e-commerce architecture"

2. **Design Systems**
   > "Add systems for: user management, product catalog, shopping cart, and checkout"

3. **Add Containers**
   > "The user management system needs an API, a database, and a cache"

4. **Define Components**
   > "The API container should have handlers for authentication, profile management, and password reset"

5. **Create Diagrams**
   > "Generate a system diagram showing all the systems and their relationships"

6. **Validate**
   > "Check if there are any issues with the architecture"

7. **Build**
   > "Build the HTML documentation so I can review it"

## Resources

- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Claude Desktop Documentation](https://claude.ai/docs)
- [loko GitHub Repository](https://github.com/madstone-tech/loko)
