package taskmaster

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ComplexityLevel represents the complexity level of a task
type ComplexityLevel string

const (
	ComplexityLow      ComplexityLevel = "low"
	ComplexityMedium   ComplexityLevel = "medium"
	ComplexityHigh     ComplexityLevel = "high"
	ComplexityVeryHigh ComplexityLevel = "veryhigh"
)

// TaskComplexity represents the analyzed complexity of a task
type TaskComplexity struct {
	TaskID      string          `json:"taskId"`
	Level       ComplexityLevel `json:"level"`
	Score       int             `json:"score"`
	Title       string          `json:"title"`                 // Stored for display purposes
	Description string          `json:"description,omitempty"` // Optional detailed description
	AnalyzedAt  time.Time       `json:"analyzedAt"`
}

// ComplexityReport represents a collection of task complexity analyses
type ComplexityReport struct {
	Tasks        []TaskComplexity `json:"tasks"`
	AnalyzedAt   time.Time        `json:"analyzedAt"`
	Scope        string           `json:"scope"` // "all", "selected", "tag:X"
	FilteredTags []string         `json:"filteredTags,omitempty"`
}

// GetColorForLevel returns the appropriate color identifier for the complexity level
func (level ComplexityLevel) GetColorForLevel() string {
	switch level {
	case ComplexityLow:
		return "green"
	case ComplexityMedium:
		return "yellow"
	case ComplexityHigh:
		return "orange"
	case ComplexityVeryHigh:
		return "red"
	default:
		return "default"
	}
}

// GetLevelFromScore determines the complexity level based on a numeric score and thresholds
func GetLevelFromScore(score int, thresholds *LevelThresholds) ComplexityLevel {
	// Use default thresholds if not provided
	t := DefaultLevelThresholds()
	if thresholds != nil {
		t = *thresholds
	}

	switch {
	case score <= t.Low:
		return ComplexityLow
	case score <= t.Medium:
		return ComplexityMedium
	case score <= t.High:
		return ComplexityHigh
	default:
		return ComplexityVeryHigh
	}
}

// AnalyzeComplexity calculates complexity scores for a set of tasks
func AnalyzeComplexity(tasks []*Task) []TaskComplexity {
	result := make([]TaskComplexity, 0, len(tasks))

	// Use default weights and thresholds
	weights := DefaultScoringWeights()
	thresholds := DefaultLevelThresholds()

	for _, task := range tasks {
		complexity := CalculateComplexityScore(task, &weights, &thresholds)
		result = append(result, complexity)
	}

	return result
}

// NewComplexityReport creates a new complexity report for the given tasks and scope
func NewComplexityReport(taskComplexities []TaskComplexity, scope string, tags []string) *ComplexityReport {
	return &ComplexityReport{
		Tasks:        taskComplexities,
		AnalyzedAt:   time.Now(),
		Scope:        scope,
		FilteredTags: tags,
	}
}

// String returns a string representation of the complexity level
func (level ComplexityLevel) String() string {
	switch level {
	case ComplexityLow:
		return "Low"
	case ComplexityMedium:
		return "Medium"
	case ComplexityHigh:
		return "High"
	case ComplexityVeryHigh:
		return "Very High"
	default:
		return "Unknown"
	}
}

// String returns a formatted summary of the task complexity
func (tc TaskComplexity) String() string {
	return fmt.Sprintf("Task %s: %s (Score: %d)", tc.TaskID, tc.Level, tc.Score)
}

// GetSummary returns a summary of the complexity report
func (report *ComplexityReport) GetSummary() string {
	counts := make(map[ComplexityLevel]int)

	for _, task := range report.Tasks {
		counts[task.Level]++
	}

	return fmt.Sprintf("Analyzed %d tasks: %d Low, %d Medium, %d High, %d Very High",
		len(report.Tasks),
		counts[ComplexityLow],
		counts[ComplexityMedium],
		counts[ComplexityHigh],
		counts[ComplexityVeryHigh],
	)
}

// FilterByLevel filters the report to only include tasks of the specified complexity levels
func (report *ComplexityReport) FilterByLevel(levels []ComplexityLevel) *ComplexityReport {
	if len(levels) == 0 {
		return report
	}

	// Create a map for quick lookup
	levelMap := make(map[ComplexityLevel]bool)
	for _, level := range levels {
		levelMap[level] = true
	}

	filteredTasks := make([]TaskComplexity, 0)
	for _, task := range report.Tasks {
		if levelMap[task.Level] {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Create a new report with filtered tasks
	filtered := &ComplexityReport{
		Tasks:        filteredTasks,
		AnalyzedAt:   report.AnalyzedAt,
		Scope:        report.Scope,
		FilteredTags: report.FilteredTags,
	}

	return filtered
}

// BuildAnalyzeComplexityArgs constructs CLI arguments for task-master analyze-complexity command
func BuildAnalyzeComplexityArgs(scope string, taskID string, tags []string) []string {
	args := []string{"analyze-complexity"}
	
	if scope != "" {
		args = append(args, "--scope", scope)
	}
	
	// Add task ID if analyzing a selected task
	if taskID != "" && scope == "selected" {
		args = append(args, "--task-id", taskID)
	}
	
	// Add tags if analyzing by tag
	if len(tags) > 0 && scope == "tag" {
		tagsStr := strings.Join(tags, ",")
		args = append(args, "--tags", tagsStr)
	}
	
	return args
}

// ParseComplexityProgress parses a single line of complexity analysis output
// and extracts progress information
func ParseComplexityProgress(line string) ComplexityProgressState {
	state := ComplexityProgressState{}
	
	line = strings.TrimSpace(line)
	
	// Filter out empty lines and noise
	if line == "" ||
		strings.Contains(line, "/.taskmaster/") ||
		strings.HasPrefix(line, "/Users/") ||
		strings.HasPrefix(line, "/home/") ||
		len(line) > 200 {
		return state
	}
	
	// Parse "Analyzing task X (Y/Z)..." pattern
	// Example: "Analyzing task 1.2 (5/47)..."
	re := regexp.MustCompile(`Analyzing task (\S+)\s+\((\d+)/(\d+)\)`)
	if matches := re.FindStringSubmatch(line); len(matches) > 3 {
		taskID := matches[1]
		analyzed, _ := strconv.Atoi(matches[2])
		total, _ := strconv.Atoi(matches[3])
		
		state.CurrentTaskID = taskID
		state.TasksAnalyzed = analyzed
		state.TotalTasks = total
		return state
	}
	
	// Parse "Progress: X/Y" pattern
	re = regexp.MustCompile(`Progress:\s*(\d+)/(\d+)`)
	if matches := re.FindStringSubmatch(line); len(matches) > 2 {
		analyzed, _ := strconv.Atoi(matches[1])
		total, _ := strconv.Atoi(matches[2])
		
		state.TasksAnalyzed = analyzed
		state.TotalTasks = total
		return state
	}
	
	return state
}

// ParseComplexityReportJSON parses JSON output from the CLI into a ComplexityReport
func ParseComplexityReportJSON(jsonStr string) (*ComplexityReport, error) {
	var report ComplexityReport
	
	// Trim any whitespace and quotes
	jsonStr = strings.TrimSpace(jsonStr)
	if strings.HasPrefix(jsonStr, "\"") && strings.HasSuffix(jsonStr, "\"") {
		jsonStr = strings.Trim(jsonStr, "\"")
		// Unescape JSON string escapes
		jsonStr = strings.ReplaceAll(jsonStr, "\\\"", "\"")
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		return nil, fmt.Errorf("failed to parse complexity report JSON: %w", err)
	}
	
	return &report, nil
}
