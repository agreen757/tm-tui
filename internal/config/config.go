package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Common environment variable names for AI API keys
const (
	EnvOpenAIKey      = "OPENAI_API_KEY"
	EnvAnthropicKey   = "ANTHROPIC_API_KEY"
	EnvPerplexityKey  = "PERPLEXITY_API_KEY"
	EnvGoogleKey      = "GOOGLE_API_KEY"
	EnvXAIKey         = "XAI_API_KEY"
	EnvOpenRouterKey  = "OPENROUTER_API_KEY"
	EnvMistralKey     = "MISTRAL_API_KEY"
	EnvAzureOpenAIKey = "AZURE_OPENAI_API_KEY"
	EnvOllamaKey      = "OLLAMA_API_KEY"
)

// Config represents the TUI configuration
type Config struct {
	TaskMasterPath      string            `json:"taskmasterPath"`
	KeyBindings         map[string]string `json:"keyBindings"`
	Theme               ThemeConfig       `json:"theme"`
	UI                  UIConfig          `json:"ui"`
	StatePath           string            `json:"statePath"`
	ProjectRegistryPath string            `json:"projectRegistryPath"`
	ModelProvider       string            `json:"modelProvider,omitempty"`
	ModelName           string            `json:"modelName,omitempty"`
	ActiveTag           string            `json:"activeTag,omitempty"` // Specific tag to use in tasks.json
}

// ThemeConfig defines color and styling options
type ThemeConfig struct {
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	AccentColor    string `json:"accentColor"`
	SuccessColor   string `json:"successColor"`
	ErrorColor     string `json:"errorColor"`
	WarningColor   string `json:"warningColor"`
}

// UIConfig defines UI behavior settings
type UIConfig struct {
	ShowLineNumbers bool   `json:"showLineNumbers"`
	DefaultView     string `json:"defaultView"`
	AutoRefresh     bool   `json:"autoRefresh"`
	RefreshInterval int    `json:"refreshInterval"`
}

// UIState represents the persisted TUI state between sessions
type UIState struct {
	ExpandedIDs      []string       `json:"expandedIds"`
	SelectedID       string         `json:"selectedId"`
	ViewMode         string         `json:"viewMode"`
	FocusedPanel     string         `json:"focusedPanel"`
	ShowDetailsPanel bool           `json:"showDetailsPanel"`
	ShowLogPanel     bool           `json:"showLogPanel"`
	PanelHeights     map[string]int `json:"panelHeights,omitempty"`
	LastPrdPath      string         `json:"lastPrdPath,omitempty"`
}

// Load loads configuration from the specified path
func Load() (*Config, error) {
	// Start with default configuration
	cfg := defaultConfig()

	// Try to find .taskmaster directory
	tmDir, err := findTaskMasterDir()
	if err != nil {
		// No taskmaster directory found, return defaults
		return cfg, nil
	}

	cfg.TaskMasterPath = tmDir
	cfg.StatePath = filepath.Join(tmDir, ".taskmaster", "tui-state.json")
	cfg.ProjectRegistryPath = filepath.Join(tmDir, ".taskmaster", "projects.json")

	// Read the active tag from .taskmaster/state.json
	stateFilePath := filepath.Join(tmDir, ".taskmaster", "state.json")
	if stateData, err := os.ReadFile(stateFilePath); err == nil {
		var state struct {
			CurrentTag string `json:"currentTag"`
		}
		if err := json.Unmarshal(stateData, &state); err == nil && state.CurrentTag != "" {
			cfg.ActiveTag = state.CurrentTag
		}
	}

	// Load default.json if it exists
	configPath := filepath.Join("configs", "default.json")
	if _, err := os.Stat(configPath); err == nil {
		if err := mergeConfigFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load default config: %w", err)
		}
	}

	// Load project-specific overrides from .taskmaster/config.json if it exists
	projectConfigPath := filepath.Join(tmDir, ".taskmaster", "config.json")
	if _, err := os.Stat(projectConfigPath); err == nil {
		if err := mergeConfigFile(cfg, projectConfigPath); err != nil {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
	}

	return cfg, nil
}

