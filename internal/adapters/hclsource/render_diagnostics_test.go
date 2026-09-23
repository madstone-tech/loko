package hclsource

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestRendererShowsSnippet is the difference between a diagnostic a reader can
// act on and a line number they have to go look up.
func TestRendererShowsSnippet(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"arch.loko.hcl": "project \"p\" {}\n\nsystem \"a\" {\n  colour = \"red\"\n}\n",
	})

	_, diags, err := New().Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	r, err := NewRendererForRoot(root, false)
	if err != nil {
		t.Fatalf("NewRendererForRoot: %v", err)
	}

	var buf bytes.Buffer
	errs, warns, writeErr := r.Write(&buf, diags)
	if writeErr != nil {
		t.Fatalf("Write: %v", writeErr)
	}
	out := buf.String()

	if errs != 1 || warns != 0 {
		t.Errorf("got %d errors %d warnings, want 1 and 0", errs, warns)
	}
	for _, want := range []string{
		"Error: Unsupported attribute",
		"on arch.loko.hcl:4:3",
		`colour = "red"`,
		"^",
		"1 error, 0 warnings",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRendererNoColourByDefault(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{"arch.loko.hcl": `system "a" { colour = "x" }`})
	_, diags, _ := New().Load(context.Background(), root)

	r, _ := NewRendererForRoot(root, false)
	var buf bytes.Buffer
	if _, _, err := r.Write(&buf, diags); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "\033[") {
		t.Error("colour codes present when colour is disabled")
	}
}

func TestRendererCleanProject(t *testing.T) {
	t.Parallel()

	root := writeTree(t, map[string]string{
		"arch.loko.hcl": "project \"p\" {}\nsystem \"a\" { description = \"ok\" }\n",
	})
	_, diags, _ := New().Load(context.Background(), root)

	r, _ := NewRendererForRoot(root, false)
	var buf bytes.Buffer
	errs, _, err := r.Write(&buf, diags)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if errs != 0 {
		t.Errorf("clean project reported %d errors:\n%s", errs, buf.String())
	}
	if !strings.Contains(buf.String(), "No problems found.") {
		t.Errorf("want the clean-run summary, got:\n%s", buf.String())
	}
}
