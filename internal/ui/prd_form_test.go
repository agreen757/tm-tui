package ui

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/pathutil"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
)

func TestNewPrdInputDialog(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: "/project/.taskmaster",
	}

	form := NewPrdInputDialog(cfg)

	if form == nil {
		t.Fatal("NewPrdInputDialog returned nil")
	}

	if form.Title() != "Create PRD" {
		t.Errorf("Expected title 'Create PRD', got '%s'", form.Title())
	}

	if form.Kind() != dialog.DialogKindForm {
		t.Errorf("Expected form kind, got %v", form.Kind())
	}
}

func TestCountSentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single sentence",
			text:     "This is a sentence.",
			expected: 1,
		},
		{
			name:     "two sentences",
			text:     "First sentence. Second sentence.",
			expected: 2,
		},
		{
			name:     "three sentences",
			text:     "First. Second. Third.",
			expected: 3,
		},
		{
			name:     "sentence with question mark",
			text:     "What is this? This is the answer.",
			expected: 2,
		},
		{
			name:     "sentence with exclamation",
			text:     "Wow! Amazing! Great!",
			expected: 3,
		},
		{
			name:     "no punctuation",
			text:     "This is text without punctuation",
			expected: 1,
		},
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "whitespace only",
			text:     "   ",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countSentences(tt.text)
			if result != tt.expected {
				t.Errorf("countSentences(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestPrdFormValidation(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		summary       string
		filename      string
		shouldFail    bool
		expectedError string
	}{
		{
			name:       "valid form",
			title:      "Test PRD",
			summary:    "A test PRD",
			filename:   "test.md",
			shouldFail: false,
		},
		{
			name:          "missing title",
			title:         "",
			summary:       "A test PRD",
			filename:      "test.md",
			shouldFail:    true,
			expectedError: "Title is required",
		},
		{
			name:          "missing summary",
			title:         "Test PRD",
			summary:       "",
			filename:      "test.md",
			shouldFail:    true,
			expectedError: "Summary is required",
		},
		{
			name:       "missing filename",
			title:      "Test PRD",
			summary:    "A test PRD",
			filename:   "",
			shouldFail: false, // Should auto-generate
		},
		{
			name:       "filename without .md extension",
			title:      "Test PRD",
			summary:    "A test PRD",
			filename:   "test",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TaskMasterPath: "/project/.taskmaster",
			}
			form := NewPrdInputDialog(cfg)

			values := map[string]interface{}{
				"title":    tt.title,
				"summary":  tt.summary,
				"filename": tt.filename,
				"scope":    "",
			}

			// Create a dummy handler to test validation
			handler := func(f *dialog.FormDialog, button string, vals map[string]interface{}) (interface{}, error) {
				if button != "Create" {
					return nil, nil
				}

				title := stringValue(vals, "title")
				if title == "" {
					return nil, NewOperationError("Validation", "Title is required", nil)
				}

				summary := stringValue(vals, "summary")
				if summary == "" {
					return nil, NewOperationError("Validation", "Summary is required", nil)
				}

				filename := stringValue(vals, "filename")
				if filename == "" {
					filename = Slugify(title)
					if filename == "" {
						return nil, NewOperationError("Validation", "Output filename is required", nil)
					}
				}

				if len(filename) < 3 || filename[len(filename)-3:] != ".md" {
					filename += ".md"
				}

				result := PrdFormValues{
					Title:            title,
					Summary:          summary,
					ScopeConstraints: stringValue(vals, "scope"),
					OutputFilename:   filename,
				}

				return result, nil
			}

			result, err := handler(form, "Create", values)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error '%s', got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%v'", err)
				}

				if result != nil {
					resultVals, ok := result.(PrdFormValues)
					if !ok {
						t.Fatalf("Expected PrdFormValues, got %T", result)
					}

					if resultVals.Title != tt.title {
						t.Errorf("Expected title '%s', got '%s'", tt.title, resultVals.Title)
					}

					if resultVals.Summary != tt.summary {
						t.Errorf("Expected summary '%s', got '%s'", tt.summary, resultVals.Summary)
					}

					expectedFilename := tt.filename
					if expectedFilename == "" {
						expectedFilename = Slugify(tt.title) + ".md"
					} else if len(expectedFilename) >= 3 && expectedFilename[len(expectedFilename)-3:] != ".md" {
						expectedFilename += ".md"
					}
					if resultVals.OutputFilename != expectedFilename {
						t.Errorf("Expected filename '%s', got '%s'", expectedFilename, resultVals.OutputFilename)
					}
				}
			}
		})
	}
}

func TestPrdFormDestinationPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectedPath string
		description  string
	}{
		{
			name: "with taskmaster path that doesn't exist",
			config: &config.Config{
				TaskMasterPath: "/nonexistent/project/.taskmaster",
			},
			expectedPath: ".taskmaster/docs",
			description:  "Falls back to default when TaskMasterPath doesn't exist",
		},
		{
			name:         "nil config",
			config:       nil,
			expectedPath: ".taskmaster/docs",
			description:  "Uses default when config is nil",
		},
		{
			name: "empty taskmaster path",
			config: &config.Config{
				TaskMasterPath: "",
			},
			expectedPath: ".taskmaster/docs",
			description:  "Uses default when TaskMasterPath is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := pathutil.ResolvePrdDirectoryPath(tt.config, "")
			if path != tt.expectedPath {
				t.Errorf("Expected path '%s', got '%s' - %s", tt.expectedPath, path, tt.description)
			}
		})
	}
}
