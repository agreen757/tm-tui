package ui

import (
	"context"
	"testing"
	"time"

	"github.com/adriangreen/tm-tui/internal/projects"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/adriangreen/tm-tui/internal/ui/dialog"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// mockTaskService creates a mock task service for testing
func mockTaskService() *mockService {
	return &mockService{
		tasks: []taskmaster.Task{
			{
				ID:          "1",
				Title:       "Simple task",
				Description: "A simple task with not much complexity",
				Status:      "pending",
				Priority:    "low",
			},
			{
				ID:           "2",
				Title:        "Medium complexity task",
				Description:  "A task with some complexity, involving integration work",
				Status:       "in-progress",
				Priority:     "medium",
				Dependencies: []string{"1"},
				Subtasks: []taskmaster.Task{
					{
						ID:    "2.1",
						Title: "Subtask 1",
					},
					{
						ID:    "2.2",
						Title: "Subtask 2",
					},
				},
			},
			{
				ID:           "3",
				Title:        "Complex refactoring task",
				Description:  "A complex refactoring with performance and security implications",
				Details:      "Will require careful testing and validation",
				Status:       "pending",
				Priority:     "high",
				Dependencies: []string{"1", "2"},
				Subtasks: []taskmaster.Task{
					{
						ID:    "3.1",
						Title: "Subtask 1",
					},
					{
						ID:    "3.2",
						Title: "Subtask 2",
					},
					{
						ID:    "3.3",
						Title: "Subtask 3",
					},
				},
			},
		},
		latestReport: nil,
		reloadCh:     make(chan struct{}),
		available:    true,
	}
}

// mockService implements a minimal version of the taskmaster.Service for testing
type mockService struct {
	tasks        []taskmaster.Task
	latestReport *taskmaster.ComplexityReport
	reloadCh     chan struct{}
	available    bool
}

func (s *mockService) GetTasks() ([]taskmaster.Task, []string) {
	return s.tasks, nil
}

func (s *mockService) IsAvailable() bool {
	return s.available
}

func (s *mockService) GetTaskByID(id string) (*taskmaster.Task, bool) {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			return &s.tasks[i], true
		}
	}
	return nil, false
}

func (s *mockService) AnalyzeComplexity(ctx context.Context, scope string, taskID string, tags []string) (*taskmaster.ComplexityReport, error) {
	// Create a simple mock report based on the scope
	var tasksToAnalyze []*taskmaster.Task

	switch scope {
	case "all":
		for i := range s.tasks {
			tasksToAnalyze = append(tasksToAnalyze, &s.tasks[i])
		}
	case "selected":
		if task, ok := s.GetTaskByID(taskID); ok {
			tasksToAnalyze = []*taskmaster.Task{task}
		}
	case "tag":
		// Simplified tag handling for tests
		for i := range s.tasks {
			tasksToAnalyze = append(tasksToAnalyze, &s.tasks[i])
		}
	}

	// Generate mock complexity scores
	complexities := make([]taskmaster.TaskComplexity, 0, len(tasksToAnalyze))
	for _, task := range tasksToAnalyze {
		var level taskmaster.ComplexityLevel
		var score int

		switch task.Priority {
		case "low":
			level = taskmaster.ComplexityLow
			score = 3
		case "medium":
			level = taskmaster.ComplexityMedium
			score = 6
		case "high":
			level = taskmaster.ComplexityHigh
			score = 10
		default:
			level = taskmaster.ComplexityLow
			score = 2
		}

		complexities = append(complexities, taskmaster.TaskComplexity{
			TaskID:     task.ID,
			Level:      level,
			Score:      score,
			Title:      task.Title,
			AnalyzedAt: time.Now(),
		})
	}

	// Create the report
	report := taskmaster.NewComplexityReport(complexities, scope, tags)
	s.latestReport = report

	return report, nil
}

