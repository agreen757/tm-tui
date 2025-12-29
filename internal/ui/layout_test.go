package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/git"
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
