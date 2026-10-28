package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runValidateOpts(t *testing.T, opts ValidateOptions) (code int, stdout, stderr string, err error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	opts.Stdout, opts.Stderr = &out, &errBuf
	if opts.Format == "" {
		opts.Format = formatText
	}
	code, err = runValidateWith(context.Background(), opts)
	return code, out.String(), errBuf.String(), err
}

const firstRun = "../testdata/projects/first-run"

// TestValidateSummary is SC-003: a summary of the reported model fits on one
// screen. The golden is testdata/validate_summary_first_run.txt.
func TestValidateSummary(t *testing.T) {
	t.Parallel()
	_, _, stderr, err := runValidateOpts(t, ValidateOptions{Root: firstRun, Summary: true})
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(stderr, "\n"); n > 30 {
		t.Errorf("summary is %d lines, want at most 30:\n%s", n, stderr)
	}
	golden := filepath.Join("testdata", "validate_summary_first_run.txt")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(stderr), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if stderr != string(want) {
		t.Errorf("summary differs from %s:\n%s", golden, stderr)
	}
	_, _, empty, _ := runValidateOpts(t, ValidateOptions{Root: project(t, cleanFiles()), Summary: true})
	if strings.Count(empty, "\n") != 1 {
		t.Errorf("summary of a clean project is not one line: %q", empty)
	}
}

func TestValidateSeverityAndIgnore(t *testing.T) {
	t.Parallel()
	code, _, stderr, _ := runValidateOpts(t, ValidateOptions{Root: firstRun, Severity: "error", Strict: true})
	if code != 2 || strings.Contains(stderr, "missing_docs") {
		t.Errorf("--severity error --strict: exit %d, stderr:\n%s", code, stderr)
	}
	all := []string{"missing_docs", "empty_system", "orphan_element", "unbound_instance"}
	if code, _, _, _ := runValidateOpts(t, ValidateOptions{Root: firstRun, Strict: true, Ignore: all}); code != 0 {
		t.Errorf("--strict with every warning ignored: exit %d, want 0", code)
	}
	if code, _, _, _ := runValidateOpts(t, ValidateOptions{Root: firstRun, Strict: true, Ignore: all[:3]}); code != 2 {
		t.Errorf("--strict with unbound_instance left: exit %d, want 2", code)
	}
	code, _, _, err := runValidateOpts(t, ValidateOptions{Root: firstRun, Ignore: []string{"syntax_error"}})
	if code != 1 || err == nil || !strings.Contains(err.Error(), "orphan_element") {
		t.Errorf("--ignore syntax_error: exit %d, %v, want 1 and the warning codes", code, err)
	}
	if _, _, _, err := runValidateOpts(t, ValidateOptions{Root: firstRun, Severity: "info"}); err == nil {
		t.Error("--severity info was accepted")
	}
}

// TestStrictEscalatesMissingDocs pins the greenfield contract: missing_docs
// is information by default, so a new model exits 0 with prose still unwritten,
// while --strict escalates it to a warning: exit 2 and shown as a warning.
func TestStrictEscalatesMissingDocs(t *testing.T) {
	t.Parallel()
	code, _, stderr, err := runValidateOpts(t, ValidateOptions{Root: firstRun})
	if err != nil || code != 0 {
		t.Fatalf("lenient exit %d, %v", code, err)
	}
	if strings.Contains(stderr, "warning  missing_docs") {
		t.Errorf("lenient run shows missing_docs as a warning:\n%s", stderr)
	}
	code, _, stderr, _ = runValidateOpts(t, ValidateOptions{Root: firstRun, Strict: true})
	if code != 2 {
		t.Errorf("strict exit %d, want 2", code)
	}
	if !strings.Contains(stderr, "⚠ missing_docs") && !strings.Contains(stderr, "Warning") {
		t.Errorf("strict run does not surface missing_docs as warned:\n%s", stderr)
	}
}

// TestBuildAndValidateStrictAgree is FR-012: the same --ignore gives the
// same strict outcome from build as from validate.
func TestBuildAndValidateStrictAgree(t *testing.T) {
	t.Parallel()
	ignore := []string{"missing_docs", "empty_system", "orphan_element", "unbound_instance"}
	vcode, _, _, _ := runValidateOpts(t, ValidateOptions{Root: firstRun, Strict: true, Ignore: ignore})
	var out, errBuf bytes.Buffer
	bcode, err := runBuildWith(context.Background(), BuildOptions{Root: firstRun, Out: t.TempDir(),
		Strict: true, Ignore: ignore, Stdout: &out, Stderr: &errBuf})
	if err != nil || bcode != vcode || vcode != 0 {
		t.Errorf("validate exit %d, build exit %d (%v)", vcode, bcode, err)
	}
}

// TestRemodelledWarnings is SC-002: with externals as external and unbuilt
// instances planned, the reported model shows at most 10 warnings, none of
// them about systems the team does not own or infrastructure not built yet.
func TestRemodelledWarnings(t *testing.T) {
	t.Parallel()
	code, out, _, err := runValidateOpts(t, ValidateOptions{Root: "../testdata/projects/first-run-remodelled", Format: formatJSON})
	if err != nil || code != 0 {
		t.Fatalf("exit %d, %v", code, err)
	}
	var doc struct {
		Diagnostics []struct {
			Severity string `json:"severity"`
			Code     string `json:"code"`
			Address  string `json:"address"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	warnings := 0
	for _, d := range doc.Diagnostics {
		if d.Severity == "warning" {
			warnings++
		}
		if d.Code == "docs_root_relative" {
			t.Errorf("the remodelled fixture writes docs paths file-relative, yet %s gets %s", d.Address, d.Code)
		}
	}
	if warnings > 10 {
		t.Errorf("%d warnings, want at most 10", warnings)
	}
	for _, d := range doc.Diagnostics {
		if d.Code == "empty_system" || d.Code == "unbound_instance" || strings.HasPrefix(d.Address, "external.") {
			t.Errorf("an expected warning remains: %s %s", d.Code, d.Address)
		}
	}
}
