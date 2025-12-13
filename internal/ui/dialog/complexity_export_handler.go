package dialog

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ExportRequest contains the details for file export
type ExportRequest struct {
	Format   string // "json", "csv"
	FilePath string // Path for the output file
}

// ExportResult contains the result of an export operation
type ExportResult struct {
	Success  bool   // Whether the export succeeded
	FilePath string // Full path of the exported file
	Error    error  // Error if any occurred
}

// ComplexityExportDialog creates a dialog for selecting export options
func ComplexityExportDialog(style *DialogStyle) (*FormDialog, error) {
	// Define format options
	formatOptions := []FormOption{
		{Value: "csv", Label: "CSV (Comma-Separated Values)"},
		{Value: "json", Label: "JSON (JavaScript Object Notation)"},
	}

	// Create fields
	fields := []FormField{
		{
			ID:       "format",
			Label:    "Export Format:",
			Type:     FormFieldTypeRadio,
			Required: true,
			Options:  formatOptions,
			Value:    "csv", // Default value
		},
		{
			ID:       "file_path",
			Label:    "Output Directory (optional):",
			Type:     FormFieldTypeText,
			Required: false,
			Value:    "",
			Help:     "Leave empty for current project directory",
		},
		{
			ID:       "filename",
			Label:    "File Name (optional):",
			Type:     FormFieldTypeText,
			Required: false,
			Value:    "",
			Help:     "Leave empty for auto-generated name",
		},
	}

	// Create export dialog
	dialog := NewFormDialog(
		"Export Complexity Results",
		"Choose format and location for export:",
		fields,
		[]string{"Export", "Cancel"},
		style,
		func(form *FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Export" {
				return nil, nil // User cancelled
			}

			// Extract format
			format, _ := values["format"].(string)
			if format == "" {
				return nil, fmt.Errorf("export format is required")
			}

			// Extract custom path if provided
			path, _ := values["file_path"].(string)

			// Extract custom filename if provided
			filename, _ := values["filename"].(string)
			if filename == "" {
				// Generate filename with timestamp
				timestamp := time.Now().Format("20060102-150405")
				filename = fmt.Sprintf("complexity-report-%s.%s", timestamp, format)
			} else {
				// Ensure correct extension
				if filepath.Ext(filename) == "" {
					filename = fmt.Sprintf("%s.%s", filename, format)
				}
			}

			// Combine path and filename if path provided
			filePath := filename
			if path != "" {
				filePath = filepath.Join(path, filename)
			}

			return ExportRequest{
				Format:   format,
				FilePath: filePath,
			}, nil
		},
	)

	return dialog, nil
}

// ShowExportCompleteDialog displays a confirmation dialog with the export results
func ShowExportCompleteDialog(result ExportResult, style *DialogStyle) *ButtonModalDialog {
	var title, message string

	if result.Success {
		title = "Export Successful"
		message = fmt.Sprintf("Complexity report exported to:\n%s", result.FilePath)
	} else {
		title = "Export Failed"
		message = fmt.Sprintf("Failed to export complexity report:\n%v", result.Error)
	}

	dialog := OkDialog(title, message)
	if style != nil {
		ApplyStyleToDialog(dialog, style)
	}
	return dialog
}

// HandleExportResult processes the result of an export operation
func HandleExportResult(result ExportResult) tea.Cmd {
	return func() tea.Msg {
		return DialogResultMsg{
			Button: "ok",
			Value:  result,
		}
	}
}
