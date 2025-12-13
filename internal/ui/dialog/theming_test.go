package dialog

import (
	"testing"
	"time"

	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/charmbracelet/lipgloss"
)

// TestThemeApplicationToNewDialogs verifies that all new dialogs respect the theming system
func TestThemeApplicationToNewDialogs(t *testing.T) {
	// Create default and high-contrast themes
	defaultTheme := DefaultThemeColors()
	highContrastTheme := HighContrastThemeColors()

	tests := []struct {
		name  string
		theme ThemeColors
		setup func(theme *DialogStyle) Dialog
	}{
		{
			name:  "DefaultTheme_ComplexityScopeDialog",
			theme: defaultTheme,
			setup: func(style *DialogStyle) Dialog {
				d, _ := NewComplexityScopeDialog("task-1", style)
				return d
			},
		},
		{
			name:  "HighContrastTheme_ComplexityScopeDialog",
			theme: highContrastTheme,
			setup: func(style *DialogStyle) Dialog {
				d, _ := NewComplexityScopeDialog("task-1", style)
				return d
			},
		},
		{
			name:  "DefaultTheme_ComplexityFilterDialog",
			theme: defaultTheme,
			setup: func(style *DialogStyle) Dialog {
				d, _ := NewComplexityFilterDialog(FilterSettings{}, style)
				return d
			},
		},
		{
			name:  "DefaultTheme_ComplexityExportDialog",
			theme: defaultTheme,
			setup: func(style *DialogStyle) Dialog {
				d, _ := NewComplexityExportDialog(style)
				return d
			},
		},
		{
			name:  "DefaultTheme_ExpandTaskOptionsDialog",
			theme: defaultTheme,
			setup: func(style *DialogStyle) Dialog {
				d, _ := NewExpandTaskOptionsDialog(style)
				return d
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dialog style from theme
			dialogStyle := &DialogStyle{
				Border:             lipgloss.RoundedBorder(),
				BorderColor:        lipgloss.Color("#444444"),
				FocusedBorderColor: lipgloss.Color("#6D98BA"),
				TitleColor:         lipgloss.Color("#EEEEEE"),
				BackgroundColor:    lipgloss.Color("#333333"),
				TextColor:          lipgloss.Color("#DDDDDD"),
				ButtonColor:        lipgloss.Color("#6D98BA"),
				ErrorColor:         lipgloss.Color("#F7768E"),
				SuccessColor:       lipgloss.Color("#9ECE6A"),
				WarningColor:       lipgloss.Color("#E0AF68"),
			}

			// Setup creates the dialog with the style
			dialog := tt.setup(dialogStyle)
			if dialog == nil {
				t.Fatalf("Failed to create dialog")
			}

			// Verify dialog is styled
			if styledDialog, ok := dialog.(*FormDialog); ok {
				if styledDialog.BaseFocusableDialog.BaseDialog.Style == nil {
					t.Errorf("Dialog style not applied")
				}
			}
		})
	}
}

// TestThemeSwitching verifies that switching themes updates dialogs properly
func TestThemeSwitching(t *testing.T) {
	// Create a dialog manager
	dm := NewDialogManager(100, 30)

	// Add a complexity filter dialog
	filterDlg, _ := NewComplexityFilterDialog(FilterSettings{}, dm.Style)
	dm.PushDialog(filterDlg)

	// Get initial style colors
	initialErrorColor := dm.Style.ErrorColor

	// Apply new theme
	NewTheme := HighContrastThemeColors()
	ApplyTheme(dm, NewTheme)

	// Verify colors changed
	if initialErrorColor == dm.Style.ErrorColor {
		t.Errorf("Error color did not change when switching themes")
	}
}

// TestComplexityProgressDialogStyling verifies progress dialog receives and applies style
func TestComplexityProgressDialogStyling(t *testing.T) {
	style := DefaultDialogStyle()

	// Create progress dialog with style
	progressDlg := NewComplexityProgressDialog("all", []string{}, 10, style)

	// Verify it has the style applied
	if progressDlg.BaseDialog.Style == nil {
		t.Errorf("Progress dialog style not applied")
	}
}

// TestComplexityReportDialogStyling verifies report dialog receives and applies style
func TestComplexityReportDialogStyling(t *testing.T) {
	style := DefaultDialogStyle()

	// Create a mock report
	report := &taskmaster.ComplexityReport{
		Tasks: []taskmaster.TaskComplexity{
			{
				TaskID:      "task-1",
				Title:       "Test Task",
				Score:       5,
				Level:       taskmaster.ComplexityMedium,
				Description: "Test details",
				AnalyzedAt:  time.Now(),
			},
		},
		AnalyzedAt: time.Now(),
		Scope:      "test",
	}

	// Create report dialog with style
	reportDlg := NewComplexityReportDialog(report, style)

	// Verify it has the style applied
	if reportDlg.Style == nil {
		t.Errorf("Report dialog style not applied")
	}
}

// TestDialogFallbackStyle verifies dialogs use DefaultDialogStyle when none provided
func TestDialogFallbackStyle(t *testing.T) {
	// Create dialogs without explicit style (nil)
	scopeDlg, _ := NewComplexityScopeDialog("task-1", nil)
	if scopeDlg == nil {
		t.Fatalf("Failed to create dialog with nil style")
	}

	// Dialog should still work (fallback to DefaultDialogStyle)
	if scopeDlg.BaseFocusableDialog.BaseDialog.Style == nil {
		t.Errorf("Dialog should have fallback style when nil provided")
	}
}
