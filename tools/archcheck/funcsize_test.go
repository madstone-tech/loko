package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func parseFuncTestFile(t *testing.T, src string) *ParsedFile {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return &ParsedFile{
		Path:    "cmd/handler.go",
		Source:  []byte(src),
		FileSet: fset,
		AST:     f,
	}
}

func makeFuncSizeRules() []FunctionSizeRule {
	return []FunctionSizeRule{
		{
			Name:              "cli-handler-func-size",
			PathPattern:       "cmd/**/*.go",
			MaxEffectiveLines: 6,
			Description:       "test func budget",
		},
	}
}

func makeFuncSizeExemptions() []Exemption {
	return []Exemption{
		{
			Kind:   "function-size",
			Match:  ExemptionMatch{PathPattern: "**/*_test.go"},
			Reason: "Tests exempt",
		},
	}
}

// shortFn has 3 effective lines inside (including func decl + closing brace).
const shortFnSrc = `package main

func ShortFn() {
	x := 1
	_ = x
}
`

// atBudgetFnSrc has exactly 5 effective lines in the function body region.
const atBudgetFnSrc = `package main

func AtBudget() {
	a := 1
	b := 2
	c := 3
	_ = a + b + c
}
`

// overBudgetFnSrc has 8 effective lines.
const overBudgetFnSrc = `package main

func OverBudget() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	_ = a + b + c + d + e + f
}
`

// commentHeavyFnSrc: lots of comments but only a few effective lines.
const commentHeavyFnSrc = `package main

// CommentHeavy does stuff.
func CommentHeavy() {
	// Step 1: set a
	a := 1
	// Step 2: set b
	b := 2
	// Step 3: use them
	_ = a + b
}
`

// methodFnSrc has a method on a type.
const methodFnSrc = `package main

type T struct{}

func (t *T) BigMethod() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	_ = a + b + c + d + e + f
}
`

// closureFnSrc: the top-level function is short; closure inside should NOT be counted separately.
const closureFnSrc = `package main

func OuterFn() {
	inner := func() {
		a := 1
		b := 2
		c := 3
		d := 4
		e := 5
		f := 6
		g := 7
		h := 8
		_ = a + b + c + d + e + f + g + h
	}
	inner()
}
`

func TestCheckFunctionSize(t *testing.T) {
	rules := makeFuncSizeRules()
	exemptions := makeFuncSizeExemptions()

	t.Run("short function passes", func(t *testing.T) {
		pf := parseFuncTestFile(t, shortFnSrc)
		pf.Path = "cmd/handler.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 0 {
			t.Errorf("expected no violations, got %v", got)
		}
	})

	t.Run("exactly at budget passes", func(t *testing.T) {
		pf := parseFuncTestFile(t, atBudgetFnSrc)
		pf.Path = "cmd/handler.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 0 {
			t.Errorf("expected no violations at budget, got %v", got)
		}
	})

	t.Run("over budget fails", func(t *testing.T) {
		pf := parseFuncTestFile(t, overBudgetFnSrc)
		pf.Path = "cmd/handler.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 1 {
			t.Fatalf("expected 1 violation, got %d: %v", len(got), got)
		}
		v := got[0]
		if v.Subject != "OverBudget" {
			t.Errorf("subject = %q, want OverBudget", v.Subject)
		}
		if v.Kind != "function-size" {
			t.Errorf("kind = %q, want function-size", v.Kind)
		}
	})

	t.Run("comment-heavy drops below limit", func(t *testing.T) {
		pf := parseFuncTestFile(t, commentHeavyFnSrc)
		pf.Path = "cmd/handler.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 0 {
			t.Errorf("comment-heavy fn should pass, got %v", got)
		}
	})

	t.Run("method on type is detected", func(t *testing.T) {
		pf := parseFuncTestFile(t, methodFnSrc)
		pf.Path = "cmd/handler.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 1 {
			t.Fatalf("expected 1 violation for method, got %d", len(got))
		}
		if got[0].Subject != "BigMethod" {
			t.Errorf("subject = %q, want BigMethod", got[0].Subject)
		}
	})

	t.Run("closure inside top-level fn NOT counted separately", func(t *testing.T) {
		pf := parseFuncTestFile(t, closureFnSrc)
		pf.Path = "cmd/handler.go"
		// OuterFn is the only top-level FuncDecl; its effective lines include
		// the closure body text but OuterFn itself may be short (the walk is
		// top-level FuncDecls only, so we should get 0 or 1 violation for OuterFn,
		// not an extra violation for the inner closure).
		got := CheckFunctionSize(pf, rules, exemptions)
		// Verify we only see at most 1 violation (OuterFn), not 2+
		for _, v := range got {
			if v.Subject != "OuterFn" {
				t.Errorf("unexpected subject %q — closure should not be reported separately", v.Subject)
			}
		}
	})

	t.Run("test file exempt", func(t *testing.T) {
		pf := parseFuncTestFile(t, overBudgetFnSrc)
		pf.Path = "cmd/handler_test.go"
		got := CheckFunctionSize(pf, rules, exemptions)
		if len(got) != 0 {
			t.Errorf("test file should be exempt, got %v", got)
		}
	})
}
