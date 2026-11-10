package authoring

import "fmt"

// FileContent is one source file before and after an edit plan. Old is nil for
// a file the plan creates; New is the full new content.
type FileContent struct {
	Path string
	Old  []byte
	New  []byte
}

// Plan is the in-memory result of applying edits: the new content of every
// changed file, sorted by path, and the addresses a cascade removed. Nothing
// is on disk until the plan is committed.
type Plan struct {
	Files   []FileContent
	Removed []string
}

// Changed reports whether any file's content differs (FR-021: a no-op saves nothing).
func (p Plan) Changed() bool {
	for _, f := range p.Files {
		if f.Old == nil || string(f.Old) != string(f.New) {
			return true
		}
	}
	return false
}

// FileChange is one changed file in a result, with a unified diff.
// Whitespace lists the new line numbers of the lines that the diff shows as
// context because only their whitespace changed (feature 023, contract W-4).
type FileChange struct {
	Path       string `json:"path" toon:"path"`
	Diff       string `json:"diff" toon:"diff"`
	Whitespace []int  `json:"whitespace,omitempty" toon:"whitespace,omitempty"`
}

// DiffMode selects how a diff shows lines that changed only in whitespace.
type DiffMode int

const (
	// DiffMeaning shows a line whose only change is whitespace outside a
	// string literal as a context line, and lists it in Whitespace.
	DiffMeaning DiffMode = iota
	// DiffExact shows every changed byte, as 1.5 did.
	DiffExact
)

// Refusal reasons. A refused write changes nothing on disk.
const (
	ReasonCompileErrors      = "compile_errors"
	ReasonStaleRevision      = "stale_revision"
	ReasonDanglingReferences = "dangling_references"
	ReasonAddressInUse       = "address_in_use"
	ReasonNotFound           = "not_found"
	ReasonInvalidEdit        = "invalid_edit"
	ReasonPathRefused        = "path_refused"
	// Feature 017: the project declaration.
	ReasonProjectExists = "project_exists"
	ReasonNoProject     = "no_project"
)

// EditError is a refused edit: which edit in the batch, why, and (for a
// dangling removal) which declarations still refer to the target.
type EditError struct {
	Index      int
	Reason     string
	Detail     string
	Dependents []string
}

func (e *EditError) Error() string {
	return fmt.Sprintf("edit %d: %s: %s", e.Index, e.Reason, e.Detail)
}
