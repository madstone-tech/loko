package entities

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "PaymentService", false},
		{"valid with spaces", "Payment Service", false},
		{"valid with hyphens", "payment-service", false},
		{"valid with underscores", "payment_service", false},
		{"valid with numbers", "Service2", false},
		{"valid starts with number", "3DRenderer", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"special chars", "Payment@Service", true},
		{"starts with hyphen", "-payment", true},
		{"starts with underscore", "_payment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "payment", false},
		{"valid with hyphens", "payment-service", false},
		{"valid with numbers", "service2", false},
		{"valid starts with number", "3drenderer", false},
		{"empty", "", true},
		{"uppercase", "Payment", true},
		{"spaces", "payment service", true},
		{"underscores", "payment_service", true},
		{"starts with hyphen", "-payment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "payment", "payment"},
		{"uppercase", "Payment", "payment"},
		{"spaces to hyphens", "Payment Service", "payment-service"},
		{"underscores to hyphens", "payment_service", "payment-service"},
		{"multiple spaces", "Payment  Service", "payment-service"},
		{"leading/trailing spaces", "  Payment  ", "payment"},
		{"mixed", "My_Cool Service", "my-cool-service"},
		{"consecutive hyphens", "payment--service", "payment-service"},
		{"leading hyphen after normalize", "  _payment", "payment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeName(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid absolute", "/home/user/project", false},
		{"valid relative", "./src/system", false},
		{"valid simple", "system.d2", false},
		{"empty", "", true},
		{"path traversal", "../../../etc/passwd", true},
		{"path traversal middle", "/home/../../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
