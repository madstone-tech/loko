package cli

import (
	"bufio"
	"strings"
	"testing"
)

func TestPromptString_WithValue(t *testing.T) {
	input := "my value\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptString("Enter text", "default")
	if result != "my value" {
		t.Errorf("expected 'my value', got %q", result)
	}
}

func TestPromptString_Empty_UseDefault(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptString("Enter text", "default")
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}
}

func TestPromptString_TrimWhitespace(t *testing.T) {
	input := "  trimmed  \n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptString("Enter text", "")
	if result != "trimmed" {
		t.Errorf("expected 'trimmed', got %q", result)
	}
}

func TestPromptStringMulti_CommaSeparated(t *testing.T) {
	input := "value1, value2, value3\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptStringMulti("Enter values")
	if len(result) != 3 {
		t.Errorf("expected 3 values, got %d", len(result))
	}
	if result[0] != "value1" {
		t.Errorf("expected 'value1', got %q", result[0])
	}
	if result[1] != "value2" {
		t.Errorf("expected 'value2', got %q", result[1])
	}
	if result[2] != "value3" {
		t.Errorf("expected 'value3', got %q", result[2])
	}
}

func TestPromptStringMulti_Empty(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptStringMulti("Enter values")
	if len(result) != 0 {
		t.Errorf("expected 0 values, got %d", len(result))
	}
}

func TestPromptStringMulti_TrimsWhitespace(t *testing.T) {
	input := "  value1  ,  value2  ,  value3  \n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptStringMulti("Enter values")
	if len(result) != 3 {
		t.Errorf("expected 3 values, got %d", len(result))
	}
	for i, val := range result {
		if strings.HasPrefix(val, " ") || strings.HasSuffix(val, " ") {
			t.Errorf("value %d not trimmed: %q", i, val)
		}
	}
}

func TestPromptYesNo_Yes(t *testing.T) {
	input := "y\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptYesNo("Continue?", false)
	if !result {
		t.Error("expected true, got false")
	}
}

func TestPromptYesNo_No(t *testing.T) {
	input := "n\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptYesNo("Continue?", true)
	if result {
		t.Error("expected false, got true")
	}
}

func TestPromptYesNo_Empty_UseDefault_True(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptYesNo("Continue?", true)
	if !result {
		t.Error("expected true (default), got false")
	}
}

func TestPromptYesNo_Empty_UseDefault_False(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptYesNo("Continue?", false)
	if result {
		t.Error("expected false (default), got true")
	}
}

func TestPromptYesNo_Full(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"yes\n", true},
		{"y\n", true},
		{"Y\n", true},
		{"YES\n", true},
		{"no\n", false},
		{"n\n", false},
		{"N\n", false},
	}

	for _, tt := range tests {
		reader := bufio.NewReader(strings.NewReader(tt.input))
		prompts := NewPrompts(reader)
		result := prompts.PromptYesNo("Continue?", false)
		if result != tt.expected {
			t.Errorf("input %q: expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestPromptSelect_ValidOption(t *testing.T) {
	input := "2\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	options := []string{"Option 1", "Option 2", "Option 3"}
	result := prompts.PromptSelect("Choose:", options)
	if result != "Option 2" {
		t.Errorf("expected 'Option 2', got %q", result)
	}
}

func TestPromptSelect_FirstOption(t *testing.T) {
	input := "1\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	options := []string{"Option 1", "Option 2"}
	result := prompts.PromptSelect("Choose:", options)
	if result != "Option 1" {
		t.Errorf("expected 'Option 1', got %q", result)
	}
}

func TestPromptSelect_InvalidOption(t *testing.T) {
	input := "99\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	options := []string{"Option 1", "Option 2"}
	result := prompts.PromptSelect("Choose:", options)
	if result != "" {
		t.Errorf("expected empty string for invalid option, got %q", result)
	}
}

func TestPromptSelect_EmptyOptions(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	result := prompts.PromptSelect("Choose:", []string{})
	if result != "" {
		t.Errorf("expected empty string for empty options, got %q", result)
	}
}

func TestPromptSelect_SingleOption(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))
	prompts := NewPrompts(reader)

	options := []string{"Only Option"}
	result := prompts.PromptSelect("Choose:", options)
	if result != "Only Option" {
		t.Errorf("expected 'Only Option', got %q", result)
	}
}
