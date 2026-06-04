// Package cligolden holds CLI golden-file regression tests for feature
// 010-constitution-compliance (T004 capture + T019 replay).
//
// Each covered subcommand is run as a black-box subprocess (the real built loko
// binary) in a fresh temp dir; its stdout, stderr, exit code, and resulting file
// tree are captured, sanitised (absolute temp paths → <TMP>), and asserted
// against a checked-in golden under tests/golden/cli/.
//
// A failure means a refactor changed observable CLI behaviour — fix the refactor,
// not the golden. To re-baseline after a deliberate, reviewed change:
//
//	go test ./tests/integration/cli/ -update
//
// COVERAGE NOTE (honest record): only `init` and `validate` are covered here.
// They are fully deterministic and portable. The `new system|container|component`
// commands render templates whose resolution depends on the binary's install
// layout (`<exeDir>/../templates` or `./templates`), which is environment-
// sensitive and unsuitable for a portable CI golden. Their logic is covered by
// the scaffold use-case unit tests (internal/core/usecases/scaffold_*_test.go,
// with a mock TemplateEngine) and by tests/integration/scaffolding_test.go.
package cligolden

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden files instead of asserting against them")

const goldenRelDir = "tests/golden/cli"

// lokoBin is the path to the built loko binary, set in TestMain.
var lokoBin string

func TestMain(m *testing.M) {
	flag.Parse()

	root, err := findRepoRoot()
	if err != nil {
		panic("cli golden: locate repo root: " + err.Error())
	}
	bin := filepath.Join(os.TempDir(), "loko-cli-golden-test")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		panic("cli golden: go build failed: " + err.Error() + "\n" + string(out))
	}
	lokoBin = bin

	code := m.Run()
	_ = os.Remove(bin)
	os.Exit(code)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// cliCase describes one subcommand invocation. setup runs first (its output is
// discarded); then args is run and captured. workdir is the subprocess cwd;
// treeDir is the directory whose file tree is recorded (relative to workdir).
type cliCase struct {
	name    string
	setup   [][]string // commands run before the captured command (output ignored)
	args    []string   // the captured command
	treeDir string     // dir (relative to workdir) whose tree is recorded; "" = none
}

func cases() []cliCase {
	return []cliCase{
		{
			name:    "init",
			args:    []string{"init", "myproj", "-d", "demo project"},
			treeDir: "myproj",
		},
		{
			name:    "validate",
			setup:   [][]string{{"init", "myproj", "-d", "demo project"}},
			args:    []string{"validate", "-p", "PROJECT"}, // PROJECT placeholder → abs path at runtime
			treeDir: "myproj",
		},
	}
}

func run(t *testing.T, workdir string, args []string) (stdout, stderr string, exit int) {
	t.Helper()
	// Replace the PROJECT placeholder with the absolute project path.
	resolved := make([]string, len(args))
	for i, a := range args {
		if a == "PROJECT" {
			resolved[i] = filepath.Join(workdir, "myproj")
		} else {
			resolved[i] = a
		}
	}
	cmd := exec.Command(lokoBin, resolved...)
	cmd.Dir = workdir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exit = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			t.Fatalf("run %v: %v", resolved, err)
		}
	}
	return outBuf.String(), errBuf.String(), exit
}

// fileTree returns a sorted, newline-joined list of file paths under dir,
// relative to dir. Returns "(none)" if dir does not exist.
func fileTree(t *testing.T, dir string) string {
	t.Helper()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "(none)"
	}
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "(empty)"
	}
	return strings.Join(files, "\n")
}

// sanitise replaces volatile, machine-specific strings with stable placeholders.
func sanitise(s, workdir string) string {
	s = strings.ReplaceAll(s, workdir, "<TMP>")
	s = strings.ReplaceAll(s, os.TempDir(), "<TMP>")
	return s
}

func render(stdout, stderr string, exit int, tree string) string {
	var b strings.Builder
	b.WriteString("=== stdout ===\n")
	b.WriteString(stdout)
	if !strings.HasSuffix(stdout, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("=== stderr ===\n")
	b.WriteString(stderr)
	if !strings.HasSuffix(stderr, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("=== exit ===\n")
	b.WriteString(strconv.Itoa(exit))
	b.WriteString("\n=== files ===\n")
	b.WriteString(tree)
	b.WriteString("\n")
	return b.String()
}

func TestCLIGolden(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	goldenDir := filepath.Join(root, goldenRelDir)

	for _, tc := range cases() {
		t.Run(tc.name, func(t *testing.T) {
			workdir := t.TempDir()
			for _, s := range tc.setup {
				run(t, workdir, s) // setup output intentionally discarded
			}
			stdout, stderr, exit := run(t, workdir, tc.args)

			tree := "(none)"
			if tc.treeDir != "" {
				tree = fileTree(t, filepath.Join(workdir, tc.treeDir))
			}

			got := render(sanitise(stdout, workdir), sanitise(stderr, workdir), exit, tree)
			path := filepath.Join(goldenDir, tc.name+".golden")

			if *update {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update to create it)", path, err)
			}
			if string(want) != got {
				t.Errorf("CLI output for %q diverged from golden %s.\n--- want ---\n%s\n--- got ---\n%s",
					strings.Join(tc.args, " "), path, want, got)
			}
		})
	}
}
