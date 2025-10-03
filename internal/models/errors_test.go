package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *ValidationError
		expected string
	}{
		{
			name: "validation error without value",
			error: &ValidationError{
				BidderID: "bidder1",
				Field:    "StartingBid",
				Message:  "starting bid cannot be negative",
			},
			expected: "validation error for bidder bidder1, field StartingBid: starting bid cannot be negative",
		},
		{
			name: "validation error with value",
			error: &ValidationError{
				BidderID: "bidder2",
				Field:    "MaxBid",
				Message:  "maximum bid cannot be negative",
				Value:    "-100.00",
			},
			expected: "validation error for bidder bidder2, field MaxBid: maximum bid cannot be negative (value: -100.00)",
		},
		{
			name: "validation error with empty bidder ID",
			error: &ValidationError{
				BidderID: "",
				Field:    "ID",
				Message:  "bidder ID is required",
				Value:    "",
			},
			expected: "validation error for bidder , field ID: bidder ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	ve := NewValidationError("bidder1", "StartingBid", "starting bid cannot be negative")

	if ve.BidderID != "bidder1" {
		t.Errorf("Expected BidderID 'bidder1', got '%s'", ve.BidderID)
	}
	if ve.Field != "StartingBid" {
		t.Errorf("Expected Field 'StartingBid', got '%s'", ve.Field)
	}
	if ve.Message != "starting bid cannot be negative" {
		t.Errorf("Expected Message 'starting bid cannot be negative', got '%s'", ve.Message)
	}
	if ve.Value != "" {
		t.Errorf("Expected empty Value, got '%s'", ve.Value)
	}
}

func TestNewValidationErrorWithValue(t *testing.T) {
	ve := NewValidationErrorWithValue("bidder1", "StartingBid", "starting bid cannot be negative", "-50.00")

	if ve.BidderID != "bidder1" {
		t.Errorf("Expected BidderID 'bidder1', got '%s'", ve.BidderID)
	}
	if ve.Field != "StartingBid" {
		t.Errorf("Expected Field 'StartingBid', got '%s'", ve.Field)
	}
	if ve.Message != "starting bid cannot be negative" {
		t.Errorf("Expected Message 'starting bid cannot be negative', got '%s'", ve.Message)
	}
	if ve.Value != "-50.00" {
		t.Errorf("Expected Value '-50.00', got '%s'", ve.Value)
	}
}

func TestAuctionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *AuctionError
		expected string
	}{
		{
			name: "simple auction error",
			error: &AuctionError{
				Type:    ErrorTypeValidation,
				Message: "validation failed",
			},
			expected: "validation error: validation failed",
		},
		{
			name: "auction error with operation",
			error: &AuctionError{
				Type:      ErrorTypeProcessing,
				Message:   "processing failed",
				Operation: "ProcessBids",
			},
			expected: "processing error: processing failed; operation: ProcessBids",
		},
		{
			name: "auction error with validation details",
			error: &AuctionError{
				Type:    ErrorTypeValidation,
				Message: "validation failed",
				Details: []*ValidationError{
					{BidderID: "bidder1", Field: "StartingBid", Message: "negative bid"},
					{BidderID: "bidder2", Field: "MaxBid", Message: "negative bid"},
				},
			},
			expected: "validation error: validation failed; validation errors: 2",
		},
		{
			name: "auction error with cause",
			error: &AuctionError{
				Type:    ErrorTypeSystem,
				Message: "system error",
				Cause:   errors.New("underlying error"),
			},
			expected: "system error: system error; caused by: underlying error",
		},
		{
			name: "auction error with all fields",
			error: &AuctionError{
				Type:      ErrorTypeProcessing,
				Message:   "complex error",
				Operation: "DetermineWinner",
				Details: []*ValidationError{
					{BidderID: "bidder1", Field: "StartingBid", Message: "invalid"},
				},
				Cause: errors.New("root cause"),
			},
			expected: "processing error: complex error; operation: DetermineWinner; validation errors: 1; caused by: root cause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAuctionError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	auctionErr := NewAuctionErrorWithCause(ErrorTypeSystem, "system error", cause)

	unwrapped := auctionErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be the original cause")
	}

	// Test with no cause
	auctionErrNoCause := NewAuctionError(ErrorTypeValidation, "validation error", nil)
	unwrappedNoCause := auctionErrNoCause.Unwrap()
	if unwrappedNoCause != nil {
		t.Errorf("Expected unwrapped error to be nil when no cause is set")
	}
}

