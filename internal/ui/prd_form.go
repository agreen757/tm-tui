package ui

import (
	"fmt"
	"strings"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/pathutil"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

// PrdFormValues holds the values submitted from the PRD form
type PrdFormValues struct {
	Title            string
	Summary          string
	ScopeConstraints string
	OutputFilename   string
}

// NewPrdInputDialog creates a form dialog to collect PRD information
func NewPrdInputDialog(cfg *config.Config) *dialog.FormDialog {
	fields := []dialog.FormField{
		{
			ID:          "title",
			Label:       "Title",
			Type:        dialog.FormFieldTypeText,
			Required:    true,
			Placeholder: "Enter PRD title",
			Help:        "The main title for your PRD",
		},
		{
			ID:          "summary",
			Label:       "Summary",
			Type:        dialog.FormFieldTypeTextArea,
			Required:    true,
			Placeholder: "Brief overview of the PRD",
			//Help:        "A concise summary of what this PRD covers",
			Rows:        6,
			Border:      true,
		},
		{
			ID:          "filename",
			Label:       "Output filename",
			Type:        dialog.FormFieldTypeText,
			Required:    true,
			Placeholder: "prd.md",
			Help:        "Filename for the PRD file (without .md extension)",
		},
		{
			ID:       "destination",
			Label:    "Destination",
			Type:     dialog.FormFieldTypeText,
			Required: false,
			Value:    pathutil.ResolvePrdDirectoryPath(cfg, ""),
			Help:     "Where the PRD file will be saved",
		},
	}

	return createPrdFormDialog(cfg, fields, nil)
}

// NewPrdInputDialogWithState creates a form dialog with pre-filled values from state
func NewPrdInputDialogWithState(cfg *config.Config, state *PrdCreationState) *dialog.FormDialog {
	fields := []dialog.FormField{
		{
			ID:          "title",
			Label:       "Title",
			Type:        dialog.FormFieldTypeText,
			Required:    true,
			Placeholder: "Enter PRD title",
			Value:       state.Title,
			Help:        "The main title for your PRD",
		},
		{
			ID:          "summary",
			Label:       "Summary (1-3 sentences)",
			Type:        dialog.FormFieldTypeText,
			Required:    true,
			Placeholder: "Brief overview of the PRD",
			Value:       state.Summary,
			//Help:        "A concise summary of what this PRD covers",
		},
		{
			ID:          "scope",
			Label:       "Scope/Constraints (optional)",
			Type:        dialog.FormFieldTypeText,
			Required:    false,
			Placeholder: "Define scope and constraints",
			Value:       state.Scope,
			Help:        "Any limitations or boundaries for this PRD",
		},
		{
			ID:          "filename",
			Label:       "Output filename",
			Type:        dialog.FormFieldTypeText,
			Required:    true,
			Placeholder: "prd.md",
			Value:       state.Filename,
			Help:        "Filename for the PRD file (without .md extension)",
		},
		{
			ID:       "destination",
			Label:    "Destination",
			Type:     dialog.FormFieldTypeText,
			Required: false,
			Value:    pathutil.ResolvePrdDirectoryPath(cfg, ""),
			Help:     "Where the PRD file will be saved",
		},
	}

	return createPrdFormDialog(cfg, fields, state)
}

// createPrdFormDialog is the common form creation logic used by both new and state-based constructors
func createPrdFormDialog(cfg *config.Config, fields []dialog.FormField, state *PrdCreationState) *dialog.FormDialog {
	// Create the form dialog
	form := dialog.NewFormDialog(
		"Create PRD",
		"Enter PRD Details",
		fields,
		[]string{"Create", "Cancel"},
		nil,
		func(f *dialog.FormDialog, button string, values map[string]interface{}) (interface{}, error) {
			if button != "Create" {
				return nil, nil
			}

			// Validate required fields
			title := stringValue(values, "title")
			if title == "" {
				return nil, fmt.Errorf("Title is required")
			}

			summary := stringValue(values, "summary")
			if summary == "" {
				return nil, fmt.Errorf("Summary is required")
			}

			filename := stringValue(values, "filename")
			if filename == "" {
				// If filename is empty, auto-generate from title
				filename = Slugify(title)
				if filename == "" {
					return nil, fmt.Errorf("Output filename is required")
				}
			}

			// Ensure filename ends with .md
			if len(filename) < 3 || filename[len(filename)-3:] != ".md" {
				filename += ".md"
			}

			result := PrdFormValues{
				Title:            title,
				Summary:          summary,
				ScopeConstraints: stringValue(values, "scope"),
				OutputFilename:   filename,
			}

			// Update the state with the form values
			if state != nil {
				state.UpdateFromFormValues(result)
			}

			return result, nil
		},
	)

	// Add event handler to track state changes
	form.AddEventHandler(func(f *dialog.FormDialog, msg tea.Msg) {
		if valueMsg, ok := msg.(dialog.FormValueChangedMsg); ok {
			if state == nil {
				return
			}

			// Update state whenever a field changes
			switch valueMsg.FieldID {
			case "title":
				if str, ok := valueMsg.NewValue.(string); ok {
					state.Title = str
				}
			case "summary":
				if str, ok := valueMsg.NewValue.(string); ok {
					state.Summary = str
				}
			case "scope":
				if str, ok := valueMsg.NewValue.(string); ok {
					state.Scope = str
				}
			case "filename":
				if str, ok := valueMsg.NewValue.(string); ok {
					state.Filename = str
				}
			}
		}
	})

	// Add custom validators
	form.AddValidator(func(values map[string]interface{}) error {
		// Validate summary is between 1-3 sentences
		summary := stringValue(values, "summary")
		if summary != "" {
			sentenceCount := countSentences(summary)
			if sentenceCount < 1 {
				return &dialog.ErrorFormValidation{
					FieldID: "summary",
					Message: "Summary must be at least 1 sentence",
				}
			}
			if sentenceCount > 3 {
				return &dialog.ErrorFormValidation{
					FieldID: "summary",
					Message: "Summary must be at most 3 sentences",
				}
			}
		}
		return nil
	})

	return form
}



// countSentences counts the approximate number of sentences in a text
func countSentences(text string) int {
	// Count sentence-ending punctuation marks
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	count := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			// Check if it's followed by whitespace or end of string
			if i == len(text)-1 || text[i+1] == ' ' || text[i+1] == '\t' || text[i+1] == '\n' {
				count++
			}
		}
	}

	// If no sentence endings found, treat the whole text as one sentence
	if count == 0 && len(text) > 0 {
		count = 1
	}

	return count
}
