# Data Model: Cobra & Viper CLI Migration

**Branch**: `002-cobra-viper` | **Date**: 2026-02-05

---

## Entities

### Config (modified from existing `ProjectConfig`)

The existing `ProjectConfig` entity in `internal/core/entities/project.go` is extended to support the full config hierarchy.

```
Config
├── Source: ConfigSource (flag | env | project | global | default)
├── Project
│   ├── Name: string
│   ├── Description: string
│   ├── Version: string
│   └── Template: string
├── Paths
│   ├── SourceDir: string (default: "./src")
│   └── OutputDir: string (default: "./dist")
├── D2
│   ├── Theme: string (default: "neutral-default")
│   ├── Layout: string (default: "elk")
│   └── Cache: bool (default: true)
├── Outputs
│   ├── HTML: bool (default: true)
│   ├── Markdown: bool (default: false)
│   └── PDF: bool (default: false)
├── Build
│   ├── Parallel: bool (default: true)
│   └── MaxWorkers: int (default: 4)
└── Server
    ├── ServePort: int (default: 8080)
    ├── APIPort: int (default: 8081)
    └── HotReload: bool (default: true)
```

**Validation rules**:
- `D2.Theme` must be a valid D2 theme name or a file in themes dir
- `D2.Layout` must be one of: dagre, elk, tala
- `Build.MaxWorkers` must be >= 1
- `Server.ServePort` and `Server.APIPort` must be 1024-65535
- `Paths.SourceDir` and `Paths.OutputDir` must be relative paths

### XDGPaths (new entity)

Resolves XDG-compliant paths for loko's data directories.

```
XDGPaths
├── ConfigHome: string  → $LOKO_CONFIG_HOME or $XDG_CONFIG_HOME/loko/ or ~/.config/loko/
├── DataHome: string    → $XDG_DATA_HOME/loko/ or ~/.local/share/loko/
├── CacheHome: string   → $XDG_CACHE_HOME/loko/ or ~/.cache/loko/
├── ConfigFile(): string → ConfigHome/config.toml
├── ThemesDir(): string  → DataHome/themes/
└── CacheDir(): string   → CacheHome/
```

**Resolution precedence for ConfigHome**:
1. `--config` flag (if set, use parent directory)
2. `LOKO_CONFIG_HOME` env var
3. `XDG_CONFIG_HOME` env var + `/loko/`
4. `~/.config/loko/` (default on all platforms)

**Validation rules**:
- All paths must be absolute after resolution
- Reject relative `XDG_*` env var values (per XDG spec)

### Theme (new entity)

```
Theme
├── Name: string        → derived from filename (e.g., "dark" from "dark.toml")
├── Path: string        → absolute path to theme file
├── D2Theme: string     → D2 theme name override
├── Colors: map[string]string → color overrides (key → hex value)
└── Styles: map[string]string → D2 style overrides
```

**Validation rules**:
- Name must match filename (no path separators)
- D2Theme must be a valid D2 built-in theme name
- Colors must be valid hex color codes

### Command (Cobra-managed, not a domain entity)

Cobra manages command definitions internally. No custom entity needed — Cobra's `cobra.Command` struct handles name, description, flags, and execution.

---

## Relationships

```
main.go
  └── rootCmd (cobra.Command)
        ├── PersistentPreRunE → initConfig()
        │     ├── resolves XDGPaths
        │     ├── loads global config (ReadInConfig)
        │     ├── merges project config (MergeInConfig)
        │     ├── enables AutomaticEnv (LOKO_*)
        │     └── binds persistent flags
        ├── PersistentFlags
        │     ├── --config (overrides config file path)
        │     ├── --project (project root dir)
        │     └── --verbose
        └── Subcommands
              ├── init → InitCommand.Execute()
              ├── new → [system|container|component]
              ├── build → BuildCommand.Execute()
              ├── serve → ServeCommand.Execute()
              ├── watch → WatchCommand.Execute()
              ├── validate → ValidateCommand.Execute()
              ├── mcp → MCPCommand.Execute()
              ├── api → APICommand.Execute()
              ├── completion (built-in)
              └── version
```

---

## State Transitions

### Config Loading Lifecycle

```
[No Config] → ReadInConfig(global) → [Global Loaded]
[Global Loaded] → MergeInConfig(project) → [Merged]
[Merged] → AutomaticEnv → [Env Applied]
[Env Applied] → BindPFlag → [Flags Applied]
[Flags Applied] → Ready
```

### XDG Directory Lifecycle

```
[Not Exists] → first write needed → os.MkdirAll(0o755) → [Created]
[Created] → read/write → [Active]
[Active] → user deletes → [Not Exists] (recreated on next write)
```

---

## TOML Config File Format

### Global config (`~/.config/loko/config.toml`)

```toml
# Global loko configuration
# Values here apply to all projects unless overridden

[d2]
theme = "neutral-default"
layout = "elk"
cache = true

[outputs]
html = true
markdown = false
pdf = false

[build]
parallel = true
max_workers = 4

[server]
serve_port = 8080
api_port = 8081
hot_reload = true
```

### Project config (`./loko.toml`)

```toml
[project]
name = "my-architecture"
description = "System architecture documentation"
version = "1.0.0"
template = "serverless"

[paths]
source = "./src"
output = "./dist"

[d2]
theme = "dark"
```

### Theme file (`~/.local/share/loko/themes/dark.toml`)

```toml
[theme]
name = "dark"
d2_theme = "dark-mauve"

[colors]
primary = "#bb86fc"
secondary = "#03dac6"
background = "#121212"
surface = "#1e1e1e"
error = "#cf6679"

[styles]
stroke_dash = "0"
opacity = "1.0"
```

---

## Environment Variable Mapping

| Viper Key | Env Var | Example |
|-----------|---------|---------|
| `d2.theme` | `LOKO_D2_THEME` | `LOKO_D2_THEME=dark` |
| `d2.layout` | `LOKO_D2_LAYOUT` | `LOKO_D2_LAYOUT=dagre` |
| `d2.cache` | `LOKO_D2_CACHE` | `LOKO_D2_CACHE=false` |
| `output.dir` | `LOKO_OUTPUT_DIR` | `LOKO_OUTPUT_DIR=./build` |
| `build.parallel` | `LOKO_BUILD_PARALLEL` | `LOKO_BUILD_PARALLEL=false` |
| `build.max_workers` | `LOKO_BUILD_MAX_WORKERS` | `LOKO_BUILD_MAX_WORKERS=8` |
| `server.serve_port` | `LOKO_SERVER_SERVE_PORT` | `LOKO_SERVER_SERVE_PORT=3000` |

Special path env vars (not Viper-managed):

| Env Var | Purpose |
|---------|---------|
| `LOKO_CONFIG_HOME` | Override config directory path |
| `XDG_CONFIG_HOME` | Standard XDG config path |
| `XDG_DATA_HOME` | Standard XDG data path |
| `XDG_CACHE_HOME` | Standard XDG cache path |