func (s *mockService) AnalyzeComplexityWithProgress(ctx context.Context, scope string, taskID string, tags []string, onProgress func(taskmaster.ComplexityProgressState)) (*taskmaster.ComplexityReport, error) {
	report, err := s.AnalyzeComplexity(ctx, scope, taskID, tags)
	if onProgress != nil && report != nil {
		total := len(report.Tasks)
		for i, task := range report.Tasks {
			onProgress(taskmaster.ComplexityProgressState{
				TasksAnalyzed: i + 1,
				TotalTasks:    total,
				CurrentTaskID: task.TaskID,
			})
		}
	}
	return report, err
}

func (s *mockService) GetLatestComplexityReport() *taskmaster.ComplexityReport {
	return s.latestReport
}

func (s *mockService) ParsePRDWithProgress(ctx context.Context, inputPath string, mode taskmaster.ParsePrdMode, onProgress func(taskmaster.ParsePrdProgressState)) error {
	if onProgress != nil {
		onProgress(taskmaster.ParsePrdProgressState{Progress: 1.0, Label: "Parsed"})
	}
	return nil
}

func (s *mockService) ExpandTaskWithProgress(ctx context.Context, taskID string, opts taskmaster.ExpandTaskOptions, prompt string, onProgress func(taskmaster.ExpandProgressState)) error {
	if onProgress != nil {
		onProgress(taskmaster.ExpandProgressState{Stage: "Expanding task...", Progress: 0.5})
		onProgress(taskmaster.ExpandProgressState{Stage: "Expansion complete", Progress: 1.0})
	}
	return nil
}

func (s *mockService) ExportComplexityReport(ctx context.Context, format string, outputPath string) (string, error) {
	// Mock export, return a dummy file path
	return "/tmp/mock-export." + format, nil
}

func (s *mockService) LoadTasks(ctx context.Context) error {
	return nil
}

func (s *mockService) ReloadEvents() <-chan struct{} {
	return s.reloadCh
}

func (s *mockService) AnalyzeDeleteImpact(taskIDs []string, opts taskmaster.DeleteOptions) (*taskmaster.DeleteImpact, error) {
	return &taskmaster.DeleteImpact{}, nil
}

func (s *mockService) DeleteTasks(ctx context.Context, taskIDs []string, opts taskmaster.DeleteOptions) (*taskmaster.DeleteResult, error) {
	return &taskmaster.DeleteResult{DeletedCount: len(taskIDs)}, nil
}

func (s *mockService) UndoAction(ctx context.Context, actionID string) error {
	return nil
}

func (s *mockService) ListTagContexts(ctx context.Context, includeMetadata bool) (*taskmaster.TagList, error) {
	return &taskmaster.TagList{}, nil
}

func (s *mockService) AddTagContext(ctx context.Context, opts taskmaster.TagAddOptions) (*taskmaster.TagOperationResult, error) {
	return &taskmaster.TagOperationResult{}, nil
}

func (s *mockService) DeleteTagContext(ctx context.Context, name string, skipConfirmation bool) (*taskmaster.TagOperationResult, error) {
	return &taskmaster.TagOperationResult{}, nil
}

func (s *mockService) UseTagContext(ctx context.Context, name string) (*taskmaster.TagOperationResult, error) {
	return &taskmaster.TagOperationResult{}, nil
}

func (s *mockService) RenameTagContext(ctx context.Context, oldName, newName string) (*taskmaster.TagOperationResult, error) {
	return &taskmaster.TagOperationResult{}, nil
}

func (s *mockService) CopyTagContext(ctx context.Context, sourceName, targetName string, opts taskmaster.TagCopyOptions) (*taskmaster.TagOperationResult, error) {
	return &taskmaster.TagOperationResult{}, nil
}

func (s *mockService) ProjectRegistry() *projects.Registry {
	return nil
}

func (s *mockService) ActiveProjectMetadata() *projects.Metadata {
	return nil
}

