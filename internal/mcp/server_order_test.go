package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
)

// TestToolsListSorted: tools/list must not depend on map iteration order.
func TestToolsListSorted(t *testing.T) {
	t.Parallel()
	server := NewServer("test", bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	for _, n := range []string{"query", "apply_edit", "validate", "move", "describe"} {
		if err := server.RegisterTool(&MockTool{NameValue: n}); err != nil {
			t.Fatal(err)
		}
	}
	want := "[apply_edit describe move query validate]"
	for range 20 {
		resp := server.handleRequest(t.Context(), map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/list"})
		tools := resp["result"].(map[string]any)["tools"].([]map[string]any)
		var names []string
		for _, tl := range tools {
			names = append(names, tl["name"].(string))
		}
		if got := fmt.Sprint(names); got != want {
			t.Fatalf("tools/list order = %s, want %s", got, want)
		}
	}
}

// TestToolCallReceivesRunContext: a cancelled server cancels the call.
func TestToolCallReceivesRunContext(t *testing.T) {
	t.Parallel()
	server := NewServer("test", bytes.NewBuffer(nil), bytes.NewBuffer(nil))
	var seen error
	_ = server.RegisterTool(&MockTool{NameValue: "probe", CallFunc: func(ctx context.Context, _ map[string]any) (any, error) {
		seen = ctx.Err()
		return "ok", nil
	}})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	server.handleRequest(ctx, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call",
		"params": map[string]any{"name": "probe"}})
	if !errors.Is(seen, context.Canceled) {
		t.Errorf("tool saw ctx.Err() = %v, want context.Canceled", seen)
	}
}
