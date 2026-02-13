# Research: Cobra & Viper CLI Migration

**Branch**: `002-cobra-viper` | **Date**: 2026-02-05

---

## Decision 1: CLI Framework

**Decision**: Use `github.com/spf13/cobra` v1.10.2

**Rationale**: Industry standard for Go CLI tools. Provides built-in shell completions, command groups in help, typo suggestions, persistent flags, and flag validation. Compatible with Go 1.25.

**Alternatives considered**:
- `urfave/cli/v2` — Less structured command hierarchy, weaker completion support
- `alecthomas/kong` — Struct-based, less community adoption
- Manual flag parsing (current) — No completions, no command groups, each command reinvents flag handling

---

## Decision 2: Configuration Management

**Decision**: Use `github.com/spf13/viper` v1.21.0

**Rationale**: Native integration with Cobra via `BindPFlag`. Built-in precedence: flags > env vars > config files > defaults. Supports TOML natively (via `pelletier/go-toml/v2`). `MergeInConfig` enables multi-file config merging.

**Alternatives considered**:
- `knadh/koanf` — More modular but no Cobra integration
- `BurntSushi/toml` (current) — No env var support, no flag binding, no multi-file merge
- `kelseyhightower/envconfig` — Env vars only, no file support

**Key patterns**:
- `ReadInConfig()` for global config, `MergeInConfig()` for project-local (deep merge)
- `SetEnvPrefix("LOKO")` + `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` + `AutomaticEnv()`
- `BindPFlag("d2.theme", cmd.Flags().Lookup("d2-theme"))` — flag value only used when `HasChanged()==true`
- Config init in `rootCmd.PersistentPreRunE`

---

## Decision 3: XDG Path Resolution

**Decision**: Manual implementation (~40 lines) in `internal/adapters/config/paths.go`

**Rationale**: loko only needs 3 XDG paths (config, data, cache). Manual implementation has zero external dependencies, fits the clean architecture (`core/` has zero deps), and avoids macOS override hacks needed by third-party libraries.

**Alternatives considered**:
- `adrg/xdg` — Most popular Go XDG library (946 stars), but returns macOS-native paths (`~/Library/Application Support`), requires runtime.GOOS override hack
- `zchee/go-xdgbasedir` — Has Unix mode that returns `~/.config` on macOS, but stale (last updated 2018, v1.0.3)
- `os.UserConfigDir()` (stdlib) — Returns `~/Library/Application Support` on macOS, only covers config + cache (no data dir)

**Path resolution**:

| Path | Env Override | Default (all platforms) |
|------|-------------|------------------------|
| Config | `$LOKO_CONFIG_HOME` > `$XDG_CONFIG_HOME/loko/` | `~/.config/loko/` |
| Data | `$XDG_DATA_HOME/loko/` | `~/.local/share/loko/` |
| Cache | `$XDG_CACHE_HOME/loko/` | `~/.cache/loko/` |

Flag override: `--config` flag overrides config path entirely.

---

## Decision 4: BurntSushi/toml Removal

**Decision**: Remove `BurntSushi/toml` dependency, rely on Viper for TOML parsing

**Rationale**: Viper uses `pelletier/go-toml/v2` internally. Having two TOML parsers is redundant. Viper handles all config reading/writing needs.

**Alternatives considered**:
- Keep both — Unnecessary dependency, divergent behavior risk
- Note: `SaveConfig` currently uses `BurntSushi/toml` encoder with comments. Viper's `WriteConfig` also supports TOML output. If custom comment formatting is needed, can use `pelletier/go-toml/v2` directly (already a transitive dep).

---

## Decision 5: Command File Layout

**Decision**: Flat layout in `cmd/` with `new_system.go`, `new_container.go`, `new_component.go` sub-files

**Rationale**: loko has <20 commands. Flat layout keeps navigation simple. Cobra convention uses `init()` for registration.

**Alternatives considered**:
- Nested packages (`cmd/new/system.go`) — Overkill for current command count, adds import complexity

**Structure**:
```
cmd/
├── root.go              # Root command, persistent flags, config init
├── init.go              # loko init
├── new.go               # loko new (parent)
├── new_system.go        # loko new system
├── new_container.go     # loko new container
├── new_component.go     # loko new component
├── build.go             # loko build
├── serve.go             # loko serve
├── watch.go             # loko watch
├── validate.go          # loko validate
├── completion.go        # loko completion (customize built-in)
├── mcp.go               # loko mcp
├── api.go               # loko api
└── version.go           # loko version (or --version on root)
```

---

## Decision 6: Config Init Location

**Decision**: Use `rootCmd.PersistentPreRunE` for Viper initialization

**Rationale**: Cobra's lifecycle guarantees `PersistentPreRunE` runs before any subcommand's `RunE`. This ensures config is loaded for all commands, including flag bindings.

**Pattern**:
1. `PersistentPreRunE` on rootCmd calls `initConfig()`
2. `initConfig()`: set defaults → ReadInConfig (global) → MergeInConfig (project-local) → AutomaticEnv
3. Per-command flag bindings via `BindPFlag` in each command's `init()`

---

## Decision 7: Lazy Directory Creation

**Decision**: Create XDG directories on first write using `os.MkdirAll(path, 0o755)`

**Rationale**: XDG spec says directories should not be created eagerly. CLI tool may never write to data or cache dirs in a given session. `os.MkdirAll` is idempotent.

**Permissions**:
- Directories: `0o755` (rwxr-xr-x, further restricted by umask)
- Config files: `0o644` (rw-r--r--)
- No `sync.Once` needed — CLI tool has sequential initialization

---

## Decision 8: Shell Completion Strategy

**Decision**: Use Cobra's built-in completion command with custom `ValidArgsFunction` for dynamic completions

**Rationale**: Cobra generates correct completion scripts for bash, zsh, fish, and PowerShell. Custom completions needed for `loko new [type]` and flag values like `--d2-theme`.

**Custom completions needed**:
- `loko new <TAB>` → `system`, `container`, `component`
- `loko new container --parent <TAB>` → list systems from project
- `--d2-theme <TAB>` → list available themes from XDG data dir
- `--template <TAB>` → list available templates

---

## Decision 9: Command Groups

**Decision**: Use Cobra's `AddGroup` for help output organization

**Groups**:
- **Scaffolding**: `init`, `new`
- **Building**: `build`, `watch`, `validate`
- **Serving**: `serve`, `api`, `mcp`
- **Other**: `completion`, `version`, `help`
