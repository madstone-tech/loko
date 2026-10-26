package usecases

import (
	"encoding/json"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// BootstrapHint is what describe offers in a directory with no source: the
// smallest valid project declaration, and the apply_edit request that creates
// it, ready to send once the placeholders are filled (feature 017, FR-020).
type BootstrapHint struct {
	Syntax  string         `json:"syntax" toon:"syntax"`
	Request map[string]any `json:"request" toon:"request"`
}

func bootstrapHint(revision string) *BootstrapHint {
	var edit map[string]any
	if err := json.Unmarshal([]byte(authoring.BootstrapEdit), &edit); err != nil {
		panic("usecases: authoring.BootstrapEdit is not JSON: " + err.Error())
	}
	return &BootstrapHint{Syntax: authoring.ProjectSyntax,
		Request: map[string]any{"base_revision": revision, "edits": []any{edit}}}
}

func containsCode(d arch.Diagnostics, code string) bool {
	return slices.ContainsFunc(d, func(x arch.Diagnostic) bool { return x.Code == code })
}
