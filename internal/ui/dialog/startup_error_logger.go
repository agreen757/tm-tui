package dialog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogStartupError logs a startup error when NewGitLogger fails
// This ensures we capture errors even when the logger can't be initialized
func LogStartupError(commandID string, args []string, tagName string, errorMsg string) error {
	// Sanitize commandID
	sanitizedID := SanitizeTaskIDForFilename(commandID)

	var logsDir string
	var logFileName string

	if tagName != "" {
		logsDir = filepath.Join(".taskmaster", tagName)
		logFileName = fmt.Sprintf("%s.log", sanitizedID)
	} else {
		logsDir = filepath.Join(".taskmaster", "logs")
		timestamp := time.Now().Format("20060102-150405")
		logFileName = fmt.Sprintf("git-command-%s-%s.log", sanitizedID, timestamp)
	}

	// Create directory if needed
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	logPath := filepath.Join(logsDir, logFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create error log file: %w", err)
	}
	defer logFile.Close()

	// Write error as JSON
	entry := map[string]interface{}{
		"level":       "error",
		"time":        time.Now().Format(time.RFC3339Nano),
		"command_id":  commandID,
		"tag":         tagName,
		"git_args":    args,
		"event":       "startup_failed",
		"command":     "git",
		"error":       errorMsg,
		"duration_ms": 0,
	}

	jsonData, _ := json.Marshal(entry)
	logFile.WriteString(string(jsonData) + "\n")

	return nil
}
