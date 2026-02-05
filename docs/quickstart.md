# Quickstart Guide

Get started with loko in 5 minutes. This guide walks you through creating your first C4 architecture documentation project.

## Prerequisites

- Go 1.25 or later
- [D2](https://d2lang.com/) diagram tool (for rendering diagrams)

```bash
# Install D2 (macOS)
brew install d2

# Install D2 (Linux)
curl -fsSL https://d2lang.com/install.sh | sh
```

## Installation

```bash
# Install loko
go install github.com/madstone-tech/loko@latest

# Verify installation
loko --help
```

## Create Your First Project

### 1. Initialize a new project

```bash
loko init my-architecture
cd my-architecture
```

This creates a `loko.toml` configuration file and the basic project structure:

```
my-architecture/
├── loko.toml           # Project configuration
└── src/                # Architecture source files
```

### 2. Create a system

```bash
loko new system "Payment Service"
```

This creates a new system with a starter template:

```
src/
└── payment-service/
    ├── system.md       # System documentation
    └── system.d2       # System diagram (D2 format)
```

### 3. Add containers to your system

```bash
loko new container "API Gateway" -parent payment-service
loko new container "Payment Processor" -parent payment-service
loko new container "Database" -parent payment-service
```

### 4. Add components (optional)

```bash
loko new component "Auth Handler" -parent api-gateway
loko new component "Request Router" -parent api-gateway
```

### 5. Build documentation

```bash
# Build HTML documentation
loko build

# Build with multiple formats
loko build -format html,markdown

# Build to custom directory
loko build -output docs
```

### 6. Preview your documentation

```bash
# Start local server
loko serve

# Open http://localhost:8080 in your browser
```

## Project Structure

After following this guide, your project will look like:

```
my-architecture/
├── loko.toml
├── src/
│   └── payment-service/
│       ├── system.md
│       ├── system.d2
│       ├── api-gateway/
│       │   ├── container.md
│       │   ├── container.d2
│       │   ├── auth-handler/
│       │   │   ├── component.md
│       │   │   └── component.d2
│       │   └── request-router/
│       │       ├── component.md
│       │       └── component.d2
│       ├── payment-processor/
│       │   ├── container.md
│       │   └── container.d2
│       └── database/
│           ├── container.md
│           └── container.d2
└── dist/               # Generated documentation
    ├── index.html
    ├── systems/
    ├── containers/
    ├── components/
    ├── diagrams/
    └── README.md       # If markdown format enabled
```

## Watch Mode

For rapid iteration, use watch mode to automatically rebuild on changes:

```bash
loko watch
```

This monitors your `src/` directory and rebuilds documentation whenever files change.

## Validation

Check your architecture for issues:

```bash
loko validate
```

This checks for:
- Empty systems (no containers)
- Missing descriptions
- Orphaned references
- Invalid hierarchy

## Using with Claude (MCP)

loko includes an MCP server for AI-assisted architecture design:

```bash
# Start MCP server (for Claude Desktop integration)
loko mcp
```

See the [MCP Integration Guide](mcp-integration.md) for setup instructions.

## Next Steps

- Read the [Configuration Reference](configuration.md) for all loko.toml options
- Explore [example projects](../examples/) for common architecture patterns
- Learn about [MCP integration](mcp-integration.md) for AI-assisted design

## Common Commands

| Command | Description |
|---------|-------------|
| `loko init <project-name>` | Initialize a new project |
| `loko new system <name>` | Create a new system |
| `loko new container <name> -parent <system>` | Create a new container |
| `loko new component <name> -parent <container>` | Create a new component |
| `loko build` | Build documentation |
| `loko build -format markdown` | Build as Markdown |
| `loko serve` | Start preview server |
| `loko watch` | Watch mode with auto-rebuild |
| `loko validate` | Validate architecture |
| `loko mcp` | Start MCP server |
| `loko api` | Start HTTP API server |

## Getting Help

```bash
# General help
loko --help

# Command-specific help
loko build --help
loko new --help
```

For issues and feature requests, visit: https://github.com/madstone-tech/loko/issues
