package dialog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agreen757/tm-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// ModelOption represents an available AI model with metadata
type ModelOption struct {
	Provider      string
	ModelID       string
	DisplayName   string
	ContextWindow int
	InputCost     float64 // per 1M tokens
	OutputCost    float64 // per 1M tokens
	Capabilities  string  // Optional: brief description of capabilities
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
	// Format context window
	contextWindow := m.option.ContextWindow
	var contextStr string
	if contextWindow >= 1000000 {
		contextStr = fmt.Sprintf("%.1fM", float64(contextWindow)/1000000)
	} else if contextWindow >= 1000 {
		contextStr = fmt.Sprintf("%dK", contextWindow/1000)
	} else {
		contextStr = fmt.Sprintf("%d", contextWindow)
	}

	// Calculate average cost and determine tier
	avgCost := (m.option.InputCost + m.option.OutputCost) / 2
	var costTier string
	if avgCost < 1 {
		costTier = "💰 Budget"
	} else if avgCost < 10 {
		costTier = "💵 Mid-range"
	} else {
		costTier = "💎 Premium"
	}

	// Format precise cost
	costStr := fmt.Sprintf("$%.2f/$%.2f", m.option.InputCost, m.option.OutputCost)
	provider := strings.ToTitle(m.option.Provider)

	return fmt.Sprintf("%s | %s tokens | %s | %s/1M", provider, contextStr, costTier, costStr)
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
	lastSelected     *ModelSelectionResult
	configPath       string
	allModels        []ModelOption   // Store all models for filtering
	providers        map[string]bool // Available providers
	selectedProvider string          // Currently selected provider filter (empty = all)
	providerList     []string        // Sorted list of providers for cycling
}

