package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileChange represents a changed file
type FileChange struct {
	Path        string    `json:"path"`
	ChangeType  string    `json:"changeType"`
	Description string    `json:"description"`
	LastChanged time.Time `json:"lastChanged"`
	CommitID    string    `json:"commitId"`
	IsPending   bool      `json:"isPending"`
}

// FileChangeMapping is the top-level storage
type FileChangeMapping struct {
	Version           string                   `json:"version"`
	LastUpdated       time.Time                `json:"lastUpdated"`
	Tasks             map[string][]FileChange  `json:"tasks"`
	UnassignedChanges []FileChange             `json:"unassignedChanges"`
}

func main() {
	logsDir := ".taskmaster/file-change-tracker"
	outputFile := ".taskmaster/file-changes.json"
	
	// Parse all log files
	mapping := &FileChangeMapping{
		Version:           "1.0",
		LastUpdated:       time.Now(),
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: []FileChange{},
	}
	
	// Read all .log files
	files, err := filepath.Glob(filepath.Join(logsDir, "*.log"))
	if err != nil {
		fmt.Printf("Error finding log files: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Found %d log files to process\n", len(files))
	
	for _, logFile := range files {
		// Extract task ID from filename (e.g., "1.1.log" -> "1.1")
		basename := filepath.Base(logFile)
		taskID := strings.TrimSuffix(basename, ".log")
		
		// Skip summary files
		if strings.Contains(strings.ToLower(taskID), "summary") {
			continue
		}
		
		fmt.Printf("Processing task %s...\n", taskID)
		
		// Parse the log file
		fileChanges, err := parseLogFile(logFile, taskID)
		if err != nil {
			fmt.Printf("  Warning: Failed to parse %s: %v\n", logFile, err)
			continue
		}
		
		if len(fileChanges) > 0 {
			mapping.Tasks[taskID] = fileChanges
			fmt.Printf("  Found %d file changes\n", len(fileChanges))
		}
	}
	
	// Write to output file
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("\nBackfill complete!\n")
	fmt.Printf("Total tasks processed: %d\n", len(mapping.Tasks))
	
	totalFiles := 0
	for _, changes := range mapping.Tasks {
		totalFiles += len(changes)
	}
	fmt.Printf("Total file changes: %d\n", totalFiles)
	fmt.Printf("Output written to: %s\n", outputFile)
}

func parseLogFile(logPath string, taskID string) ([]FileChange, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var fileChanges []FileChange
	seenFiles := make(map[string]bool)
	
	scanner := bufio.NewScanner(file)
	
	// Patterns to match file paths
	filePatterns := []*regexp.Regexp{
		// "### File: path/to/file.go"
		regexp.MustCompile(`^###\s+(?:File|Files?):\s+(.+?)(?:\s+\(NEW(?:\s+FILE)?\))?$`),
		// "**Added FileChange struct:**" followed by file info
		regexp.MustCompile(`^\*\*(?:Added|Created|Modified|Extended)\s+.+?:\*\*`),
		// Look for file paths in markdown code blocks or text
		regexp.MustCompile(`(?:^|\s)((?:internal|cmd|pkg)/[a-zA-Z0-9_/]+\.go)`),
	}
	
	inCodeChanges := false
	
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		// Track if we're in Code Changes section
		if strings.Contains(trimmed, "## Code Changes") {
			inCodeChanges = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") && !strings.Contains(trimmed, "Code Changes") {
			inCodeChanges = false
		}
		
		if !inCodeChanges {
			continue
		}
		
		// Try to extract file paths
		for _, pattern := range filePatterns {
			matches := pattern.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				filePath := strings.TrimSpace(matches[1])
				
				// Clean up the path
				filePath = strings.TrimPrefix(filePath, "**")
				filePath = strings.TrimSuffix(filePath, "**")
				filePath = strings.TrimPrefix(filePath, "### File: ")
				filePath = strings.TrimSuffix(filePath, " (NEW)")
				filePath = strings.TrimSuffix(filePath, " (NEW FILE)")
				filePath = strings.TrimSpace(filePath)
				
				// Skip if already seen or invalid
				if seenFiles[filePath] || filePath == "" || !strings.Contains(filePath, "/") {
					continue
				}
				
				// Determine change type
				changeType := "modified"
				if strings.Contains(line, "(NEW") || strings.Contains(line, "Created") || strings.Contains(line, "Added") {
					changeType = "added"
				}
				
				description := fmt.Sprintf("Implemented for task %s", taskID)
				
				fileChanges = append(fileChanges, FileChange{
					Path:        filePath,
					ChangeType:  changeType,
					Description: description,
					LastChanged: time.Now(),
					CommitID:    "", // Historical - no commit ID
					IsPending:   false,
				})
				
				seenFiles[filePath] = true
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return fileChanges, nil
}
