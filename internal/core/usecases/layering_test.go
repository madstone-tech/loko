package usecases

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestUseCasesDoNoFileIO: use cases reach the file system only through ports,
// so no non-test file here imports os (Principle I; feature 018, R9, which
// moved docs resolution into the source adapter).
func TestUseCasesDoNoFileIO(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		ast, err := parser.ParseFile(fset, f, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range ast.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); p == "os" {
				t.Errorf("%s imports os; use cases reach files through a port", f)
			}
		}
	}
}
