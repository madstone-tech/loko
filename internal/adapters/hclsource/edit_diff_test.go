package hclsource

import (
	"slices"
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
	d := NewEditor().Diff(p, authoring.DiffExact)
	if len(d) != 2 || d[0].Path != "a/c.loko.hcl" || d[1].Path != "b.loko.hcl" {
		t.Fatalf("Diff lists changed files sorted by path: %+v", d)
	}
	if again := NewEditor().Diff(p, authoring.DiffExact); again[1].Diff != d[1].Diff {
		t.Error("diff is not deterministic")
	}
}

// TestEditorDiffMeaning is contract W-4: lines that change only in
// whitespace outside string literals become context lines, listed by their
// new line numbers.
func TestEditorDiffMeaning(t *testing.T) {
	t.Parallel()
	old := "container \"api\" {\n  system = system.s\n  tags = [\"a\"]\n}\n"
	upd := "container \"api\" {\n  system      = system.s\n  tags        = [\"a\"]\n  description = \"Orders\"\n}\n"
	tests := map[string]struct {
		old, new   string
		diff       string
		whitespace []int
	}{
		"realigned siblings": {
			old: old, new: upd,
			diff: "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n" +
				"@@ -1,4 +1,5 @@\n container \"api\" {\n   system      = system.s\n   tags        = [\"a\"]\n+  description = \"Orders\"\n }\n",
			whitespace: []int{2, 3},
		},
		"meaning and whitespace": {
			old:  "a {\n  x = 1\n}\n",
			new:  "a {\n  x   = 2\n}\n",
			diff: "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n@@ -1,3 +1,3 @@\n a {\n-  x = 1\n+  x   = 2\n }\n",
		},
		"whitespace-only hunk dropped": {
			old: "a {\n  x = 1\n}\n" + strings.Repeat("z\n", 10) + "b {\n  y = 1\n}\n",
			new: "a {\n  x  = 1\n}\n" + strings.Repeat("z\n", 10) + "b {\n  y = 2\n}\n",
			diff: "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n" +
				"@@ -12,5 +12,5 @@\n z\n z\n b {\n-  y = 1\n+  y = 2\n }\n",
			whitespace: []int{2},
		},
		"only whitespace": {
			old: "a {\n  x = 1\n}\n", new: "a {\n  x  = 1\n}\n",
			whitespace: []int{2},
		},
		"whitespace inside a string is meaning": {
			old:  "a {\n  d = \"a b\"\n}\n",
			new:  "a {\n  d = \"a  b\"\n}\n",
			diff: "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n@@ -1,3 +1,3 @@\n a {\n-  d = \"a b\"\n+  d = \"a  b\"\n }\n",
		},
		"escaped quote stays inside the string": {
			old:  "a {\n  d = \"q\\\" b\"\n}\n",
			new:  "a {\n  d = \"q\\\"  b\"\n}\n",
			diff: "--- a/x.loko.hcl\n+++ b/x.loko.hcl\n@@ -1,3 +1,3 @@\n a {\n-  d = \"q\\\" b\"\n+  d = \"q\\\"  b\"\n }\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := authoring.Plan{Files: []authoring.FileContent{{Path: "x.loko.hcl", Old: []byte(tt.old), New: []byte(tt.new)}}}
			got := NewEditor().Diff(p, authoring.DiffMeaning)
			if len(got) != 1 {
				t.Fatalf("want one file, got %+v", got)
			}
			if got[0].Diff != tt.diff {
				t.Errorf("diff:\n%s\nwant:\n%s", got[0].Diff, tt.diff)
			}
			if !slices.Equal(got[0].Whitespace, tt.whitespace) {
				t.Errorf("whitespace = %v, want %v", got[0].Whitespace, tt.whitespace)
			}
			exact := NewEditor().Diff(p, authoring.DiffExact)
			if len(exact) != 1 || exact[0].Diff != unifiedDiff("x.loko.hcl", []byte(tt.old), []byte(tt.new)) || exact[0].Whitespace != nil {
				t.Errorf("exact diff is not the 1.5 diff: %+v", exact)
			}
		})
	}
}
