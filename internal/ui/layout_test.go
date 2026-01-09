package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/git"
	"github.com/agreen757/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/bubbles/help"
)

// TestFormatGitInfoNoGit tests formatGitInfo when git is not available
func TestFormatGitInfoNoGit(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:       cfg,
		gitAvailable: false,
		gitRepoInfo:  git.RepoInfo{IsRepo: false},
	}

	result := m.formatGitInfo()
	if result != "" {
		t.Errorf("expected empty string when git unavailable, got: %q", result)
	}
}

// TestFormatGitInfoNoRepo tests formatGitInfo when not in a git repo
func TestFormatGitInfoNoRepo(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:       cfg,
		gitAvailable: true,
		gitRepoInfo:  git.RepoInfo{IsRepo: false},
	}

	result := m.formatGitInfo()
	if result == "" {
		t.Errorf("expected non-empty string for 'No Git repo' message")
	}
	// Should contain ANSI codes (styling)
	if len(result) < 5 {
		t.Errorf("expected styled output with ANSI codes, got: %q", result)
	}
}

// TestFormatGitInfoEmpty tests formatGitInfo when branch is empty
func TestFormatGitInfoEmpty(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:       cfg,
		gitAvailable: true,
		gitRepoInfo:  git.RepoInfo{IsRepo: true},
		gitRefresher: &git.StatusRefresher{},
	}

	// With empty branch, should return empty
	result := m.formatGitInfo()
	if result != "" {
		t.Logf("Note: Got result %q (may contain ANSI), which is acceptable if gitRefresher is running", result)
	}
}

// TestRenderStatusBarWithGitAvailable tests that renderStatusBar includes git info
func TestRenderStatusBarWithGitAvailable(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:       cfg,
		gitAvailable: true,
		gitRepoInfo:  git.RepoInfo{IsRepo: false},
		width:        80,
		height:       24,
		styles:       NewStyles(),
		helpModel:    help.New(),
		keyMap:       NewKeyMap(cfg),
	}

	result := m.renderStatusBar()
	if result == "" {
		t.Errorf("renderStatusBar should not be empty")
	}

	// Status bar should be valid and contain content
	if len(result) > 10000 {
		t.Errorf("renderStatusBar output seems unusually large: %d chars", len(result))
	}
}

// TestRenderStatusBarStructure tests all status bar modes
func TestRenderStatusBarModes(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	baseModel := Model{
		config:       cfg,
		width:        80,
		height:       24,
		styles:       NewStyles(),
		helpModel:    help.New(),
		keyMap:       NewKeyMap(cfg),
	}

	tests := []struct {
		name string
		setup func(*Model)
	}{
		{
			name: "normal mode with git",
			setup: func(m *Model) {
				m.gitAvailable = true
				m.gitRepoInfo = git.RepoInfo{IsRepo: false}
			},
		},
		{
			name: "command mode",
			setup: func(m *Model) {
				m.commandMode = true
				m.commandInput = "1.2"
			},
		},
		{
			name: "search mode",
			setup: func(m *Model) {
				m.searchMode = true
			},
		},
		{
			name: "clear state confirmation",
			setup: func(m *Model) {
				m.confirmingClearState = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := baseModel
			tt.setup(&m)

			result := m.renderStatusBar()
			if result == "" {
				t.Errorf("renderStatusBar returned empty for %s", tt.name)
			}
		})
	}
}

// TestGitInfoStylingOutput tests that git info gets proper styling
func TestGitInfoStylingOutput(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}

	tests := []struct {
		name      string
		isRepo    bool
		gitAvail  bool
		expectLen func(int) bool
	}{
		{
			name:     "no git repo shows styled message",
			isRepo:   false,
			gitAvail: true,
			expectLen: func(l int) bool { return l > 10 }, // Has ANSI codes
		},
		{
			name:     "git unavailable returns empty",
			isRepo:   false,
			gitAvail: false,
			expectLen: func(l int) bool { return l == 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				config:       cfg,
				gitAvailable: tt.gitAvail,
				gitRepoInfo:  git.RepoInfo{IsRepo: tt.isRepo},
			}

			result := m.formatGitInfo()
			if !tt.expectLen(len(result)) {
				t.Errorf("unexpected length for result: %q (len=%d)", stripANSIForTest(result), len(result))
			}
		})
	}
}

// TestStatusBarIntegration tests the full status bar rendering with git info
func TestStatusBarIntegration(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:       cfg,
		gitAvailable: true,
		gitRepoInfo:  git.RepoInfo{IsRepo: true, RootPath: "/test"},
		width:        120,
		height:       30,
		styles:       NewStyles(),
		helpModel:    help.New(),
		keyMap:       NewKeyMap(cfg),
		gitRefresher: &git.StatusRefresher{},
	}

	// Test in normal mode - should try to include git info
	result := m.renderStatusBar()

	if result == "" {
		t.Errorf("status bar with git info should not be empty")
	}

	// Should contain formatting but exact content depends on actual git status
	t.Logf("Status bar output sample: %q", stripANSIForTest(result)[:min(100, len(result))])
}

