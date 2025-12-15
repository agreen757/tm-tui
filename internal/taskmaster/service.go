package taskmaster

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/projects"
)

// Service handles Task Master integration with thread-safe access
type Service struct {
	// RootDir is the absolute path to the directory containing .taskmaster
	RootDir string
	
	// Tasks contains the root-level tasks
	Tasks []Task
	
	// TaskIndex provides O(1) lookup by task ID
	TaskIndex map[string]*Task
	
	// config stores the application configuration
	config *config.Config
	
	// warnings stores validation warnings from the last load
	warnings []ValidationWarning
	
	// lastModTime tracks the last modification time of tasks.json for caching
	lastModTime time.Time
	
	// available indicates whether a .taskmaster directory was found
	available bool
	
	// watcher monitors tasks.json for changes
	watcher *config.Watcher
	
	// reloadChan signals when tasks should be reloaded
	reloadChan chan struct{}
	
	// mu protects concurrent access to task data
	mu sync.RWMutex
}

// ComplexityProgressState captures incremental progress during analysis.
type ComplexityProgressState struct {
	TasksAnalyzed int
	TotalTasks    int
	CurrentTaskID string
}

// ParsePrdMode defines append vs replace behavior for CLI parse workflow.
type ParsePrdMode string

const (
	ParsePrdModeAppend  ParsePrdMode = "append"
	ParsePrdModeReplace ParsePrdMode = "replace"
)

// ParsePrdProgressState represents incremental progress updates from the CLI.
type ParsePrdProgressState struct {
	Progress float64
	Label    string
}

// NewService creates a new Task Master service.
// It detects the .taskmaster directory, loads tasks, and performs initial validation.
// If no .taskmaster directory is found, returns a service with available=false
// but does not return an error, allowing the TUI to degrade gracefully.
func NewService(cfg *config.Config) (*Service, error) {
	svc := &Service{
		config:     cfg,
		TaskIndex:  make(map[string]*Task),
		available:  false,
		reloadChan: make(chan struct{}, 1),
	}
	
	// Try to detect taskmaster root
	var rootDir string
	var err error
	
	if cfg.TaskMasterPath != "" {
		// Use configured path if available
		rootDir = cfg.TaskMasterPath
		// Verify it exists
		tmPath := filepath.Join(rootDir, ".taskmaster")
		if _, err := os.Stat(tmPath); err != nil {
			return svc, nil // Not available, but not an error
		}
	} else {
		// Auto-detect from current directory
		rootDir, err = FindTaskmasterRoot()
		if err == ErrNotFound {
			return svc, nil // Not available, but not an error
		}
		if err != nil {
			return nil, fmt.Errorf("failed to detect taskmaster root: %w", err)
		}
	}
	
	svc.RootDir = rootDir
	svc.available = true
	
	// Load tasks initially
	ctx := context.Background()
	if err := svc.LoadTasks(ctx); err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	
	return svc, nil
}

// LoadTasks loads tasks from tasks.json with mutex protection.
// It checks the file modification time and skips reload if unchanged
// unless force is true (via context value).
func (s *Service) LoadTasks(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.available {
		return fmt.Errorf("taskmaster not available")
	}
	
	tasksPath := filepath.Join(s.RootDir, ".taskmaster", "tasks", "tasks.json")
	
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	// Check modification time for caching
	info, err := os.Stat(tasksPath)
	if err != nil {
		return fmt.Errorf("failed to stat tasks file: %w", err)
	}
	
	modTime := info.ModTime()
	force := ctx.Value("force") != nil && ctx.Value("force").(bool)
	
	if !force && !s.lastModTime.IsZero() && modTime.Equal(s.lastModTime) {
		// File hasn't changed, skip reload
		return nil
	}
	
	// Get the active tag from config, default to "master"
	tag := s.config.ActiveTag
	if tag == "" {
		tag = "master"
	}
	
	// Load tasks from file with the specified tag
	tasks, err := LoadTasksFromFile(s.RootDir, tag)
	if err != nil {
		return err
	}
	
	s.Tasks = tasks
	s.lastModTime = modTime
	
	// Rebuild index and validate
	s.rebuildIndexAndValidate()
	
	return nil
}

