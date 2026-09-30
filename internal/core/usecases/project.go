package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// ProseDoc is the result of reading one docs reference.
type ProseDoc struct {
	Text  string
	Found bool
}

// ProseSet holds the prose read before projection, keyed by the authored docs
// value. It is an input to the projection, never serialised.
type ProseSet map[string]ProseDoc

// Project turns the compiled architecture into the complete intermediate
// value every backend consumes. It is a pure function: the same IR, prose and
// provenance always yield a deep-equal projection (FR-010).
//
// A view model that fails its own invariants is a projection bug, reported as
// an error rather than a diagnostic.
func Project(ir *arch.IR, prose ProseSet, prov Provenance) (*viewmodel.Projection, arch.Diagnostics, error) {
	views, diags := ResolveViews(ir, prov)
	p := &viewmodel.Projection{
		Project: viewmodel.ProjectHeader{Name: ir.Project.Name, Description: ir.Project.Description},
		Sources: prov.AllFiles(),
	}
	for _, v := range views {
		vm := ProjectView(ir, v, prov)
		if err := vm.Validate(); err != nil {
			return nil, diags, fmt.Errorf("projection: %w", err)
		}
		p.Views = append(p.Views, vm)
		if v.Kind == viewmodel.KindDeploymentView {
			p.Environments = append(p.Environments, viewmodel.LinkRef{
				Address: v.Subject, Name: v.Title, Kind: "deployment", PagePath: vm.PagePath,
			})
		}
	}
	p.Pages = ProjectPages(ir, prose, prov, p.Views)
	return p, diags, nil
}

// CompileAndProject compiles the project and returns its projection, or the
// diagnostics that prevented one. It reads prose through reader when given.
// Outer layers use it to inspect the intermediate value without touching
// entity types directly.
func CompileAndProject(ctx context.Context, src ArchitectureSource, reader ProseReader,
	req CompileRequest) (*viewmodel.Projection, arch.Diagnostics, error) {

	compiled, err := CompileArchitecture(ctx, src, req)
	if err != nil {
		return nil, nil, err
	}
	if compiled.HasErrors() {
		return nil, compiled.Diags, nil
	}
	ir := BuildIR(compiled.Model, compiled.Resolved)
	prose, err := readProse(ctx, reader, req.Root, ir)
	if err != nil {
		return nil, nil, err
	}
	proj, diags, err := Project(ir, prose, BuildProvenance(compiled.Model))
	return proj, append(compiled.Diags, diags...), err
}