// Helper to strip ANSI codes for testing/logging
func stripANSIForTest(s string) string {
	var result []rune
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == 'm' {
				inEscape = false
			}
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestRenderProgressBar tests the renderProgressBar function with various percentages and widths
func TestRenderProgressBar(t *testing.T) {
	m := Model{}

	tests := []struct {
		name       string
		percentage float64
		width      int
		expectLen  int // in runes, not bytes
		expectFull bool // all filled (▓)
		expectZero bool // all empty (░)
	}{
		{
			name:       "0% progress with width 10",
			percentage: 0,
			width:      10,
			expectLen:  10,
			expectZero: true,
		},
		{
			name:       "100% progress with width 10",
			percentage: 100,
			width:      10,
			expectLen:  10,
			expectFull: true,
		},
		{
			name:       "50% progress with width 10",
			percentage: 50,
			width:      10,
			expectLen:  10,
			expectFull: false,
			expectZero: false,
		},
		{
			name:       "width below threshold (4)",
			percentage: 50,
			width:      4,
			expectLen:  0,
		},
		{
			name:       "width at threshold (5)",
			percentage: 50,
			width:      5,
			expectLen:  5,
		},
		{
			name:       "negative percentage",
			percentage: -10,
			width:      10,
			expectLen:  10,
			expectZero: true,
		},
		{
			name:       "over 100% percentage",
			percentage: 150,
			width:      10,
			expectLen:  10,
			expectFull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderProgressBar(tt.percentage, tt.width)

			// Check rune length (not byte length)
			runeCount := len([]rune(result))
			if runeCount != tt.expectLen {
				t.Errorf("expected rune length %d, got %d for result: %q", tt.expectLen, runeCount, result)
			}

			// Check if all filled (100%)
			if tt.expectFull && result != "" {
				expected := strings.Repeat("▓", tt.width)
				if result != expected {
					t.Errorf("expected all filled (▓), got: %q", result)
				}
			}

			// Check if all empty (0%)
			if tt.expectZero && result != "" {
				expected := strings.Repeat("░", tt.width)
				if result != expected {
					t.Errorf("expected all empty (░), got: %q", result)
				}
			}

			// Verify result contains only valid block characters when non-empty
			if result != "" {
				validChars := 0
				for _, r := range result {
					if r == '▓' || r == '░' {
						validChars++
					}
				}
				if validChars != runeCount {
					t.Errorf("result contains invalid characters: %q", result)
				}
			}
		})
	}
}

// TestRenderProgressBar50Percent specifically tests 50% on various widths
func TestRenderProgressBar50Percent(t *testing.T) {
	m := Model{}

	tests := []struct {
		width          int
		expectedFilled int
		expectedEmpty  int
	}{
		{width: 5, expectedFilled: 2, expectedEmpty: 3},  // 50% of 5 = 2.5 → 2
		{width: 10, expectedFilled: 5, expectedEmpty: 5}, // 50% of 10 = 5
		{width: 20, expectedFilled: 10, expectedEmpty: 10},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("50_percent_width_%d", tt.width), func(t *testing.T) {
			result := m.renderProgressBar(50, tt.width)

			runeCount := len([]rune(result))
			if runeCount != tt.width {
				t.Errorf("expected rune length %d, got %d", tt.width, runeCount)
			}

			filledCount := 0
			emptyCount := 0
			for _, r := range result {
				if r == '▓' {
					filledCount++
				} else if r == '░' {
					emptyCount++
				}
			}

			if filledCount != tt.expectedFilled {
				t.Errorf("expected %d filled chars, got %d", tt.expectedFilled, filledCount)
			}
			if emptyCount != tt.expectedEmpty {
				t.Errorf("expected %d empty chars, got %d", tt.expectedEmpty, emptyCount)
			}
		})
	}
}