func TestAuctionError_AddValidationError(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)

	auctionErr.AddValidationError("bidder1", "StartingBid", "negative bid")
	auctionErr.AddValidationError("bidder2", "MaxBid", "negative bid")

	if len(auctionErr.Details) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(auctionErr.Details))
	}

	if auctionErr.Details[0].BidderID != "bidder1" {
		t.Errorf("Expected first error bidder ID 'bidder1', got '%s'", auctionErr.Details[0].BidderID)
	}

	if auctionErr.Details[1].BidderID != "bidder2" {
		t.Errorf("Expected second error bidder ID 'bidder2', got '%s'", auctionErr.Details[1].BidderID)
	}
}

func TestAuctionError_AddValidationErrorWithValue(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)

	auctionErr.AddValidationErrorWithValue("bidder1", "StartingBid", "negative bid", "-50.00")

	if len(auctionErr.Details) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(auctionErr.Details))
	}

	if auctionErr.Details[0].Value != "-50.00" {
		t.Errorf("Expected validation error value '-50.00', got '%s'", auctionErr.Details[0].Value)
	}
}

func TestAuctionError_HasValidationErrors(t *testing.T) {
	// Test with no validation errors
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	if auctionErr.HasValidationErrors() {
		t.Error("Expected HasValidationErrors to return false when no errors present")
	}

	// Test with validation errors
	auctionErr.AddValidationError("bidder1", "StartingBid", "negative bid")
	if !auctionErr.HasValidationErrors() {
		t.Error("Expected HasValidationErrors to return true when errors present")
	}
}

func TestAuctionError_AddContext(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)

	auctionErr.AddContext("bidder_count", "5")
	auctionErr.AddContext("operation", "ValidateBidders")

	if len(auctionErr.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(auctionErr.Context))
	}

	if auctionErr.Context["bidder_count"] != "5" {
		t.Errorf("Expected context bidder_count '5', got '%s'", auctionErr.Context["bidder_count"])
	}

	if auctionErr.Context["operation"] != "ValidateBidders" {
		t.Errorf("Expected context operation 'ValidateBidders', got '%s'", auctionErr.Context["operation"])
	}
}

func TestAuctionError_GetContext(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	auctionErr.AddContext("test_key", "test_value")

	// Test existing key
	value, exists := auctionErr.GetContext("test_key")
	if !exists {
		t.Error("Expected context key to exist")
	}
	if value != "test_value" {
		t.Errorf("Expected context value 'test_value', got '%s'", value)
	}

	// Test non-existing key
	_, exists = auctionErr.GetContext("non_existing_key")
	if exists {
		t.Error("Expected context key to not exist")
	}
}

func TestAuctionError_WithOperation(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	result := auctionErr.WithOperation("TestOperation")

	if result != auctionErr {
		t.Error("Expected WithOperation to return the same instance")
	}

	if auctionErr.Operation != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got '%s'", auctionErr.Operation)
	}
}

func TestAuctionError_WithContext(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	context := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result := auctionErr.WithContext(context)

	if result != auctionErr {
		t.Error("Expected WithContext to return the same instance")
	}

	if len(auctionErr.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(auctionErr.Context))
	}

	if auctionErr.Context["key1"] != "value1" {
		t.Errorf("Expected context key1 'value1', got '%s'", auctionErr.Context["key1"])
	}
}

func TestAuctionError_GetValidationErrorsByField(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	auctionErr.AddValidationError("bidder1", "StartingBid", "negative bid")
	auctionErr.AddValidationError("bidder2", "StartingBid", "too high")
	auctionErr.AddValidationError("bidder1", "MaxBid", "negative bid")

	errorsByField := auctionErr.GetValidationErrorsByField()

	if len(errorsByField) != 2 {
		t.Errorf("Expected 2 fields with errors, got %d", len(errorsByField))
	}

	if len(errorsByField["StartingBid"]) != 2 {
		t.Errorf("Expected 2 StartingBid errors, got %d", len(errorsByField["StartingBid"]))
	}

	if len(errorsByField["MaxBid"]) != 1 {
		t.Errorf("Expected 1 MaxBid error, got %d", len(errorsByField["MaxBid"]))
	}
}

func TestAuctionError_GetValidationErrorsByBidder(t *testing.T) {
	auctionErr := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
	auctionErr.AddValidationError("bidder1", "StartingBid", "negative bid")
	auctionErr.AddValidationError("bidder1", "MaxBid", "negative bid")
	auctionErr.AddValidationError("bidder2", "StartingBid", "too high")

	errorsByBidder := auctionErr.GetValidationErrorsByBidder()

	if len(errorsByBidder) != 2 {
		t.Errorf("Expected 2 bidders with errors, got %d", len(errorsByBidder))
	}

	if len(errorsByBidder["bidder1"]) != 2 {
		t.Errorf("Expected 2 errors for bidder1, got %d", len(errorsByBidder["bidder1"]))
	}

	if len(errorsByBidder["bidder2"]) != 1 {
		t.Errorf("Expected 1 error for bidder2, got %d", len(errorsByBidder["bidder2"]))
	}
}

