package ui

import "bytes"

// PrdCreationState holds the current state of PRD creation workflow
type PrdCreationState struct {
	Title            string
	Summary          string
	Scope            string         // Renamed from ScopeConstraints for consistency
	Filename         string         // Renamed from OutputFilename for consistency
	OutputBuffer     *bytes.Buffer  // Buffer for storing process output
	GeneratedContent string         // Final generated PRD content
	InProgress       bool           // Flag indicating if PRD creation is in progress
}

// NewPrdCreationState creates a new PRD creation state with initialized OutputBuffer
func NewPrdCreationState() *PrdCreationState {
	return &PrdCreationState{
		OutputBuffer: &bytes.Buffer{},
	}
}

// UpdateFromFormValues updates the state from form submission values
func (s *PrdCreationState) UpdateFromFormValues(values PrdFormValues) {
	s.Title = values.Title
	s.Summary = values.Summary
	s.Scope = values.ScopeConstraints
	s.Filename = values.OutputFilename
}

// ToFormValues converts the state back to form values
func (s *PrdCreationState) ToFormValues() PrdFormValues {
	return PrdFormValues{
		Title:            s.Title,
		Summary:          s.Summary,
		ScopeConstraints: s.Scope,
		OutputFilename:   s.Filename,
	}
}

// IsEmpty checks if the state has any meaningful content
func (s *PrdCreationState) IsEmpty() bool {
	return s.Title == "" && s.Summary == "" && s.Scope == "" && s.Filename == ""
}

// Clear resets the state to empty values and clears the output buffer
func (s *PrdCreationState) Clear() {
	s.Title = ""
	s.Summary = ""
	s.Scope = ""
	s.Filename = ""
	if s.OutputBuffer != nil {
		s.OutputBuffer.Reset()
	}
	s.InProgress = false
}
