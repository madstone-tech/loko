// Command archcheck audits the loko repository's Go source tree against the
// Clean Architecture constitution defined in structural-rules.yaml. It checks
// layer-import rules, file-size budgets, and function-size budgets, and emits
// a human-readable text report (stderr) and an optional JSON report file.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	os.Exit(run())
}

// run is the real entry point; it returns an exit code so main can call
// os.Exit after all deferred cleanups have run.
func run() int {
	rulesFlag := flag.String("rules",
		"specs/009-constitution-compliance/contracts/structural-rules.yaml",
		"Path to structural-rules YAML file")
	formatFlag := flag.String("format", "text", "Output format: text or json")
	reportFlag := flag.String("report", "", "Path to write JSON report file (empty = no file)")
	annotateFlag := flag.String("annotate", "none", "Annotation mode: none or github")
	emitGolangciFlag := flag.Bool("emit-golangci-config", false,
		"Print depguard YAML config derived from layer rules to stdout and exit")
	flag.Parse()

	// Load rules first so we can fail fast.
	rules, err := LoadRules(*rulesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "archcheck: error loading rules: %v\n", err)
		return 2
	}

	if *emitGolangciFlag {
		printDepguardConfig(rules)
		return 0
	}

	// Determine module path by reading go.mod in the current directory.
	modulePath, err := readModulePath("go.mod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "archcheck: cannot determine module path: %v\n", err)
		return 2
	}

	// Walk the repository and parse Go files.
	parsedFiles, err := walkAndParse(".", modulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "archcheck: walk error: %v\n", err)
		return 2
	}

	// Run all checkers.
	var allViolations []Violation
	totalFunctions := 0

	for _, pf := range parsedFiles {
		allViolations = append(allViolations, CheckFileSize(pf, rules.FileSizes, rules.Exemptions)...)
		allViolations = append(allViolations, CheckFunctionSize(pf, rules.FunctionSizes, rules.Exemptions)...)
		totalFunctions += countTopLevelFuncs(pf)
		allViolations = append(allViolations, CheckLayerImports(pf, rules.Layers, modulePath)...)
	}

	// Sort violations deterministically.
	allViolations = sortedViolations(allViolations)

	exitCode := 0
	if len(allViolations) > 0 {
		exitCode = 1
	}

	report := &Report{
		Version:               "1.0",
		GeneratedAt:           time.Now().UTC().Format(time.RFC3339),
		AuditToolVersion:      toolVersion(),
		RulesPath:             *rulesFlag,
		TotalFilesScanned:     len(parsedFiles),
		TotalFunctionsScanned: totalFunctions,
		Violations:            allViolations,
		ExitCode:              exitCode,
	}

	// Write JSON report file if requested.
	if *reportFlag != "" {
		if err := writeReportFile(*reportFlag, report); err != nil {
			fmt.Fprintf(os.Stderr, "archcheck: error writing report: %v\n", err)
			return 2
		}
	}

	// Output to stdout (json) or stderr (text).
	switch *formatFlag {
	case "json":
		if err := WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "archcheck: error writing JSON: %v\n", err)
			return 2
		}
	default: // "text"
		annotateGitHub := *annotateFlag == "github"
		WriteText(os.Stderr, report, annotateGitHub)
	}

	return exitCode
}

// walkAndParse walks root finding all .go files (excluding vendor, hidden
// dirs, dist, node_modules, and tools/archcheck/testdata) and parses them.
func walkAndParse(root, modulePath string) ([]*ParsedFile, error) {
	_ = modulePath // reserved for future use
	var files []*ParsedFile

	err := filepath.WalkDir(root, func(osPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()

		// Skip directories we never want to descend into.
		if d.IsDir() {
			if name == "." {
				return nil
			}
			switch name {
			case "vendor", "dist", "node_modules":
				return filepath.SkipDir
			}
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files.
		if !strings.HasSuffix(name, ".go") {
			return nil
		}

		// Compute module-relative slash path.
		relOS, err := filepath.Rel(root, osPath)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}
		relSlash := filepath.ToSlash(relOS)

		// Skip tools/archcheck/testdata (the binary auditing itself is fine,
		// but testdata fixtures have their own rules).
		if strings.HasPrefix(relSlash, "tools/archcheck/testdata/") {
			return nil
		}

		pf, err := parseGoFile(osPath, relSlash)
		if err != nil {
			// Non-fatal: report and continue.
			fmt.Fprintf(os.Stderr, "archcheck: warning: skipping %s: %v\n", relSlash, err)
			return nil
		}
		files = append(files, pf)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort for determinism.
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

var reGeneratedHeader = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.`)

// parseGoFile reads and parses a single Go file.
func parseGoFile(osPath, relSlash string) (*ParsedFile, error) {
	src, err := os.ReadFile(osPath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, osPath, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing AST: %w", err)
	}

	imports := extractImports(astFile, fset)
	generated := reGeneratedHeader.Match(src)
	effective := CountEffectiveLines(string(src))

	return &ParsedFile{
		Path:           relSlash,
		Source:         src,
		FileSet:        fset,
		AST:            astFile,
		EffectiveLines: effective,
		Imports:        imports,
		Generated:      generated,
	}, nil
}

// extractImports pulls import paths and their line numbers from the AST.
func extractImports(f *ast.File, fset *token.FileSet) []ImportSpec {
	var specs []ImportSpec
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		line := fset.Position(imp.Path.Pos()).Line
		specs = append(specs, ImportSpec{Path: path, Line: line})
	}
	return specs
}

// readModulePath reads the first "module" directive from go.mod.
func readModulePath(gomodPath string) (string, error) {
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", fmt.Errorf("opening %s: %w", gomodPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("no module directive found in %s", gomodPath)
}

// toolVersion returns the build VCS revision or "unknown".
func toolVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				v := s.Value
				if len(v) > 12 {
					v = v[:12]
				}
				return v
			}
		}
	}
	return "unknown"
}

// countTopLevelFuncs counts *ast.FuncDecl nodes in the file.
func countTopLevelFuncs(pf *ParsedFile) int {
	n := 0
	for _, d := range pf.AST.Decls {
		if _, ok := d.(*ast.FuncDecl); ok {
			n++
		}
	}
	return n
}

// writeReportFile writes the JSON report to path.
func writeReportFile(path string, report *Report) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encoding report: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// printDepguardConfig emits a golangci depguard YAML fragment derived from
// the layer rules (implements T039).
func printDepguardConfig(rs *RuleSet) {
	fmt.Println("# Generated by archcheck --emit-golangci-config")
	fmt.Println("# Paste into .golangci.yml under linters-settings.depguard.rules")
	fmt.Println("linters-settings:")
	fmt.Println("  depguard:")
	fmt.Println("    rules:")
	for _, lr := range rs.Layers {
		slug := strings.ReplaceAll(lr.Name, "/", "-")
		fmt.Printf("      %s:\n", slug)
		fmt.Printf("        files:\n")
		fmt.Printf("          - \"**/%s\"\n", lr.PathPattern)
		if len(lr.ForbiddenImports) > 0 {
			fmt.Printf("        deny:\n")
			for _, fp := range lr.ForbiddenImports {
				fmt.Printf("          - pkg: \"%s/**\"\n", fp)
				fmt.Printf("            desc: \"%s\"\n", lr.Description)
			}
		}
	}
}
