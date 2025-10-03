package models

import "math"

// DollarsToCents converts a dollar amount to cents using precise rounding
func DollarsToCents(dollars float64) int64 {
	return int64(math.Round(dollars * 100))
}

// CentsToDollars converts cents to dollars
func CentsToDollars(cents int64) float64 {
	return float64(cents) / 100.0
}
