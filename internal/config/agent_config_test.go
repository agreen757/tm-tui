package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/agreen757/tm-tui/internal/types"
)

func TestLoadAgentType(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	configPath := filepath.Join(tmDir, "config.json")

	// Test case 1: No .taskmaster directory - should return default
	t.Run("no taskmaster directory", func(t *testing.T) {
		// Change to a directory without .taskmaster
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		emptyDir := t.TempDir()
		os.Chdir(emptyDir)

		agentType, err := LoadAgentType()
		if err != nil {
			t.Errorf("LoadAgentType() unexpected error: %v", err)
		}
		if agentType != types.AgentTypeCrush {
			t.Errorf("LoadAgentType() = %v, want %v", agentType, types.AgentTypeCrush)
		}
	})

	// Setup: Create .taskmaster directory for remaining tests
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Test case 2: No config file - should return default
	t.Run("no config file", func(t *testing.T) {
		agentType, err := LoadAgentType()
		if err != nil {
			t.Errorf("LoadAgentType() unexpected error: %v", err)
		}
		if agentType != types.AgentTypeCrush {
			t.Errorf("LoadAgentType() = %v, want %v", agentType, types.AgentTypeCrush)
		}
	})

	// Test case 3: Config with AgentTypeCrush
	t.Run("config with crush", func(t *testing.T) {
		config := map[string]interface{}{
			"agentType": "crush",
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, data, 0600)

		agentType, err := LoadAgentType()
		if err != nil {
			t.Errorf("LoadAgentType() unexpected error: %v", err)
		}
		if agentType != types.AgentTypeCrush {
			t.Errorf("LoadAgentType() = %v, want %v", agentType, types.AgentTypeCrush)
		}

		os.Remove(configPath)
	})

	// Test case 4: Config with AgentTypeGemini
	t.Run("config with gemini", func(t *testing.T) {
		config := map[string]interface{}{
			"agentType": "gemini",
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, data, 0600)

		agentType, err := LoadAgentType()
		if err != nil {
			t.Errorf("LoadAgentType() unexpected error: %v", err)
		}
		if agentType != types.AgentTypeGemini {
			t.Errorf("LoadAgentType() = %v, want %v", agentType, types.AgentTypeGemini)
		}

		os.Remove(configPath)
	})

	// Test case 5: Config with invalid agent type - should return default and error
	t.Run("config with invalid agent type", func(t *testing.T) {
		config := map[string]interface{}{
			"agentType": "invalid",
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configPath, data, 0600)

		agentType, err := LoadAgentType()
		if err == nil {
			t.Errorf("LoadAgentType() expected error for invalid agent type, got nil")
		}
		if agentType != types.AgentTypeCrush {
			t.Errorf("LoadAgentType() = %v, want default %v on error", agentType, types.AgentTypeCrush)
		}

		os.Remove(configPath)
	})

	// Test case 6: Config with malformed JSON - should return default and error
	t.Run("config with malformed json", func(t *testing.T) {
		os.WriteFile(configPath, []byte("{invalid json"), 0600)

		agentType, err := LoadAgentType()
		if err == nil {
			t.Errorf("LoadAgentType() expected error for malformed JSON, got nil")
		}
		if agentType != types.AgentTypeCrush {
			t.Errorf("LoadAgentType() = %v, want default %v on error", agentType, types.AgentTypeCrush)
		}

		os.Remove(configPath)
	})
}

