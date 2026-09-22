package usecases

import (
	"context"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// stubSource is a concrete mock of the ArchitectureSource port. No mocking
// library: Principle VII says concrete mock structs are sufficient, and the
// layer rules forbid a use-case test from importing the real adapter.
type stubSource struct {
	model *arch.SourceModel
	diags arch.Diagnostics
	err   error
}

func (s stubSource) Load(context.Context, string) (*arch.SourceModel, arch.Diagnostics, error) {
	return s.model, s.diags, s.err
}
