package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Environment variable names for Crush config
const (
	// CRUSH_CONFIG_PATH can override the config file location
	EnvCrushConfigPath = "CRUSH_CONFIG_PATH"
	// CRUSH_PROJECT_ROOT can override the project root detection
	EnvCrushProjectRoot = "CRUSH_PROJECT_ROOT"
)

// Project root markers to search for when detecting project root
var projectRootMarkers = []string{
	".taskmaster",  // Task Master project
	".git",         // Git repository
	"go.mod",       // Go module
	".crush.json",  // Existing Crush config
}


// SelectedModel represents a single model configuration entry
// Used in the models map to specify a model ID and its provider
type SelectedModel struct {
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

// CrushConfig represents the configuration stored in .crush.json
// It models the known configuration fields needed for Crush integration,
// with future-proofing for unknown/forward-compatible fields via ExtraFields.
type CrushConfig struct {
	// Schema reference (informational)
	Schema string `json:"$schema,omitempty"`

	// Models map contains named model configurations (e.g., "large", "small")
	// This is the new multi-model format compatible with Crush CLI
	Models map[string]SelectedModel `json:"models,omitempty"`

	// Options contain Crush-specific settings
	Options CrushOptions `json:"options,omitempty"`

	// Version can be used for future schema migration
	Version string `json:"version,omitempty"`

	// ExtraFields captures and preserves unknown fields for forward compatibility
	ExtraFields map[string]interface{} `json:"-"`

	// RawJSON stores the raw JSON for round-trip preservation if needed
	RawJSON json.RawMessage `json:"-"`
}

// CrushOptions contains Crush-specific configuration options
type CrushOptions struct {
	// Context paths to include when running Crush
	ContextPaths []string `json:"context_paths,omitempty"`

	// Skills paths for custom Crush skills
	SkillsPaths []string `json:"skills_paths,omitempty"`

	// Additional unknown options for forward compatibility
	ExtraOptions map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling to preserve unknown fields
func (c *CrushConfig) UnmarshalJSON(data []byte) error {
	// Store raw JSON for reference
	c.RawJSON = json.RawMessage(data)

	// Create a temporary struct without our custom fields to unmarshal known fields
	type tempCrushConfig struct {
		Schema  string                     `json:"$schema,omitempty"`
		Models  map[string]SelectedModel   `json:"models,omitempty"`
		Version string                     `json:"version,omitempty"`
		Options CrushOptions               `json:"options,omitempty"`
		Extra   map[string]interface{}     `json:"-"`
	}

	// First, unmarshal into a generic map to capture all fields
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return fmt.Errorf("failed to unmarshal CrushConfig: %w", err)
	}

	// Then unmarshal into the known structure
	temp := &tempCrushConfig{}
	if err := json.Unmarshal(data, temp); err != nil {
		return fmt.Errorf("failed to unmarshal CrushConfig fields: %w", err)
	}

	// Copy known fields
	c.Schema = temp.Schema
	c.Models = temp.Models
	c.Version = temp.Version
	c.Options = temp.Options

	// Migrate legacy single-model format to multi-model format
	// This handles backward compatibility for configs using the old "model" field
	if legacyModel, hasLegacyModel := rawMap["model"]; hasLegacyModel {
		// Only migrate if Models["large"] doesn't already exist
		// This ensures idempotent behavior and respects user-set values
		if c.Models == nil {
			c.Models = make(map[string]SelectedModel)
		}
		
		if _, hasLargeModel := c.Models["large"]; !hasLargeModel {
			// Extract model string from legacy field
			var modelStr string
			var providerStr string
			
			// Legacy format could be just a string, or a structured object
			switch v := legacyModel.(type) {
			case string:
				modelStr = v
				// Try to infer provider from model name
				providerStr = inferProviderFromModel(modelStr)
			case map[string]interface{}:
				// If it's an object, try to extract model and provider
				if model, ok := v["model"].(string); ok {
					modelStr = model
				}
				if provider, ok := v["provider"].(string); ok {
					providerStr = provider
				} else {
					providerStr = inferProviderFromModel(modelStr)
				}
			}
			
			// Migrate to Models["large"] if we have a model string
			if modelStr != "" {
				c.Models["large"] = SelectedModel{
					Model:    modelStr,
					Provider: providerStr,
				}
			}
		}
	}

	// Extract unknown fields (those not in our known schema)
	// Note: "model" is now a known field (legacy), so exclude it from ExtraFields
	knownFields := map[string]bool{
		"$schema": true,
		"models":  true,
		"model":   true, // Legacy field, should not be preserved in ExtraFields
		"version": true,
		"options": true,
	}

	c.ExtraFields = make(map[string]interface{})
	for k, v := range rawMap {
		if !knownFields[k] {
			c.ExtraFields[k] = v
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling to include unknown fields
func (c *CrushConfig) MarshalJSON() ([]byte, error) {
	// Build a map with all fields (known + extra)
	result := make(map[string]interface{})

	// Add known fields
	if c.Schema != "" {
		result["$schema"] = c.Schema
	}
	if c.Models != nil && len(c.Models) > 0 {
		result["models"] = c.Models
	}
	if c.Version != "" {
		result["version"] = c.Version
	}

	// Marshal options separately to include unknown option fields
	if c.Options.ContextPaths != nil || c.Options.SkillsPaths != nil || c.Options.ExtraOptions != nil {
		optionsMap := make(map[string]interface{})
		if c.Options.ContextPaths != nil && len(c.Options.ContextPaths) > 0 {
			optionsMap["context_paths"] = c.Options.ContextPaths
		}
		if c.Options.SkillsPaths != nil && len(c.Options.SkillsPaths) > 0 {
			optionsMap["skills_paths"] = c.Options.SkillsPaths
		}
		// Include extra options
		for k, v := range c.Options.ExtraOptions {
			optionsMap[k] = v
		}
		if len(optionsMap) > 0 {
			result["options"] = optionsMap
		}
	}

	// Add extra fields
	for k, v := range c.ExtraFields {
		result[k] = v
	}

	return json.Marshal(result)
}

// GetCrushConfigPath returns the path to the .crush.json file
// It searches upward from the current directory to find the project root.
// 
// Resolution order:
// 1. CRUSH_CONFIG_PATH environment variable (if set and valid)
// 2. CRUSH_PROJECT_ROOT environment variable (if set) + .crush.json
// 3. Project root detected by markers (.taskmaster, .git, go.mod, .crush.json)
// 4. Current working directory + .crush.json
//
// Returns error only if unable to determine current directory.
func GetCrushConfigPath() (string, error) {
	// 1. Check for explicit config path override via environment
	if configPath := os.Getenv(EnvCrushConfigPath); configPath != "" {
		return configPath, nil
	}

	// 2. Check for explicit project root override via environment
	if projectRoot := os.Getenv(EnvCrushProjectRoot); projectRoot != "" {
		return filepath.Join(projectRoot, ".crush.json"), nil
	}

	// 3. Try to detect project root using markers
	projectRoot, err := DetectProjectRoot()
	if err == nil {
		return filepath.Join(projectRoot, ".crush.json"), nil
	}

	// 4. Fall back to current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(wd, ".crush.json"), nil
}

// DetectProjectRoot finds the project root by walking up the directory tree
// and looking for marker files/directories like .git, go.mod, .taskmaster, or .crush.json.
// Returns the first directory found containing any marker.
// Returns error if no markers found or unable to walk directory tree.
func DetectProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return detectProjectRootFrom(wd)
}

// detectProjectRootFrom walks up from the given directory looking for project root markers.
// Returns the first directory containing any marker.
// Returns error if no markers found or path is invalid.
func detectProjectRootFrom(startDir string) (string, error) {
	dir := startDir

	// Ensure we have an absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		// Check for each marker in this directory
		for _, marker := range projectRootMarkers {
			markerPath := filepath.Join(absDir, marker)
			if _, err := os.Stat(markerPath); err == nil {
				// Found a marker, this is the project root
				return absDir, nil
			}
		}

		// Move to parent directory
		parent := filepath.Dir(absDir)
		if parent == absDir {
			// Reached filesystem root without finding any marker
			return "", fmt.Errorf("no project root marker found")
		}

		absDir = parent
	}
}

// FindProjectRootMarker returns the path of the first marker found when walking up
// from the current directory. Returns the marker name and path, or error if none found.
func FindProjectRootMarker() (marker string, path string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get working directory: %w", err)
	}

	absDir, err := filepath.Abs(wd)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		for _, m := range projectRootMarkers {
			markerPath := filepath.Join(absDir, m)
			if _, err := os.Stat(markerPath); err == nil {
				return m, markerPath, nil
			}
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			return "", "", fmt.Errorf("no project root marker found")
		}

		absDir = parent
	}
}

