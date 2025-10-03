package models

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors that can occur
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeProcessing ErrorType = "processing"
	ErrorTypeSystem     ErrorType = "system"
	ErrorTypeInput      ErrorType = "input"
	ErrorTypeTimeout    ErrorType = "timeout"
)

// ValidationError represents a validation error for a specific bidder and field
type ValidationError struct {
	BidderID string `json:"bidder_id"` // ID of the bidder with validation error
	Field    string `json:"field"`     // Field that failed validation
	Message  string `json:"message"`   // Error message describing the validation failure
	Value    string `json:"value"`     // The invalid value that caused the error
}

// NewValidationError creates a new ValidationError
func NewValidationError(bidderID, field, message string) *ValidationError {
	return &ValidationError{
		BidderID: bidderID,
		Field:    field,
		Message:  message,
	}
}

// NewValidationErrorWithValue creates a new ValidationError with the invalid value
func NewValidationErrorWithValue(bidderID, field, message, value string) *ValidationError {
	return &ValidationError{
		BidderID: bidderID,
		Field:    field,
		Message:  message,
		Value:    value,
	}
}

// Error implements the error interface for ValidationError
func (ve *ValidationError) Error() string {
	if ve.Value != "" {
		return fmt.Sprintf("validation error for bidder %s, field %s: %s (value: %s)", ve.BidderID, ve.Field, ve.Message, ve.Value)
	}
	return fmt.Sprintf("validation error for bidder %s, field %s: %s", ve.BidderID, ve.Field, ve.Message)
}

// AuctionError represents different types of errors that can occur during auction processing
type AuctionError struct {
	Type      ErrorType          `json:"type"`      // Type of error (validation, processing, system, etc.)
	Message   string             `json:"message"`   // Main error message
	Details   []*ValidationError `json:"details"`   // Detailed validation errors
	Cause     error              `json:"-"`         // Underlying cause of the error (not serialized)
	Context   map[string]string  `json:"context"`   // Additional context information
	Operation string             `json:"operation"` // Operation that was being performed when error occurred
}

// NewAuctionError creates a new AuctionError
func NewAuctionError(errorType ErrorType, message string, details []*ValidationError) *AuctionError {
	return &AuctionError{
		Type:    errorType,
		Message: message,
		Details: details,
		Context: make(map[string]string),
	}
}

// NewAuctionErrorWithCause creates a new AuctionError with an underlying cause
func NewAuctionErrorWithCause(errorType ErrorType, message string, cause error) *AuctionError {
	return &AuctionError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]string),
	}
}

// Error implements the error interface for AuctionError
func (ae *AuctionError) Error() string {
	var parts []string

	// Add error type and message
	parts = append(parts, fmt.Sprintf("%s error: %s", ae.Type, ae.Message))

	// Add operation context if available
	if ae.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation: %s", ae.Operation))
	}

	// Add validation error count if present
	if len(ae.Details) > 0 {
		parts = append(parts, fmt.Sprintf("validation errors: %d", len(ae.Details)))
	}

	// Add underlying cause if present
	if ae.Cause != nil {
		parts = append(parts, fmt.Sprintf("caused by: %s", ae.Cause.Error()))
	}

	return strings.Join(parts, "; ")
}

// Unwrap returns the underlying cause of the error for error wrapping
func (ae *AuctionError) Unwrap() error {
	return ae.Cause
}

// AddValidationError adds a validation error to the auction error
func (ae *AuctionError) AddValidationError(bidderID, field, message string) {
	ae.Details = append(ae.Details, NewValidationError(bidderID, field, message))
}

// AddValidationErrorWithValue adds a validation error with the invalid value
func (ae *AuctionError) AddValidationErrorWithValue(bidderID, field, message, value string) {
	ae.Details = append(ae.Details, NewValidationErrorWithValue(bidderID, field, message, value))
}

// HasValidationErrors returns true if there are validation errors
func (ae *AuctionError) HasValidationErrors() bool {
	return len(ae.Details) > 0
}

