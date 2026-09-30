package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
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

// ---------------------------------------------------------------------------
// Render ports (feature 014-viewmodel-renderers, contracts/ports.md)
// ---------------------------------------------------------------------------

// Backend turns a projection into the bytes of one output format.
//
// A backend MUST read nothing but its arguments — no file system, no IR, no
// source (FR-011); MUST NOT import or call another backend (FR-012); MUST make
// no styling decision of its own (FR-008); and MUST be a pure function of its
// arguments, so the same projection always yields identical bytes (FR-021).
type Backend interface {
	Format() viewmodel.Format
	Render(ctx context.Context, in *viewmodel.Projection, opts RenderOptions) ([]viewmodel.Artifact, error)
}

// RenderOptions carries the non-architectural inputs a backend may need.
type RenderOptions struct {
	// Theme holds the user's overrides, sorted by Name. Only the html backend
	// reads it.
	Theme []viewmodel.ThemeFile
}

// ThemeError is returned by a backend for a malformed theme override. The build
// turns it into a theme_invalid diagnostic naming the file (FR-035).
type ThemeError struct {
	File    string // project-relative, e.g. "templates/partials.gohtml"
	Line    int    // 0 when unknown
	Message string
}

func (e *ThemeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
	}
	return e.File + ": " + e.Message
}

// ProseReader returns the text of a docs reference, resolved against the
// project root with the same containment rule as the docs_not_found warning.
// found=false is not an error (FR-026).
type ProseReader interface {
	ReadProse(ctx context.Context, root, docs string) (text string, found bool, err error)
}

// ThemeSource returns the override files in <root>/templates/, sorted by name.
// A missing directory returns (nil, nil).
type ThemeSource interface {
	LoadTheme(ctx context.Context, root string) ([]viewmodel.ThemeFile, error)
}

// ArtifactStore commits a complete artifact set to an output directory,
// pruning files the previous build owned and nothing else (FR-023). It MUST
// write nothing if it cannot write everything.
type ArtifactStore interface {
	Commit(ctx context.Context, outDir string, artifacts []viewmodel.Artifact, sources []string) (CommitReport, error)
}

// CommitReport says what a commit did. Each list is sorted and relative to the
// output directory.
type CommitReport struct {
	Written, Unchanged, Removed []string
}

// ChangeWatcher signals once per settled burst of changes to the watched set.
// The channel closes when ctx is cancelled.
type ChangeWatcher interface {
	Watch(ctx context.Context, spec WatchSpec) (<-chan struct{}, error)
}

// WatchSpec says what to watch.
type WatchSpec struct {
	Root string
	// ExtraFiles returns further files to watch — the prose referenced by the
	// last good build. It is re-read on every tick.
	ExtraFiles func() []string
	// Exclude lists absolute directories never to watch, such as the output
	// directory (FR-032).
	Exclude []string
}

// PreviewServer publishes build results to connected browsers.
type PreviewServer interface {
	// Publish switches to the ok state and tells browsers to reload.
	Publish(artifacts []viewmodel.Artifact)
	// Fail switches to the error state, in which every page shows the
	// diagnostics text, and tells browsers to reload (FR-030).
	Fail(diagnosticsText string)
}
