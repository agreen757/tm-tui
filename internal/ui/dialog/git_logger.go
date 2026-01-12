package dialog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Global log rotator for git operations
var (
	globalLogRotator *LogRotator
	rotatorMutex     sync.Once
)

// getLogRotator returns the global log rotator, initializing if needed
func getLogRotator() *LogRotator {
	rotatorMutex.Do(func() {
		// 10 logs max, 10MB per file
		globalLogRotator = NewLogRotator(10, 10*1024*1024)
	})
	return globalLogRotator
}

// GitLogger wraps zerolog for git operations logging
type GitLogger struct {
	logger  *zerolog.Logger
	file    io.WriteCloser
	start   time.Time
	logPath string
}

// NewGitLogger creates a new structured logger for git operations
// commandID is used for log correlation and filename
// tagName is used for organizing logs in .taskmaster/<tagName>/ directory
func NewGitLogger(commandID string, args []string, tagName string) (*GitLogger, string, error) {
	// Sanitize commandID to ensure valid filename
	sanitizedID := SanitizeTaskIDForFilename(commandID)

	var logsDir string
	var logFileName string

	if tagName != "" {
		// Use tag-based directory structure
		// .taskmaster/<tag-name>/<command-id>.log
		logsDir = filepath.Join(".taskmaster", tagName)
		logFileName = fmt.Sprintf("%s.log", sanitizedID)
	} else {
		// Fallback to generic logs directory with timestamp
		logsDir = filepath.Join(".taskmaster", "logs")
		timestamp := time.Now().Format("20060102-150405")
		logFileName = fmt.Sprintf("git-command-%s-%s.log", sanitizedID, timestamp)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create logs directory: %w", err)
	}

	logPath := filepath.Join(logsDir, logFileName)

	// Create the log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create log file: %w", err)
	}

	// Configure zerolog for JSON output
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(logFile).
		With().
		Timestamp().
		Str("command_id", commandID).
		Str("tag", tagName).
		Str("git_args", fmt.Sprintf("%v", args)).
		Logger()

	// Log the start event
	logger.Info().
		Str("event", "start").
		Str("command", "git").
		Strs("args", args).
		Msg("Git command started")

	// Check and rotate logs if needed
	rotator := getLogRotator()
	if err := rotator.CheckAndRotate(logPath); err != nil {
		// Non-fatal error, continue with logging
		fmt.Fprintf(os.Stderr, "Warning: Log rotation check failed: %v\n", err)
	}

	return &GitLogger{
		logger:  &logger,
		file:    logFile,
		start:   time.Now(),
		logPath: logPath,
	}, logPath, nil
}

// LogOutput logs a line of output from the git command
func (gl *GitLogger) LogOutput(source string, line string) {
	gl.logger.Debug().
		Str("source", source).
		Str("output", line).
		Msg("Command output")
}

// LogWarning logs a warning during execution
func (gl *GitLogger) LogWarning(msg string, err error) {
	gl.logger.Warn().
		Err(err).
		Msg(msg)
}

// LogError logs an error during execution
func (gl *GitLogger) LogError(msg string, err error) {
	gl.logger.Error().
		Err(err).
		Msg(msg)
}

// LogCompletion logs the completion of the git command
func (gl *GitLogger) LogCompletion(exitCode int, err error) {
	duration := time.Since(gl.start)
	event := gl.logger.Info()

	if err != nil {
		event = gl.logger.Error()
		event.Err(err)
	}

	event.
		Int("exit_code", exitCode).
		Int64("duration_ms", duration.Milliseconds()).
		Str("event", "completion").
		Msg("Git command completed")
}

// LogCancellation logs the cancellation of the git command
func (gl *GitLogger) LogCancellation() {
	duration := time.Since(gl.start)
	gl.logger.Warn().
		Int64("duration_ms", duration.Milliseconds()).
		Str("event", "cancellation").
		Msg("Git command cancelled")
}

// Close closes the log file and ensures all data is written
func (gl *GitLogger) Close() error {
	if gl.file != nil {
		return gl.file.Close()
	}
	return nil
}
