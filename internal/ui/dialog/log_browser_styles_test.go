package dialog

import (
	"strings"
	"testing"
)

// TestNewLogBrowserStyles verifies that styles are created with correct properties
func TestNewLogBrowserStyles(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	if styles == nil {
		t.Fatal("NewLogBrowserStyles returned nil")
	}
	
	// Verify all styles are initialized by checking if they can render
	testContent := "Test"
	
	if styles.PanelStyle.Render(testContent) == "" {
		t.Error("PanelStyle not initialized")
	}
	if styles.FocusedPanelStyle.Render(testContent) == "" {
		t.Error("FocusedPanelStyle not initialized")
	}
	if styles.UnfocusedPanelStyle.Render(testContent) == "" {
		t.Error("UnfocusedPanelStyle not initialized")
	}
	if styles.TitleStyle.Render(testContent) == "" {
		t.Error("TitleStyle not initialized")
	}
	if styles.SubtitleStyle.Render(testContent) == "" {
		t.Error("SubtitleStyle not initialized")
	}
	if styles.HintStyle.Render(testContent) == "" {
		t.Error("HintStyle not initialized")
	}
	if styles.EmptyStateStyle.Render(testContent) == "" {
		t.Error("EmptyStateStyle not initialized")
	}
	if styles.LoadingStyle.Render(testContent) == "" {
		t.Error("LoadingStyle not initialized")
	}
	if styles.ErrorStyle.Render(testContent) == "" {
		t.Error("ErrorStyle not initialized")
	}
	if styles.CurrentLineStyle.Render(testContent) == "" {
		t.Error("CurrentLineStyle not initialized")
	}
}

// TestGetEmptyStateMessage verifies correct empty state messages
func TestGetEmptyStateMessage(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	tests := []struct {
		name      string
		panelType string
		want      string
	}{
		{
			name:      "file browser empty state",
			panelType: "file_browser",
			want:      "No log files found",
		},
		{
			name:      "tag selector empty state",
			panelType: "tag_selector",
			want:      "No tags available",
		},
		{
			name:      "log viewer empty state",
			panelType: "log_viewer",
			want:      "Select a log file to preview",
		},
		{
			name:      "unknown panel type",
			panelType: "unknown",
			want:      "No content available",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := styles.GetEmptyStateMessage(tt.panelType)
			if !strings.Contains(got, tt.want) {
				t.Errorf("GetEmptyStateMessage(%q) = %q, want to contain %q", tt.panelType, got, tt.want)
			}
		})
	}
}

// TestRenderEmptyState verifies empty state rendering
func TestRenderEmptyState(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	tests := []struct {
		name      string
		panelType string
		width     int
		height    int
	}{
		{
			name:      "file browser",
			panelType: "file_browser",
			width:     40,
			height:    20,
		},
		{
			name:      "tag selector",
			panelType: "tag_selector",
			width:     30,
			height:    20,
		},
		{
			name:      "log viewer",
			panelType: "log_viewer",
			width:     50,
			height:    20,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := styles.RenderEmptyState(tt.panelType, tt.width, tt.height)
			if rendered == "" {
				t.Error("RenderEmptyState returned empty string")
			}
			
			// The rendered output contains ANSI codes, so just verify it's not empty
			// and has reasonable length (message + styling)
			if len(rendered) < 10 {
				t.Errorf("RenderEmptyState output too short: %d chars", len(rendered))
			}
		})
	}
}

// TestRenderLoadingState verifies loading state rendering
func TestRenderLoadingState(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	message := "Loading log files..."
	spinnerView := "⠋"
	
	rendered := styles.RenderLoadingState(message, spinnerView)
	
	if rendered == "" {
		t.Error("RenderLoadingState returned empty string")
	}
	
	if !strings.Contains(rendered, message) {
		t.Errorf("RenderLoadingState output doesn't contain message: %q", message)
	}
	
	if !strings.Contains(rendered, spinnerView) {
		t.Errorf("RenderLoadingState output doesn't contain spinner: %q", spinnerView)
	}
}

// TestStyleConsistency verifies that focused and unfocused styles are different
func TestStyleConsistency(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	// Render the same content with both styles
	content := "Test Content"
	focused := styles.FocusedPanelStyle.Render(content)
	unfocused := styles.UnfocusedPanelStyle.Render(content)
	
	// They should render differently (different border colors)
	if focused == unfocused {
		t.Error("Focused and unfocused panel styles render identically")
	}
}

