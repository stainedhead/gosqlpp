package preprocessor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConditionalStack(t *testing.T) {
	stack := NewConditionalStack()
	
	if !stack.IsEmpty() {
		t.Error("Expected stack to be empty initially")
	}
	
	if stack.Depth() != 0 {
		t.Error("Expected depth to be 0 initially")
	}
	
	if !stack.ShouldInclude() {
		t.Error("Expected ShouldInclude to be true when stack is empty")
	}
	
	// Push an active block
	block1 := ConditionalBlock{
		Type:      "ifdef",
		Variable:  "TEST",
		StartLine: 1,
		Active:    true,
	}
	stack.Push(block1)
	
	if stack.IsEmpty() {
		t.Error("Expected stack not to be empty after push")
	}
	
	if stack.Depth() != 1 {
		t.Error("Expected depth to be 1 after push")
	}
	
	if !stack.ShouldInclude() {
		t.Error("Expected ShouldInclude to be true with active block")
	}
	
	// Push an inactive block
	block2 := ConditionalBlock{
		Type:      "ifndef",
		Variable:  "OTHER",
		StartLine: 5,
		Active:    false,
	}
	stack.Push(block2)
	
	if stack.Depth() != 2 {
		t.Error("Expected depth to be 2 after second push")
	}
	
	if stack.ShouldInclude() {
		t.Error("Expected ShouldInclude to be false with inactive block")
	}
	
	// Pop blocks
	poppedBlock, err := stack.Pop()
	if err != nil {
		t.Errorf("Unexpected error popping: %v", err)
	}
	
	if poppedBlock.Variable != "OTHER" {
		t.Errorf("Expected popped block variable 'OTHER', got %s", poppedBlock.Variable)
	}
	
	if !stack.ShouldInclude() {
		t.Error("Expected ShouldInclude to be true after popping inactive block")
	}
	
	// Pop last block
	_, err = stack.Pop()
	if err != nil {
		t.Errorf("Unexpected error popping: %v", err)
	}
	
	if !stack.IsEmpty() {
		t.Error("Expected stack to be empty after popping all blocks")
	}
	
	// Try to pop from empty stack
	_, err = stack.Pop()
	if err == nil {
		t.Error("Expected error when popping from empty stack")
	}
}

func TestProcessIfdef(t *testing.T) {
	p := NewPreprocessor()
	p.SetDefine("DEFINED_VAR", "value")
	
	tests := []struct {
		line        string
		expectError bool
		expectActive bool
	}{
		{
			line:         "#ifdef DEFINED_VAR",
			expectError:  false,
			expectActive: true,
		},
		{
			line:         "#ifdef UNDEFINED_VAR",
			expectError:  false,
			expectActive: false,
		},
		{
			line:         "#ifdef DEFINED_VAR // with comment",
			expectError:  false,
			expectActive: true,
		},
		{
			line:        "#ifdef",
			expectError: true,
		},
		{
			line:        "#ifdef INVALID SYNTAX",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			// Reset preprocessor state
			p.conditionalStack = NewConditionalStack()
			
			lines, locations, err := p.processIfdef(tt.line, "test.sql", 1)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			// #ifdef lines should not produce output
			if len(lines) != 0 {
				t.Errorf("Expected no output lines, got %d", len(lines))
			}
			
			if len(locations) != 0 {
				t.Errorf("Expected no locations, got %d", len(locations))
			}
			
			// Check if conditional block was created correctly
			if p.conditionalStack.IsEmpty() {
				t.Error("Expected conditional block to be pushed")
			} else {
				shouldInclude := p.conditionalStack.ShouldInclude()
				if shouldInclude != tt.expectActive {
					t.Errorf("Expected active=%t, got %t", tt.expectActive, shouldInclude)
				}
			}
		})
	}
}

