package usecases

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// compileSrc writes a project and compiles it through the real parser. These
// are the end-to-end cases; rule-level cases live in the table tests, which
// need no files.
func compileSrc(t *testing.T, files map[string]string) *CompileResult {
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
	res, err := CompileArchitecture(context.Background(), hclsource.New(),
		CompileRequest{Root: root, BuildVersion: "1.0.0"})
	if err != nil {
		t.Fatalf("CompileArchitecture: %v", err)
	}
	return res
}

// TestCompileReportsEveryErrorInOneRun is FR-031 and SC-008. A compiler that
// reports one problem per invocation turns a ten-minute fix into a ten-round
// conversation.
func TestCompileReportsEveryErrorInOneRun(t *testing.T) {
	t.Parallel()

	res := compileSrc(t, map[string]string{"arch.loko.hcl": `
project "p" {}

system "payments" {}

container "api" {
  system = system.payments
  colour = "red"

  uses "orders" {
    target = container.does_not_exist
  }
}

component "loose" {
  container = system.payments
}

for_each "nope" {}

system "other" { description = trimspace("  x  ") }
`})

	want := map[string]bool{
		arch.CodeUnknownAttribute:    false, // colour
		arch.CodeUnresolvedReference: false, // container.does_not_exist
		arch.CodeWrongReferenceKind:  false, // component parented to a system
		arch.CodeUnknownBlock:        false, // for_each
		arch.CodeUnknownFunction:     false, // trimspace
	}
	for _, d := range res.Diags {
		if _, tracked := want[d.Code]; tracked {
			want[d.Code] = true
		}
	}

	missing := 0
	for code, found := range want {
		if !found {
			t.Errorf("five independent mistakes, but %s was not reported", code)
			missing++
		}
	}
	if missing > 0 {
		for _, d := range res.Diags {
			t.Logf("  got %s: %s", d.Code, d.Summary)
		}
	}
}

// TestCompileCleanProject: the happy path produces no errors, and the warnings
// it does produce leave the exit code at success.
func TestCompileCleanProject(t *testing.T) {
	t.Parallel()

	res := compileSrc(t, map[string]string{
		"arch.loko.hcl": `
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
`,
		"docs/p.md":   "prose",
		"docs/api.md": "prose",
		"docs/db.md":  "prose",
	})

	for _, d := range res.Diags {
		if d.Severity == arch.SeverityError {
			t.Errorf("unexpected error %s at %s: %s", d.Code, d.Range, d.Summary)
		}
	}
	if got := res.Diags.ExitCode(false); got != arch.ExitSuccess {
		t.Errorf("lenient exit code = %d, want %d", got, arch.ExitSuccess)
	}
}

// TestCompileWarningsEscalateUnderStrict covers FR-034 and FR-038 together.
func TestCompileWarningsEscalateUnderStrict(t *testing.T) {
	t.Parallel()

	res := compileSrc(t, map[string]string{"arch.loko.hcl": `
project "p" {}
system "payments" {}
container "api" { system = system.payments }
container "db"  { system = system.payments }
`})

	if res.Diags.HasErrors() {
		t.Fatalf("unexpected errors: %v", codes(res.Diags))
	}
	if res.Diags.CountBySeverity(arch.SeverityWarning) == 0 {
		t.Fatal("expected warnings for elements with no edges and no prose")
	}
	if got := res.Diags.ExitCode(false); got != arch.ExitSuccess {
		t.Errorf("lenient exit = %d, want %d", got, arch.ExitSuccess)
	}
	if got := res.Diags.ExitCode(true); got != arch.ExitWarnings {
		t.Errorf("strict exit = %d, want %d", got, arch.ExitWarnings)
	}
}

func TestCompileWarningCodes(t *testing.T) {
	t.Parallel()

	res := compileSrc(t, map[string]string{"arch.loko.hcl": `
project "p" {}

system "empty" {}

container "lonely" {
  system = system.other
  docs   = "./missing.md"
}

system "other" {}

container "selfish" {
  system = system.other
  uses "me" { target = container.selfish }
}

deployment "prod" {
  instance "x" { of = container.selfish }
}
`})

	for _, want := range []string{
		arch.CodeEmptySystem,
		arch.CodeOrphanElement,
		arch.CodeDocsNotFound,
		arch.CodeMissingDocs,
		arch.CodeSelfRelationship,
		arch.CodeUnboundInstance,
	} {
		if !hasCode(res.Diags, want) {
			t.Errorf("warning %s not reported; got %v", want, codes(res.Diags))
		}
	}
}

// TestCompileDeterministicDiagnosticOrder covers FR-033.
func TestCompileDeterministicDiagnosticOrder(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"b.loko.hcl": "container \"x\" { colour = \"a\" }\n",
		"a.loko.hcl": "project \"p\" {}\nsystem \"s\" { colour = \"b\" }\n",
	}

	first := codes(compileSrc(t, files).Diags.SortedForOutput())
	for i := 0; i < 5; i++ {
		got := codes(compileSrc(t, files).Diags.SortedForOutput())
		if len(got) != len(first) {
			t.Fatalf("run %d produced %d diagnostics, first run produced %d", i, len(got), len(first))
		}
		for j := range got {
			if got[j] != first[j] {
				t.Fatalf("run %d differs at %d: %s vs %s", i, j, got[j], first[j])
			}
		}
	}
}
