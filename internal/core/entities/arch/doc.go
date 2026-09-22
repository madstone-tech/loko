// Package arch holds the v1 compiler's domain model: addresses, diagnostics,
// the unresolved SourceModel produced by the parser, and the resolved IR every
// command and later stage consumes.
//
// It sits under internal/core/entities/ so the entities layer rule applies —
// standard library only, and none of the parsing or rendering libraries
// (FR-044). It is a sub-package rather than part of entities itself because
// the v0 model still occupies that package until it is deleted in the final
// phase of this feature, and two of the names collide (Project, Relationship).
// Once the v0 model is gone this package may be flattened into its parent;
// nothing outside depends on the distinction.
package arch
