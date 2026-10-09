package usecases

import (
	"context"
	"fmt"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// CompileRequest names what to compile and against which build version.
type CompileRequest struct {
	// Root is the project directory to discover source beneath.
	Root string
	// BuildVersion is the running loko version, checked against the project's
	// loko_version constraint.
	BuildVersion string
}

// CompileResult is everything one compilation produced.
type CompileResult struct {
	Model    *arch.SourceModel
	Resolved *Resolved
	Diags    arch.Diagnostics
}

// HasErrors reports whether compilation found anything that blocks artefact
// production.
func (r CompileResult) HasErrors() bool { return r.Diags.HasErrors() }

// CompileArchitecture runs the pipeline: source -> resolve -> validate.
//
// Every stage contributes diagnostics and NONE of them return early on error.
// That is FR-031: five independent mistakes must yield five diagnostics from
// one invocation, because a compiler that reports one problem per run turns a
// ten-minute fix into a ten-round conversation.
//
// Building the IR is a separate step (BuildIR), so a caller that only wants
// diagnostics does not pay for construction.
func CompileArchitecture(ctx context.Context, src ArchitectureSource, req CompileRequest) (*CompileResult, error) {
	if src == nil {
		return nil, fmt.Errorf("compile: no architecture source configured")
	}

	model, diags, err := src.Load(ctx, req.Root)
	if err != nil {
		return nil, fmt.Errorf("compile: loading %s: %w", req.Root, err)
	}
	if model == nil {
		model = &arch.SourceModel{}
	}

	// The version constraint is checked first so an incompatible project says
	// so plainly, rather than burying it under errors caused by language
	// features this build does not understand.
	diags = append(diags, ValidateProjectVersion(model, req.BuildVersion)...)

	res, resolveDiags := ResolveModel(model)
	diags = append(diags, resolveDiags...)
	diags = append(diags, ValidateStructure(model, res)...)
	diags = append(diags, ValidateDeployment(model, res)...)
	diags = append(diags, ValidateMoved(model)...)
	diags = append(diags, ValidateRenderAttributes(model)...)
	diags = append(diags, ValidateWarnings(model, res, req.Root)...)

	return &CompileResult{Model: model, Resolved: res, Diags: diags}, nil
}

// Exit codes, re-exported so outer layers never import entities directly
// (Constitution v1.2.0, Dependency Direction). There are exactly three, and no
// command may add a fourth (FR-038).
const (
	ExitSuccess  = arch.ExitSuccess
	ExitErrors   = arch.ExitErrors
	ExitWarnings = arch.ExitWarnings
)

// ExitCode maps this result to a process exit code.
func (r CompileResult) ExitCode(strict bool) int { return r.Diags.ExitCode(strict) }

// ErrorCount and WarningCount report the diagnostic tallies without exposing
// the entity severity type to callers.
func (r CompileResult) ErrorCount() int   { return r.Diags.CountBySeverity(arch.SeverityError) }
func (r CompileResult) WarningCount() int { return r.Diags.CountBySeverity(arch.SeverityWarning) }
