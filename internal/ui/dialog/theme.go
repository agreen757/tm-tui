package dialog

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeColors contains all the color definitions for a dialog theme
type ThemeColors struct {
	Border          string
	FocusedBorder   string
	Title           string
	Background      string
	Text            string
	Button          string
	Error           string
	Success         string
	Warning         string
	InputField      string
	PlaceholderText string
	FocusedText     string
	HighlightText   string
}

// DefaultThemeColors returns the default theme colors
func DefaultThemeColors() ThemeColors {
	return ThemeColors{
		Border:          "#444444",
		FocusedBorder:   "#6D98BA", // Highlight color
		Title:           "#EEEEEE",
		Background:      "#333333",
		Text:            "#DDDDDD",
		Button:          "#6D98BA", // Highlight color
		Error:           "#F7768E", // Error color (red)
		Success:         "#9ECE6A", // Success color (green)
		Warning:         "#E0AF68", // Warning color (orange)
		InputField:      "#444444",
		PlaceholderText: "#666666",
		FocusedText:     "#FFFFFF",
		HighlightText:   "#00FFFF",
	}
}

// HighContrastThemeColors returns a high contrast theme for accessibility
func HighContrastThemeColors() ThemeColors {
	return ThemeColors{
		Border:          "#FFFFFF",
		FocusedBorder:   "#FFFF00", // Yellow for high contrast
		Title:           "#FFFFFF",
		Background:      "#000000",
		Text:            "#FFFFFF",
		Button:          "#FFFF00", // Yellow for high contrast
		Error:           "#FF0000", // Bright red
		Success:         "#00FF00", // Bright green
		Warning:         "#FFFF00", // Bright yellow
		InputField:      "#FFFFFF",
		PlaceholderText: "#CCCCCC",
		FocusedText:     "#FFFFFF",
		HighlightText:   "#FFFF00", // Yellow for high contrast
	}
}

// ApplyTheme applies a theme to a dialog manager and all its dialogs
func ApplyTheme(dm *DialogManager, theme ThemeColors) {
	// Create a new dialog style from theme colors
	dialogStyle := &DialogStyle{
		Border:             lipgloss.RoundedBorder(),
		BorderColor:        lipgloss.Color(theme.Border),
		FocusedBorderColor: lipgloss.Color(theme.FocusedBorder),
		TitleColor:         lipgloss.Color(theme.Title),
		BackgroundColor:    lipgloss.Color(theme.Background),
		TextColor:          lipgloss.Color(theme.Text),
		ButtonColor:        lipgloss.Color(theme.Button),
		ErrorColor:         lipgloss.Color(theme.Error),
		SuccessColor:       lipgloss.Color(theme.Success),
		WarningColor:       lipgloss.Color(theme.Warning),
	}

	// Store style on manager for easy access
	dm.Style = dialogStyle

	// Apply the style to all dialogs in the stack
	for _, entry := range dm.dialogs {
		ApplyStyleToDialog(entry.dialog, dialogStyle)
	}
}

// ApplyStyleToDialog applies a dialog style to a specific dialog
func ApplyStyleToDialog(d Dialog, style *DialogStyle) {
	switch dialog := d.(type) {
	case *ModalDialog:
		dialog.BaseDialog.Style = style
	case *ButtonModalDialog:
		dialog.ModalDialog.BaseDialog.Style = style
	case *FormDialog:
		dialog.BaseFocusableDialog.BaseDialog.Style = style
	case *ListDialog:
		dialog.BaseFocusableDialog.BaseDialog.Style = style
	case *ConfirmationDialog:
		dialog.BaseFocusableDialog.BaseDialog.Style = style
	case *ProgressDialog:
		dialog.BaseDialog.Style = style
	}
}

// CreateDialogStyleFromAppStyles creates a DialogStyle from application styles
func CreateDialogStyleFromAppStyles(borderColor, focusedColor, titleColor, backgroundColor,
	textColor, buttonColor, errorColor, successColor, warningColor string) *DialogStyle {
	return &DialogStyle{
		Border:             lipgloss.RoundedBorder(),
		BorderColor:        lipgloss.Color(borderColor),
		FocusedBorderColor: lipgloss.Color(focusedColor),
		TitleColor:         lipgloss.Color(titleColor),
		BackgroundColor:    lipgloss.Color(backgroundColor),
		TextColor:          lipgloss.Color(textColor),
		ButtonColor:        lipgloss.Color(buttonColor),
		ErrorColor:         lipgloss.Color(errorColor),
		SuccessColor:       lipgloss.Color(successColor),
		WarningColor:       lipgloss.Color(warningColor),
	}
}

// InitializeDialogManager creates and initializes a dialog manager with the given style
func InitializeDialogManager(width, height int, style *DialogStyle) *DialogManager {
	manager := NewDialogManager(width, height)

	// Set the default style for all new dialogs
	DefaultDialogStyle = func() *DialogStyle {
		return style
	}

	manager.Style = style

	return manager
}
