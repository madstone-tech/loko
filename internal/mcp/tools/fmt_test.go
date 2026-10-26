package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

const misformatted = "system \"s\" {\ndescription=\"x\"\n}\n"

func TestFmtTool(t *testing.T) {
	t.Parallel()
	root := fixture(t, "two-systems")
	path := filepath.Join(root, "extra.loko.hcl")
	if err := os.WriteFile(path, []byte(misformatted), 0o644); err != nil {
		t.Fatal(err)
	}
	svc, enc := newService(root), encoding.NewEncoder()
	base := callJSON(t, NewValidateTool(svc, enc), map[string]any{})["revision"].(string)
	tool := NewFmtTool(svc, enc)

	prev := callJSON(t, tool, map[string]any{"base_revision": base, "preview": true})
	if prev["ok"] != true || len(prev["files"].([]any)) != 1 {
		t.Fatalf("preview: %v", prev)
	}
	if got, _ := os.ReadFile(path); string(got) != misformatted {
		t.Error("preview wrote the file")
	}
	done := callJSON(t, tool, map[string]any{"base_revision": base})
	if done["ok"] != true {
		t.Fatalf("apply: %v", done)
	}
	if got, _ := os.ReadFile(path); string(got) != "system \"s\" {\n  description = \"x\"\n}\n" {
		t.Errorf("formatted file:\n%s", got)
	}
	stale := callJSON(t, tool, map[string]any{"base_revision": "r1-0000000000000000"})
	if stale["ok"] != false || stale["refusal"].(map[string]any)["reason"] != "stale_revision" {
		t.Errorf("stale: %v", stale)
	}
}