// mergeConfigFile loads a config file and merges its values into the target config
func mergeConfigFile(target *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var partial Config
	if err := json.Unmarshal(data, &partial); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge non-zero values from partial into target
	if partial.ModelProvider != "" {
		target.ModelProvider = partial.ModelProvider
	}
	if partial.ModelName != "" {
		target.ModelName = partial.ModelName
	}

	// Merge key bindings
	if len(partial.KeyBindings) > 0 {
		for key, value := range partial.KeyBindings {
			target.KeyBindings[key] = value
		}
	}

	// Merge theme settings
	if partial.Theme.PrimaryColor != "" {
		target.Theme.PrimaryColor = partial.Theme.PrimaryColor
	}
	if partial.Theme.SecondaryColor != "" {
		target.Theme.SecondaryColor = partial.Theme.SecondaryColor
	}
	if partial.Theme.AccentColor != "" {
		target.Theme.AccentColor = partial.Theme.AccentColor
	}
	if partial.Theme.SuccessColor != "" {
		target.Theme.SuccessColor = partial.Theme.SuccessColor
	}
	if partial.Theme.ErrorColor != "" {
		target.Theme.ErrorColor = partial.Theme.ErrorColor
	}
	if partial.Theme.WarningColor != "" {
		target.Theme.WarningColor = partial.Theme.WarningColor
	}

	// Merge UI settings (with explicit zero-value handling for bools)
	// For bools, we need to check if the field was explicitly set
	// We'll merge the default view if it's not empty
	if partial.UI.DefaultView != "" {
		target.UI.DefaultView = partial.UI.DefaultView
	}
	if partial.UI.RefreshInterval > 0 {
		target.UI.RefreshInterval = partial.UI.RefreshInterval
	}

	if partial.ProjectRegistryPath != "" {
		target.ProjectRegistryPath = partial.ProjectRegistryPath
	}

	return nil
}

// findTaskMasterDir searches for .taskmaster directory in current or parent directories
func findTaskMasterDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		tmPath := filepath.Join(dir, ".taskmaster")
		if info, err := os.Stat(tmPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".taskmaster directory not found")
		}
		dir = parent
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		TaskMasterPath: "",
		KeyBindings: map[string]string{
			"quit":       "q",
			"help":       "?",
			"next":       "n",
			"refresh":    "r",
			"expand":     "e",
			"details":    "d",
			"inProgress": "i",
			"done":       "x",
			"blocked":    "b",
			"cancelled":  "c",
		},
		Theme: ThemeConfig{
			PrimaryColor:   "#7d56f4",
			SecondaryColor: "#EE6FF8",
			AccentColor:    "#F780E2",
			SuccessColor:   "#04B575",
			ErrorColor:     "#EF4146",
			WarningColor:   "#FF9800",
		},
		UI: UIConfig{
			ShowLineNumbers: true,
			DefaultView:     "tree",
			AutoRefresh:     true,
			RefreshInterval: 5,
		},
	}
}

// ConfigManager handles configuration with file watching capabilities
type ConfigManager struct {
	config      *Config
	watcher     *Watcher
	reloadChan  chan struct{}
	mu          sync.RWMutex
	configPaths []string
}

// NewConfigManager creates a new config manager with optional file watching
func NewConfigManager() (*ConfigManager, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	cm := &ConfigManager{
		config:     cfg,
		reloadChan: make(chan struct{}, 1),
	}

	// Determine config paths to watch
	paths := []string{
		filepath.Join("configs", "default.json"),
	}

	// Add .taskmaster/config.json if it exists
	if cfg.TaskMasterPath != "" {
		tmConfigPath := filepath.Join(cfg.TaskMasterPath, ".taskmaster", "config.json")
		if _, err := os.Stat(tmConfigPath); err == nil {
			paths = append(paths, tmConfigPath)
		}
	}

	cm.configPaths = paths

	return cm, nil
}

// GetConfig returns the current configuration (thread-safe)
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Reload loads the configuration from disk
func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cfg, err := Load()
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	cm.config = cfg
	return nil
}

