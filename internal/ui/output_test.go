package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestOutput_Success(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	out.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Error("Expected success checkmark")
	}
	if !strings.Contains(output, "Operation completed") {
		t.Error("Expected message in output")
	}
}

func TestOutput_Error(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithErrWriter(&buf)

	out.Error("Something went wrong")

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Error("Expected error X mark")
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Error("Expected message in output")
	}
}

func TestOutput_Warning(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithErrWriter(&buf)

	out.Warning("This is a warning")

	output := buf.String()
	if !strings.Contains(output, "⚠") {
		t.Error("Expected warning symbol")
	}
}

func TestOutput_Progress(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	out.Progress(50, 100, "Processing files")

	output := buf.String()
	if !strings.Contains(output, "50%") {
		t.Error("Expected percentage in output")
	}
	if !strings.Contains(output, "Processing files") {
		t.Error("Expected message in output")
	}
}

func TestOutput_Progress_ZeroTotal(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	out.Progress(0, 0, "No progress")

	output := buf.String()
	if !strings.Contains(output, "No progress") {
		t.Error("Expected message in output")
	}
}

func TestOutput_Debug_Verbose(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf).WithVerbose(true)

	out.Debug("Debug message")

	output := buf.String()
	if !strings.Contains(output, "Debug message") {
		t.Error("Expected debug message when verbose")
	}
}

func TestOutput_Debug_NotVerbose(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf).WithVerbose(false)

	out.Debug("Debug message")

	output := buf.String()
	if strings.Contains(output, "Debug message") {
		t.Error("Debug message should not appear when not verbose")
	}
}

func TestOutput_List(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	out.List([]string{"Item 1", "Item 2", "Item 3"})

	output := buf.String()
	if !strings.Contains(output, "• Item 1") {
		t.Error("Expected bullet point for Item 1")
	}
	if !strings.Contains(output, "• Item 2") {
		t.Error("Expected bullet point for Item 2")
	}
	if !strings.Contains(output, "• Item 3") {
		t.Error("Expected bullet point for Item 3")
	}
}

func TestOutput_Table(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"System A", "Active"},
		{"System B", "Inactive"},
	}

	out.Table(headers, rows)

	output := buf.String()
	if !strings.Contains(output, "Name") {
		t.Error("Expected header Name")
	}
	if !strings.Contains(output, "Status") {
		t.Error("Expected header Status")
	}
	if !strings.Contains(output, "System A") {
		t.Error("Expected System A in output")
	}
	if !strings.Contains(output, "Active") {
		t.Error("Expected Active in output")
	}
}

func TestOutput_KeyValue(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithWriter(&buf)

	out.KeyValue("Version", "1.0.0")

	output := buf.String()
	if !strings.Contains(output, "Version") {
		t.Error("Expected key in output")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("Expected value in output")
	}
}

func TestFormatError(t *testing.T) {
	err := errors.New("test error")
	formatted := FormatError(err)

	if !strings.Contains(formatted, "test error") {
		t.Error("Expected error message in formatted output")
	}
}

func TestFormatError_Nil(t *testing.T) {
	formatted := FormatError(nil)
	if formatted != "" {
		t.Error("Expected empty string for nil error")
	}
}

func TestFormatSuccess(t *testing.T) {
	formatted := FormatSuccess("Done")

	if !strings.Contains(formatted, "✓") {
		t.Error("Expected checkmark in formatted output")
	}
	if !strings.Contains(formatted, "Done") {
		t.Error("Expected message in formatted output")
	}
}

func TestFormatWarning(t *testing.T) {
	formatted := FormatWarning("Caution")

	if !strings.Contains(formatted, "⚠") {
		t.Error("Expected warning symbol in formatted output")
	}
	if !strings.Contains(formatted, "Caution") {
		t.Error("Expected message in formatted output")
	}
}

func TestOutput_ErrorWithDetails(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutput().WithErrWriter(&buf)

	out.ErrorWithDetails("Operation failed", "Check your permissions")

	output := buf.String()
	if !strings.Contains(output, "Operation failed") {
		t.Error("Expected error message")
	}
	if !strings.Contains(output, "Check your permissions") {
		t.Error("Expected details in output")
	}
}
