// Package cli provides command-line interface utilities.
package cli

import (
	"bufio"
	"fmt"
	"strings"
)

// Prompts provides interactive CLI prompts for gathering user input.
type Prompts struct {
	reader *bufio.Reader
}

// NewPrompts creates a new Prompts instance reading from stdin.
func NewPrompts(reader *bufio.Reader) *Prompts {
	return &Prompts{reader: reader}
}

// PromptString asks the user for a string input with optional default value.
func (p *Prompts) PromptString(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// PromptStringMulti asks the user for multiple comma-separated values.
// Returns a slice of trimmed strings.
func (p *Prompts) PromptStringMulti(prompt string) []string {
	fmt.Printf("%s (comma-separated): ", prompt)
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return []string{}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// PromptYesNo asks the user for a yes/no response.
func (p *Prompts) PromptYesNo(prompt string, defaultYes bool) bool {
	defaultStr := "n"
	if defaultYes {
		defaultStr = "y"
	}

	fmt.Printf("%s [%s/n]: ", prompt, defaultStr)
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

// PromptSelect asks the user to select from options.
// Returns the selected option or empty string if cancelled.
func (p *Prompts) PromptSelect(prompt string, options []string) string {
	if len(options) == 0 {
		return ""
	}

	if len(options) == 1 {
		return options[0]
	}

	fmt.Printf("%s\n", prompt)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Printf("Select (1-%d): ", len(options))

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return ""
	}

	input = strings.TrimSpace(input)
	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(options) {
		return ""
	}

	return options[idx-1]
}
