package main

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
)

// CheckFunctionSize walks top-level FuncDecl nodes in file.AST and reports any
// functions whose effective-line count exceeds the matching FunctionSizeRule's
// MaxEffectiveLines. Test files matching a function-size exemption are skipped
// entirely.
func CheckFunctionSize(file *ParsedFile, rules []FunctionSizeRule, exemptions []Exemption) []Violation {
	if isFuncSizeExempt(file, exemptions) {
		return nil
	}

	var violations []Violation

	for _, rule := range rules {
		if !matchPath(rule.PathPattern, file.Path) {
			continue
		}

		for _, decl := range file.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Body == nil {
				continue
			}

			// Extract the source bytes for this function declaration.
			startOff := file.FileSet.Position(fn.Pos()).Offset
			endOff := file.FileSet.Position(fn.End()).Offset
			if startOff < 0 || endOff > len(file.Source) || startOff >= endOff {
				continue
			}
			funcSrc := string(file.Source[startOff:endOff])
			effective := CountEffectiveLines(funcSrc)

			if effective > rule.MaxEffectiveLines {
				line := file.FileSet.Position(fn.Pos()).Line
				msg := fmt.Sprintf("%s:%d: %s exceeds %s (%d > %d)",
					file.Path, line, fn.Name.Name, rule.Name, effective, rule.MaxEffectiveLines)
				violations = append(violations, Violation{
					Rule:    rule.Name,
					Kind:    "function-size",
					File:    file.Path,
					Line:    line,
					Subject: fn.Name.Name,
					Actual:  effective,
					Limit:   rule.MaxEffectiveLines,
					Message: msg,
				})
			}
		}
	}

	return violations
}

// isFuncSizeExempt returns true if the file matches any function-size exemption.
func isFuncSizeExempt(file *ParsedFile, exemptions []Exemption) bool {
	basename := filepath.Base(file.Path)

	for _, ex := range exemptions {
		if ex.Kind != "function-size" {
			continue
		}
		for _, bn := range ex.Match.Basename {
			if basename == bn {
				return true
			}
		}
		if ex.Match.PathPattern != "" && matchPath(ex.Match.PathPattern, file.Path) {
			return true
		}
	}

	// Fallback: always exempt _test.go for function-size
	if strings.HasSuffix(file.Path, "_test.go") {
		for _, ex := range exemptions {
			if ex.Kind == "function-size" && ex.Match.PathPattern == "**/*_test.go" {
				return true
			}
		}
	}

	return false
}
