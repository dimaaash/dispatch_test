package models

import (
	"testing"
)

// TestDollarsToCents tests the conversion from dollars to cents
func TestDollarsToCents(t *testing.T) {
	tests := []struct {
		name     string
		dollars  float64
		expected int64
	}{
		{"Whole dollar", 1.00, 100},
		{"Dollar with cents", 1.23, 123},
		{"Zero", 0.00, 0},
		{"Small amount", 0.01, 1},
		{"Large amount", 999999.99, 99999999},
		{"Fractional cent rounds up", 1.006, 101},
		{"Fractional cent rounds down", 1.004, 100},
		{"Negative amount", -1.23, -123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DollarsToCents(tt.dollars)
			if result != tt.expected {
				t.Errorf("DollarsToCents(%.3f) = %d, expected %d", tt.dollars, result, tt.expected)
			}
		})
	}
}

// TestCentsToDollars tests the conversion from cents to dollars
func TestCentsToDollars(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		expected float64
	}{
		{"Whole dollar", 100, 1.00},
		{"Dollar with cents", 123, 1.23},
		{"Zero", 0, 0.00},
		{"Single cent", 1, 0.01},
		{"Large amount", 99999999, 999999.99},
		{"Negative amount", -123, -1.23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CentsToDollars(tt.cents)
			if result != tt.expected {
				t.Errorf("CentsToDollars(%d) = %.2f, expected %.2f", tt.cents, result, tt.expected)
			}
		})
	}
}

// TestRoundTripConversion tests that converting dollars to cents and back preserves precision
func TestRoundTripConversion(t *testing.T) {
	tests := []struct {
		name      string
		dollars   float64
		tolerance float64 // Acceptable difference due to rounding
	}{
		{"Exact cents", 1.23, 0.0},
		{"Rounded up", 1.006, 0.005},
		{"Rounded down", 1.004, 0.005},
		{"Large amount", 999999.99, 0.0},
		{"Small amount", 0.01, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cents := DollarsToCents(tt.dollars)
			backToDollars := CentsToDollars(cents)

			diff := backToDollars - tt.dollars
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("Round-trip conversion failed: %.3f -> %d -> %.3f (diff: %.6f, tolerance: %.6f)",
					tt.dollars, cents, backToDollars, diff, tt.tolerance)
			}
		})
	}
}

// TestPrecisionEdgeCases tests edge cases that could cause precision issues
func TestPrecisionEdgeCases(t *testing.T) {
	// Test very small amounts
	result := DollarsToCents(0.001)
	if result != 0 {
		t.Errorf("Expected 0.001 to round to 0 cents, got %d", result)
	}

	result = DollarsToCents(0.006)
	if result != 1 {
		t.Errorf("Expected 0.006 to round to 1 cent, got %d", result)
	}

	// Test maximum precision
	result = DollarsToCents(0.999)
	if result != 100 {
		t.Errorf("Expected 0.999 to convert to 100 cents, got %d", result)
	}

	// Test negative rounding (math.Round rounds ties to even)
	result = DollarsToCents(-1.005)
	if result != -100 {
		t.Errorf("Expected -1.005 to round to -100 cents, got %d", result)
	}
}
