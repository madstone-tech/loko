package hclsource

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

var update = flag.Bool("update", false, "rewrite golden files from current behaviour")

// Golden fixtures pair a source tree with the diagnostics it must produce.
//
// Each case is a directory under testdata/golden/<name>/ holding the *.loko.hcl
// inputs plus expected_diagnostics.json. Adding a language feature or a
// validation rule means adding a case, which is what makes SC-004's coverage
// claim checkable rather than aspirational.
//
// Only parse-stage diagnostics appear here. Resolution and semantic rules are
// core's job and are table-tested there against SourceModel literals, so a
// failure points at one layer rather than two.
func TestGolden(t *testing.T) {
	t.Parallel()

	root := filepath.Join("testdata", "golden")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading %s: %v", root, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGolden(t, filepath.Join(root, name))
		})
	}
}

func runGolden(t *testing.T, dir string) {
	t.Helper()

	inputDir := filepath.Join(dir, "input")
	_, diags, err := New().Load(context.Background(), inputDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	got := normaliseDiags(diags)
	goldenPath := filepath.Join(dir, "expected_diagnostics.json")

	if *update {
		writeGolden(t, goldenPath, got)
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden (run with -update to create): %v", err)
	}
	var want []goldenDiag
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("parsing golden: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d\ngot:  %s\nwant: %s",
			len(got), len(want), mustJSON(t, got), mustJSON(t, want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("diagnostic %d:\n got %+v\nwant %+v", i, got[i], want[i])
		}
	}
}

// goldenDiag is the comparable projection of a diagnostic. Detail is excluded
// on purpose: the code and the source range are the contract, while the prose
// is free to improve without invalidating every fixture.
type goldenDiag struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Summary  string `json:"summary"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func normaliseDiags(diags arch.Diagnostics) []goldenDiag {
	out := make([]goldenDiag, 0, len(diags))
	for _, d := range diags.SortedForOutput() {
		out = append(out, goldenDiag{
			Severity: d.Severity.String(),
			Code:     d.Code,
			Summary:  d.Summary,
			File:     d.Range.File,
			Line:     d.Range.StartLine,
			Column:   d.Range.StartColumn,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Code < out[j].Code
	})
	return out
}

func writeGolden(t *testing.T, path string, got []goldenDiag) {
	t.Helper()
	if err := os.WriteFile(path, mustJSON(t, got), 0o644); err != nil {
		t.Fatalf("writing golden: %v", err)
	}
	t.Logf("updated %s", path)
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return append(b, '\n')
}

// TestGoldenCasesCoverParseRules asserts every parse-stage diagnostic code has
// at least one fixture. Without this the fixture set silently stops keeping
// pace with the language.
func TestGoldenCasesCoverParseRules(t *testing.T) {
	t.Parallel()

	// The codes the parser itself can emit; the rest belong to core.
	parseCodes := []string{
		arch.CodeSyntaxError,
		arch.CodeNoSourceFound,
		arch.CodeUnknownBlock,
		arch.CodeUnknownAttribute,
		arch.CodeUnknownFunction,
		arch.CodeWrongReferenceKind,
		arch.CodeDuplicateDeclaration,
	}

	seen := map[string]bool{}
	root := filepath.Join("testdata", "golden")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading %s: %v", root, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		b, readErr := os.ReadFile(filepath.Join(root, entry.Name(), "expected_diagnostics.json"))
		if readErr != nil {
			continue
		}
		var diags []goldenDiag
		if json.Unmarshal(b, &diags) != nil {
			continue
		}
		for _, d := range diags {
			seen[d.Code] = true
		}
	}

	var missing []string
	for _, code := range parseCodes {
		if !seen[code] {
			missing = append(missing, code)
		}
	}
	if len(missing) > 0 {
		t.Errorf("no golden fixture produces these parse diagnostics: %s",
			strings.Join(missing, ", "))
	}
}
