package models

import (
	"testing"
	"time"
)

// TestNewBidResult tests the NewBidResult constructor
func TestNewBidResult(t *testing.T) {
	winner := NewBidder("1", "Alice", 10.00, 20.00, 5.00)
	winningBid := 15.50
	totalBidders := 3
	biddingRounds := 5

	allBidders := []Bidder{
		*winner,
		*NewBidder("2", "Bob", 12.00, 18.00, 3.00),
		*NewBidder("3", "Charlie", 8.00, 16.00, 2.00),
	}

	result := NewBidResult(winner, winningBid, totalBidders, biddingRounds, allBidders)

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	if result.Winner.ID != winner.ID {
		t.Errorf("Expected winner ID '%s', got '%s'", winner.ID, result.Winner.ID)
	}

	if result.WinningBid != winningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", winningBid, result.WinningBid)
	}

	if result.TotalBidders != totalBidders {
		t.Errorf("Expected total bidders %d, got %d", totalBidders, result.TotalBidders)
	}

	if result.BiddingRounds != biddingRounds {
		t.Errorf("Expected bidding rounds %d, got %d", biddingRounds, result.BiddingRounds)
	}

	if len(result.AllBidders) != len(allBidders) {
		t.Errorf("Expected %d bidders, got %d", len(allBidders), len(result.AllBidders))
	}

	// Test that winning bid cents are properly set
	expectedCents := DollarsToCents(winningBid)
	if result.GetWinningBidCents() != expectedCents {
		t.Errorf("Expected winning bid cents %d, got %d", expectedCents, result.GetWinningBidCents())
	}

	// Test that all bidders have synced float fields
	for i, bidder := range result.AllBidders {
		expectedCurrentBid := CentsToDollars(bidder.GetCurrentBidCents())
		if bidder.CurrentBid != expectedCurrentBid {
			t.Errorf("Bidder %d: expected current bid %.2f, got %.2f", i, expectedCurrentBid, bidder.CurrentBid)
		}
	}
}

// TestNewBidResultFromCents tests the NewBidResultFromCents constructor
func TestNewBidResultFromCents(t *testing.T) {
	winner := NewBidder("1", "Alice", 10.00, 20.00, 5.00)
	winningBidCents := int64(1550) // 15.50 in cents
	totalBidders := 2
	biddingRounds := 3

	allBidders := []Bidder{
		*winner,
		*NewBidder("2", "Bob", 12.00, 18.00, 3.00),
	}

	result := NewBidResultFromCents(winner, winningBidCents, totalBidders, biddingRounds, allBidders)

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	if result.Winner.ID != winner.ID {
		t.Errorf("Expected winner ID '%s', got '%s'", winner.ID, result.Winner.ID)
	}

	expectedWinningBid := CentsToDollars(winningBidCents)
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", expectedWinningBid, result.WinningBid)
	}

	if result.TotalBidders != totalBidders {
		t.Errorf("Expected total bidders %d, got %d", totalBidders, result.TotalBidders)
	}

	if result.BiddingRounds != biddingRounds {
		t.Errorf("Expected bidding rounds %d, got %d", biddingRounds, result.BiddingRounds)
	}

	if len(result.AllBidders) != len(allBidders) {
		t.Errorf("Expected %d bidders, got %d", len(allBidders), len(result.AllBidders))
	}

	// Test that winning bid cents are properly set
	if result.GetWinningBidCents() != winningBidCents {
		t.Errorf("Expected winning bid cents %d, got %d", winningBidCents, result.GetWinningBidCents())
	}
}

// TestBidResult_WinnerCheck tests checking for winner presence
func TestBidResult_WinnerCheck(t *testing.T) {
	tests := []struct {
		name     string
		winner   *Bidder
		expected bool
	}{
		{
			name:     "Has winner",
			winner:   NewBidder("1", "Alice", 10.00, 20.00, 5.00),
			expected: true,
		},
		{
			name:     "No winner",
			winner:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewBidResult(tt.winner, 15.00, 1, 0, []Bidder{})

			hasWinner := result.Winner != nil
			if hasWinner != tt.expected {
				t.Errorf("Expected winner != nil = %v, got %v", tt.expected, hasWinner)
			}
		})
	}
}

// TestBidResult_GetWinningBidCents tests the GetWinningBidCents method
func TestBidResult_GetWinningBidCents(t *testing.T) {
	tests := []struct {
		name          string
		winningBid    float64
		expectedCents int64
	}{
		{
			name:          "Whole dollar amount",
			winningBid:    15.00,
			expectedCents: 1500,
		},
		{
			name:          "Fractional amount",
			winningBid:    12.34,
			expectedCents: 1234,
		},
		{
			name:          "Small amount",
			winningBid:    0.99,
			expectedCents: 99,
		},
		{
			name:          "Large amount",
			winningBid:    999.99,
			expectedCents: 99999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := NewBidder("1", "Alice", 10.00, 20.00, 5.00)
			result := NewBidResult(winner, tt.winningBid, 1, 0, []Bidder{*winner})

			cents := result.GetWinningBidCents()
			if cents != tt.expectedCents {
				t.Errorf("Expected winning bid cents %d, got %d", tt.expectedCents, cents)
			}
		})
	}
}