// rebuildIndexAndValidate rebuilds the task index and performs validation.
// Must be called with write lock held.
func (s *Service) rebuildIndexAndValidate() {
	// Build index and collect warnings
	index, indexWarnings := buildTaskIndex(s.Tasks)
	s.TaskIndex = index
	
	// Validate tasks
	validationWarnings := validateTasks(s.Tasks, index)
	
	// Combine all warnings
	s.warnings = append(indexWarnings, validationWarnings...)
}

// GetTasks returns all root-level tasks and any validation warnings.
// Thread-safe for concurrent access.
func (s *Service) GetTasks() ([]Task, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Convert warnings to strings
	warningStrs := make([]string, len(s.warnings))
	for i, w := range s.warnings {
		warningStrs[i] = w.String()
	}
	
	return s.Tasks, warningStrs
}

// GetTaskByID returns a task by ID using the index for O(1) lookup.
// Returns the task pointer and true if found, nil and false otherwise.
// Thread-safe for concurrent access.
func (s *Service) GetTaskByID(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, ok := s.TaskIndex[id]
	return task, ok
}

// GetValidationWarnings returns all validation warnings from the last load.
// Thread-safe for concurrent access.
func (s *Service) GetValidationWarnings() []ValidationWarning {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	warnings := make([]ValidationWarning, len(s.warnings))
	copy(warnings, s.warnings)
	return warnings
}

// IsAvailable returns true if a .taskmaster directory was found and tasks are loaded.
func (s *Service) IsAvailable() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.available
}

// GetTask returns a task by ID (backwards compatibility wrapper for GetTaskByID).
// Returns an error if the task is not found.
func (s *Service) GetTask(id string) (*Task, error) {
	task, ok := s.GetTaskByID(id)
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// GetNextTask finds the next available task to work on.
// A task is available if it has status "pending" and all its dependencies are complete.
func (s *Service) GetNextTask() (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for i := range s.Tasks {
		if task := s.findNextTask(&s.Tasks[i]); task != nil {
			return task, nil
		}
	}
	return nil, fmt.Errorf("no available tasks found")
}

// findNextTask recursively finds the next pending task with no pending dependencies.
// Must be called with read lock held.
func (s *Service) findNextTask(task *Task) *Task {
	// Check if this task is available
	if task.Status == StatusPending {
		// Check if dependencies are satisfied
		if !task.HasBlockedDependencies(s.TaskIndex) {
			return task
		}
	}
	
	// Check subtasks
	for i := range task.Subtasks {
		if found := s.findNextTask(&task.Subtasks[i]); found != nil {
			return found
		}
	}
	
	return nil
}

// GetTaskCount returns count of tasks by status.
// Thread-safe for concurrent access.
func (s *Service) GetTaskCount() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	counts := map[string]int{
		StatusPending:    0,
		StatusInProgress: 0,
		StatusDone:       0,
		StatusBlocked:    0,
		StatusDeferred:   0,
		StatusCancelled:  0,
	}
	
	for i := range s.Tasks {
		s.countTaskStatus(&s.Tasks[i], counts)
	}
	
	return counts
}

// countTaskStatus recursively counts tasks by status.
// Must be called with read lock held.
func (s *Service) countTaskStatus(task *Task, counts map[string]int) {
	counts[task.Status]++
	for i := range task.Subtasks {
		s.countTaskStatus(&task.Subtasks[i], counts)
	}
}

// StartWatcher begins watching tasks.json for changes with a 300ms debounce.
// Changes trigger automatic reloads that can be consumed via ReloadEvents().
// Returns an error if watcher cannot be started or if service is not available.
func (s *Service) StartWatcher(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.available {
		return fmt.Errorf("taskmaster not available")
	}

	if s.watcher != nil {
		return fmt.Errorf("watcher already started")
	}

	// Create watcher for tasks.json
	tasksPath := filepath.Join(s.RootDir, ".taskmaster", "tasks", "tasks.json")
	watcher, err := config.NewWatcher(ctx, tasksPath)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Start watching with 300ms debounce
	if err := watcher.Start(300 * time.Millisecond); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	s.watcher = watcher

	// Start goroutine to handle file change events
	go s.handleFileChanges(ctx)

	return nil
}

