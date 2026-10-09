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
type FileChange struct {
	Path string `json:"path" toon:"path"`
	Diff string `json:"diff" toon:"diff"`
}

// Refusal reasons. A refused write changes nothing on disk.
const (
	ReasonCompileErrors      = "compile_errors"
	ReasonStaleRevision      = "stale_revision"
	ReasonDanglingReferences = "dangling_references"
	ReasonAddressInUse       = "address_in_use"
	ReasonNotFound           = "not_found"
	ReasonInvalidEdit        = "invalid_edit"
	ReasonPathRefused        = "path_refused"
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