// TestBidResult_NoWinnerScenario tests result with no winner
func TestBidResult_NoWinnerScenario(t *testing.T) {
	// Create a result with no winner (empty auction)
	result := NewBidResult(nil, 0.0, 0, 0, []Bidder{})

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner != nil {
		t.Error("Expected winner to be nil")
	}

	if result.WinningBid != 0.0 {
		t.Errorf("Expected winning bid 0.0, got %.2f", result.WinningBid)
	}

	if result.GetWinningBidCents() != 0 {
		t.Errorf("Expected winning bid cents 0, got %d", result.GetWinningBidCents())
	}

	if result.TotalBidders != 0 {
		t.Errorf("Expected total bidders 0, got %d", result.TotalBidders)
	}

	if len(result.AllBidders) != 0 {
		t.Errorf("Expected 0 bidders, got %d", len(result.AllBidders))
	}
}

// TestBidResult_ComplexScenario tests a complex auction result
func TestBidResult_ComplexScenario(t *testing.T) {
	baseTime := time.Now()

	// Create multiple bidders with different states
	alice := NewBidder("1", "Alice", 100.00, 500.00, 50.00)
	bob := NewBidder("2", "Bob", 110.00, 450.00, 40.00)
	charlie := NewBidder("3", "Charlie", 90.00, 300.00, 30.00)

	// Set entry times
	alice.EntryTime = baseTime
	bob.EntryTime = baseTime.Add(1 * time.Second)
	charlie.EntryTime = baseTime.Add(2 * time.Second)

	// Simulate some bidding activity
	alice.Increment()   // 150.00
	alice.Increment()   // 200.00
	bob.Increment()     // 150.00
	charlie.Increment() // 120.00

	allBidders := []Bidder{*alice, *bob, *charlie}
	winningBid := 470.00 // Alice wins, pays Bob's max + her increment

	result := NewBidResult(alice, winningBid, 3, 8, allBidders)

	// Verify result structure
	if result.Winner == nil {
		t.Fatal("Expected to have winner")
	}

	if result.Winner.ID != "1" {
		t.Errorf("Expected winner '1', got '%s'", result.Winner.ID)
	}

	if result.WinningBid != winningBid {
		t.Errorf("Expected winning bid %.2f, got %.2f", winningBid, result.WinningBid)
	}

	if result.TotalBidders != 3 {
		t.Errorf("Expected total bidders 3, got %d", result.TotalBidders)
	}

	if result.BiddingRounds != 8 {
		t.Errorf("Expected bidding rounds 8, got %d", result.BiddingRounds)
	}

	if len(result.AllBidders) != 3 {
		t.Errorf("Expected 3 bidders in result, got %d", len(result.AllBidders))
	}

	// Verify that all bidders have proper float field sync
	for i, bidder := range result.AllBidders {
		expectedCurrent := CentsToDollars(bidder.GetCurrentBidCents())
		if bidder.CurrentBid != expectedCurrent {
			t.Errorf("Bidder %d: current bid not synced, expected %.2f, got %.2f",
				i, expectedCurrent, bidder.CurrentBid)
		}

		expectedStarting := CentsToDollars(bidder.GetStartingBidCents())
		if bidder.StartingBid != expectedStarting {
			t.Errorf("Bidder %d: starting bid not synced, expected %.2f, got %.2f",
				i, expectedStarting, bidder.StartingBid)
		}

		expectedMax := CentsToDollars(bidder.GetMaxBidCents())
		if bidder.MaxBid != expectedMax {
			t.Errorf("Bidder %d: max bid not synced, expected %.2f, got %.2f",
				i, expectedMax, bidder.MaxBid)
		}

		expectedIncrement := CentsToDollars(bidder.GetAutoIncrementCents())
		if bidder.AutoIncrement != expectedIncrement {
			t.Errorf("Bidder %d: auto increment not synced, expected %.2f, got %.2f",
				i, expectedIncrement, bidder.AutoIncrement)
		}
	}

	// Test cents precision
	expectedCents := DollarsToCents(winningBid)
	if result.GetWinningBidCents() != expectedCents {
		t.Errorf("Expected winning bid cents %d, got %d", expectedCents, result.GetWinningBidCents())
	}
}

// TestBidResult_PrecisionConsistency tests precision consistency between constructors
func TestBidResult_PrecisionConsistency(t *testing.T) {
	winner := NewBidder("1", "Alice", 10.01, 20.99, 0.33)
	winningBid := 15.67
	allBidders := []Bidder{*winner}

	// Create result using float constructor
	result1 := NewBidResult(winner, winningBid, 1, 0, allBidders)

	// Create result using cents constructor
	winningBidCents := DollarsToCents(winningBid)
	result2 := NewBidResultFromCents(winner, winningBidCents, 1, 0, allBidders)

	// Both should have the same cents value
	if result1.GetWinningBidCents() != result2.GetWinningBidCents() {
		t.Errorf("Cents mismatch: result1=%d, result2=%d",
			result1.GetWinningBidCents(), result2.GetWinningBidCents())
	}

	// Both should have the same dollar value (within precision tolerance)
	if result1.WinningBid != result2.WinningBid {
		t.Errorf("Dollar mismatch: result1=%.2f, result2=%.2f",
			result1.WinningBid, result2.WinningBid)
	}

	// Both should have same winner status
	hasWinner1 := result1.Winner != nil
	hasWinner2 := result2.Winner != nil
	if hasWinner1 != hasWinner2 {
		t.Error("Winner status mismatch between constructors")
	}
}
