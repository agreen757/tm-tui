package ui

import (
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// ErrorCategoryIO represents file I/O errors
	ErrorCategoryIO ErrorCategory = "io"
	// ErrorCategoryParsing represents parsing/validation errors
	ErrorCategoryParsing ErrorCategory = "parsing"
	// ErrorCategoryOperation represents operation/execution errors
	ErrorCategoryOperation ErrorCategory = "operation"
	// ErrorCategoryValidation represents user input validation errors
	ErrorCategoryValidation ErrorCategory = "validation"
	// ErrorCategoryDependency represents missing dependency errors
	ErrorCategoryDependency ErrorCategory = "dependency"
)

// AppError represents a standardized application error
type AppError struct {
	Category   ErrorCategory
	Title      string
	Message    string
	Details    string
	RecoveryHints []string
	Underlying error
}

// NewAppError creates a new app error
func NewAppError(category ErrorCategory, title, message string, underlying error) *AppError {
	return &AppError{
		Category:      category,
		Title:         title,
		Message:       message,
		Underlying:    underlying,
		RecoveryHints: []string{},
	}
}

// NewIOError creates a standardized I/O error
func NewIOError(title, message string, underlying error) *AppError {
	err := NewAppError(ErrorCategoryIO, title, message, underlying)
	err.RecoveryHints = []string{
		"Check file permissions",
		"Verify the file path is correct",
		"Try again or contact support if the issue persists",
	}
	return err
}

// NewParsingError creates a standardized parsing error
func NewParsingError(title, message string, underlying error) *AppError {
	err := NewAppError(ErrorCategoryParsing, title, message, underlying)
	err.RecoveryHints = []string{
		"Check the file format is correct",
		"Verify the file content is valid",
		"Try uploading a different file",
	}
	return err
}

// NewOperationError creates a standardized operation error
func NewOperationError(title, message string, underlying error) *AppError {
	err := NewAppError(ErrorCategoryOperation, title, message, underlying)
	err.RecoveryHints = []string{
		"Retry the operation",
		"Check if all prerequisites are satisfied",
		"Restart the application if the error persists",
	}
	return err
}

// NewValidationError creates a standardized validation error
func NewValidationError(title, message string, underlying error) *AppError {
	err := NewAppError(ErrorCategoryValidation, title, message, underlying)
	err.RecoveryHints = []string{
		"Review your input and try again",
		"Check the field requirements",
	}
	return err
}

// NewDependencyError creates a standardized dependency error
func NewDependencyError(title, message string, underlying error) *AppError {
	err := NewAppError(ErrorCategoryDependency, title, message, underlying)
	err.RecoveryHints = []string{
		"Ensure all required services are available",
		"Check your installation or configuration",
		"Restart the application",
	}
	return err
}

// WithDetails adds detailed error information
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithRecoveryHints sets recovery hints
func (e *AppError) WithRecoveryHints(hints ...string) *AppError {
	e.RecoveryHints = hints
	return e
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Title, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Title, e.Message)
}

// GetDisplayMessage returns a user-friendly message for the error
func (e *AppError) GetDisplayMessage() string {
	if e.Details != "" {
		return fmt.Sprintf("%s\n\n%s", e.Message, e.Details)
	}
	return e.Message
}

// GetRecoveryMessage returns recovery hints as a formatted string
func (e *AppError) GetRecoveryMessage() string {
	if len(e.RecoveryHints) == 0 {
		return "Please try again or contact support."
	}
	msg := "To recover:\n"
	for i, hint := range e.RecoveryHints {
		msg += fmt.Sprintf("%d. %s\n", i+1, hint)
	}
	return msg
}