func TestProcessIfndef(t *testing.T) {
	p := NewPreprocessor()
	p.SetDefine("DEFINED_VAR", "value")
	
	tests := []struct {
		line         string
		expectError  bool
		expectActive bool
	}{
		{
			line:         "#ifndef DEFINED_VAR",
			expectError:  false,
			expectActive: false,
		},
		{
			line:         "#ifndef UNDEFINED_VAR",
			expectError:  false,
			expectActive: true,
		},
		{
			line:         "#ifndef UNDEFINED_VAR // with comment",
			expectError:  false,
			expectActive: true,
		},
		{
			line:        "#ifndef",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			// Reset preprocessor state
			p.conditionalStack = NewConditionalStack()
			
			lines, locations, err := p.processIfndef(tt.line, "test.sql", 1)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			// #ifndef lines should not produce output
			if len(lines) != 0 {
				t.Errorf("Expected no output lines, got %d", len(lines))
			}
			
			if len(locations) != 0 {
				t.Errorf("Expected no locations, got %d", len(locations))
			}
			
			// Check if conditional block was created correctly
			if p.conditionalStack.IsEmpty() {
				t.Error("Expected conditional block to be pushed")
			} else {
				shouldInclude := p.conditionalStack.ShouldInclude()
				if shouldInclude != tt.expectActive {
					t.Errorf("Expected active=%t, got %t", tt.expectActive, shouldInclude)
				}
			}
		})
	}
}

func TestConditionalProcessing(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	
	// Create test file with conditionals
	testFile := filepath.Join(tempDir, "conditional.sql")
	testContent := `#define DEBUG 1
SELECT 'before conditional' as status;
#ifdef DEBUG
SELECT 'debug mode enabled' as debug_status;
#end
#ifndef PRODUCTION
SELECT 'not in production' as env_status;
#end
#ifdef UNDEFINED_VAR
SELECT 'this should not appear' as hidden;
#end
SELECT 'after conditionals' as status;`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Process the file
	p := NewPreprocessor()
	lines, _, err := p.ProcessFile(testFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}
	
	// Check results
	expectedLines := []string{
		"SELECT 'before conditional' as status;",
		"SELECT 'debug mode enabled' as debug_status;",
		"SELECT 'not in production' as env_status;",
		"SELECT 'after conditionals' as status;",
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

func TestNestedConditionals(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	
	// Create test file with nested conditionals
	testFile := filepath.Join(tempDir, "nested.sql")
	testContent := `#define OUTER_VAR 1
#define INNER_VAR 1
SELECT 'start' as status;
#ifdef OUTER_VAR
SELECT 'outer block' as status;
#ifdef INNER_VAR
SELECT 'inner block' as status;
#end
SELECT 'after inner' as status;
#end
SELECT 'end' as status;`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Process the file
	p := NewPreprocessor()
	lines, _, err := p.ProcessFile(testFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}
	
	// Check results
	expectedLines := []string{
		"SELECT 'start' as status;",
		"SELECT 'outer block' as status;",
		"SELECT 'inner block' as status;",
		"SELECT 'after inner' as status;",
		"SELECT 'end' as status;",
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

func TestUnclosedConditional(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	
	// Create test file with unclosed conditional
	testFile := filepath.Join(tempDir, "unclosed.sql")
	testContent := `#define TEST_VAR 1
SELECT 'start' as status;
#ifdef TEST_VAR
SELECT 'in conditional' as status;
-- Missing #end`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Process the file - should fail
	p := NewPreprocessor()
	_, _, err = p.ProcessFile(testFile)
	if err == nil {
		t.Error("Expected error for unclosed conditional block")
	}
}

func TestEndWithoutIfdef(t *testing.T) {
	p := NewPreprocessor()
	
	lines, locations, err := p.processEnd("#end", "test.sql", 1)
	if err == nil {
		t.Error("Expected error for #end without matching #ifdef")
	}
	
	if len(lines) != 0 {
		t.Errorf("Expected no output lines, got %d", len(lines))
	}
	
	if len(locations) != 0 {
		t.Errorf("Expected no locations, got %d", len(locations))
	}
}
