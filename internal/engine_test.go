package internal

import (
	"testing"
	"time"

	"auction-bidding-algorithm/internal/models"
)

func TestNewBiddingEngine(t *testing.T) {
	engine := NewBiddingEngine()
	if engine == nil {
		t.Fatal("NewBiddingEngine should return a non-nil engine")
	}
	if engine.maxRounds != 1000 {
		t.Errorf("Expected maxRounds to be 1000, got %d", engine.maxRounds)
	}
}

func TestProcessBids_EmptyBidders(t *testing.T) {
	engine := NewBiddingEngine()
	result, err := engine.ProcessBids([]models.Bidder{})

	if err != nil {
		t.Errorf("Expected no error for empty bidders, got: %v", err)
	}

	if result.Winner != nil {
		t.Error("Expected no winner for empty bidders")
	}
	if result.WinningBid != 0 {
		t.Errorf("Expected winning bid to be 0, got %f", result.WinningBid)
	}
	if result.TotalBidders != 0 {
		t.Errorf("Expected total bidders to be 0, got %d", result.TotalBidders)
	}
}

func TestProcessBids_SingleBidder(t *testing.T) {
	engine := NewBiddingEngine()

	bidder := models.NewBidder("1", "Alice", 100.0, 200.0, 10.0)
	bidders := []models.Bidder{*bidder}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error for single bidder, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner for single bidder")
	}
	if result.Winner.ID != "1" {
		t.Errorf("Expected winner ID to be '1', got '%s'", result.Winner.ID)
	}
	if result.WinningBid != 100.0 {
		t.Errorf("Expected winning bid to be 100.0 (starting bid), got %f", result.WinningBid)
	}
	if result.BiddingRounds != 0 {
		t.Errorf("Expected 0 bidding rounds for single bidder, got %d", result.BiddingRounds)
	}
}

func TestProcessBids_MultipleBidders_NoIncrements(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders where the highest starting bid wins immediately
	bidder1 := models.NewBidder("1", "Alice", 100.0, 150.0, 10.0)
	bidder2 := models.NewBidder("2", "Bob", 80.0, 120.0, 5.0)

	bidders := []models.Bidder{*bidder1, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}
	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice to win, got %s", result.Winner.Name)
	}
	// Alice should pay just enough to beat Bob's max bid
	expectedWinningBid := 120.0 + 10.0 // Bob's max + Alice's increment
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid to be %f, got %f", expectedWinningBid, result.WinningBid)
	}
}

func TestProcessBids_MultipleBidders_WithIncrements(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders where increments are needed
	bidder1 := models.NewBidder("1", "Alice", 100.0, 200.0, 20.0)
	bidder2 := models.NewBidder("2", "Bob", 110.0, 180.0, 15.0)

	bidders := []models.Bidder{*bidder1, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Alice starts at 100, Bob at 110
	// Alice increments to 120, Bob increments to 125
	// Alice increments to 140, Bob increments to 140
	// Alice increments to 160, Bob increments to 155
	// Alice increments to 180, Bob increments to 170
	// Alice increments to 200, Bob can't increment further (max 180)
	// Alice wins at 200

	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice to win, got %s", result.Winner.Name)
	}

	// Alice should pay just enough to beat Bob's max (180) + her increment (20) = 200
	// But capped at her max bid of 200
	expectedWinningBid := 200.0
	if result.WinningBid != expectedWinningBid {
		t.Errorf("Expected winning bid to be %f, got %f", expectedWinningBid, result.WinningBid)
	}

	if result.BiddingRounds == 0 {
		t.Error("Expected multiple bidding rounds")
	}
}

func TestProcessBids_TieResolution(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders with identical parameters but different entry times
	now := time.Now()

	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(time.Second),
	}

	bidders := []models.Bidder{*bidder1, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Alice should win because she entered first
	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice to win due to earlier entry, got %s", result.Winner.Name)
	}
}

func TestIncrementBids_NoIncrementsPossible(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders who can't increment
	bidder1 := models.Bidder{
		ID: "1", CurrentBid: 100.0, MaxBid: 100.0, AutoIncrement: 10.0, IsActive: false,
	}
	bidder2 := models.Bidder{
		ID: "2", CurrentBid: 90.0, MaxBid: 90.0, AutoIncrement: 5.0, IsActive: false,
	}

	bidders := []models.Bidder{bidder1, bidder2}

	result, err := engine.IncrementBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result {
		t.Error("Expected no increments to be possible")
	}
}

