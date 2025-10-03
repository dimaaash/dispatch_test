package auction

import (
	"errors"
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

// MockValidator for testing error handling paths
type MockValidator struct {
	shouldReturnError      bool
	shouldReturnNonAuction bool
}

func (mv *MockValidator) ValidateBidder(bidder models.Bidder) error {
	if mv.shouldReturnError {
		if mv.shouldReturnNonAuction {
			return errors.New("unexpected validation error")
		}
		return models.NewAuctionError(models.ErrorTypeValidation, "mock validation error", nil)
	}
	return nil
}

func (mv *MockValidator) ValidateBidders(bidders []models.Bidder) error {
	if mv.shouldReturnError {
		if mv.shouldReturnNonAuction {
			return errors.New("unexpected validation error")
		}
		return models.NewAuctionError(models.ErrorTypeValidation, "mock validation error", nil)
	}
	return nil
}

// MockEngine for testing processing error paths
type MockEngine struct {
	shouldReturnError      bool
	shouldReturnNonAuction bool
	shouldReturnNilResult  bool
}

func (me *MockEngine) ProcessBids(bidders []models.Bidder) (*models.BidResult, error) {
	if me.shouldReturnError {
		if me.shouldReturnNonAuction {
			return nil, errors.New("unexpected processing error")
		}
		return nil, models.NewAuctionError(models.ErrorTypeProcessing, "mock processing error", nil)
	}
	if me.shouldReturnNilResult {
		return nil, nil
	}
	// Return a valid result
	winner := &bidders[0]
	return models.NewBidResult(winner, winner.StartingBid, len(bidders), 0, bidders), nil
}

// TestDetermineWinner_ValidationErrorWrapping tests validation error wrapping
func TestDetermineWinner_ValidationErrorWrapping(t *testing.T) {
	mockValidator := &MockValidator{shouldReturnError: true, shouldReturnNonAuction: false}
	mockEngine := &MockEngine{}

	// Create service with mock dependencies manually
	service := &AuctionService{
		validator: mockValidator,
		engine:    mockEngine,
	}

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

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when validation fails")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError, got %T", err)
	}

	if auctionErr.Operation != "DetermineWinner.Validation" {
		t.Errorf("Expected operation 'DetermineWinner.Validation', got '%s'", auctionErr.Operation)
	}

	service_context, exists := auctionErr.GetContext("service")
	if !exists || service_context != "AuctionService" {
		t.Errorf("Expected service context 'AuctionService', got '%s'", service_context)
	}
}

// TestDetermineWinner_UnexpectedValidationError tests handling of non-AuctionError validation errors
func TestDetermineWinner_UnexpectedValidationError(t *testing.T) {
	mockValidator := &MockValidator{shouldReturnError: true, shouldReturnNonAuction: true}
	mockEngine := &MockEngine{}

	// Create service with mock dependencies manually
	service := &AuctionService{
		validator: mockValidator,
		engine:    mockEngine,
	}

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

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when validation fails")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected wrapped AuctionError, got %T", err)
	}

	if auctionErr.Type != models.ErrorTypeValidation {
		t.Errorf("Expected validation error type, got '%s'", auctionErr.Type)
	}

	if auctionErr.Message != "unexpected validation error" {
		t.Errorf("Expected 'unexpected validation error', got '%s'", auctionErr.Message)
	}

	if auctionErr.Operation != "DetermineWinner.Validation" {
		t.Errorf("Expected operation 'DetermineWinner.Validation', got '%s'", auctionErr.Operation)
	}

	// Check that the original error is wrapped
	if auctionErr.Unwrap() == nil {
		t.Error("Expected wrapped error to be available")
	}
}

// TestDetermineWinner_ProcessingErrorWrapping tests processing error wrapping
func TestDetermineWinner_ProcessingErrorWrapping(t *testing.T) {
	mockValidator := &MockValidator{shouldReturnError: false}
	mockEngine := &MockEngine{shouldReturnError: true, shouldReturnNonAuction: false}

	// Create service with mock dependencies manually
	service := &AuctionService{
		validator: mockValidator,
		engine:    mockEngine,
	}

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

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected processing error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when processing fails")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError, got %T", err)
	}

	if auctionErr.Operation != "DetermineWinner.Processing" {
		t.Errorf("Expected operation 'DetermineWinner.Processing', got '%s'", auctionErr.Operation)
	}

	service_context, exists := auctionErr.GetContext("service")
	if !exists || service_context != "AuctionService" {
		t.Errorf("Expected service context 'AuctionService', got '%s'", service_context)
	}
}

// TestDetermineWinner_UnexpectedProcessingError tests handling of non-AuctionError processing errors
func TestDetermineWinner_UnexpectedProcessingError(t *testing.T) {
	mockValidator := &MockValidator{shouldReturnError: false}
	mockEngine := &MockEngine{shouldReturnError: true, shouldReturnNonAuction: true}

	// Create service with mock dependencies manually
	service := &AuctionService{
		validator: mockValidator,
		engine:    mockEngine,
	}

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

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected processing error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when processing fails")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected wrapped AuctionError, got %T", err)
	}

	if auctionErr.Type != models.ErrorTypeProcessing {
		t.Errorf("Expected processing error type, got '%s'", auctionErr.Type)
	}

	if auctionErr.Message != "unexpected processing error" {
		t.Errorf("Expected 'unexpected processing error', got '%s'", auctionErr.Message)
	}

	if auctionErr.Operation != "DetermineWinner.Processing" {
		t.Errorf("Expected operation 'DetermineWinner.Processing', got '%s'", auctionErr.Operation)
	}

	// Check that the original error is wrapped
	if auctionErr.Unwrap() == nil {
		t.Error("Expected wrapped error to be available")
	}
}

