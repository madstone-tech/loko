package arch

import "testing"

func TestConstraintSatisfied(t *testing.T) {
	t.Parallel()

	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{"= 1.0.0", "1.0.0", true},
		{"1.0.0", "1.0.0", true},
		{"= 1.0.0", "1.0.1", false},
		{"!= 1.0.0", "1.0.1", true},
		{"> 1.0.0", "1.0.1", true},
		{"> 1.0.0", "1.0.0", false},
		{">= 1.0.0", "1.0.0", true},
		{"< 2.0.0", "1.9.9", true},
		{"< 2.0.0", "2.0.0", false},
		{"<= 2.0.0", "2.0.0", true},

		// ~> allows the rightmost stated segment to move.
		{"~> 1.0", "1.0.0", true},
		{"~> 1.0", "1.4.2", true},
		{"~> 1.0", "2.0.0", false},
		{"~> 1.0", "0.9.0", false},
		{"~> 1.2.3", "1.2.9", true},
		{"~> 1.2.3", "1.3.0", false},
		{"~> 1.2.3", "1.2.2", false},
		{"~> 1", "1.9.9", true},
		{"~> 1", "2.0.0", false},

		// Conjunctions: every term must hold.
		{">= 1.0, < 2.0", "1.5.0", true},
		{">= 1.0, < 2.0", "2.0.1", false},
		{">= 1.0, < 2.0", "0.9.0", false},

		// Pre-release sorts below the release it precedes.
		{">= 1.0.0", "1.0.0-rc1", false},
		{"> 1.0.0-rc1", "1.0.0", true},
		{"= 1.0.0-rc1", "1.0.0-rc1", true},

		// Shorthand versions and the v prefix.
		{">= 1", "1.0.0", true},
		{">= v1.2", "1.2.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+" vs "+tt.version, func(t *testing.T) {
			t.Parallel()
			got, err := ConstraintSatisfied(tt.constraint, tt.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ConstraintSatisfied(%q, %q) = %v, want %v",
					tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

func TestConstraintErrors(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"", ">= ", "~> 1.2.*", ">= abc", "1.2.3.4", ">= 1.0,"} {
		t.Run(bad, func(t *testing.T) {
			t.Parallel()
			if _, err := ConstraintSatisfied(bad, "1.0.0"); err == nil {
				t.Errorf("ConstraintSatisfied(%q, ...) returned no error", bad)
			}
		})
	}

	if _, err := ConstraintSatisfied(">= 1.0", "not-a-version"); err == nil {
		t.Error("an invalid running version returned no error")
	}
}

func TestVersionCompareOrdering(t *testing.T) {
	t.Parallel()

	ordered := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-beta", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0"}
	for i := 1; i < len(ordered); i++ {
		a, _ := ParseVersion(ordered[i-1])
		b, _ := ParseVersion(ordered[i])
		if a.Compare(b) >= 0 {
			t.Errorf("%s should sort before %s", ordered[i-1], ordered[i])
		}
	}

	// Build metadata is ignored for ordering.
	x, _ := ParseVersion("1.0.0+build1")
	y, _ := ParseVersion("1.0.0+build2")
	if x.Compare(y) != 0 {
		t.Error("build metadata affected ordering")
	}
}