// AddContext adds context information to the error
func (ae *AuctionError) AddContext(key, value string) {
	if ae.Context == nil {
		ae.Context = make(map[string]string)
	}
	ae.Context[key] = value
}

// GetContext retrieves context information from the error
func (ae *AuctionError) GetContext(key string) (string, bool) {
	if ae.Context == nil {
		return "", false
	}
	value, exists := ae.Context[key]
	return value, exists
}

// WithOperation sets the operation context for the error
func (ae *AuctionError) WithOperation(operation string) *AuctionError {
	ae.Operation = operation
	return ae
}

// WithContext adds multiple context values to the error
func (ae *AuctionError) WithContext(context map[string]string) *AuctionError {
	if ae.Context == nil {
		ae.Context = make(map[string]string)
	}
	for k, v := range context {
		ae.Context[k] = v
	}
	return ae
}

// GetValidationErrorsByField returns validation errors grouped by field
func (ae *AuctionError) GetValidationErrorsByField() map[string][]*ValidationError {
	result := make(map[string][]*ValidationError)
	for _, detail := range ae.Details {
		result[detail.Field] = append(result[detail.Field], detail)
	}
	return result
}

// GetValidationErrorsByBidder returns validation errors grouped by bidder ID
func (ae *AuctionError) GetValidationErrorsByBidder() map[string][]*ValidationError {
	result := make(map[string][]*ValidationError)
	for _, detail := range ae.Details {
		result[detail.BidderID] = append(result[detail.BidderID], detail)
	}
	return result
}

// ProcessingError represents errors that occur during bid processing
type ProcessingError struct {
	*AuctionError
	BidderCount  int    `json:"bidder_count"`
	CurrentRound int    `json:"current_round"`
	FailedBidder string `json:"failed_bidder,omitempty"`
}

// NewProcessingError creates a new ProcessingError
func NewProcessingError(message string, bidderCount, currentRound int) *ProcessingError {
	return &ProcessingError{
		AuctionError: NewAuctionError(ErrorTypeProcessing, message, nil),
		BidderCount:  bidderCount,
		CurrentRound: currentRound,
	}
}

// NewProcessingErrorWithCause creates a new ProcessingError with an underlying cause
func NewProcessingErrorWithCause(message string, cause error, bidderCount, currentRound int) *ProcessingError {
	return &ProcessingError{
		AuctionError: NewAuctionErrorWithCause(ErrorTypeProcessing, message, cause),
		BidderCount:  bidderCount,
		CurrentRound: currentRound,
	}
}

// SystemError represents system-level errors
type SystemError struct {
	*AuctionError
	Component string `json:"component"`
	Severity  string `json:"severity"` // "low", "medium", "high", "critical"
}

// NewSystemError creates a new SystemError
func NewSystemError(message, component, severity string) *SystemError {
	return &SystemError{
		AuctionError: NewAuctionError(ErrorTypeSystem, message, nil),
		Component:    component,
		Severity:     severity,
	}
}

// NewSystemErrorWithCause creates a new SystemError with an underlying cause
func NewSystemErrorWithCause(message, component, severity string, cause error) *SystemError {
	return &SystemError{
		AuctionError: NewAuctionErrorWithCause(ErrorTypeSystem, message, cause),
		Component:    component,
		Severity:     severity,
	}
}

// InputError represents errors in user input
type InputError struct {
	*AuctionError
	InputField string      `json:"input_field"`
	InputValue interface{} `json:"input_value"`
}

// NewInputError creates a new InputError
func NewInputError(message, inputField string, inputValue interface{}) *InputError {
	return &InputError{
		AuctionError: NewAuctionError(ErrorTypeInput, message, nil),
		InputField:   inputField,
		InputValue:   inputValue,
	}
}

// TimeoutError represents timeout errors during processing
type TimeoutError struct {
	*AuctionError
	TimeoutDuration string `json:"timeout_duration"`
	Operation       string `json:"operation"`
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(message, operation, duration string) *TimeoutError {
	return &TimeoutError{
		AuctionError:    NewAuctionError(ErrorTypeTimeout, message, nil),
		TimeoutDuration: duration,
		Operation:       operation,
	}
}
