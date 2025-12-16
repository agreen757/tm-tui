package dialog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agreen757/tm-tui/internal/config"
)

// ModelOption represents an available AI model with metadata
type ModelOption struct {
	Provider      string
	ModelID       string
	DisplayName   string
	ContextWindow int
	InputCost     float64  // per 1M tokens
	OutputCost    float64  // per 1M tokens
}

// ModelSelectionResult is the result of model selection
type ModelSelectionResult struct {
	Provider string // "anthropic", "openai", etc.
	ModelID  string // "claude-3-5-sonnet-20241022"
}

// ModelSelectionMsg is sent when a model is selected from the dialog
type ModelSelectionMsg struct {
	Provider  string
	ModelName string
	ModelID   string
}

// ModelSelectionListItem wraps ModelOption to implement ListItem interface
type ModelSelectionListItem struct {
	option ModelOption
}

// Title returns the title of the model item
func (m *ModelSelectionListItem) Title() string {
	return m.option.DisplayName
}

// Description returns formatted metadata for the model
func (m *ModelSelectionListItem) Description() string {
	contextStr := fmt.Sprintf("%dK ctx", m.option.ContextWindow/1000)
	costStr := fmt.Sprintf("$%.2f/$%.2f per 1M", m.option.InputCost, m.option.OutputCost)
	provider := strings.ToTitle(m.option.Provider)
	return fmt.Sprintf("%s | %s | %s", provider, contextStr, costStr)
}

// FilterValue returns the value to use for filtering
func (m *ModelSelectionListItem) FilterValue() string {
	return m.option.DisplayName + " " + m.option.ModelID + " " + m.option.Provider
}

// GetOption returns the underlying ModelOption
func (m *ModelSelectionListItem) GetOption() ModelOption {
	return m.option
}

// ModelSelectionDialog is a dialog for selecting AI models
type ModelSelectionDialog struct {
	*ListDialog
	lastSelected *ModelSelectionResult
	configPath   string
}

// NewModelSelectionDialog creates a new model selection dialog
func NewModelSelectionDialog(width, height int, configPath string) *ModelSelectionDialog {
	// Load available models
	options := loadAvailableModels()
	
	// Convert to ListItem interface
	items := make([]ListItem, len(options))
	for i, opt := range options {
		items[i] = &ModelSelectionListItem{option: opt}
	}

	// Create base list dialog
	listDialog := NewListDialog("Select AI Model", width, height, items)
	listDialog.showDescription = true

	dialog := &ModelSelectionDialog{
		ListDialog: listDialog,
		configPath: configPath,
	}

	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Enter", Label: "Select"},
		ShortcutHint{Key: "Esc", Label: "Cancel"},
	)

	// Load last selection if available
	dialog.loadLastSelection()

	return dialog
}

// loadAvailableModels loads available models from config system
func loadAvailableModels() []ModelOption {
	models := []ModelOption{
		{
			Provider:      "anthropic",
			ModelID:       "claude-3-7-sonnet-20250219",
			DisplayName:   "Claude 3.7 Sonnet",
			ContextWindow: 200000,
			InputCost:     3.0,
			OutputCost:    15.0,
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-3-5-sonnet-20241022",
			DisplayName:   "Claude 3.5 Sonnet",
			ContextWindow: 200000,
			InputCost:     3.0,
			OutputCost:    15.0,
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-3-5-haiku-20241022",
			DisplayName:   "Claude 3.5 Haiku",
			ContextWindow: 200000,
			InputCost:     0.8,
			OutputCost:    4.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4o",
			DisplayName:   "GPT-4o",
			ContextWindow: 128000,
			InputCost:     5.0,
			OutputCost:    15.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4-turbo",
			DisplayName:   "GPT-4 Turbo",
			ContextWindow: 128000,
			InputCost:     10.0,
			OutputCost:    30.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4o-mini",
			DisplayName:   "GPT-4o Mini",
			ContextWindow: 128000,
			InputCost:     0.15,
			OutputCost:    0.6,
		},
		{
			Provider:      "perplexity",
			ModelID:       "sonar-pro",
			DisplayName:   "Perplexity Sonar Pro",
			ContextWindow: 200000,
			InputCost:     20.0,
			OutputCost:    20.0,
		},
		{
			Provider:      "google",
			ModelID:       "gemini-2.0-flash",
			DisplayName:   "Gemini 2.0 Flash",
			ContextWindow: 1000000,
			InputCost:     0.075,
			OutputCost:    0.3,
		},
		{
			Provider:      "google",
			ModelID:       "gemini-1.5-pro",
			DisplayName:   "Gemini 1.5 Pro",
			ContextWindow: 1000000,
			InputCost:     1.25,
			OutputCost:    5.0,
		},
	}
	return models
}

// loadLastSelection loads the last selected model from config
func (d *ModelSelectionDialog) loadLastSelection() {
	// Read the config file
	data, err := os.ReadFile(d.configPath)
	if err != nil {
		return
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	// If there's a saved selection, find and select it
	if cfg.ModelProvider != "" && cfg.ModelName != "" {
		d.lastSelected = &ModelSelectionResult{
			Provider: cfg.ModelProvider,
			ModelID:  cfg.ModelName,
		}

		// Try to find and pre-select this model in the list
		for i, item := range d.items {
			if listItem, ok := item.(*ModelSelectionListItem); ok {
				opt := listItem.GetOption()
				if opt.Provider == cfg.ModelProvider && opt.ModelID == cfg.ModelName {
					d.selectedIndex = i
					break
				}
			}
		}
	}
}

// GetSelectedModel returns the selected model with provider
func (d *ModelSelectionDialog) GetSelectedModel() *ModelSelectionResult {
	if d.selectedIndex < 0 || d.selectedIndex >= len(d.items) {
		return nil
	}

	item, ok := d.items[d.selectedIndex].(*ModelSelectionListItem)
	if !ok {
		return nil
	}

	opt := item.GetOption()
	return &ModelSelectionResult{
		Provider: opt.Provider,
		ModelID:  opt.ModelID,
	}
}

// WriteSelectionToConfig persists the selected model to config file
func (d *ModelSelectionDialog) WriteSelectionToConfig() error {
	result := d.GetSelectedModel()
	if result == nil {
		return fmt.Errorf("no model selected")
	}

	// Read existing config
	data, err := os.ReadFile(d.configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg config.Config
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Update model selection
	cfg.ModelProvider = result.Provider
	cfg.ModelName = result.ModelID

	// Write back to file
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(d.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(d.configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// NewModelSelectionDialogSimple creates a model selection dialog with default settings
// This is a convenience constructor for callers that don't need custom dimensions
func NewModelSelectionDialogSimple() *ModelSelectionDialog {
	return NewModelSelectionDialog(60, 20, "")
}


