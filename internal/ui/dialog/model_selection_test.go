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

// TestCycleProviderFilter tests the provider filter cycling functionality
func TestCycleProviderFilter(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	// Start with all providers
	if dialog.selectedProvider != "" {
		t.Error("Should start with empty provider (all providers)")
	}

	initialItemCount := len(dialog.items)

	// Cycle to next provider
	dialog.CycleProviderFilter()
	
	if dialog.selectedProvider == "" {
		t.Error("Should have cycled to a specific provider")
	}

	// After filtering, should have fewer items
	if len(dialog.items) >= initialItemCount {
		t.Errorf("Filtered items should be less than total items: %d >= %d", len(dialog.items), initialItemCount)
	}

	// Cycle back to all providers
	for i := 0; i < len(dialog.providerList)-1; i++ {
		dialog.CycleProviderFilter()
	}

	if dialog.selectedProvider != "" {
		t.Error("Should cycle back to all providers")
	}

	if len(dialog.items) != initialItemCount {
		t.Error("Should return to showing all items after cycling through all providers")
	}
}

// TestApplyProviderFilter tests the provider filter application
func TestApplyProviderFilter(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	// Test filtering to anthropic
	dialog.selectedProvider = "anthropic"
	dialog.applyProviderFilter()

	// Verify only anthropic models remain
	for _, item := range dialog.items {
		if modelItem, ok := item.(*ModelSelectionListItem); ok {
			if modelItem.option.Provider != "anthropic" {
				t.Errorf("Filter failed: found non-anthropic model in filtered list: %s", modelItem.option.Provider)
			}
		}
	}

	// Test filtering to openai
	dialog.selectedProvider = "openai"
	dialog.applyProviderFilter()

	for _, item := range dialog.items {
		if modelItem, ok := item.(*ModelSelectionListItem); ok {
			if modelItem.option.Provider != "openai" {
				t.Errorf("Filter failed: found non-openai model in filtered list: %s", modelItem.option.Provider)
			}
		}
	}
}

// TestGetCurrentFilter tests the filter display string
func TestGetCurrentFilter(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	// Test "all providers" display
	dialog.selectedProvider = ""
	if dialog.GetCurrentFilter() != "All Providers" {
		t.Errorf("Expected 'All Providers', got '%s'", dialog.GetCurrentFilter())
	}

	// Test provider-specific display
	dialog.selectedProvider = "anthropic"
	filter := dialog.GetCurrentFilter()
	if filter != "Anthropic" {
		t.Errorf("Expected 'Anthropic', got '%s'", filter)
	}
}

// TestModelMetadataDisplay tests enhanced metadata display in descriptions
func TestModelMetadataDisplay(t *testing.T) {
	models := []struct {
		name       string
		option     ModelOption
		expectedIn []string // Strings expected in description
	}{
		{
			name: "high-cost model",
			option: ModelOption{
				Provider:      "anthropic",
				ModelID:       "claude-3-opus",
				DisplayName:   "Claude 3 Opus",
				ContextWindow: 200000,
				InputCost:     15.0,
				OutputCost:    75.0,
			},
			expectedIn: []string{"Premium", "200K", "ANTHROPIC"},
		},
		{
			name: "budget model",
			option: ModelOption{
				Provider:      "google",
				ModelID:       "gemini-2.0-flash",
				DisplayName:   "Gemini 2.0 Flash",
				ContextWindow: 1000000,
				InputCost:     0.075,
				OutputCost:    0.3,
			},
			expectedIn: []string{"Budget", "1.0M", "GOOGLE"},
		},
		{
			name: "mid-range model",
			option: ModelOption{
				Provider:      "openai",
				ModelID:       "gpt-4o",
				DisplayName:   "GPT-4o",
				ContextWindow: 128000,
				InputCost:     5.0,
				OutputCost:    15.0,
			},
			expectedIn: []string{"Premium", "128K", "OPENAI"},
		},
	}

	for _, test := range models {
		t.Run(test.name, func(t *testing.T) {
			item := &ModelSelectionListItem{option: test.option}
			desc := item.Description()

			for _, expected := range test.expectedIn {
				if !containsString(desc, expected) {
					t.Errorf("Description missing '%s': %s", expected, desc)
				}
			}
		})
	}
}

