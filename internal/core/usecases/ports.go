package usecases

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// Logger is the logging port. Implemented by internal/adapters/logging.
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, err error, keysAndValues ...any)
	WithContext(ctx context.Context) Logger
	WithFields(keysAndValues ...any) Logger
}

// OutputEncoder serialises values for machine consumption. Implemented by
// internal/adapters/encoding.
type OutputEncoder interface {
	EncodeJSON(value any) ([]byte, error)
	EncodeTOON(value any) ([]byte, error)
	DecodeJSON(data []byte, value any) error
	DecodeTOON(data []byte, value any) error
}

// ---------------------------------------------------------------------------
// v1 compiler ports (feature 013-hcl-compiler-core)
// ---------------------------------------------------------------------------

// ArchitectureSource discovers and parses the authored architecture, returning
// an unresolved arch.SourceModel plus any syntax-level diagnostics.
//
// The contract is deliberately narrow: an implementation reports what it could
// not parse — unreadable files, invalid syntax, unknown blocks, unknown
// attributes, unknown functions — and nothing else. Reference resolution and
// every semantic rule belong to the use-case layer, which is why references
// come back as raw strings in arch.Reference rather than resolved addresses
// (research R2 and R7).
//
// Implementations MUST:
//   - return a SourceModel even when diagnostics contain errors, so the
//     compiler can keep validating whatever parsed and report everything in
//     one run (FR-031);
//   - return ranges whose File is project-relative with forward slashes on
//     every platform, so exports stay byte-identical across machines (FR-036c);
//   - reserve the error return for failures that make the whole run
//     impossible, such as an unreadable project root. An individual bad file
//     is a diagnostic, not an error.
type ArchitectureSource interface {
	// Load discovers every architecture source file beneath root and parses
	// them into one SourceModel.
	Load(ctx context.Context, root string) (*arch.SourceModel, arch.Diagnostics, error)
}

// SourceFormatter rewrites authored source into canonical form, preserving
// comments and declaration order (FR-035).
//
// Check mode exists so a continuous-integration job can fail on unformatted
// source without inspecting version-control state (FR-035a).
type SourceFormatter interface {
	// Format rewrites non-canonical files in place and returns the paths it
	// changed, project-relative and sorted.
	Format(ctx context.Context, root string) ([]string, arch.Diagnostics, error)

	// Check reports which files are not canonically formatted without writing
	// anything. The returned paths are project-relative and sorted.
	Check(ctx context.Context, root string) ([]string, arch.Diagnostics, error)
}