func TestSaveAgentType(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	configPath := filepath.Join(tmDir, "config.json")

	// Setup: Create .taskmaster directory
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Test case 1: Save AgentTypeCrush
	t.Run("save crush agent type", func(t *testing.T) {
		err := SaveAgentType(types.AgentTypeCrush)
		if err != nil {
			t.Fatalf("SaveAgentType() unexpected error: %v", err)
		}

		// Verify file was written correctly
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		if config["agentType"] != "crush" {
			t.Errorf("Config agentType = %v, want 'crush'", config["agentType"])
		}

		os.Remove(configPath)
	})

	// Test case 2: Save AgentTypeGemini
	t.Run("save gemini agent type", func(t *testing.T) {
		err := SaveAgentType(types.AgentTypeGemini)
		if err != nil {
			t.Fatalf("SaveAgentType() unexpected error: %v", err)
		}

		// Verify file was written correctly
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		if config["agentType"] != "gemini" {
			t.Errorf("Config agentType = %v, want 'gemini'", config["agentType"])
		}

		os.Remove(configPath)
	})

	// Test case 3: Save invalid agent type - should return error
	t.Run("save invalid agent type", func(t *testing.T) {
		invalidAgent := types.AgentType(999)
		err := SaveAgentType(invalidAgent)
		if err == nil {
			t.Errorf("SaveAgentType() expected error for invalid agent type, got nil")
		}
	})

	// Test case 4: Update existing config with agent type
	t.Run("update existing config", func(t *testing.T) {
		// Create initial config with other fields
		initialConfig := map[string]interface{}{
			"modelProvider": "anthropic",
			"modelName":     "claude-3-5-sonnet",
		}
		data, _ := json.MarshalIndent(initialConfig, "", "  ")
		os.WriteFile(configPath, data, 0600)

		// Save agent type
		err := SaveAgentType(types.AgentTypeGemini)
		if err != nil {
			t.Fatalf("SaveAgentType() unexpected error: %v", err)
		}

		// Verify existing fields were preserved
		data, err = os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		if config["agentType"] != "gemini" {
			t.Errorf("Config agentType = %v, want 'gemini'", config["agentType"])
		}
		if config["modelProvider"] != "anthropic" {
			t.Errorf("Config modelProvider = %v, want 'anthropic'", config["modelProvider"])
		}
		if config["modelName"] != "claude-3-5-sonnet" {
			t.Errorf("Config modelName = %v, want 'claude-3-5-sonnet'", config["modelName"])
		}

		os.Remove(configPath)
	})
}

func TestGetDefaultAgentType(t *testing.T) {
	defaultAgent := GetDefaultAgentType()
	if defaultAgent != types.AgentTypeCrush {
		t.Errorf("GetDefaultAgentType() = %v, want %v", defaultAgent, types.AgentTypeCrush)
	}
}

func TestValidateAgentType(t *testing.T) {
	tests := []struct {
		name      string
		agentType types.AgentType
		wantError bool
	}{
		{
			name:      "valid AgentTypeCrush",
			agentType: types.AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "valid AgentTypeGemini",
			agentType: types.AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "invalid agent type",
			agentType: types.AgentType(999),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentType(tt.agentType)
			if tt.wantError && err == nil {
				t.Errorf("ValidateAgentType() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateAgentType() unexpected error: %v", err)
			}
		})
	}
}

func TestGetAgentTypeName(t *testing.T) {
	tests := []struct {
		name      string
		agentType types.AgentType
		expected  string
	}{
		{
			name:      "AgentTypeCrush",
			agentType: types.AgentTypeCrush,
			expected:  "Crush",
		},
		{
			name:      "AgentTypeGemini",
			agentType: types.AgentTypeGemini,
			expected:  "Gemini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAgentTypeName(tt.agentType)
			if result != tt.expected {
				t.Errorf("GetAgentTypeName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseAgentType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  types.AgentType
		wantError bool
	}{
		{
			name:      "parse crush",
			input:     "crush",
			expected:  types.AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "parse CRUSH (case insensitive)",
			input:     "CRUSH",
			expected:  types.AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "parse gemini",
			input:     "gemini",
			expected:  types.AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "parse GEMINI (case insensitive)",
			input:     "GEMINI",
			expected:  types.AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "parse invalid",
			input:     "invalid",
			expected:  types.AgentType(-1),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAgentType(tt.input)

			if tt.wantError && err == nil {
				t.Errorf("ParseAgentType() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ParseAgentType() unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("ParseAgentType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
