package models

import (
	"testing"
)

// TestNewBidder tests the NewBidder constructor
func TestNewBidder(t *testing.T) {
	id := "bidder1"
	name := "Alice"
	startingBid := 10.50
	maxBid := 25.75
	autoIncrement := 2.25

	bidder := NewBidder(id, name, startingBid, maxBid, autoIncrement)

	if bidder == nil {
		t.Fatal("Expected bidder, got nil")
	}

	if bidder.ID != id {
		t.Errorf("Expected ID '%s', got '%s'", id, bidder.ID)
	}

	if bidder.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, bidder.Name)
	}

	if bidder.StartingBid != startingBid {
		t.Errorf("Expected starting bid %.2f, got %.2f", startingBid, bidder.StartingBid)
	}

	if bidder.MaxBid != maxBid {
		t.Errorf("Expected max bid %.2f, got %.2f", maxBid, bidder.MaxBid)
	}

	if bidder.AutoIncrement != autoIncrement {
		t.Errorf("Expected auto increment %.2f, got %.2f", autoIncrement, bidder.AutoIncrement)
	}

	if bidder.CurrentBid != startingBid {
		t.Errorf("Expected current bid to equal starting bid %.2f, got %.2f", startingBid, bidder.CurrentBid)
	}

	if !bidder.IsActive {
		t.Error("Expected bidder to be active initially")
	}

	// Test that entry time is set
	if bidder.EntryTime.IsZero() {
		t.Error("Expected entry time to be set")
	}

	// Test that cents values are properly initialized
	expectedStartingCents := DollarsToCents(startingBid)
	if bidder.GetStartingBidCents() != expectedStartingCents {
		t.Errorf("Expected starting bid cents %d, got %d", expectedStartingCents, bidder.GetStartingBidCents())
	}

	expectedMaxCents := DollarsToCents(maxBid)
	if bidder.GetMaxBidCents() != expectedMaxCents {
		t.Errorf("Expected max bid cents %d, got %d", expectedMaxCents, bidder.GetMaxBidCents())
	}

	expectedIncrementCents := DollarsToCents(autoIncrement)
	if bidder.GetAutoIncrementCents() != expectedIncrementCents {
		t.Errorf("Expected auto increment cents %d, got %d", expectedIncrementCents, bidder.GetAutoIncrementCents())
	}

	expectedCurrentCents := DollarsToCents(startingBid)
	if bidder.GetCurrentBidCents() != expectedCurrentCents {
		t.Errorf("Expected current bid cents %d, got %d", expectedCurrentCents, bidder.GetCurrentBidCents())
	}
}

