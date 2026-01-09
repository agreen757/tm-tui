package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agreen757/tm-tui/internal/types"
)

// LoadAgentType loads the configured agent type from .taskmaster/config.json
// Returns AgentTypeCrush (default) if config file doesn't exist or is invalid
func LoadAgentType() (types.AgentType, error) {
	tmDir, err := findTaskMasterDir()
	if err != nil {
		// Return default if .taskmaster directory not found
		return types.AgentTypeCrush, nil
	}

	configPath := filepath.Join(tmDir, ".taskmaster", "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default if file doesn't exist
		return types.AgentTypeCrush, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default if read fails
		return types.AgentTypeCrush, fmt.Errorf("failed to read config file: %w", err)
	}

	// Attempt to unmarshal the config
	var fullConfig map[string]interface{}
	if err := json.Unmarshal(data, &fullConfig); err != nil {
		// Return default if JSON is invalid
		return types.AgentTypeCrush, fmt.Errorf("invalid config format: %w", err)
	}

	// Extract agent type if present
	if agentTypeStr, ok := fullConfig["agentType"].(string); ok && agentTypeStr != "" {
		agentType, err := types.AgentTypeFromString(agentTypeStr)
		if err != nil {
			// Return default if invalid agent type in config
			return types.AgentTypeCrush, fmt.Errorf("invalid agent type in config: %w", err)
		}
		return agentType, nil
	}

	// Return default if not found in config
	return types.AgentTypeCrush, nil
}

// SaveAgentType saves the agent type selection to .taskmaster/config.json
// Creates the config file if it doesn't exist, updates it if it does
func SaveAgentType(agentType types.AgentType) error {
	if !agentType.IsValid() {
		return fmt.Errorf("invalid agent type: %v", agentType)
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

	// Update agent type setting (using string representation for readability)
	agentTypeJSON, err := agentType.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal agent type: %w", err)
	}

	// agentTypeJSON is a JSON string like "\"crush\"", so we need to unmarshal it to get the raw string
	var agentTypeStr string
	if err := json.Unmarshal(agentTypeJSON, &agentTypeStr); err != nil {
		return fmt.Errorf("failed to unmarshal agent type string: %w", err)
	}

	config["agentType"] = agentTypeStr

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

// GetDefaultAgentType returns the default agent type
func GetDefaultAgentType() types.AgentType {
	return types.AgentTypeCrush
}

// ValidateAgentType checks if an agent type is valid and supported
func ValidateAgentType(agentType types.AgentType) error {
	if !agentType.IsValid() {
		return fmt.Errorf("invalid agent type: %v (valid values: Crush, Gemini)", agentType)
	}
	return nil
}

// GetAgentTypeName returns the human-readable name for an agent type
func GetAgentTypeName(agentType types.AgentType) string {
	return agentType.String()
}

// ParseAgentType parses a string into an AgentType
// This is a convenience wrapper around types.AgentTypeFromString
func ParseAgentType(s string) (types.AgentType, error) {
	return types.AgentTypeFromString(s)
}
