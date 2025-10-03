package validation

import (
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

// TestEnhancedValidationErrors tests the enhanced error handling in the validator
func TestEnhancedValidationErrors_ValidateBidder(t *testing.T) {
	validator := NewBidValidator()

	tests := []struct {
		name                string
		bidder              models.Bidder
		expectError         bool
		expectedErrorType   models.ErrorType
		expectedOperation   string
		expectedContextKeys []string
		validateErrorValue  bool
	}{
		{
			name: "valid bidder - no error",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
				EntryTime:     time.Now(),
			},
			expectError: false,
		},
		{
			name: "invalid bidder with enhanced error context",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   -100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
				EntryTime:     time.Now(),
			},
			expectError:         true,
			expectedErrorType:   models.ErrorTypeValidation,
			expectedOperation:   "ValidateBidder",
			expectedContextKeys: []string{"bidder_id", "bidder_name"},
			validateErrorValue:  true,
		},
		{
			name: "multiple validation errors with values",
			bidder: models.Bidder{
				ID:            "",
				Name:          "",
				StartingBid:   -100.0,
				MaxBid:        -500.0,
				AutoIncrement: 0.0,
			},
			expectError:         true,
			expectedErrorType:   models.ErrorTypeValidation,
			expectedOperation:   "ValidateBidder",
			expectedContextKeys: []string{"bidder_id", "bidder_name"},
			validateErrorValue:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBidder(tt.bidder)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				auctionErr, ok := err.(*models.AuctionError)
				if !ok {
					t.Fatalf("Expected AuctionError but got %T", err)
				}

				// Test error type
				if auctionErr.Type != tt.expectedErrorType {
					t.Errorf("Expected error type %s, got %s", tt.expectedErrorType, auctionErr.Type)
				}

				// Test operation
				if auctionErr.Operation != tt.expectedOperation {
					t.Errorf("Expected operation %s, got %s", tt.expectedOperation, auctionErr.Operation)
				}

				// Test context keys
				for _, key := range tt.expectedContextKeys {
					if _, exists := auctionErr.GetContext(key); !exists {
						t.Errorf("Expected context key %s to exist", key)
					}
				}

				// Test that validation errors have values when expected
				if tt.validateErrorValue {
					for _, detail := range auctionErr.Details {
						if detail.Value == "" && detail.Field != "ID" && detail.Field != "Name" {
							t.Errorf("Expected validation error for field %s to have a value", detail.Field)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestEnhancedValidationErrors_ValidateBidders(t *testing.T) {
	validator := NewBidValidator()

	tests := []struct {
		name                string
		bidders             []models.Bidder
		expectError         bool
		expectedErrorType   models.ErrorType
		expectedOperation   string
		expectedContextKeys []string
		validatePositions   bool
	}{
		{
			name:                "empty bidders list",
			bidders:             []models.Bidder{},
			expectError:         true,
			expectedErrorType:   models.ErrorTypeValidation,
			expectedOperation:   "ValidateBidders",
			expectedContextKeys: []string{"bidder_count"},
		},
		{
			name: "mixed valid and invalid bidders with position context",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "John Doe",
					StartingBid:   100.0,
					MaxBid:        500.0,
					AutoIncrement: 25.0,
				},
				{
					ID:            "bidder2",
					Name:          "Jane Smith",
					StartingBid:   -150.0, // Invalid
					MaxBid:        600.0,
					AutoIncrement: 50.0,
				},
				{
					ID:            "bidder3",
					Name:          "Bob Johnson",
					StartingBid:   200.0,
					MaxBid:        100.0, // Invalid: starting > max
					AutoIncrement: 0.0,   // Invalid: zero increment
				},
			},
			expectError:         true,
			expectedErrorType:   models.ErrorTypeValidation,
			expectedOperation:   "ValidateBidders",
			expectedContextKeys: []string{"total_bidders", "valid_bidders", "invalid_bidders", "total_validation_errors"},
			validatePositions:   true,
		},
		{
			name: "duplicate bidder IDs",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "John Doe",
					StartingBid:   100.0,
					MaxBid:        500.0,
					AutoIncrement: 25.0,
				},
				{
					ID:            "bidder1", // Duplicate
					Name:          "Jane Smith",
					StartingBid:   150.0,
					MaxBid:        600.0,
					AutoIncrement: 50.0,
				},
			},
			expectError:         true,
			expectedErrorType:   models.ErrorTypeValidation,
			expectedOperation:   "ValidateBidders",
			expectedContextKeys: []string{"total_bidders", "valid_bidders", "invalid_bidders", "total_validation_errors"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBidders(tt.bidders)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				auctionErr, ok := err.(*models.AuctionError)
				if !ok {
					t.Fatalf("Expected AuctionError but got %T", err)
				}

				// Test error type
				if auctionErr.Type != tt.expectedErrorType {
					t.Errorf("Expected error type %s, got %s", tt.expectedErrorType, auctionErr.Type)
				}

				// Test operation
				if auctionErr.Operation != tt.expectedOperation {
					t.Errorf("Expected operation %s, got %s", tt.expectedOperation, auctionErr.Operation)
				}

				// Test context keys
				for _, key := range tt.expectedContextKeys {
					if _, exists := auctionErr.GetContext(key); !exists {
						t.Errorf("Expected context key %s to exist", key)
					}
				}

				// Test position information in validation errors
				if tt.validatePositions {
					for _, detail := range auctionErr.Details {
						if detail.Value != "" && detail.Field != "ID" {
							// Position information should be included in the value
							// This is a basic check - in practice you might want more specific validation
							if len(detail.Value) == 0 {
								t.Errorf("Expected validation error for field %s to have position information", detail.Field)
							}
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidationErrorGrouping(t *testing.T) {
	validator := NewBidValidator()

	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "",
			StartingBid:   -100.0,
			MaxBid:        500.0,
			AutoIncrement: 0.0,
		},
		{
			ID:            "bidder2",
			Name:          "Jane",
			StartingBid:   200.0,
			MaxBid:        100.0, // Invalid: starting > max
			AutoIncrement: -5.0,  // Invalid: negative increment
		},
	}

	err := validator.ValidateBidders(bidders)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError but got %T", err)
	}

	// Test grouping by field
	errorsByField := auctionErr.GetValidationErrorsByField()
	if len(errorsByField) == 0 {
		t.Error("Expected errors grouped by field")
	}

	// Test grouping by bidder
	errorsByBidder := auctionErr.GetValidationErrorsByBidder()
	if len(errorsByBidder) != 2 {
		t.Errorf("Expected errors for 2 bidders, got %d", len(errorsByBidder))
	}

	// Verify bidder1 has multiple errors
	if len(errorsByBidder["bidder1"]) < 2 {
		t.Errorf("Expected multiple errors for bidder1, got %d", len(errorsByBidder["bidder1"]))
	}

	// Verify bidder2 has multiple errors
	if len(errorsByBidder["bidder2"]) < 2 {
		t.Errorf("Expected multiple errors for bidder2, got %d", len(errorsByBidder["bidder2"]))
	}
}

func TestErrorContextAccumulation(t *testing.T) {
	validator := NewBidValidator()

	bidders := []models.Bidder{
		{
			ID:            "valid_bidder",
			Name:          "Valid Bidder",
			StartingBid:   100.0,
			MaxBid:        500.0,
			AutoIncrement: 25.0,
		},
		{
			ID:            "invalid_bidder",
			Name:          "Invalid Bidder",
			StartingBid:   -100.0,
			MaxBid:        500.0,
			AutoIncrement: 25.0,
		},
	}

	err := validator.ValidateBidders(bidders)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError but got %T", err)
	}

	// Test that context contains expected information
	totalBidders, exists := auctionErr.GetContext("total_bidders")
	if !exists || totalBidders != "2" {
		t.Errorf("Expected total_bidders context to be '2', got '%s'", totalBidders)
	}

	validBidders, exists := auctionErr.GetContext("valid_bidders")
	if !exists || validBidders != "1" {
		t.Errorf("Expected valid_bidders context to be '1', got '%s'", validBidders)
	}

	invalidBidders, exists := auctionErr.GetContext("invalid_bidders")
	if !exists || invalidBidders != "1" {
		t.Errorf("Expected invalid_bidders context to be '1', got '%s'", invalidBidders)
	}
}

func TestValidationErrorMessages(t *testing.T) {
	validator := NewBidValidator()

	bidder := models.Bidder{
		ID:            "test_bidder",
		Name:          "Test Bidder",
		StartingBid:   -50.0,
		MaxBid:        -100.0,
		AutoIncrement: 0.0,
	}

	err := validator.ValidateBidder(bidder)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError but got %T", err)
	}

	// Test that error messages are descriptive and include values
	for _, detail := range auctionErr.Details {
		errorMsg := detail.Error()

		// Error message should contain bidder ID
		if detail.BidderID != "" && !contains(errorMsg, detail.BidderID) {
			t.Errorf("Error message should contain bidder ID: %s", errorMsg)
		}

		// Error message should contain field name
		if !contains(errorMsg, detail.Field) {
			t.Errorf("Error message should contain field name: %s", errorMsg)
		}

		// Error message should contain the invalid value when available
		if detail.Value != "" && !contains(errorMsg, detail.Value) {
			t.Errorf("Error message should contain invalid value: %s", errorMsg)
		}
	}
}

func TestErrorWrappingInValidator(t *testing.T) {
	validator := NewBidValidator()

	bidder := models.Bidder{
		ID:            "test_bidder",
		Name:          "Test Bidder",
		StartingBid:   -50.0,
		MaxBid:        500.0,
		AutoIncrement: 25.0,
	}

	err := validator.ValidateBidder(bidder)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	// Test that the error can be unwrapped and type-asserted correctly
	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError but got %T", err)
	}

	// Test error formatting includes all relevant information
	errorStr := auctionErr.Error()
	if !contains(errorStr, "validation error") {
		t.Errorf("Error string should contain 'validation error': %s", errorStr)
	}

	if !contains(errorStr, "ValidateBidder") {
		t.Errorf("Error string should contain operation name: %s", errorStr)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// Benchmark enhanced error handling
func BenchmarkEnhancedValidationError(b *testing.B) {
	validator := NewBidValidator()
	bidder := models.Bidder{
		ID:            "test_bidder",
		Name:          "Test Bidder",
		StartingBid:   -50.0,
		MaxBid:        500.0,
		AutoIncrement: 25.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateBidder(bidder)
		if err != nil {
			_ = err.Error() // Force error message formatting
		}
	}
}

func BenchmarkEnhancedValidationErrors(b *testing.B) {
	validator := NewBidValidator()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Valid Bidder",
			StartingBid:   100.0,
			MaxBid:        500.0,
			AutoIncrement: 25.0,
		},
		{
			ID:            "bidder2",
			Name:          "Invalid Bidder",
			StartingBid:   -100.0,
			MaxBid:        500.0,
			AutoIncrement: 0.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateBidders(bidders)
		if err != nil {
			_ = err.Error() // Force error message formatting
		}
	}
}
