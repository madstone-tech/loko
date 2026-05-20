package main

import (
	"testing"
)

func makeFileSizeRules() []FileSizeRule {
	return []FileSizeRule{
		{
			Name:              "entity-file-size",
			PathPattern:       "internal/core/entities/**/*.go",
			MaxEffectiveLines: 5,
			Description:       "test entity budget",
		},
	}
}

func makeFileSizeExemptions() []Exemption {
	return []Exemption{
		{
			Kind:   "file-size",
			Match:  ExemptionMatch{Basename: []string{"schemas.go", "registry.go", "helpers.go", "constants.go"}},
			Reason: "Pure-data files",
		},
		{
			Kind:   "file-size",
			Match:  ExemptionMatch{PathPattern: "**/*_cobra.go"},
			Reason: "Cobra flag wiring",
		},
		{
			Kind:   "file-size",
			Match:  ExemptionMatch{PathPattern: "**/*_test.go"},
			Reason: "Tests exempt",
		},
		{
			Kind:   "file-size",
			Match:  ExemptionMatch{GeneratedHeader: true},
			Reason: "Generated files",
		},
	}
}

func TestCheckFileSize(t *testing.T) {
	rules := makeFileSizeRules()
	exemptions := makeFileSizeExemptions()

	t.Run("under budget passes", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/small.go",
			EffectiveLines: 3,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations, got %v", violations)
		}
	})

	t.Run("exactly at budget passes", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/exact.go",
			EffectiveLines: 5,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations at exact budget, got %v", violations)
		}
	})

	t.Run("over budget fails", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/big.go",
			EffectiveLines: 10,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(violations))
		}
		v := violations[0]
		if v.Rule != "entity-file-size" {
			t.Errorf("rule = %q, want %q", v.Rule, "entity-file-size")
		}
		if v.Kind != "file-size" {
			t.Errorf("kind = %q, want %q", v.Kind, "file-size")
		}
		if v.Actual != 10 {
			t.Errorf("actual = %d, want 10", v.Actual)
		}
		if v.Limit != 5 {
			t.Errorf("limit = %d, want 5", v.Limit)
		}
		if v.Line != 1 {
			t.Errorf("line = %d, want 1", v.Line)
		}
		if v.Subject != "big.go" {
			t.Errorf("subject = %q, want %q", v.Subject, "big.go")
		}
	})

	t.Run("exempt basename schemas.go", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/schemas.go",
			EffectiveLines: 100,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations for schemas.go, got %v", violations)
		}
	})

	t.Run("exempt *_cobra.go via pathPattern", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/new_cobra.go",
			EffectiveLines: 100,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations for *_cobra.go, got %v", violations)
		}
	})

	t.Run("exempt *_test.go", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/foo_test.go",
			EffectiveLines: 100,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations for *_test.go, got %v", violations)
		}
	})

	t.Run("exempt generated header", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "internal/core/entities/generated.go",
			EffectiveLines: 100,
			Generated:      true,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations for generated file, got %v", violations)
		}
	})

	t.Run("non-matching path produces no violations", func(t *testing.T) {
		f := &ParsedFile{
			Path:           "cmd/main.go",
			EffectiveLines: 1000,
		}
		violations := CheckFileSize(f, rules, exemptions)
		if len(violations) != 0 {
			t.Errorf("expected no violations for unmatched path, got %v", violations)
		}
	})
}
