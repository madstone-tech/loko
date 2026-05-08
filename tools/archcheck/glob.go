package main

import (
	"regexp"
	"strings"
)

// matchPath reports whether the slash-separated path matches the glob pattern.
//
// Supported glob syntax:
//   - "**"  matches any number of path segments (including zero)
//   - "*"   matches any sequence of characters within a single segment
//   - "?"   matches any single character within a segment
//   - All other characters are literal (regex metacharacters are escaped).
//
// Pattern "foo/**" matches "foo", "foo/bar", and "foo/bar/baz".
// Pattern "**/*_test.go" matches "foo_test.go" and "cmd/foo_test.go".
func matchPath(pattern, path string) bool {
	re, err := globToRegexp(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(path)
}

// globToRegexp converts a glob pattern to an anchored regular expression.
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	var sb strings.Builder
	sb.WriteString(`^`)

	i := 0
	for i < len(pattern) {
		// Check for "/**/" — zero-or-more intermediate segments
		if strings.HasPrefix(pattern[i:], "/**/") {
			// Matches "/" optionally followed by segments and "/"
			sb.WriteString(`(?:/[^/]+)*/`)
			i += 4
			continue
		}
		// Check for "/**" at end of pattern — optional trailing path
		if strings.HasPrefix(pattern[i:], "/**") && i+3 == len(pattern) {
			sb.WriteString(`(?:/.*)?`)
			i += 3
			continue
		}
		// Check for standalone "**" at start (possibly followed by "/")
		if strings.HasPrefix(pattern[i:], "**/") {
			sb.WriteString(`(?:[^/]+/)*`)
			i += 3
			continue
		}
		// Standalone "**" with nothing after (treat as wildcard)
		if strings.HasPrefix(pattern[i:], "**") && i+2 == len(pattern) {
			sb.WriteString(`.*`)
			i += 2
			continue
		}
		// Single "*" — matches within one segment
		if pattern[i] == '*' {
			sb.WriteString(`[^/]*`)
			i++
			continue
		}
		// "?" — matches single non-separator char
		if pattern[i] == '?' {
			sb.WriteString(`[^/]`)
			i++
			continue
		}
		// Literal character — escape regex metacharacters
		ch := string(pattern[i])
		sb.WriteString(regexp.QuoteMeta(ch))
		i++
	}

	sb.WriteString(`$`)
	return regexp.Compile(sb.String())
}
