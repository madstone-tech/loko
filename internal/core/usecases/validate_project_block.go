package usecases

import (
	"fmt"
	"slices"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// validateIgnoreWarnings checks the project's ignore_warnings list and returns
// the codes that may be ignored. An error code is an error: errors cannot be
// ignored. An unknown code is only a warning, so a release that retires a code
// does not break a project that names it (feature 017, FR-014).
func validateIgnoreWarnings(model *arch.SourceModel) (arch.Diagnostics, []string) {
	var (
		diags arch.Diagnostics
		valid []string
	)
	p := model.Project
	for _, c := range p.IgnoreWarnings {
		switch {
		case slices.Contains(arch.WarningCodes, c):
			valid = append(valid, c)
		case slices.Contains(arch.AllCodes, c):
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     arch.CodeInvalidAttributeValue,
				Summary:  "Errors cannot be ignored",
				Detail:   fmt.Sprintf("ignore_warnings lists %q, which is an error. Only warnings can be ignored: %s.", c, strings.Join(arch.WarningCodes, ", ")),
				Range:    p.IgnoreWarningsRange,
			})
		default:
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning,
				Code:     arch.CodeUnknownWarningCode,
				Summary:  "Unknown warning code",
				Detail:   fmt.Sprintf("ignore_warnings lists %q, which is not a code loko emits. Remove or correct it; warning codes are: %s.", c, strings.Join(arch.WarningCodes, ", ")),
				Range:    p.IgnoreWarningsRange,
			})
		}
	}
	return diags, valid
}