func (s *mockService) SwitchProject(ctx context.Context, projectPath string) (*projects.Metadata, error) {
	return nil, nil
}

func (s *mockService) DiscoverProjects(ctx context.Context, roots []string) (int, error) {
	return 0, nil
}

// TestComplexityAnalysisWorkflow tests the entire complexity analysis workflow
func TestComplexityAnalysisWorkflow(t *testing.T) {
	// Set up the model with mock services
	mockService := mockTaskService()
	keyMap := NewKeyMap(nil)
	dm := dialog.InitializeDialogManager(800, 600, dialog.DefaultDialogStyle())
	model := Model{
		taskService: mockService,
		appState:    NewAppState(dm, &keyMap),
		keyMap:      keyMap,
	}

	// Test 1: Show scope selection dialog
	model.showComplexityScopeDialog()
	if manager := model.dialogManager(); manager == nil || !manager.HasDialogs() {
		t.Error("Expected scope selection dialog to be shown")
	}

	// Test 2: Handle scope selection message
	msg := ComplexityScopeSelectedMsg{
		Scope:  "all",
		TaskID: "",
		Tags:   []string{},
	}
	cmd := model.handleComplexityScopeSelected(msg)
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned from handleComplexityScopeSelected")
	}

	// Test 3: Handle analysis completion
	report, err := mockService.AnalyzeComplexity(context.Background(), "all", "", nil)
	if err != nil {
		t.Errorf("Expected no error from AnalyzeComplexity, got %v", err)
	}

	completedMsg := ComplexityAnalysisCompletedMsg{
		Report: report,
		Error:  nil,
	}
	cmd = model.handleComplexityAnalysisCompleted(completedMsg)
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned from handleComplexityAnalysisCompleted")
	}

	// Verify report dialog was created
	if manager := model.dialogManager(); manager == nil || !manager.HasDialogs() {
		t.Error("Expected report dialog to be shown")
	}

	// Test 4: Filter dialog
	actionMsg := ComplexityReportActionMsg{
		Action: "filter",
	}
	cmd = model.handleComplexityReportAction(actionMsg)
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned from handleComplexityReportAction")
	}

	// Test 5: Export dialog
	actionMsg = ComplexityReportActionMsg{
		Action: "export",
	}
	cmd = model.handleComplexityReportAction(actionMsg)
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned from handleComplexityReportAction")
	}
}

// TestComplexityScoringSorting tests sorting functionality for complexity results
func TestComplexityScoringSorting(t *testing.T) {
	// Create a sample report with known values
	complexities := []taskmaster.TaskComplexity{
		{
			TaskID: "1",
			Level:  taskmaster.ComplexityLow,
			Score:  2,
			Title:  "Task 1",
		},
		{
			TaskID: "3",
			Level:  taskmaster.ComplexityHigh,
			Score:  10,
			Title:  "Task 3",
		},
		{
			TaskID: "2",
			Level:  taskmaster.ComplexityMedium,
			Score:  5,
			Title:  "Task 2",
		},
	}

	report := taskmaster.NewComplexityReport(complexities, "all", nil)

	// Create the dialog
	reportDialog := dialog.NewComplexityReportDialog(report, nil)

	// Test SortByTaskID
	reportDialog.SortOrder = dialog.SortByTaskID
	reportDialog.ApplyFiltersAndSortForTest()
	if reportDialog.FilteredTasks[0].TaskID != "1" ||
		reportDialog.FilteredTasks[1].TaskID != "2" ||
		reportDialog.FilteredTasks[2].TaskID != "3" {
		t.Error("SortByTaskID failed to sort correctly")
	}

	// Test SortByScoreAsc
	reportDialog.SortOrder = dialog.SortByScoreAsc
	reportDialog.ApplyFiltersAndSortForTest()
	if reportDialog.FilteredTasks[0].Score != 2 ||
		reportDialog.FilteredTasks[1].Score != 5 ||
		reportDialog.FilteredTasks[2].Score != 10 {
		t.Error("SortByScoreAsc failed to sort correctly")
	}

	// Test SortByScoreDesc
	reportDialog.SortOrder = dialog.SortByScoreDesc
	reportDialog.ApplyFiltersAndSortForTest()
	if reportDialog.FilteredTasks[0].Score != 10 ||
		reportDialog.FilteredTasks[1].Score != 5 ||
		reportDialog.FilteredTasks[2].Score != 2 {
		t.Error("SortByScoreDesc failed to sort correctly")
	}
}

