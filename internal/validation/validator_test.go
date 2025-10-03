package validation

import (
	"fmt"
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

func TestDefaultBidValidator_ValidateBidder(t *testing.T) {
	validator := NewBidValidator()

	tests := []struct {
		name        string
		bidder      models.Bidder
		expectError bool
		errorCount  int
		description string
	}{
		{
			name: "valid bidder",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
				EntryTime:     time.Now(),
			},
			expectError: false,
			errorCount:  0,
			description: "should pass validation with all valid parameters",
		},
		{
			name: "missing bidder ID",
			bidder: models.Bidder{
				ID:            "",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when bidder ID is empty",
		},
		{
			name: "missing bidder name",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when bidder name is empty",
		},
		{
			name: "negative starting bid",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   -100.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when starting bid is negative (Requirement 6.3)",
		},
		{
			name: "negative maximum bid",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        -500.0,
				AutoIncrement: 25.0,
			},
			expectError: true,
			errorCount:  2, // MaxBid negative + StartingBid > MaxBid
			description: "should fail when maximum bid is negative (Requirement 6.3)",
		},
		{
			name: "zero auto-increment",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 0.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when auto-increment is zero (Requirement 6.2)",
		},
		{
			name: "negative auto-increment",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: -25.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when auto-increment is negative (Requirement 6.2)",
		},
		{
			name: "starting bid greater than max bid",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   600.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when starting bid exceeds maximum bid (Requirement 6.1)",
		},
		{
			name: "multiple validation errors",
			bidder: models.Bidder{
				ID:            "",
				Name:          "",
				StartingBid:   -100.0,
				MaxBid:        -500.0,
				AutoIncrement: 0.0,
			},
			expectError: true,
			errorCount:  6, // ID, Name, StartingBid negative, MaxBid negative, AutoIncrement zero, StartingBid > MaxBid
			description: "should collect multiple validation errors",
		},
		{
			name: "edge case: starting bid equals max bid",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   500.0,
				MaxBid:        500.0,
				AutoIncrement: 25.0,
			},
			expectError: false,
			errorCount:  0,
			description: "should pass when starting bid equals maximum bid",
		},
		{
			name: "edge case: very small auto-increment",
			bidder: models.Bidder{
				ID:            "bidder1",
				Name:          "John Doe",
				StartingBid:   100.0,
				MaxBid:        500.0,
				AutoIncrement: 0.01,
			},
			expectError: false,
			errorCount:  0,
			description: "should pass with very small positive auto-increment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBidder(tt.bidder)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none: %s", tt.description)
					return
				}

				auctionErr, ok := err.(*models.AuctionError)
				if !ok {
					t.Errorf("expected AuctionError but got %T", err)
					return
				}

				if len(auctionErr.Details) != tt.errorCount {
					t.Errorf("expected %d validation errors but got %d", tt.errorCount, len(auctionErr.Details))
				}

				if auctionErr.Type != models.ErrorTypeValidation {
					t.Errorf("expected error type 'validation' but got '%s'", auctionErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v (%s)", err, tt.description)
				}
			}
		})
	}
}

func TestDefaultBidValidator_ValidateBidders(t *testing.T) {
	validator := NewBidValidator()

	tests := []struct {
		name        string
		bidders     []models.Bidder
		expectError bool
		errorCount  int
		description string
	}{
		{
			name: "valid multiple bidders",
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
					StartingBid:   150.0,
					MaxBid:        600.0,
					AutoIncrement: 50.0,
				},
			},
			expectError: false,
			errorCount:  0,
			description: "should pass validation with multiple valid bidders",
		},
		{
			name:        "empty bidders list",
			bidders:     []models.Bidder{},
			expectError: true,
			errorCount:  0,
			description: "should fail when no bidders are provided (Requirement 6.4)",
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
					ID:            "bidder1",
					Name:          "Jane Smith",
					StartingBid:   150.0,
					MaxBid:        600.0,
					AutoIncrement: 50.0,
				},
			},
			expectError: true,
			errorCount:  1,
			description: "should fail when bidder IDs are duplicated",
		},
		{
			name: "mixed valid and invalid bidders",
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
					StartingBid:   -150.0, // Invalid: negative starting bid
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
			expectError: true,
			errorCount:  3,
			description: "should collect errors from multiple invalid bidders",
		},
		{
			name: "single bidder",
			bidders: []models.Bidder{
				{
					ID:            "bidder1",
					Name:          "John Doe",
					StartingBid:   100.0,
					MaxBid:        500.0,
					AutoIncrement: 25.0,
				},
			},
			expectError: false,
			errorCount:  0,
			description: "should pass validation with single valid bidder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBidders(tt.bidders)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none: %s", tt.description)
					return
				}

				auctionErr, ok := err.(*models.AuctionError)
				if !ok {
					t.Errorf("expected AuctionError but got %T", err)
					return
				}

				if len(auctionErr.Details) != tt.errorCount {
					t.Errorf("expected %d validation errors but got %d", tt.errorCount, len(auctionErr.Details))
				}

				if auctionErr.Type != models.ErrorTypeValidation {
					t.Errorf("expected error type 'validation' but got '%s'", auctionErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v (%s)", err, tt.description)
				}
			}
		})
	}
}

// TestValidationErrorDetails tests that validation errors contain proper details
func TestValidationErrorDetails(t *testing.T) {
	validator := NewBidValidator()

	bidder := models.Bidder{
		ID:            "",
		Name:          "",
		StartingBid:   -100.0,
		MaxBid:        -500.0,
		AutoIncrement: 0.0,
	}

	err := validator.ValidateBidder(bidder)
	if err == nil {
		t.Fatal("expected validation error but got none")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("expected AuctionError but got %T", err)
	}

	// Check that we have the expected validation errors
	expectedFields := map[string]bool{
		"ID":            false,
		"Name":          false,
		"StartingBid":   false,
		"MaxBid":        false,
		"AutoIncrement": false,
	}

	for _, detail := range auctionErr.Details {
		if _, exists := expectedFields[detail.Field]; exists {
			expectedFields[detail.Field] = true
		}
	}

	for field, found := range expectedFields {
		if !found {
			t.Errorf("expected validation error for field %s but didn't find it", field)
		}
	}
}

// BenchmarkValidateBidder benchmarks the validation performance
func BenchmarkValidateBidder(b *testing.B) {
	validator := NewBidValidator()
	bidder := models.Bidder{
		ID:            "bidder1",
		Name:          "John Doe",
		StartingBid:   100.0,
		MaxBid:        500.0,
		AutoIncrement: 25.0,
		EntryTime:     time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateBidder(bidder)
	}
}

// BenchmarkValidateBidders benchmarks validation of multiple bidders
func BenchmarkValidateBidders(b *testing.B) {
	validator := NewBidValidator()
	bidders := make([]models.Bidder, 100)

	for i := range 100 {
		bidders[i] = models.Bidder{
			ID:            fmt.Sprintf("bidder%d", i),
			Name:          fmt.Sprintf("Bidder %d", i),
			StartingBid:   float64(100 + i),
			MaxBid:        float64(500 + i*10),
			AutoIncrement: 25.0,
			EntryTime:     time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateBidders(bidders)
	}
}