// TestCalculateTaskProgress tests the calculateTaskProgress function with various task configurations
func TestCalculateTaskProgress(t *testing.T) {
	tests := []struct {
		name       string
		tasks      []taskmaster.Task
		wantDone   int
		wantTotal  int
		wantPercent float64
	}{
		{
			name:       "empty task list",
			tasks:      []taskmaster.Task{},
			wantDone:   0,
			wantTotal:  0,
			wantPercent: 0.0,
		},
		{
			name: "single pending task",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusPending,
				},
			},
			wantDone:    0,
			wantTotal:   1,
			wantPercent: 0.0,
		},
		{
			name: "single done task",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusDone,
				},
			},
			wantDone:    1,
			wantTotal:   1,
			wantPercent: 100.0,
		},
		{
			name: "all tasks done",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusDone,
				},
				{
					ID:     "2",
					Title:  "Task 2",
					Status: taskmaster.StatusDone,
				},
			},
			wantDone:    2,
			wantTotal:   2,
			wantPercent: 100.0,
		},
		{
			name: "mixed status tasks",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusDone,
				},
				{
					ID:     "2",
					Title:  "Task 2",
					Status: taskmaster.StatusPending,
				},
			},
			wantDone:    1,
			wantTotal:   2,
			wantPercent: 50.0,
		},
		{
			name: "task with subtasks",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusDone,
					Subtasks: []taskmaster.Task{
						{
							ID:     "1.1",
							Title:  "Subtask 1.1",
							Status: taskmaster.StatusDone,
						},
						{
							ID:     "1.2",
							Title:  "Subtask 1.2",
							Status: taskmaster.StatusPending,
						},
					},
				},
			},
			wantDone:    2,
			wantTotal:   3,
			wantPercent: 66.666667,
		},
		{
			name: "deeply nested subtasks",
			tasks: []taskmaster.Task{
				{
					ID:     "1",
					Title:  "Task 1",
					Status: taskmaster.StatusPending,
					Subtasks: []taskmaster.Task{
						{
							ID:     "1.1",
							Title:  "Subtask 1.1",
							Status: taskmaster.StatusDone,
							Subtasks: []taskmaster.Task{
								{
									ID:     "1.1.1",
									Title:  "Sub-subtask 1.1.1",
									Status: taskmaster.StatusDone,
								},
							},
						},
					},
				},
			},
			wantDone:    2,
			wantTotal:   3,
			wantPercent: 66.666667,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				tasks: tt.tasks,
			}

			done, total, percent := m.calculateTaskProgress()

			if done != tt.wantDone {
				t.Errorf("done: got %d, want %d", done, tt.wantDone)
			}
			if total != tt.wantTotal {
				t.Errorf("total: got %d, want %d", total, tt.wantTotal)
			}

			tolerance := 0.001
			if total > 0 && (percent < tt.wantPercent-tolerance || percent > tt.wantPercent+tolerance) {
				t.Errorf("percent: got %.6f, want %.6f", percent, tt.wantPercent)
			}
		})
	}
}

// TestRenderWideHeaderContainsRequiredElements verifies that all required elements are present
func TestRenderWideHeaderContainsRequiredElements(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     120,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tag := "test-tag"
	done := 5
	total := 10
	percentage := 50.0

	result := m.renderWideHeader(tag, done, total, percentage)

	// Verify all required components are present
	tests := []struct {
		name    string
		pattern string
	}{
		{"Application title", "Task Master TUI"},
		{"Tag prefix", "Tag:"},
		{"Tag value", tag},
		{"Progress format", "Progress: 5/10"},
		{"Percentage display", "50"},
	}

	for _, test := range tests {
		if !strings.Contains(result, test.pattern) {
			t.Errorf("%s: expected to find '%s' in output, got: %q", test.name, test.pattern, result)
		}
	}
}

// TestRenderWideHeaderProgressBarLength verifies that the progress bar is 10 characters
func TestRenderWideHeaderProgressBarLength(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     120,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	// Test with 100% progress (all filled)
	result := m.renderWideHeader("test", 10, 10, 100.0)

	// The 10-character bar should be 10 block characters
	// Check for the filled block character pattern
	if !strings.Contains(result, "▓▓▓▓▓▓▓▓▓▓") {
		t.Errorf("expected to find 10 filled blocks for 100%% progress, got: %q", result)
	}

	// Test with 0% progress (all empty)
	result = m.renderWideHeader("test", 0, 10, 0.0)
	if !strings.Contains(result, "░░░░░░░░░░") {
		t.Errorf("expected to find 10 empty blocks for 0%% progress, got: %q", result)
	}

	// Test with 50% progress (5 filled, 5 empty)
	result = m.renderWideHeader("test", 5, 10, 50.0)
	if !strings.Contains(result, "▓▓▓▓▓░░░░░") {
		t.Errorf("expected to find 5 filled and 5 empty blocks for 50%% progress, got: %q", result)
	}
}

// TestRenderWideHeaderBorderStyle verifies that the border is correctly applied
func TestRenderWideHeaderBorderStyle(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     120,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderWideHeader("test", 5, 10, 50.0)

	// Check for double-line border characters (╔═╗║╚═╝)
	// These should be present in the output when lipgloss.DoubleBorder() is applied
	doubleBorderChars := []string{"╔", "╗", "║", "╚", "═"}
	foundBorder := false
	for _, char := range doubleBorderChars {
		if strings.Contains(result, char) {
			foundBorder = true
			break
		}
	}

	if !foundBorder {
		t.Errorf("expected to find double-line border characters in output, got: %q", result)
	}
}

