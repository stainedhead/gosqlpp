package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.DefaultConnection != "" {
		t.Errorf("Expected empty default connection, got %s", config.DefaultConnection)
	}
	
	if config.EndOnError != false {
		t.Errorf("Expected EndOnError to be false, got %t", config.EndOnError)
	}
	
	if config.Output != "table" {
		t.Errorf("Expected output to be 'table', got %s", config.Output)
	}
	
	if config.Connections == nil {
		t.Error("Expected connections map to be initialized")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with single connection",
			config: &Config{
				DefaultConnection: "test",
				EndOnError:        false,
				Output:            "table",
				Connections: map[string]Connection{
					"test": {
						Driver:           "sqlite3",
						ConnectionString: "test.db",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with auto-default connection",
			config: &Config{
				DefaultConnection: "",
				EndOnError:        false,
				Output:            "json",
				Connections: map[string]Connection{
					"test": {
						Driver:           "postgres",
						ConnectionString: "postgres://localhost/test",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid output format",
			config: &Config{
				DefaultConnection: "test",
				EndOnError:        false,
				Output:            "invalid",
				Connections: map[string]Connection{
					"test": {
						Driver:           "sqlite3",
						ConnectionString: "test.db",
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid output format",
		},
		{
			name: "no connections",
			config: &Config{
				DefaultConnection: "",
				EndOnError:        false,
				Output:            "table",
				Connections:       map[string]Connection{},
			},
			expectError: true,
			errorMsg:    "no database connections defined",
		},
		{
			name: "missing driver",
			config: &Config{
				DefaultConnection: "test",
				EndOnError:        false,
				Output:            "table",
				Connections: map[string]Connection{
					"test": {
						Driver:           "",
						ConnectionString: "test.db",
					},
				},
			},
			expectError: true,
			errorMsg:    "missing driver",
		},
		{
			name: "missing connection string",
			config: &Config{
				DefaultConnection: "test",
				EndOnError:        false,
				Output:            "table",
				Connections: map[string]Connection{
					"test": {
						Driver:           "sqlite3",
						ConnectionString: "",
					},
				},
			},
			expectError: true,
			errorMsg:    "missing connection-string",
		},
		{
			name: "default connection not found",
			config: &Config{
				DefaultConnection: "missing",
				EndOnError:        false,
				Output:            "table",
				Connections: map[string]Connection{
					"test": {
						Driver:           "sqlite3",
						ConnectionString: "test.db",
					},
				},
			},
			expectError: true,
			errorMsg:    "not found in connections",
		},
		{
			name: "multiple connections without default",
			config: &Config{
				DefaultConnection: "",
				EndOnError:        false,
				Output:            "table",
				Connections: map[string]Connection{
					"test1": {
						Driver:           "sqlite3",
						ConnectionString: "test1.db",
					},
					"test2": {
						Driver:           "sqlite3",
						ConnectionString: "test2.db",
					},
				},
			},
			expectError: true,
			errorMsg:    "no default-connection specified",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && len(tt.errorMsg) > 0 {
					if len(err.Error()) == 0 || len(tt.errorMsg) == 0 {
						t.Errorf("Error message check failed: err='%s', expected='%s'", err.Error(), tt.errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestGetConnection(t *testing.T) {
	config := &Config{
		DefaultConnection: "default",
		Connections: map[string]Connection{
			"default": {
				Driver:           "sqlite3",
				ConnectionString: "default.db",
			},
			"test": {
				Driver:           "postgres",
				ConnectionString: "postgres://localhost/test",
			},
		},
	}
	
	// Test getting default connection
	conn, err := config.GetConnection("")
	if err != nil {
		t.Errorf("Expected no error getting default connection, got %v", err)
	}
	if conn.Driver != "sqlite3" {
		t.Errorf("Expected driver 'sqlite3', got %s", conn.Driver)
	}
	
	// Test getting named connection
	conn, err = config.GetConnection("test")
	if err != nil {
		t.Errorf("Expected no error getting named connection, got %v", err)
	}
	if conn.Driver != "postgres" {
		t.Errorf("Expected driver 'postgres', got %s", conn.Driver)
	}
	
	// Test getting non-existent connection
	_, err = config.GetConnection("missing")
	if err == nil {
		t.Error("Expected error getting non-existent connection")
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	// Change to a temporary directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)
	
	config, err := LoadConfig()
	if err != nil {
		t.Errorf("Expected no error when config file doesn't exist, got %v", err)
	}
	
	if config.Output != "table" {
		t.Errorf("Expected default output 'table', got %s", config.Output)
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Change to a temporary directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)
	
	// Create a test config
	originalConfig := &Config{
		DefaultConnection: "test",
		EndOnError:        true,
		Output:            "json",
		Connections: map[string]Connection{
			"test": {
				Driver:           "sqlite3",
				ConnectionString: "test.db",
			},
		},
	}
	
	// Save config
	err := originalConfig.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Load config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Compare configs
	if loadedConfig.DefaultConnection != originalConfig.DefaultConnection {
		t.Errorf("DefaultConnection mismatch: expected %s, got %s", 
			originalConfig.DefaultConnection, loadedConfig.DefaultConnection)
	}
	
	if loadedConfig.EndOnError != originalConfig.EndOnError {
		t.Errorf("EndOnError mismatch: expected %t, got %t", 
			originalConfig.EndOnError, loadedConfig.EndOnError)
	}
	
	if loadedConfig.Output != originalConfig.Output {
		t.Errorf("Output mismatch: expected %s, got %s", 
			originalConfig.Output, loadedConfig.Output)
	}
	
	testConn, exists := loadedConfig.Connections["test"]
	if !exists {
		t.Error("Test connection not found in loaded config")
	} else {
		if testConn.Driver != "sqlite3" {
			t.Errorf("Driver mismatch: expected sqlite3, got %s", testConn.Driver)
		}
		if testConn.ConnectionString != "test.db" {
			t.Errorf("ConnectionString mismatch: expected test.db, got %s", testConn.ConnectionString)
		}
	}
}

func TestGetConfigPath(t *testing.T) {
	expected := filepath.Join(".", ".sqlppconfig")
	actual := GetConfigPath()
	
	if actual != expected {
		t.Errorf("Expected config path %s, got %s", expected, actual)
	}
}
