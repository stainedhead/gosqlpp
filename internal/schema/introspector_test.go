package schema

import (
	"bytes"
	"testing"

	"gosqlpp/internal/database"
	"gosqlpp/internal/output"
)

func TestIsSchemaCommand(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"@schema-tables", true},
		{"@schema-views", true},
		{"@schema-procedures", true},
		{"@schema-functions", true},
		{"@schema-all", true},
		{"@drivers", true},
		{"@drivers \"s\"", true},
		{"SELECT * FROM table", false},
		{"@unknown-command", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := IsSchemaCommand(tt.line)
			if result != tt.expected {
				t.Errorf("IsSchemaCommand(%s) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParseSchemaCommand(t *testing.T) {
	tests := []struct {
		line            string
		expectedCommand string
		expectedFilter  string
	}{
		{"@schema-tables", "@schema-tables", ""},
		{"@schema-tables \"user\"", "@schema-tables", "user"},
		{"@schema-tables 'test'", "@schema-tables", "test"},
		{"@drivers", "@drivers", ""},
		{"@drivers \"postgres\"", "@drivers", "postgres"},
		{"@schema-all", "@schema-all", ""},
		{"", "", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			command, filter := ParseSchemaCommand(tt.line)
			if command != tt.expectedCommand {
				t.Errorf("ParseSchemaCommand(%s) command = %s, expected %s", tt.line, command, tt.expectedCommand)
			}
			if filter != tt.expectedFilter {
				t.Errorf("ParseSchemaCommand(%s) filter = %s, expected %s", tt.line, filter, tt.expectedFilter)
			}
		})
	}
}

func TestProcessDriversCommand(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	formatter := output.NewFormatter("table", &buf)
	
	// Create introspector with mock connection
	// Note: We can't easily test this without a real connection, but we can test the command recognition
	conn := &database.Connection{
		Driver: "sqlite3",
		Name:   "test",
	}
	
	introspector := NewIntrospector(conn, formatter)
	
	// Test that the command is recognized
	err := introspector.ProcessSchemaCommand("@drivers", "")
	if err != nil {
		t.Errorf("ProcessSchemaCommand(@drivers) returned error: %v", err)
	}
	
	// Check that some output was generated
	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected some output from @drivers command")
	}
}

func TestFilterNames(t *testing.T) {
	names := []string{"postgres", "mysql", "sqlite3", "sqlserver"}
	
	tests := []struct {
		filter   string
		expected []string
	}{
		{"", []string{"postgres", "mysql", "sqlite3", "sqlserver"}},
		{"s", []string{"sqlite3", "sqlserver"}},
		{"p", []string{"postgres"}},
		{"m", []string{"mysql"}},
		{"xyz", []string{}},
	}
	
	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			result := filterNames(names, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("filterNames with filter '%s' returned %d items, expected %d", tt.filter, len(result), len(tt.expected))
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("filterNames with filter '%s' item %d = %s, expected %s", tt.filter, i, result[i], expected)
				}
			}
		})
	}
}