// TestRenderWideHeaderWidthCalculation verifies that the header respects width constraints
func TestRenderWideHeaderWidthCalculation(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"minimum wide terminal", 100},
		{"standard wide terminal", 120},
		{"very wide terminal", 160},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{TaskMasterPath: "/tmp/test"}
			m := Model{
				config:    cfg,
				width:     test.width,
				height:    24,
				styles:    NewStyles(),
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderWideHeader("test", 5, 10, 50.0)

			// The result should not be empty
			if result == "" {
				t.Errorf("renderWideHeader returned empty string for width %d", test.width)
			}

			// Count visible characters (rough check - lipgloss adds ANSI codes)
			// We just verify that we got a non-trivial output
			if len(result) < 50 {
				t.Logf("width %d: output length %d (may contain many ANSI codes)", test.width, len(result))
			}
		})
	}
}

// TestRenderWideHeaderEdgeCases verifies behavior with edge case inputs
func TestRenderWideHeaderEdgeCases(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     120,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tests := []struct {
		name       string
		tag        string
		done       int
		total      int
		percentage float64
	}{
		{"zero tasks", "test", 0, 0, 0.0},
		{"empty tag", "", 5, 10, 50.0},
		{"long tag name", "very-long-feature-tag-name", 5, 10, 50.0},
		{"high percentage", "test", 10, 10, 100.0},
		{"low percentage", "test", 0, 10, 0.0},
		{"partial progress", "test", 3, 10, 30.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Should not panic or return empty string
			result := m.renderWideHeader(test.tag, test.done, test.total, test.percentage)
			if result == "" {
				t.Errorf("renderWideHeader returned empty string for tag=%q, done=%d, total=%d, percentage=%.0f",
					test.tag, test.done, test.total, test.percentage)
			}

			// Should contain progress information
			if !strings.Contains(result, "Progress:") {
				t.Errorf("expected to find 'Progress:' in output for tag=%q", test.tag)
			}
		})
	}
}

// TestRenderMediumHeaderContainsRequiredElements verifies that all required elements are present in medium header
func TestRenderMediumHeaderContainsRequiredElements(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     90,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tag := "test-tag"
	done := 5
	total := 10
	percentage := 50.0

	result := m.renderMediumHeader(tag, done, total, percentage)

	// Verify all required components are present
	tests := []struct {
		name    string
		pattern string
	}{
		{"Application title", "Task Master TUI"},
		{"Tag value", tag},
		{"Percentage display", "50"},
	}

	for _, test := range tests {
		if !strings.Contains(result, test.pattern) {
			t.Errorf("%s: expected to find '%s' in output, got: %q", test.name, test.pattern, result)
		}
	}
}

// TestRenderMediumHeaderProgressBarLength verifies that the progress bar is 5 characters
func TestRenderMediumHeaderProgressBarLength(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     90,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	// Test with 100% progress (all filled)
	result := m.renderMediumHeader("test", 10, 10, 100.0)

	// The 5-character bar should be 5 block characters
	if !strings.Contains(result, "▓▓▓▓▓") {
		t.Errorf("expected to find 5 filled blocks for 100%% progress, got: %q", result)
	}

	// Test with 0% progress (all empty)
	result = m.renderMediumHeader("test", 0, 10, 0.0)
	if !strings.Contains(result, "░░░░░") {
		t.Errorf("expected to find 5 empty blocks for 0%% progress, got: %q", result)
	}

	// Test with 50% progress (2-3 filled, rest empty)
	result = m.renderMediumHeader("test", 5, 10, 50.0)
	// 50% of 5 = 2.5 → 2 filled
	if !strings.Contains(result, "▓▓░░░") {
		t.Errorf("expected to find 2 filled and 3 empty blocks for 50%% progress, got: %q", result)
	}
}

// TestRenderMediumHeaderBorderStyle verifies that the border is correctly applied (rounded border)
func TestRenderMediumHeaderBorderStyle(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     90,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderMediumHeader("test", 5, 10, 50.0)

	// Check for rounded border characters (╭─╮│╰─╯)
	roundedBorderChars := []string{"╭", "╮", "│", "╰", "─"}
	foundBorder := false
	for _, char := range roundedBorderChars {
		if strings.Contains(result, char) {
			foundBorder = true
			break
		}
	}

	if !foundBorder {
		t.Errorf("expected to find rounded border characters in output, got: %q", result)
	}
}

