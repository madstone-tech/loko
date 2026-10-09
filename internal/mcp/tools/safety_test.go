package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

// handwritten copies the comment-heavy, deliberately non-canonical fixture.
func handwritten(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	if err := os.CopyFS(dst, os.DirFS(filepath.Join("..", "..", "adapters", "hclsource", "testdata", "handwritten"))); err != nil {
		t.Fatal(err)
	}
	return dst
}

// tree hashes every file under root, keyed by relative path.
func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		sum := sha256.Sum256(b)
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestRefusedWritesChangeNothing is SC-004: every kind of refused write
// leaves every file byte-identical.
func TestRefusedWritesChangeNothing(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		prepare func(t *testing.T, root string)
		edit    map[string]any
		stale   bool
		reason  string
	}{
		"does not compile": {edit: edit("update", "element", "container.web", "set", map[string]any{"system": "system.nowhere"}),
			reason: "compile_errors"},
		"stale revision": {edit: edit("update", "element", "system.shop", "set", map[string]any{"owner": "x"}),
			stale: true, reason: "stale_revision"},
		"dangling removal": {edit: edit("remove", "element", "container.api"), reason: "dangling_references"},
		"refused path":     {edit: edit("add", "element", "system.x", "file", "../outside.loko.hcl"), reason: "path_refused"},
		"another file has a syntax error": {
			prepare: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "broken.loko.hcl"), []byte("system \"oops\" {\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			edit: edit("update", "element", "system.shop", "set", map[string]any{"owner": "x"}), reason: "compile_errors"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := handwritten(t)
			svc, enc := newService(root), encoding.NewEncoder()
			rev := readRevision(t, NewValidateTool(svc, enc))
			if tc.prepare != nil {
				tc.prepare(t, root)
			}
			if tc.stale {
				p := filepath.Join(root, "main.loko.hcl")
				b, _ := os.ReadFile(p)
				_ = os.WriteFile(p, append(b, []byte("\n# edited by hand meanwhile\n")...), 0o644)
			}
			before := tree(t, root)
			got := callJSON(t, NewApplyEditTool(svc, enc), map[string]any{"base_revision": rev, "edits": []any{tc.edit}})
			if r, _ := got["refusal"].(map[string]any); got["ok"] != false || r["reason"] != tc.reason {
				t.Fatalf("got %v, want refusal %s", got, tc.reason)
			}
			if tc.reason == "compile_errors" && got["diagnostics"] == nil {
				t.Error("a compile refusal carries the diagnostics")
			}
			if after := tree(t, root); !reflect.DeepEqual(before, after) {
				t.Errorf("a refused write changed files")
			}
		})
	}
}

// TestToolsWriteOnlyHCL is SC-005, FR-026 and FR-026a: driving every tool,
// with every kind of edit, creates or changes nothing but *.loko.hcl files,
// and setting docs never creates the file it names.
func TestToolsWriteOnlyHCL(t *testing.T) {
	t.Parallel()
	root := handwritten(t)
	svc, enc := newService(root), encoding.NewEncoder()
	validate, apply := NewValidateTool(svc, enc), NewApplyEditTool(svc, enc)
	before := tree(t, root)
	for _, tl := range []toolCaller{NewDescribeTool(svc, enc), NewQueryTool(svc, enc), validate} {
		callJSON(t, tl, map[string]any{"kind": "orphans", "level": "full"})
	}
	batches := [][]any{
		{edit("add", "element", "container.cache", "set", map[string]any{"system": "system.shop", "docs": "docs/cache.md", "tags": []any{"x"}})},
		{edit("add", "relationship", "container.cache.uses.api", "set", map[string]any{"target": "container.api"})},
		{edit("add", "environment", "deployment.dev", "set", map[string]any{"region": "eu-west-1"})},
		{edit("add", "group", "deployment.dev.node.vpc")},
		{edit("add", "instance", "deployment.dev.node.vpc.instance.cache", "set", map[string]any{"of": "container.cache",
			"attributes": map[string]any{"memory": 512.0}})},
		{edit("add", "binding", "deployment.dev.instance.cache", "binding", map[string]any{"kind": "terraform"},
			"set", map[string]any{"address": "module.cache"})},
		{edit("update", "element", "system.shop", "set", map[string]any{"docs": "docs/shop/index.md"}, "clear", []any{"owner"})},
		{edit("add", "element", "system.extra", "file", "more/extra.loko.hcl")},
		{edit("remove", "element", "container.cache", "cascade", true)},
		{edit("remove", "environment", "deployment.dev", "cascade", true)},
	}
	for i, b := range batches {
		res := callJSON(t, apply, map[string]any{"base_revision": readRevision(t, validate), "edits": b})
		if res["ok"] != true {
			t.Fatalf("batch %d refused: %v", i, res["refusal"])
		}
	}
	moved := callJSON(t, NewMoveTool(svc, enc), map[string]any{"from": "system.payments", "to": "system.billing",
		"base_revision": readRevision(t, validate)})
	if moved["ok"] != true {
		t.Fatalf("move refused: %v", moved["refusal"])
	}
	after := tree(t, root)
	for _, p := range slices.Sorted(maps.Keys(after)) {
		if before[p] != after[p] && !strings.HasSuffix(p, ".loko.hcl") {
			t.Errorf("a tool wrote %s, which is not architecture source", p)
		}
	}
	for _, p := range []string{"docs/cache.md", "docs/shop/index.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			t.Errorf("setting docs created %s (FR-026a)", p)
		}
	}
	if _, ok := after["more/extra.loko.hcl"]; !ok {
		t.Error("an explicit file was not created")
	}
}

// TestStaleRevisionIsPerFile is US3 scenario 4: an edit based on an old
// revision is refused if the file it changes was modified since, and the
// modification survives; an edit to an unchanged file still succeeds.
func TestStaleRevisionIsPerFile(t *testing.T) {
	t.Parallel()
	root := handwritten(t)
	svc, enc := newService(root), encoding.NewEncoder()
	apply := NewApplyEditTool(svc, enc)
	old := readRevision(t, NewValidateTool(svc, enc))
	views := filepath.Join(root, "views.loko.hcl")
	theirs := "# reviewed by hand\nview \"payments-path\" {\n  include = [system.payments, container.api]\n}\n"
	if err := os.WriteFile(views, []byte(theirs), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := callJSON(t, apply, map[string]any{"base_revision": old, "edits": []any{
		edit("update", "element", "system.payments", "set", map[string]any{"owner": "x"}),
		edit("remove", "element", "container.api", "cascade", true),
	}})
	r, _ := stale["refusal"].(map[string]any)
	if stale["ok"] != false || r["reason"] != "stale_revision" || !strings.Contains(r["detail"].(string), "views.loko.hcl") {
		t.Fatalf("edit touching the changed file: %v", stale)
	}
	if b, _ := os.ReadFile(views); string(b) != theirs {
		t.Error("the external change was overwritten")
	}
	ok := callJSON(t, apply, map[string]any{"base_revision": old, "edits": []any{
		edit("update", "element", "system.payments", "set", map[string]any{"owner": "x"}),
	}})
	if ok["ok"] != true {
		t.Errorf("an edit to an unchanged file must succeed on the old revision: %v", ok["refusal"])
	}
}
