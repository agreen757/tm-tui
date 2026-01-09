package main

import (
	"fmt"
	"log"
	"os"

	"github.com/agreen757/tm-tui/internal/cli"
	"github.com/agreen757/tm-tui/internal/config"
)

func main() {
	// Initialize Crush configuration before starting TUI
	// This creates .crush.json with defaults if missing (non-destructive)
	// Existing configs are never modified or overwritten
	if _, err := config.InitCrushConfig(); err != nil {
		log.Printf("Warning: Failed to initialize Crush config: %v\n", err)
		// Continue execution - this is not fatal
	}

	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
