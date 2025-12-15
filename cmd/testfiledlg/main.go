package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

func main() {
	// Get the docs directory path
	docsDir := filepath.Join(".taskmaster", "docs")
	
	// Check if the directory exists
	if _, err := os.Stat(docsDir); err != nil {
		fmt.Printf("Error with docs directory: %v\n", err)
		return
	}
	
	fmt.Printf("Testing with directory: %s\n", docsDir)
	
	// Check what files are in the directory directly
	entries, err := os.ReadDir(docsDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}
	
	fmt.Printf("Files in directory (direct ReadDir):\n")
	for _, entry := range entries {
		fmt.Printf("  %s (isDir: %v)\n", entry.Name(), entry.IsDir())
	}
	
	// Now use the dialog's function
	fmt.Printf("\nTesting file selection dialog with extensions [.md, .txt]\n")
	filters := make(map[string]struct{})
	for _, ext := range []string{".md", ".txt"} {
		filters[ext] = struct{}{}
	}
	
	fmt.Printf("Filter map: %v\n", filters)
	
	results, err := dialog.TestReadDirectoryEntries(docsDir, filters)
	if err != nil {
		fmt.Printf("Error from dialog's function: %v\n", err)
		return
	}
	
	fmt.Printf("\nFiles found by dialog's function:\n")
	for _, entry := range results {
		fmt.Printf("  %s (isDir: %v)\n", entry.Name, entry.IsDir)
	}
}