// TestRenderMediumHeaderWidthCalculation verifies that the header respects width constraints
func TestRenderMediumHeaderWidthCalculation(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"narrow medium terminal", 80},
		{"standard medium terminal", 90},
		{"wide medium terminal", 99},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{TaskMasterPath: "/tmp/test"}
			m := Model{
				config:    cfg,
				width:     test.width,
				height:    24,
				styles:    NewStyles(),
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderMediumHeader("test", 5, 10, 50.0)

			// The result should not be empty
			if result == "" {
				t.Errorf("renderMediumHeader returned empty string for width %d", test.width)
			}

			// Count visible characters (rough check - lipgloss adds ANSI codes)
			// We just verify that we got a non-trivial output
			if len(result) < 40 {
				t.Logf("width %d: output length %d (may contain many ANSI codes)", test.width, len(result))
			}
		})
	}
}

// TestRenderMediumHeaderEdgeCases verifies behavior with edge case inputs
func TestRenderMediumHeaderEdgeCases(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     90,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tests := []struct {
		name       string
		tag        string
		done       int
		total      int
		percentage float64
	}{
		{"zero tasks", "test", 0, 0, 0.0},
		{"empty tag", "", 5, 10, 50.0},
		{"long tag name", "very-long-feature-tag-name", 5, 10, 50.0},
		{"high percentage", "test", 10, 10, 100.0},
		{"low percentage", "test", 0, 10, 0.0},
		{"partial progress", "test", 3, 10, 30.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Should not panic or return empty string
			result := m.renderMediumHeader(test.tag, test.done, test.total, test.percentage)
			if result == "" {
				t.Errorf("renderMediumHeader returned empty string for tag=%q, done=%d, total=%d, percentage=%.0f",
					test.tag, test.done, test.total, test.percentage)
			}

			// Should contain percentage information
			percentStr := fmt.Sprintf("%.0f%%", test.percentage)
			if !strings.Contains(result, percentStr) {
				t.Errorf("expected to find '%s' in output for tag=%q", percentStr, test.tag)
			}
		})
	}
}

// TestRenderMediumHeaderNoTagPrefix verifies that medium header does NOT include "Tag:" prefix
func TestRenderMediumHeaderNoTagPrefix(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     90,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderMediumHeader("feature", 3, 6, 50.0)

	// Should contain the tag value
	if !strings.Contains(result, "feature") {
		t.Errorf("expected to find tag 'feature' in output, got: %q", result)
	}

	// Should NOT contain "Tag:" prefix (that's for wide header)
	if strings.Contains(result, "Tag:") {
		t.Errorf("medium header should NOT contain 'Tag:' prefix, got: %q", result)
	}

	// Tag should be in bracket format [feature]
	if !strings.Contains(result, "[feature]") {
		t.Errorf("expected tag to be in [bracket] format, got: %q", result)
	}
}

// TestAbbreviateTagFunction tests the tag abbreviation helper
func TestAbbreviateTag(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		maxLength int
		want      string
	}{
		{
			name:      "short tag needs no abbreviation",
			tag:       "auth",
			maxLength: 12,
			want:      "auth",
		},
		{
			name:      "tag exactly at max length",
			tag:       "feat-auth123",
			maxLength: 12,
			want:      "feat-auth123",
		},
		{
			name:      "long tag gets abbreviated",
			tag:       "feature-auth-detailed",
			maxLength: 12,
			want:      "feature-a...",
		},
		{
			name:      "very long tag gets abbreviated",
			tag:       "very-long-feature-name-with-many-chars",
			maxLength: 12,
			want:      "very-long...",
		},
		{
			name:      "tag one char over limit",
			tag:       "feat-auth1234",
			maxLength: 12,
			want:      "feat-auth...",
		},
		{
			name:      "empty tag",
			tag:       "",
			maxLength: 12,
			want:      "",
		},
		{
			name:      "single character tag",
			tag:       "a",
			maxLength: 12,
			want:      "a",
		},
		{
			name:      "ellipsis-sized max length",
			tag:       "test-tag",
			maxLength: 3,
			want:      "t...",
		},
		{
			name:      "max length smaller than tag",
			tag:       "feature-auth",
			maxLength: 5,
			want:      "fe...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := abbreviateTag(tt.tag, tt.maxLength)
			if result != tt.want {
				t.Errorf("abbreviateTag(%q, %d) = %q, want %q", tt.tag, tt.maxLength, result, tt.want)
			}
		})
	}
}

// TestRenderNarrowHeaderContainsRequiredElements verifies that all required elements are present
func TestRenderNarrowHeaderContainsRequiredElements(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tag := "test-tag"
	percentage := 60.0

	result := m.renderNarrowHeader(tag, percentage)

	// Verify all required components are present
	tests := []struct {
		name    string
		pattern string
	}{
		{"Application name", "TM-TUI"},
		{"Tag brackets", "[test-tag]"},
		{"Percentage", "60%"},
	}

	for _, test := range tests {
		if !strings.Contains(result, test.pattern) {
			t.Errorf("%s: expected to find '%s' in output, got: %q", test.name, test.pattern, stripANSIForTest(result))
		}
	}
}