// GetCrushConfigPathOrDefault returns GetCrushConfigPath result or the second parameter as fallback
// Useful for operations that want a guaranteed path even if detection fails
func GetCrushConfigPathOrDefault(fallback string) string {
	if path, err := GetCrushConfigPath(); err == nil {
		return path
	}
	return fallback
}


// LoadCrushConfig loads and parses the .crush.json file
// Returns default configuration if file doesn't exist
func LoadCrushConfig() (*CrushConfig, error) {
	configPath, err := GetCrushConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return defaults
		return getDefaultCrushConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse JSON
	config := &CrushConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return config, nil
}

// SaveCrushConfig writes the configuration to .crush.json atomically
// Creates the file if it doesn't exist, preserves unknown fields if updating
// Uses atomic write semantics: writes to a temporary file and then renames to ensure
// the configuration file is never left in a truncated or invalid state.
func SaveCrushConfig(config *CrushConfig) error {
	configPath, err := GetCrushConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine config path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config with proper indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Get existing file permissions to preserve them
	var perm os.FileMode = 0644
	if fileInfo, err := os.Stat(configPath); err == nil {
		perm = fileInfo.Mode()
	}

	// Write atomically using a temporary file in the same directory
	// This ensures we never have a truncated/invalid config file
	tempFile, err := os.CreateTemp(dir, ".crush.json.tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tempPath := tempFile.Name()
	defer func() {
		// Clean up temp file if we exit early
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Close temp file to flush data
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Set permissions on temp file to match original (if exists)
	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Atomically rename temp file to actual config path
	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to config path: %w", err)
	}

	return nil
}

// UpdateCrushModel updates a model entry in .crush.json
// This is the primary API for updating model configurations in Crush.
// It validates inputs, loads existing config (or defaults), updates the specified
// model type, and saves atomically while preserving all other config fields.
func UpdateCrushModel(modelID, provider, modelType string) error {
	// Validate inputs
	if modelID == "" {
		return fmt.Errorf("modelID cannot be empty")
	}
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if modelType == "" {
		return fmt.Errorf("modelType cannot be empty")
	}

	// Load existing config (or defaults if doesn't exist)
	config, err := LoadCrushConfig()
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	// Update the specific model type
	config.SetModelByType(modelType, modelID, provider)

	// Save back to file
	return SaveCrushConfig(config)
}

// InitCrushConfig creates a .crush.json file with default configuration if it doesn't exist
// Returns the configuration that was created or loaded
func InitCrushConfig() (*CrushConfig, error) {
	configPath, err := GetCrushConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, load and return it
		return LoadCrushConfig()
	}

	// File doesn't exist, create new one with defaults
	config := getDefaultCrushConfig()

	// Save the default configuration
	if err := SaveCrushConfig(config); err != nil {
		return nil, fmt.Errorf("failed to initialize config file: %w", err)
	}

	return config, nil
}

