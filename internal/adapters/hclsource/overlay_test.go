package hclsource

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func handwrittenCopy(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	if err := os.CopyFS(dst, os.DirFS("testdata/handwritten")); err != nil {
		t.Fatal(err)
	}
	return dst
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestLoadOverlay(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	before := snapshot(t, root)
	src := New()
	ctx := context.Background()

	base, _, err := src.Load(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	same, _, _ := src.LoadOverlay(ctx, root, nil)
	if !reflect.DeepEqual(base, same) {
		t.Error("an empty overlay must equal Load")
	}

	views := strings.Replace(before["views.loko.hcl"], `"payments-path"`, `"renamed"`, 1)
	m, diags, err := src.LoadOverlay(ctx, root, []authoring.FileContent{
		{Path: "views.loko.hcl", New: []byte(views)},
		{Path: "extra/new.loko.hcl", New: []byte("system \"added\" {}\n")},
	})
	if err != nil || diags.HasErrors() {
		t.Fatalf("LoadOverlay: %v %v", err, diags)
	}
	if len(m.Views) != 1 || m.Views[0].Name != "renamed" {
		t.Errorf("overlay content not compiled: views %+v", m.Views)
	}
	if !strings.Contains(strings.Join(m.Files, ","), "extra/new.loko.hcl") {
		t.Errorf("overlay-only file missing from model files %v", m.Files)
	}
	found := false
	for _, e := range m.Elements {
		found = found || e.Name == "added"
	}
	if !found {
		t.Error("overlay-only file was not decoded")
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Error("LoadOverlay changed files on disk")
	}
}
