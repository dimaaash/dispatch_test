package validation

import (
	"fmt"
	"strings"

	"auction-bidding-algorithm/internal/models"
)

// BidValidator interface defines methods for validating bidder information
type BidValidator interface {
	ValidateBidder(bidder models.Bidder) error
	ValidateBidders(bidders []models.Bidder) error
}

// DefaultBidValidator implements the BidValidator interface with standard validation rules
type DefaultBidValidator struct{}

// NewBidValidator creates a new instance of DefaultBidValidator
func NewBidValidator() BidValidator {
	return &DefaultBidValidator{}
}

// ValidateBidder validates a single bidder's parameters according to auction rules
func (v *DefaultBidValidator) ValidateBidder(bidder models.Bidder) error {
	var validationErrors []*models.ValidationError

	// Validate required fields
	if strings.TrimSpace(bidder.ID) == "" {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue("", "ID", "bidder ID is required", bidder.ID))
	}

	if strings.TrimSpace(bidder.Name) == "" {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue(bidder.ID, "Name", "bidder name is required", bidder.Name))
	}

	// Validate bid amounts are non-negative (Requirement 6.3)
	if bidder.StartingBid < 0 {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue(bidder.ID, "StartingBid", "starting bid cannot be negative", fmt.Sprintf("%.2f", bidder.StartingBid)))
	}

	if bidder.MaxBid < 0 {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue(bidder.ID, "MaxBid", "maximum bid cannot be negative", fmt.Sprintf("%.2f", bidder.MaxBid)))
	}

	// Validate auto-increment is positive (Requirement 6.2)
	if bidder.AutoIncrement <= 0 {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue(bidder.ID, "AutoIncrement", "auto-increment amount must be greater than zero", fmt.Sprintf("%.2f", bidder.AutoIncrement)))
	}

	// Validate starting bid does not exceed maximum bid (Requirement 6.1)
	if bidder.StartingBid > bidder.MaxBid {
		validationErrors = append(validationErrors, models.NewValidationErrorWithValue(bidder.ID, "StartingBid", "starting bid cannot be greater than maximum bid", fmt.Sprintf("starting: %.2f, max: %.2f", bidder.StartingBid, bidder.MaxBid)))
	}

	// If there are validation errors, return them as an AuctionError
	if len(validationErrors) > 0 {
		auctionErr := models.NewAuctionError(models.ErrorTypeValidation, fmt.Sprintf("validation failed for bidder %s", bidder.ID), validationErrors)
		auctionErr.WithOperation("ValidateBidder")
		auctionErr.AddContext("bidder_id", bidder.ID)
		auctionErr.AddContext("bidder_name", bidder.Name)
		return auctionErr
	}

	return nil
}

// ValidateBidders validates multiple bidders and collects all validation errors
func (v *DefaultBidValidator) ValidateBidders(bidders []models.Bidder) error {
	if len(bidders) == 0 {
		auctionErr := models.NewAuctionError(models.ErrorTypeValidation, "no bidders provided", nil)
		auctionErr.WithOperation("ValidateBidders")
		auctionErr.AddContext("bidder_count", "0")
		return auctionErr
	}

	var allValidationErrors []*models.ValidationError
	bidderIDs := make(map[string]bool)
	validBidderCount := 0

	for i, bidder := range bidders {
		// Check for duplicate bidder IDs
		if bidderIDs[bidder.ID] {
			allValidationErrors = append(allValidationErrors, models.NewValidationErrorWithValue(bidder.ID, "ID", "duplicate bidder ID", bidder.ID))
			continue
		}
		bidderIDs[bidder.ID] = true

		// Validate individual bidder
		if err := v.ValidateBidder(bidder); err != nil {
			if auctionErr, ok := err.(*models.AuctionError); ok {
				// Add position context to each validation error
				for _, detail := range auctionErr.Details {
					detail.Value = fmt.Sprintf("position %d: %s", i+1, detail.Value)
				}
				allValidationErrors = append(allValidationErrors, auctionErr.Details...)
			} else {
				// Handle unexpected error types
				allValidationErrors = append(allValidationErrors, models.NewValidationErrorWithValue(bidder.ID, "unknown", "unexpected validation error", err.Error()))
			}
		} else {
			validBidderCount++
		}
	}

	// If there are validation errors, return them as an AuctionError
	if len(allValidationErrors) > 0 {
		errorsByBidder := make(map[string]int)
		for _, err := range allValidationErrors {
			errorsByBidder[err.BidderID]++
		}

		auctionErr := models.NewAuctionError(models.ErrorTypeValidation, fmt.Sprintf("validation failed for %d out of %d bidders", len(errorsByBidder), len(bidders)), allValidationErrors)
		auctionErr.WithOperation("ValidateBidders")
		auctionErr.AddContext("total_bidders", fmt.Sprintf("%d", len(bidders)))
		auctionErr.AddContext("valid_bidders", fmt.Sprintf("%d", validBidderCount))
		auctionErr.AddContext("invalid_bidders", fmt.Sprintf("%d", len(errorsByBidder)))
		auctionErr.AddContext("total_validation_errors", fmt.Sprintf("%d", len(allValidationErrors)))

		return auctionErr
	}

	return nil
}