// getDefaultCrushConfig returns a CrushConfig with sensible defaults
func getDefaultCrushConfig() *CrushConfig {
	return &CrushConfig{
		Schema:      "https://charm.land/crush.json",
		Models:      make(map[string]SelectedModel), // Initialize as empty map, not nil
		Version:     "1.0",
		ExtraFields: make(map[string]interface{}),
		Options: CrushOptions{
			ContextPaths:  []string{},
			SkillsPaths:   []string{"./.crush/skills"},
			ExtraOptions:  make(map[string]interface{}),
		},
	}
}

// GetModelByType retrieves a model configuration by its type key (e.g., "large", "small")
// Returns the SelectedModel and a boolean indicating if the model type exists
func (c *CrushConfig) GetModelByType(modelType string) (SelectedModel, bool) {
	if c.Models == nil {
		return SelectedModel{}, false
	}
	model, ok := c.Models[modelType]
	return model, ok
}

// SetModelByType sets or updates a model configuration for a given type
// Initializes the Models map if it's nil
func (c *CrushConfig) SetModelByType(modelType, modelID, provider string) {
	if c.Models == nil {
		c.Models = make(map[string]SelectedModel)
	}
	c.Models[modelType] = SelectedModel{
		Model:    modelID,
		Provider: provider,
	}
}

