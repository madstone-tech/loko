package authoring

// MaxBatch is the most edits one request may carry (FR-012a).
const MaxBatch = 100

// Batch is an ordered list of edits applied together: compiled once, saved
// all-or-nothing (FR-012a), or only previewed (FR-012b).
type Batch struct {
	Edits        []Edit
	Preview      bool
	BaseRevision Revision
}

// NewBatch validates the batch's shape. Each edit must already have passed NewEdit.
func NewBatch(edits []Edit, preview bool, base Revision) (Batch, error) {
	switch {
	case len(edits) == 0:
		return Batch{}, fieldError("edits", "at least one edit is required")
	case len(edits) > MaxBatch:
		return Batch{}, fieldError("edits", "%d edits exceed the limit of %d", len(edits), MaxBatch)
	case base.IsZero():
		return Batch{}, fieldError("base_revision", "required; take it from any read result")
	}
	return Batch{Edits: edits, Preview: preview, BaseRevision: base}, nil
}
