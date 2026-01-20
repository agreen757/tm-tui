package dialog

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogRotator manages log file rotation and compression
type LogRotator struct {
	maxFiles       int
	maxSizeBytes   int64
	mu             sync.Mutex
	lastCheckTime  map[string]time.Time
	cleanupCounter int
}

// NewLogRotator creates a new log rotator with sensible defaults
// maxFiles: number of logs to keep per operation type (e.g., 10)
// maxSizeBytes: maximum size before rotation (e.g., 10MB = 10*1024*1024)
func NewLogRotator(maxFiles int, maxSizeBytes int64) *LogRotator {
	if maxFiles <= 0 {
		maxFiles = 10
	}
	if maxSizeBytes <= 0 {
		maxSizeBytes = 10 * 1024 * 1024 // 10MB default
	}

	return &LogRotator{
		maxFiles:      maxFiles,
		maxSizeBytes:  maxSizeBytes,
		lastCheckTime: make(map[string]time.Time),
	}
}

// CheckAndRotate checks if log rotation is needed for a given log file
// It performs rotation if:
// 1. The current file exceeds maxSizeBytes
// 2. There are more than maxFiles logs of this type
// This is thread-safe and should be called before writing to a log file
func (lr *LogRotator) CheckAndRotate(logPath string) error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	// Increment cleanup counter
	lr.cleanupCounter++

	// Run cleanup every 100 operations
	if lr.cleanupCounter >= 100 {
		lr.cleanupLastCheckTimeMap()
	}

	// Extract the log directory and pattern
	logDir := filepath.Dir(logPath)
	logName := filepath.Base(logPath)

	// Get the pattern (e.g., "git-*.log" from "git-status.log")
	pattern := lr.extractLogPattern(logName)

	// Only check rotation once per minute per directory
	now := time.Now()
	lastCheck, exists := lr.lastCheckTime[logDir]
	if exists && now.Sub(lastCheck) < time.Minute {
		return nil
	}
	lr.lastCheckTime[logDir] = now

	// Check if current file needs rotation based on size
	if fileExists(logPath) {
		fileInfo, err := os.Stat(logPath)
		if err == nil && fileInfo.Size() > lr.maxSizeBytes {
			// Rotate current log
			if err := lr.rotateLog(logPath); err != nil {
				// Log rotation errors are non-fatal
				fmt.Fprintf(os.Stderr, "Warning: Failed to rotate log: %v\n", err)
			}
		}
	}

	// Clean up old logs based on pattern
	if err := lr.cleanupOldLogs(logDir, pattern); err != nil {
		// Cleanup errors are non-fatal
		fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup old logs: %v\n", err)
	}

	return nil
}

// cleanupLastCheckTimeMap removes entries for directories that no longer exist
func (lr *LogRotator) cleanupLastCheckTimeMap() {
	for dir := range lr.lastCheckTime {
		// Check if directory still exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Directory no longer exists, remove from map
			delete(lr.lastCheckTime, dir)
		}
	}
	// Reset counter after cleanup
	lr.cleanupCounter = 0
}

// rotateLog renames current log and optionally compresses it
func (lr *LogRotator) rotateLog(logPath string) error {
	if !fileExists(logPath) {
		return nil
	}

	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(logPath)
	name := filepath.Base(logPath)
	nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))

	// Create rotated filename with timestamp
	rotatedPath := filepath.Join(dir, fmt.Sprintf("%s.%s.log", nameWithoutExt, timestamp))

	// Rename current log
	if err := os.Rename(logPath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rename log: %w", err)
	}

	// Compress in background to avoid blocking
	go func() {
		if err := lr.compressLog(rotatedPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to compress log: %v\n", err)
		}
	}()

	return nil
}

// compressLog compresses a log file with gzip
func (lr *LogRotator) compressLog(logPath string) error {
	if !fileExists(logPath) {
		return nil
	}

	// Read original file
	inputFile, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log for compression: %w", err)
	}
	defer inputFile.Close()

	// Create compressed file
	compressedPath := logPath + ".gz"
	outputFile, err := os.Create(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to create compressed log: %w", err)
	}
	defer outputFile.Close()

	// Compress with gzip
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, inputFile); err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	// Remove original uncompressed file
	if err := os.Remove(logPath); err != nil {
		return fmt.Errorf("failed to remove original log after compression: %w", err)
	}

	return nil
}

// cleanupOldLogs removes logs beyond the maxFiles limit
func (lr *LogRotator) cleanupOldLogs(logDir string, pattern string) error {
	// Read directory
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	// Find and collect matching log files
	var logFiles []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Match log files for this pattern (e.g., "git-*.log" or "git-*.log.gz")
		if matchesPattern(name, pattern) {
			info, err := entry.Info()
			if err == nil {
				logFiles = append(logFiles, info)
			}
		}
	}

	// If we have more logs than maxFiles, remove oldest ones
	if len(logFiles) > lr.maxFiles {
		// Sort by modification time (oldest first)
		sort.Slice(logFiles, func(i, j int) bool {
			return logFiles[i].ModTime().Before(logFiles[j].ModTime())
		})

		// Remove oldest logs
		numToRemove := len(logFiles) - lr.maxFiles
		for i := 0; i < numToRemove; i++ {
			filePath := filepath.Join(logDir, logFiles[i].Name())
			if err := os.Remove(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to remove old log %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

// extractLogPattern extracts the pattern from a log filename
// E.g., "git-status.log" -> "git-*.log"
func (lr *LogRotator) extractLogPattern(logName string) string {
	// Remove timestamp and .gz if present
	name := strings.TrimSuffix(logName, ".gz")
	name = strings.TrimSuffix(name, ".log")

	// Remove timestamp suffix (e.g., ".20060102-150405")
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		// Check if last part looks like a timestamp
		if isTimestamp(parts[len(parts)-1]) {
			name = strings.Join(parts[:len(parts)-1], ".")
		}
	}

	// Return pattern (e.g., "git-status*")
	return name + "*"
}

// matchesPattern checks if a filename matches a pattern
func matchesPattern(filename string, pattern string) bool {
	// Simple pattern matching: "git-*" matches any file starting with "git-"
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(filename, prefix)
	}
	return filename == pattern
}

// isTimestamp checks if a string looks like a timestamp
func isTimestamp(s string) bool {
	// Check if format matches YYYYMMdd-HHMMSS
	if len(s) != 15 {
		return false
	}
	if s[8] != '-' {
		return false
	}
	// Simplified check - in production might want more validation
	return true
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// RotateOnStartup checks for and rotates any existing logs
// This should be called during application initialization
func (lr *LogRotator) RotateOnStartup(basePath string) error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	// Handle both .taskmaster/logs and tag-specific directories
	logsDir := filepath.Join(basePath, "logs")
	if fileExists(logsDir) {
		if err := lr.cleanupOldLogs(logsDir, "git-*"); err != nil {
			// Non-fatal
			fmt.Fprintf(os.Stderr, "Warning: Failed to rotate logs on startup: %v\n", err)
		}
	}

	// For tag-specific directories, we'd need a registry, so just handle generic logs

	return nil
}
