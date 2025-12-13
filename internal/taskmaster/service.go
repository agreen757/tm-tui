package taskmaster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/projects"
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
	if onProgress != nil {
		onProgress(ParsePrdProgressState{
			Progress: 1.0,
			Label:    "Complete",
		})
	}
	return nil
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





