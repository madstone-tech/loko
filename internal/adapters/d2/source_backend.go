package d2

import (
	"context"

	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// SourceBackend is the "d2" backend: one D2 source file per view.
type SourceBackend struct{}

// NewSourceBackend returns the "d2" backend.
func NewSourceBackend() *SourceBackend { return &SourceBackend{} }

// Format implements usecases.Backend.
func (*SourceBackend) Format() vm.Format { return vm.FormatD2 }

// Render implements usecases.Backend.
func (*SourceBackend) Render(_ context.Context, in *vm.Projection, _ usecases.RenderOptions) ([]vm.Artifact, error) {
	out := make([]vm.Artifact, 0, len(in.Views))
	for _, v := range in.Views {
		out = append(out, vm.Artifact{
			Path:   vm.DiagramFile(v.View.ID, "d2"),
			Format: vm.FormatD2,
			Bytes:  Emit(v),
			Owner:  ownerOf(v),
		})
	}
	return out, nil
}

// ownerOf names what a view depicts, for output path collision reports.
func ownerOf(v vm.ViewModel) string {
	if v.View.Subject != "" {
		return v.View.Subject
	}
	return "view " + string(v.View.ID)
}
