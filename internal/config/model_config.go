package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ModelConfig represents the model selection configuration
type ModelConfig struct {
	Provider  string `json:"modelProvider"`
	ModelName string `json:"modelName"`
}

// AvailableModel represents a model available from a provider
type AvailableModel struct {
	Provider  string
	ModelName string
	ModelID   string
}

// LoadModelConfig loads the model configuration from .taskmaster/config.json
// Returns default values if config file doesn't exist or is invalid
func LoadModelConfig() (provider string, modelName string, err error) {
	tmDir, err := findTaskMasterDir()
	if err != nil {
		// Return defaults if .taskmaster directory not found
		return getDefaultModelProvider(), getDefaultModelName(), nil
	}

	configPath := filepath.Join(tmDir, ".taskmaster", "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return defaults if file doesn't exist
		return getDefaultModelProvider(), getDefaultModelName(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return defaults if read fails
		return getDefaultModelProvider(), getDefaultModelName(), fmt.Errorf("failed to read config file: %w", err)
	}

	// Attempt to unmarshal the config
	var fullConfig map[string]interface{}
	if err := json.Unmarshal(data, &fullConfig); err != nil {
		// Return defaults if JSON is invalid
		return getDefaultModelProvider(), getDefaultModelName(), fmt.Errorf("invalid config format: %w", err)
	}

	// Extract model configuration if present
	if provider, ok := fullConfig["modelProvider"].(string); ok && provider != "" {
		if modelName, ok := fullConfig["modelName"].(string); ok && modelName != "" {
			return provider, modelName, nil
		}
	}

	// Return defaults if not found in config
	return getDefaultModelProvider(), getDefaultModelName(), nil
}

// SaveModelConfig saves the model selection to .taskmaster/config.json
// Creates the config file if it doesn't exist, updates it if it does
func SaveModelConfig(provider string, modelName string) error {
	if provider == "" || modelName == "" {
		return fmt.Errorf("provider and modelName cannot be empty")
	}

	tmDir, err := findTaskMasterDir()
	if err != nil {
		return fmt.Errorf("could not find .taskmaster directory: %w", err)
	}

	configPath := filepath.Join(tmDir, ".taskmaster", "config.json")

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config or start with empty object
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			// If unmarshal fails, start fresh
			config = make(map[string]interface{})
		}
	} else {
		// File doesn't exist, create new config
		config = make(map[string]interface{})
	}

	// Update model settings
	config["modelProvider"] = provider
	config["modelName"] = modelName

	// Marshal config to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file with user-only permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultModelProvider returns the default model provider
func getDefaultModelProvider() string {
	return "anthropic"
}

// GetDefaultModelName returns the default model name
func getDefaultModelName() string {
	return "claude-3-5-sonnet-20241022"
}

// ListAvailableModels returns a list of available models by provider
// This can be expanded to read from a configuration file or API
func ListAvailableModels() map[string][]AvailableModel {
	models := make(map[string][]AvailableModel)

	// Anthropic models
	models["anthropic"] = []AvailableModel{
		{Provider: "anthropic", ModelName: "Claude 3.5 Sonnet", ModelID: "claude-3-5-sonnet-20241022"},
		{Provider: "anthropic", ModelName: "Claude 3 Sonnet", ModelID: "claude-3-sonnet-20240229"},
		{Provider: "anthropic", ModelName: "Claude 3 Opus", ModelID: "claude-3-opus-20240229"},
		{Provider: "anthropic", ModelName: "Claude 3 Haiku", ModelID: "claude-3-haiku-20240307"},
	}

	// OpenAI models
	models["openai"] = []AvailableModel{
		{Provider: "openai", ModelName: "GPT-4", ModelID: "gpt-4"},
		{Provider: "openai", ModelName: "GPT-4 Turbo", ModelID: "gpt-4-turbo-preview"},
		{Provider: "openai", ModelName: "GPT-3.5 Turbo", ModelID: "gpt-3.5-turbo"},
	}

	// Perplexity models
	models["perplexity"] = []AvailableModel{
		{Provider: "perplexity", ModelName: "Sonar Pro", ModelID: "sonar-pro"},
		{Provider: "perplexity", ModelName: "Sonar", ModelID: "sonar"},
	}

	// Google models
	models["google"] = []AvailableModel{
		{Provider: "google", ModelName: "Gemini Pro", ModelID: "gemini-pro"},
		{Provider: "google", ModelName: "Gemini Pro Vision", ModelID: "gemini-pro-vision"},
	}

	// Mistral models
	models["mistral"] = []AvailableModel{
		{Provider: "mistral", ModelName: "Mistral Large", ModelID: "mistral-large-latest"},
		{Provider: "mistral", ModelName: "Mistral Medium", ModelID: "mistral-medium-latest"},
		{Provider: "mistral", ModelName: "Mistral Small", ModelID: "mistral-small-latest"},
	}

	return models
}

// ValidateModelSelection checks if a given provider and model combination is valid
func ValidateModelSelection(provider string, modelName string) bool {
	if provider == "" || modelName == "" {
		return false
	}

	availableModels := ListAvailableModels()
	providerModels, exists := availableModels[provider]
	if !exists {
		return false
	}

	for _, model := range providerModels {
		if model.ModelID == modelName {
			return true
		}
	}

	return false
}

// GetModelsByProvider returns all available models for a specific provider
func GetModelsByProvider(provider string) []AvailableModel {
	availableModels := ListAvailableModels()
	if models, exists := availableModels[provider]; exists {
		return models
	}
	return []AvailableModel{}
}

// GetAvailableProviders returns all available providers
func GetAvailableProviders() []string {
	availableModels := ListAvailableModels()
	providers := make([]string, 0, len(availableModels))
	for provider := range availableModels {
		providers = append(providers, provider)
	}
	return providers
}
