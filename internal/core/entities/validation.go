package entities

import (
	"regexp"
	"strings"
)

// Validation patterns
var (
	// namePattern allows alphanumeric, hyphens, underscores, and spaces
	// Must start with a letter or number
	namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\- ]*$`)
	
	// idPattern is stricter: lowercase alphanumeric and hyphens only
	// Used for file/directory names
	idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]*$`)
)

// ValidateName checks if a name is valid for display purposes.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if !namePattern.MatchString(name) {
		return ErrInvalidName
	}
	return nil
}

// ValidateID checks if an ID is valid for file/directory names.
func ValidateID(id string) error {
	if id == "" {
		return ErrEmptyID
	}
	if !idPattern.MatchString(id) {
		return ErrInvalidName
	}
	return nil
}

// NormalizeName converts a display name to a valid ID.
// "Payment Service" -> "payment-service"
func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	
	// Remove consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	
	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")
	
	return name
}

// ValidatePath checks if a path is valid (non-empty, no traversal).
func ValidatePath(path string) error {
	if path == "" {
		return ErrEmptyPath
	}
	// Basic path traversal protection
	if strings.Contains(path, "..") {
		return ErrInvalidName
	}
	return nil
}