func TestProcessingError(t *testing.T) {
	processingErr := NewProcessingError("processing failed", 5, 10)

	if processingErr.Type != ErrorTypeProcessing {
		t.Errorf("Expected error type processing, got %s", processingErr.Type)
	}

	if processingErr.BidderCount != 5 {
		t.Errorf("Expected bidder count 5, got %d", processingErr.BidderCount)
	}

	if processingErr.CurrentRound != 10 {
		t.Errorf("Expected current round 10, got %d", processingErr.CurrentRound)
	}
}

func TestProcessingErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	processingErr := NewProcessingErrorWithCause("processing failed", cause, 5, 10)

	if processingErr.Cause != cause {
		t.Error("Expected cause to be set correctly")
	}

	if processingErr.BidderCount != 5 {
		t.Errorf("Expected bidder count 5, got %d", processingErr.BidderCount)
	}
}

func TestSystemError(t *testing.T) {
	systemErr := NewSystemError("system failure", "BiddingEngine", "critical")

	if systemErr.Type != ErrorTypeSystem {
		t.Errorf("Expected error type system, got %s", systemErr.Type)
	}

	if systemErr.Component != "BiddingEngine" {
		t.Errorf("Expected component 'BiddingEngine', got '%s'", systemErr.Component)
	}

	if systemErr.Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", systemErr.Severity)
	}
}

func TestSystemErrorWithCause(t *testing.T) {
	cause := errors.New("underlying system error")
	systemErr := NewSystemErrorWithCause("system failure", "BiddingEngine", "critical", cause)

	if systemErr.Cause != cause {
		t.Error("Expected cause to be set correctly")
	}

	if systemErr.Component != "BiddingEngine" {
		t.Errorf("Expected component 'BiddingEngine', got '%s'", systemErr.Component)
	}
}

func TestInputError(t *testing.T) {
	inputErr := NewInputError("invalid input", "bidders", []string{"bidder1", "bidder2"})

	if inputErr.Type != ErrorTypeInput {
		t.Errorf("Expected error type input, got %s", inputErr.Type)
	}

	if inputErr.InputField != "bidders" {
		t.Errorf("Expected input field 'bidders', got '%s'", inputErr.InputField)
	}

	if inputErr.InputValue == nil {
		t.Error("Expected input value to be set")
	}
}

func TestTimeoutError(t *testing.T) {
	timeoutErr := NewTimeoutError("operation timed out", "ProcessBids", "30 seconds")

	if timeoutErr.Type != ErrorTypeTimeout {
		t.Errorf("Expected error type timeout, got %s", timeoutErr.Type)
	}

	if timeoutErr.Operation != "ProcessBids" {
		t.Errorf("Expected operation 'ProcessBids', got '%s'", timeoutErr.Operation)
	}

	if timeoutErr.TimeoutDuration != "30 seconds" {
		t.Errorf("Expected timeout duration '30 seconds', got '%s'", timeoutErr.TimeoutDuration)
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// Test that error type constants are defined correctly
	if ErrorTypeValidation != "validation" {
		t.Errorf("Expected ErrorTypeValidation to be 'validation', got '%s'", ErrorTypeValidation)
	}

	if ErrorTypeProcessing != "processing" {
		t.Errorf("Expected ErrorTypeProcessing to be 'processing', got '%s'", ErrorTypeProcessing)
	}

	if ErrorTypeSystem != "system" {
		t.Errorf("Expected ErrorTypeSystem to be 'system', got '%s'", ErrorTypeSystem)
	}

	if ErrorTypeInput != "input" {
		t.Errorf("Expected ErrorTypeInput to be 'input', got '%s'", ErrorTypeInput)
	}

	if ErrorTypeTimeout != "timeout" {
		t.Errorf("Expected ErrorTypeTimeout to be 'timeout', got '%s'", ErrorTypeTimeout)
	}
}

// Test error wrapping compatibility with Go's error handling
func TestErrorWrapping(t *testing.T) {
	cause := errors.New("root cause")
	auctionErr := NewAuctionErrorWithCause(ErrorTypeSystem, "system error", cause)

	// Test that errors.Is works
	if !errors.Is(auctionErr, cause) {
		t.Error("Expected errors.Is to find the root cause")
	}

	// Test that errors.As works
	var targetErr *AuctionError
	if !errors.As(auctionErr, &targetErr) {
		t.Error("Expected errors.As to find AuctionError")
	}

	if targetErr.Message != "system error" {
		t.Errorf("Expected message 'system error', got '%s'", targetErr.Message)
	}
}

// Benchmark error creation and formatting
func BenchmarkAuctionErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := NewAuctionError(ErrorTypeValidation, "validation failed", nil)
		err.AddValidationError("bidder1", "StartingBid", "negative bid")
		_ = err.Error()
	}
}

func BenchmarkValidationErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := NewValidationErrorWithValue("bidder1", "StartingBid", "negative bid", "-50.00")
		_ = err.Error()
	}
}

func BenchmarkComplexErrorFormatting(b *testing.B) {
	err := NewAuctionError(ErrorTypeValidation, "complex validation failed", nil)
	err.WithOperation("ValidateBidders")
	err.AddContext("bidder_count", "100")
	err.AddContext("operation", "test")

	for i := 0; i < 10; i++ {
		err.AddValidationError(fmt.Sprintf("bidder%d", i), "StartingBid", "negative bid")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func TestAuctionError_AddContext_NilContext(t *testing.T) {
	err := &AuctionError{
		Type:    ErrorTypeValidation,
		Message: "test",
		Context: nil, // Explicitly nil
	}

	// Should initialize context map and add the value
	err.AddContext("key", "value")

	if err.Context == nil {
		t.Fatal("Expected context to be initialized")
	}

	value, exists := err.GetContext("key")
	if !exists {
		t.Error("Expected key to exist after AddContext")
	}
	if value != "value" {
		t.Errorf("Expected value 'value', got '%s'", value)
	}
}

// TestAuctionError_GetContext_NilContext tests GetContext with nil context map
func TestAuctionError_GetContext_NilContext(t *testing.T) {
	err := &AuctionError{
		Type:    ErrorTypeValidation,
		Message: "test",
		Context: nil, // Explicitly nil
	}

	value, exists := err.GetContext("nonexistent")
	if exists {
		t.Error("Expected key to not exist in nil context")
	}
	if value != "" {
		t.Errorf("Expected empty value, got '%s'", value)
	}
}

// TestAuctionError_WithContext_NilInitialContext tests WithContext with nil initial context
func TestAuctionError_WithContext_NilInitialContext(t *testing.T) {
	err := &AuctionError{
		Type:    ErrorTypeValidation,
		Message: "test",
		Context: nil, // Explicitly nil
	}

	newContext := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result := err.WithContext(newContext)

	// Should return the same error instance
	if result != err {
		t.Error("Expected WithContext to return the same error instance")
	}

	// Should initialize context and add values
	if err.Context == nil {
		t.Fatal("Expected context to be initialized")
	}

	for key, expectedValue := range newContext {
		actualValue, exists := err.GetContext(key)
		if !exists {
			t.Errorf("Expected context key '%s' to exist", key)
		}
		if actualValue != expectedValue {
			t.Errorf("Expected context value '%s', got '%s'", expectedValue, actualValue)
		}
	}
}

// TestAuctionError_WithContext_ExistingContext tests WithContext with existing context
func TestAuctionError_WithContext_ExistingContext(t *testing.T) {
	err := NewAuctionError(ErrorTypeValidation, "test", nil)
	err.AddContext("existing", "old_value")

	newContext := map[string]string{
		"existing": "new_value", // Should overwrite
		"new_key":  "new_value", // Should add
	}

	result := err.WithContext(newContext)

	// Should return the same error instance
	if result != err {
		t.Error("Expected WithContext to return the same error instance")
	}

	// Check overwritten value
	value, exists := err.GetContext("existing")
	if !exists {
		t.Error("Expected existing key to still exist")
	}
	if value != "new_value" {
		t.Errorf("Expected overwritten value 'new_value', got '%s'", value)
	}

	// Check new value
	value, exists = err.GetContext("new_key")
	if !exists {
		t.Error("Expected new key to exist")
	}
	if value != "new_value" {
		t.Errorf("Expected new value 'new_value', got '%s'", value)
	}
}

// TestAuctionError_WithContext_EmptyContext tests WithContext with empty context
func TestAuctionError_WithContext_EmptyContext(t *testing.T) {
	err := NewAuctionError(ErrorTypeValidation, "test", nil)
	err.AddContext("existing", "value")

	emptyContext := map[string]string{}
	result := err.WithContext(emptyContext)

	// Should return the same error instance
	if result != err {
		t.Error("Expected WithContext to return the same error instance")
	}

	// Existing context should remain
	value, exists := err.GetContext("existing")
	if !exists {
		t.Error("Expected existing context to remain")
	}
	if value != "value" {
		t.Errorf("Expected existing value 'value', got '%s'", value)
	}
}

// TestAuctionError_ContextChaining tests method chaining with context operations
func TestAuctionError_ContextChaining(t *testing.T) {
	err := NewAuctionError(ErrorTypeValidation, "test", nil).
		WithOperation("TestOp").
		WithContext(map[string]string{"key1": "value1"})

	err.AddContext("key2", "value2")

	// Test all values are present
	if err.Operation != "TestOp" {
		t.Errorf("Expected operation 'TestOp', got '%s'", err.Operation)
	}

	value1, exists := err.GetContext("key1")
	if !exists || value1 != "value1" {
		t.Error("Expected key1 to have value1")
	}

	value2, exists := err.GetContext("key2")
	if !exists || value2 != "value2" {
		t.Error("Expected key2 to have value2")
	}
}

// TestErrorTypeConstants_Coverage ensures all error types are tested
func TestErrorTypeConstants_Coverage(t *testing.T) {
	errorTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeProcessing,
		ErrorTypeSystem,
		ErrorTypeInput,
		ErrorTypeTimeout,
	}

	expectedValues := []string{
		"validation",
		"processing",
		"system",
		"input",
		"timeout",
	}

	if len(errorTypes) != len(expectedValues) {
		t.Fatalf("Mismatch in error type count: expected %d, got %d", len(expectedValues), len(errorTypes))
	}

	for i, errorType := range errorTypes {
		if string(errorType) != expectedValues[i] {
			t.Errorf("Error type %d: expected '%s', got '%s'", i, expectedValues[i], string(errorType))
		}
	}
}

// TestSpecializedErrorTypes_Coverage tests all specialized error constructors
func TestSpecializedErrorTypes_Coverage(t *testing.T) {
	// Test ProcessingError
	procErr := NewProcessingError("processing failed", 5, 10)
	if procErr.BidderCount != 5 {
		t.Errorf("Expected bidder count 5, got %d", procErr.BidderCount)
	}
	if procErr.CurrentRound != 10 {
		t.Errorf("Expected current round 10, got %d", procErr.CurrentRound)
	}

	// Test ProcessingErrorWithCause
	cause := NewAuctionError(ErrorTypeSystem, "underlying error", nil)
	procErrWithCause := NewProcessingErrorWithCause("processing failed", cause, 3, 7)
	if procErrWithCause.BidderCount != 3 {
		t.Errorf("Expected bidder count 3, got %d", procErrWithCause.BidderCount)
	}
	if procErrWithCause.CurrentRound != 7 {
		t.Errorf("Expected current round 7, got %d", procErrWithCause.CurrentRound)
	}
	if procErrWithCause.Unwrap() != cause {
		t.Error("Expected cause to be wrapped")
	}

	// Test SystemError
	sysErr := NewSystemError("system failure", "TestComponent", "high")
	if sysErr.Component != "TestComponent" {
		t.Errorf("Expected component 'TestComponent', got '%s'", sysErr.Component)
	}
	if sysErr.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", sysErr.Severity)
	}

	// Test SystemErrorWithCause
	sysErrWithCause := NewSystemErrorWithCause("system failure", "TestComponent", "critical", cause)
	if sysErrWithCause.Component != "TestComponent" {
		t.Errorf("Expected component 'TestComponent', got '%s'", sysErrWithCause.Component)
	}
	if sysErrWithCause.Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", sysErrWithCause.Severity)
	}
	if sysErrWithCause.Unwrap() != cause {
		t.Error("Expected cause to be wrapped")
	}

	// Test InputError
	inputErr := NewInputError("invalid input", "testField", "testValue")
	if inputErr.InputField != "testField" {
		t.Errorf("Expected input field 'testField', got '%s'", inputErr.InputField)
	}
	if inputErr.InputValue != "testValue" {
		t.Errorf("Expected input value 'testValue', got '%v'", inputErr.InputValue)
	}

	// Test TimeoutError
	timeoutErr := NewTimeoutError("operation timed out", "TestOperation", "30 seconds")
	if timeoutErr.Operation != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got '%s'", timeoutErr.Operation)
	}
	if timeoutErr.TimeoutDuration != "30 seconds" {
		t.Errorf("Expected timeout duration '30 seconds', got '%s'", timeoutErr.TimeoutDuration)
	}
}