// TestBidder_CanIncrement tests the CanIncrement method
func TestBidder_CanIncrement(t *testing.T) {
	tests := []struct {
		name          string
		startingBid   float64
		maxBid        float64
		autoIncrement float64
		isActive      bool
		expected      bool
	}{
		{
			name:          "Can increment - active bidder with room",
			startingBid:   10.00,
			maxBid:        20.00,
			autoIncrement: 5.00,
			isActive:      true,
			expected:      true,
		},
		{
			name:          "Cannot increment - inactive bidder",
			startingBid:   10.00,
			maxBid:        20.00,
			autoIncrement: 5.00,
			isActive:      false,
			expected:      false,
		},
		{
			name:          "Cannot increment - would exceed max bid",
			startingBid:   18.00,
			maxBid:        20.00,
			autoIncrement: 5.00,
			isActive:      true,
			expected:      false,
		},
		{
			name:          "Can increment - exactly at max after increment",
			startingBid:   15.00,
			maxBid:        20.00,
			autoIncrement: 5.00,
			isActive:      true,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidder := NewBidder("1", "Test", tt.startingBid, tt.maxBid, tt.autoIncrement)
			bidder.IsActive = tt.isActive

			result := bidder.CanIncrement()
			if result != tt.expected {
				t.Errorf("Expected CanIncrement() = %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestBidder_Increment tests the Increment method
func TestBidder_Increment(t *testing.T) {
	tests := []struct {
		name            string
		startingBid     float64
		maxBid          float64
		autoIncrement   float64
		expectedSuccess bool
		expectedNewBid  float64
		expectedActive  bool
	}{
		{
			name:            "Successful increment with room remaining",
			startingBid:     10.00,
			maxBid:          20.00,
			autoIncrement:   5.00,
			expectedSuccess: true,
			expectedNewBid:  15.00,
			expectedActive:  true,
		},
		{
			name:            "Increment to max bid - becomes inactive",
			startingBid:     15.00,
			maxBid:          20.00,
			autoIncrement:   5.00,
			expectedSuccess: true,
			expectedNewBid:  20.00,
			expectedActive:  false,
		},
		{
			name:            "Cannot increment - already at max",
			startingBid:     20.00,
			maxBid:          20.00,
			autoIncrement:   5.00,
			expectedSuccess: false,
			expectedNewBid:  20.00,
			expectedActive:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidder := NewBidder("1", "Test", tt.startingBid, tt.maxBid, tt.autoIncrement)

			// If starting at max, set inactive
			if tt.startingBid >= tt.maxBid {
				bidder.IsActive = false
			}

			success := bidder.Increment()
			if success != tt.expectedSuccess {
				t.Errorf("Expected Increment() = %v, got %v", tt.expectedSuccess, success)
			}

			if bidder.CurrentBid != tt.expectedNewBid {
				t.Errorf("Expected current bid %.2f, got %.2f", tt.expectedNewBid, bidder.CurrentBid)
			}

			if bidder.IsActive != tt.expectedActive {
				t.Errorf("Expected IsActive = %v, got %v", tt.expectedActive, bidder.IsActive)
			}

			// Verify cents are in sync
			expectedCents := DollarsToCents(tt.expectedNewBid)
			if bidder.GetCurrentBidCents() != expectedCents {
				t.Errorf("Expected current bid cents %d, got %d", expectedCents, bidder.GetCurrentBidCents())
			}
		})
	}
}

// TestBidder_GetMethods tests all the getter methods
func TestBidder_GetMethods(t *testing.T) {
	startingBid := 12.34
	maxBid := 56.78
	autoIncrement := 9.01

	bidder := NewBidder("1", "Test", startingBid, maxBid, autoIncrement)

	// Test GetCurrentBidCents
	expectedCurrentCents := DollarsToCents(startingBid)
	if bidder.GetCurrentBidCents() != expectedCurrentCents {
		t.Errorf("Expected current bid cents %d, got %d", expectedCurrentCents, bidder.GetCurrentBidCents())
	}

	// Test GetMaxBidCents
	expectedMaxCents := DollarsToCents(maxBid)
	if bidder.GetMaxBidCents() != expectedMaxCents {
		t.Errorf("Expected max bid cents %d, got %d", expectedMaxCents, bidder.GetMaxBidCents())
	}

	// Test GetAutoIncrementCents
	expectedIncrementCents := DollarsToCents(autoIncrement)
	if bidder.GetAutoIncrementCents() != expectedIncrementCents {
		t.Errorf("Expected auto increment cents %d, got %d", expectedIncrementCents, bidder.GetAutoIncrementCents())
	}

	// Test GetStartingBidCents
	expectedStartingCents := DollarsToCents(startingBid)
	if bidder.GetStartingBidCents() != expectedStartingCents {
		t.Errorf("Expected starting bid cents %d, got %d", expectedStartingCents, bidder.GetStartingBidCents())
	}
}

// TestBidder_SyncFloatFields tests the SyncFloatFields method
func TestBidder_SyncFloatFields(t *testing.T) {
	bidder := NewBidder("1", "Test", 10.00, 20.00, 2.50)

	// Increment the bidder to change internal state
	bidder.Increment()

	// Manually modify the float fields to test sync
	bidder.CurrentBid = 999.99
	bidder.StartingBid = 888.88
	bidder.MaxBid = 777.77
	bidder.AutoIncrement = 666.66

	// Sync should restore correct values from cents
	bidder.SyncFloatFields()

	expectedCurrent := CentsToDollars(bidder.GetCurrentBidCents())
	if bidder.CurrentBid != expectedCurrent {
		t.Errorf("Expected current bid %.2f after sync, got %.2f", expectedCurrent, bidder.CurrentBid)
	}

	expectedStarting := CentsToDollars(bidder.GetStartingBidCents())
	if bidder.StartingBid != expectedStarting {
		t.Errorf("Expected starting bid %.2f after sync, got %.2f", expectedStarting, bidder.StartingBid)
	}

	expectedMax := CentsToDollars(bidder.GetMaxBidCents())
	if bidder.MaxBid != expectedMax {
		t.Errorf("Expected max bid %.2f after sync, got %.2f", expectedMax, bidder.MaxBid)
	}

	expectedIncrement := CentsToDollars(bidder.GetAutoIncrementCents())
	if bidder.AutoIncrement != expectedIncrement {
		t.Errorf("Expected auto increment %.2f after sync, got %.2f", expectedIncrement, bidder.AutoIncrement)
	}
}

// TestBidder_MultipleIncrements tests multiple increments
func TestBidder_MultipleIncrements(t *testing.T) {
	bidder := NewBidder("1", "Test", 10.00, 25.00, 5.00)

	// First increment: 10.00 -> 15.00
	success := bidder.Increment()
	if !success {
		t.Fatal("Expected first increment to succeed")
	}
	if bidder.CurrentBid != 15.00 {
		t.Errorf("Expected current bid 15.00, got %.2f", bidder.CurrentBid)
	}
	if !bidder.IsActive {
		t.Error("Expected bidder to remain active")
	}

	// Second increment: 15.00 -> 20.00
	success = bidder.Increment()
	if !success {
		t.Fatal("Expected second increment to succeed")
	}
	if bidder.CurrentBid != 20.00 {
		t.Errorf("Expected current bid 20.00, got %.2f", bidder.CurrentBid)
	}
	if !bidder.IsActive {
		t.Error("Expected bidder to remain active")
	}

	// Third increment: 20.00 -> 25.00 (max bid, should become inactive)
	success = bidder.Increment()
	if !success {
		t.Fatal("Expected third increment to succeed")
	}
	if bidder.CurrentBid != 25.00 {
		t.Errorf("Expected current bid 25.00, got %.2f", bidder.CurrentBid)
	}
	if bidder.IsActive {
		t.Error("Expected bidder to become inactive at max bid")
	}

	// Fourth increment: should fail
	success = bidder.Increment()
	if success {
		t.Error("Expected fourth increment to fail")
	}
	if bidder.CurrentBid != 25.00 {
		t.Errorf("Expected current bid to remain 25.00, got %.2f", bidder.CurrentBid)
	}
}

// TestBidder_PrecisionHandling tests precision handling with fractional cents
func TestBidder_PrecisionHandling(t *testing.T) {
	// Use values that might cause floating-point precision issues
	bidder := NewBidder("1", "Test", 10.01, 20.99, 0.33)

	// Test that cents conversion is accurate
	expectedStartingCents := int64(1001) // 10.01 * 100
	if bidder.GetStartingBidCents() != expectedStartingCents {
		t.Errorf("Expected starting bid cents %d, got %d", expectedStartingCents, bidder.GetStartingBidCents())
	}

	expectedMaxCents := int64(2099) // 20.99 * 100
	if bidder.GetMaxBidCents() != expectedMaxCents {
		t.Errorf("Expected max bid cents %d, got %d", expectedMaxCents, bidder.GetMaxBidCents())
	}

	expectedIncrementCents := int64(33) // 0.33 * 100
	if bidder.GetAutoIncrementCents() != expectedIncrementCents {
		t.Errorf("Expected auto increment cents %d, got %d", expectedIncrementCents, bidder.GetAutoIncrementCents())
	}

	// Test increment with precision
	success := bidder.Increment()
	if !success {
		t.Fatal("Expected increment to succeed")
	}

	expectedNewCents := int64(1034) // 1001 + 33
	if bidder.GetCurrentBidCents() != expectedNewCents {
		t.Errorf("Expected current bid cents %d, got %d", expectedNewCents, bidder.GetCurrentBidCents())
	}

	expectedNewDollars := 10.34 // Should be precise
	if bidder.CurrentBid != expectedNewDollars {
		t.Errorf("Expected current bid %.2f, got %.2f", expectedNewDollars, bidder.CurrentBid)
	}
}