// TestEmptyStateMessagesHaveHints verifies that empty states include helpful hints
func TestEmptyStateMessagesHaveHints(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	panelTypes := []string{"file_browser", "tag_selector", "log_viewer"}
	
	for _, panelType := range panelTypes {
		t.Run(panelType, func(t *testing.T) {
			message := styles.GetEmptyStateMessage(panelType)
			
			// All messages should contain "Tip:" to provide helpful guidance
			if !strings.Contains(message, "Tip:") {
				t.Errorf("Empty state message for %q should contain 'Tip:' for user guidance", panelType)
			}
		})
	}
}

// TestStyleColors verifies that styles use consistent color scheme
func TestStyleColors(t *testing.T) {
	styles := NewLogBrowserStyles()
	
	// Verify key colors are set (by checking render output contains ANSI codes)
	testContent := "Test"
	
	// Focused border should use cyan (#8be9fd)
	focused := styles.FocusedPanelStyle.Render(testContent)
	if !strings.Contains(focused, "\x1b[") {
		t.Error("FocusedPanelStyle doesn't apply ANSI color codes")
	}
	
	// Error style should use red
	error := styles.ErrorStyle.Render(testContent)
	if !strings.Contains(error, "\x1b[") {
		t.Error("ErrorStyle doesn't apply ANSI color codes")
	}
	
	// Loading style should use cyan
	loading := styles.LoadingStyle.Render(testContent)
	if !strings.Contains(loading, "\x1b[") {
		t.Error("LoadingStyle doesn't apply ANSI color codes")
	}
}

// BenchmarkRenderEmptyState benchmarks empty state rendering performance
func BenchmarkRenderEmptyState(b *testing.B) {
	styles := NewLogBrowserStyles()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = styles.RenderEmptyState("file_browser", 80, 24)
	}
}

// BenchmarkGetEmptyStateMessage benchmarks message retrieval
func BenchmarkGetEmptyStateMessage(b *testing.B) {
	styles := NewLogBrowserStyles()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = styles.GetEmptyStateMessage("file_browser")
	}
}

// TestHighContrastTheme verifies high-contrast theme creation
func TestHighContrastTheme(t *testing.T) {
	styles := NewLogBrowserStylesWithTheme(true)
	
	if !styles.HighContrast {
		t.Error("High contrast flag not set")
	}
	
	// Verify styles are initialized
	testContent := "Test"
	if styles.FocusedPanelStyle.Render(testContent) == "" {
		t.Error("High contrast focused style not initialized")
	}
	if styles.ErrorStyle.Render(testContent) == "" {
		t.Error("High contrast error style not initialized")
	}
}

// TestGetDefaultTheme verifies default theme configuration
func TestGetDefaultTheme(t *testing.T) {
	theme := GetDefaultTheme()
	
	if theme.FocusedBorder == "" {
		t.Error("FocusedBorder not set")
	}
	if theme.Title == "" {
		t.Error("Title color not set")
	}
	if theme.Error == "" {
		t.Error("Error color not set")
	}
}

// TestGetHighContrastTheme verifies high-contrast theme configuration
func TestGetHighContrastTheme(t *testing.T) {
	theme := GetHighContrastTheme()
	
	if theme.FocusedBorder == "" {
		t.Error("FocusedBorder not set")
	}
	if theme.Title == "" {
		t.Error("Title color not set")
	}
	if theme.Error == "" {
		t.Error("Error color not set")
	}
	
	// High contrast should use different colors than default
	defaultTheme := GetDefaultTheme()
	if theme.FocusedBorder == defaultTheme.FocusedBorder {
		t.Error("High contrast focused border should differ from default")
	}
	if theme.LineHighlight == defaultTheme.LineHighlight {
		t.Error("High contrast line highlight should differ from default")
	}
}

// TestThemeConsistency verifies theme colors are valid hex codes
func TestThemeConsistency(t *testing.T) {
	themes := []struct {
		name  string
		theme ThemeConfig
	}{
		{"default", GetDefaultTheme()},
		{"high-contrast", GetHighContrastTheme()},
	}
	
	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			// Check all colors start with #
			if !strings.HasPrefix(tt.theme.FocusedBorder, "#") {
				t.Errorf("%s: FocusedBorder should be hex color", tt.name)
			}
			if !strings.HasPrefix(tt.theme.Title, "#") {
				t.Errorf("%s: Title should be hex color", tt.name)
			}
			if !strings.HasPrefix(tt.theme.Error, "#") {
				t.Errorf("%s: Error should be hex color", tt.name)
			}
		})
	}
}
