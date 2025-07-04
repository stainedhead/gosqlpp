package database

import (
	"database/sql"
	"fmt"

	"gosqlpp/internal/config"
	
	// Import database drivers
	_ "github.com/lib/pq"                // PostgreSQL
	_ "github.com/go-sql-driver/mysql"   // MySQL/MariaDB
	_ "github.com/mattn/go-sqlite3"      // SQLite
	_ "github.com/microsoft/go-mssqldb"  // SQL Server
)

// Connection wraps a database connection with metadata
type Connection struct {
	DB     *sql.DB
	Driver string
	Name   string
}

// Manager handles database connections
type Manager struct {
	connections map[string]*Connection
}

// NewManager creates a new database connection manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
	}
}

// Connect establishes a database connection using the provided configuration
func (m *Manager) Connect(name string, conn config.Connection) error {
	// Check if connection already exists
	if _, exists := m.connections[name]; exists {
		return fmt.Errorf("connection '%s' already exists", name)
	}
	
	// Validate driver
	if !isDriverSupported(conn.Driver) {
		return fmt.Errorf("unsupported database driver: %s", conn.Driver)
	}
	
	// Open database connection
	db, err := sql.Open(conn.Driver, conn.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open database connection '%s': %w", name, err)
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database '%s': %w", name, err)
	}
	
	// Store the connection
	m.connections[name] = &Connection{
		DB:     db,
		Driver: conn.Driver,
		Name:   name,
	}
	
	return nil
}

// GetConnection returns a database connection by name
func (m *Manager) GetConnection(name string) (*Connection, error) {
	conn, exists := m.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}
	return conn, nil
}

// CloseConnection closes a specific database connection
func (m *Manager) CloseConnection(name string) error {
	conn, exists := m.connections[name]
	if !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}
	
	err := conn.DB.Close()
	delete(m.connections, name)
	return err
}

// CloseAll closes all database connections
func (m *Manager) CloseAll() error {
	var lastErr error
	for name, conn := range m.connections {
		if err := conn.DB.Close(); err != nil {
			lastErr = err
		}
		delete(m.connections, name)
	}
	return lastErr
}

// ListConnections returns a list of active connection names
func (m *Manager) ListConnections() []string {
	names := make([]string, 0, len(m.connections))
	for name := range m.connections {
		names = append(names, name)
	}
	return names
}

// isDriverSupported checks if a database driver is supported
func isDriverSupported(driver string) bool {
	supportedDrivers := map[string]bool{
		"postgres":  true,
		"mysql":     true,
		"sqlite3":   true,
		"sqlserver": true,
	}
	return supportedDrivers[driver]
}

// GetSupportedDrivers returns a list of supported database drivers
func GetSupportedDrivers() []string {
	return []string{"postgres", "mysql", "sqlite3", "sqlserver"}
}
