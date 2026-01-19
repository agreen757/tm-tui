package dialog

import (
	"strings"
	"testing"
)

func TestGenerateCrushPrdPrompt_BasicFormat(t *testing.T) {
	summary := "This is a test summary"
	prompt := GenerateCrushPrdPrompt("Test PRD", summary, "")

	// Verify prompt starts with the expected format
	if !strings.HasPrefix(prompt, "create a complete and concise PRD from this summary:") {
		t.Errorf("Expected prompt to start with 'create a complete and concise PRD from this summary:', got: %s", prompt)
	}

	// Verify summary is included
	if !strings.Contains(prompt, summary) {
		t.Errorf("Summary not included in prompt. Got: %s", prompt)
	}
}

func TestGenerateCrushPrdPrompt_SummaryIncluded(t *testing.T) {
	summary := "This is my unique summary with special chars: !@#$%"
	prompt := GenerateCrushPrdPrompt("Title", summary, "Scope")

	if !strings.Contains(prompt, summary) {
		t.Errorf("Summary not included verbatim. Got: %s", prompt)
	}
}

func TestGenerateCrushPrdPrompt_OnlySummaryMatters(t *testing.T) {
	// Title and scope should not affect the prompt - only summary matters
	prompt1 := GenerateCrushPrdPrompt("Title1", "Summary", "Scope1")
	prompt2 := GenerateCrushPrdPrompt("Title2", "Summary", "Scope2")
	prompt3 := GenerateCrushPrdPrompt("Title3", "Summary", "")

	// All prompts should be identical since they have the same summary
	if prompt1 != prompt2 || prompt1 != prompt3 {
		t.Error("Prompts with same summary should be identical regardless of title/scope")
	}
}

func TestGenerateCrushPrdPrompt_EmptySummary(t *testing.T) {
	prompt := GenerateCrushPrdPrompt("Title", "", "Scope")

	expected := "create a complete and concise PRD from this summary: "
	if prompt != expected {
		t.Errorf("Expected %q, got %q", expected, prompt)
	}
}

func TestGenerateCrushPrdPrompt_LongSummary(t *testing.T) {
	longSummary := strings.Repeat("This is a very long summary. ", 50)
	prompt := GenerateCrushPrdPrompt("Title", longSummary, "")

	if !strings.Contains(prompt, longSummary) {
		t.Error("Long summary should be included in full")
	}

	if !strings.HasPrefix(prompt, "create a complete and concise PRD from this summary:") {
		t.Error("Prompt should still have correct prefix with long summary")
	}
}

func TestGenerateCrushPrdPrompt_SpecialCharacters(t *testing.T) {
	summary := "Summary with quotes \"like this\" and newlines\nand tabs\t"
	prompt := GenerateCrushPrdPrompt("Title", summary, "")

	if !strings.Contains(prompt, summary) {
		t.Error("Summary with special characters should be included exactly as provided")
	}
}

func TestGenerateCrushPrdPrompt_ReturnType(t *testing.T) {
	result := GenerateCrushPrdPrompt("Test", "Summary", "")

	if result == "" {
		t.Error("GenerateCrushPrdPrompt should return a non-empty string")
	}
}