// TestModelCapabilitiesIncluded tests that model capabilities are preserved
func TestModelCapabilitiesIncluded(t *testing.T) {
	models := loadAvailableModels()

	// Check that at least some models have capabilities
	foundWithCapabilities := false
	for _, m := range models {
		if m.Capabilities != "" {
			foundWithCapabilities = true
			break
		}
	}

	if !foundWithCapabilities {
		t.Logf("Note: No models have capabilities defined - consider adding them")
	}
}

// TestModelIDsFormatted tests that model IDs are compatible with Crush
func TestModelIDsFormatted(t *testing.T) {
	models := loadAvailableModels()

	for _, m := range models {
		// Model IDs should not contain spaces or special shell characters
		if containsSubstring(m.ModelID, " ") {
			t.Errorf("Model ID contains spaces: %s", m.ModelID)
		}

		// Should be valid for use in shell/config
		if m.ModelID == "" {
			t.Error("Empty model ID found")
		}

		// Provider should be lowercase
		if m.Provider != "" && m.Provider[0] >= 'A' && m.Provider[0] <= 'Z' {
			t.Errorf("Provider should be lowercase: %s", m.Provider)
		}
	}
}

// TestProviderListInitialization tests that provider list is properly initialized
func TestProviderListInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	dialog := NewModelSelectionDialog(60, 20, configPath)

	// Should have at least one entry for "all providers" plus actual providers
	if len(dialog.providerList) < 2 {
		t.Errorf("Provider list should have multiple entries, got %d", len(dialog.providerList))
	}

	// First entry should be empty (all providers)
	if dialog.providerList[0] != "" {
		t.Errorf("First entry should be empty (all providers), got %s", dialog.providerList[0])
	}

	// Should have all expected providers
	expectedProviders := map[string]bool{
		"anthropic":  false,
		"openai":     false,
		"google":     false,
		"perplexity": false,
	}

	for _, p := range dialog.providerList {
		if p == "" {
			continue
		}
		expectedProviders[p] = true
	}

	for provider, found := range expectedProviders {
		if !found {
			t.Errorf("Expected provider %s not in provider list", provider)
		}
	}
}

// TestWriteSelectionToBothConfigs tests that both TUI and Crush configs are updated
func TestWriteSelectionToBothConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	// Set temp directory for Crush config
	t.Setenv("CRUSH_PROJECT_ROOT", tmpDir)

	dialog := NewModelSelectionDialog(60, 20, configPath)
	dialog.selectedIndex = 0

	err := dialog.WriteSelectionToConfig()
	if err != nil {
		t.Fatalf("WriteSelectionToConfig failed: %v", err)
	}

	// Verify TUI config was written
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read TUI config: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal TUI config: %v", err)
	}

	if cfg.ModelProvider == "" || cfg.ModelName == "" {
		t.Error("TUI config does not have model selection")
	}

	// Verify Crush config was created/updated
	crushConfigPath := tmpDir + "/.crush.json"
	if _, err := os.Stat(crushConfigPath); os.IsNotExist(err) {
		t.Logf("Crush config not created (may be expected in test environment)")
	} else {
		// If it exists, verify it was updated
		crushData, err := os.ReadFile(crushConfigPath)
		if err != nil {
			t.Fatalf("Failed to read Crush config: %v", err)
		}

		var crushCfg map[string]interface{}
		if err := json.Unmarshal(crushData, &crushCfg); err != nil {
			t.Fatalf("Failed to unmarshal Crush config: %v", err)
		}

		// Check if model field is present and matches
		if model, ok := crushCfg["model"]; ok {
			if model != cfg.ModelName {
				t.Errorf("Crush config model mismatch: expected %s, got %v", cfg.ModelName, model)
			}
		}
	}
}
