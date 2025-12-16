package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadModelConfig tests loading model configuration
func TestLoadModelConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "model-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster dir: %v", err)
	}

	// Change to temp directory for the test
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalCwd)

	os.Chdir(tmpDir)

	t.Run("load defaults when config file not found", func(t *testing.T) {
		provider, modelName, err := LoadModelConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider != getDefaultModelProvider() {
			t.Errorf("Expected default provider %s, got %s", getDefaultModelProvider(), provider)
		}
		if modelName != getDefaultModelName() {
			t.Errorf("Expected default modelName %s, got %s", getDefaultModelName(), modelName)
		}
	})

	t.Run("load config with model settings", func(t *testing.T) {
		config := map[string]interface{}{
			"modelProvider": "openai",
			"modelName":     "gpt-4",
		}
		configPath := filepath.Join(tmDir, "config.json")
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		provider, modelName, err := LoadModelConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider != "openai" {
			t.Errorf("Expected provider 'openai', got %s", provider)
		}
		if modelName != "gpt-4" {
			t.Errorf("Expected modelName 'gpt-4', got %s", modelName)
		}
	})

	t.Run("load config without model settings returns defaults", func(t *testing.T) {
		config := map[string]interface{}{
			"otherSetting": "value",
		}
		configPath := filepath.Join(tmDir, "config.json")
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		provider, modelName, err := LoadModelConfig()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if provider != getDefaultModelProvider() {
			t.Errorf("Expected default provider, got %s", provider)
		}
		if modelName != getDefaultModelName() {
			t.Errorf("Expected default modelName, got %s", modelName)
		}
	})

	t.Run("load invalid JSON config returns defaults", func(t *testing.T) {
		configPath := filepath.Join(tmDir, "config.json")
		if err := os.WriteFile(configPath, []byte("invalid json"), 0600); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		provider, modelName, _ := LoadModelConfig()
		// Should not error, should return defaults
		if provider != getDefaultModelProvider() {
			t.Errorf("Expected default provider, got %s", provider)
		}
		if modelName != getDefaultModelName() {
			t.Errorf("Expected default modelName, got %s", modelName)
		}
	})
}

// TestSaveModelConfig tests saving model configuration
func TestSaveModelConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "model-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster dir: %v", err)
	}

	// Change to temp directory for the test
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalCwd)

	os.Chdir(tmpDir)

	t.Run("save creates new config file", func(t *testing.T) {
		err := SaveModelConfig("anthropic", "claude-3-5-sonnet-20241022")
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		configPath := filepath.Join(tmDir, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		// Verify content
		data, _ := os.ReadFile(configPath)
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			t.Errorf("Saved config is not valid JSON: %v", err)
		}

		if provider, ok := config["modelProvider"].(string); !ok || provider != "anthropic" {
			t.Error("modelProvider not saved correctly")
		}
		if modelName, ok := config["modelName"].(string); !ok || modelName != "claude-3-5-sonnet-20241022" {
			t.Error("modelName not saved correctly")
		}
	})

	t.Run("save updates existing config", func(t *testing.T) {
		// First save
		err := SaveModelConfig("openai", "gpt-4")
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Update config
		err = SaveModelConfig("anthropic", "claude-3-opus-20240229")
		if err != nil {
			t.Fatalf("Failed to update config: %v", err)
		}

		// Verify updated content
		configPath := filepath.Join(tmDir, "config.json")
		data, _ := os.ReadFile(configPath)
		var config map[string]interface{}
		json.Unmarshal(data, &config)

		if provider, ok := config["modelProvider"].(string); !ok || provider != "anthropic" {
			t.Error("modelProvider was not updated")
		}
		if modelName, ok := config["modelName"].(string); !ok || modelName != "claude-3-opus-20240229" {
			t.Error("modelName was not updated")
		}
	})

	t.Run("save preserves other config values", func(t *testing.T) {
		// Create initial config with other values
		configPath := filepath.Join(tmDir, "config.json")
		config := map[string]interface{}{
			"otherSetting":  "value",
			"modelProvider": "old",
			"modelName":     "old",
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, data, 0600)

		// Save new model config
		err := SaveModelConfig("google", "gemini-pro")
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify other values preserved
		data, _ = os.ReadFile(configPath)
		var savedConfig map[string]interface{}
		json.Unmarshal(data, &savedConfig)

		if otherVal, ok := savedConfig["otherSetting"].(string); !ok || otherVal != "value" {
			t.Error("Other config values were not preserved")
		}
		if provider, ok := savedConfig["modelProvider"].(string); !ok || provider != "google" {
			t.Error("modelProvider was not updated")
		}
	})

	t.Run("save rejects empty provider", func(t *testing.T) {
		err := SaveModelConfig("", "claude-3-5-sonnet-20241022")
		if err == nil {
			t.Error("Expected error for empty provider")
		}
	})

	t.Run("save rejects empty modelName", func(t *testing.T) {
		err := SaveModelConfig("anthropic", "")
		if err == nil {
			t.Error("Expected error for empty modelName")
		}
	})
}

