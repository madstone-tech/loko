package hclsource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestParseSyntaxErrorIsDistinct covers FR-028a: a file that does not parse is
// reported as syntax_error, separate from unknown_block and unknown_attribute,
// which apply to files that DO parse but declare something undefined.
func TestParseSyntaxErrorIsDistinct(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"broken.loko.hcl": "system \"oops\" {\n", // unclosed block
	})
	files, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	p := newParser(root)
	parsed, diags := p.parseAll(files)

	if len(parsed) != 0 {
		t.Errorf("got %d parsed files, want 0 — an unparseable file must not reach decoding", len(parsed))
	}
	if !diags.HasErrors() {
		t.Fatal("no diagnostics for an unclosed block")
	}
	for _, d := range diags {
		if d.Code != arch.CodeSyntaxError {
			t.Errorf("code = %q, want %q", d.Code, arch.CodeSyntaxError)
		}
		if d.Range.File != "broken.loko.hcl" {
			t.Errorf("range file = %q, want project-relative %q", d.Range.File, "broken.loko.hcl")
		}
		if d.Range.StartLine < 1 {
			t.Errorf("diagnostic has no line number: %+v", d)
		}
	}
}

// TestParseContinuesPastBadFile is the FR-031 guarantee at the parse stage: one
// invocation reports everything it can, so a single typo does not hide the rest
// of the project's problems.
func TestParseContinuesPastBadFile(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"a_broken.loko.hcl": "system \"oops\" {\n",
		"b_good.loko.hcl":   "system \"payments\" {\n  description = \"fine\"\n}\n",
		"c_broken.loko.hcl": "container \"x\" {{\n",
	})
	files, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	p := newParser(root)
	parsed, diags := p.parseAll(files)

	if len(parsed) != 1 || parsed[0].rel != "b_good.loko.hcl" {
		t.Fatalf("got %d parsed files, want only b_good.loko.hcl", len(parsed))
	}

	seen := map[string]bool{}
	for _, d := range diags {
		seen[d.Range.File] = true
	}
	for _, want := range []string{"a_broken.loko.hcl", "c_broken.loko.hcl"} {
		if !seen[want] {
			t.Errorf("no diagnostic for %s; a bad file must not mask its siblings", want)
		}
	}
}

// TestParseRangesAreProjectRelative guards FR-036c. An absolute path in a range
// would leak a machine-specific string into exports and break SC-003.
func TestParseRangesAreProjectRelative(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"nested/dir/x.loko.hcl": "system \"a\" {\n  bad =\n}\n",
	})
	files, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	_, diags := newParser(root).parseAll(files)
	if len(diags) == 0 {
		t.Fatal("expected a syntax diagnostic")
	}
	for _, d := range diags {
		if filepath.IsAbs(d.Range.File) {
			t.Errorf("range file %q is absolute", d.Range.File)
		}
		if got, want := d.Range.File, "nested/dir/x.loko.hcl"; got != want {
			t.Errorf("range file = %q, want %q (forward slashes on every platform)", got, want)
		}
	}
}

func TestParseUnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; permission bits do not apply")
	}
	t.Parallel()

	root := writeTree(t, map[string]string{
		"locked.loko.hcl": "system \"a\" {}\n",
		"open.loko.hcl":   "system \"b\" {}\n",
	})
	if err := os.Chmod(filepath.Join(root, "locked.loko.hcl"), 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "locked.loko.hcl"), 0o644) })

	files, _, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	parsed, diags := newParser(root).parseAll(files)

	if len(parsed) != 1 || parsed[0].rel != "open.loko.hcl" {
		t.Errorf("got %d parsed files, want only open.loko.hcl", len(parsed))
	}
	if !diags.HasErrors() {
		t.Error("unreadable file produced no diagnostic")
	}
}
