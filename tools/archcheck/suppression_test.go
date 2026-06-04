package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	return p
}

func TestLoadSuppressions_MissingFileReturnsEmpty(t *testing.T) {
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	got, errs, err := LoadSuppressions("/does/not/exist.yaml", now, nil)
	if err != nil {
		t.Fatalf("unexpected I/O error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected validation errors: %v", errs)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 entries, got %d", len(got))
	}
}

func TestLoadSuppressions_RejectsLongExpiry(t *testing.T) {
	yaml := `
- rule: cli-handler-func-size
  file: cmd/legacy.go
  function: runLegacy
  owner: "@andhi"
  expires_on: "2099-01-01"
  reason: "Far-future expiry must be rejected by 90-day cap"
`
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	path := writeTempFile(t, "suppressions.yaml", yaml)
	known := map[string]bool{"cli-handler-func-size": true}
	_, errs, err := LoadSuppressions(path, now, known)
	if err != nil {
		t.Fatalf("I/O error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected validation error for 90-day cap")
	}
}

func TestLoadSuppressions_RejectsUnknownRule(t *testing.T) {
	yaml := `
- rule: not-a-real-rule
  file: cmd/foo.go
  owner: "@andhi"
  expires_on: "2026-07-01"
  reason: "Twenty-character reason here please"
`
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	path := writeTempFile(t, "suppressions.yaml", yaml)
	known := map[string]bool{"cli-handler-func-size": true}
	_, errs, err := LoadSuppressions(path, now, known)
	if err != nil {
		t.Fatalf("I/O error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected validation error for unknown rule")
	}
}

func TestApplySuppressions_MatchesByRuleAndFile(t *testing.T) {
	violations := []Violation{
		{Rule: "cli-handler-func-size", File: "cmd/new.go", Subject: "runNew", Kind: "function-size"},
		{Rule: "cli-handler-func-size", File: "cmd/build.go", Subject: "runBuild", Kind: "function-size"},
	}
	supps := []Suppression{{
		Rule: "cli-handler-func-size", File: "cmd/new.go", Function: "runNew",
		Owner: "@andhi", ExpiresOn: "2026-07-01",
		Reason: "Twenty-character reason here please",
	}}
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	kept, suppressed, stale := ApplySuppressions(violations, supps, now)

	if len(kept) != 1 || kept[0].File != "cmd/build.go" {
		t.Fatalf("want cmd/build.go kept, got %+v", kept)
	}
	if len(suppressed) != 1 || suppressed[0].File != "cmd/new.go" {
		t.Fatalf("want cmd/new.go suppressed, got %+v", suppressed)
	}
	if len(stale) != 0 {
		t.Fatalf("want 0 stale, got %d", len(stale))
	}
}

func TestApplySuppressions_ExpiredDoesNotSuppress(t *testing.T) {
	violations := []Violation{{
		Rule: "cli-handler-func-size", File: "cmd/old.go", Subject: "runOld", Kind: "function-size",
	}}
	supps := []Suppression{{
		Rule: "cli-handler-func-size", File: "cmd/old.go", Function: "runOld",
		Owner: "@andhi", ExpiresOn: "2026-04-01", // before now
		Reason: "Was supposed to be fixed by April",
	}}
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	kept, suppressed, stale := ApplySuppressions(violations, supps, now)
	if len(kept) != 1 {
		t.Fatalf("expired suppression must not silence violation; got kept=%d", len(kept))
	}
	if len(suppressed) != 0 {
		t.Fatalf("expected 0 suppressed, got %d", len(suppressed))
	}
	// 50 days past expiry → stale warning
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale, got %d", len(stale))
	}
}

func TestApplySuppressions_FileGlob(t *testing.T) {
	violations := []Violation{
		{Rule: "outer-no-entities", File: "internal/mcp/tools/build_docs.go", Kind: "layer-import"},
		{Rule: "outer-no-entities", File: "internal/mcp/tools/analyze.go", Kind: "layer-import"},
		{Rule: "outer-no-entities", File: "internal/api/handlers/handlers.go", Kind: "layer-import"},
	}
	supps := []Suppression{{
		Rule: "outer-no-entities", File: "internal/mcp/tools/*.go",
		Owner: "@andhi", ExpiresOn: "2026-07-01",
		Reason: "Legacy MCP tools — tracking #789",
	}}
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	kept, suppressed, _ := ApplySuppressions(violations, supps, now)
	if len(kept) != 1 || kept[0].File != "internal/api/handlers/handlers.go" {
		t.Fatalf("api file should remain kept; got %+v", kept)
	}
	if len(suppressed) != 2 {
		t.Fatalf("both mcp files should be suppressed; got %d", len(suppressed))
	}
}
