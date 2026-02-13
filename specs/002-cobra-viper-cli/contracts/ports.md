# Port Interface Changes

**Branch**: `002-cobra-viper` | **Date**: 2026-02-05

---

## Modified Interfaces

### ConfigLoader (in `internal/core/usecases/ports.go`)

Current:
```go
type ConfigLoader interface {
    LoadConfig(ctx context.Context, projectRoot string) (*entities.ProjectConfig, error)
    SaveConfig(ctx context.Context, projectRoot string, config *entities.ProjectConfig) error
}
```

Updated (adds XDG-aware path resolution):
```go
type ConfigLoader interface {
    LoadConfig(ctx context.Context, projectRoot string) (*entities.ProjectConfig, error)
    SaveConfig(ctx context.Context, projectRoot string, config *entities.ProjectConfig) error
    LoadGlobalConfig(ctx context.Context) (*entities.ProjectConfig, error)
    SaveGlobalConfig(ctx context.Context, config *entities.ProjectConfig) error
}
```

### No changes to other existing ports

All 17 remaining interfaces in `ports.go` are unchanged. The Cobra migration only affects the CLI layer (`cmd/`, `main.go`) and the config adapter.

---

## New Interfaces

### PathResolver (new port in `internal/core/usecases/ports.go`)

```go
// PathResolver resolves XDG-compliant paths for application data.
type PathResolver interface {
    ConfigDir() string   // $XDG_CONFIG_HOME/loko/ or ~/.config/loko/
    DataDir() string     // $XDG_DATA_HOME/loko/ or ~/.local/share/loko/
    CacheDir() string    // $XDG_CACHE_HOME/loko/ or ~/.cache/loko/
    ConfigFile() string  // ConfigDir()/config.toml
    ThemesDir() string   // DataDir()/themes/
}
```

### ThemeLoader (new port in `internal/core/usecases/ports.go`)

```go
// ThemeLoader loads and lists available themes.
type ThemeLoader interface {
    LoadTheme(ctx context.Context, name string) (*entities.Theme, error)
    ListThemes(ctx context.Context) ([]string, error)
}
```

---

## Adapter Implementations

### `internal/adapters/config/paths.go` (new)

Implements `PathResolver` with XDG resolution logic:
- Checks `LOKO_CONFIG_HOME` → `XDG_CONFIG_HOME/loko/` → `~/.config/loko/`
- Checks `XDG_DATA_HOME/loko/` → `~/.local/share/loko/`
- Checks `XDG_CACHE_HOME/loko/` → `~/.cache/loko/`
- Provides `EnsureDir()` for lazy directory creation

### `internal/adapters/config/loader.go` (modified)

Updated to:
- Accept `PathResolver` dependency
- Use Viper for TOML parsing instead of `BurntSushi/toml`
- Implement `LoadGlobalConfig` / `SaveGlobalConfig`
- Support `ReadInConfig` + `MergeInConfig` pattern

### `internal/adapters/config/theme.go` (new)

Implements `ThemeLoader`:
- Reads TOML theme files from `PathResolver.ThemesDir()`
- Lists available themes by scanning directory
