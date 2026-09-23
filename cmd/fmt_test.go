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

const messySrc = `system    "billing"   {
# keep me
      description="Billing"
}
`

const canonicalSrc = "system \"billing\" {\n  # keep me\n  description = \"Billing\"\n}\n"

func runFmtIn(t *testing.T, root string, check bool) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code, err := runFmtWith(context.Background(), FmtOptions{
		Root: root, Check: check, Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("runFmtWith: %v", err)
	}
	return code, stdout.String(), stderr.String()
}

// TestFmtCheckFailsOnUnformatted covers FR-035a: this is the shape a CI job
// needs — paths listed, nothing written, non-zero exit.
func TestFmtCheckFailsOnUnformatted(t *testing.T) {
	t.Parallel()

	root := project(t, map[string]string{
		"a.loko.hcl": messySrc,
		"b.loko.hcl": messySrc,
		"c.loko.hcl": canonicalSrc,
	})

	code, stdout, stderr := runFmtIn(t, root, true)

	if code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", code, usecases.ExitErrors)
	}
	for _, want := range []string{"a.loko.hcl", "b.loko.hcl"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not list %s:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "c.loko.hcl") {
		t.Errorf("a canonical file was listed:\n%s", stdout)
	}
	if !strings.Contains(stderr, "loko fmt") {
		t.Errorf("stderr does not say how to fix it:\n%s", stderr)
	}

	// Nothing on disk changed.
	for _, name := range []string{"a.loko.hcl", "b.loko.hcl"} {
		if got := mustRead(t, root, name); got != messySrc {
			t.Errorf("--check modified %s", name)
		}
	}
}

func TestFmtCheckPassesOnCanonicalProject(t *testing.T) {
	t.Parallel()

	root := project(t, map[string]string{"a.loko.hcl": canonicalSrc})
	code, stdout, _ := runFmtIn(t, root, true)

	if code != usecases.ExitSuccess {
		t.Errorf("exit = %d, want %d", code, usecases.ExitSuccess)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("a canonical project listed paths:\n%s", stdout)
	}
}

// TestFmtCheckReusesErrorExitCode guards FR-038: exactly three exit codes, and
// --check must not introduce a fourth.
func TestFmtCheckReusesErrorExitCode(t *testing.T) {
	t.Parallel()

	root := project(t, map[string]string{"a.loko.hcl": messySrc})
	code, _, _ := runFmtIn(t, root, true)

	if code != usecases.ExitErrors {
		t.Errorf("--check exit = %d; it must reuse %d rather than add a fourth code",
			code, usecases.ExitErrors)
	}
}

func TestFmtWritesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	root := project(t, map[string]string{"a.loko.hcl": messySrc})

	code, stdout, _ := runFmtIn(t, root, false)
	if code != usecases.ExitSuccess {
		t.Errorf("exit = %d, want %d", code, usecases.ExitSuccess)
	}
	if !strings.Contains(stdout, "a.loko.hcl") {
		t.Errorf("the rewritten path was not reported:\n%s", stdout)
	}

	after := mustRead(t, root, "a.loko.hcl")
	if !strings.Contains(after, "# keep me") {
		t.Errorf("the comment was lost:\n%s", after)
	}

	// Second run is a no-op, and --check now passes.
	code, stdout, _ = runFmtIn(t, root, false)
	if code != usecases.ExitSuccess || strings.TrimSpace(stdout) != "" {
		t.Errorf("second run reported changes: exit=%d\n%s", code, stdout)
	}
	if got := mustRead(t, root, "a.loko.hcl"); got != after {
		t.Error("the second run rewrote the file")
	}
	if code, _, _ := runFmtIn(t, root, true); code != usecases.ExitSuccess {
		t.Errorf("--check fails after fmt: exit = %d", code)
	}
}

func TestFmtLeavesUnparseableFileAlone(t *testing.T) {
	t.Parallel()

	const broken = "system \"oops\" {\n"
	root := project(t, map[string]string{"broken.loko.hcl": broken})

	code, _, stderr := runFmtIn(t, root, false)

	if code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", code, usecases.ExitErrors)
	}
	if got := mustRead(t, root, "broken.loko.hcl"); got != broken {
		t.Errorf("an unparseable file was rewritten:\n%q", got)
	}
	if !strings.Contains(stderr, "broken.loko.hcl") {
		t.Errorf("the diagnostic does not name the file:\n%s", stderr)
	}
}

func mustRead(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return string(b)
}
