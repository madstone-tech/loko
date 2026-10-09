package tools

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/adapters/encoding"
)

func TestArgumentHelpers(t *testing.T) {
	t.Parallel()
	args := map[string]any{"s": "x", "empty": "", "n": 3.0, "frac": 1.5, "neg": -1.0, "b": true, "notbool": "yes"}
	if v, err := str(args, "s"); err != nil || v != "x" {
		t.Errorf("str: %q %v", v, err)
	}
	for _, k := range []string{"missing", "empty", "n"} {
		if _, err := str(args, k); err == nil || !strings.Contains(err.Error(), k) {
			t.Errorf("str(%q) = %v, want an error naming it", k, err)
		}
	}
	if v, err := optStr(args, "missing", "def"); err != nil || v != "def" {
		t.Errorf("optStr default: %q %v", v, err)
	}
	if v, err := intArg(args, "n"); err != nil || v != 3 {
		t.Errorf("intArg: %d %v", v, err)
	}
	for _, k := range []string{"frac", "neg", "s"} {
		if _, err := intArg(args, k); err == nil {
			t.Errorf("intArg(%q) accepted it", k)
		}
	}
	if v, err := boolArg(args, "b"); err != nil || !v {
		t.Errorf("boolArg: %v %v", v, err)
	}
	if _, err := boolArg(args, "notbool"); err == nil {
		t.Error("boolArg accepted a string")
	}
}

func TestFormatArg(t *testing.T) {
	t.Parallel()
	if f, _ := formatArg(map[string]any{}); f != "toon" {
		t.Errorf("default format = %q, want toon (FR-008)", f)
	}
	if _, err := formatArg(map[string]any{"format": "yaml"}); err == nil {
		t.Error("formatArg accepted yaml")
	}
}

func TestDecodeEdits(t *testing.T) {
	t.Parallel()
	edits, err := decodeEdits(map[string]any{"edits": []any{
		map[string]any{"op": "add", "target": "element", "address": "system.s",
			"set":     map[string]any{"description": "d", "tags": []any{"a"}},
			"binding": map[string]any{"kind": "terraform", "index": 1.0}},
	}})
	if err != nil || len(edits) != 1 || edits[0].Set["description"] != "d" || edits[0].Binding.Index != 1 {
		t.Fatalf("decodeEdits: %+v %v", edits, err)
	}
	for name, v := range map[string]any{
		"missing": nil, "empty": []any{}, "not an array": "x",
		"unknown field": []any{map[string]any{"op": "add", "adress": "system.s"}},
	} {
		args := map[string]any{"edits": v}
		if v == nil {
			args = map[string]any{}
		}
		if _, err := decodeEdits(args); err == nil {
			t.Errorf("%s: decodeEdits accepted it", name)
		}
	}
}

func TestEncodeResult(t *testing.T) {
	t.Parallel()
	enc := encoding.NewEncoder()
	v := struct {
		OK   bool   `json:"ok" toon:"ok"`
		Name string `json:"name" toon:"name"`
	}{true, "shop"}
	j, err := encodeResult(enc, v, "json")
	if err != nil || !strings.Contains(j.(string), `"ok":true`) {
		t.Errorf("json: %v %v", j, err)
	}
	tn, err := encodeResult(enc, v, "toon")
	if err != nil || strings.Contains(tn.(string), "{") || !strings.Contains(tn.(string), "shop") {
		t.Errorf("toon: %v %v", tn, err)
	}
}
