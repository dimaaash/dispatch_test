package internal

import (
	"fmt"
	"sort"

	"auction-bidding-algorithm/internal/models"
)

// BiddingEngine handles the core auction bidding algorithm
type BiddingEngine struct {
	maxRounds int // Maximum number of bidding rounds to prevent infinite loops
}

// NewBiddingEngine creates a new BiddingEngine with default settings
func NewBiddingEngine() *BiddingEngine {
	return &BiddingEngine{
		maxRounds: 1000, // Reasonable limit to prevent infinite loops
	}
}

// ProcessBids executes the core bidding algorithm and returns the result
func (be *BiddingEngine) ProcessBids(bidders []models.Bidder) (*models.BidResult, error) {
	if len(bidders) == 0 {
		return models.NewBidResult(nil, 0, 0, 0, bidders), nil
	}

	// Make a copy of bidders to avoid modifying the original slice
	workingBidders := make([]models.Bidder, len(bidders))
	copy(workingBidders, bidders)

	// Initialize current bids to starting bids and sort by entry time for tie resolution
	for i := range workingBidders {
		// Reinitialize the bidder to ensure precise calculations
		bidder := &workingBidders[i]
		originalEntryTime := bidder.EntryTime // Preserve original entry time
		*bidder = *models.NewBidder(bidder.ID, bidder.Name, bidder.StartingBid, bidder.MaxBid, bidder.AutoIncrement)
		bidder.EntryTime = originalEntryTime // Restore original entry time
	}

	// Sort bidders by entry time for consistent tie resolution
	sort.Slice(workingBidders, func(i, j int) bool {
		return workingBidders[i].EntryTime.Before(workingBidders[j].EntryTime)
	})

	rounds := 0

	// Iterative bidding process with timeout protection
	for rounds < be.maxRounds {
		// Check if any losing bidders can increment
		incremented, err := be.IncrementBids(workingBidders)
		if err != nil {
			processingErr := models.NewProcessingErrorWithCause("failed to increment bids", err, len(bidders), rounds)
			processingErr.WithOperation("ProcessBids.IncrementBids")
			processingErr.AddContext("round", fmt.Sprintf("%d", rounds))
			processingErr.AddContext("max_rounds", fmt.Sprintf("%d", be.maxRounds))
			return nil, processingErr
		}

		if !incremented {
			break // No more increments possible
		}
		rounds++
	}

	// Check for timeout condition
	if rounds >= be.maxRounds {
		timeoutErr := models.NewTimeoutError("bidding process exceeded maximum rounds", "ProcessBids", fmt.Sprintf("%d rounds", be.maxRounds))
		timeoutErr.WithOperation("ProcessBids.TimeoutCheck")
		timeoutErr.AddContext("bidder_count", fmt.Sprintf("%d", len(bidders)))
		timeoutErr.AddContext("final_round", fmt.Sprintf("%d", rounds))
		return nil, timeoutErr
	}

	// Find the winner (highest current bid, earliest entry time for ties)
	winner, err := be.findWinner(workingBidders)
	if err != nil {
		processingErr := models.NewProcessingErrorWithCause("failed to determine winner", err, len(bidders), rounds)
		processingErr.WithOperation("ProcessBids.FindWinner")
		processingErr.AddContext("rounds_completed", fmt.Sprintf("%d", rounds))
		return nil, processingErr
	}

	if winner == nil {
		return models.NewBidResult(nil, 0, len(bidders), rounds, workingBidders), nil
	}

	// Calculate minimum winning bid using precise arithmetic
	winningBidCents, err := be.CalculateMinimumWinningBidCents(workingBidders, winner)
	if err != nil {
		processingErr := models.NewProcessingErrorWithCause("failed to calculate minimum winning bid", err, len(bidders), rounds)
		processingErr.WithOperation("ProcessBids.CalculateMinimumWinningBidCents")
		processingErr.AddContext("winner_id", winner.ID)
		processingErr.AddContext("winner_current_bid", fmt.Sprintf("%.2f", winner.CurrentBid))
		return nil, processingErr
	}

	return models.NewBidResultFromCents(winner, winningBidCents, len(bidders), rounds, workingBidders), nil
}

