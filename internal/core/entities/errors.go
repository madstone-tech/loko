// Package entities contains the domain entities for loko.
// These are pure Go structs with validation logic and zero external dependencies.
package entities

import (
	"errors"
	"fmt"
	"strings"
)

// Common domain errors
var (
	ErrEmptyName        = errors.New("name cannot be empty")
	ErrInvalidName      = errors.New("name contains invalid characters")
	ErrEmptyID          = errors.New("id cannot be empty")
	ErrEmptyPath        = errors.New("path cannot be empty")
	ErrEmptySource      = errors.New("source cannot be empty")
	ErrDuplicateSystem  = errors.New("system already exists")
	ErrDuplicateContainer = errors.New("container already exists")
	ErrDuplicateComponent = errors.New("component already exists")
	ErrSystemNotFound   = errors.New("system not found")
	ErrContainerNotFound = errors.New("container not found")
	ErrComponentNotFound = errors.New("component not found")
	ErrInvalidHierarchy = errors.New("invalid C4 hierarchy")
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Entity  string // Entity type (e.g., "System", "Container")
	Field   string // Field that failed validation
	Value   string // The invalid value (may be truncated)
	Message string // Human-readable error message
	Err     error  // Underlying error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s.%s: %s", e.Entity, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Entity, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error.
func NewValidationError(entity, field, value, message string, err error) *ValidationError {
	// Truncate value if too long
	if len(value) > 50 {
		value = value[:47] + "..."
	}
	return &ValidationError{
		Entity:  entity,
		Field:   field,
		Value:   value,
		Message: message,
		Err:     err,
	}
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d validation errors:\n", len(ve)))
	for i, err := range ve {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return b.String()
}

// HasErrors returns true if there are validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add appends a validation error to the collection.
func (ve *ValidationErrors) Add(entity, field, value, message string, err error) {
	*ve = append(*ve, NewValidationError(entity, field, value, message, err))
}

// NotFoundError represents an entity not found error.
type NotFoundError struct {
	Entity string
	ID     string
	Parent string // Optional parent context
}

func (e *NotFoundError) Error() string {
	if e.Parent != "" {
		return fmt.Sprintf("%s '%s' not found in %s", e.Entity, e.ID, e.Parent)
	}
	return fmt.Sprintf("%s '%s' not found", e.Entity, e.ID)
}

// DuplicateError represents a duplicate entity error.
type DuplicateError struct {
	Entity string
	ID     string
	Parent string
}

func (e *DuplicateError) Error() string {
	if e.Parent != "" {
		return fmt.Sprintf("%s '%s' already exists in %s", e.Entity, e.ID, e.Parent)
	}
	return fmt.Sprintf("%s '%s' already exists", e.Entity, e.ID)
}
