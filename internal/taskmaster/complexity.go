package taskmaster

import (
	"fmt"
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
