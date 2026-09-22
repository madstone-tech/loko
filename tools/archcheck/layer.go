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
//   - External imports (not under modulePath) are checked against
//     ForbiddenExternalImports; a match is a violation, otherwise they are
//     allowed. AllowedImports never constrains them.
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

		// External imports are allowed unless a ForbiddenExternalImports
		// pattern matches the full import path.
		if !strings.HasPrefix(impPath, prefix) && impPath != modulePath {
			for _, forbidPat := range matched.ForbiddenExternalImports {
				if matchPath(forbidPat, impPath) {
					violations = append(violations,
						newLayerViolation(file, matched, imp.Line, impPath))
					break
				}
			}
			continue
		}

		// Compute module-relative import dir (slash-separated).
		relDir := strings.TrimPrefix(impPath, prefix)
		// Normalise using path.Clean so we match exactly.
		relDir = path.Clean(relDir)

		// Check forbidden imports (takes priority).
		for _, forbidPat := range matched.ForbiddenImports {
			if matchPath(forbidPat, relDir) {
				violations = append(violations,
					newLayerViolation(file, matched, imp.Line, impPath))
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
			violations = append(violations,
				newLayerViolation(file, matched, imp.Line, impPath))
		}
		// If AllowedImports is empty AND no forbidden match, the file is
		// restricted to stdlib plus any third-party package not named in
		// ForbiddenExternalImports, both handled in the external branch above.

	nextImport:
	}

	return violations
}

// newLayerViolation builds the Violation for an import that a layer rule
// rejects. All three rejection paths (forbidden internal, forbidden external,
// and no matching allowed pattern) share it so the message format cannot drift.
func newLayerViolation(file *ParsedFile, rule *LayerRule, line int, impPath string) Violation {
	desc := strings.TrimSpace(rule.Description)
	return Violation{
		Rule:    rule.Name,
		Kind:    "layer",
		File:    file.Path,
		Line:    line,
		Subject: impPath,
		Actual:  0,
		Limit:   0,
		Message: fmt.Sprintf("%s:%d: layer '%s' may not import '%s' (rule: %s)",
			file.Path, line, rule.Name, impPath, desc),
	}
}
