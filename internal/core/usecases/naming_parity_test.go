package usecases

import (
	"fmt"
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

// TestRenderValueSetParity keeps authoring's copies of the rendering value
// sets (feature 016) in step with arch's.
func TestRenderValueSetParity(t *testing.T) {
	t.Parallel()
	if fmt.Sprint(authoring.Shapes) != fmt.Sprint(arch.Shapes) ||
		fmt.Sprint(authoring.RelationshipKinds) != fmt.Sprint(arch.RelationshipKinds) ||
		fmt.Sprint(authoring.Directions) != fmt.Sprint(arch.Directions) ||
		fmt.Sprint(authoring.Layouts) != fmt.Sprint(arch.Layouts) {
		t.Error("authoring and arch disagree on the rendering value sets")
	}
}
