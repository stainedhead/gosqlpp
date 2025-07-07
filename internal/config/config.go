package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Connection represents a database connection configuration
type Connection struct {
	Driver           string `yaml:"driver"`
	Notes            string `yaml:"notes,omitempty"`
	ConnectionString string `yaml:"connection-string"`
}

// Config represents the application configuration
type Config struct {
	DefaultConnection string                `yaml:"default-connection"`
	EndOnError        bool                  `yaml:"end-on-error"`
	Output            string                `yaml:"output"`
	Connections       map[string]Connection `yaml:"connections"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DefaultConnection: "",
		EndOnError:        false,
		Output:            "table",
		Connections:       make(map[string]Connection),
	}
}

// configDirFunc is a function that returns the directory to look for config files
// It can be overridden for testing purposes
var configDirFunc = getExecutableDir

// getExecutableDir returns the directory where the executable is located
func getExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

// LoadConfig loads configuration from .sqlppconfig file in the executable's directory
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Get the directory to look for config files
	configDir, err := configDirFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ".sqlppconfig")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file exists, return default config
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to .sqlppconfig file in the executable's directory
func (c *Config) SaveConfig() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Get the directory to save config files
	configDir, err := configDirFunc()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ".sqlppconfig")

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate output format
	validOutputs := map[string]bool{
		"table": true,
		"json":  true,
		"yaml":  true,
		"csv":   true,
	}

	if !validOutputs[c.Output] {
		return fmt.Errorf("invalid output format '%s', must be one of: table, json, yaml, csv", c.Output)
	}

	// Validate connections
	if len(c.Connections) == 0 {
		return fmt.Errorf("no database connections defined")
	}

	for name, conn := range c.Connections {
		if conn.Driver == "" {
			return fmt.Errorf("connection '%s' missing driver", name)
		}
		if conn.ConnectionString == "" {
			return fmt.Errorf("connection '%s' missing connection-string", name)
		}
	}

	// Validate default connection
	if c.DefaultConnection != "" {
		if _, exists := c.Connections[c.DefaultConnection]; !exists {
			return fmt.Errorf("default-connection '%s' not found in connections", c.DefaultConnection)
		}
	} else if len(c.Connections) == 1 {
		// If only one connection and no default specified, use it as default
		for name := range c.Connections {
			c.DefaultConnection = name
			break
		}
	} else if len(c.Connections) > 1 {
		return fmt.Errorf("multiple connections defined but no default-connection specified")
	}

	return nil
}

// GetConnection returns the connection configuration for the given name
func (c *Config) GetConnection(name string) (Connection, error) {
	if name == "" {
		name = c.DefaultConnection
	}

	conn, exists := c.Connections[name]
	if !exists {
		return Connection{}, fmt.Errorf("connection '%s' not found", name)
	}

	return conn, nil
}

// GetConfigPath returns the path to the configuration file in the executable's directory
func GetConfigPath() string {
	configDir, err := configDirFunc()
	if err != nil {
		// Fallback to current directory if we can't determine config directory
		return filepath.Join(".", ".sqlppconfig")
	}

	return filepath.Join(configDir, ".sqlppconfig")
}

// ConnectionInfo represents connection information for display
type ConnectionInfo struct {
	Name      string `json:"name" yaml:"name"`
	Driver    string `json:"driver" yaml:"driver"`
	Notes     string `json:"notes" yaml:"notes"`
	IsDefault bool   `json:"is_default" yaml:"is_default"`
}

// GetConnectionNames returns a slice of all connection names
func (c *Config) GetConnectionNames() []string {
	names := make([]string, 0, len(c.Connections))
	for name := range c.Connections {
		names = append(names, name)
	}
	return names
}

// GetConnectionInfos returns detailed information about all connections
func (c *Config) GetConnectionInfos() []ConnectionInfo {
	infos := make([]ConnectionInfo, 0, len(c.Connections))
	for name, conn := range c.Connections {
		info := ConnectionInfo{
			Name:      name,
			Driver:    conn.Driver,
			Notes:     conn.Notes,
			IsDefault: name == c.DefaultConnection,
		}
		infos = append(infos, info)
	}
	return infos
}