// handleFileChanges processes file change notifications from the watcher
func (s *Service) handleFileChanges(ctx context.Context) {
	if s.watcher == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-s.watcher.Events():
			// Reload tasks on file change
			if err := s.LoadTasks(ctx); err != nil {
				// Log error but don't stop watching
				fmt.Fprintf(os.Stderr, "Error reloading tasks: %v\n", err)
				continue
			}

			// Signal that tasks were reloaded
			select {
			case s.reloadChan <- struct{}{}:
			default:
				// Channel full, reload notification already pending
			}

		case err := <-s.watcher.Errors():
			// Log error but continue watching
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
}

// StopWatcher stops the file watcher if it's running
func (s *Service) StopWatcher() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcher == nil {
		return nil
	}

	err := s.watcher.Stop()
	s.watcher = nil
	return err
}

// ReloadEvents returns a channel that signals when tasks have been reloaded
// from disk due to file changes. UI components should listen to this channel
// and refresh their view when signaled.
func (s *Service) ReloadEvents() <-chan struct{} {
	return s.reloadChan
}



// ActiveProjectMetadata returns the currently active project metadata
func (s *Service) ActiveProjectMetadata() *projects.Metadata {
	return nil
}

// AnalyzeComplexity performs complexity analysis on tasks
func (s *Service) AnalyzeComplexity(ctx context.Context, scope string, taskID string, tags []string) (*ComplexityReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.available {
		return nil, fmt.Errorf("taskmaster not available")
	}
	
	tasksToAnalyze := make([]*Task, 0)
	for i := range s.Tasks {
		tasksToAnalyze = append(tasksToAnalyze, &s.Tasks[i])
	}
	
	complexities := AnalyzeComplexity(tasksToAnalyze)
	report := NewComplexityReport(complexities, scope, tags)
	return report, nil
}

// AnalyzeComplexityWithProgress performs complexity analysis with progress reporting
func (s *Service) AnalyzeComplexityWithProgress(ctx context.Context, scope string, taskID string, tags []string, onProgress func(ComplexityProgressState)) (*ComplexityReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.available {
		return nil, fmt.Errorf("taskmaster not available")
	}
	
	tasksToAnalyze := make([]*Task, 0)
	for i := range s.Tasks {
		tasksToAnalyze = append(tasksToAnalyze, &s.Tasks[i])
	}
	
	if onProgress != nil {
		onProgress(ComplexityProgressState{
			TasksAnalyzed: len(tasksToAnalyze),
			TotalTasks:    len(tasksToAnalyze),
		})
	}
	
	complexities := AnalyzeComplexity(tasksToAnalyze)
	report := NewComplexityReport(complexities, scope, tags)
	return report, nil
}

// GetLatestComplexityReport returns the latest cached complexity report
func (s *Service) GetLatestComplexityReport() *ComplexityReport {
	return nil
}

