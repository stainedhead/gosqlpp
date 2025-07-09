// Test program to demonstrate the new config search functionality
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gosqlpp/internal/config"
)

func main() {
	fmt.Println("Testing config search functionality...")
	fmt.Println("=====================================")

	// Get current working directory
	cwd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", cwd)

	// Get executable directory
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	fmt.Printf("Executable directory: %s\n", execDir)

	// Get user home directory
	homeDir, _ := os.UserHomeDir()
	fmt.Printf("User home directory: %s\n", homeDir)

	fmt.Println()

	// Test the config search
	fmt.Println("Config search order:")
	fmt.Println("1. Current working directory")
	fmt.Println("2. Executable directory")
	fmt.Println("3. User home directory")
	fmt.Println()

	// Load config and see what happens
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Config loaded successfully!\n")
	fmt.Printf("Number of connections: %d\n", len(cfg.Connections))
	fmt.Printf("Output format: %s\n", cfg.Output)

	// Show where config would be saved/read
	configPath := config.GetConfigPath()
	fmt.Printf("Config file path: %s\n", configPath)

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file exists at: %s\n", configPath)
	} else {
		fmt.Printf("No config file found (using defaults)\n")
	}
}
