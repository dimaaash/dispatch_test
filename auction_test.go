package auction

import (
	"fmt"
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

func TestAuctionService_DetermineWinner_SingleBidder(t *testing.T) {
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

	if result.Winner.ID != "bidder1" {
		t.Errorf("Expected winner ID 'bidder1', got '%s'", result.Winner.ID)
	}

	// Single bidder should pay their starting bid
	if result.WinningBid != 100.0 {
		t.Errorf("Expected winning bid 100.0, got %f", result.WinningBid)
	}

	if result.TotalBidders != 1 {
		t.Errorf("Expected total bidders 1, got %d", result.TotalBidders)
	}
}

func TestAuctionService_DetermineWinner_MultipleBidders_NoIncrements(t *testing.T) {
	service := NewAuctionService()

	baseTime := time.Now()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        150.0,
			AutoIncrement: 10.0,
			EntryTime:     baseTime,
		},
		{
			ID:            "bidder2",
			Name:          "Bob",
			StartingBid:   80.0,
			MaxBid:        120.0,
			AutoIncrement: 5.0,
			EntryTime:     baseTime.Add(1 * time.Second),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Winner.ID != "bidder1" {
		t.Errorf("Expected winner 'bidder1', got '%s'", result.Winner.ID)
	}

	// Winner should pay just enough to beat the second bidder
	expectedWinningBid := 120.0 + 10.0 // Bob's max + Alice's increment
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %f, got %f", expectedWinningBid, result.WinningBid)
	}
}

func TestAuctionService_DetermineWinner_ComplexIncrementingScenario(t *testing.T) {
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
		{
			ID:            "bidder3",
			Name:          "Charlie",
			StartingBid:   90.0,
			MaxBid:        200.0,
			AutoIncrement: 15.0,
			EntryTime:     baseTime.Add(2 * time.Second),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Winner.ID != "bidder1" {
		t.Errorf("Expected winner 'bidder1', got '%s'", result.Winner.ID)
	}

	// Alice should win and pay just enough to beat Bob
	expectedWinningBid := 250.0 + 25.0 // Bob's max + Alice's increment
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid %f, got %f", expectedWinningBid, result.WinningBid)
	}

	if result.BiddingRounds == 0 {
		t.Error("Expected multiple bidding rounds")
	}
}

func TestAuctionService_DetermineWinner_TieResolution(t *testing.T) {
	service := NewAuctionService()

	baseTime := time.Now()
	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        150.0,
			AutoIncrement: 10.0,
			EntryTime:     baseTime.Add(1 * time.Second), // Later entry
		},
		{
			ID:            "bidder2",
			Name:          "Bob",
			StartingBid:   100.0,
			MaxBid:        150.0,
			AutoIncrement: 10.0,
			EntryTime:     baseTime, // Earlier entry
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Bob should win due to earlier entry time
	if result.Winner.ID != "bidder2" {
		t.Errorf("Expected winner 'bidder2' (earlier entry), got '%s'", result.Winner.ID)
	}
}

func TestAuctionService_DetermineWinner_ValidationErrors(t *testing.T) {
	service := NewAuctionService()

	// Test with invalid bidders
	bidders := []models.Bidder{
		{
			ID:            "", // Invalid: empty ID
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        200.0,
			AutoIncrement: 10.0,
			EntryTime:     time.Now(),
		},
		{
			ID:            "bidder2",
			Name:          "Bob",
			StartingBid:   200.0,
			MaxBid:        100.0, // Invalid: starting bid > max bid
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

	// Check that it's an AuctionError with validation details
	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError, got %T", err)
	}

	if auctionErr.Type != "validation" {
		t.Errorf("Expected validation error type, got '%s'", auctionErr.Type)
	}

	if len(auctionErr.Details) == 0 {
		t.Error("Expected validation error details")
	}
}

func TestAuctionService_DetermineWinner_EmptyBidders(t *testing.T) {
	service := NewAuctionService()

	result, err := service.DetermineWinner([]models.Bidder{})

	if err == nil {
		t.Fatal("Expected error for empty bidders, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for empty bidders")
	}

	auctionErr, ok := err.(*models.AuctionError)
	if !ok {
		t.Fatalf("Expected AuctionError, got %T", err)
	}

	if auctionErr.Type != "validation" {
		t.Errorf("Expected validation error type, got '%s'", auctionErr.Type)
	}
}

func TestAuctionService_DetermineWinner_NegativeBids(t *testing.T) {
	service := NewAuctionService()

	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   -100.0, // Invalid: negative starting bid
			MaxBid:        200.0,
			AutoIncrement: 10.0,
			EntryTime:     time.Now(),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected validation error for negative bid, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when validation fails")
	}
}

func TestAuctionService_DetermineWinner_ZeroAutoIncrement(t *testing.T) {
	service := NewAuctionService()

	bidders := []models.Bidder{
		{
			ID:            "bidder1",
			Name:          "Alice",
			StartingBid:   100.0,
			MaxBid:        200.0,
			AutoIncrement: 0.0, // Invalid: zero auto-increment
			EntryTime:     time.Now(),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected validation error for zero auto-increment, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when validation fails")
	}
}

func TestAuctionService_DetermineWinner_DuplicateBidderIDs(t *testing.T) {
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
		{
			ID:            "bidder1", // Duplicate ID
			Name:          "Bob",
			StartingBid:   110.0,
			MaxBid:        220.0,
			AutoIncrement: 15.0,
			EntryTime:     time.Now(),
		},
	}

	result, err := service.DetermineWinner(bidders)

	if err == nil {
		t.Fatal("Expected validation error for duplicate IDs, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when validation fails")
	}
}

func TestAuctionService_DetermineWinner_LargeBidderSet(t *testing.T) {
	service := NewAuctionService()

	// Create a large set of bidders to test performance
	bidders := make([]models.Bidder, 100)
	baseTime := time.Now()

	for i := 0; i < 100; i++ {
		bidders[i] = models.Bidder{
			ID:            fmt.Sprintf("bidder%d", i+1),
			Name:          fmt.Sprintf("Bidder %d", i+1),
			StartingBid:   float64(100 + i),
			MaxBid:        float64(200 + i*2),
			AutoIncrement: float64(5 + i%10),
			EntryTime:     baseTime.Add(time.Duration(i) * time.Millisecond),
		}
	}

	result, err := service.DetermineWinner(bidders)

	if err != nil {
		t.Fatalf("Expected no error with large bidder set, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	if result.TotalBidders != 100 {
		t.Errorf("Expected total bidders 100, got %d", result.TotalBidders)
	}

	// The last bidder should have the highest max bid and should win
	if result.Winner.ID != "bidder100" {
		t.Errorf("Expected winner 'bidder100', got '%s'", result.Winner.ID)
	}
}
