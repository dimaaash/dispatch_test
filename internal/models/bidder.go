package models

import (
	"time"
)

// Bidder represents a participant in the auction with their bidding parameters
type Bidder struct {
	ID            string    `json:"id" validate:"required"`                  // Unique identifier
	Name          string    `json:"name" validate:"required"`                // Bidder name
	StartingBid   float64   `json:"starting_bid" validate:"required,gt=0"`   // Initial bid amount
	MaxBid        float64   `json:"max_bid" validate:"required,gt=0"`        // Maximum willing to pay
	AutoIncrement float64   `json:"auto_increment" validate:"required,gt=0"` // Increment amount
	CurrentBid    float64   `json:"current_bid"`                             // Current active bid
	EntryTime     time.Time `json:"entry_time"`                              // When bid was submitted
	IsActive      bool      `json:"is_active"`                               // Whether bidder can still increment

	// Internal fields for precise calculations (stored as cents)
	startingBidCents   int64 // Starting bid in cents
	maxBidCents        int64 // Maximum bid in cents
	autoIncrementCents int64 // Auto increment in cents
	currentBidCents    int64 // Current bid in cents
}

// NewBidder creates a new Bidder with the provided parameters
func NewBidder(id, name string, startingBid, maxBid, autoIncrement float64) *Bidder {
	bidder := &Bidder{
		ID:            id,
		Name:          name,
		StartingBid:   startingBid,
		MaxBid:        maxBid,
		AutoIncrement: autoIncrement,
		CurrentBid:    startingBid,
		EntryTime:     time.Now(),
		IsActive:      true,
	}

	// Convert to cents for precise calculations
	bidder.startingBidCents = DollarsToCents(startingBid)
	bidder.maxBidCents = DollarsToCents(maxBid)
	bidder.autoIncrementCents = DollarsToCents(autoIncrement)
	bidder.currentBidCents = bidder.startingBidCents

	return bidder
}

// CanIncrement checks if the bidder can increment their current bid
func (b *Bidder) CanIncrement() bool {
	return b.IsActive && (b.currentBidCents+b.autoIncrementCents) <= b.maxBidCents
}

// Increment increases the bidder's current bid by their auto-increment amount
func (b *Bidder) Increment() bool {
	if !b.CanIncrement() {
		return false
	}
	b.currentBidCents += b.autoIncrementCents
	if b.currentBidCents >= b.maxBidCents {
		b.currentBidCents = b.maxBidCents
		b.IsActive = false
	}
	// Update the float64 field for external API compatibility
	b.CurrentBid = CentsToDollars(b.currentBidCents)
	return true
}

// GetCurrentBidCents returns the current bid in cents for precise calculations
func (b *Bidder) GetCurrentBidCents() int64 {
	return b.currentBidCents
}

// GetMaxBidCents returns the maximum bid in cents for precise calculations
func (b *Bidder) GetMaxBidCents() int64 {
	return b.maxBidCents
}

// GetAutoIncrementCents returns the auto increment in cents for precise calculations
func (b *Bidder) GetAutoIncrementCents() int64 {
	return b.autoIncrementCents
}

// GetStartingBidCents returns the starting bid in cents for precise calculations
func (b *Bidder) GetStartingBidCents() int64 {
	return b.startingBidCents
}

// SyncFloatFields updates the float64 fields from the precise cent values
func (b *Bidder) SyncFloatFields() {
	b.CurrentBid = CentsToDollars(b.currentBidCents)
	b.StartingBid = CentsToDollars(b.startingBidCents)
	b.MaxBid = CentsToDollars(b.maxBidCents)
	b.AutoIncrement = CentsToDollars(b.autoIncrementCents)
}
