package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func makeTestReport() *Report {
	return &Report{
		Version:               "1.0",
		GeneratedAt:           "2026-01-01T00:00:00Z",
		AuditToolVersion:      "abc123",
		RulesPath:             "specs/rules.yaml",
		TotalFilesScanned:     10,
		TotalFunctionsScanned: 5,
		Violations: []Violation{
			{
				Rule:    "entity-file-size",
				Kind:    "file-size",
				File:    "internal/core/entities/graph.go",
				Line:    1,
				Subject: "graph.go",
				Actual:  400,
				Limit:   300,
				Message: "internal/core/entities/graph.go:1: graph.go exceeds entity-file-size (400 > 300)",
			},
			{
				Rule:    "cli-handler-func-size",
				Kind:    "function-size",
				File:    "cmd/new.go",
				Line:    42,
				Subject: "runScaffold",
				Actual:  67,
				Limit:   50,
				Message: "cmd/new.go:42: runScaffold exceeds cli-handler-func-size (67 > 50)",
			},
			{
				Rule:    "core/usecases",
				Kind:    "layer",
				File:    "internal/core/usecases/build_docs.go",
				Line:    8,
				Subject: "github.com/madstone-tech/loko/cmd",
				Actual:  0,
				Limit:   0,
				Message: "internal/core/usecases/build_docs.go:8: layer 'core/usecases' may not import 'github.com/madstone-tech/loko/cmd' (rule: Use-case layer)",
			},
		},
		ExitCode: 1,
	}
}

func TestWriteText_Order(t *testing.T) {
	report := makeTestReport()
	var buf bytes.Buffer
	WriteText(&buf, report, false)
	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// Should have 4 lines: 3 violations + 1 summary
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), output)
	}

	// Sorted order: layer < file-size < function-size
	if !strings.Contains(lines[0], "layer") || !strings.Contains(lines[0], "core/usecases") {
		t.Errorf("expected layer violation first, got: %s", lines[0])
	}
	// Summary line
	if !strings.Contains(lines[3], "archcheck:") {
		t.Errorf("expected summary last, got: %s", lines[3])
	}
	if !strings.Contains(lines[3], "10 file(s)") {
		t.Errorf("summary missing file count: %s", lines[3])
	}
	if !strings.Contains(lines[3], "3 violation(s)") {
		t.Errorf("summary missing violation count: %s", lines[3])
	}
}

func TestWriteText_NoAnnotations(t *testing.T) {
	report := makeTestReport()
	var buf bytes.Buffer
	WriteText(&buf, report, false)
	output := buf.String()
	if strings.Contains(output, "::error") {
		t.Errorf("no GitHub annotations expected, but found ::error in output")
	}
}

func TestWriteText_WithAnnotations(t *testing.T) {
	report := makeTestReport()
	var buf bytes.Buffer
	WriteText(&buf, report, true)
	output := buf.String()
	count := strings.Count(output, "::error")
	if count != len(report.Violations) {
		t.Errorf("expected %d ::error annotations, got %d", len(report.Violations), count)
	}
}

func TestWriteText_SummaryAlwaysPrinted(t *testing.T) {
	report := &Report{
		TotalFilesScanned:     3,
		TotalFunctionsScanned: 1,
		Violations:            nil,
		ExitCode:              0,
	}
	var buf bytes.Buffer
	WriteText(&buf, report, false)
	output := buf.String()
	if !strings.Contains(output, "archcheck:") {
		t.Error("summary must always be printed even with no violations")
	}
	if !strings.Contains(output, "0 violation(s)") {
		t.Errorf("expected 0 violations in summary, got: %s", output)
	}
}

func TestWriteJSON_RoundTrip(t *testing.T) {
	original := makeTestReport()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, original); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf("version mismatch: %q != %q", decoded.Version, original.Version)
	}
	if decoded.TotalFilesScanned != original.TotalFilesScanned {
		t.Errorf("files scanned mismatch: %d != %d", decoded.TotalFilesScanned, original.TotalFilesScanned)
	}
	if len(decoded.Violations) != len(original.Violations) {
		t.Errorf("violation count mismatch: %d != %d", len(decoded.Violations), len(original.Violations))
	}
	for i, v := range decoded.Violations {
		if v.Rule != original.Violations[i].Rule {
			t.Errorf("[%d] rule mismatch: %q != %q", i, v.Rule, original.Violations[i].Rule)
		}
	}
}

func TestSortedViolations(t *testing.T) {
	input := []Violation{
		{Kind: "layer", File: "b.go", Line: 2, Subject: "x"},
		{Kind: "file-size", File: "a.go", Line: 1, Subject: "y"},
		{Kind: "function-size", File: "a.go", Line: 1, Subject: "z"},
		{Kind: "file-size", File: "a.go", Line: 1, Subject: "a"},
	}
	sorted := sortedViolations(input)

	if sorted[0].Kind != "layer" || sorted[0].Subject != "x" {
		t.Errorf("expected first: layer x, got %s/%s", sorted[0].Kind, sorted[0].Subject)
	}
	if sorted[1].Kind != "file-size" || sorted[1].Subject != "a" {
		t.Errorf("expected second: file-size a, got %s/%s", sorted[1].Kind, sorted[1].Subject)
	}
	if sorted[2].Kind != "file-size" || sorted[2].Subject != "y" {
		t.Errorf("expected third: file-size y, got %s/%s", sorted[2].Kind, sorted[2].Subject)
	}
	if sorted[3].Kind != "function-size" {
		t.Errorf("expected fourth: function-size, got %s", sorted[3].Kind)
	}
}
