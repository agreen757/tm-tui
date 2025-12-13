package dialog

import (
	"fmt"
	"os"
	"time"
)

// DebugLogger is a simple file logger for debugging
type DebugLogger struct {
	file *os.File
}

var debugLogger *DebugLogger

// InitDebugLogger initializes the debug logger
func InitDebugLogger(filename string) error {
	// Close existing logger if any
	if debugLogger != nil && debugLogger.file != nil {
		debugLogger.file.Close()
	}

	// Create or append to the log file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	debugLogger = &DebugLogger{file: file}
	
	// Write header
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	debugLogger.Logf("=== Debug session started at %s ===\n", timestamp)
	
	return nil
}

// Log writes a message to the debug log
func (d *DebugLogger) Log(message string) {
	if d.file == nil {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(d.file, "[%s] %s\n", timestamp, message)
}

// Logf writes a formatted message to the debug log
func (d *DebugLogger) Logf(format string, args ...interface{}) {
	if d.file == nil {
		return
	}
	
	message := fmt.Sprintf(format, args...)
	d.Log(message)
}

// Log is a convenience function to write to the global debug logger
func Log(message string) {
	if debugLogger != nil {
		debugLogger.Log(message)
	}
}

// Logf is a convenience function to write to the global debug logger
func Logf(format string, args ...interface{}) {
	if debugLogger != nil {
		debugLogger.Logf(format, args...)
	}
}

// Close closes the debug logger
func CloseDebugLogger() {
	if debugLogger != nil && debugLogger.file != nil {
		debugLogger.file.Close()
		debugLogger.file = nil
	}
}