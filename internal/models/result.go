package models

// BidResult represents the outcome of an auction bidding process
type BidResult struct {
	Winner        *Bidder  `json:"winner"`         // Winning bidder
	WinningBid    float64  `json:"winning_bid"`    // Final winning amount
	TotalBidders  int      `json:"total_bidders"`  // Number of participants
	BiddingRounds int      `json:"bidding_rounds"` // Number of increment rounds
	AllBidders    []Bidder `json:"all_bidders"`    // Final state of all bidders

	// Internal field for precise calculations
	winningBidCents int64 // Winning bid in cents
}

// NewBidResult creates a new BidResult with the provided parameters
func NewBidResult(winner *Bidder, winningBid float64, totalBidders, biddingRounds int, allBidders []Bidder) *BidResult {
	result := &BidResult{
		Winner:        winner,
		WinningBid:    winningBid,
		TotalBidders:  totalBidders,
		BiddingRounds: biddingRounds,
		AllBidders:    allBidders,
	}

	// Store precise winning bid in cents
	result.winningBidCents = DollarsToCents(winningBid)

	// Ensure all bidders have synced float fields
	for i := range result.AllBidders {
		result.AllBidders[i].SyncFloatFields()
	}

	return result
}

// NewBidResultFromCents creates a new BidResult with winning bid specified in cents
func NewBidResultFromCents(winner *Bidder, winningBidCents int64, totalBidders, biddingRounds int, allBidders []Bidder) *BidResult {
	result := &BidResult{
		Winner:          winner,
		WinningBid:      CentsToDollars(winningBidCents),
		TotalBidders:    totalBidders,
		BiddingRounds:   biddingRounds,
		AllBidders:      allBidders,
		winningBidCents: winningBidCents,
	}

	// Ensure all bidders have synced float fields
	for i := range result.AllBidders {
		result.AllBidders[i].SyncFloatFields()
	}

	return result
}

// GetWinningBidCents returns the winning bid in cents for precise calculations
func (br *BidResult) GetWinningBidCents() int64 {
	return br.winningBidCents
}
