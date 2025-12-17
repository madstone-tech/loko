package entities

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with field",
			err: &ValidationError{
				Entity:  "System",
				Field:   "Name",
				Value:   "test",
				Message: "invalid name",
			},
			expected: "System.Name: invalid name",
		},
		{
			name: "without field",
			err: &ValidationError{
				Entity:  "Project",
				Message: "validation failed",
			},
			expected: "Project: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &ValidationError{
		Entity:  "Test",
		Message: "test error",
		Err:     underlying,
	}

	if !errors.Is(err, underlying) {
		t.Error("Unwrap() should return underlying error")
	}
}

func TestNewValidationError_TruncatesLongValue(t *testing.T) {
	longValue := "this is a very long value that should be truncated because it exceeds fifty characters"
	err := NewValidationError("Test", "Field", longValue, "too long", nil)

	if len(err.Value) > 50 {
		t.Errorf("Value should be truncated, got length %d", len(err.Value))
	}
	if err.Value[len(err.Value)-3:] != "..." {
		t.Error("Truncated value should end with ...")
	}
}

func TestValidationErrors(t *testing.T) {
	var errs ValidationErrors

	if errs.HasErrors() {
		t.Error("Empty ValidationErrors should not have errors")
	}

	errs.Add("System", "Name", "", "name required", ErrEmptyName)
	errs.Add("System", "ID", "bad id!", "invalid id", ErrInvalidName)

	if !errs.HasErrors() {
		t.Error("ValidationErrors should have errors after Add")
	}

	if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	}

	errStr := errs.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestValidationErrors_SingleError(t *testing.T) {
	var errs ValidationErrors
	errs.Add("Test", "Field", "value", "single error", nil)

	// Single error should not say "X validation errors:"
	errStr := errs.Error()
	if errStr != "Test.Field: single error" {
		t.Errorf("Single error format unexpected: %s", errStr)
	}
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFoundError
		expected string
	}{
		{
			name:     "without parent",
			err:      &NotFoundError{Entity: "System", ID: "payment"},
			expected: "System 'payment' not found",
		},
		{
			name:     "with parent",
			err:      &NotFoundError{Entity: "Container", ID: "api", Parent: "PaymentSystem"},
			expected: "Container 'api' not found in PaymentSystem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDuplicateError(t *testing.T) {
	tests := []struct {
		name     string
		err      *DuplicateError
		expected string
	}{
		{
			name:     "without parent",
			err:      &DuplicateError{Entity: "System", ID: "payment"},
			expected: "System 'payment' already exists",
		},
		{
			name:     "with parent",
			err:      &DuplicateError{Entity: "Container", ID: "api", Parent: "PaymentSystem"},
			expected: "Container 'api' already exists in PaymentSystem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}
