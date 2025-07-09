package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the name of the configuration file
const ConfigFileName = ".sqlppconfig"

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

// configDirFunc is a function that returns the directories to look for config files
// It can be overridden for testing purposes
var configDirFunc = getConfigSearchDirs

// getConfigSearchDirs returns a list of directories to search for config files, in order of preference
// This ensures compatibility with both direct execution and MCP client child process usage
func getConfigSearchDirs() ([]string, error) {
	var dirs []string

	// 1. Current working directory (for MCP client child process usage)
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	// 2. Executable directory (original behavior for direct execution)
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		dirs = append(dirs, execDir)
	}

	// 3. User home directory (fallback)
	if homeDir, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, homeDir)
	}

	if len(dirs) == 0 {
		return nil, fmt.Errorf("could not determine any config search directories")
	}

	return dirs, nil
}

// LoadConfig loads configuration from ConfigFileName file, searching multiple directories
// Searches in order: current working directory, executable directory, user home directory
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Get the directories to search for config files
	configDirs, err := configDirFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to get config search directories: %w", err)
	}

	// Search for config file in each directory
	var configPath string
	var configFound bool

	for _, dir := range configDirs {
		candidatePath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(candidatePath); err == nil {
			configPath = candidatePath
			configFound = true
			break
		}
	}

	if !configFound {
		// No config file exists in any search directory, return default config
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate basic configuration only (not connections)
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to ConfigFileName file in the first available directory
func (c *Config) SaveConfig() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Get the directories to save config files
	configDirs, err := configDirFunc()
	if err != nil {
		return fmt.Errorf("failed to get config directories: %w", err)
	}

	// Use the first directory (current working directory if available)
	configDir := configDirs[0]
	configPath := filepath.Join(configDir, ConfigFileName)

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := c.ValidateBasic(); err != nil {
		return err
	}
	return c.ValidateConnections()
}

// ValidateBasic validates basic configuration that doesn't require database connections
func (c *Config) ValidateBasic() error {
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

	return nil
}

// ValidateConnections validates database connection configuration
func (c *Config) ValidateConnections() error {
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

// GetConfigPath returns the path to the configuration file, searching multiple directories
// Returns the path of the first existing config file, or the path in the first search directory if none exist
func GetConfigPath() string {
	configDirs, err := configDirFunc()
	if err != nil {
		// Fallback to current directory if we can't determine config directories
		return filepath.Join(".", ConfigFileName)
	}

	// Search for existing config file
	for _, dir := range configDirs {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// If no existing config file found, return path in first directory
	return filepath.Join(configDirs[0], ConfigFileName)
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

// RequiresDatabaseConnection checks if the given input requires a database connection
func RequiresDatabaseConnection(input string) bool {
	// Commands that don't require database connections
	connectionlessCommands := []string{
		"@drivers",
	}

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		// Check if this line is a connectionless command
		isConnectionless := false
		for _, cmd := range connectionlessCommands {
			if strings.HasPrefix(trimmedLine, cmd) {
				isConnectionless = true
				break
			}
		}

		// If we find any line that requires a connection, return true
		if !isConnectionless {
			return true
		}
	}

	// All non-empty, non-comment lines are connectionless commands
	return false
}