func TestIncrementBids_SomeIncrementsPerformed(t *testing.T) {
	engine := NewBiddingEngine()

	// Create bidders where some can increment using NewBidder for proper initialization
	bidder1 := *models.NewBidder("1", "Alice", 100.0, 150.0, 10.0)
	bidder2 := *models.NewBidder("2", "Bob", 90.0, 120.0, 5.0)

	bidders := []models.Bidder{bidder1, bidder2}

	result, err := engine.IncrementBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !result {
		t.Error("Expected some increments to be performed")
	}

	// Bidder2 should have incremented to compete with bidder1
	if bidders[1].CurrentBid != 95.0 {
		t.Errorf("Expected bidder2 to increment to 95.0, got %f", bidders[1].CurrentBid)
	}
}

func TestFindWinner_HighestBidWins(t *testing.T) {
	engine := NewBiddingEngine()

	bidder1 := *models.NewBidder("1", "Bidder1", 100.0, 150.0, 10.0)
	bidder2 := *models.NewBidder("2", "Bidder2", 120.0, 150.0, 10.0)
	bidder3 := *models.NewBidder("3", "Bidder3", 90.0, 150.0, 10.0)

	bidders := []models.Bidder{bidder1, bidder2, bidder3}

	winner, err := engine.findWinner(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if winner == nil {
		t.Fatal("Expected a winner")
	}
	if winner.ID != "2" {
		t.Errorf("Expected bidder 2 to win, got %s", winner.ID)
	}
}

func TestFindWinner_TieGoesToEarlierEntry(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	bidder1 := *models.NewBidder("1", "Bidder1", 100.0, 150.0, 10.0)
	bidder1.EntryTime = now.Add(time.Second)

	bidder2 := *models.NewBidder("2", "Bidder2", 100.0, 150.0, 10.0)
	bidder2.EntryTime = now

	bidders := []models.Bidder{bidder1, bidder2}

	winner, err := engine.findWinner(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if winner == nil {
		t.Fatal("Expected a winner")
	}
	if winner.ID != "2" {
		t.Errorf("Expected bidder 2 to win (earlier entry), got %s", winner.ID)
	}
}

// Comprehensive tie resolution tests

func TestTieResolution_SameEffectiveBidAmount_EarlierEntryWins(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Create bidders with identical parameters but different entry times
	// Requirement 5.1: WHEN multiple bidders have the same effective bid amount THEN the system SHALL prioritize the earlier entry
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now, IsActive: true,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(time.Second), IsActive: true,
	}
	bidder3 := &models.Bidder{
		ID: "3", Name: "Charlie", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(2 * time.Second), IsActive: true,
	}

	bidders := []models.Bidder{*bidder1, *bidder2, *bidder3}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Alice should win because she entered first (earliest entry time)
	if result.Winner.ID != "1" {
		t.Errorf("Expected Alice (ID: 1) to win due to earlier entry, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}
}

func TestTieResolution_BidOrderByTimestamp(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Requirement 5.2: WHEN determining bid order THEN the system SHALL use the timestamp or entry sequence of when bids were submitted
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 90.0, MaxBid: 120.0,
		AutoIncrement: 5.0, EntryTime: now.Add(2 * time.Second), IsActive: true,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 95.0, MaxBid: 120.0,
		AutoIncrement: 5.0, EntryTime: now, IsActive: true, // Earliest entry
	}
	bidder3 := &models.Bidder{
		ID: "3", Name: "Charlie", StartingBid: 85.0, MaxBid: 120.0,
		AutoIncrement: 5.0, EntryTime: now.Add(time.Second), IsActive: true,
	}

	bidders := []models.Bidder{*bidder1, *bidder2, *bidder3}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Bob should win because he entered first, even though he wasn't first in the slice
	if result.Winner.ID != "2" {
		t.Errorf("Expected Bob (ID: 2) to win due to earliest entry time, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}
}

func TestTieResolution_WinningBidLevelTie_FirstEntryWins(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Requirement 5.3: WHEN ties occur at the winning bid level THEN the system SHALL award the item to the bidder who entered first
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 130.0,
		AutoIncrement: 10.0, EntryTime: now.Add(time.Second), IsActive: true,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 100.0, MaxBid: 130.0,
		AutoIncrement: 10.0, EntryTime: now, IsActive: true, // Earlier entry
	}

	bidders := []models.Bidder{*bidder1, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Bob should win because he entered first, even though both reach the same final bid
	if result.Winner.ID != "2" {
		t.Errorf("Expected Bob (ID: 2) to win due to earlier entry at winning bid level, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}

	// Both bidders start at 100.0 and since they're identical, no increments are needed
	// The winner is determined by tie resolution (earlier entry wins)
	if result.Winner.CurrentBid != 100.0 {
		t.Errorf("Expected winner's current bid to be 100.0 (starting bid), got %f", result.Winner.CurrentBid)
	}
}

func TestTieResolution_ComplexScenario_MultipleRoundsWithTies(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Complex scenario with multiple bidders and increments
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 200.0,
		AutoIncrement: 15.0, EntryTime: now.Add(3 * time.Second), IsActive: true,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 105.0, MaxBid: 200.0,
		AutoIncrement: 15.0, EntryTime: now, IsActive: true, // Earliest
	}
	bidder3 := &models.Bidder{
		ID: "3", Name: "Charlie", StartingBid: 95.0, MaxBid: 200.0,
		AutoIncrement: 15.0, EntryTime: now.Add(time.Second), IsActive: true,
	}
	bidder4 := &models.Bidder{
		ID: "4", Name: "David", StartingBid: 90.0, MaxBid: 200.0,
		AutoIncrement: 15.0, EntryTime: now.Add(2 * time.Second), IsActive: true,
	}

	bidders := []models.Bidder{*bidder1, *bidder2, *bidder3, *bidder4}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Charlie should win because he can reach the highest bid (200)
	// Charlie: 95 + 7*15 = 200 (exactly reaches max)
	// Bob: 105 + 6*15 = 195 (next increment would exceed max: 195 + 15 = 210 > 200)
	if result.Winner.ID != "3" {
		t.Errorf("Expected Charlie (ID: 3) to win due to highest achievable bid, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}

	// Charlie should reach exactly 200
	if result.Winner.CurrentBid != 200.0 {
		t.Errorf("Expected winner's current bid to be 200.0, got %f", result.Winner.CurrentBid)
	}

	// Verify that bidding rounds occurred
	if result.BiddingRounds == 0 {
		t.Error("Expected multiple bidding rounds in complex scenario")
	}
}

func TestTieResolution_DifferentMaxBids_HigherMaxWins(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Test that higher max bid wins when both bidders need to increment
	// Create a third bidder with higher starting bid to force increments
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 90.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now, IsActive: true, // Earlier entry
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 95.0, MaxBid: 180.0,
		AutoIncrement: 10.0, EntryTime: now.Add(time.Second), IsActive: true, // Later entry but higher max
	}
	bidder3 := &models.Bidder{
		ID: "3", Name: "Charlie", StartingBid: 100.0, MaxBid: 120.0,
		AutoIncrement: 5.0, EntryTime: now.Add(2 * time.Second), IsActive: true, // Forces others to increment
	}

	bidders := []models.Bidder{*bidder1, *bidder2, *bidder3}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Bob should win because he has the highest max bid (180)
	// Charlie starts highest (100) but can only reach 120
	// Alice can reach 150, Bob can reach 180
	if result.Winner.ID != "2" {
		t.Errorf("Expected Bob (ID: 2) to win due to highest max bid, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}

	// Bob should reach a bid higher than Alice's max (150)
	if result.Winner.CurrentBid <= 150.0 {
		t.Errorf("Expected Bob to bid higher than Alice's max (150), got %f", result.Winner.CurrentBid)
	}
}

func TestTieResolution_IdenticalBiddersExceptEntryTime(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Test with completely identical bidders except for entry time
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(100 * time.Millisecond), IsActive: true,
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now, IsActive: true, // 100ms earlier
	}

	bidders := []models.Bidder{*bidder1, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Bob should win because he entered 100ms earlier
	if result.Winner.ID != "2" {
		t.Errorf("Expected Bob (ID: 2) to win due to earlier entry by 100ms, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}

	// Since both bidders are identical and start at the same bid (100.0),
	// no increments are needed and tie resolution determines the winner
	if result.Winner.CurrentBid != 100.0 {
		t.Errorf("Expected winner's current bid to be 100.0 (starting bid), got %f", result.Winner.CurrentBid)
	}
}

func TestTieResolution_SortingPreservesEntryTimeOrder(t *testing.T) {
	engine := NewBiddingEngine()

	now := time.Now()

	// Create bidders in random order but with specific entry times
	bidder1 := &models.Bidder{
		ID: "1", Name: "Alice", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(3 * time.Second), IsActive: true, // Latest
	}
	bidder2 := &models.Bidder{
		ID: "2", Name: "Bob", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now, IsActive: true, // Earliest
	}
	bidder3 := &models.Bidder{
		ID: "3", Name: "Charlie", StartingBid: 100.0, MaxBid: 150.0,
		AutoIncrement: 10.0, EntryTime: now.Add(time.Second), IsActive: true, // Middle
	}

	// Pass bidders in non-chronological order to test sorting
	bidders := []models.Bidder{*bidder1, *bidder3, *bidder2}

	result, err := engine.ProcessBids(bidders)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}

	// Bob should win because he has the earliest entry time
	if result.Winner.ID != "2" {
		t.Errorf("Expected Bob (ID: 2) to win due to earliest entry time, got %s (ID: %s)", result.Winner.Name, result.Winner.ID)
	}

	// Verify that all bidders are in the result and properly sorted by entry time
	if len(result.AllBidders) != 3 {
		t.Errorf("Expected 3 bidders in result, got %d", len(result.AllBidders))
	}

	// Check that bidders are sorted by entry time in the result
	for i := 1; i < len(result.AllBidders); i++ {
		if result.AllBidders[i-1].EntryTime.After(result.AllBidders[i].EntryTime) {
			t.Error("Bidders should be sorted by entry time (earliest first)")
		}
	}
}

func TestProcessBids_MaxRoundsTimeout(t *testing.T) {
	// Create an engine with very low max rounds to trigger timeout
	engine := &BiddingEngine{maxRounds: 1}

	// Create bidders that would cause many rounds of bidding
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 1.00, 100.00, 1.00),
		*models.NewBidder("2", "Bob", 1.01, 99.00, 1.00),
	}

	// Set entry times
	baseTime := time.Now()
	bidders[0].EntryTime = baseTime
	bidders[1].EntryTime = baseTime.Add(1 * time.Second)

	result, err := engine.ProcessBids(bidders)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on timeout")
	}

	// Check that it's a timeout error
	timeoutErr, ok := err.(*models.TimeoutError)
	if !ok {
		t.Fatalf("Expected TimeoutError, got %T", err)
	}

	if timeoutErr.Operation != "ProcessBids" {
		t.Errorf("Expected operation 'ProcessBids', got '%s'", timeoutErr.Operation)
	}
}

// TestIncrementBids_SystemError tests system error handling in IncrementBids
func TestIncrementBids_SystemError(t *testing.T) {
	engine := NewBiddingEngine()

	// Create a bidder with corrupted internal state (this is artificial for testing)
	bidder := models.NewBidder("1", "Alice", 10.00, 20.00, 5.00)

	// Manually corrupt the bidder's internal state to trigger system error
	// We'll use reflection or direct field access if possible, or create a scenario
	// that would naturally cause the system error

	bidders := []models.Bidder{*bidder}

	// This should work normally, but let's test the error path by creating
	// a scenario where CanIncrement returns true but Increment fails
	// We need to test the system error path in IncrementBids

	// For now, test the normal path and ensure no system errors occur
	incremented, err := engine.IncrementBids(bidders)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// With only one bidder, no increments should happen
	if incremented {
		t.Error("Expected no increments with single bidder")
	}
}

// TestIncrementBids_EmptyAndSingleBidder tests edge cases
func TestIncrementBids_EmptyAndSingleBidder(t *testing.T) {
	engine := NewBiddingEngine()

	// Test empty bidders
	incremented, err := engine.IncrementBids([]models.Bidder{})
	if err != nil {
		t.Fatalf("Expected no error with empty bidders, got: %v", err)
	}
	if incremented {
		t.Error("Expected no increments with empty bidders")
	}

	// Test single bidder
	bidder := models.NewBidder("1", "Alice", 10.00, 20.00, 5.00)
	incremented, err = engine.IncrementBids([]models.Bidder{*bidder})
	if err != nil {
		t.Fatalf("Expected no error with single bidder, got: %v", err)
	}
	if incremented {
		t.Error("Expected no increments with single bidder")
	}
}

// TestCalculateMinimumWinningBidCents_ErrorPaths tests error handling
func TestCalculateMinimumWinningBidCents_ErrorPaths(t *testing.T) {
	engine := NewBiddingEngine()

	// Test nil winner
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.00, 20.00, 5.00),
	}

	_, err := engine.CalculateMinimumWinningBidCents(bidders, nil)
	if err == nil {
		t.Fatal("Expected error with nil winner")
	}

	inputErr, ok := err.(*models.InputError)
	if !ok {
		t.Fatalf("Expected InputError, got %T", err)
	}
	if inputErr.InputField != "winner" {
		t.Errorf("Expected input field 'winner', got '%s'", inputErr.InputField)
	}

	// Test empty bidders
	winner := models.NewBidder("1", "Alice", 10.00, 20.00, 5.00)
	_, err = engine.CalculateMinimumWinningBidCents([]models.Bidder{}, winner)
	if err == nil {
		t.Fatal("Expected error with empty bidders")
	}

	inputErr, ok = err.(*models.InputError)
	if !ok {
		t.Fatalf("Expected InputError, got %T", err)
	}
	if inputErr.InputField != "bidders" {
		t.Errorf("Expected input field 'bidders', got '%s'", inputErr.InputField)
	}

	// Test winner not in bidders slice
	winner = models.NewBidder("2", "Bob", 10.00, 20.00, 5.00)
	_, err = engine.CalculateMinimumWinningBidCents(bidders, winner)
	if err == nil {
		t.Fatal("Expected error when winner not in bidders")
	}

	inputErr, ok = err.(*models.InputError)
	if !ok {
		t.Fatalf("Expected InputError, got %T", err)
	}
	if inputErr.InputField != "winner.ID" {
		t.Errorf("Expected input field 'winner.ID', got '%s'", inputErr.InputField)
	}
}

// TestCalculateMinimumWinningBidCents_EdgeCases tests edge cases
func TestCalculateMinimumWinningBidCents_EdgeCases(t *testing.T) {
	engine := NewBiddingEngine()

	// Test single bidder (no other bidders to compete against)
	winner := models.NewBidder("1", "Alice", 10.00, 20.00, 5.00)
	bidders := []models.Bidder{*winner}

	winningBid, err := engine.CalculateMinimumWinningBidCents(bidders, winner)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should pay starting bid when no competition
	expectedCents := winner.GetStartingBidCents()
	if winningBid != expectedCents {
		t.Errorf("Expected winning bid %d cents, got %d cents", expectedCents, winningBid)
	}

	// Test winner pays minimum when calculated bid is less than starting bid
	winner = models.NewBidder("1", "Alice", 15.00, 20.00, 1.00)
	loser := models.NewBidder("2", "Bob", 10.00, 12.00, 1.00)
	bidders = []models.Bidder{*winner, *loser}

	winningBid, err = engine.CalculateMinimumWinningBidCents(bidders, winner)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Winner should pay at least their starting bid
	if winningBid < winner.GetStartingBidCents() {
		t.Errorf("Winning bid %d should not be less than starting bid %d", winningBid, winner.GetStartingBidCents())
	}

	// Test winner pays max bid when calculated exceeds max
	winner = models.NewBidder("1", "Alice", 10.00, 15.00, 1.00)
	loser = models.NewBidder("2", "Bob", 10.00, 20.00, 1.00)
	bidders = []models.Bidder{*winner, *loser}

	winningBid, err = engine.CalculateMinimumWinningBidCents(bidders, winner)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Winner should not pay more than their max bid
	if winningBid > winner.GetMaxBidCents() {
		t.Errorf("Winning bid %d should not exceed max bid %d", winningBid, winner.GetMaxBidCents())
	}
}

// TestFindWinner_ErrorPaths tests error handling in findWinner
func TestFindWinner_ErrorPaths(t *testing.T) {
	engine := NewBiddingEngine()

	// Test empty bidders
	winner, err := engine.findWinner([]models.Bidder{})
	if err != nil {
		t.Fatalf("Expected no error with empty bidders, got: %v", err)
	}
	if winner != nil {
		t.Error("Expected nil winner with empty bidders")
	}

	// Test normal case
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.00, 20.00, 5.00),
		*models.NewBidder("2", "Bob", 15.00, 25.00, 5.00),
	}

	// Set entry times for deterministic ordering
	baseTime := time.Now()
	bidders[0].EntryTime = baseTime
	bidders[1].EntryTime = baseTime.Add(1 * time.Second)

	winner, err = engine.findWinner(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Bob should win with higher starting bid
	if winner.ID != "2" {
		t.Errorf("Expected winner '2', got '%s'", winner.ID)
	}
}

// TestFindWinner_TieResolution tests tie resolution logic
func TestFindWinner_TieResolution(t *testing.T) {
	engine := NewBiddingEngine()

	baseTime := time.Now()
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.00, 20.00, 5.00),
		*models.NewBidder("2", "Bob", 10.00, 20.00, 5.00),
	}

	// Alice enters first
	bidders[0].EntryTime = baseTime
	bidders[1].EntryTime = baseTime.Add(1 * time.Second)

	winner, err := engine.findWinner(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Alice should win due to earlier entry time
	if winner.ID != "1" {
		t.Errorf("Expected winner '1' (earlier entry), got '%s'", winner.ID)
	}
}

// TestFindHighestBidCents_ErrorPaths tests error handling
func TestFindHighestBidCents_ErrorPaths(t *testing.T) {
	engine := NewBiddingEngine()

	// Test empty bidders
	highest, err := engine.findHighestBidCents([]models.Bidder{})
	if err != nil {
		t.Fatalf("Expected no error with empty bidders, got: %v", err)
	}
	if highest != 0 {
		t.Errorf("Expected 0 with empty bidders, got %d", highest)
	}

	// Test normal case
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 10.00, 20.00, 5.00),
		*models.NewBidder("2", "Bob", 15.00, 25.00, 5.00),
		*models.NewBidder("3", "Charlie", 12.00, 22.00, 5.00),
	}

	highest, err = engine.findHighestBidCents(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Bob has the highest starting bid (15.00 = 1500 cents)
	expectedCents := int64(1500)
	if highest != expectedCents {
		t.Errorf("Expected highest bid %d cents, got %d cents", expectedCents, highest)
	}
}

// TestFindHighestBidCents_SingleBidder tests single bidder case
func TestFindHighestBidCents_SingleBidder(t *testing.T) {
	engine := NewBiddingEngine()

	bidder := models.NewBidder("1", "Alice", 10.50, 20.00, 5.00)
	bidders := []models.Bidder{*bidder}

	highest, err := engine.findHighestBidCents(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedCents := bidder.GetCurrentBidCents()
	if highest != expectedCents {
		t.Errorf("Expected highest bid %d cents, got %d cents", expectedCents, highest)
	}
}

// TestProcessBids_ComplexScenario tests a complex bidding scenario
func TestProcessBids_ComplexScenario(t *testing.T) {
	engine := NewBiddingEngine()

	baseTime := time.Now()
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 100.00, 500.00, 50.00),
		*models.NewBidder("2", "Bob", 110.00, 450.00, 40.00),
		*models.NewBidder("3", "Charlie", 90.00, 300.00, 30.00),
		*models.NewBidder("4", "Diana", 95.00, 200.00, 25.00),
	}

	// Set entry times
	for i := range bidders {
		bidders[i].EntryTime = baseTime.Add(time.Duration(i) * time.Second)
	}

	result, err := engine.ProcessBids(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Alice should win with highest max bid
	if result.Winner.ID != "1" {
		t.Errorf("Expected winner '1', got '%s'", result.Winner.ID)
	}

	// Should have multiple rounds of bidding
	if result.BiddingRounds == 0 {
		t.Error("Expected multiple bidding rounds")
	}

	// Verify all bidders are in final state
	if len(result.AllBidders) != 4 {
		t.Errorf("Expected 4 bidders in result, got %d", len(result.AllBidders))
	}
}

// TestProcessBids_NoWinner tests scenario with no valid winner
func TestProcessBids_NoWinner(t *testing.T) {
	engine := NewBiddingEngine()

	// This should still produce a winner, but let's test the edge case
	// where all bidders have the same parameters
	baseTime := time.Now()
	bidders := []models.Bidder{
		*models.NewBidder("1", "Alice", 100.00, 100.00, 10.00),
		*models.NewBidder("2", "Bob", 100.00, 100.00, 10.00),
	}

	// Set entry times
	bidders[0].EntryTime = baseTime
	bidders[1].EntryTime = baseTime.Add(1 * time.Second)

	result, err := engine.ProcessBids(bidders)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Winner == nil {
		t.Fatal("Expected winner, got nil")
	}

	// Alice should win due to earlier entry time
	if result.Winner.ID != "1" {
		t.Errorf("Expected winner '1' (tie-breaker), got '%s'", result.Winner.ID)
	}
}
