package arch

import (
	"fmt"
	"strconv"
	"strings"
)

// Semantic-version constraint evaluation, standard library only.
//
// A dependency would do this in one line, but the check is one of FR-028's
// errors and so belongs in the validation layer, and core takes no
// dependencies (Constitution, Principle I). Putting it behind a port to
// satisfy the letter of the rule would mean an interface, an adapter, a mock
// and wiring for arithmetic on three integers, which Principle VII forbids.
//
// Supported operators: =, !=, >, >=, <, <=, ~>, and comma-separated
// conjunctions such as ">= 1.0, < 2.0". An operator-less constraint means
// exact equality.

// Version is a parsed semantic version.
type Version struct {
	Major, Minor, Patch int
	PreRelease          string
}

// ParseVersion parses "1.2.3", "1.2", "1", or any of those with a "v" prefix
// and an optional "-prerelease" or "+build" suffix.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return Version{}, fmt.Errorf("empty version")
	}

	// Build metadata never affects ordering, so it is discarded.
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}

	var pre string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		pre, s = s[i+1:], s[:i]
	}

	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return Version{}, fmt.Errorf("too many version segments in %q", s)
	}

	v := Version{PreRelease: pre}
	targets := []*int{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid version segment %q", p)
		}
		*targets[i] = n
	}
	return v, nil
}

// String renders the canonical form.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	return s
}

// Compare orders two versions per semantic-versioning precedence: a
// pre-release sorts before the release it precedes.
func (v Version) Compare(o Version) int {
	if c := cmpInts(v.Major, o.Major); c != 0 {
		return c
	}
	if c := cmpInts(v.Minor, o.Minor); c != 0 {
		return c
	}
	if c := cmpInts(v.Patch, o.Patch); c != 0 {
		return c
	}
	return comparePreRelease(v.PreRelease, o.PreRelease)
}

// comparePreRelease implements the semver rule that a version WITH a
// pre-release is lower than the same version without one.
func comparePreRelease(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	}

	ai, bi := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(ai) && i < len(bi); i++ {
		if c := comparePreSegment(ai[i], bi[i]); c != 0 {
			return c
		}
	}
	return cmpInts(len(ai), len(bi))
}

// comparePreSegment compares two dot-separated identifiers: numeric ones
// compare numerically and sort below alphanumeric ones.
func comparePreSegment(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	switch {
	case aErr == nil && bErr == nil:
		return cmpInts(an, bn)
	case aErr == nil:
		return -1
	case bErr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func cmpInts(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// ConstraintSatisfied reports whether version satisfies constraint. Every
// comma-separated term must hold.
func ConstraintSatisfied(constraint, version string) (bool, error) {
	v, err := ParseVersion(version)
	if err != nil {
		return false, fmt.Errorf("running version %q: %w", version, err)
	}

	terms := strings.SplitSeq(constraint, ",")
	for term := range terms {
		ok, termErr := satisfiesTerm(strings.TrimSpace(term), v)
		if termErr != nil {
			return false, termErr
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// operators are tried longest-first so ">=" is not read as ">".
var operators = []string{"~>", ">=", "<=", "!=", "=", ">", "<"}

func satisfiesTerm(term string, v Version) (bool, error) {
	if term == "" {
		return false, fmt.Errorf("empty constraint term")
	}

	op, rest := "=", term
	for _, candidate := range operators {
		if strings.HasPrefix(term, candidate) {
			op, rest = candidate, strings.TrimSpace(term[len(candidate):])
			break
		}
	}

	if strings.ContainsAny(rest, "*x") {
		return false, fmt.Errorf(
			"wildcard constraint %q is not supported; use ~> for a compatible-version range", term)
	}

	want, err := ParseVersion(rest)
	if err != nil {
		return false, fmt.Errorf("constraint %q: %w", term, err)
	}

	c := v.Compare(want)
	switch op {
	case "=":
		return c == 0, nil
	case "!=":
		return c != 0, nil
	case ">":
		return c > 0, nil
	case ">=":
		return c >= 0, nil
	case "<":
		return c < 0, nil
	case "<=":
		return c <= 0, nil
	case "~>":
		return satisfiesPessimistic(rest, want, v), nil
	default:
		return false, fmt.Errorf("unknown operator in constraint %q", term)
	}
}

// satisfiesPessimistic implements ~>, which allows the rightmost stated
// segment to increase and pins everything to its left.
//
//	~> 1.0    allows >= 1.0.0 and < 2.0.0
//	~> 1.2.3  allows >= 1.2.3 and < 1.3.0
func satisfiesPessimistic(raw string, want, v Version) bool {
	if v.Compare(want) < 0 {
		return false
	}

	segments := len(strings.Split(strings.TrimPrefix(strings.TrimSpace(raw), "v"), "."))
	switch segments {
	case 1:
		// ~> 1 is every 1.x.y.
		return v.Major == want.Major
	case 2:
		return v.Major == want.Major
	default:
		return v.Major == want.Major && v.Minor == want.Minor
	}
}
