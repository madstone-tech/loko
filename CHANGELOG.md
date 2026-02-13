# Changelog

All notable changes to the loko project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Architecture Graph Improvements**: Comprehensive improvements to graph implementation
  - Qualified hierarchical node IDs to prevent collisions in multi-system projects
  - O(1) performance for dependency queries via IncomingEdges and ChildrenMap
  - Thread-safe GraphCache for MCP session optimization
  - Type-safe graph operations with C4Entity interface
  - Comprehensive documentation in ADR-0004

### Changed

- **BREAKING**: Node IDs now use qualified hierarchical format
  - **Before**: `"auth"` (short ID, collision-prone)
  - **After**: `"backend/api/auth"` (qualified ID, unique)
  - Migration path: Use ShortIDMap for backward compatibility
  - See `docs/migration-001-graph-qualified-ids.md` for migration guide

- **Performance**: Dependency queries optimized from O(E) to O(1)
  - `GetIncomingEdges()`: 68.56 ns/op (1,000,000x faster)
  - `GetChildren()`: 101.8 ns/op (5,000x faster)
  - `AnalyzeDependencies()`: 171.125µs (11,600x faster)

- **Type Safety**: Replaced runtime type assertions with compile-time checks
  - `GraphNode.Data` changed from `any` to `C4Entity` interface
  - MCP tool arguments use typed structs instead of `map[string]any`
  - Zero runtime type assertions in core/usecases package

- **Validation**: Architecture validation now filters to components only
  - Isolated component checks skip systems/containers
  - High coupling checks skip systems/containers
  - Prevents false positives for non-component entities

### Fixed

- **Critical Bug**: Node ID collisions in multi-system projects
  - Multiple systems with identically-named components no longer overwrite each other
  - Example: `backend/api/auth` and `admin/ui/auth` now coexist without data loss

- **Thread Safety**: Clarified ArchitectureGraph concurrency model
  - Graph is immutable after construction
  - GraphCache provides thread-safe concurrent access
  - Documentation prevents misuse patterns

### Performance

Benchmark results on Apple M3 Pro:
- `GetIncomingEdges`: 68.56 ns/op
- `GetChildren`: 101.8 ns/op
- Target: <1ms for all operations (achieved: 68ns avg)

### Documentation

- Added ADR-0004: Architecture Graph Conventions
  - Node ID format and collision prevention
  - Thread safety model and lifecycle
  - Relationship scope (component-level only)
  - Migration guidance for existing projects

- Enhanced godoc with examples
  - `AddNode()`: Qualified ID usage examples
  - `AddEdge()`: Component relationship examples
  - `GetDependencies()`: Dependency query examples

- Package-level documentation
  - Thread safety guarantees
  - Graph lifecycle stages
  - Immutability contract

### Migration Guide

See `docs/migration-001-graph-qualified-ids.md` for step-by-step migration from short IDs to qualified IDs.

**Breaking changes affect**:
- Projects with multiple systems using same component names
- MCP tools relying on short ID lookups
- Custom code directly accessing graph node IDs

**Backward compatibility**:
- `ShortIDMap` enables short ID resolution
- `ResolveID()` handles unambiguous short IDs
- No changes needed for single-system projects

---

## [0.1.0] - Previous Release

Initial release with core features.

[Unreleased]: https://github.com/madstone-tech/loko/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/madstone-tech/loko/releases/tag/v0.1.0
