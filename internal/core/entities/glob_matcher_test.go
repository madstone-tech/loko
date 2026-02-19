package entities

import (
	"testing"
)

func TestGlobMatcher_ExactMatch(t *testing.T) {
	m := NewGlobMatcher("payment-service")

	tests := []struct {
		text     string
		expected bool
	}{
		{"payment-service", true},
		{"payment-api", false},
		{"order-service", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_PrefixWildcard(t *testing.T) {
	m := NewGlobMatcher("payment*")

	tests := []struct {
		text     string
		expected bool
	}{
		{"payment", true},
		{"payment-service", true},
		{"payment-api", true},
		{"payments", true},
		{"order-service", false},
		{"user-payment", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_SuffixWildcard(t *testing.T) {
	m := NewGlobMatcher("*-service")

	tests := []struct {
		text     string
		expected bool
	}{
		{"payment-service", true},
		{"order-service", true},
		{"user-service", true},
		{"service", false}, // No hyphen before "service"
		{"payment-api", false},
		{"service-payment", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_MiddleWildcard(t *testing.T) {
	m := NewGlobMatcher("api-*-handler")

	tests := []struct {
		text     string
		expected bool
	}{
		{"api-auth-handler", true},
		{"api-payment-handler", true},
		{"api-handler", true},
		{"api-auth", false},
		{"auth-handler", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_MultipleWildcards(t *testing.T) {
	m := NewGlobMatcher("*-api-*")

	tests := []struct {
		text     string
		expected bool
	}{
		{"payment-api-service", true},
		{"user-api-handler", true},
		{"api-service", false}, // No hyphen before "api"
		{"payment-api", false}, // No hyphen after "api"
		{"payment-service", false},
		{"-api-", true}, // Minimal match
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_MatchAll(t *testing.T) {
	m := NewGlobMatcher("*")

	tests := []string{
		"payment-service",
		"api",
		"",
		"any-thing-at-all",
	}

	for _, text := range tests {
		if !m.Match(text) {
			t.Errorf("Match(%q) = false, want true (should match everything)", text)
		}
	}
}

func TestGlobMatcher_SingleCharWildcard(t *testing.T) {
	m := NewGlobMatcher("api-?")

	tests := []struct {
		text     string
		expected bool
	}{
		{"api-1", true},
		{"api-a", true},
		{"api-x", true},
		{"api-10", false},
		{"api-", false},
		{"api", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestGlobMatcher_MixedWildcards(t *testing.T) {
	m := NewGlobMatcher("api-?-*")

	tests := []struct {
		text     string
		expected bool
	}{
		{"api-1-service", true},
		{"api-a-handler", true},
		{"api-x-", true},
		{"api-10-service", false}, // ? matches single char
		{"api-service", false},
	}

	for _, tt := range tests {
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("Match(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestMatchAny(t *testing.T) {
	patterns := []string{"payment*", "*-service", "api-?"}

	tests := []struct {
		text     string
		expected bool
	}{
		{"payment-api", true},     // matches payment*
		{"order-service", true},   // matches *-service
		{"api-1", true},           // matches api-?
		{"payment-service", true}, // matches both payment* and *-service
		{"user-handler", false},   // matches none
	}

	for _, tt := range tests {
		if got := MatchAny(tt.text, patterns); got != tt.expected {
			t.Errorf("MatchAny(%q, %v) = %v, want %v", tt.text, patterns, got, tt.expected)
		}
	}
}

func TestGlobMatcher_EdgeCases(t *testing.T) {
	tests := []struct {
		pattern  string
		text     string
		expected bool
	}{
		{"", "", true},           // empty matches empty
		{"", "text", false},      // empty doesn't match non-empty
		{"text", "", false},      // non-empty doesn't match empty
		{"**", "anything", true}, // multiple wildcards
		{"*a*b*", "aXb", true},   // overlapping wildcards
		{"*a*b*", "ab", true},    // consecutive letters
		{"*a*b*", "ba", false},   // wrong order
	}

	for _, tt := range tests {
		m := NewGlobMatcher(tt.pattern)
		if got := m.Match(tt.text); got != tt.expected {
			t.Errorf("NewGlobMatcher(%q).Match(%q) = %v, want %v",
				tt.pattern, tt.text, got, tt.expected)
		}
	}
}
