package internal

import (
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

// TestFloatingPointPrecision tests that monetary calculations are handled with proper precision
func TestFloatingPointPrecision(t *testing.T) {
	tests := []struct {
		name               string
		bidders            []models.Bidder
		expectedWinnerID   string
		expectedWinningBid float64
		description        string
	}{
		{
			name: "Precise decimal calculations with small increments",
			bidders: []models.Bidder{
				*models.NewBidder("1", "Alice", 10.01, 10.99, 0.01),
				*models.NewBidder("2", "Bob", 10.02, 10.98, 0.01),
			},
			expectedWinnerID:   "1",
			expectedWinningBid: 10.99, // Alice's max bid since she can outbid Bob
			description:        "Small increments should be handled precisely",
		},
		{
			name: "Precision with fractional cents that round",
			bidders: []models.Bidder{
				*models.NewBidder("1", "Alice", 10.01, 20.01, 0.01), // Use valid cent increments
				*models.NewBidder("2", "Bob", 10.00, 19.99, 0.01),
			},
			expectedWinnerID:   "1",   // Alice has higher max bid
			expectedWinningBid: 20.00, // Alice pays Bob's max (19.99) + increment (0.01) = 20.00
			description:        "Fractional cents should round properly",
		},
		{
			name: "Avoid floating point accumulation errors",
			bidders: []models.Bidder{
				*models.NewBidder("1", "Alice", 0.1, 1.0, 0.1),
				*models.NewBidder("2", "Bob", 0.2, 0.9, 0.1),
			},
			expectedWinnerID:   "1",
			expectedWinningBid: 1.0, // Alice should win with her max bid
			description:        "Repeated 0.1 additions should not cause precision errors",
		},
		{
			name: "Large monetary values with small increments",
			bidders: []models.Bidder{
				*models.NewBidder("1", "Alice", 999999.99, 1000000.00, 0.01),
				*models.NewBidder("2", "Bob", 999999.98, 999999.99, 0.01),
			},
			expectedWinnerID:   "1",
			expectedWinningBid: 1000000.00, // Alice's max bid
			description:        "Large values should maintain precision",
		},
		{
			name: "Minimum winning bid calculation precision",
			bidders: []models.Bidder{
				*models.NewBidder("1", "Alice", 10.00, 15.00, 0.25),
				*models.NewBidder("2", "Bob", 10.00, 12.50, 0.25),
			},
			expectedWinnerID:   "1",
			expectedWinningBid: 12.75, // Bob's max (12.50) + Alice's increment (0.25)
			description:        "Minimum winning bid should be calculated precisely",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set entry times to ensure deterministic ordering
			baseTime := time.Now()
			for i := range tt.bidders {
				tt.bidders[i].EntryTime = baseTime.Add(time.Duration(i) * time.Second)
			}

			engine := NewBiddingEngine()
			result, err := engine.ProcessBids(tt.bidders)

			if err != nil {
				t.Fatalf("ProcessBids failed: %v", err)
			}

			if result.Winner == nil {
				t.Fatal("Expected a winner but got nil")
			}

			if result.Winner.ID != tt.expectedWinnerID {
				t.Errorf("Expected winner %s, got %s", tt.expectedWinnerID, result.Winner.ID)
			}

			// Check winning bid with reasonable precision tolerance (1 cent)
			if abs(result.WinningBid-tt.expectedWinningBid) > 0.01 {
				t.Errorf("Expected winning bid %.2f, got %.2f (diff: %.4f)",
					tt.expectedWinningBid, result.WinningBid,
					abs(result.WinningBid-tt.expectedWinningBid))
			}

			t.Logf("Test passed: %s", tt.description)
		})
	}
}

