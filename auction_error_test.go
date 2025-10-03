package auction

import (
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

// MockBidValidator for testing error scenarios
type MockBidValidator struct {
	shouldError bool
	errorType   models.ErrorType
	errorMsg    string
}

func (m *MockBidValidator) ValidateBidder(bidder models.Bidder) error {
	if m.shouldError {
		return models.NewAuctionError(m.errorType, m.errorMsg, nil)
	}
	return nil
}

func (m *MockBidValidator) ValidateBidders(bidders []models.Bidder) error {
	if m.shouldError {
		auctionErr := models.NewAuctionError(m.errorType, m.errorMsg, nil)
		auctionErr.WithOperation("ValidateBidders")
		return auctionErr
	}
	return nil
}

// MockBiddingEngine for testing error scenarios
type MockBiddingEngine struct {
	shouldError bool
	errorType   models.ErrorType
	errorMsg    string
	result      *models.BidResult
}

func (m *MockBiddingEngine) ProcessBids(bidders []models.Bidder) (*models.BidResult, error) {
	if m.shouldError {
		return nil, models.NewAuctionError(m.errorType, m.errorMsg, nil)
	}
	return m.result, nil
}

// CustomErrorValidator for testing unexpected error types
type CustomErrorValidator struct {
	standardError error
}

func (c *CustomErrorValidator) ValidateBidder(bidder models.Bidder) error {
	return c.standardError
}

func (c *CustomErrorValidator) ValidateBidders(bidders []models.Bidder) error {
	return c.standardError
}

// ContextualErrorValidator for testing context propagation
type ContextualErrorValidator struct {
	baseError *models.AuctionError
}

func (c *ContextualErrorValidator) ValidateBidder(bidder models.Bidder) error {
	return c.baseError
}

func (c *ContextualErrorValidator) ValidateBidders(bidders []models.Bidder) error {
	return c.baseError
}

// TestRealWorldErrorScenarios tests realistic error scenarios
func TestRealWorldErrorScenarios(t *testing.T) {
	service := NewAuctionService()

	tests := []struct {
		name              string
		bidders           []models.Bidder
		expectError       bool
		expectedErrorType models.ErrorType
		validateContext   func(*models.AuctionError) bool
	}{
		{
			name:              "empty bidders list",
			bidders:           []models.Bidder{},
			expectError:       true,
			expectedErrorType: models.ErrorTypeValidation,
			validateContext: func(err *models.AuctionError) bool {
				count, exists := err.GetContext("bidder_count")
				return exists && count == "0"
			},
		},
		{
			name: "invalid bidder data",
			bidders: []models.Bidder{
				{
					ID:            "",
					Name:          "",
					StartingBid:   -100.0,
					MaxBid:        -500.0,
					AutoIncrement: 0.0,
				},
			},
			expectError:       true,
			expectedErrorType: models.ErrorTypeValidation,
			validateContext: func(err *models.AuctionError) bool {
				totalBidders, exists1 := err.GetContext("total_bidders")
				validBidders, exists2 := err.GetContext("valid_bidders")
				return exists1 && exists2 && totalBidders == "1" && validBidders == "0"
			},
		},
		{
			name: "duplicate bidder IDs",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "Alice",
					StartingBid:   100.0,
					MaxBid:        200.0,
					AutoIncrement: 10.0,
				},
				{
					ID:            "bidder1", // Duplicate
					Name:          "Bob",
					StartingBid:   110.0,
					MaxBid:        220.0,
					AutoIncrement: 15.0,
				},
			},
			expectError:       true,
			expectedErrorType: models.ErrorTypeValidation,
			validateContext: func(err *models.AuctionError) bool {
				totalBidders, exists1 := err.GetContext("total_bidders")
				invalidBidders, exists2 := err.GetContext("invalid_bidders")
				return exists1 && exists2 && totalBidders == "2" && invalidBidders == "1"
			},
		},
		{
			name: "valid single bidder",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "Alice",
					StartingBid:   100.0,
					MaxBid:        200.0,
					AutoIncrement: 10.0,
					EntryTime:     time.Now(),
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.DetermineWinner(tt.bidders)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				auctionErr, ok := err.(*models.AuctionError)
				if !ok {
					t.Fatalf("Expected AuctionError but got %T", err)
				}

				if auctionErr.Type != tt.expectedErrorType {
					t.Errorf("Expected error type %s, got %s", tt.expectedErrorType, auctionErr.Type)
				}

				// Validate context if provided
				if tt.validateContext != nil && !tt.validateContext(auctionErr) {
					t.Error("Context validation failed")
				}

				// Check that error message is descriptive
				errorMsg := auctionErr.Error()
				if len(errorMsg) < 10 {
					t.Errorf("Error message seems too short: %s", errorMsg)
				}

				if result != nil {
					t.Error("Expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

// TestErrorMessageFormatting tests that error messages are well-formatted
func TestErrorMessageFormatting(t *testing.T) {
	service := NewAuctionService()

	bidders := []models.Bidder{
		{
			ID:            "",
			Name:          "",
			StartingBid:   -100.0,
			MaxBid:        -500.0,
			AutoIncrement: 0.0,
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected error but got none")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError but got %T", err)
	}

	errorMsg := auctionErr.Error()

	// Check that error message contains key information
	expectedParts := []string{
		"validation error",
		"DetermineWinner.Validation",
		"validation errors:",
	}

	for _, part := range expectedParts {
		if !contains(errorMsg, part) {
			t.Errorf("Error message should contain '%s': %s", part, errorMsg)
		}
	}

	// Check that validation details are present
	if len(auctionErr.Details) == 0 {
		t.Error("Expected validation error details")
	}

	// Check that each validation error has proper formatting
	for _, detail := range auctionErr.Details {
		detailMsg := detail.Error()
		if !contains(detailMsg, "validation error") {
			t.Errorf("Validation detail should contain 'validation error': %s", detailMsg)
		}
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
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

// Benchmark error handling in auction service
func BenchmarkAuctionService_ErrorHandling(b *testing.B) {
	service := NewAuctionService()
	bidders := []models.Bidder{
		{
			ID:            "",
			Name:          "",
			StartingBid:   -100.0,
			MaxBid:        500.0,
			AutoIncrement: 0.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := service.DetermineWinner(bidders)
		if err != nil {
			_ = err.Error() // Force error formatting
		}
		_ = result
	}
}

func BenchmarkAuctionService_SuccessPath(b *testing.B) {
	service := NewAuctionService()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        200.0,
			AutoIncrement: 10.0,
			EntryTime:     time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := service.DetermineWinner(bidders)
		_ = result
		_ = err
	}
}
