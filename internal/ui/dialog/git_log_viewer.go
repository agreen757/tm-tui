package dialog

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GitLogEntry represents a single structured log entry
type GitLogEntry struct {
	Level      string   `json:"level"`
	Time       string   `json:"time"`
	CommandID  string   `json:"command_id"`
	Tag        string   `json:"tag"`
	GitArgs    string   `json:"git_args"`
	Event      string   `json:"event"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Source     string   `json:"source"`
	Output     string   `json:"output"`
	DurationMs int64    `json:"duration_ms"`
	ExitCode   int      `json:"exit_code"`
	Message    string   `json:"message"`
	Error      string   `json:"error"`
	Raw        string   `json:"-"` // Raw JSON line
}

// GitLogViewer provides methods to read and view git logs
type GitLogViewer struct {
	logsBaseDir string
}

// NewGitLogViewer creates a new log viewer
func NewGitLogViewer(logsBaseDir string) *GitLogViewer {
	if logsBaseDir == "" {
		logsBaseDir = ".taskmaster"
	}
	return &GitLogViewer{
		logsBaseDir: logsBaseDir,
	}
}

// ViewRecentLogs returns the most recent log entries, optionally filtered
func (glv *GitLogViewer) ViewRecentLogs(limit int, logType string) ([]GitLogEntry, error) {
	var entries []GitLogEntry

	// Read logs from all directories
	baseDir := glv.logsBaseDir
	if err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Match log files
		name := info.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			return nil
		}

		// Filter by type if specified
		if logType != "" && !strings.Contains(name, logType) {
			return nil
		}

		// Read log file
		fileEntries, err := glv.readLogFile(path)
		if err == nil {
			entries = append(entries, fileEntries...)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Sort by time descending (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	// Limit results
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// ViewLogsByTimeRange returns logs within a time range
func (glv *GitLogViewer) ViewLogsByTimeRange(since, until time.Time) ([]GitLogEntry, error) {
	var entries []GitLogEntry

	// Read logs from all directories
	baseDir := glv.logsBaseDir
	if err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Match log files
		name := info.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			return nil
		}

		// Check file modification time
		modTime := info.ModTime()
		if modTime.Before(since) || modTime.After(until) {
			return nil
		}

		// Read log file
		fileEntries, err := glv.readLogFile(path)
		if err == nil {
			entries = append(entries, fileEntries...)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Sort by time descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	return entries, nil
}

// SearchLogs searches logs for a specific string in various fields
func (glv *GitLogViewer) SearchLogs(searchTerm string) ([]GitLogEntry, error) {
	var entries []GitLogEntry

	// Read logs from all directories
	baseDir := glv.logsBaseDir
	if err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Match log files
		name := info.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			return nil
		}

		// Read log file
		fileEntries, err := glv.readLogFile(path)
		if err == nil {
			// Filter by search term
			for _, entry := range fileEntries {
				if glv.matchesSearch(entry, searchTerm) {
					entries = append(entries, entry)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Sort by time descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	return entries, nil
}

// readLogFile reads a log file (supports gzip)
func (glv *GitLogViewer) readLogFile(path string) ([]GitLogEntry, error) {
	var file io.ReadCloser
	var err error

	// Open file (handle gzip)
	if strings.HasSuffix(path, ".gz") {
		gzFile, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer gzFile.Close()

		file, err = gzip.NewReader(gzFile)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var entries []GitLogEntry

	// Read lines as JSON
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry GitLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip non-JSON lines (they might be regular text logs)
			continue
		}

		entry.Raw = line
		entries = append(entries, entry)
	}

	return entries, nil
}

// matchesSearch checks if a log entry matches the search term
func (glv *GitLogViewer) matchesSearch(entry GitLogEntry, searchTerm string) bool {
	searchLower := strings.ToLower(searchTerm)

	checks := []string{
		strings.ToLower(entry.CommandID),
		strings.ToLower(entry.Command),
		strings.ToLower(entry.Message),
		strings.ToLower(entry.Output),
		strings.ToLower(entry.Error),
		strings.ToLower(entry.Tag),
		strings.ToLower(entry.Event),
	}

	for _, check := range checks {
		if strings.Contains(check, searchLower) {
			return true
		}
	}

	return false
}

// FormatAsText formats log entries as human-readable text
func (glv *GitLogViewer) FormatAsText(entries []GitLogEntry) string {
	var output strings.Builder

	for _, entry := range entries {
		output.WriteString(fmt.Sprintf("[%s] %s - %s\n", entry.Time, entry.CommandID, entry.Message))

		if entry.Command != "" {
			output.WriteString(fmt.Sprintf("  Command: %s %v\n", entry.Command, entry.Args))
		}

		if entry.Output != "" {
			output.WriteString(fmt.Sprintf("  Output: %s\n", entry.Output))
		}

		if entry.Error != "" {
			output.WriteString(fmt.Sprintf("  Error: %s\n", entry.Error))
		}

		if entry.DurationMs > 0 {
			output.WriteString(fmt.Sprintf("  Duration: %dms\n", entry.DurationMs))
		}

		if entry.ExitCode != 0 {
			output.WriteString(fmt.Sprintf("  Exit Code: %d\n", entry.ExitCode))
		}

		output.WriteString("\n")
	}

	return output.String()
}

// FormatAsJSON formats log entries as JSON
func (glv *GitLogViewer) FormatAsJSON(entries []GitLogEntry, pretty bool) (string, error) {
	var data []map[string]interface{}

	for _, entry := range entries {
		if entry.Raw != "" {
			// Use original JSON
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(entry.Raw), &obj); err == nil {
				data = append(data, obj)
				continue
			}
		}

		// Fall back to marshaling the entry
		obj := map[string]interface{}{
			"level":      entry.Level,
			"time":       entry.Time,
			"command_id": entry.CommandID,
			"message":    entry.Message,
		}

		if entry.Command != "" {
			obj["command"] = entry.Command
		}
		if len(entry.Args) > 0 {
			obj["args"] = entry.Args
		}
		if entry.DurationMs > 0 {
			obj["duration_ms"] = entry.DurationMs
		}

		data = append(data, obj)
	}

	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	return string(output), err
}

// GetLogTypes returns available log types (based on file names)
func (glv *GitLogViewer) GetLogTypes() ([]string, error) {
	typesMap := make(map[string]bool)

	// Walk through logs directory
	baseDir := glv.logsBaseDir
	if err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			return nil
		}

		// Extract type from filename (e.g., "git-status.log" -> "git-status")
		nameWithoutExt := strings.TrimSuffix(name, ".gz")
		nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".log")

		// Remove timestamp suffix if present
		parts := strings.Split(nameWithoutExt, ".")
		if len(parts) > 1 && isTimestamp(parts[len(parts)-1]) {
			nameWithoutExt = strings.Join(parts[:len(parts)-1], ".")
		}

		typesMap[nameWithoutExt] = true

		return nil
	}); err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	var types []string
	for t := range typesMap {
		types = append(types, t)
	}
	sort.Strings(types)

	return types, nil
}
