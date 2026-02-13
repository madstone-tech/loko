# Feature Specification: Cobra & Viper CLI Migration

**Feature Branch**: `002-cobra-viper-cli`
**Created**: 2026-02-05
**Status**: Draft
**Spec Version**: 0.1.0

---

## Overview

Migrate loko's CLI from manual flag parsing to the Cobra/Viper framework to enable a professional-grade command-line experience with shell completions, hierarchical configuration, and extensible command structure.

---

## Clarifications

### Session 2026-02-05

- Q: Which loko artifacts map to which XDG directories? → A: `XDG_CONFIG_HOME/loko/` for config.toml; `XDG_DATA_HOME/loko/themes/` for themes; `XDG_CACHE_HOME/loko/` for build cache
- Q: How should loko handle macOS platform paths? → A: XDG paths on all platforms (`~/.config/loko/`, `~/.local/share/loko/`, `~/.cache/loko/`)
- Q: How should config path override precedence work? → A: `--config` flag > `LOKO_CONFIG_HOME` env > `XDG_CONFIG_HOME/loko/` > `~/.config/loko/` for path resolution; value hierarchy unchanged
- Q: What is a theme concretely? → A: Single TOML file per theme (e.g., `themes/dark.toml`) with D2 style overrides; designed to later support TOML includes/imports for structured composition
- Q: What happens on first run when XDG dirs don't exist? → A: Lazy creation — only create a directory when loko first needs to write to it

---

## User Scenarios & Testing

### User Story 1 - Shell Completions (Priority: P1)

As a developer using loko daily, I want tab completion in my shell so I can discover commands and options without memorizing them or consulting documentation.

**Why this priority**: Shell completions are the most visible user-facing improvement and dramatically improve daily workflow efficiency. This is table-stakes for modern CLI tools.

**Independent Test**: User can type `loko <TAB>` in bash/zsh/fish and see all available commands; `loko build --<TAB>` shows available flags

**Acceptance Scenarios**:

1. **Given** loko is installed, **When** user runs `loko completion bash` and sources the output, **Then** tab completion works for all commands and flags
2. **Given** completion is configured, **When** user types `loko n<TAB>`, **Then** shell completes to `loko new`
3. **Given** completion is configured, **When** user types `loko new <TAB>`, **Then** shell shows `system`, `container`, `component` options
4. **Given** completion is configured, **When** user types `loko build --<TAB>`, **Then** shell shows available flags like `--clean`, `--output`, `--project`

---

### User Story 2 - Persistent Configuration (Priority: P1)

As a developer working on multiple loko projects, I want global defaults that apply across all projects (like preferred D2 theme) while allowing per-project overrides, so I don't have to specify common options repeatedly.

**Why this priority**: Reduces repetitive flag usage and enables team-wide consistency through shared configuration files.

**Independent Test**: User can set `d2.theme = "dark"` in `$XDG_CONFIG_HOME/loko/config.toml` (default `~/.config/loko/config.toml`) and all projects use dark theme unless overridden locally

**Acceptance Scenarios**:

1. **Given** global config exists at `$XDG_CONFIG_HOME/loko/config.toml`, **When** user runs `loko build` without flags, **Then** global settings are applied
2. **Given** both global and project `loko.toml` exist, **When** they have conflicting values, **Then** project config takes precedence
3. **Given** a flag is passed on command line, **When** it conflicts with config file, **Then** command-line flag takes precedence
4. **Given** environment variable `LOKO_D2_THEME` is set, **When** user runs loko, **Then** env var overrides config file but not CLI flags

---

### User Story 3 - Nested Subcommands (Priority: P2)

As a power user, I want logically grouped commands (like `loko graph analyze`, `loko export pdf`) so the CLI remains intuitive as features grow.

**Why this priority**: Enables clean organization for v0.2.0+ features (API, export formats, graph analysis) without cluttering the top-level command space.

**Independent Test**: User can run `loko export pdf` and `loko export markdown` as distinct subcommands under a shared `export` parent

**Acceptance Scenarios**:

1. **Given** export command exists, **When** user runs `loko export --help`, **Then** shows available export subcommands (html, markdown, pdf)
2. **Given** graph commands exist, **When** user runs `loko graph analyze`, **Then** runs architecture graph analysis
3. **Given** nested command, **When** user runs `loko export` without subcommand, **Then** shows help with available subcommands

---

### User Story 4 - Command Aliases (Priority: P3)

As a frequent user, I want short aliases for common commands (e.g., `loko b` for `loko build`) so I can work faster.

**Why this priority**: Nice-to-have convenience that reduces keystrokes for power users without complicating the interface.

**Independent Test**: User can run `loko b` and it executes `loko build`

**Acceptance Scenarios**:

1. **Given** alias `b` is defined for `build`, **When** user runs `loko b`, **Then** executes build command
2. **Given** alias exists, **When** user runs `loko help b`, **Then** shows build command help
3. **Given** multiple aliases exist, **When** user runs `loko --help`, **Then** aliases are shown alongside full command names

---

### User Story 5 - Improved Help System (Priority: P2)

As a new user, I want contextual, well-organized help that shows examples and groups related options, so I can learn the tool quickly.

**Why this priority**: Good documentation reduces support burden and improves adoption. Auto-generated help ensures it stays current.

