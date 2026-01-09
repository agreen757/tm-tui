package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AgentType represents the type of AI agent used for task execution.
// It provides type-safe identification of different AI agents throughout
// the application, supporting configuration, UI display, and execution logic.
type AgentType int

const (
	// AgentTypeCrush represents the Crush AI agent (default).
	// This is the zero value for backward compatibility.
	AgentTypeCrush AgentType = iota

	// AgentTypeGemini represents the Google Gemini AI agent.
	AgentTypeGemini
)

// String returns the human-readable name of the agent type.
// This is primarily used for display in UI components and logging.
//
// Returns:
//   - "Crush" for AgentTypeCrush
//   - "Gemini" for AgentTypeGemini
//   - "Unknown" for any unrecognized value
func (a AgentType) String() string {
	switch a {
	case AgentTypeCrush:
		return "Crush"
	case AgentTypeGemini:
		return "Gemini"
	default:
		return "Unknown"
	}
}

// IsValid checks if the agent type is a recognized, valid value.
// This should be used to validate AgentType values before use,
// especially when loading from configuration or user input.
//
// Returns:
//   - true if the agent type is AgentTypeCrush or AgentTypeGemini
//   - false for any other value
func (a AgentType) IsValid() bool {
	return a == AgentTypeCrush || a == AgentTypeGemini
}

// AgentTypeFromString converts a string to an AgentType.
// The conversion is case-insensitive for user convenience.
// This is primarily used when parsing configuration files or user input.
//
// Parameters:
//   - s: The string to parse (e.g., "crush", "Crush", "CRUSH", "gemini")
//
// Returns:
//   - The corresponding AgentType value and nil error on success
//   - AgentType(-1) and an error on failure
//
// Examples:
//
//	agent, err := AgentTypeFromString("crush")    // Returns AgentTypeCrush, nil
//	agent, err := AgentTypeFromString("GEMINI")   // Returns AgentTypeGemini, nil
//	agent, err := AgentTypeFromString("unknown")  // Returns AgentType(-1), error
func AgentTypeFromString(s string) (AgentType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "crush":
		return AgentTypeCrush, nil
	case "gemini":
		return AgentTypeGemini, nil
	default:
		return AgentType(-1), fmt.Errorf("invalid agent type: %q (valid values: \"crush\", \"gemini\")", s)
	}
}

// MarshalJSON implements json.Marshaler for AgentType.
// This encodes the agent type as a JSON string (e.g., "crush", "gemini")
// rather than a numeric value for better readability in configuration files.
//
// Returns:
//   - JSON-encoded string representation and nil error for valid types
//   - Error if the agent type is invalid
//
// Example JSON output:
//
//	{"agentType": "crush"}
//	{"agentType": "gemini"}
func (a AgentType) MarshalJSON() ([]byte, error) {
	if !a.IsValid() {
		return nil, fmt.Errorf("cannot marshal invalid agent type: %d", a)
	}
	return json.Marshal(strings.ToLower(a.String()))
}

// UnmarshalJSON implements json.Unmarshaler for AgentType.
// This decodes a JSON string (e.g., "crush", "gemini") into an AgentType value.
// The parsing is case-insensitive.
//
// Parameters:
//   - data: JSON-encoded string data
//
// Returns:
//   - nil on success (AgentType is updated in place)
//   - Error if the JSON is malformed or contains an invalid agent type
//
// Example JSON input:
//
//	{"agentType": "crush"}   -> AgentTypeCrush
//	{"agentType": "GEMINI"}  -> AgentTypeGemini
func (a *AgentType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("agent type must be a string: %w", err)
	}

	parsed, err := AgentTypeFromString(s)
	if err != nil {
		return err
	}

	*a = parsed
	return nil
}
