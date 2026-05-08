package main

import (
	"fmt"
	"path"
	"strings"
)

// CheckLayerImports checks file's imports against the applicable LayerRule.
// The first rule whose pathPattern matches the file's path is used (first match
// wins). Files matched by no rule are unconstrained and return nil.
//
// For each import:
//   - External imports (not under modulePath) are always allowed.
//   - ForbiddenImports patterns are checked first; a match is a violation.
//   - AllowedImports patterns are checked next; if none match, it is a violation.
func CheckLayerImports(file *ParsedFile, rules []LayerRule, modulePath string) []Violation {
	// Find first matching rule.
	var matched *LayerRule
	for i := range rules {
		if matchPath(rules[i].PathPattern, file.Path) {
			matched = &rules[i]
			break
		}
	}
	if matched == nil {
		return nil
	}

	var violations []Violation
	prefix := modulePath + "/"

	for _, imp := range file.Imports {
		impPath := imp.Path

		// External imports are always allowed.
		if !strings.HasPrefix(impPath, prefix) && impPath != modulePath {
			continue
		}

		// Compute module-relative import dir (slash-separated).
		relDir := strings.TrimPrefix(impPath, prefix)
		// Normalise using path.Clean so we match exactly.
		relDir = path.Clean(relDir)

		// Check forbidden imports (takes priority).
		for _, forbidPat := range matched.ForbiddenImports {
			if matchPath(forbidPat, relDir) {
				desc := strings.TrimSpace(matched.Description)
				msg := fmt.Sprintf("%s:%d: layer '%s' may not import '%s' (rule: %s)",
					file.Path, imp.Line, matched.Name, impPath, desc)
				violations = append(violations, Violation{
					Rule:    matched.Name,
					Kind:    "layer",
					File:    file.Path,
					Line:    imp.Line,
					Subject: impPath,
					Actual:  0,
					Limit:   0,
					Message: msg,
				})
				goto nextImport
			}
		}

		// Check allowed imports. When the list is empty ALL internal imports are
		// forbidden (entities layer). When non-empty, the import must match at
		// least one pattern.
		for _, allowPat := range matched.AllowedImports {
			if matchPath(allowPat, relDir) {
				goto nextImport // allowed
			}
		}
		{
			// No allowed pattern matched (or list is empty) — violation.
			desc := strings.TrimSpace(matched.Description)
			msg := fmt.Sprintf("%s:%d: layer '%s' may not import '%s' (rule: %s)",
				file.Path, imp.Line, matched.Name, impPath, desc)
			violations = append(violations, Violation{
				Rule:    matched.Name,
				Kind:    "layer",
				File:    file.Path,
				Line:    imp.Line,
				Subject: impPath,
				Actual:  0,
				Limit:   0,
				Message: msg,
			})
		}
		// If AllowedImports is empty AND no forbidden match, file is restricted to
		// stdlib only (which is already handled by the external-import skip above).

	nextImport:
	}

	return violations
}