// TestDetermineWinner_NilResultError tests handling when engine returns nil result
func TestDetermineWinner_NilResultError(t *testing.T) {
	mockValidator := &MockValidator{shouldReturnError: false}
	mockEngine := &MockEngine{shouldReturnNilResult: true}

	// Create service with mock dependencies manually
	service := &AuctionService{
		validator: mockValidator,
		engine:    mockEngine,
	}

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

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected processing error for nil result, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when engine returns nil")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError, got %T", err)
	}

	if auctionErr.Type != models.ErrorTypeProcessing {
		t.Errorf("Expected processing error type, got '%s'", auctionErr.Type)
	}

	if auctionErr.Message != "failed to process bids: result is nil" {
		t.Errorf("Expected 'failed to process bids: result is nil', got '%s'", auctionErr.Message)
	}

	if auctionErr.Operation != "DetermineWinner.ResultValidation" {
		t.Errorf("Expected operation 'DetermineWinner.ResultValidation', got '%s'", auctionErr.Operation)
	}

	// Check context
	service_context, exists := auctionErr.GetContext("service")
	if !exists || service_context != "AuctionService" {
		t.Errorf("Expected service context 'AuctionService', got '%s'", service_context)
	}

	bidder_count, exists := auctionErr.GetContext("bidder_count")
	if !exists || bidder_count != "1" {
		t.Errorf("Expected bidder_count context '1', got '%s'", bidder_count)
	}
}

// TestDetermineWinner_SuccessfulPath tests the complete successful execution path
func TestDetermineWinner_SuccessfulPath(t *testing.T) {
	service := NewAuctionService()

	baseTime := time.Now()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        300.0,
			AutoIncrement: 25.0,
			EntryTime:     baseTime,
		},
		{
			ID:            "bidder2",
			Name:          "Bob",
			StartingBid:   110.0,
			MaxBid:        250.0,
			AutoIncrement: 20.0,
			EntryTime:     baseTime.Add(1 * time.Second),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Verify the result structure
	if result.TotalBidders != 2 {
		t.Errorf("Expected total bidders 2, got %d", result.TotalBidders)
	}

	if len(result.AllBidders) != 2 {
		t.Errorf("Expected 2 bidders in result, got %d", len(result.AllBidders))
	}

	// Alice should win with higher max bid
	if result.Winner.ID != "bidder1" {
		t.Errorf("Expected winner 'bidder1', got '%s'", result.Winner.ID)
	}
}

// TestDetermineWinner_EdgeCaseScenarios tests various edge cases
func TestDetermineWinner_EdgeCaseScenarios(t *testing.T) {
	tests := []struct {
		name           string
		bidders        []models.Bidder
		expectError    bool
		expectedWinner string
	}{
		{
			name: "Single bidder with minimum values",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "Alice",
					StartingBid:   0.01,
					MaxBid:        0.01,
					AutoIncrement: 0.01,
					EntryTime:     time.Now(),
				},
			},
			expectError:    false,
			expectedWinner: "bidder1",
		},
		{
			name: "Multiple bidders with same starting bid",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "Alice",
					StartingBid:   100.0,
					MaxBid:        200.0,
					AutoIncrement: 10.0,
					EntryTime:     time.Now(),
				},
				{
					ID:            "bidder2",
					Name:          "Bob",
					StartingBid:   100.0,
					MaxBid:        180.0,
					AutoIncrement: 5.0,
					EntryTime:     time.Now().Add(1 * time.Second),
				},
			},
			expectError:    false,
			expectedWinner: "bidder1", // Higher max bid
		},
		{
			name: "Bidders with very large values",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "Alice",
					StartingBid:   999999.99,
					MaxBid:        1000000.00,
					AutoIncrement: 0.01,
					EntryTime:     time.Now(),
				},
			},
			expectError:    false,
			expectedWinner: "bidder1",
		},
	}

	service := NewAuctionService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.DetermineWinner(tt.bidders)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Winner == nil {
				t.Fatal("Expected winner, got nil")
			}

			if result.Winner.ID != tt.expectedWinner {
				t.Errorf("Expected winner '%s', got '%s'", tt.expectedWinner, result.Winner.ID)
			}
		})
	}
}

// TestDetermineWinner_PrecisionHandling tests that the service handles precision correctly
func TestDetermineWinner_PrecisionHandling(t *testing.T) {
	service := NewAuctionService()

	baseTime := time.Now()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   10.01,
			MaxBid:        20.99,
			AutoIncrement: 0.25,
			EntryTime:     baseTime,
		},
		{
			ID:            "bidder2",
			Name:          "Bob",
			StartingBid:   10.02,
			MaxBid:        19.98,
			AutoIncrement: 0.33,
			EntryTime:     baseTime.Add(1 * time.Second),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Alice should win with higher max bid
	if result.Winner.ID != "bidder1" {
		t.Errorf("Expected winner 'bidder1', got '%s'", result.Winner.ID)
	}

	// Verify winning bid is reasonable (should be Bob's max + Alice's increment)
	expectedWinningBid := 19.98 + 0.25 // 20.23
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", expectedWinningBid, result.WinningBid)
	}
}