// TestCentsConversion tests the conversion between dollars and cents
func TestCentsConversion(t *testing.T) {
	tests := []struct {
		dollars       float64
		expectedCents int64
		description   string
	}{
		{1.00, 100, "Simple dollar amount"},
		{1.01, 101, "Dollar with cents"},
		{0.99, 99, "Less than a dollar"},
		{10.005, 1001, "Fractional cent rounds up"},
		{10.004, 1000, "Fractional cent rounds down"},
		{999999.99, 99999999, "Large amount"},
		{0.01, 1, "Single cent"},
		{0.001, 0, "Sub-cent rounds to zero"},
		{0.006, 1, "Sub-cent rounds to one"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cents := models.DollarsToCents(tt.dollars)
			if cents != tt.expectedCents {
				t.Errorf("dollarsToCents(%.3f) = %d, expected %d", tt.dollars, cents, tt.expectedCents)
			}

			// Test round-trip conversion (allowing for rounding)
			backToDollars := models.CentsToDollars(cents)
			if abs(backToDollars-tt.dollars) > 0.005 { // Allow for rounding tolerance
				t.Errorf("Round-trip conversion failed: %.3f -> %d -> %.3f", tt.dollars, cents, backToDollars)
			}
		})
	}
}

// TestBidderPrecisionMethods tests the precision methods on Bidder
func TestBidderPrecisionMethods(t *testing.T) {
	bidder := models.NewBidder("1", "Alice", 10.01, 20.99, 0.25)

	// Test that cents values are properly initialized
	if bidder.GetStartingBidCents() != 1001 {
		t.Errorf("Expected starting bid cents 1001, got %d", bidder.GetStartingBidCents())
	}

	if bidder.GetMaxBidCents() != 2099 {
		t.Errorf("Expected max bid cents 2099, got %d", bidder.GetMaxBidCents())
	}

	if bidder.GetAutoIncrementCents() != 25 {
		t.Errorf("Expected auto increment cents 25, got %d", bidder.GetAutoIncrementCents())
	}

	if bidder.GetCurrentBidCents() != 1001 {
		t.Errorf("Expected current bid cents 1001, got %d", bidder.GetCurrentBidCents())
	}

	// Test increment precision
	originalCents := bidder.GetCurrentBidCents()
	bidder.Increment()
	newCents := bidder.GetCurrentBidCents()

	if newCents != originalCents+25 {
		t.Errorf("Expected increment of 25 cents, got %d -> %d", originalCents, newCents)
	}

	// Test that float field is properly synced
	expectedFloat := models.CentsToDollars(newCents)
	if abs(bidder.CurrentBid-expectedFloat) > 0.001 {
		t.Errorf("Float field not properly synced: expected %.3f, got %.3f", expectedFloat, bidder.CurrentBid)
	}
}

// TestMinimumWinningBidPrecision tests precise minimum winning bid calculation
func TestMinimumWinningBidPrecision(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders with values that could cause floating-point precision issues
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.01, 15.33, 0.17),
		*models.NewBidder("2", "Bob", 10.02, 14.44, 0.11),
	}

	// Set entry times
	baseTime := time.Now()
	for i := range bidders {
		bidders[i].EntryTime = baseTime.Add(time.Duration(i) * time.Second)
	}

	result, err := engine.ProcessBids(bidders)
	if err != nil {
		t.Fatalf("ProcessBids failed: %v", err)
	}

	// Test that the winning bid calculation is precise
	// Alice should win, and should pay Bob's max (14.44) + Alice's increment (0.17) = 14.61
	expectedWinningBid := 14.61
	if abs(result.WinningBid-expectedWinningBid) > 0.01 {
		t.Errorf("Expected winning bid %.2f, got %.2f", expectedWinningBid, result.WinningBid)
	}

	// Test the cents-based calculation directly
	winningBidCents, err := engine.CalculateMinimumWinningBidCents(bidders, result.Winner)
	if err != nil {
		t.Fatalf("CalculateMinimumWinningBidCents failed: %v", err)
	}

	expectedCents := int64(1461) // 14.61 in cents
	if winningBidCents != expectedCents {
		t.Errorf("Expected winning bid cents %d, got %d", expectedCents, winningBidCents)
	}
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
