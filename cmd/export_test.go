package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

func runExportIn(t *testing.T, root, format, out string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code, err := runExportWith(context.Background(), ExportOptions{
		Root: root, Format: format, Out: out, Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("runExportWith: %v", err)
	}
	return code, stdout.String(), stderr.String()
}

// TestExportNoArtifactOnError covers FR-037, including the part that matters
// most: an existing --out file is left untouched rather than truncated, so a
// broken project cannot destroy the last good export.
func TestExportNoArtifactOnError(t *testing.T) {
	t.Parallel()

	files := cleanFiles()
	files["arch.loko.hcl"] = strings.Replace(cleanProject, "container.orders_db", "container.nope", 1)
	root := project(t, files)

	outPath := filepath.Join(t.TempDir(), "ir.json")
	const sentinel = "PREVIOUS GOOD EXPORT"
	if err := os.WriteFile(outPath, []byte(sentinel), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runExportIn(t, root, "json", outPath)

	if code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", code, usecases.ExitErrors)
	}
	if stdout != "" {
		t.Errorf("an artefact was written to stdout despite errors:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Unresolvable reference") {
		t.Errorf("diagnostics missing from stderr:\n%s", stderr)
	}

	after, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("the --out file was removed: %v", err)
	}
	if string(after) != sentinel {
		t.Errorf("the --out file was overwritten with %q; a failed export must leave it alone", after)
	}
}

// TestExportArtifactToStdoutDiagnosticsToStderr is what lets
// `loko export --format json | conftest test -` work unfiltered.
func TestExportArtifactToStdoutDiagnosticsToStderr(t *testing.T) {
	t.Parallel()

	files := cleanFiles()
	delete(files, "docs/api.md") // a warning, not an error

	code, stdout, stderr := runExportIn(t, project(t, files), "json", "")
	if code != usecases.ExitSuccess {
		t.Errorf("exit = %d, want %d", code, usecases.ExitSuccess)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("stdout is not the artefact:\n%s", stdout)
	}
	if strings.Contains(stdout, "Warning") {
		t.Error("a diagnostic leaked into stdout, which would corrupt a pipeline")
	}
	if !strings.Contains(stderr, "Warning") {
		t.Errorf("the warning did not reach stderr:\n%s", stderr)
	}
}

// TestExportIsStateless covers FR-027: no lock file, cache, or state file. A
// future cache added without thought would be caught here.
func TestExportIsStateless(t *testing.T) {
	t.Parallel()

	root := project(t, cleanFiles())
	before := snapshot(t, root)

	if code, _, _ := runExportIn(t, root, "json", ""); code != usecases.ExitSuccess {
		t.Fatalf("export failed")
	}
	if code, _, _ := runValidateIn(t, root, formatText, false); code != usecases.ExitSuccess {
		t.Fatalf("validate failed")
	}

	after := snapshot(t, root)
	for path := range after {
		if _, existed := before[path]; !existed {
			t.Errorf("%s was created inside the project root; the compiler is stateless", path)
		}
	}
	for path, size := range before {
		if after[path] != size {
			t.Errorf("%s changed size; the compiler must not write to its input", path)
		}
	}
}

func snapshot(t *testing.T, root string) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		out[rel] = info.Size()
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return out
}

func TestExportBothFormats(t *testing.T) {
	t.Parallel()

	root := project(t, cleanFiles())
	for _, format := range []string{"json", "toon"} {
		code, stdout, _ := runExportIn(t, root, format, "")
		if code != usecases.ExitSuccess {
			t.Errorf("%s: exit = %d", format, code)
		}
		if !strings.Contains(stdout, "schemaVersion") {
			t.Errorf("%s output carries no schemaVersion:\n%s", format, stdout)
		}
	}
}

func TestExportWritesToOutFile(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "nested", "dir", "ir.json")
	code, stdout, _ := runExportIn(t, project(t, cleanFiles()), "json", outPath)

	if code != usecases.ExitSuccess {
		t.Fatalf("exit = %d", code)
	}
	if stdout != "" {
		t.Errorf("--out was given but stdout also received the artefact:\n%s", stdout)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading --out: %v", err)
	}
	if !strings.Contains(string(data), `"schemaVersion": 1`) {
		t.Errorf("the written file is not an export:\n%s", data)
	}
}
