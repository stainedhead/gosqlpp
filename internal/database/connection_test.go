package database

import (
	"testing"

	"gosqlpp/internal/config"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Error("Expected manager to be created")
	}
	
	if manager.connections == nil {
		t.Error("Expected connections map to be initialized")
	}
}

func TestGetSupportedDrivers(t *testing.T) {
	drivers := GetSupportedDrivers()
	expectedDrivers := []string{"postgres", "mysql", "sqlite3", "sqlserver"}
	
	if len(drivers) != len(expectedDrivers) {
		t.Errorf("Expected %d drivers, got %d", len(expectedDrivers), len(drivers))
	}
	
	for _, expected := range expectedDrivers {
		found := false
		for _, driver := range drivers {
			if driver == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected driver %s not found", expected)
		}
	}
}

func TestIsDriverSupported(t *testing.T) {
	tests := []struct {
		driver   string
		expected bool
	}{
		{"sqlite3", true},
		{"postgres", true},
		{"mysql", true},
		{"sqlserver", true},
		{"unsupported", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := isDriverSupported(tt.driver)
			if result != tt.expected {
				t.Errorf("isDriverSupported(%s) = %v, expected %v", tt.driver, result, tt.expected)
			}
		})
	}
}

func TestConnectUnsupportedDriver(t *testing.T) {
	manager := NewManager()
	conn := config.Connection{
		Driver:           "unsupported",
		ConnectionString: "test",
	}
	
	err := manager.Connect("test", conn)
	if err == nil {
		t.Error("Expected error for unsupported driver")
	}
}

func TestConnectSQLite(t *testing.T) {
	manager := NewManager()
	defer manager.CloseAll()
	
	conn := config.Connection{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	}
	
	err := manager.Connect("test", conn)
	if err != nil {
		t.Errorf("Expected no error connecting to SQLite, got %v", err)
	}
	
	// Test getting the connection
	dbConn, err := manager.GetConnection("test")
	if err != nil {
		t.Errorf("Expected no error getting connection, got %v", err)
	}
	
	if dbConn.Driver != "sqlite3" {
		t.Errorf("Expected driver 'sqlite3', got %s", dbConn.Driver)
	}
	
	if dbConn.Name != "test" {
		t.Errorf("Expected name 'test', got %s", dbConn.Name)
	}
}

func TestListConnections(t *testing.T) {
	manager := NewManager()
	defer manager.CloseAll()
	
	// Initially should be empty
	connections := manager.ListConnections()
	if len(connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(connections))
	}
	
	// Add a connection
	conn := config.Connection{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	}
	
	err := manager.Connect("test", conn)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	
	connections = manager.ListConnections()
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}
	
	if connections[0] != "test" {
		t.Errorf("Expected connection name 'test', got %s", connections[0])
	}
}
