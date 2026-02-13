# MCP Integration Setup Guide for Claude Code

This guide explains how to integrate loko's MCP server with Claude Code (OpenCode) for conversational architecture design.

## Prerequisites

- loko binary built and available in your PATH
- Claude Code installed and updated
- A loko project initialized (`loko init`)

## Quick Start

### 1. Build the loko Binary

First, ensure you have the latest loko binary built:

```bash
cd /path/to/loko
go build -o loko .
```

Or if you have it in your path:
```bash
which loko
# Should show: /path/to/loko/loko
```

### 2. Add loko MCP to Claude Code Configuration

Claude Code uses MCP servers defined in a configuration file. The configuration location depends on your OS:

**macOS & Linux:**
```bash
~/.claude/config.json
```

**Windows:**
```
%APPDATA%\Claude\config.json
```

### 3. Configure the MCP Server

Add the loko MCP server to your Claude Code configuration:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/path/to/loko",
      "args": ["mcp", "-project", "/path/to/your/project"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

**Key parameters:**
- `command`: Full path to the loko binary
- `args`: 
  - `mcp` - Start MCP server mode
  - `-project` - Path to your loko project root
- `env`:
  - `LOG_LEVEL`: Set to "debug" for verbose output, "info" for normal

### 4. Example Configuration

Here's a complete example configuration file:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp", "-project", "/Users/you/my-architecture-project"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### 5. Verify the Setup

Test the MCP connection by starting Claude Code:

1. Open Claude Code
2. In the composer, you should see "loko" in the available MCP servers list
3. Click on "loko" to enable it

If you see errors, check:
- Path to loko binary is correct
- Project path exists and is a valid loko project
- loko binary is executable: `chmod +x /path/to/loko`

## Available Tools

Once connected, Claude Code can use these loko tools:

### 1. **query_project**
Get current project metadata and statistics.

```
Project: My Architecture
Systems: 3
Containers: 7
Components: 23
```

### 2. **query_architecture**
Query architecture with configurable detail levels (saves tokens):

- **summary** (~50 tokens) - Project overview with counts
- **structure** (~100 tokens) - Systems and their containers
- **full** (~200+ tokens) - Complete details with all components

Usage:
```
Claude: "Show me the architecture at structure level"
→ Returns: Systems with containers and descriptions
```

### 3. **create_system**
Create a new system in your architecture.

```
Claude: "Create a new system called 'Payment Processing'"
→ Creates system.md and system.d2 files
```

### 4. **create_container**
Create a new container in a system.

```
Claude: "Add a container called 'Payment API' to the Payment Processing system"
→ Creates container.md file
```

### 5. **create_component**
Create a new component in a container.

```
Claude: "Add a component called 'Transaction Handler' to the Payment API"
→ Creates component structure
```

### 6. **update_diagram**
Update D2 diagram source code.

```
Claude: "Update the Payment Processing diagram with nodes for Auth and Ledger"
→ Updates system.d2 file
```

### 7. **build_docs**
Trigger documentation build.

```
Claude: "Build the documentation"
→ Generates HTML site in dist/
```

### 8. **validate**
Validate architecture for errors and warnings.

```
Claude: "Validate the architecture"
→ Checks for missing containers, orphaned systems, etc.
```

## Typical Workflow

Here's how you'd use loko with Claude Code:

### 1. Initialize a Project
```bash
loko init my-microservices
cd my-microservices
```

### 2. Start Claude Code with loko
- Configure loko MCP in ~/.claude/config.json
- Open Claude Code
- Enable the loko MCP server

### 3. Design via Conversation

**You:**
> I'm building a microservices architecture for an e-commerce platform. I need payment processing, order management, and inventory systems.

**Claude:**
> I'll create the three systems for you:
> - Payment Processing System
> - Order Management System  
> - Inventory System
> 
> [Uses create_system tool 3 times]

**You:**
> What should each system contain?

**Claude:**
> Let me show you the current structure at a summary level:
> [Uses query_architecture with "summary" detail]
> 
> The Payment Processing system needs:
> - Payment API container
> - Ledger service
> - Notification service

**You:**
> Add those containers

**Claude:**
> [Uses create_container for each one]

**You:**
> Draw a diagram showing how payment flows through the system

**Claude:**
> [Uses update_diagram to add D2 diagram]

**You:**
> Validate the architecture and build the documentation

**Claude:**
> [Uses validate and build_docs tools]
> Your architecture is valid! Documentation generated at dist/index.html

### 4. Review Generated Files

```bash
# View the generated documentation
open dist/index.html

# Or view the markdown files
cat src/payment-processing/system.md
cat src/payment-processing/payment-api/container.md
```

## Troubleshooting

### Issue: "loko not found"
**Solution:** Ensure loko is in your PATH:
```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="/path/to/loko:$PATH"
```

### Issue: "Permission denied" when running loko
**Solution:** Make the binary executable:
```bash
chmod +x /path/to/loko/loko
```

### Issue: MCP server won't start
**Solution:** Test manually:
```bash
/path/to/loko mcp -project /path/to/your/project
```

If you see errors, check:
- Project path is correct: `ls /path/to/your/project/loko.toml`
- loko binary is in PATH: `which loko`
- No port conflicts: The MCP server uses stdio, not ports

### Issue: Tools aren't showing up in Claude Code
**Solution:** 
1. Restart Claude Code completely
2. Check config.json syntax: `cat ~/.claude/config.json | python3 -m json.tool`
3. Check loko logs: Look for any errors when starting the MCP server

## Advanced Configuration

### Environment Variables

You can pass environment variables to loko:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp", "-project", "."],
      "env": {
        "LOG_LEVEL": "debug",
        "LOKO_CACHE": "true"
      }
    }
  }
}
```

### Using Relative Paths

For projects in your current directory:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp", "-project", "."]
    }
  }
}
```

