package taskmaster

import (
	"testing"
	"time"
)

func TestCalculateComplexityScore(t *testing.T) {
	// Mock the current time function for testing
	mockTime := time.Date(2025, 12, 9, 12, 0, 0, 0, time.UTC)
	originalNow := Now
	Now = func() time.Time {
		return mockTime
	}
	defer func() { Now = originalNow }() // Restore the original function

	// Override complexity keywords for test predictability
	originalKeywords := ComplexityKeywords
	ComplexityKeywords = []string{"integration", "migration", "refactor", "complex", "algorithm", "performance"}
	defer func() { ComplexityKeywords = originalKeywords }() // Restore the original keywords

	tests := []struct {
		name             string
		task             Task
		expectedLevel    ComplexityLevel
		expectedScore    int
		customWeights    *ScoringWeights
		customThresholds *LevelThresholds
	}{
		{
			name: "Simple task - Low complexity",
			task: Task{
				ID:           "1",
				Title:        "Simple task",
				Description:  "This is a simple task with minimal description.",
				Priority:     PriorityLow,
				Dependencies: []string{},
				Subtasks:     []Task{},
			},
			expectedLevel: ComplexityLow,
			expectedScore: 0,
		},
		{
			name: "Medium complexity task",
			task: Task{
				ID:           "2",
				Title:        "Medium complexity task",
				Description:  "This task requires integration with an external system and has several steps to complete. The implementation will need careful planning.",
				Priority:     PriorityMedium,
				Dependencies: []string{"1", "3"},
				Subtasks:     []Task{{ID: "2.1"}, {ID: "2.2"}},
			},
			expectedLevel: ComplexityMedium,
			expectedScore: 6,
		},
		{
			name: "High complexity task",
			task: Task{
				ID:             "3",
				Title:          "Complex refactoring task",
				Description:    "This task involves a major refactoring of the core system. It will require careful planning and execution to ensure compatibility.",
				Details:        "The refactoring will touch multiple components and will need to maintain backward compatibility. Performance optimization and security considerations are important.",
				Priority:       PriorityHigh,
				Dependencies:   []string{"1", "2", "4", "5"},
				Subtasks:       []Task{{ID: "3.1"}, {ID: "3.2"}, {ID: "3.3"}},
				EstimatedHours: 24,
			},
			expectedLevel: ComplexityHigh,
			expectedScore: 12,
		},
		{
			name: "Very high complexity task",
			task: Task{
				ID:             "4",
				Title:          "System-wide performance optimization",
				Description:    "This task involves complex algorithm optimization and distributed system performance improvements. It requires deep understanding of the system architecture and dependencies.",
				Details:        "The optimization will focus on concurrent processing, distributed computing, and real-time data flow. Security and scalability are critical considerations.",
				Priority:       PriorityCritical,
				Dependencies:   []string{"1", "2", "3", "5", "6"},
				Subtasks:       []Task{{ID: "4.1"}, {ID: "4.2"}, {ID: "4.3"}, {ID: "4.4"}},
				EstimatedHours: 40,
			},
			expectedLevel: ComplexityVeryHigh,
			expectedScore: 26,
		},
		{
			name: "Custom weights test",
			task: Task{
				ID:           "5",
				Title:        "Task with custom weights",
				Description:  "This is a test for custom weights.",
				Dependencies: []string{"1", "2"},
				Subtasks:     []Task{{ID: "5.1"}, {ID: "5.2"}},
			},
			customWeights: &ScoringWeights{
				DescriptionLength:  0.5,  // Lower weight for description
				SubtaskCount:       3.0,  // Higher weight for subtasks
				DependencyCount:    4.0,  // Higher weight for dependencies
				ComplexityKeywords: 1.0,  // Lower weight for keywords
				EstimatedHours:     0.25, // Lower weight for hours
				PriorityFactor:     1.0,  // No priority factor
			},
			expectedLevel: ComplexityHigh,
			expectedScore: 10,
		},
		{
			name: "Custom thresholds test",
			task: Task{
				ID:           "6",
				Title:        "Task with custom thresholds",
				Description:  "This is a test for custom thresholds.",
				Dependencies: []string{"1", "2"},
				Subtasks:     []Task{{ID: "6.1"}, {ID: "6.2"}},
			},
			customThresholds: &LevelThresholds{
				Low:      2, // 0-2: Low
				Medium:   5, // 3-5: Medium
				High:     8, // 6-8: High
				VeryHigh: 8, // 9+: Very High
			},
			expectedLevel: ComplexityHigh,
			expectedScore: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateComplexityScore(&tt.task, tt.customWeights, tt.customThresholds)

			if result.Level != tt.expectedLevel {
				t.Errorf("Expected complexity level %s, got %s", tt.expectedLevel, result.Level)
			}

			if result.Score != tt.expectedScore {
				t.Errorf("Expected complexity score %d, got %d", tt.expectedScore, result.Score)
			}

			if result.TaskID != tt.task.ID {
				t.Errorf("Expected task ID %s, got %s", tt.task.ID, result.TaskID)
			}

			if result.Title != tt.task.Title {
				t.Errorf("Expected task title %s, got %s", tt.task.Title, result.Title)
			}

			if !result.AnalyzedAt.Equal(mockTime) {
				t.Errorf("Expected analyzed time %v, got %v", mockTime, result.AnalyzedAt)
			}
		})
	}
}