// NewModelSelectionDialog creates a new model selection dialog
func NewModelSelectionDialog(width, height int, configPath string) *ModelSelectionDialog {
	// Load available models
	allModels := loadAvailableModels()

	// Convert to ListItem interface
	items := make([]ListItem, len(allModels))
	for i, opt := range allModels {
		items[i] = &ModelSelectionListItem{option: opt}
	}

	// Create base list dialog
	listDialog := NewListDialog("Select AI Model", width, height, items)
	listDialog.showDescription = true

	// Build provider list
	providers := make(map[string]bool)
	for _, model := range allModels {
		providers[model.Provider] = true
	}

	providerList := make([]string, 0, len(providers))
	providerList = append(providerList, "") // Empty string means "all providers"
	for provider := range providers {
		providerList = append(providerList, provider)
	}
	// Sort providers for consistent ordering
	if len(providerList) > 1 {
		// Simple bubble sort for small list
		for i := 1; i < len(providerList); i++ {
			for j := i; j > 1 && providerList[j] < providerList[j-1]; j-- {
				providerList[j], providerList[j-1] = providerList[j-1], providerList[j]
			}
		}
	}

	dialog := &ModelSelectionDialog{
		ListDialog:       listDialog,
		configPath:       configPath,
		allModels:        allModels,
		providers:        providers,
		selectedProvider: "", // Start with all providers
		providerList:     providerList,
	}

	// Set dialog ID for result handling
	dialog.ListDialog.BaseDialog.ID = "model_selection_dialog"

	dialog.SetFooterHints(
		ShortcutHint{Key: "↑/↓", Label: "Navigate"},
		ShortcutHint{Key: "Tab", Label: "Filter provider"},
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
		// Anthropic Claude 4.x Models (January 2026 - Current Generation)
		{
			Provider:      "anthropic",
			ModelID:       "claude-opus-4-5-20251101",
			DisplayName:   "Claude Opus 4.5",
			ContextWindow: 200000,
			InputCost:     15.0,
			OutputCost:    75.0,
			Capabilities:  "Flagship • Maximum intelligence • Complex agents",
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-sonnet-4-5-20250929",
			DisplayName:   "Claude Sonnet 4.5",
			ContextWindow: 200000,
			InputCost:     3.0,
			OutputCost:    15.0,
			Capabilities:  "High performance • Coding & agents • Recommended",
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-haiku-4-5-20251001",
			DisplayName:   "Claude Haiku 4.5",
			ContextWindow: 200000,
			InputCost:     0.8,
			OutputCost:    4.0,
			Capabilities:  "Fastest • Cost-effective • High volume",
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-opus-4-1-20250805",
			DisplayName:   "Claude Opus 4.1",
			ContextWindow: 200000,
			InputCost:     15.0,
			OutputCost:    75.0,
			Capabilities:  "Stable 4.1 release • High intelligence",
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-sonnet-4-20250514",
			DisplayName:   "Claude Sonnet 4",
			ContextWindow: 200000,
			InputCost:     3.0,
			OutputCost:    15.0,
			Capabilities:  "Stable mid-tier • First Claude 4 release",
		},
		{
			Provider:      "anthropic",
			ModelID:       "claude-opus-4-20250514",
			DisplayName:   "Claude Opus 4",
			ContextWindow: 200000,
			InputCost:     15.0,
			OutputCost:    75.0,
			Capabilities:  "Original Opus 4 • Still active",
		},
		// OpenAI Models
		{
			Provider:      "openai",
			ModelID:       "gpt-4o",
			DisplayName:   "GPT-4o",
			ContextWindow: 128000,
			InputCost:     5.0,
			OutputCost:    15.0,
			Capabilities:  "Vision • Best OpenAI • Multimodal",
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4o-2024-11-20",
			DisplayName:   "GPT-4o (2024-11-20)",
			ContextWindow: 128000,
			InputCost:     5.0,
			OutputCost:    15.0,
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
			Provider:      "openai",
			ModelID:       "gpt-4-turbo",
			DisplayName:   "GPT-4 Turbo",
			ContextWindow: 128000,
			InputCost:     10.0,
			OutputCost:    30.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4-turbo-2024-04-09",
			DisplayName:   "GPT-4 Turbo (2024-04-09)",
			ContextWindow: 128000,
			InputCost:     10.0,
			OutputCost:    30.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-4",
			DisplayName:   "GPT-4",
			ContextWindow: 8192,
			InputCost:     30.0,
			OutputCost:    60.0,
		},
		{
			Provider:      "openai",
			ModelID:       "gpt-3.5-turbo",
			DisplayName:   "GPT-3.5 Turbo",
			ContextWindow: 16384,
			InputCost:     0.5,
			OutputCost:    1.5,
		},
		{
			Provider:      "openai",
			ModelID:       "o1-preview",
			DisplayName:   "o1 Preview",
			ContextWindow: 128000,
			InputCost:     15.0,
			OutputCost:    60.0,
		},
		{
			Provider:      "openai",
			ModelID:       "o1-mini",
			DisplayName:   "o1 Mini",
			ContextWindow: 128000,
			InputCost:     3.0,
			OutputCost:    12.0,
		},
		// Google Gemini Models
		{
			Provider:      "google",
			ModelID:       "gemini-2.0-flash",
			DisplayName:   "Gemini 2.0 Flash",
			ContextWindow: 1000000,
			InputCost:     0.075,
			OutputCost:    0.3,
			Capabilities:  "Budget-friendly • 1M context • Very fast",
		},
		{
			Provider:      "google",
			ModelID:       "gemini-2.0-flash-exp",
			DisplayName:   "Gemini 2.0 Flash Exp",
			ContextWindow: 1000000,
			InputCost:     0.075,
			OutputCost:    0.3,
		},
		{
			Provider:      "google",
			ModelID:       "gemini-1.5-pro",
			DisplayName:   "Gemini 1.5 Pro",
			ContextWindow: 2000000,
			InputCost:     1.25,
			OutputCost:    5.0,
		},
		{
			Provider:      "google",
			ModelID:       "gemini-1.5-flash",
			DisplayName:   "Gemini 1.5 Flash",
			ContextWindow: 1000000,
			InputCost:     0.075,
			OutputCost:    0.3,
		},
		{
			Provider:      "google",
			ModelID:       "gemini-pro",
			DisplayName:   "Gemini Pro",
			ContextWindow: 32000,
			InputCost:     0.5,
			OutputCost:    1.5,
		},
		// Perplexity Models
		{
			Provider:      "perplexity",
			ModelID:       "sonar-pro",
			DisplayName:   "Perplexity Sonar Pro",
			ContextWindow: 200000,
			InputCost:     20.0,
			OutputCost:    20.0,
			Capabilities:  "Web search • Research • Real-time data",
		},
		{
			Provider:      "perplexity",
			ModelID:       "sonar",
			DisplayName:   "Perplexity Sonar",
			ContextWindow: 127000,
			InputCost:     0.2,
			OutputCost:    0.2,
		},
		{
			Provider:      "perplexity",
			ModelID:       "sonar-reasoning-pro",
			DisplayName:   "Perplexity Sonar Reasoning Pro",
			ContextWindow: 200000,
			InputCost:     20.0,
			OutputCost:    60.0,
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

// CycleProviderFilter cycles to the next provider filter
func (d *ModelSelectionDialog) CycleProviderFilter() {
	// Find current position in provider list
	currentIndex := 0
	for i, p := range d.providerList {
		if p == d.selectedProvider {
			currentIndex = i
			break
		}
	}

	// Move to next provider
	nextIndex := (currentIndex + 1) % len(d.providerList)
	d.selectedProvider = d.providerList[nextIndex]

	// Reapply filter to items
	d.applyProviderFilter()

	// Reset selection to top of filtered list
	d.selectedIndex = 0
	d.offset = 0
}

// applyProviderFilter filters the items list based on selected provider
func (d *ModelSelectionDialog) applyProviderFilter() {
	if d.selectedProvider == "" {
		// Show all models
		items := make([]ListItem, len(d.allModels))
		for i, opt := range d.allModels {
			items[i] = &ModelSelectionListItem{option: opt}
		}
		d.items = items
	} else {
		// Filter to selected provider
		var filtered []ModelOption
		for _, model := range d.allModels {
			if model.Provider == d.selectedProvider {
				filtered = append(filtered, model)
			}
		}

		items := make([]ListItem, len(filtered))
		for i, opt := range filtered {
			items[i] = &ModelSelectionListItem{option: opt}
		}
		d.items = items
	}
	d.numElements = len(d.items)
}

// GetCurrentFilter returns the current provider filter display string
func (d *ModelSelectionDialog) GetCurrentFilter() string {
	if d.selectedProvider == "" {
		return "All Providers"
	}
	// Capitalize first letter
	if len(d.selectedProvider) > 0 {
		return strings.ToUpper(d.selectedProvider[:1]) + d.selectedProvider[1:]
	}
	return d.selectedProvider
}

// WriteSelectionToConfig persists the selected model to both TUI and Crush config files
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

	// Update model selection in TUI config
	cfg.ModelProvider = result.Provider
	cfg.ModelName = result.ModelID

	// Write back to TUI config file
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

	// Also update Crush config with the selected model
	// This is a non-fatal operation - log warning but don't fail if it errors
	// Use "large" as the default model type for the primary model selection
	if err := config.UpdateCrushModel(result.ModelID, result.Provider, "large"); err != nil {
		// Log warning but don't propagate error - TUI config was successfully saved
		fmt.Fprintf(os.Stderr, "Warning: failed to update Crush config: %v\n", err)
	}

	return nil
}

// Update processes keyboard input including Tab for filtering
func (d *ModelSelectionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Cycle through provider filters
			d.CycleProviderFilter()
			return d, nil
		}
	case ListSelectionMsg:
		// Handle confirmation via ListSelectionMsg from ListDialog
		if msg.SelectedItem != nil {
			if modelItem, ok := msg.SelectedItem.(*ModelSelectionListItem); ok {
				opt := modelItem.GetOption()
				// Emit DialogResultMsg when model is selected
				return d, func() tea.Msg {
					return DialogResultMsg{
						ID:     "model_selection_dialog",
						Button: "confirm",
						Value:  &ModelSelectionResult{Provider: opt.Provider, ModelID: opt.ModelID},
					}
				}
			}
		}
	}
	// Let parent handle other messages
	return d.ListDialog.Update(msg)
}

// View renders the dialog with provider filter indicator
func (d *ModelSelectionDialog) View() string {
	// Get the parent view (already includes title and footer)
	// The filter is accessible via GetCurrentFilter() if needed for customization
	return d.ListDialog.View()
}

// NewModelSelectionDialogSimple creates a model selection dialog with default settings
// This is a convenience constructor for callers that don't need custom dimensions
func NewModelSelectionDialogSimple() *ModelSelectionDialog {
	return NewModelSelectionDialog(60, 20, "")
}
