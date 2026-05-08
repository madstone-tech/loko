package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CheckFileSize checks file against all FileSizeRules, returning any violations.
// For each rule whose pathPattern matches the file's path, the effective-line
// count is compared to the rule's MaxEffectiveLines. If the file is exempt (via
// any matching file-size exemption), no violation is emitted.
func CheckFileSize(file *ParsedFile, rules []FileSizeRule, exemptions []Exemption) []Violation {
	var violations []Violation

	for _, rule := range rules {
		if !matchPath(rule.PathPattern, file.Path) {
			continue
		}
		if isFileSizeExempt(file, exemptions) {
			continue
		}
		if file.EffectiveLines > rule.MaxEffectiveLines {
			basename := filepath.Base(file.Path)
			msg := fmt.Sprintf("%s:1: %s exceeds %s (%d > %d)",
				file.Path, basename, rule.Name, file.EffectiveLines, rule.MaxEffectiveLines)
			violations = append(violations, Violation{
				Rule:    rule.Name,
				Kind:    "file-size",
				File:    file.Path,
				Line:    1,
				Subject: basename,
				Actual:  file.EffectiveLines,
				Limit:   rule.MaxEffectiveLines,
				Message: msg,
			})
		}
	}

	return violations
}

// isFileSizeExempt returns true if the file matches any file-size exemption.
func isFileSizeExempt(file *ParsedFile, exemptions []Exemption) bool {
	basename := filepath.Base(file.Path)

	for _, ex := range exemptions {
		if ex.Kind != "file-size" {
			continue
		}
		// Generated header exemption
		if ex.Match.GeneratedHeader && file.Generated {
			return true
		}
		// Basename exemption
		for _, bn := range ex.Match.Basename {
			if basename == bn {
				return true
			}
		}
		// PathPattern exemption
		if ex.Match.PathPattern != "" && matchPath(ex.Match.PathPattern, file.Path) {
			return true
		}
	}

	// Also exempt test files via the standard _test.go suffix check
	// (handled by pathPattern exemption in the YAML, but check as fallback)
	if strings.HasSuffix(file.Path, "_test.go") {
		for _, ex := range exemptions {
			if ex.Kind == "file-size" && ex.Match.PathPattern == "**/*_test.go" {
				return true
			}
		}
	}

	return false
}
