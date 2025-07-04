package preprocessor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPreprocessor(t *testing.T) {
	p := NewPreprocessor()
	if p == nil {
		t.Error("Expected preprocessor to be created")
	}
	
	if p.defines == nil {
		t.Error("Expected defines map to be initialized")
	}
	
	if len(p.defines) != 0 {
		t.Error("Expected defines map to be empty initially")
	}
}

func TestProcessDefine(t *testing.T) {
	p := NewPreprocessor()
	
	tests := []struct {
		line     string
		expected Define
		hasError bool
	}{
		{
			line:     `#define TRUE Y`,
			expected: Define{Name: "TRUE", Value: "Y"},
			hasError: false,
		},
		{
			line:     `#define MAX_LOOPS 100`,
			expected: Define{Name: "MAX_LOOPS", Value: "100"},
			hasError: false,
		},
		{
			line:     `#define PROC_NAME "stored-proc-name"`,
			expected: Define{Name: "PROC_NAME", Value: "stored-proc-name"},
			hasError: false,
		},
		{
			line:     `#define FOOBAR "some text to use" // some comment`,
			expected: Define{Name: "FOOBAR", Value: "some text to use"},
			hasError: false,
		},
		{
			line:     `#define SINGLE 'value'`,
			expected: Define{Name: "SINGLE", Value: "value"},
			hasError: false,
		},
		{
			line:     `#define INVALID`,
			hasError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			lines, locations, err := p.processDefine(tt.line, "test.sql", 1)
			
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			// #define lines should not produce output
			if len(lines) != 0 {
				t.Errorf("Expected no output lines, got %d", len(lines))
			}
			
			if len(locations) != 0 {
				t.Errorf("Expected no locations, got %d", len(locations))
			}
			
			// Check if define was stored
			if !p.HasDefine(tt.expected.Name) {
				t.Errorf("Expected define %s to be stored", tt.expected.Name)
			}
			
			define := p.defines[tt.expected.Name]
			if define.Value != tt.expected.Value {
				t.Errorf("Expected value %s, got %s", tt.expected.Value, define.Value)
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	p := NewPreprocessor()
	p.SetDefine("TRUE", "Y")
	p.SetDefine("MAX_LOOPS", "100")
	p.SetDefine("PROC_NAME", "stored-proc-name")
	
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "SELECT * FROM table WHERE active = TRUE",
			expected: "SELECT * FROM table WHERE active = Y",
		},
		{
			input:    "LIMIT MAX_LOOPS",
			expected: "LIMIT 100",
		},
		{
			input:    "EXEC PROC_NAME",
			expected: "EXEC stored-proc-name",
		},
		{
			input:    "No substitution needed",
			expected: "No substitution needed",
		},
		{
			input:    "TRUE and MAX_LOOPS in same line",
			expected: "Y and 100 in same line",
		},
		{
			input:    "TRUELY should not be replaced",
			expected: "TRUELY should not be replaced",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.substituteVariables(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestProcessInclude(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	
	// Create main file
	mainFile := filepath.Join(tempDir, "main.sql")
	mainContent := `SELECT 'main file' as source;
#include "included.sqi"
SELECT 'after include' as source;`
	
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}
	
	// Create included file
	includedFile := filepath.Join(tempDir, "included.sqi")
	includedContent := `SELECT 'included file' as source;
SELECT 'second line' as source;`
	
	err = os.WriteFile(includedFile, []byte(includedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create included file: %v", err)
	}
	
	// Process the main file
	p := NewPreprocessor()
	lines, locations, err := p.ProcessFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}
	
	// Check results
	expectedLines := []string{
		"SELECT 'main file' as source;",
		"SELECT 'included file' as source;",
		"SELECT 'second line' as source;",
		"SELECT 'after include' as source;",
	}
	
	if len(lines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}
	
	for i, expected := range expectedLines {
		if i < len(lines) && lines[i] != expected {
			t.Errorf("Line %d: expected %s, got %s", i, expected, lines[i])
		}
	}
	
	// Check locations
	if len(locations) != len(lines) {
		t.Errorf("Expected %d locations, got %d", len(lines), len(locations))
	}
	
	// First line should be from main file
	if locations[0].FileName != mainFile {
		t.Errorf("Expected first line from %s, got %s", mainFile, locations[0].FileName)
	}
	
	// Second line should be from included file
	if locations[1].FileName != includedFile {
		t.Errorf("Expected second line from %s, got %s", includedFile, locations[1].FileName)
	}
}

func TestDefineAndIncludeTogether(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	
	// Create main file with defines and include
	mainFile := filepath.Join(tempDir, "main.sql")
	mainContent := `#define TABLE_NAME users
#define LIMIT_COUNT 10
SELECT * FROM TABLE_NAME;
#include "query.sqi"`
	
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}
	
	// Create included file that uses defines
	includedFile := filepath.Join(tempDir, "query.sqi")
	includedContent := `SELECT COUNT(*) FROM TABLE_NAME LIMIT LIMIT_COUNT;`
	
	err = os.WriteFile(includedFile, []byte(includedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create included file: %v", err)
	}
	
	// Process the main file
	p := NewPreprocessor()
	lines, _, err := p.ProcessFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}
	
	// Check results
	expectedLines := []string{
		"SELECT * FROM users;",
		"SELECT COUNT(*) FROM users LIMIT 10;",
	}
	
	if len(lines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(lines))
		for i, line := range lines {
			t.Logf("Line %d: %s", i, line)
		}
	}
	
	for i, expected := range expectedLines {
		if i < len(lines) && lines[i] != expected {
			t.Errorf("Line %d: expected %s, got %s", i, expected, lines[i])
		}
	}
}

func TestGetDefines(t *testing.T) {
	p := NewPreprocessor()
	p.SetDefine("TEST1", "value1")
	p.SetDefine("TEST2", "value2")
	
	defines := p.GetDefines()
	
	if len(defines) != 2 {
		t.Errorf("Expected 2 defines, got %d", len(defines))
	}
	
	if defines["TEST1"].Value != "value1" {
		t.Errorf("Expected TEST1 = value1, got %s", defines["TEST1"].Value)
	}
	
	if defines["TEST2"].Value != "value2" {
		t.Errorf("Expected TEST2 = value2, got %s", defines["TEST2"].Value)
	}
	
	// Modify returned map should not affect original
	defines["TEST1"] = Define{Name: "TEST1", Value: "modified"}
	
	if p.defines["TEST1"].Value != "value1" {
		t.Error("Original defines should not be modified")
	}
}

func TestClearDefines(t *testing.T) {
	p := NewPreprocessor()
	p.SetDefine("TEST1", "value1")
	p.SetDefine("TEST2", "value2")
	
	if len(p.defines) != 2 {
		t.Error("Expected 2 defines before clear")
	}
	
	p.ClearDefines()
	
	if len(p.defines) != 0 {
		t.Error("Expected 0 defines after clear")
	}
}
