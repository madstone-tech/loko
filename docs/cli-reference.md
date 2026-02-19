# CLI Reference

Complete reference for all `loko` commands and flags.

## Global Flags

| Flag | Description |
|------|-------------|
| `--help, -h` | Show help for any command |
| `--version, -v` | Show loko version |

---

## loko init

Initialize a new loko project in the current directory.

```bash
loko init [project-name] [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--template` | string | `standard-3layer` | Project template (`standard-3layer`, `serverless`) |
| `--description` | string | `""` | Project description |

**Examples**:
```bash
loko init my-project
loko init payment-service --template serverless
```

---

## loko new

Create a new architecture element (system, container, or component).

### loko new system

```bash
loko new system [flags]
```

**Flags**:

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--name` | string | Yes | System display name |
| `--description` | string | No | System description |

### loko new container

```bash
loko new container [flags]
```

**Flags**:

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--name` | string | Yes | Container display name |
| `--technology` | string | No | Technology stack |
| `--system` | string | Yes | Parent system name |
| `--description` | string | No | Container description |

### loko new component

```bash
loko new component [flags]
```

**Flags**:

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--name` | string | Yes | Component display name |
| `--technology` | string | No | Technology (used for template auto-selection) |
| `--container` | string | Yes | Parent container name |
| `--system` | string | Yes | Parent system name |
| `--description` | string | No | Component description |
| `--template` | string | No | **NEW v0.2.0** — Override auto-selected template (e.g., `compute`, `datastore`, `messaging`) |
| `--preview` | bool | No | **NEW v0.2.0** — Render and display a D2 diagram preview after creation |

**Template auto-selection** (v0.2.0):
- `AWS Lambda` → `compute`
- `DynamoDB`, `RDS` → `datastore`
- `SQS`, `SNS`, `Kinesis` → `messaging`
- `API Gateway`, `REST` → `api`
- `EventBridge`, `Step Functions` → `event`
- `S3`, `EFS` → `storage`
- (anything else) → `generic`

**Examples**:
```bash
# Auto-select template based on technology
loko new component --name "Payment Processor" \
                   --technology "AWS Lambda" \
                   --container api-gateway \
                   --system payment-service

# Override template selection
loko new component --name "Cache Manager" \
                   --technology "Redis" \
                   --template datastore \
                   --container backend \
                   --system my-service

# Show diagram preview after creation
loko new component --name "Auth Handler" \
                   --technology "Go" \
                   --container api \
                   --system auth-service \
                   --preview
```

---

## loko build

Build architecture documentation.

```bash
loko build [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `html` | Output format: `html`, `markdown`, `pdf`, `toon` |
| `--output` | string | `./docs/output` | Output directory |
| `--project` | string | `.` | Project root directory |

**Examples**:
```bash
loko build
loko build --format markdown --output ./docs
loko build --format pdf
loko build --format toon
```

---

## loko validate

Validate the architecture for consistency issues.

```bash
loko validate [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check-drift` | bool | `false` | **NEW v0.2.0** — Check for inconsistencies between D2 diagrams and frontmatter |
| `--project` | string | `.` | Project root directory |

**Drift detection** (`--check-drift`):
- Reports `DriftDescriptionMismatch` as WARNING (D2 tooltip ≠ frontmatter description)
- Reports `DriftMissingComponent` as ERROR (D2 arrow targets non-existent component)
- Reports `DriftOrphanedRelationship` as ERROR (frontmatter relationship to deleted component)
- Exit code `1` if any ERROR-level drift is found; `0` otherwise

**Examples**:
```bash
loko validate
loko validate --check-drift
loko validate --check-drift --project /path/to/project
```

**Sample output** (with drift):
```
❌ Validation failed - Critical drift detected

Issues found:
  auth-handler (ERROR): Orphaned relationship - target 'old-service' not found

Summary:
  Components checked: 17
  Drift issues found: 1 (0 warnings, 1 error)
```

---

## loko serve

Start the local documentation server.

```bash
loko serve [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | `3000` | Port to listen on |
| `--host` | string | `localhost` | Host address |
| `--project` | string | `.` | Project root directory |

---

## loko mcp

Start the MCP (Model Context Protocol) server for AI assistant integration.

```bash
loko mcp [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project` | string | `.` | Project root directory |

See the [MCP Integration Guide](./guides/mcp-integration-guide.md) for setup instructions.

---

## loko watch

Watch for file changes and rebuild documentation automatically.

```bash
loko watch [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `html` | Output format to rebuild on changes |
| `--project` | string | `.` | Project root directory |

---

## loko export

Export architecture data to various formats.

```bash
loko export [flags]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `json` | Export format: `json`, `toon` |
| `--output` | string | `stdout` | Output file path |

---

## loko completion

Generate shell completion scripts.

```bash
loko completion [bash|zsh|fish|powershell]
```

**Examples**:
```bash
# Bash
loko completion bash > /etc/bash_completion.d/loko

# Zsh
loko completion zsh > "${fpath[1]}/_loko"

# Fish
loko completion fish > ~/.config/fish/completions/loko.fish
```

---

## loko version

Print the current version.

```bash
loko version
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LOKO_CONFIG_HOME` | Override config directory (default: `~/.config/loko`) |
| `LOKO_PROJECT_ROOT` | Override project root detection |
| `XDG_CONFIG_HOME` | XDG config base directory |
| `XDG_DATA_HOME` | XDG data base directory |
| `XDG_CACHE_HOME` | XDG cache base directory |