// TestComplexityFiltering tests filtering functionality for complexity results
func TestComplexityFiltering(t *testing.T) {
	// Create a sample report with known values
	complexities := []taskmaster.TaskComplexity{
		{
			TaskID: "1",
			Level:  taskmaster.ComplexityLow,
			Score:  2,
			Title:  "Task 1",
		},
		{
			TaskID: "2",
			Level:  taskmaster.ComplexityMedium,
			Score:  5,
			Title:  "Task 2",
		},
		{
			TaskID: "3",
			Level:  taskmaster.ComplexityHigh,
			Score:  10,
			Title:  "Task 3",
		},
		{
			TaskID: "4",
			Level:  taskmaster.ComplexityVeryHigh,
			Score:  15,
			Title:  "Task 4",
		},
	}

	report := taskmaster.NewComplexityReport(complexities, "all", nil)

	// Create the dialog
	reportDialog := dialog.NewComplexityReportDialog(report, nil)

	// Test default filters (all levels enabled)
	if len(reportDialog.FilteredTasks) != 4 {
		t.Error("Default filter should show all tasks")
	}

	// Test filtering by just low complexity
	reportDialog.FilterSettings.Levels = map[taskmaster.ComplexityLevel]bool{
		taskmaster.ComplexityLow:      true,
		taskmaster.ComplexityMedium:   false,
		taskmaster.ComplexityHigh:     false,
		taskmaster.ComplexityVeryHigh: false,
	}
	reportDialog.ApplyFiltersAndSortForTest()
	if len(reportDialog.FilteredTasks) != 1 || reportDialog.FilteredTasks[0].TaskID != "1" {
		t.Error("Filter by Low complexity failed")
	}

	// Test filtering by high and very high complexity
	reportDialog.FilterSettings.Levels = map[taskmaster.ComplexityLevel]bool{
		taskmaster.ComplexityLow:      false,
		taskmaster.ComplexityMedium:   false,
		taskmaster.ComplexityHigh:     true,
		taskmaster.ComplexityVeryHigh: true,
	}
	reportDialog.ApplyFiltersAndSortForTest()
	if len(reportDialog.FilteredTasks) != 2 ||
		(reportDialog.FilteredTasks[0].TaskID != "3" && reportDialog.FilteredTasks[1].TaskID != "4") {
		t.Error("Filter by High and Very High complexity failed")
	}
}

// TestUpdateMethodsForComplexity tests the model's update methods for complexity
func TestUpdateMethodsForComplexity(t *testing.T) {
	// Set up a model with mock service
	mockService := mockTaskService()
	keyMap := NewKeyMap(nil)
	dm := dialog.InitializeDialogManager(800, 600, dialog.DefaultDialogStyle())
	model := Model{
		taskService: mockService,
		appState:    NewAppState(dm, &keyMap),
		keyMap:      keyMap,
		width:       800,
		height:      600,
	}

	// Create a test message
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}, Alt: true}

	// Test keymap for Alt+C
	model.keyMap = keyMap

	// Check if Alt+C is bound to AnalyzeComplexity
	matched := key.Matches(keyMsg, model.keyMap.AnalyzeComplexity)
	if !matched {
		t.Error("Expected Alt+C to match AnalyzeComplexity keybinding")
	}
}
