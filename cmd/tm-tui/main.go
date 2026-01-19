package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/agreen757/tm-tui/internal/cli"
	"github.com/agreen757/tm-tui/internal/config"
)

func main() {
	// Setup debug logging to file instead of stdout to avoid interfering with TUI
	logDir := ".taskmaster/logs"
	if err := os.MkdirAll(logDir, 0755); err == nil {
		logFile, err := os.OpenFile(filepath.Join(logDir, "tui-debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.SetOutput(logFile)
			log.Printf("=== TUI Debug Session Started ===")
		}
	}

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
