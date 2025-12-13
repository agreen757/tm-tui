package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// The specific path that's causing problems
	docsPath := "/Users/adriangreen/Work/taskmaster-crush-fork/.taskmaster/docs"
	
	fmt.Printf("Testing access to: %s\n", docsPath)
	
	// Check if the path exists
	info, err := os.Stat(docsPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	if !info.IsDir() {
		fmt.Printf("Path exists but is not a directory\n")
		return
	}
	
	fmt.Printf("Path exists and is a directory\n")
	
	// Try to read the directory
	entries, err := os.ReadDir(docsPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}
	
	fmt.Printf("Directory contains %d entries:\n", len(entries))
	for _, entry := range entries {
		name := entry.Name()
		isDir := entry.IsDir()
		
		fullPath := filepath.Join(docsPath, name)
		size := int64(0)
		if !isDir {
			if info, err := os.Stat(fullPath); err == nil {
				size = info.Size()
			}
		}
		
		fmt.Printf("  - %s (isDir: %v, size: %d bytes)\n", name, isDir, size)
	}
}