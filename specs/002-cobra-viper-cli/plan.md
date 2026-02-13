# Implementation Plan: Cobra & Viper CLI Migration

**Branch**: `002-cobra-viper` | **Date**: 2026-02-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-cobra-viper-cli/spec.md`

## Summary

Migrate loko's CLI from manual `flag.FlagSet` parsing (~377 lines in `main.go`) to Cobra/Viper framework. This enables shell completions, hierarchical XDG-compliant configuration (`CLI flags > env vars > project config > global config > defaults`), command groups in help, and typo suggestions. XDG paths used on all platforms with lazy directory creation.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: spf13/cobra v1.10.2, spf13/viper v1.21.0 (new); charmbracelet/lipgloss, fsnotify (existing)
**Storage**: Filesystem — XDG dirs (`~/.config/loko/`, `~/.local/share/loko/`, `~/.cache/loko/`) + project `loko.toml`
**Testing**: `go test`, table-driven tests, `t.Setenv()` for XDG path tests
**Target Platform**: Linux, macOS, Windows WSL (XDG paths on all)
**Project Type**: Single CLI binary
**Performance Goals**: CLI startup < 50ms, config parse < 10ms, completion gen < 100ms
**Constraints**: < 5MB additional memory, 100% backward compatibility with existing CLI syntax
**Scale/Scope**: 9 existing commands, ~1900 lines in cmd/ + main.go

## Constitution Check

*GATE: Constitution is an unfilled template — no project-specific gates defined. Proceeding with clean architecture principles from CLAUDE.md.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Clean Architecture (core imports nothing from adapters) | PASS | XDG paths implemented as adapter, new ports added to core |
| Interface-First (deps through ports.go) | PASS | PathResolver, ThemeLoader added as ports; ConfigLoader extended |
| Thin Handlers (cmd < 50 lines) | PASS | Cobra commands wire deps and delegate to use cases |
| Entity Validation (in entities, not use cases) | PASS | Theme validation in entities, config validation in entities |

## Project Structure

### Documentation (this feature)

```text
specs/002-cobra-viper-cli/
├── plan.md              # This file
├── research.md          # Phase 0: Cobra/Viper/XDG research
├── data-model.md        # Phase 1: Entity and config data models
├── quickstart.md        # Phase 1: Migration quickstart guide
├── contracts/
│   ├── cli-commands.md  # Phase 1: Full CLI command contract
│   └── ports.md         # Phase 1: Interface changes
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
main.go                          # Simplified to ~10 lines: cmd.Execute()
cmd/
├── root.go                      # NEW: Root command, persistent flags, initConfig()
├── init.go                      # MODIFIED: Cobra command wrapping existing InitCommand
├── new.go                       # MODIFIED: Parent command for new subcommands
├── new_system.go                # NEW: loko new system (extracted from new.go)
├── new_container.go             # NEW: loko new container (extracted from new.go)
├── new_component.go             # NEW: loko new component (extracted from new.go)
├── build.go                     # MODIFIED: Cobra command wrapping existing BuildCommand
├── serve.go                     # MODIFIED: Cobra command wrapping existing ServeCommand
├── watch.go                     # MODIFIED: Cobra command wrapping existing WatchCommand
├── validate.go                  # MODIFIED: Cobra command wrapping existing ValidateCommand
├── mcp.go                       # MODIFIED: Cobra command wrapping existing MCPCommand
├── api.go                       # MODIFIED: Cobra command wrapping existing APICommand
├── d2_generator.go              # MODIFIED: Cobra command (if exposed as CLI)
├── completion.go                # NEW: Customize built-in completion command
└── version.go                   # NEW: loko version command

internal/
├── core/
│   ├── entities/
│   │   ├── project.go           # MODIFIED: Add Theme entity
│   │   └── xdg.go               # NEW: XDGPaths value object
│   └── usecases/
│       └── ports.go             # MODIFIED: Add PathResolver, ThemeLoader; extend ConfigLoader
├── adapters/
│   ├── config/
│   │   ├── loader.go            # MODIFIED: Viper-based, XDG-aware config loading
│   │   ├── paths.go             # NEW: XDG path resolver (implements PathResolver)
│   │   └── theme.go             # NEW: Theme loader (implements ThemeLoader)
│   └── ...                      # (unchanged adapters)
└── ...

go.mod                           # MODIFIED: Add cobra, viper; remove BurntSushi/toml
go.sum                           # MODIFIED: Updated checksums
```

**Structure Decision**: Flat `cmd/` layout with `new_*.go` sub-files for the `new` subcommand hierarchy. Existing command structs (InitCommand, BuildCommand, etc.) remain in `cmd/` — Cobra commands are thin wrappers that create and execute them.

## Complexity Tracking

| Concern | Resolution |
|---------|-----------|
| Viper replaces BurntSushi/toml | Viper uses pelletier/go-toml/v2 internally; config format stays identical |
| XDG path resolver in core vs adapters | Value object `XDGPaths` in entities (pure, no I/O); resolver adapter in adapters/config |
| Cobra init() pattern vs clean architecture | Cobra commands in `cmd/` call existing command structs — no business logic in Cobra layer |
