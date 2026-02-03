# loko MCP Quick Start (5 minutes)

## Step 1: Build loko Binary

```bash
cd /path/to/loko
go build -o loko .
```

Verify it works:
```bash
./loko --help
```

## Step 2: Create/Initialize a Project

```bash
./loko init my-architecture
cd my-architecture
```

This creates:
- `loko.toml` - Project configuration
- `src/` - Directory for systems

## Step 3: Configure Claude Code

Open or create: `~/.claude/config.json`

Add this configuration:

```json
{
  "mcpServers": {
    "loko": {
      "command": "/path/to/loko/loko",
      "args": ["mcp", "-project", "/path/to/my-architecture"]
    }
  }
}
```

**Replace:**
- `/path/to/loko/loko` with full path to your loko binary
- `/path/to/my-architecture` with full path to your project

## Step 4: Restart Claude Code

Close and reopen Claude Code completely.

## Step 5: Try It Out

In Claude Code, you should now see "loko" in the available MCP servers.

**Example conversation:**

```
You: "I need to design a payment system architecture. 
      Can you create a Payment Service system with API and Database containers?"

Claude: [Uses create_system and create_container tools]
        "I've created the Payment Service system with:
         - Payment API (handles transactions)
         - Payment Database (stores transaction data)"

You: "Show me the architecture"

Claude: [Uses query_architecture tool]
        "Your architecture has 1 system with 2 containers..."

You: "Validate and build documentation"

Claude: [Uses validate and build_docs tools]
        "Architecture validated! Docs built to dist/"
```

## Quick Reference: Available Tools

| Tool | Purpose |
|------|---------|
| `query_project` | Get project metadata |
| `query_architecture` | View architecture (summary/structure/full) |
| `create_system` | Create new system |
| `create_container` | Create container in system |
| `create_component` | Create component in container |
| `update_diagram` | Update D2 diagram |
| `build_docs` | Generate HTML documentation |
| `validate` | Check for errors |

## Verify Generated Files

```bash
# View the created files
ls -la src/

# View generated documentation
open dist/index.html

# View markdown
cat src/[system-name]/system.md
cat src/[system-name]/[container-name]/container.md
```

## Troubleshooting

### "loko not found"
Make sure the path in config.json is absolute and correct:
```bash
which loko  # Copy this full path into config.json
```

### Config file not found
Create the directory:
```bash
mkdir -p ~/.claude
```

### MCP server not appearing
1. Verify config.json syntax: `cat ~/.claude/config.json | jq .`
2. Restart Claude Code completely
3. Check the file is saved: `cat ~/.claude/config.json`

### "Project not found"
Ensure the project path in config.json has `loko.toml`:
```bash
ls /your/project/path/loko.toml
```

## Next Steps

1. **Read MCP_SETUP.md** for detailed documentation
2. **Design your architecture** with Claude
3. **Share your feedback** on the tools and MCP integration
4. **Check out Phase 5** (TOON format for token efficiency)

---

That's it! You're ready to design architectures with Claude. 🎉

For more info: See [MCP_SETUP.md](./MCP_SETUP.md)