### Multiple Projects

You can add multiple loko MCP servers for different projects:

```json
{
  "mcpServers": {
    "loko-backend": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp", "-project", "/Users/you/backend-architecture"]
    },
    "loko-frontend": {
      "command": "/usr/local/bin/loko",
      "args": ["mcp", "-project", "/Users/you/frontend-architecture"]
    }
  }
}
```

## Token Efficiency

The loko MCP tools are designed to be token-efficient:

- **query_architecture** with "summary" uses ~50 tokens
- **query_architecture** with "structure" uses ~100 tokens
- **query_architecture** with "full" uses ~200+ tokens

This allows Claude to efficiently query your architecture without consuming excessive context window.

## Best Practices

1. **Start with summary level queries**: When asking Claude about your architecture, let it use "summary" detail first
2. **Validate frequently**: After Claude makes changes, use the validate tool to catch errors
3. **Build docs regularly**: Keep documentation up-to-date with `build_docs`
4. **Use descriptive names**: When creating systems/containers, use clear names for better diagrams
5. **Review generated files**: Always review the generated markdown and D2 files to ensure quality

## Example: Complete Architecture Design Session

```
You: I need to design an API gateway architecture. Can you help?

Claude: Of course! Let me create the foundational structure.
[Uses create_system to add "API Gateway System"]

You: Create a load balancer, auth service, and rate limiter container

Claude: [Uses create_container 3 times with appropriate names]

You: Show me what we have at the structure level

Claude: [Uses query_architecture with detail="structure"]
Your API Gateway System has:
- Load Balancer: Distributes traffic across instances
- Auth Service: Handles authentication and tokens
- Rate Limiter: Prevents abuse

You: Add components to the Auth Service - JWT validator and token cache

Claude: [Uses create_component twice]

You: Draw a diagram with connections between components

Claude: [Uses update_diagram with D2 code]

You: Validate everything and build docs

Claude: [Uses validate and build_docs]
Your architecture is valid! 
Docs built to dist/
No missing containers or orphaned systems.
```

## Support & Issues

If you encounter issues:

1. Check that loko is built: `loko version` or `loko --help`
2. Test MCP manually: `/path/to/loko mcp -project /path/to/project`
3. Check Claude Code logs: Look in the Claude Code application logs
4. File an issue: https://github.com/madstone-tech/loko/issues

## What's Next?

Once you've set up the MCP integration:

1. **Explore Phase 5 features** (TOON format, advanced queries)
2. **Export to multiple formats** (HTML, Markdown, PDF in future phases)
3. **Integrate with CI/CD** (REST API in Phase 6)
4. **Share architectures** (version control your C4 diagrams)

---

**Happy designing! 🎨**

For more information, see:
- [loko Documentation](https://github.com/madstone-tech/loko)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Claude Code Documentation](https://claude.ai/docs)
