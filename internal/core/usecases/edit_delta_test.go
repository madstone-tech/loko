package usecases

import (
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func deltaDiag(sev arch.Severity, code, addr, file, detail string, line int) arch.Diagnostic {
	return arch.Diagnostic{Severity: sev, Code: code, Address: arch.Address(addr), Detail: detail,
		Range: arch.SourceRange{File: file, StartLine: line}}
}

func TestDiagnosticDelta(t *testing.T) {
	t.Parallel()
	warn := func(addr string, line int) arch.Diagnostic {
		return deltaDiag(arch.SeverityWarning, "missing_description", addr, "main.loko.hcl", "", line)
	}
	info := func(detail string, line int) arch.Diagnostic {
		return deltaDiag(arch.SeverityInfo, "unresolved_reference", "", "main.loko.hcl", detail, line)
	}
	tests := map[string]struct {
		before, after  arch.Diagnostics
		added, removed arch.Diagnostics
		unchanged      DiagnosticCounts
	}{
		"lines shift": {
			before:    arch.Diagnostics{warn("system.a", 3), info("x", 9)},
			after:     arch.Diagnostics{warn("system.a", 5), info("x", 11)},
			unchanged: DiagnosticCounts{Warnings: 1, Info: 1},
		},
		"removed keeps its old position": {
			before:    arch.Diagnostics{warn("system.a", 3), warn("system.b", 7)},
			after:     arch.Diagnostics{warn("system.a", 3)},
			removed:   arch.Diagnostics{warn("system.b", 7)},
			unchanged: DiagnosticCounts{Warnings: 1},
		},
		"added": {
			before:    arch.Diagnostics{warn("system.a", 3)},
			after:     arch.Diagnostics{warn("system.a", 3), warn("system.c", 12)},
			added:     arch.Diagnostics{warn("system.c", 12)},
			unchanged: DiagnosticCounts{Warnings: 1},
		},
		"detail is part of the key": {
			before:  arch.Diagnostics{info("x", 2)},
			after:   arch.Diagnostics{info("y", 2)},
			added:   arch.Diagnostics{info("y", 2)},
			removed: arch.Diagnostics{info("x", 2)},
		},
		"equal keys pair in source order": {
			before:    arch.Diagnostics{info("x", 2), info("x", 8)},
			after:     arch.Diagnostics{info("x", 4)},
			removed:   arch.Diagnostics{info("x", 8)},
			unchanged: DiagnosticCounts{Info: 1},
		},
		"no change": {
			before:    arch.Diagnostics{warn("system.a", 3), deltaDiag(arch.SeverityError, "e", "", "f", "", 1)},
			after:     arch.Diagnostics{warn("system.a", 3), deltaDiag(arch.SeverityError, "e", "", "f", "", 1)},
			unchanged: DiagnosticCounts{Errors: 1, Warnings: 1},
		},
		"nothing at all": {},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			added, removed, unchanged := diagnosticDelta(tt.before, tt.after)
			if !slices.EqualFunc(added, tt.added, diagEqual) {
				t.Errorf("added = %v, want %v", added, tt.added)
			}
			if !slices.EqualFunc(removed, tt.removed, diagEqual) {
				t.Errorf("removed = %v, want %v", removed, tt.removed)
			}
			if unchanged != tt.unchanged {
				t.Errorf("unchanged = %+v, want %+v", unchanged, tt.unchanged)
			}
			if p := unchanged.orNil(); (p == nil) != (unchanged == DiagnosticCounts{}) {
				t.Errorf("orNil(%+v) = %v: want nil exactly when every count is 0", unchanged, p)
			}
		})
	}
}

func diagEqual(a, b arch.Diagnostic) bool {
	return a.Code == b.Code && a.Severity == b.Severity && a.Address == b.Address &&
		a.Detail == b.Detail && a.Range == b.Range
}