// ExportComplexityReport exports a complexity report in the specified format
func (s *Service) ExportComplexityReport(ctx context.Context, format string, outputPath string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// ParsePRDWithProgress parses a PRD file and generates tasks with progress reporting
func (s *Service) ParsePRDWithProgress(ctx context.Context, inputPath string, mode ParsePrdMode, onProgress func(ParsePrdProgressState)) error {
	if !s.available {
		return fmt.Errorf("taskmaster not available")
	}

	// Build CLI command args
	args := []string{"parse-prd", inputPath}
	
	// Add mode flag
	if mode == ParsePrdModeAppend {
		args = append(args, "--append")
	}

	// Execute command with streaming output
	cmd := exec.CommandContext(ctx, "task-master", args...)
	cmd.Dir = s.RootDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output and parse progress
	progressCh := make(chan ParsePrdProgressState, 10)
	errCh := make(chan error, 1)

	// Parse stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if state := parseParsePrdProgress(line); state.Label != "" {
				select {
				case progressCh <- state:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Capture stderr
	go func() {
		errOutput, _ := io.ReadAll(stderr)
		if len(errOutput) > 0 {
			select {
			case errCh <- fmt.Errorf("CLI error: %s", string(errOutput)):
			default:
			}
		}
	}()

	// Forward progress updates to callback
	go func() {
		for {
			select {
			case state := <-progressCh:
				if onProgress != nil {
					onProgress(state)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion
	cmdErr := cmd.Wait()

	// Check for errors
	if cmdErr != nil {
		if ctx.Err() == context.Canceled {
			return context.Canceled
		}
		select {
		case err := <-errCh:
			return fmt.Errorf("command failed: %w: %v", cmdErr, err)
		default:
			return fmt.Errorf("command failed: %w", cmdErr)
		}
	}

	// Send completion update
	if onProgress != nil {
		onProgress(ParsePrdProgressState{
			Progress: 1.0,
			Label:    "Complete",
		})
	}

	// Reload tasks after parsing
	reloadCtx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(reloadCtx)
}

// ExpandTaskWithProgress expands a task with subtasks based on AI analysis and progress reporting
func (s *Service) ExpandTaskWithProgress(ctx context.Context, taskID string, opts ExpandTaskOptions, prompt string, onProgress func(ExpandProgressState)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.available {
		return fmt.Errorf("taskmaster not available")
	}
	
	task, ok := s.TaskIndex[taskID]
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	
	if onProgress != nil {
		onProgress(ExpandProgressState{
			Stage:    "Generating subtask drafts...",
			Progress: 0.3,
		})
	}
	
	drafts := ExpandTaskDrafts(task, opts)
	if len(drafts) == 0 {
		return fmt.Errorf("no subtasks generated")
	}
	
	if onProgress != nil {
		onProgress(ExpandProgressState{
			Stage:    "Applying subtasks...",
			Progress: 0.7,
		})
	}
	
	_, err := ApplySubtaskDrafts(task, drafts)
	if err != nil {
		return fmt.Errorf("failed to apply subtasks: %w", err)
	}
	
	if onProgress != nil {
		onProgress(ExpandProgressState{
			Stage:    "Complete",
			Progress: 1.0,
		})
	}
	
	return nil
}

// ExecuteExpandWithProgress executes the task-master expand CLI command with
// real-time progress reporting. It supports multiple expansion scopes:
//   - "single": Expand a single task by ID
//   - "all": Expand all tasks in the project
//   - "range": Expand tasks within a specified ID range
//   - "tag": Expand tasks matching specified tags
//
// The onProgress callback receives updates during execution, including stage,
// progress percentage, and current task being processed. The context can be
// used for cancellation.
//
// After successful expansion, tasks are automatically reloaded.
func (s *Service) ExecuteExpandWithProgress(
	ctx context.Context,
	scope string,
	taskID string,
	fromID string,
	toID string,
	tags []string,
	opts ExpandTaskOptions,
	onProgress func(ExpandProgressState),
) error {
	if !s.available {
		return fmt.Errorf("taskmaster not available")
	}

	// Build CLI command args
	args := []string{"expand"}

	switch scope {
	case "single":
		if taskID == "" {
			return fmt.Errorf("task ID is required for single scope")
		}
		args = append(args, fmt.Sprintf("--id=%s", taskID))
	case "all":
		args = append(args, "--all")
	case "range":
		if fromID != "" {
			args = append(args, fmt.Sprintf("--from=%s", fromID))
		}
		if toID != "" {
			args = append(args, fmt.Sprintf("--to=%s", toID))
		}
		if fromID == "" && toID == "" {
			return fmt.Errorf("at least one of from/to ID is required for range scope")
		}
	case "tag":
		if len(tags) == 0 {
			return fmt.Errorf("tags are required for tag scope")
		}
		// Note: CLI may not support --tag flag yet, this is a future enhancement
		for _, tag := range tags {
			args = append(args, fmt.Sprintf("--tag=%s", tag))
		}
	default:
		return fmt.Errorf("invalid scope: %s", scope)
	}

	// Add optional flags
	if opts.UseAI {
		args = append(args, "--research")
	}
	if opts.NumSubtasks > 0 {
		args = append(args, fmt.Sprintf("--num=%d", opts.NumSubtasks))
	}
	if opts.Force {
		args = append(args, "--force")
	}

	// Execute command with streaming output
	cmd := exec.CommandContext(ctx, "task-master", args...)
	cmd.Dir = s.RootDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output and parse progress
	progressCh := make(chan ExpandProgressState, 10)
	errCh := make(chan error, 1)

	// Parse stdout in goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if state := parseExpandProgress(line); state.Message != "" {
				select {
				case progressCh <- state:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Capture stderr
	go func() {
		errOutput, _ := io.ReadAll(stderr)
		if len(errOutput) > 0 {
			select {
			case errCh <- fmt.Errorf("CLI error: %s", string(errOutput)):
			default:
			}
		}
	}()

	// Forward progress updates to callback
	go func() {
		for {
			select {
			case state := <-progressCh:
				if onProgress != nil {
					onProgress(state)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion
	cmdErr := cmd.Wait()

	// Check for errors
	if cmdErr != nil {
		if ctx.Err() == context.Canceled {
			return context.Canceled
		}
		select {
		case err := <-errCh:
			return fmt.Errorf("command failed: %w: %v", cmdErr, err)
		default:
			return fmt.Errorf("command failed: %w", cmdErr)
		}
	}

	// Send completion update
	if onProgress != nil {
		onProgress(ExpandProgressState{
			Stage:    "Complete",
			Progress: 1.0,
			Message:  "Expansion complete, reloading tasks...",
		})
	}

	// Reload tasks after expansion
	reloadCtx := context.WithValue(context.Background(), "force", true)
	return s.LoadTasks(reloadCtx)
}

// parseExpandProgress parses progress information from CLI output
func parseExpandProgress(line string) ExpandProgressState {
	state := ExpandProgressState{}

	line = strings.TrimSpace(line)
	
	// Filter out raw file paths and CLI noise
	if strings.Contains(line, "/.taskmaster/") ||
	   strings.HasPrefix(line, "/Users/") ||
	   strings.HasPrefix(line, "/home/") ||
	   strings.HasPrefix(line, "/opt/") ||
	   strings.HasPrefix(line, "C:\\") ||
	   strings.HasPrefix(line, "D:\\") ||
	   len(line) > 200 ||
	   line == "" {
		return state // Empty state will be ignored
	}

	// Set message only for recognized patterns
	state.Message = line

	// Parse "Expanding task X..."
	if strings.Contains(line, "Expanding task") {
		re := regexp.MustCompile(`Expanding task (\S+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			state.Stage = "Expanding"
			state.CurrentTask = matches[1]
			state.Progress = 0.3
		}
		return state
	}

	// Parse "Generated N subtasks"
	if strings.Contains(line, "Generated") && strings.Contains(line, "subtask") {
		re := regexp.MustCompile(`Generated (\d+) subtask`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				state.SubtasksCreated = count
				state.Stage = "Generating"
				state.Progress = 0.6
			}
		}
		return state
	}

	// Parse "Progress: X/Y"
	if strings.Contains(line, "Progress:") {
		re := regexp.MustCompile(`Progress:\s*(\d+)/(\d+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 2 {
			if current, err1 := strconv.Atoi(matches[1]); err1 == nil {
				if total, err2 := strconv.Atoi(matches[2]); err2 == nil && total > 0 {
					state.TasksExpanded = current
					state.TotalTasks = total
					state.Progress = float64(current) / float64(total)
					state.Stage = "Processing"
				}
			}
		}
		return state
	}

	// Parse "Analyzing..."
	if strings.Contains(line, "Analyzing") {
		state.Stage = "Analyzing"
		state.Progress = 0.1
		return state
	}

	// Parse "Applying..."
	if strings.Contains(line, "Applying") {
		state.Stage = "Applying"
		state.Progress = 0.8
		return state
	}

	// Filter out generic/unrecognized output - return empty state to ignore
	// Only pass through messages that seem intentional/informative
	if len(line) < 10 || 
	   strings.Contains(line, "npm") ||
	   strings.Contains(line, "node") ||
	   strings.Contains(line, "installed") {
		state.Message = "" // Will be filtered out
		return state
	}

	// Default: pass through recognized informational messages
	state.Stage = "Processing"
	state.Progress = 0.5
	return state
}

// parseParsePrdProgress parses progress information from parse-prd CLI output
func parseParsePrdProgress(line string) ParsePrdProgressState {
	state := ParsePrdProgressState{}

	line = strings.TrimSpace(line)
	
	// Filter out empty lines
	if line == "" {
		return state
	}
	
	// Filter out Node.js warnings and deprecation notices
	if strings.Contains(line, "DeprecationWarning") ||
	   strings.Contains(line, "(node:") ||
	   strings.Contains(line, "Use `node --trace") ||
	   strings.Contains(line, "punycode") {
		return state
	}
	
	// Strip [INFO] prefix if present
	line = strings.TrimPrefix(line, "[INFO] ")
	line = strings.TrimSpace(line)
	
	// Parse "Parsing PRD file:" with path
	if strings.Contains(line, "Parsing PRD file:") {
		state.Label = "Reading PRD file..."
		state.Progress = 0.1
		return state
	}
	
	// Parse "Reading PRD content"
	if strings.Contains(line, "Reading PRD content") {
		state.Label = "Reading PRD content..."
		state.Progress = 0.2
		return state
	}
	
	// Parse "Calling AI service" or "Generating"
	if strings.Contains(line, "Calling AI service") || 
	   (strings.Contains(line, "Generating") && strings.Contains(line, "task")) {
		state.Label = "Generating tasks with AI..."
		state.Progress = 0.4
		return state
	}
	
	// Parse "New AI service call"
	if strings.Contains(line, "New AI service call") {
		state.Label = "Calling AI service..."
		state.Progress = 0.5
		return state
	}

	// Parse "Generated N tasks"
	if strings.Contains(line, "Generated") && strings.Contains(line, "task") {
		re := regexp.MustCompile(`Generated (\d+) task`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				state.Label = fmt.Sprintf("Generated %d tasks", count)
				state.Progress = 0.8
			}
		}
		return state
	}

	// Parse "Saving" or "Writing" or "Appending"
	if strings.Contains(line, "Saving") || 
	   strings.Contains(line, "Writing") ||
	   strings.Contains(line, "Appending to existing") {
		state.Label = "Saving tasks..."
		state.Progress = 0.9
		return state
	}
	
	// Parse "Successfully generated" or completion messages
	if strings.Contains(line, "Successfully") || strings.Contains(line, "Complete") {
		state.Label = "Complete"
		state.Progress = 1.0
		return state
	}

	// Filter out very long lines (likely file paths or stack traces)
	if len(line) > 150 {
		return state
	}
	
	// Filter out npm/node installation noise
	if strings.Contains(line, "npm") ||
	   strings.Contains(line, "installed") {
		return state
	}
	
	// Filter out emoji and tag lines
	if strings.HasPrefix(line, "🏷️") || strings.HasPrefix(line, "✓") {
		return state
	}

	// Pass through other informational messages with default progress
	if len(line) >= 10 {
		state.Label = line
		state.Progress = 0.3
		return state
	}
	
	return state
}

// SwitchProject switches to a different project
func (s *Service) SwitchProject(ctx context.Context, projectPath string) (*projects.Metadata, error) {
	return nil, fmt.Errorf("not implemented")
}

// DiscoverProjects discovers projects in specified roots
func (s *Service) DiscoverProjects(ctx context.Context, roots []string) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

// ProjectRegistry returns the project registry
func (s *Service) ProjectRegistry() *projects.Registry {
	return nil
}





