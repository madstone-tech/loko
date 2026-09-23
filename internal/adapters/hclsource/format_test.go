package hclsource

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

const messy = `system    "billing"   {
# keep this comment exactly here
      description="Billing"
        owner = "finance"   # trailing comment
}

   container "api" {
system=system.billing
}
`

func TestFormatCanonicalises(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{"a.loko.hcl": messy})
	changed, diags, err := NewFormatter().Format(context.Background(), root)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(changed) != 1 || changed[0] != "a.loko.hcl" {
		t.Fatalf("changed = %v, want [a.loko.hcl]", changed)
	}

	out := readFile(t, root, "a.loko.hcl")

	// Comments survive, in place. This is the property the whole authoring
	// stage depends on: an agent editing a hand-written file must not destroy
	// the human's annotations.
	for _, want := range []string{
		"# keep this comment exactly here",
		"# trailing comment",
	} {
		if !containsStr(out, want) {
			t.Errorf("comment lost: %q\n%s", want, out)
		}
	}
	// Declaration order is preserved.
	if indexOfStr(out, `system "billing"`) > indexOfStr(out, `container "api"`) {
		t.Errorf("declarations were reordered:\n%s", out)
	}
	// Alignment is canonical.
	if containsStr(out, `description="Billing"`) {
		t.Errorf("assignment not canonicalised:\n%s", out)
	}
}

// TestFormatIsIdempotent covers SC-007. A formatter that keeps changing its
// own output cannot be used as a CI gate.
func TestFormatIsIdempotent(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{"a.loko.hcl": messy})
	f := NewFormatter()

	if _, _, err := f.Format(context.Background(), root); err != nil {
		t.Fatalf("first Format: %v", err)
	}
	afterFirst := readFile(t, root, "a.loko.hcl")

	changed, _, err := f.Format(context.Background(), root)
	if err != nil {
		t.Fatalf("second Format: %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("second run reported %v as changed; formatting is not idempotent", changed)
	}
	if readFile(t, root, "a.loko.hcl") != afterFirst {
		t.Error("the second run rewrote the file")
	}
}

func TestFormatLeavesCanonicalFilesAlone(t *testing.T) {
	t.Parallel()

	canonical := "system \"billing\" {\n  description = \"Billing\"\n}\n"
	root := writeTree(t, map[string]string{"a.loko.hcl": canonical})

	changed, _, err := NewFormatter().Format(context.Background(), root)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("a canonical file was reported as changed: %v", changed)
	}
	if got := readFile(t, root, "a.loko.hcl"); got != canonical {
		t.Errorf("a canonical file was rewritten:\n%q", got)
	}
}

// TestFormatRefusesUnparseableFile covers FR-035. hclwrite.Format is purely
// lexical and would happily rewrite a broken file, so the parse check is what
// makes running fmt over a whole project safe.
func TestFormatRefusesUnparseableFile(t *testing.T) {
	t.Parallel()

	broken := "system \"oops\" {\n"
	root := writeTree(t, map[string]string{
		"broken.loko.hcl": broken,
		"fine.loko.hcl":   messy,
	})

	changed, diags, err := NewFormatter().Format(context.Background(), root)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}

	if got := readFile(t, root, "broken.loko.hcl"); got != broken {
		t.Errorf("an unparseable file was rewritten:\n%q", got)
	}
	if !diags.HasErrors() {
		t.Error("no diagnostic for the unparseable file")
	}
	for _, d := range diags {
		if d.Code != arch.CodeSyntaxError {
			t.Errorf("code = %q, want %q", d.Code, arch.CodeSyntaxError)
		}
		if d.Range.StartLine < 1 {
			t.Error("the diagnostic has no source position")
		}
	}
	// A broken file must not stop its siblings being formatted.
	if len(changed) != 1 || changed[0] != "fine.loko.hcl" {
		t.Errorf("changed = %v, want [fine.loko.hcl]", changed)
	}
}

// TestCheckWritesNothing covers FR-035a.
func TestCheckWritesNothing(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"a.loko.hcl": messy,
		"b.loko.hcl": messy,
		"c.loko.hcl": "system \"ok\" {\n  description = \"fine\"\n}\n",
	})

	changed, _, err := NewFormatter().Check(context.Background(), root)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(changed) != 2 || changed[0] != "a.loko.hcl" || changed[1] != "b.loko.hcl" {
		t.Errorf("changed = %v, want [a.loko.hcl b.loko.hcl] in sorted order", changed)
	}
	for _, name := range []string{"a.loko.hcl", "b.loko.hcl"} {
		if readFile(t, root, name) != messy {
			t.Errorf("%s was modified by Check", name)
		}
	}
}

// TestFormatPreservesFileMode: a formatter should be the least invasive tool
// in the chain.
func TestFormatPreservesFileMode(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{"a.loko.hcl": messy})
	path := filepath.Join(root, "a.loko.hcl")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := NewFormatter().Format(context.Background(), root); err != nil {
		t.Fatalf("Format: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("file mode = %v, want 0600", got)
	}
}

func readFile(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return string(b)
}

func containsStr(s, sub string) bool { return indexOfStr(s, sub) >= 0 }

func indexOfStr(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