**Independent Test**: User runs `loko build --help` and sees grouped flags with examples

**Acceptance Scenarios**:

1. **Given** any command, **When** user runs `loko <cmd> --help`, **Then** shows description, usage, flags grouped by category, and examples
2. **Given** user runs `loko help`, **Then** shows all commands organized by category (Scaffolding, Building, Integration)
3. **Given** user makes a typo, **When** running `loko buidl`, **Then** suggests "Did you mean 'build'?"

---

### Edge Cases

- What happens when config file has invalid TOML syntax? → Clear error message with line number
- What happens when environment variable has invalid value? → Warning message, falls back to default
- What happens when user has very old shell without completion support? → Graceful degradation, completions simply don't work
- What happens when global and local config define the same nested key differently? → Deep merge with local taking precedence
- What happens when XDG environment variables are not set? → Fall back to XDG defaults (`~/.config/`, `~/.local/share/`, `~/.cache/`)
- What happens when `--config` flag points to a nonexistent path? → Error with clear message showing the resolved path
- What happens on first run with no XDG dirs? → Directories created lazily on first write (no pre-creation)
- What happens when config is read but no config file exists? → Silent fallback to defaults (no error)

---

## Requirements

### Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-001 | System MUST provide shell completion scripts for bash, zsh, fish, and PowerShell | P1 |
| FR-002 | System MUST support configuration hierarchy: CLI flags > `LOKO_*` env vars (values) > project config > global config > defaults. Config path resolution: `--config` flag > `LOKO_CONFIG_HOME` env > `XDG_CONFIG_HOME/loko/` > `~/.config/loko/` | P1 |
| FR-003 | System MUST load global configuration from `$XDG_CONFIG_HOME/loko/config.toml` (default `~/.config/loko/config.toml`), themes from `$XDG_DATA_HOME/loko/themes/` (default `~/.local/share/loko/themes/`), and build cache from `$XDG_CACHE_HOME/loko/` (default `~/.cache/loko/`). Paths overridable by `--config` flag and `LOKO_CONFIG_HOME` env var. | P1 |
| FR-004 | System MUST support environment variables with `LOKO_` prefix (e.g., `LOKO_D2_THEME`) | P1 |
| FR-005 | System MUST support nested subcommands (at least 2 levels deep) | P2 |
| FR-006 | System MUST provide command aliases configurable via config file | P3 |
| FR-007 | System MUST auto-generate help text from command definitions | P1 |
| FR-008 | System MUST group flags by category in help output | P2 |
| FR-009 | System MUST suggest corrections for mistyped commands | P2 |
| FR-010 | System MUST maintain backward compatibility with existing command syntax | P1 |
| FR-011 | System MUST support persistent flags that apply to all subcommands (e.g., `--verbose`, `--project`) | P1 |
| FR-012 | System MUST provide `loko completion <shell>` command to generate completion scripts | P1 |

### Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-001 | CLI startup time | < 50ms (no regression from current) |
| NFR-002 | Completion script generation | < 100ms |
| NFR-003 | Config file parsing | < 10ms |
| NFR-004 | Memory overhead vs current | < 5MB additional |

### Key Entities

- **Command**: A CLI action with name, description, flags, and execution logic
- **Flag**: A command-line option with name, shorthand, type, default, and description
- **Config**: Hierarchical configuration with global, project, and runtime layers
- **Theme**: A single TOML file in `$XDG_DATA_HOME/loko/themes/` containing D2 style overrides and color values (e.g., `dark.toml`). Future: support TOML includes/imports for structured composition
- **Completion**: Shell-specific script that enables tab completion

---

## Success Criteria

### Measurable Outcomes

| ID | Criterion | Target |
|----|-----------|--------|
| SC-001 | Shell completion available for major shells | bash, zsh, fish, PowerShell |
| SC-002 | All existing commands work identically | 100% backward compatibility |
| SC-003 | Help text accuracy | Auto-generated, always in sync with code |
| SC-004 | Config hierarchy works correctly | All 4 layers respected in correct order |
| SC-005 | New command addition effort | < 30 lines of code per command |
| SC-006 | CLI startup time | No regression (< 50ms) |

---

## Scope & Exclusions

### In Scope

- Migration of all existing commands to Cobra
- Viper integration for configuration management
- Shell completion generation
- Persistent flags (`--project`, `--verbose`)
- Command grouping in help
- Typo suggestions

### Out of Scope (Future)

- Interactive prompts/wizards (v0.3.0+)
- Plugin system for custom commands (v1.0.0+)
- Remote configuration loading
- Configuration encryption
- TOML includes/imports for theme composition (future enhancement to theme system)

---

## Dependencies & Assumptions

### Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management

### Assumptions

- Users have standard shells (bash 4+, zsh 5+, fish 3+, PowerShell 5+)
- Existing `loko.toml` format is compatible with Viper TOML parsing
- No breaking changes to existing command-line interface
- XDG paths are used on all platforms (Linux, macOS, Windows WSL); no platform-specific path conventions

---

## External References

- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- [Viper Documentation](https://github.com/spf13/viper#readme)
- [12-Factor App Config](https://12factor.net/config)
- Current loko CLI: `main.go` (~366 lines), `cmd/*.go` (~1500 lines)
