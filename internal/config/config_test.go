package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadState tests loading UI state from disk
func TestLoadState(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "tui-state.json")

	t.Run("load non-existent state returns defaults", func(t *testing.T) {
		state, err := LoadState(statePath)
		if err != nil {
			t.Errorf("Expected no error for non-existent file, got: %v", err)
		}
		if state == nil {
			t.Fatal("Expected default state, got nil")
		}
		if state.ViewMode != "tree" {
			t.Errorf("Expected default ViewMode 'tree', got: %s", state.ViewMode)
		}
		if state.FocusedPanel != "taskList" {
			t.Errorf("Expected default FocusedPanel 'taskList', got: %s", state.FocusedPanel)
		}
		if !state.ShowDetailsPanel {
			t.Error("Expected ShowDetailsPanel to be true by default")
		}
	})

	t.Run("save and load state round-trip", func(t *testing.T) {
		// Create test state
		testState := &UIState{
			ExpandedIDs:      []string{"1", "1.1", "2"},
			SelectedID:       "1.2",
			ViewMode:         "list",
			FocusedPanel:     "details",
			ShowDetailsPanel: false,
			ShowLogPanel:     true,
			PanelHeights:     map[string]int{"tasks": 50, "details": 30},
		}

		// Save state
		err := SaveState(statePath, testState)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Load state
		loadedState, err := LoadState(statePath)
		if err != nil {
			t.Fatalf("Failed to load state: %v", err)
		}

		// Verify all fields match
		if len(loadedState.ExpandedIDs) != len(testState.ExpandedIDs) {
			t.Errorf("ExpandedIDs length mismatch: expected %d, got %d",
				len(testState.ExpandedIDs), len(loadedState.ExpandedIDs))
		}
		for i, id := range testState.ExpandedIDs {
			if loadedState.ExpandedIDs[i] != id {
				t.Errorf("ExpandedIDs[%d] mismatch: expected %s, got %s",
					i, id, loadedState.ExpandedIDs[i])
			}
		}

		if loadedState.SelectedID != testState.SelectedID {
			t.Errorf("SelectedID mismatch: expected %s, got %s",
				testState.SelectedID, loadedState.SelectedID)
		}

		if loadedState.ViewMode != testState.ViewMode {
			t.Errorf("ViewMode mismatch: expected %s, got %s",
				testState.ViewMode, loadedState.ViewMode)
		}

		if loadedState.FocusedPanel != testState.FocusedPanel {
			t.Errorf("FocusedPanel mismatch: expected %s, got %s",
				testState.FocusedPanel, loadedState.FocusedPanel)
		}

		if loadedState.ShowDetailsPanel != testState.ShowDetailsPanel {
			t.Errorf("ShowDetailsPanel mismatch: expected %v, got %v",
				testState.ShowDetailsPanel, loadedState.ShowDetailsPanel)
		}

		if loadedState.ShowLogPanel != testState.ShowLogPanel {
			t.Errorf("ShowLogPanel mismatch: expected %v, got %v",
				testState.ShowLogPanel, loadedState.ShowLogPanel)
		}
	})

	t.Run("save state with empty path returns error", func(t *testing.T) {
		state := &UIState{}
		err := SaveState("", state)
		if err == nil {
			t.Error("Expected error for empty state path, got nil")
		}
	})

	t.Run("load state with empty path returns error", func(t *testing.T) {
		_, err := LoadState("")
		if err == nil {
			t.Error("Expected error for empty state path, got nil")
		}
	})
}

// TestSaveState tests saving UI state to disk
func TestSaveState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("save creates directory if missing", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "subdir", "state.json")
		state := &UIState{
			ExpandedIDs: []string{"1"},
			SelectedID:  "1",
		}

		err := SaveState(statePath, state)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(statePath); os.IsNotExist(err) {
			t.Error("State file was not created")
		}
	})

	t.Run("save creates valid JSON", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "valid.json")
		state := &UIState{
			ExpandedIDs:      []string{"1", "2"},
			SelectedID:       "1",
			ViewMode:         "tree",
			FocusedPanel:     "taskList",
			ShowDetailsPanel: true,
			ShowLogPanel:     false,
		}

		err := SaveState(statePath, state)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Read and validate JSON
		data, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("Failed to read state file: %v", err)
		}

		var loaded UIState
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Errorf("Saved state is not valid JSON: %v", err)
		}
	})
}

