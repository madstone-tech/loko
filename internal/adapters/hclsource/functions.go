package hclsource

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// Functions is the complete function set of the language (FR-017): exactly
// five, plus native string interpolation.
//
// All five come from cty's standard library rather than being hand-written.
// They already handle null and unknown values correctly and carry upstream
// tests; reimplementing them would add surface and subtract correctness.
//
// The set is deliberately under-shipped. Adding a function later is painless;
// removing one is a breaking change for every file in every user's repository,
// and this language has no migration command.
var Functions = map[string]function.Function{
	"join":    stdlib.JoinFunc,
	"split":   stdlib.SplitFunc,
	"lower":   stdlib.LowerFunc,
	"upper":   stdlib.UpperFunc,
	"replace": stdlib.ReplaceFunc,
}

// FunctionNames returns the supported function names, sorted. Error messages
// list them so the diagnostic itself documents the available set rather than
// sending the reader to the manual.
func FunctionNames() []string {
	names := make([]string, 0, len(Functions))
	for name := range Functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// unknownFunctionDetail builds the detail text for an unknown-function error
// (FR-017a).
func unknownFunctionDetail(called string) string {
	return fmt.Sprintf(
		"%q is not a function in this language. The five available functions are: %s. "+
			"Iteration, dynamic blocks, variables, and modules are deliberately excluded from v1.0.",
		called, strings.Join(FunctionNames(), ", "))
}