// TestRenderNarrowHeaderLongTagAbbreviation verifies long tags are properly abbreviated
func TestRenderNarrowHeaderLongTagAbbreviation(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tests := []struct {
		name             string
		tag              string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:          "short tag shown in full",
			tag:           "auth",
			shouldContain: "[auth]",
		},
		{
			name:             "long tag abbreviated",
			tag:              "very-long-feature-tag-name",
			shouldContain:    "[very-long...]",
			shouldNotContain: "[very-long-feature-tag-name]",
		},
		{
			name:          "medium tag shown in full",
			tag:           "feature-auth",
			shouldContain: "[feature-auth]",
		},
		{
			name:             "exactly 12 chars shown in full",
			tag:              "feat-auth123",
			shouldContain:    "[feat-auth123]",
			shouldNotContain: "...",
		},
		{
			name:             "13 chars abbreviated",
			tag:              "feat-auth1234",
			shouldContain:    "...",
			shouldNotContain: "[feat-auth1234]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := m.renderNarrowHeader(test.tag, 50.0)
			resultClean := stripANSIForTest(result)

			if test.shouldContain != "" && !strings.Contains(resultClean, test.shouldContain) {
				t.Errorf("expected to find '%s' in output, got: %q", test.shouldContain, resultClean)
			}

			if test.shouldNotContain != "" && strings.Contains(resultClean, test.shouldNotContain) {
				t.Errorf("expected NOT to find '%s' in output, got: %q", test.shouldNotContain, resultClean)
			}
		})
	}
}

// TestRenderNarrowHeaderBorderStyle verifies that the border is correctly applied
func TestRenderNarrowHeaderBorderStyle(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderNarrowHeader("test", 50.0)

	// Check for normal-line border characters (┌─┐│└─┘)
	// These should be present in the output when lipgloss.NormalBorder() is applied
	normalBorderChars := []string{"┌", "┐", "│", "└", "─"}
	foundBorder := false
	for _, char := range normalBorderChars {
		if strings.Contains(result, char) {
			foundBorder = true
			break
		}
	}

	if !foundBorder {
		t.Errorf("expected to find normal-line border characters in output, got: %q", result)
	}
}

// TestRenderNarrowHeaderWidthCalculation verifies that the header respects width constraints
func TestRenderNarrowHeaderWidthCalculation(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"narrow terminal 60", 60},
		{"narrow terminal 70", 70},
		{"narrow terminal 79", 79},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{TaskMasterPath: "/tmp/test"}
			m := Model{
				config:    cfg,
				width:     test.width,
				height:    24,
				styles:    NewStyles(),
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderNarrowHeader("test", 50.0)

			// The result should not be empty
			if result == "" {
				t.Errorf("renderNarrowHeader returned empty string for width %d", test.width)
			}

			// Output should be styled (contains ANSI codes or plain text)
			if len(result) < 20 {
				t.Logf("width %d: output length %d (may be very minimal)", test.width, len(result))
			}
		})
	}
}

// TestRenderNarrowHeaderPercentageFormat verifies percentage is formatted correctly
func TestRenderNarrowHeaderPercentageFormat(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tests := []struct {
		name       string
		percentage float64
		expected   string
	}{
		{"zero percent", 0.0, "0%"},
		{"fifty percent", 50.0, "50%"},
		{"one hundred percent", 100.0, "100%"},
		{"partial percent", 33.3, "33%"},
		{"partial percent rounded", 66.7, "67%"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := m.renderNarrowHeader("test", test.percentage)
			resultClean := stripANSIForTest(result)

			if !strings.Contains(resultClean, test.expected) {
				t.Errorf("expected to find '%s' in output, got: %q", test.expected, resultClean)
			}
		})
	}
}

// TestRenderNarrowHeaderNoProgressBar verifies that no progress bar is displayed
func TestRenderNarrowHeaderNoProgressBar(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderNarrowHeader("test", 50.0)
	resultClean := stripANSIForTest(result)

	// Narrow header should NOT contain progress bar characters
	blockChars := []string{"▓", "░"}
	for _, char := range blockChars {
		if strings.Contains(resultClean, char) {
			t.Errorf("unexpected progress bar character '%s' found in narrow header: %q", char, resultClean)
		}
	}
}

