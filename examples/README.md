# loko Examples

This directory contains example projects demonstrating different architecture patterns and loko features.

## Examples

### [simple-project](./simple-project/)

A minimal project with a single system demonstrating:
- Basic project structure
- System documentation with markdown
- D2 diagram creation
- loko.toml configuration

**Use case**: Getting started with loko

### [3layer-app](./3layer-app/)

A three-tier web application demonstrating:
- Multiple systems (Frontend, API, Database)
- Container-level decomposition
- Inter-system dependencies
- Multiple output formats

**Use case**: Traditional web application architecture

### [microservices](./microservices/)

A microservices architecture demonstrating:
- Multiple independent services
- Service mesh patterns
- API gateway
- Event-driven communication

**Use case**: Distributed systems documentation

## Running Examples

Each example can be built and viewed:

```bash
# Navigate to example
cd simple-project

# Build documentation
loko build

# Preview in browser
loko serve
# Open http://localhost:8080

# Or build with watch mode
loko watch
```

## Creating Your Own

Use these examples as templates:

```bash
# Copy an example
cp -r simple-project my-project
cd my-project

# Edit loko.toml with your project details
vim loko.toml

# Start designing
loko new system "My System"
```

## Project Structure

All examples follow the same structure:

```
example-name/
├── loko.toml           # Project configuration
├── src/                # Architecture source files
│   └── system-name/
│       ├── system.md   # System documentation
│       ├── system.d2   # System diagram
│       └── containers/ # Container subdirectories
└── dist/               # Generated documentation (after build)
```
