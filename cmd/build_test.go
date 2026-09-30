package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/usecases"
)

const fixtures = "../testdata/projects"

// copyFixture copies a testdata project into a temporary directory, so a test
// may modify it or build into it.
func copyFixture(t *testing.T, name string) string {
	t.Helper()
	dst := t.TempDir()
	src := filepath.Join(fixtures, name)
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			if rel == "expected" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

type buildRun struct {
	code           int
	stdout, stderr string
}

func runBuildIn(t *testing.T, root, out string, formats ...string) buildRun {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code, err := runBuildWith(context.Background(), BuildOptions{
		Root: root, Out: out, Formats: formats, Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("runBuildWith: %v\nstderr:\n%s", err, stderr.String())
	}
	return buildRun{code, stdout.String(), stderr.String()}
}

// listFiles returns every file under dir, relative and slash-separated.
func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func hashTree(t *testing.T, dir string) map[string][32]byte {
	t.Helper()
	m := map[string][32]byte{}
	for _, f := range listFiles(t, dir) {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f)))
		if err != nil {
			t.Fatal(err)
		}
		m[f] = sha256.Sum256(b)
	}
	return m
}

func withPrefix(files []string, prefix string) []string {
	var out []string
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			out = append(out, f)
		}
	}
	return out
}

// TestBuildZeroConfig is US1: diagrams for the landscape, each system with
// containers, each container with components, and each environment, with no
// configuration (AC1–AC3).
func TestBuildZeroConfig(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	r := runBuildIn(t, filepath.Join(fixtures, "two-systems"), out, "d2,svg")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	var want []string
	for _, id := range []string{"container-api", "container-gateway", "deployment-prod", "deployment-staging",
		"landscape", "system-payments", "system-shop"} {
		want = append(want, "diagrams/"+id+".d2", "diagrams/"+id+".svg")
	}
	sort.Strings(want)
	if got := withPrefix(listFiles(t, out), "diagrams/"); !reflect.DeepEqual(got, want) {
		t.Errorf("diagrams =\n%v\nwant\n%v", got, want)
	}
	if !strings.Contains(r.stdout, "built "+out+": ") {
		t.Errorf("summary missing:\n%s", r.stdout)
	}

	prod := readString(t, filepath.Join(out, "diagrams", "deployment-prod.d2"))
	if !strings.Contains(prod, `"deployment__prod__node__vpc-main__subnet-a": "subnet-a" {`) {
		t.Errorf("prod view lacks both levels of node nesting (AC2):\n%s", prod)
	}
	// staging has no ledger instance, so gateway's record relationship leaves
	// the view and must reach its boundary (AC4, FR-009).
	staging := readString(t, filepath.Join(out, "diagrams", "deployment-staging.d2"))
	if !strings.Contains(staging, `-> "outside"`) {
		t.Errorf("staging view lacks a boundary connection (AC4):\n%s", staging)
	}
}

// TestBuildCompileErrorLeavesOutputUntouched is US1/AC5 and FR-024.
func TestBuildCompileErrorLeavesOutputUntouched(t *testing.T) {
	t.Parallel()
	root := copyFixture(t, "two-systems")
	out := t.TempDir()
	if r := runBuildIn(t, root, out, "d2"); r.code != usecases.ExitSuccess {
		t.Fatalf("first build exit = %d\n%s", r.code, r.stderr)
	}
	before := hashTree(t, out)

	f := filepath.Join(root, "main.loko.hcl")
	src := readString(t, f)
	if err := os.WriteFile(f, []byte(strings.Replace(src, "container.orders_db", "container.nope", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runBuildIn(t, root, out, "d2")
	if r.code != usecases.ExitErrors {
		t.Errorf("exit = %d, want %d", r.code, usecases.ExitErrors)
	}
	if !strings.Contains(r.stderr, "main.loko.hcl:") {
		t.Errorf("diagnostics lack a source position:\n%s", r.stderr)
	}
	if after := hashTree(t, out); !reflect.DeepEqual(before, after) {
		t.Error("a failed build changed the output directory")
	}
}

func TestBuildUnknownFormat(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	_, err := runBuildWith(context.Background(), BuildOptions{
		Root: filepath.Join(fixtures, "two-systems"), Out: t.TempDir(), Formats: []string{"pdf"},
		Stdout: &stdout, Stderr: &stderr,
	})
	if err == nil || !strings.Contains(err.Error(), `unsupported format "pdf": supported formats are d2, svg, md, html`) {
		t.Fatalf("err = %v", err)
	}
}

func runBuildStrict(t *testing.T, root, out string) int {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code, err := runBuildWith(context.Background(), BuildOptions{
		Root: root, Out: out, Strict: true, Stdout: &stdout, Stderr: &stderr,
	})
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func readString(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestBuildSite is US4/AC1–AC3 end to end.
func TestBuildSite(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	r := runBuildIn(t, filepath.Join(fixtures, "two-systems"), out, "html")
	if r.code != usecases.ExitSuccess {
		t.Fatalf("exit = %d\n%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "added svg (required by html)") {
		t.Errorf("the added format is not reported:\n%s", r.stdout)
	}
	api := readString(t, filepath.Join(out, "element", "container", "api.html"))
	for _, want := range []string{"Prose for <strong>api</strong>", "<h2>Uses</h2>", "../../diagrams/container-api.svg"} {
		if !strings.Contains(api, want) {
			t.Errorf("api page lacks %q", want)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "diagrams", "container-api.svg")); err != nil {
		t.Errorf("the embedded diagram was not produced: %v", err)
	}
	fraud := readString(t, filepath.Join(out, "element", "container", "fraud.html"))
	if !strings.Contains(fraud, "not found") {
		t.Error("the page for an element with missing prose must note it (AC3)")
	}
}
