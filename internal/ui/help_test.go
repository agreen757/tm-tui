package ui

import (
	"strings"
	"testing"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/executor"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpOverlay(t *testing.T) {
	// Create a test model
	cfg := &config.Config{
		KeyBindings: make(map[string]string),
		TaskMasterPath: ".",
	}
	configManager, _ := config.NewConfigManager()
	taskService := &taskmaster.Service{}
	execService, _ := executor.NewService(cfg)
	
	model := NewModel(cfg, configManager, taskService, execService)
	model.ready = true  // Set model as ready
	model.width = 100   // Set dimensions
	model.height = 40
	
	// Test help toggle on
	assert.False(t, model.showHelp, "Help should be initially hidden")
	
	// Simulate pressing '?' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	assert.True(t, m.showHelp, "Help should be shown after pressing '?'")
	
	// Test help overlay content
	view := m.View()
	
	// Check that help overlay contains expected sections
	assert.Contains(t, view, "Task Master TUI Help", "Should contain help title")
	assert.Contains(t, view, "Navigation", "Should contain Navigation section")
	assert.Contains(t, view, "Task Operations", "Should contain Task Operations section")
	assert.Contains(t, view, "Status Changes", "Should contain Status Changes section")
	assert.Contains(t, view, "Panels & Views", "Should contain Panels & Views section")
	assert.Contains(t, view, "General", "Should contain General section")
	assert.Contains(t, view, "About", "Should contain About section")
	
	// Check for key bindings display
	assert.Contains(t, view, "Move up", "Should show move up help")
	assert.Contains(t, view, "Move down", "Should show move down help")
	assert.Contains(t, view, "Toggle expand/collapse", "Should show expand help")
	assert.Contains(t, view, "Get next available task", "Should show next task help")
	
	// Test help toggle off with '?'
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ = m.Update(msg)
	m2 := updatedModel.(Model)
	assert.False(t, m2.showHelp, "Help should be hidden after pressing '?' again")
	
	// Test help toggle off with Escape  
	m.showHelp = true
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(msg)
	m3 := updatedModel.(Model)
	assert.False(t, m3.showHelp, "Help should be hidden after pressing Escape")
}

func TestHelpKeyBindings(t *testing.T) {
	// Test KeyMap methods
	km := DefaultKeyMap()
	
	// Test ShortHelp
	shortHelp := km.ShortHelp()
	assert.NotEmpty(t, shortHelp, "ShortHelp should return bindings")
	assert.Greater(t, len(shortHelp), 5, "ShortHelp should contain several bindings")
	
	// Test FullHelp
	fullHelp := km.FullHelp()
	assert.NotEmpty(t, fullHelp, "FullHelp should return binding groups")
	assert.Greater(t, len(fullHelp), 5, "FullHelp should contain several groups")
	
	// Verify specific key bindings are included
	foundHelp := false
	for _, binding := range shortHelp {
		if strings.Contains(binding.Help().Desc, "help") {
			foundHelp = true
			break
		}
	}
	assert.True(t, foundHelp, "ShortHelp should include help binding")
}

func TestRenderBinding(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: ".",
	}
	configManager, _ := config.NewConfigManager()
	taskService := &taskmaster.Service{}
	execService, _ := executor.NewService(cfg)
	
	model := NewModel(cfg, configManager, taskService, execService)
	
	// Test single key binding
	binding := model.keyMap.Help
	rendered := model.renderBinding(binding)
	assert.Contains(t, rendered, "?", "Should render help key")
	
	// Test multi-key binding
	binding = model.keyMap.Up
	rendered = model.renderBinding(binding)
	// Should contain either 'up' or 'k'
	assert.True(t, strings.Contains(rendered, "up") || strings.Contains(rendered, "k"), 
		"Should render up key binding")
}

func TestHelpOverlayInteraction(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: ".",
	}
	configManager, _ := config.NewConfigManager()
	taskService := &taskmaster.Service{}
	execService, _ := executor.NewService(cfg)
	
	model := NewModel(cfg, configManager, taskService, execService)
	
	// Test that other keys are ignored when help is shown
	model.showHelp = true
	
	// Try pressing 'q' (quit) - should be ignored
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	assert.True(t, m.showHelp, "Help should still be shown")
	assert.Nil(t, cmd, "Should not quit when help is shown")
	
	// Try pressing 'n' (next task) - should be ignored
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ = model.Update(msg)
	m = updatedModel.(Model)
	
	assert.True(t, m.showHelp, "Help should still be shown")
	// Verify no command was executed by checking that execService is not running
	assert.False(t, execService.IsRunning(), "Should not execute commands when help is shown")
}

func TestHelpOverlayStyling(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: ".",
	}
	configManager, _ := config.NewConfigManager()
	taskService := &taskmaster.Service{}
	execService, _ := executor.NewService(cfg)
	
	model := NewModel(cfg, configManager, taskService, execService)
	model.width = 100
	model.height = 40
	model.ready = true
	
	// Enable help
	model.showHelp = true
	
	// Get the help overlay
	overlay := model.renderHelpOverlay()
	
	// Check for border styling - should contain border characters
	assert.True(t, 
		strings.Contains(overlay, "═") || 
		strings.Contains(overlay, "║") ||
		strings.Contains(overlay, "╔") ||
		strings.Contains(overlay, "╗") ||
		strings.Contains(overlay, "╚") ||
		strings.Contains(overlay, "╝"),
		"Should have border styling")
	
	// Check footer instruction
	assert.Contains(t, overlay, "Press '?' or 'Esc' to close help", 
		"Should show close instructions")
}

func TestKeyStyles(t *testing.T) {
	styles := NewStyles()
	
	// Test that Key style exists and is configured
	require.NotNil(t, styles.Key, "Key style should be defined")
	
	// Test rendering a key with the style
	rendered := styles.Key.Render("Enter")
	assert.NotEmpty(t, rendered, "Should render key with styling")
}