# Tasks: Cobra & Viper CLI Migration

**Input**: Design documents from `/specs/002-cobra-viper-cli/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks grouped by user story. User stories US1 and US2 are both P1 but independent.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Add dependencies and create root command foundation

- [x] T001 Add spf13/cobra v1.10.2 and spf13/viper v1.21.0 dependencies to go.mod
- [x] T002 Create root command with persistent flags (--config, --project, --verbose) and basic initConfig() stub in cmd/root.go

---

## Phase 2: Foundational (Core Cobra Migration)

**Purpose**: Migrate all existing commands from manual flag parsing to Cobra. MUST complete before any user story.

**CRITICAL**: No user story work can begin until this phase is complete.

### Core Infrastructure

- [x] T003 Add PathResolver and ThemeLoader interfaces and extend ConfigLoader with LoadGlobalConfig/SaveGlobalConfig in internal/core/usecases/ports.go
- [x] T004 [P] Create XDGPaths value object with ConfigHome/DataHome/CacheHome resolution and validation in internal/core/entities/xdg.go
- [x] T005 [P] Create Theme entity with Name, Path, D2Theme, Colors, Styles fields and validation in internal/core/entities/theme.go
- [x] T006 Implement XDG path resolver adapter (PathResolver interface) with env var fallbacks and EnsureDir() in internal/adapters/config/paths.go

### Command Migration

- [x] T007 Simplify main.go to ~10 lines: import cmd package, call cmd.Execute(), pass version/commit/date via ldflags
- [x] T008 [P] Migrate init command to Cobra: create cobra.Command with --description, --path flags wrapping existing InitCommand in cmd/init_cobra.go
- [x] T009 Create new parent command (no RunE, just Groups subcommands) with system/container/component subcommands in cmd/new_cobra.go
- [x] T010 [P] Extract new system subcommand to Cobra with --description, --technology, --template flags in cmd/new_cobra.go
- [x] T011 [P] Extract new container subcommand to Cobra with --description, --technology, --parent, --template flags in cmd/new_cobra.go
- [x] T012 [P] Extract new component subcommand to Cobra with --description, --technology, --parent, --template flags in cmd/new_cobra.go
- [x] T013 [P] Migrate build command to Cobra with --clean, --output, --format, --d2-theme, --d2-layout flags in cmd/build_cobra.go
- [x] T014 [P] Migrate serve command to Cobra with --output, --address, --port flags in cmd/serve_cobra.go
- [x] T015 [P] Migrate watch command to Cobra with --output, --debounce flags in cmd/watch_cobra.go
- [x] T016 [P] Migrate validate command to Cobra (persistent flags only) in cmd/validate_cobra.go
- [x] T017 [P] Migrate mcp command to Cobra with --env flag in cmd/mcp_cobra.go
- [x] T018 [P] Migrate api command to Cobra (persistent flags only) in cmd/api_cobra.go
- [x] T019 Create version command displaying version/commit/date/builtBy in cmd/version_cobra.go
- [x] T020 Verify backward compatibility: all existing command syntax works identically (loko init, new, build, serve, watch, validate, mcp --help)

**Checkpoint**: Cobra-based CLI is fully functional with all existing commands. No config hierarchy or completions yet.

---

## Phase 3: User Story 1 - Shell Completions (Priority: P1) MVP

**Goal**: Enable tab completion in bash, zsh, fish, and PowerShell for all commands and flags.

**Independent Test**: Run `loko completion bash | source /dev/stdin` then `loko <TAB>` shows all commands; `loko build --<TAB>` shows flags.

### Implementation for User Story 1

- [x] T021 [US1] Configure and customize Cobra's built-in completion command in cmd/completion_cobra.go
- [x] T022 [US1] Add ValidArgsFunction to new parent command returning system/container/component with descriptions in cmd/new_cobra.go
- [x] T023 [P] [US1] Add RegisterFlagCompletionFunc for --template flag listing available templates from filesystem in cmd/new_cobra.go
- [x] T024 [P] [US1] Add RegisterFlagCompletionFunc for --d2-theme (list themes) and --d2-layout (dagre/elk/tala) in cmd/build_cobra.go
- [x] T025 [P] [US1] Add RegisterFlagCompletionFunc for --format flag (html/markdown/pdf) in cmd/build_cobra.go
- [x] T026 [US1] Add dynamic --parent flag completion: list systems for new container, list containers for new component in cmd/new_cobra.go

**Checkpoint**: `loko completion bash` generates valid script. Tab completion works for all commands, subcommands, and flags.

---

## Phase 4: User Story 2 - Persistent Configuration (Priority: P1)

**Goal**: Hierarchical config: CLI flags > LOKO_* env vars > project loko.toml > global XDG config.toml > defaults.

**Independent Test**: Set `d2.theme = "dark"` in `~/.config/loko/config.toml`, run `loko build` — dark theme is used. Override with `LOKO_D2_THEME=light` — light theme used. Override with `--d2-theme terminal` — terminal theme used.

### Implementation for User Story 2

- [x] T027 [US2] Migrate config loader from BurntSushi/toml to Viper-based loading in internal/adapters/config/loader.go
- [x] T028 [US2] Implement full initConfig() in cmd/root.go: SetDefaults → ReadInConfig(global XDG path) → MergeInConfig(project loko.toml) → SetEnvPrefix("LOKO") → SetEnvKeyReplacer("."→"_") → AutomaticEnv()
- [x] T029 [US2] Implement LoadGlobalConfig and SaveGlobalConfig methods using PathResolver in internal/adapters/config/loader.go
- [x] T030 [US2] Add BindPFlag calls in each command's init() to bind Cobra flags to Viper keys (d2.theme, d2.layout, output.dir, etc.) in cmd/build_cobra.go, cmd/serve_cobra.go
- [x] T031 [US2] Wire Viper config values into existing command structs — replace flag defaults with viper.GetString calls in cmd/build_cobra.go
- [x] T032 [US2] Implement theme loader adapter (LoadTheme, ListThemes) reading TOML files from ThemesDir() in internal/adapters/config/theme.go
- [x] T033 [US2] Remove BurntSushi/toml dependency from go.mod and update all imports
- [x] T034 [US2] Handle config edge cases: invalid TOML (Viper error), missing config file (silent fallback), --config pointing to nonexistent path (clear error)

**Checkpoint**: Full config hierarchy works. Global config at `~/.config/loko/config.toml` is loaded, project `loko.toml` overrides it, `LOKO_*` env vars override both, `--flag` overrides everything.

---

## Phase 5: User Story 5 - Improved Help System (Priority: P2)

**Goal**: Contextual, grouped help with examples and typo suggestions.

**Independent Test**: Run `loko help` — commands grouped by category. Run `loko buidl` — suggests "Did you mean 'build'?"

### Implementation for User Story 5

- [x] T035 [US5] Configure command groups (Scaffolding, Building, Serving) via rootCmd.AddGroup() in cmd/root.go
- [x] T036 [US5] Set GroupID on all commands: init+new → "scaffolding", build+watch+validate → "building", serve+api+mcp → "serving" in each cmd/*_cobra.go file
- [x] T037 [P] [US5] Add Example field with usage examples to each Cobra command definition in cmd/init_cobra.go, cmd/build_cobra.go, cmd/serve_cobra.go, cmd/watch_cobra.go, cmd/validate_cobra.go
- [x] T038 [US5] Verify Cobra's built-in typo suggestion works (SuggestionsMinimumDistance defaults to 2) — verified: "loko buidl" suggests "build"

**Checkpoint**: `loko help` shows grouped commands. `loko buidl` suggests correction. Each `--help` shows examples.

---

## Phase 6: User Story 3 - Nested Subcommands (Priority: P2)

**Goal**: Support 2-level deep subcommands for future extensibility (export, graph).

**Independent Test**: Run `loko export --help` — shows html/markdown/pdf subcommands. Run `loko export html` — executes HTML export.

### Implementation for User Story 3

- [x] T039 [P] [US3] Create export parent command (no RunE, shows help when run alone) in cmd/export_cobra.go
- [x] T040 [P] [US3] Create export html subcommand wrapping existing HTML build logic in cmd/export_cobra.go
- [x] T041 [P] [US3] Create export markdown subcommand wrapping existing markdown build logic in cmd/export_cobra.go
- [x] T042 [P] [US3] Create export pdf subcommand wrapping existing PDF build logic in cmd/export_cobra.go
- [x] T043 [US3] Verify nested subcommands work with shell completions — `loko export <TAB>` shows html/markdown/pdf

**Checkpoint**: `loko export html`, `loko export markdown`, `loko export pdf` all work. 2-level subcommands functional.

---

## Phase 7: User Story 4 - Command Aliases (Priority: P3)

**Goal**: Short aliases for common commands (e.g., `loko b` for `loko build`).

**Independent Test**: Run `loko b` — executes build. Run `loko help b` — shows build help.

### Implementation for User Story 4

- [x] T044 [US4] Add Aliases field to Cobra command definitions: b→build, n→new, val→validate, s→serve, w→watch in cmd/build_cobra.go, cmd/new_cobra.go, cmd/validate_cobra.go, cmd/serve_cobra.go, cmd/watch_cobra.go
- [x] T045 [US4] Load custom alias overrides from config file [aliases] section in cmd/root.go
- [x] T046 [US4] Verify aliases appear in help output alongside full command names

**Checkpoint**: `loko b` runs build. `loko --help` shows aliases.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Cleanup, verification, documentation

- [ ] T047 [P] Update README.md with new CLI usage, shell completion setup instructions, and XDG config paths (skipped — docs update deferred to user request)
- [x] T048 Run go vet and golangci-lint, fix any issues across all modified files
- [x] T049 Verify NFR-001: CLI startup time < 50ms with `time loko --version` (30ms first run)
- [x] T050 Run quickstart.md validation — verify all listed commands produce expected output

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 Shell Completions (Phase 3)**: Depends on Foundational (all Cobra commands must exist)
- **US2 Persistent Config (Phase 4)**: Depends on Foundational (XDG paths, Viper init)
- **US5 Improved Help (Phase 5)**: Depends on Foundational (command groups require Cobra commands)
- **US3 Nested Subcommands (Phase 6)**: Depends on Foundational; benefits from US1 (completions) being done
- **US4 Command Aliases (Phase 7)**: Depends on Foundational; can start anytime after Phase 2
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Independent — only needs Cobra commands from Phase 2
- **US2 (P1)**: Independent — only needs XDG paths and root.go from Phase 2
- **US5 (P2)**: Independent — only needs Cobra commands from Phase 2
- **US3 (P2)**: Independent — creates new commands, benefits from US1 completions
- **US4 (P3)**: Independent — adds aliases to existing commands

### Within Each User Story

- Core implementation before integration
- Each story testable independently after completion
- Commit after each task or logical group

### Parallel Opportunities

**Phase 2 (Foundational)**: T008-T018 can ALL run in parallel (different files, each wraps one command)

**Phase 3 (US1)**: T023, T024, T025 can run in parallel (different command files)

**Phase 4 (US2)**: T027→T028→T030→T031 must be sequential (Viper setup flows through); T032 is independent

**Phase 5 (US5)**: T037 items can run in parallel across command files

**Phase 6 (US3)**: T039, T040, T041, T042 can ALL run in parallel (independent files)

**Cross-story**: US1, US2, US5 can ALL start simultaneously after Phase 2 completes

---

## Parallel Example: Phase 2 Command Migration

```
# All command migrations are independent (different files):
T008: Migrate init command → cmd/init.go
T010: Extract new system → cmd/new_system.go
T011: Extract new container → cmd/new_container.go
T012: Extract new component → cmd/new_component.go
T013: Migrate build → cmd/build.go
T014: Migrate serve → cmd/serve.go
T015: Migrate watch → cmd/watch.go
T016: Migrate validate → cmd/validate.go
T017: Migrate mcp → cmd/mcp.go
T018: Migrate api → cmd/api.go
```

## Parallel Example: After Phase 2 Completes

```
# Three P1/P2 stories can start simultaneously:
Developer A: US1 (Shell Completions) → T021-T026
Developer B: US2 (Persistent Config) → T027-T034
Developer C: US5 (Improved Help) → T035-T038
```

---

## Implementation Strategy

### MVP First (Phase 1 + 2 + US1)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational — all commands migrated to Cobra (T003-T020)
3. Complete Phase 3: US1 Shell Completions (T021-T026)
4. **STOP and VALIDATE**: `loko completion bash` works, all commands have tab completion
5. This is the minimum shippable increment — backward compatible with completions

### Incremental Delivery

1. Setup + Foundational → Working Cobra CLI (backward compatible)
2. Add US1 (Completions) → Ship MVP
3. Add US2 (Config hierarchy) → XDG paths, env vars, multi-file config
4. Add US5 (Help) → Grouped commands, typo suggestions
5. Add US3 (Nested subcommands) → Export commands
6. Add US4 (Aliases) → Power user convenience
7. Polish → Docs, lint, perf verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- BurntSushi/toml removal (T033) should be last config-related task to avoid breaking intermediate states