// IncrementBids increments the bids of losing bidders who can afford to increment
// Returns true if any bids were incremented, false if no more increments are possible
func (be *BiddingEngine) IncrementBids(bidders []models.Bidder) (bool, error) {
	if len(bidders) <= 1 {
		return false, nil
	}

	// Find current highest bid using precise arithmetic
	highestBidCents, err := be.findHighestBidCents(bidders)
	if err != nil {
		return false, models.NewProcessingErrorWithCause("failed to find highest bid", err, len(bidders), 0)
	}

	anyIncremented := false

	// Increment losing bidders who can afford it
	for i := range bidders {
		bidder := &bidders[i]

		// Skip if this bidder is already at the highest bid or can't increment
		if bidder.GetCurrentBidCents() >= highestBidCents || !bidder.CanIncrement() {
			continue
		}

		// Increment the bidder
		if bidder.Increment() {
			anyIncremented = true
		} else {
			// This shouldn't happen if CanIncrement() returned true
			systemErr := models.NewSystemError("bidder increment failed despite CanIncrement() returning true", "BiddingEngine", "medium")
			systemErr.WithOperation("IncrementBids")
			systemErr.AddContext("bidder_id", bidder.ID)
			systemErr.AddContext("current_bid", fmt.Sprintf("%.2f", bidder.CurrentBid))
			systemErr.AddContext("max_bid", fmt.Sprintf("%.2f", bidder.MaxBid))
			systemErr.AddContext("auto_increment", fmt.Sprintf("%.2f", bidder.AutoIncrement))
			return false, systemErr
		}
	}

	return anyIncremented, nil
}

// CalculateMinimumWinningBidCents determines the lowest amount the winner needs to pay in cents
func (be *BiddingEngine) CalculateMinimumWinningBidCents(bidders []models.Bidder, winner *models.Bidder) (int64, error) {
	if winner == nil {
		inputErr := models.NewInputError("winner cannot be nil", "winner", nil)
		inputErr.WithOperation("CalculateMinimumWinningBidCents")
		return 0, inputErr
	}

	if len(bidders) == 0 {
		inputErr := models.NewInputError("bidders slice cannot be empty", "bidders", len(bidders))
		inputErr.WithOperation("CalculateMinimumWinningBidCents")
		return 0, inputErr
	}

	// Validate winner exists in bidders slice
	winnerFound := false
	for _, bidder := range bidders {
		if bidder.ID == winner.ID {
			winnerFound = true
			break
		}
	}
	if !winnerFound {
		inputErr := models.NewInputError("winner not found in bidders slice", "winner.ID", winner.ID)
		inputErr.WithOperation("CalculateMinimumWinningBidCents")
		inputErr.AddContext("winner_id", winner.ID)
		return 0, inputErr
	}

	// Find the second highest maximum possible bid using precise arithmetic
	var secondHighestCents int64 = 0
	secondHighestBidderID := ""
	for _, bidder := range bidders {
		if bidder.ID == winner.ID {
			continue // Skip the winner
		}

		// Consider the maximum possible bid this bidder could have reached
		maxPossibleBidCents := bidder.GetMaxBidCents()
		if maxPossibleBidCents > secondHighestCents {
			secondHighestCents = maxPossibleBidCents
			secondHighestBidderID = bidder.ID
		}
	}

	// If no other bidders, winner pays their starting bid
	if secondHighestCents == 0 {
		return winner.GetStartingBidCents(), nil
	}

	// Winner pays just enough to beat the second highest bidder
	minWinningBidCents := secondHighestCents + winner.GetAutoIncrementCents()

	// But never more than their maximum bid
	if minWinningBidCents > winner.GetMaxBidCents() {
		minWinningBidCents = winner.GetMaxBidCents()
	}

	// And never less than their starting bid
	if minWinningBidCents < winner.GetStartingBidCents() {
		minWinningBidCents = winner.GetStartingBidCents()
	}

	// Validate the calculated bid is reasonable
	if minWinningBidCents < 0 {
		systemErr := models.NewSystemError("calculated minimum winning bid is negative", "BiddingEngine", "high")
		systemErr.WithOperation("CalculateMinimumWinningBidCents")
		systemErr.AddContext("calculated_bid_cents", fmt.Sprintf("%d", minWinningBidCents))
		systemErr.AddContext("calculated_bid_dollars", fmt.Sprintf("%.2f", models.CentsToDollars(minWinningBidCents)))
		systemErr.AddContext("winner_id", winner.ID)
		systemErr.AddContext("second_highest_cents", fmt.Sprintf("%d", secondHighestCents))
		systemErr.AddContext("second_highest_bidder", secondHighestBidderID)
		return 0, systemErr
	}

	return minWinningBidCents, nil
}