// TestMergeConfigFile tests configuration merging
func TestMergeConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("merge key bindings", func(t *testing.T) {
		// Create base config
		baseConfig := &Config{
			KeyBindings: map[string]string{
				"quit": "q",
				"help": "?",
			},
		}

		// Create config file with overrides
		overrideConfig := map[string]interface{}{
			"keyBindings": map[string]string{
				"quit":    "Q", // Override
				"refresh": "r", // New key
			},
		}

		configPath := filepath.Join(tmpDir, "override.json")
		data, _ := json.MarshalIndent(overrideConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Merge config
		err := mergeConfigFile(baseConfig, configPath)
		if err != nil {
			t.Fatalf("Failed to merge config: %v", err)
		}

		// Verify merged result
		if baseConfig.KeyBindings["quit"] != "Q" {
			t.Errorf("Expected quit key to be 'Q', got %s", baseConfig.KeyBindings["quit"])
		}
		if baseConfig.KeyBindings["help"] != "?" {
			t.Errorf("Expected help key to be preserved as '?', got %s", baseConfig.KeyBindings["help"])
		}
		if baseConfig.KeyBindings["refresh"] != "r" {
			t.Errorf("Expected new refresh key to be 'r', got %s", baseConfig.KeyBindings["refresh"])
		}
	})

	t.Run("merge theme colors", func(t *testing.T) {
		baseConfig := &Config{
			Theme: ThemeConfig{
				PrimaryColor: "#7d56f4",
				SuccessColor: "#04B575",
			},
		}

		overrideConfig := map[string]interface{}{
			"theme": map[string]string{
				"primaryColor": "#FF0000",
			},
		}

		configPath := filepath.Join(tmpDir, "theme.json")
		data, _ := json.MarshalIndent(overrideConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		err := mergeConfigFile(baseConfig, configPath)
		if err != nil {
			t.Fatalf("Failed to merge config: %v", err)
		}

		if baseConfig.Theme.PrimaryColor != "#FF0000" {
			t.Errorf("Expected primary color to be overridden, got %s", baseConfig.Theme.PrimaryColor)
		}
		if baseConfig.Theme.SuccessColor != "#04B575" {
			t.Errorf("Expected success color to be preserved, got %s", baseConfig.Theme.SuccessColor)
		}
	})
}

// TestGetAPIKeys tests API key retrieval from environment
func TestGetAPIKeys(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{EnvOpenAIKey, EnvAnthropicKey, EnvPerplexityKey}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore original env vars after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("get configured API keys", func(t *testing.T) {
		// Set test API keys
		os.Setenv(EnvOpenAIKey, "test-openai-key")
		os.Setenv(EnvAnthropicKey, "test-anthropic-key")
		os.Unsetenv(EnvPerplexityKey)

		keys := GetAPIKeys()

		if keys[EnvOpenAIKey] != "test-openai-key" {
			t.Errorf("Expected OpenAI key to be 'test-openai-key', got %s", keys[EnvOpenAIKey])
		}
		if keys[EnvAnthropicKey] != "test-anthropic-key" {
			t.Errorf("Expected Anthropic key to be 'test-anthropic-key', got %s", keys[EnvAnthropicKey])
		}
		if _, exists := keys[EnvPerplexityKey]; exists {
			t.Error("Expected Perplexity key to not be present")
		}
	})

	t.Run("no API keys configured", func(t *testing.T) {
		// Unset all keys
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		keys := GetAPIKeys()
		if len(keys) != 0 {
			t.Errorf("Expected no API keys, got %d", len(keys))
		}
	})
}

// TestHasAPIKey tests checking for specific API keys
func TestHasAPIKey(t *testing.T) {
	// Save original env var
	originalValue := os.Getenv(EnvOpenAIKey)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(EnvOpenAIKey)
		} else {
			os.Setenv(EnvOpenAIKey, originalValue)
		}
	}()

	t.Run("key exists", func(t *testing.T) {
		os.Setenv(EnvOpenAIKey, "test-key")
		if !HasAPIKey(EnvOpenAIKey) {
			t.Error("Expected HasAPIKey to return true")
		}
	})

	t.Run("key does not exist", func(t *testing.T) {
		os.Unsetenv(EnvOpenAIKey)
		if HasAPIKey(EnvOpenAIKey) {
			t.Error("Expected HasAPIKey to return false")
		}
	})
}

// TestGetConfiguredProviders tests getting list of configured providers
func TestGetConfiguredProviders(t *testing.T) {
	// Save original env vars
	envVars := []string{EnvOpenAIKey, EnvAnthropicKey, EnvGoogleKey}
	originalEnv := make(map[string]string)
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore original env vars after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("multiple providers configured", func(t *testing.T) {
		// Unset all first
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		// Set specific providers
		os.Setenv(EnvOpenAIKey, "test-key-1")
		os.Setenv(EnvAnthropicKey, "test-key-2")

		providers := GetConfiguredProviders()

		// Should have openai and anthropic
		hasOpenAI := false
		hasAnthropic := false
		for _, provider := range providers {
			if provider == "openai" {
				hasOpenAI = true
			}
			if provider == "anthropic" {
				hasAnthropic = true
			}
		}

		if !hasOpenAI {
			t.Error("Expected OpenAI in configured providers")
		}
		if !hasAnthropic {
			t.Error("Expected Anthropic in configured providers")
		}
	})

	t.Run("no providers configured", func(t *testing.T) {
		// Unset all keys
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		providers := GetConfiguredProviders()
		
		// Check that our unset providers are not in the list
		for _, provider := range providers {
			if provider == "openai" || provider == "anthropic" || provider == "google" {
				t.Errorf("Expected %s to not be in configured providers", provider)
			}
		}
	})
}

// TestGetEnv tests environment variable retrieval with fallback
func TestGetEnv(t *testing.T) {
	testKey := "TEST_CONFIG_VAR"
	
	// Save original value
	originalValue := os.Getenv(testKey)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(testKey)
		} else {
			os.Setenv(testKey, originalValue)
		}
	}()

	t.Run("returns env var when set", func(t *testing.T) {
		os.Setenv(testKey, "test-value")
		result := GetEnv(testKey, "fallback")
		if result != "test-value" {
			t.Errorf("Expected 'test-value', got %s", result)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		os.Unsetenv(testKey)
		result := GetEnv(testKey, "fallback")
		if result != "fallback" {
			t.Errorf("Expected 'fallback', got %s", result)
		}
	})
}