// TestValidateModelSelection tests model selection validation
func TestValidateModelSelection(t *testing.T) {
	t.Run("validate known model", func(t *testing.T) {
		isValid := ValidateModelSelection("anthropic", "claude-3-5-sonnet-20241022")
		if !isValid {
			t.Error("Expected valid model to return true")
		}
	})

	t.Run("validate unknown model", func(t *testing.T) {
		isValid := ValidateModelSelection("anthropic", "unknown-model")
		if isValid {
			t.Error("Expected unknown model to return false")
		}
	})

	t.Run("validate unknown provider", func(t *testing.T) {
		isValid := ValidateModelSelection("unknown-provider", "some-model")
		if isValid {
			t.Error("Expected unknown provider to return false")
		}
	})

	t.Run("validate empty provider", func(t *testing.T) {
		isValid := ValidateModelSelection("", "some-model")
		if isValid {
			t.Error("Expected empty provider to return false")
		}
	})

	t.Run("validate empty model", func(t *testing.T) {
		isValid := ValidateModelSelection("anthropic", "")
		if isValid {
			t.Error("Expected empty model to return false")
		}
	})
}

// TestListAvailableModels tests listing available models
func TestListAvailableModels(t *testing.T) {
	models := ListAvailableModels()

	t.Run("returns models map", func(t *testing.T) {
		if len(models) == 0 {
			t.Error("Expected non-empty models map")
		}
	})

	t.Run("anthropic models exist", func(t *testing.T) {
		if _, exists := models["anthropic"]; !exists {
			t.Error("Expected anthropic models to exist")
		}
		if len(models["anthropic"]) == 0 {
			t.Error("Expected anthropic to have models")
		}
	})

	t.Run("all models have required fields", func(t *testing.T) {
		for provider, providerModels := range models {
			for i, model := range providerModels {
				if model.Provider == "" {
					t.Errorf("Model %d in %s has empty Provider", i, provider)
				}
				if model.ModelName == "" {
					t.Errorf("Model %d in %s has empty ModelName", i, provider)
				}
				if model.ModelID == "" {
					t.Errorf("Model %d in %s has empty ModelID", i, provider)
				}
			}
		}
	})
}

// TestGetModelsByProvider tests getting models by provider
func TestGetModelsByProvider(t *testing.T) {
	t.Run("get existing provider models", func(t *testing.T) {
		models := GetModelsByProvider("anthropic")
		if len(models) == 0 {
			t.Error("Expected to get anthropic models")
		}
	})

	t.Run("get non-existing provider returns empty", func(t *testing.T) {
		models := GetModelsByProvider("non-existent")
		if len(models) != 0 {
			t.Error("Expected empty list for non-existent provider")
		}
	})
}

// TestGetAvailableProviders tests getting available providers
func TestGetAvailableProviders(t *testing.T) {
	providers := GetAvailableProviders()

	t.Run("returns multiple providers", func(t *testing.T) {
		if len(providers) < 3 {
			t.Errorf("Expected at least 3 providers, got %d", len(providers))
		}
	})

	t.Run("includes anthropic", func(t *testing.T) {
		found := false
		for _, p := range providers {
			if p == "anthropic" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected anthropic in available providers")
		}
	})
}

// TestModelConfigRoundTrip tests saving and loading model config
func TestModelConfigRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "model-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster dir: %v", err)
	}

	// Change to temp directory for the test
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalCwd)

	os.Chdir(tmpDir)

	t.Run("save and load round trip", func(t *testing.T) {
		// Save
		err := SaveModelConfig("mistral", "mistral-large-latest")
		if err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		// Load
		provider, modelName, err := LoadModelConfig()
		if err != nil {
			t.Fatalf("Failed to load: %v", err)
		}

		if provider != "mistral" {
			t.Errorf("Expected provider 'mistral', got %s", provider)
		}
		if modelName != "mistral-large-latest" {
			t.Errorf("Expected modelName 'mistral-large-latest', got %s", modelName)
		}
	})
}