// findWinner identifies the bidder with the highest current bid using precise arithmetic
// In case of ties, the earliest entry wins
func (be *BiddingEngine) findWinner(bidders []models.Bidder) (*models.Bidder, error) {
	if len(bidders) == 0 {
		return nil, nil
	}

	winner := &bidders[0]

	for i := 1; i < len(bidders); i++ {
		current := &bidders[i]

		// Validate bidder data integrity using precise values
		if current.GetCurrentBidCents() < 0 {
			systemErr := models.NewSystemError("bidder has negative current bid", "BiddingEngine", "high")
			systemErr.WithOperation("findWinner")
			systemErr.AddContext("bidder_id", current.ID)
			systemErr.AddContext("current_bid_cents", fmt.Sprintf("%d", current.GetCurrentBidCents()))
			systemErr.AddContext("current_bid_dollars", fmt.Sprintf("%.2f", current.CurrentBid))
			return nil, systemErr
		}

		// Higher bid wins (using precise comparison)
		if current.GetCurrentBidCents() > winner.GetCurrentBidCents() {
			winner = current
		} else if current.GetCurrentBidCents() == winner.GetCurrentBidCents() {
			// In case of tie, earlier entry wins (bidders are already sorted by entry time)
			if current.EntryTime.Before(winner.EntryTime) {
				winner = current
			}
		}
	}

	// Final validation of winner
	if winner.GetCurrentBidCents() < 0 {
		systemErr := models.NewSystemError("winner has negative current bid", "BiddingEngine", "critical")
		systemErr.WithOperation("findWinner")
		systemErr.AddContext("winner_id", winner.ID)
		systemErr.AddContext("winner_current_bid_cents", fmt.Sprintf("%d", winner.GetCurrentBidCents()))
		systemErr.AddContext("winner_current_bid_dollars", fmt.Sprintf("%.2f", winner.CurrentBid))
		return nil, systemErr
	}

	return winner, nil
}

// findHighestBidCents returns the highest current bid among all bidders in cents
func (be *BiddingEngine) findHighestBidCents(bidders []models.Bidder) (int64, error) {
	if len(bidders) == 0 {
		return 0, nil
	}

	highestCents := bidders[0].GetCurrentBidCents()
	highestBidderID := bidders[0].ID

	// Validate first bidder's bid
	if highestCents < 0 {
		systemErr := models.NewSystemError("bidder has negative current bid", "BiddingEngine", "high")
		systemErr.WithOperation("findHighestBidCents")
		systemErr.AddContext("bidder_id", bidders[0].ID)
		systemErr.AddContext("current_bid_cents", fmt.Sprintf("%d", highestCents))
		systemErr.AddContext("current_bid_dollars", fmt.Sprintf("%.2f", bidders[0].CurrentBid))
		return 0, systemErr
	}

	for _, bidder := range bidders[1:] {
		bidderCents := bidder.GetCurrentBidCents()

		// Validate each bidder's bid
		if bidderCents < 0 {
			systemErr := models.NewSystemError("bidder has negative current bid", "BiddingEngine", "high")
			systemErr.WithOperation("findHighestBidCents")
			systemErr.AddContext("bidder_id", bidder.ID)
			systemErr.AddContext("current_bid_cents", fmt.Sprintf("%d", bidderCents))
			systemErr.AddContext("current_bid_dollars", fmt.Sprintf("%.2f", bidder.CurrentBid))
			return 0, systemErr
		}

		if bidderCents > highestCents {
			highestCents = bidderCents
			highestBidderID = bidder.ID
		}
	}

	// Final validation
	if highestCents < 0 {
		systemErr := models.NewSystemError("calculated highest bid is negative", "BiddingEngine", "critical")
		systemErr.WithOperation("findHighestBidCents")
		systemErr.AddContext("highest_bid_cents", fmt.Sprintf("%d", highestCents))
		systemErr.AddContext("highest_bid_dollars", fmt.Sprintf("%.2f", models.CentsToDollars(highestCents)))
		systemErr.AddContext("highest_bidder_id", highestBidderID)
		return 0, systemErr
	}

	return highestCents, nil
}
