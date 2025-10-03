// Package auction provides an automated bidding algorithm for computerized auction sites.
// It determines winning bids by automatically incrementing bidders' current bids based on
// their maximum bid limits and auto-increment amounts.
package auction

import (
	"fmt"

	"auction-bidding-algorithm/internal"
	"auction-bidding-algorithm/internal/models"
	"auction-bidding-algorithm/internal/validation"
)

// AuctionProcessor defines the interface for processing auction bids
type AuctionProcessor interface {
	DetermineWinner(bidders []models.Bidder) (*models.BidResult, error)
}

// BiddingEngine defines the interface for processing bids
type BiddingEngine interface {
	ProcessBids(bidders []models.Bidder) (*models.BidResult, error)
}

// AuctionService orchestrates the entire auction process including validation and bid processing
type AuctionService struct {
	validator validation.BidValidator
	engine    BiddingEngine
}

// NewAuctionService creates a new AuctionService with default validator and engine
func NewAuctionService() *AuctionService {
	return &AuctionService{
		validator: validation.NewBidValidator(),
		engine:    internal.NewBiddingEngine(),
	}
}

// DetermineWinner validates inputs and processes bids to determine the auction winner
// This method implements the main orchestration logic for the auction process
func (as *AuctionService) DetermineWinner(bidders []models.Bidder) (*models.BidResult, error) {
	// Validate all bidders first (Requirement 1.1)
	if err := as.validator.ValidateBidders(bidders); err != nil {
		// Wrap validation error with additional context
		if auctionErr, ok := err.(*models.AuctionError); ok {
			auctionErr.WithOperation("DetermineWinner.Validation")
			auctionErr.AddContext("service", "AuctionService")
			return nil, auctionErr
		}
		// Handle unexpected error types
		wrappedErr := models.NewAuctionErrorWithCause(models.ErrorTypeValidation, "unexpected validation error", err)
		wrappedErr.WithOperation("DetermineWinner.Validation")
		wrappedErr.AddContext("service", "AuctionService")
		return nil, wrappedErr
	}

	// Process the bids using the bidding engine (Requirement 1.2)
	result, err := as.engine.ProcessBids(bidders)
	if err != nil {
		// Wrap processing error with additional context
		if auctionErr, ok := err.(*models.AuctionError); ok {
			auctionErr.WithOperation("DetermineWinner.Processing")
			auctionErr.AddContext("service", "AuctionService")
			return nil, auctionErr
		}
		// Handle unexpected error types
		wrappedErr := models.NewAuctionErrorWithCause(models.ErrorTypeProcessing, "unexpected processing error", err)
		wrappedErr.WithOperation("DetermineWinner.Processing")
		wrappedErr.AddContext("service", "AuctionService")
		return nil, wrappedErr
	}

	// Ensure proper result formatting (Requirement 1.3)
	if result == nil {
		processingErr := models.NewAuctionError(models.ErrorTypeProcessing, "failed to process bids: result is nil", nil)
		processingErr.WithOperation("DetermineWinner.ResultValidation")
		processingErr.AddContext("service", "AuctionService")
		processingErr.AddContext("bidder_count", fmt.Sprintf("%d", len(bidders)))
		return nil, processingErr
	}

	return result, nil
}
