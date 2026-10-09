package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// TestNamingRuleParity keeps authoring.ValidName, a copy that exists because
// entity packages may not import one another, in step with arch.ValidName.
func TestNamingRuleParity(t *testing.T) {
	t.Parallel()
	names := []string{
		"a", "A", "_", "_x", "x_", "a-", "a-b", "a_b", "ab1", "a1-2_c", "Z9", "__", "_-",
		"1a", "9", "-a", "-", "", "a.b", "a b", "a/b", "a:b", "a@b", "a+b", "a$", "é", "aé",
		"日本", "a\x00", "a\n", "a\t", "a,b", "a'b", `a"b`, "a=b", "(a)", "a*", "a~b", "a!", "x" + string(rune(0x2010)),
	}
	if len(names) < 40 {
		t.Fatalf("table has %d names, want at least 40", len(names))
	}
	for _, n := range names {
		if a, b := arch.ValidName(n), authoring.ValidName(n); a != b {
			t.Errorf("%q: arch.ValidName=%v authoring.ValidName=%v", n, a, b)
		}
	}
}
