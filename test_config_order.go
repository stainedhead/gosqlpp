package main

import (
	"fmt"
	"gosqlpp/internal/config"
)

func main() {
	// Load the current config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Save it back to see the new field ordering
	err = cfg.SaveConfig()
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Println("Configuration saved successfully with new field ordering")
}
