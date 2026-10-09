package hclsource

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func TestRevision(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	ed := NewEditor()
	a, err := ed.Revision(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files()) != 3 || a.Files()[0].Path != "deploy.loko.hcl" {
		t.Fatalf("revision files: %+v", a.Files())
	}
	b, _ := ed.Revision(t.Context(), root)
	if a.Token() != b.Token() {
		t.Error("revision is not deterministic")
	}
	p := filepath.Join(root, "views.loko.hcl")
	src, _ := os.ReadFile(p)
	if err := os.WriteFile(p, append(src, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	c, _ := ed.Revision(t.Context(), root)
	ha, _ := a.Hash("views.loko.hcl")
	hc, _ := c.Hash("views.loko.hcl")
	hm, _ := c.Hash("main.loko.hcl")
	hma, _ := a.Hash("main.loko.hcl")
	if ha == hc || hm != hma {
		t.Error("only the changed file's hash may change")
	}
}

func planFor(t *testing.T, root string, files ...authoring.FileContent) (authoring.Plan, authoring.Revision) {
	t.Helper()
	rev, err := NewEditor().Revision(t.Context(), root)
	if err != nil {
		t.Fatal(err)
	}
	for i, f := range files {
		if f.Old == nil {
			if b, err := os.ReadFile(filepath.Join(root, f.Path)); err == nil {
				files[i].Old = b
			}
		}
	}
	return authoring.Plan{Files: files}, rev
}

func TestCommitWritesAtomically(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	if err := os.Chmod(filepath.Join(root, "views.loko.hcl"), 0o600); err != nil {
		t.Fatal(err)
	}
	p, rev := planFor(t, root,
		authoring.FileContent{Path: "views.loko.hcl", New: []byte("# v\n")},
		authoring.FileContent{Path: "new/dir/n.loko.hcl", New: []byte("# n\n")})
	if err := NewEditor().Commit(t.Context(), root, p, rev); err != nil {
		t.Fatal(err)
	}
	snap := snapshot(t, root)
	if snap["views.loko.hcl"] != "# v\n" || snap["new/dir/n.loko.hcl"] != "# n\n" {
		t.Errorf("commit wrote %v", snap)
	}
	if fi, _ := os.Stat(filepath.Join(root, "views.loko.hcl")); fi.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want preserved 0600", fi.Mode().Perm())
	}
	for p := range snap {
		if strings.Contains(p, ".tmp") {
			t.Errorf("temporary file left behind: %s", p)
		}
	}
}

func TestCommitRefusesStale(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p, rev := planFor(t, root, authoring.FileContent{Path: "views.loko.hcl", New: []byte("# v\n")})
	if err := os.WriteFile(filepath.Join(root, "views.loko.hcl"), []byte("# someone else\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, root)
	err := NewEditor().Commit(t.Context(), root, p, rev)
	var ee *authoring.EditError
	if !errors.As(err, &ee) || ee.Reason != authoring.ReasonStaleRevision || !strings.Contains(ee.Detail, "views.loko.hcl") {
		t.Fatalf("err = %v, want stale_revision naming the file", err)
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Error("a stale commit changed files")
	}
	// A file created by someone else since the read is stale too.
	p2, rev2 := planFor(t, root, authoring.FileContent{Path: "late.loko.hcl", New: []byte("# mine\n")})
	_ = os.WriteFile(filepath.Join(root, "late.loko.hcl"), []byte("# theirs\n"), 0o644)
	if err := NewEditor().Commit(t.Context(), root, p2, rev2); !errors.As(err, &ee) || ee.Reason != authoring.ReasonStaleRevision {
		t.Errorf("a file that appeared since the read: %v", err)
	}
}

func TestCommitRefusesPaths(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	for _, path := range []string{"../out.loko.hcl", "notes.md", "/abs.loko.hcl", "a/../../x.loko.hcl", "dist/x.d2"} {
		p, rev := planFor(t, root, authoring.FileContent{Path: path, New: []byte("x")})
		var ee *authoring.EditError
		if err := NewEditor().Commit(t.Context(), root, p, rev); !errors.As(err, &ee) || ee.Reason != authoring.ReasonPathRefused {
			t.Errorf("%s: err = %v, want path_refused", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "out.loko.hcl")); err == nil {
		t.Error("a refused path was written")
	}
}

func TestCommitRollsBack(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	root := handwrittenCopy(t)
	locked := filepath.Join(root, "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "z.loko.hcl"), []byte("# z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, rev := planFor(t, root,
		authoring.FileContent{Path: "deploy.loko.hcl", New: []byte("# changed\n")},
		authoring.FileContent{Path: "locked/z.loko.hcl", New: []byte("# changed\n")})
	before := snapshot(t, root)
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	if err := NewEditor().Commit(t.Context(), root, p, rev); err == nil {
		t.Fatal("commit into a read-only directory succeeded")
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Error("a failed commit was not rolled back: the first file kept its new content")
	}
}
