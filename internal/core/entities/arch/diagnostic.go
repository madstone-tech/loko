package arch

import (
	"sort"
	"strings"
)

// Severity classifies a diagnostic. Errors suppress artefact production;
// warnings do not, unless strict mode is requested (FR-034, FR-038).
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

// String returns the lowercase name used in machine-readable output. It must
// match the `severity` enum in contracts/diagnostics.schema.json.
func (s Severity) String() string {
	if s == SeverityWarning {
		return "warning"
	}
	return "error"
}

// Exit codes. Exactly three, and no command may introduce a fourth (FR-038):
// `fmt --check` reuses ExitErrors rather than defining its own. Codes 3 and
// above are reserved for later stages (3 = reconcile below minimum coverage).
const (
	ExitSuccess  = 0
	ExitErrors   = 1
	ExitWarnings = 2
)

// Diagnostic codes. These are the stable machine identifiers consumers match
// on; Summary and Detail may be reworded in any release without a schema
// change. The set must stay in step with the `code` enum in
// contracts/diagnostics.schema.json — TestAllCodesAreKnown enforces the count,
// and the rule-coverage test enforces that each has a fixture (SC-004).
const (
	// Errors (FR-028, FR-028a).
	CodeSyntaxError           = "syntax_error"
	CodeNoSourceFound         = "no_source_found"
	CodeDuplicateDeclaration  = "duplicate_declaration"
	CodeUnresolvedReference   = "unresolved_reference"
	CodeWrongReferenceKind    = "wrong_reference_kind"
	CodeWrongParentKind       = "wrong_parent_kind"
	CodeContainmentCycle      = "containment_cycle"
	CodeDuplicateClaim        = "duplicate_claim"
	CodeDuplicateInstanceName = "duplicate_instance_name"
	CodeUnknownBlock          = "unknown_block"
	CodeUnknownAttribute      = "unknown_attribute"
	CodeUnknownFunction       = "unknown_function"
	CodeVersionUnsatisfied    = "version_unsatisfied"

	// Warnings (FR-029).
	CodeOrphanElement    = "orphan_element"
	CodeEmptySystem      = "empty_system"
	CodeMissingDocs      = "missing_docs"
	CodeDocsNotFound     = "docs_not_found"
	CodeUnboundInstance  = "unbound_instance"
	CodeSelfRelationship = "self_relationship"

	// Render stage (feature 014-viewmodel-renderers). Each joins AllCodes in
	// the task that first produces it, so the rule-coverage test never lists a
	// code nothing emits yet.
	CodeOutputPathCollision = "output_path_collision" // error, FR-028
	CodeThemeInvalid        = "theme_invalid"         // error, FR-035
	CodeViewEmpty           = "view_empty"            // warning, FR-005
	CodeViewShadowed        = "view_shadowed"         // warning, FR-004

	// Renames (feature 015-mcp-hcl-authoring): the moved block's rules.
	CodeMovedFromDeclared  = "moved_from_declared"
	CodeMovedToUnresolved  = "moved_to_unresolved"
	CodeMovedDuplicateFrom = "moved_duplicate_from"

	// Rendering attributes (feature 016-rendering-fidelity).
	CodeInvalidAttributeValue = "invalid_attribute_value"
	CodeShapeNotAllowed       = "shape_not_allowed"
	CodeEmptyTitle            = "empty_title"
)

// AllCodes lists every diagnostic code this release can emit, errors first.
// The rule-coverage test walks it to assert that no rule ships without a
// fixture, which is what turns SC-004's "100% of the rule list" into a fact
// rather than a claim.
var AllCodes = []string{
	CodeSyntaxError,
	CodeNoSourceFound,
	CodeDuplicateDeclaration,
	CodeUnresolvedReference,
	CodeWrongReferenceKind,
	CodeWrongParentKind,
	CodeContainmentCycle,
	CodeDuplicateClaim,
	CodeDuplicateInstanceName,
	CodeUnknownBlock,
	CodeUnknownAttribute,
	CodeUnknownFunction,
	CodeVersionUnsatisfied,
	CodeOutputPathCollision,
	CodeMovedFromDeclared,
	CodeMovedToUnresolved,
	CodeMovedDuplicateFrom,
	CodeInvalidAttributeValue,
	CodeShapeNotAllowed,
	CodeEmptyTitle,
	CodeThemeInvalid,
	CodeOrphanElement,
	CodeEmptySystem,
	CodeMissingDocs,
	CodeDocsNotFound,
	CodeUnboundInstance,
	CodeSelfRelationship,
	CodeViewEmpty,
	CodeViewShadowed,
}

// RelatedRange is a secondary location that explains a diagnostic — the other
// declaration in a duplicate, for instance (FR-004, FR-012a).
type RelatedRange struct {
	Message string      `json:"message,omitempty" toon:"message,omitempty"`
	Range   SourceRange `json:"range" toon:"range"`
}

// Diagnostic is one problem found during compilation. Parsing, resolution, and
// validation all produce these, so a single run can report everything it finds
// rather than stopping at the first error (FR-031).
type Diagnostic struct {
	Severity Severity       `json:"severity" toon:"severity"`
	Code     string         `json:"code" toon:"code"`
	Summary  string         `json:"summary" toon:"summary"`
	Detail   string         `json:"detail,omitempty" toon:"detail,omitempty"`
	Address  Address        `json:"address,omitempty" toon:"address,omitempty"`
	Range    SourceRange    `json:"range" toon:"range"`
	Related  []RelatedRange `json:"related,omitempty" toon:"related,omitempty"`
}

// Diagnostics is a collection with a defined output order. The zero value is
// usable: an empty set has no errors and exits successfully.
type Diagnostics []Diagnostic

// HasErrors reports whether any diagnostic is an error.
func (d Diagnostics) HasErrors() bool {
	for _, diag := range d {
		if diag.Severity == SeverityError {
			return true
		}
	}
	return false
}

// CountBySeverity returns how many diagnostics carry the given severity.
func (d Diagnostics) CountBySeverity(s Severity) int {
	n := 0
	for _, diag := range d {
		if diag.Severity == s {
			n++
		}
	}
	return n
}

// ExitCode maps the set to a process exit code (FR-038). Errors always
// outrank warnings.
func (d Diagnostics) ExitCode(strict bool) int {
	if d.HasErrors() {
		return ExitErrors
	}
	if strict && d.CountBySeverity(SeverityWarning) > 0 {
		return ExitWarnings
	}
	return ExitSuccess
}

// SortedForOutput returns a copy ordered by file, then start byte, then code
// (FR-033). It copies rather than sorting in place because callers render the
// same set more than once — text to the terminal, JSON to a file — and both
// renderings must agree.
func (d Diagnostics) SortedForOutput() Diagnostics {
	out := make(Diagnostics, len(d))
	copy(out, d)
	sort.SliceStable(out, func(i, j int) bool {
		if c := out[i].Range.Compare(out[j].Range); c != 0 {
			return c < 0
		}
		return strings.Compare(out[i].Code, out[j].Code) < 0
	})
	return out
}
