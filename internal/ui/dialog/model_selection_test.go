package dialog

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
)

func TestLoadAvailableModels(t *testing.T) {
	models := loadAvailableModels()

	if len(models) == 0 {
		t.Fatal("Expected models to be loaded, got empty list")
	}

	// Check that we have expected providers
	providers := make(map[string]bool)
	for _, m := range models {
		providers[m.Provider] = true
		if m.DisplayName == "" {
			t.Errorf("Model %s has empty DisplayName", m.ModelID)
		}
		if m.ModelID == "" {
			t.Error("Model has empty ModelID")
		}
		if m.ContextWindow <= 0 {
			t.Errorf("Model %s has invalid ContextWindow: %d", m.ModelID, m.ContextWindow)
		}
		if m.InputCost < 0 || m.OutputCost < 0 {
			t.Errorf("Model %s has negative costs: input=%f, output=%f", m.ModelID, m.InputCost, m.OutputCost)
		}
	}

	expectedProviders := []string{"anthropic", "openai", "perplexity", "google"}
	for _, provider := range expectedProviders {
		if !providers[provider] {
			t.Errorf("Expected provider %s not found in loaded models", provider)
		}
	}
}

func TestModelSelectionListItem(t *testing.T) {
	option := ModelOption{
		Provider:      "anthropic",
		ModelID:       "claude-3-5-sonnet-20241022",
		DisplayName:   "Claude 3.5 Sonnet",
		ContextWindow: 200000,
		InputCost:     3.0,
		OutputCost:    15.0,
	}

	item := &ModelSelectionListItem{option: option}

	if item.Title() != "Claude 3.5 Sonnet" {
		t.Errorf("Expected Title() to return 'Claude 3.5 Sonnet', got '%s'", item.Title())
	}

	desc := item.Description()
	if desc == "" {
		t.Error("Expected Description() to return non-empty string")
	}

	// Check that description contains expected parts
	expectedParts := []string{"ANTHROPIC", "200K", "3.00", "15.00"}
	for _, part := range expectedParts {
		if !containsString(desc, part) {
			t.Errorf("Expected Description() to contain '%s', got '%s'", part, desc)
		}
	}

	filter := item.FilterValue()
	if !containsString(filter, "Claude") || !containsString(filter, "claude-3-5-sonnet") || !containsString(filter, "anthropic") {
		t.Errorf("FilterValue() missing expected values: %s", filter)
	}

	if item.GetOption() != option {
		t.Error("GetOption() did not return the original option")
	}
}

func TestNewModelSelectionDialog(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	if dialog == nil {
		t.Fatal("NewModelSelectionDialog returned nil")
	}

	if dialog.ListDialog == nil {
		t.Fatal("ListDialog not initialized")
	}

	if len(dialog.items) == 0 {
		t.Fatal("Dialog has no items")
	}

	if dialog.showDescription != true {
		t.Error("showDescription should be true for model selection")
	}
}

func TestGetSelectedModel(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)
	dialog.selectedIndex = 0

	result := dialog.GetSelectedModel()
	if result == nil {
		t.Fatal("GetSelectedModel returned nil")
	}

	if result.Provider == "" || result.ModelID == "" {
		t.Error("GetSelectedModel returned incomplete result")
	}
}

func TestWriteSelectionToConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)
	dialog.selectedIndex = 0

	err := dialog.WriteSelectionToConfig()
	if err != nil {
		t.Fatalf("WriteSelectionToConfig failed: %v", err)
	}

	// Verify config was written
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read written config: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if cfg.ModelProvider == "" || cfg.ModelName == "" {
		t.Error("Config does not have model selection")
	}

	// Verify it matches what GetSelectedModel returns
	selected := dialog.GetSelectedModel()
	if cfg.ModelProvider != selected.Provider || cfg.ModelName != selected.ModelID {
		t.Error("Written config does not match selected model")
	}
}

func TestLoadLastSelection(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	// Create a config with a pre-selected model
	cfg := config.Config{
		ModelProvider: "anthropic",
		ModelName:     "claude-3-5-sonnet-20241022",
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	// Create a new dialog and verify it loads the selection
	dialog := NewModelSelectionDialog(60, 20, configPath)

	if dialog.lastSelected == nil {
		t.Fatal("loadLastSelection did not load the saved selection")
	}

	if dialog.lastSelected.Provider != "anthropic" || dialog.lastSelected.ModelID != "claude-3-5-sonnet-20241022" {
		t.Error("loadLastSelection loaded incorrect selection")
	}

	// Verify the dialog's selected index matches the loaded model
	if dialog.selectedIndex < 0 {
		t.Error("selectedIndex should be set to a valid value after loading")
	}
}

func TestModelSelectionDialogKeyboardNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	// Test that we can navigate through the list
	if len(dialog.items) > 1 {
		// Move down would be handled through Update() with arrow key messages
		// This is a basic test that the structure supports navigation
		if dialog.selectedIndex >= len(dialog.items) {
			t.Error("Selected index out of range")
		}
	}
}

func TestModelSelectionDialogVariousTerminalSizes(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	sizes := []struct {
		width  int
		height int
	}{
		{40, 15},
		{60, 20},
		{120, 40},
		{20, 10},
	}

	for _, size := range sizes {
		dialog := NewModelSelectionDialog(size.width, size.height, configPath)
		if dialog == nil {
			t.Errorf("Failed to create dialog with size %dx%d", size.width, size.height)
		}
		if dialog.width != size.width || dialog.height != size.height {
			t.Errorf("Dialog dimensions not set correctly: expected %dx%d, got %dx%d",
				size.width, size.height, dialog.width, dialog.height)
		}
	}
}

func TestModelSelectionResultStructure(t *testing.T) {
	result := &ModelSelectionResult{
		Provider: "anthropic",
		ModelID:  "claude-3-5-sonnet-20241022",
	}

	if result.Provider != "anthropic" {
		t.Error("Provider field not working correctly")
	}

	if result.ModelID != "claude-3-5-sonnet-20241022" {
		t.Error("ModelID field not working correctly")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