// RemoveModelByType removes a model configuration by type
// No-op if Models is nil or the key doesn't exist
func (c *CrushConfig) RemoveModelByType(modelType string) {
	if c.Models == nil {
		return
	}
	delete(c.Models, modelType)
}

// GetContextPaths returns the list of context paths
func (c *CrushConfig) GetContextPaths() []string {
	return c.Options.ContextPaths
}

// SetContextPaths sets the context paths
func (c *CrushConfig) SetContextPaths(paths []string) {
	c.Options.ContextPaths = paths
}

// GetSkillsPaths returns the list of skills paths
func (c *CrushConfig) GetSkillsPaths() []string {
	return c.Options.SkillsPaths
}

// SetSkillsPaths sets the skills paths
func (c *CrushConfig) SetSkillsPaths(paths []string) {
	c.Options.SkillsPaths = paths
}

// AddContextPath appends a context path if not already present
func (c *CrushConfig) AddContextPath(path string) {
	for _, p := range c.Options.ContextPaths {
		if p == path {
			return // Already present
		}
	}
	c.Options.ContextPaths = append(c.Options.ContextPaths, path)
}

// inferProviderFromModel attempts to infer the provider from the model name
// Returns empty string if provider cannot be inferred
func inferProviderFromModel(model string) string {
	if model == "" {
		return ""
	}
	
	// Check for common provider patterns in model names
	modelLower := strings.ToLower(model)
	
	// OpenAI models
	if strings.Contains(modelLower, "gpt") || 
	   strings.Contains(modelLower, "davinci") || 
	   strings.Contains(modelLower, "curie") ||
	   strings.Contains(modelLower, "babbage") ||
	   strings.Contains(modelLower, "ada") {
		return "openai"
	}
	
	// Anthropic models
	if strings.Contains(modelLower, "claude") {
		return "anthropic"
	}
	
	// Google models
	if strings.Contains(modelLower, "gemini") || 
	   strings.Contains(modelLower, "palm") ||
	   strings.Contains(modelLower, "bard") {
		return "google"
	}
	
	// Meta models
	if strings.Contains(modelLower, "llama") {
		return "meta"
	}
	
	// Mistral models
	if strings.Contains(modelLower, "mistral") ||
	   strings.Contains(modelLower, "mixtral") {
		return "mistral"
	}
	
	// Cohere models
	if strings.Contains(modelLower, "cohere") ||
	   strings.Contains(modelLower, "command") {
		return "cohere"
	}
	
	// Default: cannot infer, return empty
	return ""
}

// RemoveContextPath removes a context path
func (c *CrushConfig) RemoveContextPath(path string) {
	newPaths := make([]string, 0)
	for _, p := range c.Options.ContextPaths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	c.Options.ContextPaths = newPaths
}