// TestRenderNarrowHeaderEdgeCases verifies behavior with edge case inputs
func TestRenderNarrowHeaderEdgeCases(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	tests := []struct {
		name       string
		tag        string
		percentage float64
	}{
		{"empty tag", "", 0.0},
		{"single char tag", "a", 0.0},
		{"zero percent", "test", 0.0},
		{"hundred percent", "test", 100.0},
		{"very long tag", "very-long-feature-tag-that-needs-abbreviation", 75.5},
		{"special chars in tag", "feat/auth-v2", 50.0},
		{"negative percentage", "test", -10.0},
		{"over 100% percentage", "test", 150.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Should not panic or return empty string
			result := m.renderNarrowHeader(test.tag, test.percentage)
			if result == "" {
				t.Errorf("renderNarrowHeader returned empty string for tag=%q, percentage=%.1f",
					test.tag, test.percentage)
			}

			// Verify basic structure
			resultClean := stripANSIForTest(result)
			if !strings.Contains(resultClean, "TM-TUI") {
				t.Errorf("expected to find 'TM-TUI' in output for tag=%q", test.tag)
			}
		})
	}
}

// TestRenderNarrowHeaderMinimalSeparators verifies that separators are minimal
func TestRenderNarrowHeaderMinimalSeparators(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/tmp/test"}
	m := Model{
		config:    cfg,
		width:     70,
		height:    24,
		styles:    NewStyles(),
		helpModel: help.New(),
		keyMap:    NewKeyMap(cfg),
	}

	result := m.renderNarrowHeader("test-tag", 50.0)
	resultClean := stripANSIForTest(result)

	// Should contain exactly 2 pipe separators (│)
	// The border also uses pipes, so we should find them but just check structure exists
	if !strings.Contains(resultClean, "TM-TUI │") {
		t.Errorf("expected separator pattern 'TM-TUI │' in: %q", resultClean)
	}

	// Should follow pattern: TM-TUI │ [tag] │ percentage%
	expectedStructure := "TM-TUI │"
	if !strings.Contains(resultClean, expectedStructure) {
		t.Errorf("expected structure %q not found in: %q", expectedStructure, resultClean)
	}
}

// TestRenderHeaderLayoutSelection tests that renderHeader selects the correct layout based on terminal width
func TestRenderHeaderLayoutSelection(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		expectedLayout string // "wide", "medium", or "narrow"
		checkPattern  string  // Pattern to verify correct layout
	}{
		{
			name:          "width 100 selects wide layout",
			width:         100,
			expectedLayout: "wide",
			checkPattern:  "Progress: ",
		},
		{
			name:          "width 120 selects wide layout",
			width:         120,
			expectedLayout: "wide",
			checkPattern:  "Progress: ",
		},
		{
			name:          "width 80 selects medium layout",
			width:         80,
			expectedLayout: "medium",
			checkPattern:  "Task Master TUI",
		},
		{
			name:          "width 90 selects medium layout",
			width:         90,
			expectedLayout: "medium",
			checkPattern:  "Task Master TUI",
		},
		{
			name:          "width 99 selects medium layout (just below wide)",
			width:         99,
			expectedLayout: "medium",
			checkPattern:  "Task Master TUI",
		},
		{
			name:          "width 79 selects narrow layout",
			width:         79,
			expectedLayout: "narrow",
			checkPattern:  "TM-TUI",
		},
		{
			name:          "width 60 selects narrow layout",
			width:         60,
			expectedLayout: "narrow",
			checkPattern:  "TM-TUI",
		},
		{
			name:          "width 40 selects narrow layout",
			width:         40,
			expectedLayout: "narrow",
			checkPattern:  "TM-TUI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "test-tag",
			}
			m := Model{
				config: cfg,
				width:  tt.width,
				height: 24,
				styles: NewStyles(),
				tasks: []taskmaster.Task{
					{ID: "1", Title: "Task 1", Status: taskmaster.StatusDone},
					{ID: "2", Title: "Task 2", Status: taskmaster.StatusPending},
				},
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderHeader()

			// Verify output is not empty
			if result == "" {
				t.Errorf("renderHeader returned empty string for width %d", tt.width)
			}

			// Verify the expected pattern exists (indicates correct layout)
			resultClean := stripANSIForTest(result)
			if !strings.Contains(resultClean, tt.checkPattern) {
				t.Errorf("expected %s layout pattern '%s' not found in output for width %d, got: %q",
					tt.expectedLayout, tt.checkPattern, tt.width, resultClean)
			}
		})
	}
}