// StartWatcher begins watching config files for changes with a 300ms debounce
func (cm *ConfigManager) StartWatcher(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.watcher != nil {
		return fmt.Errorf("watcher already started")
	}

	if len(cm.configPaths) == 0 {
		return fmt.Errorf("no config paths to watch")
	}

	// Create watcher for config files
	watcher, err := NewWatcher(ctx, cm.configPaths...)
	if err != nil {
		return fmt.Errorf("failed to create config watcher: %w", err)
	}

	// Start watching with 300ms debounce
	if err := watcher.Start(300 * time.Millisecond); err != nil {
		return fmt.Errorf("failed to start config watcher: %w", err)
	}

	cm.watcher = watcher

	// Start goroutine to handle config change events
	go cm.handleConfigChanges(ctx)

	return nil
}

// handleConfigChanges processes config file change notifications
func (cm *ConfigManager) handleConfigChanges(ctx context.Context) {
	if cm.watcher == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-cm.watcher.Events():
			// Reload config on file change
			if err := cm.Reload(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
				continue
			}

			// Signal that config was reloaded
			select {
			case cm.reloadChan <- struct{}{}:
			default:
				// Channel full, reload notification already pending
			}

		case err := <-cm.watcher.Errors():
			fmt.Fprintf(os.Stderr, "Config watcher error: %v\n", err)
		}
	}
}

// StopWatcher stops the config file watcher if it's running
func (cm *ConfigManager) StopWatcher() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.watcher == nil {
		return nil
	}

	err := cm.watcher.Stop()
	cm.watcher = nil
	return err
}

// ReloadEvents returns a channel that signals when config has been reloaded
func (cm *ConfigManager) ReloadEvents() <-chan struct{} {
	return cm.reloadChan
}

// SaveState persists the UI state to disk
func SaveState(statePath string, state *UIState) error {
	if statePath == "" {
		return fmt.Errorf("state path is empty")
	}

	// Ensure the directory exists
	dir := filepath.Dir(statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal state to JSON with indentation for readability
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write state file with user-only permissions
	if err := os.WriteFile(statePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads the UI state from disk
func LoadState(statePath string) (*UIState, error) {
	if statePath == "" {
		return nil, fmt.Errorf("state path is empty")
	}

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Return default state if file doesn't exist
		return &UIState{
			ExpandedIDs:      []string{},
			SelectedID:       "",
			ViewMode:         "tree",
			FocusedPanel:     "taskList",
			ShowDetailsPanel: true,
			ShowLogPanel:     false,
			PanelHeights:     make(map[string]int),
		}, nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var state UIState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Ensure maps are initialized
	if state.PanelHeights == nil {
		state.PanelHeights = make(map[string]int)
	}
	if state.ExpandedIDs == nil {
		state.ExpandedIDs = []string{}
	}

	return &state, nil
}

// GetEnv retrieves an environment variable with an optional fallback value
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetAPIKeys returns a map of all configured API keys from environment variables
// This is intended for passing to child processes (e.g., task-master CLI)
func GetAPIKeys() map[string]string {
	keys := make(map[string]string)

	envVars := []string{
		EnvOpenAIKey,
		EnvAnthropicKey,
		EnvPerplexityKey,
		EnvGoogleKey,
		EnvXAIKey,
		EnvOpenRouterKey,
		EnvMistralKey,
		EnvAzureOpenAIKey,
		EnvOllamaKey,
	}

	for _, key := range envVars {
		if value := os.Getenv(key); value != "" {
			keys[key] = value
		}
	}

	return keys
}

// HasAPIKey checks if a specific API key environment variable is set
func HasAPIKey(key string) bool {
	return os.Getenv(key) != ""
}

// GetConfiguredProviders returns a list of providers that have API keys configured
func GetConfiguredProviders() []string {
	var providers []string

	providerMap := map[string]string{
		"openai":     EnvOpenAIKey,
		"anthropic":  EnvAnthropicKey,
		"perplexity": EnvPerplexityKey,
		"google":     EnvGoogleKey,
		"xai":        EnvXAIKey,
		"openrouter": EnvOpenRouterKey,
		"mistral":    EnvMistralKey,
		"azure":      EnvAzureOpenAIKey,
		"ollama":     EnvOllamaKey,
	}

	for provider, envVar := range providerMap {
		if HasAPIKey(envVar) {
			providers = append(providers, provider)
		}
	}

	return providers
}
