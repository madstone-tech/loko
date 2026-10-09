package hclsource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeLayout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := "project \"p\" {\n  layout = \"elk\"\n}\n\nsystem \"s\" {}\n\nview \"v\" {\n  include = [system.s]\n  layout  = \"dagre\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.loko.hcl"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	m, diags, err := New().Load(t.Context(), dir)
	if err != nil || diags.HasErrors() {
		t.Fatal(err, diags)
	}
	if m.Project.Layout != "elk" || m.Project.LayoutRange.StartLine != 2 {
		t.Errorf("project: %+v", m.Project)
	}
	if v := m.Views[0]; v.Layout != "dagre" || v.LayoutRange.StartLine != 9 {
		t.Errorf("view: %+v", v)
	}
}