func TestRecalculateComplexityThresholds(t *testing.T) {
	tests := []struct {
		name         string
		complexities []TaskComplexity
		expected     LevelThresholds
	}{
		{
			name:         "Empty complexities",
			complexities: []TaskComplexity{},
			expected:     DefaultLevelThresholds(),
		},
		{
			name: "Too few tasks",
			complexities: []TaskComplexity{
				{Score: 5},
				{Score: 10},
			},
			expected: DefaultLevelThresholds(),
		},
		{
			name: "Normal distribution",
			complexities: []TaskComplexity{
				{Score: 1},
				{Score: 2},
				{Score: 3},
				{Score: 4},
				{Score: 5},
				{Score: 6},
				{Score: 7},
				{Score: 8},
				{Score: 9},
				{Score: 10},
				{Score: 11},
				{Score: 12},
			},
			expected: LevelThresholds{
				Low:      3,
				Medium:   6,
				High:     9,
				VeryHigh: 9,
			},
		},
		{
			name: "Skewed distribution",
			complexities: []TaskComplexity{
				{Score: 1},
				{Score: 2},
				{Score: 2},
				{Score: 3},
				{Score: 4},
				{Score: 5},
				{Score: 6},
				{Score: 15},
				{Score: 18},
				{Score: 20},
				{Score: 22},
				{Score: 25},
			},
			expected: LevelThresholds{
				Low:      3,
				Medium:   5,
				High:     18,
				VeryHigh: 18,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RecalculateComplexityThresholds(tt.complexities)

			if result.Low != tt.expected.Low {
				t.Errorf("Expected Low threshold %d, got %d", tt.expected.Low, result.Low)
			}

			if result.Medium != tt.expected.Medium {
				t.Errorf("Expected Medium threshold %d, got %d", tt.expected.Medium, result.Medium)
			}

			if result.High != tt.expected.High {
				t.Errorf("Expected High threshold %d, got %d", tt.expected.High, result.High)
			}

			if result.VeryHigh != tt.expected.VeryHigh {
				t.Errorf("Expected VeryHigh threshold %d, got %d", tt.expected.VeryHigh, result.VeryHigh)
			}
		})
	}
}

func TestGetLevelFromScore(t *testing.T) {
	tests := []struct {
		name       string
		score      int
		thresholds *LevelThresholds
		expected   ComplexityLevel
	}{
		{
			name:       "Low with default thresholds",
			score:      2,
			thresholds: nil,
			expected:   ComplexityLow,
		},
		{
			name:       "Medium with default thresholds",
			score:      5,
			thresholds: nil,
			expected:   ComplexityMedium,
		},
		{
			name:       "High with default thresholds",
			score:      10,
			thresholds: nil,
			expected:   ComplexityHigh,
		},
		{
			name:       "Very High with default thresholds",
			score:      15,
			thresholds: nil,
			expected:   ComplexityVeryHigh,
		},
		{
			name:       "Custom thresholds test - Low",
			score:      1,
			thresholds: &LevelThresholds{Low: 2, Medium: 5, High: 10, VeryHigh: 10},
			expected:   ComplexityLow,
		},
		{
			name:       "Custom thresholds test - Medium",
			score:      4,
			thresholds: &LevelThresholds{Low: 2, Medium: 5, High: 10, VeryHigh: 10},
			expected:   ComplexityMedium,
		},
		{
			name:       "Custom thresholds test - High",
			score:      7,
			thresholds: &LevelThresholds{Low: 2, Medium: 5, High: 10, VeryHigh: 10},
			expected:   ComplexityHigh,
		},
		{
			name:       "Custom thresholds test - Very High",
			score:      12,
			thresholds: &LevelThresholds{Low: 2, Medium: 5, High: 10, VeryHigh: 10},
			expected:   ComplexityVeryHigh,
		},
		{
			name:       "Boundary test - Low/Medium",
			score:      3,
			thresholds: &LevelThresholds{Low: 3, Medium: 7, High: 12, VeryHigh: 12},
			expected:   ComplexityLow,
		},
		{
			name:       "Boundary test - Medium/High",
			score:      7,
			thresholds: &LevelThresholds{Low: 3, Medium: 7, High: 12, VeryHigh: 12},
			expected:   ComplexityMedium,
		},
		{
			name:       "Boundary test - High/Very High",
			score:      12,
			thresholds: &LevelThresholds{Low: 3, Medium: 7, High: 12, VeryHigh: 12},
			expected:   ComplexityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevelFromScore(tt.score, tt.thresholds)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSortScores(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse sorted",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Mixed order",
			input:    []int{3, 1, 5, 2, 4},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Duplicates",
			input:    []int{3, 1, 3, 2, 1},
			expected: []int{1, 1, 2, 3, 3},
		},
		{
			name:     "Single element",
			input:    []int{1},
			expected: []int{1},
		},
		{
			name:     "Empty array",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortScores(tt.input)

			// Check if arrays have same length
			if len(tt.input) != len(tt.expected) {
				t.Fatalf("Expected array length %d, got %d", len(tt.expected), len(tt.input))
			}

			// Check if arrays have same elements in same order
			for i := 0; i < len(tt.input); i++ {
				if tt.input[i] != tt.expected[i] {
					t.Errorf("At index %d, expected %d, got %d", i, tt.expected[i], tt.input[i])
				}
			}
		})
	}
}
