package types

import (
	"encoding/json"
	"testing"
)

// TestAgentType_String tests the String() method for all valid agent types
func TestAgentType_String(t *testing.T) {
	tests := []struct {
		name     string
		agent    AgentType
		expected string
	}{
		{
			name:     "AgentTypeCrush returns Crush",
			agent:    AgentTypeCrush,
			expected: "Crush",
		},
		{
			name:     "AgentTypeGemini returns Gemini",
			agent:    AgentTypeGemini,
			expected: "Gemini",
		},
		{
			name:     "Invalid agent type returns Unknown",
			agent:    AgentType(999),
			expected: "Unknown",
		},
		{
			name:     "Negative agent type returns Unknown",
			agent:    AgentType(-1),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.agent.String()
			if result != tt.expected {
				t.Errorf("AgentType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestAgentType_IsValid tests the IsValid() method
func TestAgentType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		agent    AgentType
		expected bool
	}{
		{
			name:     "AgentTypeCrush is valid",
			agent:    AgentTypeCrush,
			expected: true,
		},
		{
			name:     "AgentTypeGemini is valid",
			agent:    AgentTypeGemini,
			expected: true,
		},
		{
			name:     "Invalid positive value is not valid",
			agent:    AgentType(999),
			expected: false,
		},
		{
			name:     "Invalid negative value is not valid",
			agent:    AgentType(-1),
			expected: false,
		},
		{
			name:     "Out of range value is not valid",
			agent:    AgentType(100),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.agent.IsValid()
			if result != tt.expected {
				t.Errorf("AgentType.IsValid() = %v, want %v for agent %d", result, tt.expected, tt.agent)
			}
		})
	}
}

// TestAgentTypeFromString tests the AgentTypeFromString() function
func TestAgentTypeFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  AgentType
		wantError bool
	}{
		{
			name:      "lowercase crush",
			input:     "crush",
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "uppercase CRUSH",
			input:     "CRUSH",
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "mixed case Crush",
			input:     "Crush",
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "lowercase gemini",
			input:     "gemini",
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "uppercase GEMINI",
			input:     "GEMINI",
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "mixed case Gemini",
			input:     "Gemini",
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "with leading whitespace",
			input:     "  crush",
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "with trailing whitespace",
			input:     "gemini  ",
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "with surrounding whitespace",
			input:     "  crush  ",
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "empty string returns error",
			input:     "",
			expected:  AgentType(-1),
			wantError: true,
		},
		{
			name:      "invalid agent name returns error",
			input:     "unknown",
			expected:  AgentType(-1),
			wantError: true,
		},
		{
			name:      "invalid agent openai returns error",
			input:     "openai",
			expected:  AgentType(-1),
			wantError: true,
		},
		{
			name:      "typo crussh returns error",
			input:     "crussh",
			expected:  AgentType(-1),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AgentTypeFromString(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("AgentTypeFromString(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("AgentTypeFromString(%q) unexpected error: %v", tt.input, err)
				}
			}

			if result != tt.expected {
				t.Errorf("AgentTypeFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestAgentType_MarshalJSON tests JSON marshaling
func TestAgentType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		agent     AgentType
		expected  string
		wantError bool
	}{
		{
			name:      "AgentTypeCrush marshals to crush",
			agent:     AgentTypeCrush,
			expected:  `"crush"`,
			wantError: false,
		},
		{
			name:      "AgentTypeGemini marshals to gemini",
			agent:     AgentTypeGemini,
			expected:  `"gemini"`,
			wantError: false,
		},
		{
			name:      "Invalid agent type returns error",
			agent:     AgentType(999),
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.agent)

			if tt.wantError {
				if err == nil {
					t.Errorf("AgentType.MarshalJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("AgentType.MarshalJSON() unexpected error: %v", err)
				return
			}

			if string(result) != tt.expected {
				t.Errorf("AgentType.MarshalJSON() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

// TestAgentType_UnmarshalJSON tests JSON unmarshaling
func TestAgentType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		expected  AgentType
		wantError bool
	}{
		{
			name:      "crush unmarshals to AgentTypeCrush",
			json:      `"crush"`,
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "CRUSH unmarshals to AgentTypeCrush (case insensitive)",
			json:      `"CRUSH"`,
			expected:  AgentTypeCrush,
			wantError: false,
		},
		{
			name:      "gemini unmarshals to AgentTypeGemini",
			json:      `"gemini"`,
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "GEMINI unmarshals to AgentTypeGemini (case insensitive)",
			json:      `"GEMINI"`,
			expected:  AgentTypeGemini,
			wantError: false,
		},
		{
			name:      "invalid agent type returns error",
			json:      `"invalid"`,
			expected:  AgentType(0),
			wantError: true,
		},
		{
			name:      "numeric value returns error",
			json:      `0`,
			expected:  AgentType(0),
			wantError: true,
		},
		{
			name:      "empty string returns error",
			json:      `""`,
			expected:  AgentType(0),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var agent AgentType
			err := json.Unmarshal([]byte(tt.json), &agent)

			if tt.wantError {
				if err == nil {
					t.Errorf("AgentType.UnmarshalJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("AgentType.UnmarshalJSON() unexpected error: %v", err)
				return
			}

			if agent != tt.expected {
				t.Errorf("AgentType.UnmarshalJSON() = %v, want %v", agent, tt.expected)
			}
		})
	}
}

// TestAgentType_JSONRoundTrip tests marshaling and unmarshaling together
func TestAgentType_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		agent AgentType
	}{
		{
			name:  "AgentTypeCrush round trip",
			agent: AgentTypeCrush,
		},
		{
			name:  "AgentTypeGemini round trip",
			agent: AgentTypeGemini,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.agent)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			var decoded AgentType
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Verify
			if decoded != tt.agent {
				t.Errorf("Round trip failed: got %v, want %v", decoded, tt.agent)
			}
		})
	}
}

// TestAgentType_ZeroValue tests that the zero value is AgentTypeCrush
func TestAgentType_ZeroValue(t *testing.T) {
	var agent AgentType

	if agent != AgentTypeCrush {
		t.Errorf("Zero value of AgentType = %v, want AgentTypeCrush", agent)
	}

	if agent.String() != "Crush" {
		t.Errorf("Zero value String() = %q, want %q", agent.String(), "Crush")
	}

	if !agent.IsValid() {
		t.Errorf("Zero value IsValid() = false, want true")
	}
}
