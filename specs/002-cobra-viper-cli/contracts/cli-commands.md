# CLI Command Contract

**Branch**: `002-cobra-viper` | **Date**: 2026-02-05

---

## Root Command

```
loko [flags]
loko [command]
```

### Persistent Flags (all subcommands)

| Flag | Short | Type | Default | Env Var | Description |
|------|-------|------|---------|---------|-------------|
| `--config` | | string | "" | `LOKO_CONFIG_HOME` | Path to config file or directory |
| `--project` | `-p` | string | "." | | Project root directory |
| `--verbose` | `-v` | bool | false | `LOKO_VERBOSE` | Enable verbose output |

### Command Groups

| Group ID | Title | Commands |
|----------|-------|----------|
| scaffolding | Scaffolding | init, new |
| building | Building | build, watch, validate |
| serving | Serving | serve, api, mcp |

---

## Scaffolding Commands

### `loko init <project-name>`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | `-d` | string | "" | Project description |
| `--path` | | string | "" | Project path (defaults to project name) |
| `--template` | `-t` | string | "standard-3layer" | Template to use |

**Completions**: `--template` → list templates from filesystem

### `loko new system <name>`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | `-d` | string | "" | System description |
| `--technology` | | string | "" | Technology stack |
| `--template` | `-t` | string | "" | Template override |

### `loko new container <name>`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | `-d` | string | "" | Container description |
| `--technology` | | string | "" | Technology stack |
| `--parent` | | string | "" | **Required.** Parent system name |
| `--template` | `-t` | string | "" | Template override |

**Completions**: `--parent` → list systems from project

### `loko new component <name>`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | `-d` | string | "" | Component description |
| `--technology` | | string | "" | Technology stack |
| `--parent` | | string | "" | **Required.** Parent container (format: `system/container`) |
| `--template` | `-t` | string | "" | Template override |

**Completions**: `--parent` → list containers from project

---

## Building Commands

### `loko build`

| Flag | Short | Type | Default | Env Var | Description |
|------|-------|------|---------|---------|-------------|
| `--clean` | | bool | false | | Rebuild everything |
| `--output` | `-o` | string | "dist" | `LOKO_OUTPUT_DIR` | Output directory |
| `--format` | `-f` | []string | ["html"] | | Output formats (html,markdown,pdf) |
| `--d2-theme` | | string | "neutral-default" | `LOKO_D2_THEME` | D2 diagram theme |
| `--d2-layout` | | string | "elk" | `LOKO_D2_LAYOUT` | D2 layout engine |

**Completions**: `--format` → `html`, `markdown`, `pdf`; `--d2-theme` → list themes; `--d2-layout` → `dagre`, `elk`, `tala`

### `loko watch`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | "dist" | Output directory |
| `--debounce` | | int | 500 | Debounce delay (ms) |

### `loko validate`

No additional flags beyond persistent flags.

---

## Serving Commands

### `loko serve`

| Flag | Short | Type | Default | Env Var | Description |
|------|-------|------|---------|---------|-------------|
| `--output` | `-o` | string | "dist" | | Directory to serve |
| `--address` | | string | "localhost" | | Server address |
| `--port` | | string | "8080" | `LOKO_SERVER_SERVE_PORT` | Server port |

### `loko mcp`

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--env` | | string | "" | Environment variable (KEY=VALUE) |

### `loko api`

No additional flags beyond persistent flags.

---

## Utility Commands

### `loko completion <shell>`

Shells: `bash`, `zsh`, `fish`, `powershell`

No additional flags. Outputs completion script to stdout.

### `loko version`

No flags. Outputs: `loko <version> (commit: <hash>, built: <date> by <builder>)`

---

## Config Hierarchy Contract

**Value precedence** (highest → lowest):

```
1. CLI flag (only when explicitly passed, HasChanged==true)
2. LOKO_* environment variable
3. Project config (./loko.toml, merged via MergeInConfig)
4. Global config ($XDG_CONFIG_HOME/loko/config.toml, via ReadInConfig)
5. Built-in default (viper.SetDefault)
```

**Path precedence** for config file location:

```
1. --config flag
2. LOKO_CONFIG_HOME env var
3. XDG_CONFIG_HOME/loko/ env var
4. ~/.config/loko/ (default)
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid args, command failure) |
| 2 | Config error (invalid TOML, missing required config) |
