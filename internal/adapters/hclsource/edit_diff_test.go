package hclsource

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func TestUnifiedDiff(t *testing.T) {
	t.Parallel()
	letters := strings.Split("abcdefghijklmnopqrst", "")
	old := strings.Join(letters, "\n") + "\n"
	letters[4] = "E"
	new := strings.Join(append(letters, "u"), "\n") + "\n"
	want := "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n" +
		"@@ -2,7 +2,7 @@\n b\n c\n d\n-e\n+E\n f\n g\n h\n" +
		"@@ -18,3 +18,4 @@\n r\n s\n t\n+u\n"
	near := unifiedDiff("x.loko.hcl", []byte("a\nb\nc\nd\n"), []byte("A\nb\nc\nD\n"))
	if !strings.Contains(near, "@@ -1,4 +1,4 @@\n-a\n+A\n b\n c\n-d\n+D\n") || strings.Count(near, "@@") != 2 {
		t.Errorf("nearby changes share one hunk:\n%s", near)
	}
	if got := unifiedDiff("x.loko.hcl", []byte(old), []byte(new)); got != want {
		t.Errorf("diff:\n%s\nwant:\n%s", got, want)
	}
	if got := unifiedDiff("x.loko.hcl", []byte(old), []byte(old)); got != "" {
		t.Errorf("identical inputs gave %q", got)
	}
	created := unifiedDiff("n.loko.hcl", nil, []byte("x\ny\n"))
	if !strings.HasPrefix(created, "--- /dev/null\n+++ b/n.loko.hcl\n@@ -0,0 +1,2 @@\n+x\n+y\n") {
		t.Errorf("new file diff:\n%s", created)
	}
	noEOL := unifiedDiff("x.loko.hcl", []byte("a\n"), []byte("a\nb"))
	if !strings.Contains(noEOL, "+b\n\\ No newline at end of file\n") {
		t.Errorf("missing final newline is not marked:\n%s", noEOL)
	}
}

func TestEditorDiffSortedAndDeterministic(t *testing.T) {
	t.Parallel()
	p := authoring.Plan{Files: []authoring.FileContent{
		{Path: "b.loko.hcl", Old: []byte("x\n"), New: []byte("y\n")},
		{Path: "a.loko.hcl", Old: []byte("x\n"), New: []byte("x\n")},
		{Path: "a/c.loko.hcl", Old: []byte("1\n"), New: []byte("2\n")},
	}}
	d := NewEditor().Diff(p)
	if len(d) != 2 || d[0].Path != "a/c.loko.hcl" || d[1].Path != "b.loko.hcl" {
		t.Fatalf("Diff lists changed files sorted by path: %+v", d)
	}
	if again := NewEditor().Diff(p); again[1].Diff != d[1].Diff {
		t.Error("diff is not deterministic")
	}
}
