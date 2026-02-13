package entities

import (
	"strings"
)

// GlobMatcher provides glob pattern matching with wildcards (* and ?).
type GlobMatcher struct {
	pattern string
}

// NewGlobMatcher creates a new glob pattern matcher.
func NewGlobMatcher(pattern string) *GlobMatcher {
	return &GlobMatcher{pattern: pattern}
}

// Match returns true if the text matches the glob pattern.
// Supports wildcards:
//   - * matches zero or more characters
//   - ? matches exactly one character
//
// Examples:
//   - "payment*" matches "payment-service", "payment-api"
//   - "*-service" matches "payment-service", "order-service"
//   - "api-?" matches "api-1", "api-a" but not "api-10"
//   - "*" matches everything
func (m *GlobMatcher) Match(text string) bool {
	return globMatch(m.pattern, text)
}

// globMatch implements glob pattern matching.
// This is a simple implementation that handles * and ? wildcards.
func globMatch(pattern, text string) bool {
	// Fast path: exact match
	if pattern == text {
		return true
	}

	// Fast path: match everything
	if pattern == "*" {
		return true
	}

	// If pattern has ?, handle it separately
	if strings.Contains(pattern, "?") && !strings.Contains(pattern, "*") {
		return matchWithSingleChar(pattern, text)
	}

	// Mixed wildcards: handle ? within * segments
	if strings.Contains(pattern, "?") && strings.Contains(pattern, "*") {
		return globMatchMixed(pattern, text)
	}

	// Convert pattern to segments split by *
	segments := strings.Split(pattern, "*")

	// If no wildcards, pattern must equal text
	if len(segments) == 1 {
		return pattern == text
	}

	textPos := 0

	// Check first segment (before first *)
	if len(segments[0]) > 0 {
		if !strings.HasPrefix(text, segments[0]) {
			return false
		}
		textPos = len(segments[0])
	}

	// Check last segment (after last *)
	lastSegment := segments[len(segments)-1]
	if len(lastSegment) > 0 {
		if !strings.HasSuffix(text, lastSegment) {
			return false
		}
		// Reduce search space
		text = text[:len(text)-len(lastSegment)]
	}

	// Check middle segments
	for i := 1; i < len(segments)-1; i++ {
		segment := segments[i]
		if segment == "" {
			continue
		}

		idx := strings.Index(text[textPos:], segment)
		if idx == -1 {
			return false
		}
		textPos += idx + len(segment)
	}

	return true
}

// globMatchMixed handles patterns with both * and ? wildcards.
func globMatchMixed(pattern, text string) bool {
	pi, ti := 0, 0
	starIdx, matchIdx := -1, 0

	for ti < len(text) {
		if pi < len(pattern) {
			if pattern[pi] == '*' {
				starIdx = pi
				matchIdx = ti
				pi++
				continue
			} else if pattern[pi] == '?' || pattern[pi] == text[ti] {
				pi++
				ti++
				continue
			}
		}

		if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			ti = matchIdx
			continue
		}

		return false
	}

	// Handle remaining * in pattern
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}

	return pi == len(pattern)
}

// matchWithSingleChar matches pattern with ? wildcard (single character).
func matchWithSingleChar(pattern, text string) bool {
	if len(pattern) != len(text) {
		return false
	}

	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '?' && pattern[i] != text[i] {
			return false
		}
	}

	return true
}

// MatchAny returns true if the text matches any of the glob patterns.
func MatchAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if NewGlobMatcher(pattern).Match(text) {
			return true
		}
	}
	return false
}
