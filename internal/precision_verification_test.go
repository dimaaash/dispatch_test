package internal

import (
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

// TestPrecisionHandlingImplementation verifies that all monetary calculations use precise arithmetic
func TestPrecisionHandlingImplementation(t *testing.T) {
	// Test case that would fail with floating-point precision issues
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.01, 10.99, 0.01),
		*models.NewBidder("2", "Bob", 10.02, 10.98, 0.01),
	}

	// Set entry times to ensure deterministic ordering
	baseTime := time.Now()
	for i := range bidders {
		bidders[i].EntryTime = baseTime.Add(time.Duration(i) * time.Second)
	}

	engine := NewBiddingEngine()
	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Fatalf("ProcessBids failed: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner but got nil")
	}

	// Verify that Alice wins (she has higher max bid)
	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice (ID: 1) to win, got %s", result.Winner.ID)
	}

	// Verify that the winning bid is calculated precisely
	// Alice should pay Bob's max (10.98) + Alice's increment (0.01) = 10.99
	expectedWinningBid := 10.99
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", expectedWinningBid, result.WinningBid)
	}

	// Verify that cents-based calculations are used internally
	winningBidCents := result.GetWinningBidCents()
	expectedCents := int64(1099) // 10.99 in cents
	if winningBidCents != expectedCents {
		t.Errorf("Expected winning bid cents %d, got %d", expectedCents, winningBidCents)
	}

	t.Logf("Precision test passed: Winner=%s, WinningBid=%.2f, WinningBidCents=%d",
		result.Winner.ID, result.WinningBid, winningBidCents)
}

// TestDecimalArithmeticAccuracy tests that decimal arithmetic is accurate for monetary calculations
func TestDecimalArithmeticAccuracy(t *testing.T) {
	tests := []struct {
		name          string
		dollars       float64
		expectedCents int64
		description   string
	}{
		{"Simple conversion", 1.00, 100, "Basic dollar to cents"},
		{"With cents", 1.23, 123, "Dollar with cents"},
		{"Small increment", 0.01, 1, "Single cent"},
		{"Fractional rounding up", 10.005, 1001, "Rounds 0.5 cents up"},
		{"Fractional rounding down", 10.004, 1000, "Rounds 0.4 cents down"},
		{"Large amount", 999999.99, 99999999, "Large monetary value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cents := models.DollarsToCents(tt.dollars)
			if cents != tt.expectedCents {
				t.Errorf("DollarsToCents(%.3f) = %d, expected %d", tt.dollars, cents, tt.expectedCents)
			}

			// Test round-trip conversion
			backToDollars := models.CentsToDollars(cents)
			// Allow small tolerance for rounding
			diff := backToDollars - tt.dollars
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.005 { // 0.5 cent tolerance
				t.Errorf("Round-trip failed: %.3f -> %d -> %.3f (diff: %.6f)",
					tt.dollars, cents, backToDollars, diff)
			}

			t.Logf("%s: %.3f -> %d cents -> %.3f", tt.description, tt.dollars, cents, backToDollars)
		})
	}
}

// TestBidderPrecisionOperations tests that bidder operations use precise arithmetic
func TestBidderPrecisionOperations(t *testing.T) {
	bidder := models.NewBidder("1", "Alice", 10.01, 20.99, 0.25)

	// Verify initial state
	if bidder.GetStartingBidCents() != 1001 {
		t.Errorf("Expected starting bid cents 1001, got %d", bidder.GetStartingBidCents())
	}

	if bidder.GetMaxBidCents() != 2099 {
		t.Errorf("Expected max bid cents 2099, got %d", bidder.GetMaxBidCents())
	}

	if bidder.GetAutoIncrementCents() != 25 {
		t.Errorf("Expected auto increment cents 25, got %d", bidder.GetAutoIncrementCents())
	}

	// Test increment operation
	originalCents := bidder.GetCurrentBidCents()
	success := bidder.Increment()

	if !success {
		t.Fatal("Expected increment to succeed")
	}

	newCents := bidder.GetCurrentBidCents()
	expectedNewCents := originalCents + 25

	if newCents != expectedNewCents {
		t.Errorf("Expected current bid cents %d after increment, got %d", expectedNewCents, newCents)
	}

	// Verify that float field is properly synced
	expectedFloat := models.CentsToDollars(newCents)
	if bidder.CurrentBid != expectedFloat {
		t.Errorf("Float field not synced: expected %.2f, got %.2f", expectedFloat, bidder.CurrentBid)
	}

	t.Logf("Bidder precision test passed: %d cents -> %d cents (increment: %d)",
		originalCents, newCents, bidder.GetAutoIncrementCents())
}

// TestPrecisionInComplexScenario tests precision in a complex bidding scenario
func TestPrecisionInComplexScenario(t *testing.T) {
	// Create a scenario that would be problematic with floating-point arithmetic
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 0.10, 1.00, 0.10), // Repeated 0.1 additions
		*models.NewBidder("2", "Bob", 0.20, 0.90, 0.10),
		*models.NewBidder("3", "Charlie", 0.15, 0.85, 0.05),
	}

	// Set entry times
	baseTime := time.Now()
	for i := range bidders {
		bidders[i].EntryTime = baseTime.Add(time.Duration(i) * time.Second)
	}

	engine := NewBiddingEngine()
	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Fatalf("ProcessBids failed: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner but got nil")
	}

	// Alice should win because she has the highest max bid (1.00)
	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice (ID: 1) to win, got %s", result.Winner.ID)
	}

	// Verify that the winning bid calculation is precise
	// Alice should pay Bob's max (0.90) + Alice's increment (0.10) = 1.00
	expectedWinningBid := 1.00
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", expectedWinningBid, result.WinningBid)
	}

	// Verify cents calculation
	expectedCents := int64(100) // 1.00 in cents
	if result.GetWinningBidCents() != expectedCents {
		t.Errorf("Expected winning bid cents %d, got %d", expectedCents, result.GetWinningBidCents())
	}

	t.Logf("Complex precision test passed: Winner=%s, WinningBid=%.2f, Rounds=%d",
		result.Winner.ID, result.WinningBid, result.BiddingRounds)
}
