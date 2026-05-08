package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestEndToEnd builds the archcheck binary and runs it against the fixture
// tree. It asserts exit code 1, exactly 3 violations, and that each violation
// message matches the expected contract format.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}

	// Build the binary into a temp dir.
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "archcheck")

	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = filepath.Join(".") // tools/archcheck is the cwd from test
	// Resolve the module root (two levels up from tools/archcheck).
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	build.Dir = repoRoot
	build.Args = append(build.Args[:0], "go", "build", "-o", binPath, "./tools/archcheck")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Run archcheck against the fixture.
	fixtureDir := filepath.Join("testdata", "fixture")
	reportFile := filepath.Join(tmpDir, "report.json")

	cmd := exec.Command(binPath,
		"--rules=rules.yaml",
		"--format=text",
		"--report="+reportFile,
	)
	cmd.Dir, err = filepath.Abs(fixtureDir)
	if err != nil {
		t.Fatalf("resolving fixture dir: %v", err)
	}

	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
	}

	t.Logf("archcheck output:\n%s", string(out))

	// Must exit 1 (violations found).
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	// Parse the JSON report.
	reportData, err := os.ReadFile(reportFile)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}
	var report Report
	if err := json.Unmarshal(reportData, &report); err != nil {
		t.Fatalf("parsing report JSON: %v", err)
	}

	// Expect exactly 3 violations.
	if len(report.Violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(report.Violations))
		for i, v := range report.Violations {
			t.Logf("[%d] kind=%s file=%s subject=%s", i, v.Kind, v.File, v.Subject)
		}
	}

	// Verify violation kinds: one file-size, one layer, one function-size.
	kinds := map[string]int{}
	for _, v := range report.Violations {
		kinds[v.Kind]++
	}
	for _, k := range []string{"file-size", "layer", "function-size"} {
		if kinds[k] != 1 {
			t.Errorf("expected 1 %s violation, got %d", k, kinds[k])
		}
	}

	// Verify message formats match the contract.
	output := string(out)
	if !strings.Contains(output, "archcheck:") {
		t.Error("expected summary line in output")
	}
	if !strings.Contains(output, "3 violation(s)") {
		t.Errorf("expected '3 violation(s)' in summary, got: %s", output)
	}

	// Verify report exit_code field.
	if report.ExitCode != 1 {
		t.Errorf("report.exit_code = %d, want 1", report.ExitCode)
	}
}
