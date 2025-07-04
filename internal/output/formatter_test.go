package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gosqlpp/internal/database"
)

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()
	expected := []string{"table", "json", "yaml", "csv"}
	
	if len(formats) != len(expected) {
		t.Errorf("Expected %d formats, got %d", len(expected), len(formats))
	}
	
	for _, expectedFormat := range expected {
		found := false
		for _, format := range formats {
			if format == expectedFormat {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found", expectedFormat)
		}
	}
}

func TestIsFormatSupported(t *testing.T) {
	tests := []struct {
		format   string
		expected bool
	}{
		{"table", true},
		{"json", true},
		{"yaml", true},
		{"csv", true},
		{"TABLE", true}, // case insensitive
		{"JSON", true},
		{"unsupported", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsFormatSupported(tt.format)
			if result != tt.expected {
				t.Errorf("IsFormatSupported(%s) = %v, expected %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "NULL"},
		{"hello", "hello"},
		{[]byte("bytes"), "bytes"},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
	}
	
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("json", &buf)
	
	result := &database.ExecutionResult{
		Columns: []string{"id", "name"},
		Rows: [][]interface{}{
			{1, "John"},
			{2, "Jane"},
		},
	}
	
	err := formatter.FormatResult(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "John") {
		t.Error("Expected output to contain 'John'")
	}
	if !strings.Contains(output, "Jane") {
		t.Error("Expected output to contain 'Jane'")
	}
}

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("csv", &buf)
	
	result := &database.ExecutionResult{
		Columns: []string{"id", "name"},
		Rows: [][]interface{}{
			{1, "John"},
			{2, "Jane"},
		},
	}
	
	err := formatter.FormatResult(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	
	if lines[0] != "id,name" {
		t.Errorf("Expected header 'id,name', got %s", lines[0])
	}
}

func TestFormatError(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("table", &buf)
	
	result := &database.ExecutionResult{
		Error:      fmt.Errorf("test error"),
		FileName:   "test.sql",
		LineNumber: 10,
	}
	
	err := formatter.FormatResult(result)
	if err != nil {
		t.Errorf("Expected no error formatting error result, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "test.sql:10: error: test error") {
		t.Errorf("Expected error format, got: %s", output)
	}
}
