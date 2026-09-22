package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

const cleanProject = `
project "acme" { loko_version = "~> 1.0" }

system "payments" { docs = "./docs/p.md" }

container "api" {
  system = system.payments
  docs   = "./docs/api.md"
  uses "db" { target = container.orders_db }
}

container "orders_db" {
  system = system.payments
  docs   = "./docs/db.md"
}
`

func project(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func cleanFiles() map[string]string {
	return map[string]string{
		"arch.loko.hcl": cleanProject,
		"docs/p.md":     "prose",
		"docs/api.md":   "prose",
		"docs/db.md":    "prose",
	}
}

func runValidateIn(t *testing.T, root, format string, strict bool) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code, err := runValidateWith(context.Background(), ValidateOptions{
		Root: root, Strict: strict, Format: format, Stdout: &out, Stderr: &errBuf,
	})
	if err != nil {
		t.Fatalf("runValidateWith: %v", err)
	}
	return code, out.String(), errBuf.String()
}

// Quickstart Scenario 1.
func TestValidateCleanProjectExitsZero(t *testing.T) {
	t.Parallel()

	code, _, stderr := runValidateIn(t, project(t, cleanFiles()), formatText, false)
	if code != usecases.ExitSuccess {
		t.Errorf("exit = %d, want %d\n%s", code, usecases.ExitSuccess, stderr)
	}
	if !strings.Contains(stderr, "No problems found.") {
		t.Errorf("want the clean summary, got:\n%s", stderr)
	}
}

func TestValidateBrokenReferenceExitsOne(t *testing.T) {
	t.Parallel()

	files := cleanFiles()
	files["arch.loko.hcl"] = strings.Replace(cleanProject, "container.orders_db", "container.ordrs_db", 1)

	code, _, stderr := runValidateIn(t, project(t, files), formatText, false)
	if code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", code, usecases.ExitErrors)
	}
	// The diagnostic must name the file, the line, AND the column of the
	// reference — not of the enclosing block.
	if !strings.Contains(stderr, "Unresolvable reference") {
		t.Errorf("missing the error:\n%s", stderr)
	}
	if !strings.Contains(stderr, "arch.loko.hcl:9") {
		t.Errorf("diagnostic does not point at the reference's line:\n%s", stderr)
	}
	if !strings.Contains(stderr, "container.orders_db?") {
		t.Errorf("no near-miss suggestion:\n%s", stderr)
	}
}

func TestValidateStrictEscalatesWarnings(t *testing.T) {
	t.Parallel()

	files := cleanFiles()
	delete(files, "docs/api.md") // produces docs_not_found

	root := project(t, files)
	if code, _, _ := runValidateIn(t, root, formatText, false); code != usecases.ExitSuccess {
		t.Errorf("lenient exit = %d, want %d", code, usecases.ExitSuccess)
	}
	if code, _, _ := runValidateIn(t, root, formatText, true); code != usecases.ExitWarnings {
		t.Errorf("strict exit = %d, want %d", code, usecases.ExitWarnings)
	}
}

// FR-034a: exit codes are identical whichever output form is selected.
func TestValidateExitCodesMatchAcrossFormats(t *testing.T) {
	t.Parallel()

	cases := map[string]map[string]string{
		"clean": cleanFiles(),
		"error": func() map[string]string {
			f := cleanFiles()
			f["arch.loko.hcl"] = strings.Replace(cleanProject, "container.orders_db", "container.nope", 1)
			return f
		}(),
	}

	for name, files := range cases {
		for _, strict := range []bool{false, true} {
			root := project(t, files)
			textCode, _, _ := runValidateIn(t, root, formatText, strict)
			jsonCode, _, _ := runValidateIn(t, root, formatJSON, strict)
			if textCode != jsonCode {
				t.Errorf("%s strict=%v: text exit %d but json exit %d", name, strict, textCode, jsonCode)
			}
		}
	}
}

func TestValidateJSONGoesToStdout(t *testing.T) {
	t.Parallel()

	code, stdout, _ := runValidateIn(t, project(t, cleanFiles()), formatJSON, false)
	if code != usecases.ExitSuccess {
		t.Errorf("exit = %d, want %d", code, usecases.ExitSuccess)
	}

	var doc struct {
		SchemaVersion int `json:"schemaVersion"`
		Diagnostics   []struct {
			Code     string `json:"code"`
			Severity string `json:"severity"`
		} `json:"diagnostics"`
		Summary struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if doc.SchemaVersion != 1 {
		t.Errorf("schemaVersion = %d, want 1", doc.SchemaVersion)
	}
	// An empty run must emit [] rather than null, so consumers need no
	// special case for the clean project.
	if !strings.Contains(stdout, `"diagnostics": []`) {
		t.Errorf("clean run should emit an empty array, got:\n%s", stdout)
	}
}
