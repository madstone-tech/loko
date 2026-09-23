package arch

import (
	"path/filepath"
	"testing"
)

// TestNormalizeFileIsIdempotent is the regression guard for a bug that only
// appeared with a relative --project: the parser names each file relatively,
// so normalising the result a second time against the root produced
// "../../arch.loko.hcl". With an absolute root the second pass silently failed
// and the path survived, which is why absolute-path tests never caught it.
func TestNormalizeFileIsIdempotent(t *testing.T) {
	t.Parallel()

	roots := []string{
		"/abs/project",
		"./relative/project",
		"relative/project",
		"",
	}
	for _, root := range roots {
		t.Run("root="+root, func(t *testing.T) {
			t.Parallel()
			once := NormalizeFile(root, "arch.loko.hcl")
			if once != "arch.loko.hcl" {
				t.Errorf("NormalizeFile(%q, %q) = %q, want it unchanged", root, "arch.loko.hcl", once)
			}
			if twice := NormalizeFile(root, once); twice != once {
				t.Errorf("normalising twice changed %q to %q", once, twice)
			}
		})
	}
}

func TestNormalizeFileMakesAbsolutePathsRelative(t *testing.T) {
	t.Parallel()

	root := filepath.Join(string(filepath.Separator), "tmp", "proj")
	abs := filepath.Join(root, "sub", "arch.loko.hcl")

	if got, want := NormalizeFile(root, abs), "sub/arch.loko.hcl"; got != want {
		t.Errorf("NormalizeFile = %q, want %q", got, want)
	}
}

func TestNormalizeFileNestedRelative(t *testing.T) {
	t.Parallel()

	if got, want := NormalizeFile("/abs/project", "systems/payments.loko.hcl"), "systems/payments.loko.hcl"; got != want {
		t.Errorf("NormalizeFile = %q, want %q", got, want)
	}
}