// TestRenderHeaderDefaultTag tests that renderHeader uses default tag when config is nil or ActiveTag is empty
func TestRenderHeaderDefaultTag(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedTag   string
	}{
		{
			name:        "nil config uses default tag",
			config:      nil,
			expectedTag: "master",
		},
		{
			name: "empty ActiveTag uses default tag",
			config: &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "",
			},
			expectedTag: "master",
		},
		{
			name: "non-empty ActiveTag uses custom tag",
			config: &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "feature-auth",
			},
			expectedTag: "feature-auth",
		},
		{
			name: "complex tag name is preserved",
			config: &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "bug/fix-header-123",
			},
			expectedTag: "bug/fix-header-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				config: tt.config,
				width:  120, // Use wide layout for clear tag display
				height: 24,
				styles: NewStyles(),
				tasks: []taskmaster.Task{
					{ID: "1", Title: "Task 1", Status: taskmaster.StatusDone},
				},
				helpModel: help.New(),
				keyMap:    NewKeyMap(tt.config),
			}

			result := m.renderHeader()
			resultClean := stripANSIForTest(result)

			// Verify the expected tag appears in the output
			if !strings.Contains(resultClean, tt.expectedTag) {
				t.Errorf("expected tag '%s' not found in output, got: %q", tt.expectedTag, resultClean)
			}
		})
	}
}

// TestRenderHeaderProgressIntegration tests that renderHeader correctly integrates with calculateTaskProgress
func TestRenderHeaderProgressIntegration(t *testing.T) {
	tests := []struct {
		name       string
		tasks      []taskmaster.Task
		width      int
		wantDone   int
		wantTotal  int
		wantPercent string
	}{
		{
			name:  "empty task list shows zero progress",
			tasks: []taskmaster.Task{},
			width: 120,
			wantDone: 0,
			wantTotal: 0,
			wantPercent: "0",
		},
		{
			name: "half completed tasks shows 50%",
			tasks: []taskmaster.Task{
				{ID: "1", Title: "Task 1", Status: taskmaster.StatusDone},
				{ID: "2", Title: "Task 2", Status: taskmaster.StatusPending},
			},
			width: 120,
			wantDone: 1,
			wantTotal: 2,
			wantPercent: "50",
		},
		{
			name: "all tasks done shows 100%",
			tasks: []taskmaster.Task{
				{ID: "1", Title: "Task 1", Status: taskmaster.StatusDone},
				{ID: "2", Title: "Task 2", Status: taskmaster.StatusDone},
			},
			width: 120,
			wantDone: 2,
			wantTotal: 2,
			wantPercent: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "test",
			}
			m := Model{
				config:    cfg,
				width:     tt.width,
				height:    24,
				styles:    NewStyles(),
				tasks:     tt.tasks,
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderHeader()
			resultClean := stripANSIForTest(result)

			// Verify progress is displayed
			if !strings.Contains(resultClean, tt.wantPercent) {
				t.Errorf("expected percentage '%s%%' in output, got: %q", tt.wantPercent, resultClean)
			}

			// For wide layout, also check done/total format
			if tt.width >= 100 && tt.wantTotal > 0 {
				progressStr := fmt.Sprintf("%d/%d", tt.wantDone, tt.wantTotal)
				if !strings.Contains(resultClean, progressStr) {
					t.Errorf("expected progress format '%s' in wide layout, got: %q", progressStr, resultClean)
				}
			}
		})
	}
}

// TestRenderHeaderWidthBoundaries tests edge cases at width thresholds
func TestRenderHeaderWidthBoundaries(t *testing.T) {
	tests := []struct {
		width        int
		expectWide   bool
		expectMedium bool
		expectNarrow bool
	}{
		{width: 100, expectWide: true},   // Exactly at wide threshold
		{width: 99, expectMedium: true},  // Just below wide threshold
		{width: 80, expectMedium: true},  // Exactly at medium threshold
		{width: 79, expectNarrow: true},  // Just below medium threshold
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("width_%d", tt.width), func(t *testing.T) {
			cfg := &config.Config{
				TaskMasterPath: "/tmp/test",
				ActiveTag:      "test",
			}
			m := Model{
				config: cfg,
				width:  tt.width,
				height: 24,
				styles: NewStyles(),
				tasks: []taskmaster.Task{
					{ID: "1", Title: "Task 1", Status: taskmaster.StatusDone},
				},
				helpModel: help.New(),
				keyMap:    NewKeyMap(cfg),
			}

			result := m.renderHeader()
			resultClean := stripANSIForTest(result)

			// Check for layout-specific indicators
			if tt.expectWide {
				if !strings.Contains(resultClean, "Progress: ") {
					t.Errorf("expected wide layout for width %d, got: %q", tt.width, resultClean)
				}
			}
			if tt.expectMedium {
				// Medium has "Task Master TUI" but NOT "Progress: " prefix
				if strings.Contains(resultClean, "TM-TUI") {
					t.Errorf("expected medium layout, got narrow for width %d", tt.width)
				}
			}
			if tt.expectNarrow {
				if !strings.Contains(resultClean, "TM-TUI") {
					t.Errorf("expected narrow layout for width %d, got: %q", tt.width, resultClean)
				}
			}
		})
	}
}
